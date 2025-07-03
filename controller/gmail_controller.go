package controller

import (
	"biostat/constant"
	"biostat/models"
	"biostat/service"
	"context"
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
}

func NewGmailSyncController(gmailSyncService service.GmailSyncService, service service.TblMedicalRecordService, gTokenService service.UserService) *GmailSyncController {
	return &GmailSyncController{
		gmailSyncService: gmailSyncService,
		service:          service,
		gTokenService:    gTokenService,
	}
}

func (gc *GmailSyncController) GmailLoginHandler(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	if userID == "" {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "User Id missing in request", nil, errors.New("User Id missing in request"))
		return
	}
	authURL := gc.gmailSyncService.GetGmailAuthURL(userID)
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "", gin.H{"auth_url": authURL}, nil, nil)
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

	ctx.Redirect(http.StatusFound, fmt.Sprintf(os.Getenv("APP_URL")+"/dashboard/medical-reports?status=processing"))

	go func() {
		if err := c.gmailSyncService.SyncGmail(userIDInt64, code); err != nil {
			log.Println("Gmail sync error:", err)
		}
	}()
}

func (c *GmailSyncController) FetchEmailsHandler(ctx *gin.Context) {
	var req models.GmailSyncRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "invalid request body", nil, err)
		return
	}
	context := context.Background()
	gmailService, err := c.gmailSyncService.CreateGmailServiceFromToken(context, req.AccessToken)
	if err != nil {
		log.Println("FetchEmailsHandler->CreateGmailServiceFromToken:", err)
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "failed to create Gmail service", nil, err)
		return
	}

	emailMedRecord, err := c.gmailSyncService.FetchEmailsWithAttachments(gmailService, req.UserID)
	if err != nil {
		log.Println("FetchEmailsHandler->FetchEmailsWithAttachments:", err)
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "failed to fetch emails", nil, err)
		return
	}

	err = c.service.SaveMedicalRecords(&emailMedRecord, req.UserID)
	if err != nil {
		log.Println("FetchEmailsHandler->SaveMedicalRecords:", err)
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Error while saving email data:", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "gmail sync completed", emailMedRecord, nil, nil)
	return

}
