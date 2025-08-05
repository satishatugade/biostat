package controller

import (
	"biostat/constant"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GmailSyncController struct {
	gmailSyncService service.GmailSyncService
	service          service.TblMedicalRecordService
	gTokenService    service.UserService
	healthMonitor    *service.HealthMonitorService
}

func NewGmailSyncController(gmailSyncService service.GmailSyncService, service service.TblMedicalRecordService,
	gTokenService service.UserService, healthMonitor *service.HealthMonitorService) *GmailSyncController {
	return &GmailSyncController{
		gmailSyncService: gmailSyncService,
		service:          service,
		gTokenService:    gTokenService,
		healthMonitor:    healthMonitor,
	}
}

func (gc *GmailSyncController) GmailLoginHandler(ctx *gin.Context) {
	userId := utils.GetParamAsUInt(ctx, "user_id")
	if userId == 0 {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "Error while syncing", nil, errors.New("User Id missing in request"))
		return
	}
	authURL, err := gc.gmailSyncService.GetGmailAuthURL(userId)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Gmail couldn't be synced", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "", gin.H{"auth_url": authURL}, nil, nil)
	return
}

func (c *GmailSyncController) GmailCallbackHandler(ctx *gin.Context) {
	code := ctx.Query("code")
	userID := ctx.Query("state")

	if code == "" {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Error authenticating", nil, errors.New("Missing auth code"))
		return
	}

	userIDInt64, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Error while authenticating", nil, errors.New("Invalid user id"))
	}
	if !c.healthMonitor.IsServiceUp() {
		log.Printf("[GmailCallback] AI service is down. Request: %s\n", ctx.Request.URL.String())
		redirectURL := fmt.Sprintf("%s/dashboard/medical-reports?status=ai_down", os.Getenv("APP_URL"))
		ctx.Redirect(http.StatusFound, redirectURL)
		return
	}
	log.Printf("[GmailCallback] AI service is UP. Request: %s%s\n", ctx.Request.Host, ctx.Request.URL.String())
	ctx.Redirect(http.StatusFound, fmt.Sprintf(os.Getenv("APP_URL")+"/dashboard/medical-reports?status=processing"))

	go func() {
		if err := c.gmailSyncService.SyncGmailWeb(userIDInt64, code); err != nil {
			log.Println("Gmail sync error:", err)
		}
	}()
}

func (c *GmailSyncController) FetchEmailsHandlerApp(ctx *gin.Context) {
	var req models.GmailSyncRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "invalid request body", nil, err)
		return
	}
	gmailService, err := c.gmailSyncService.CreateGmailServiceForApp(req.UserID, req.AccessToken)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "Gmail couldn't be synced", nil, err)
		return
	}
	go func() {
		if err := c.gmailSyncService.SyncGmailApp(req.UserID, gmailService); err != nil {
			log.Println("FetchEmailsHandlerApp:", err)
		}
	}()

	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "", gin.H{"message": "Gmail syncing process started will update you once done"}, nil, nil)
	return

}
