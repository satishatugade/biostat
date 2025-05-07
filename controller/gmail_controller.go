package controller

import (
	"biostat/constant"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GmailSyncController struct {
	service       service.TblMedicalRecordService
	gTokenService service.UserService
}

func NewGmailSyncController(service service.TblMedicalRecordService, gTokenService service.UserService) *GmailSyncController {
	return &GmailSyncController{
		service:       service,
		gTokenService: gTokenService,
	}
}

func GmailLoginHandler(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	if userID == "" {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "User Id missing in request", nil, errors.New("User Id missing in request"))
		return
	}
	authURL := service.GetGmailAuthURL(userID)
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "", gin.H{"auth_url": authURL}, nil, nil)
}

func (c *GmailSyncController) GmailCallbackHandler(ctx *gin.Context) {
	code := ctx.Query("code")
	userID := ctx.Query("state")

	if code == "" {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Error authenticating", nil, errors.New("Missing auth code"))
		return
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URI")
	var googleOauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"https://mail.google.com/"},
		Endpoint:     google.Endpoint,
	}

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Error while authenticating", nil, errors.New("Token exchange failed"))
		return
	}

	userIDInt64, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Error while authenticating", nil, errors.New("Invalid user id"))
	}
	c.gTokenService.CreateTblUserToken(&models.TblUserToken{UserId: userIDInt64, AuthToken: token.AccessToken, Provider: "Gmail"})

	ctx.Redirect(http.StatusFound, fmt.Sprintf(os.Getenv("APP_URL")+"dashboard/medical-reports?status=processing"))

	go func(userID uint64, authToken string) {
		log.Println("Starting background email sync for user:", userID)

		gmailService, err := service.CreateGmailServiceClient(authToken, googleOauthConfig)
		if err != nil {
			log.Println("Failed to create Gmail client:", err)
			return
		}
		accessToken, _ := c.gTokenService.GetSingleTblUserToken(userID, "DigiLocker")

		emailMedRecord, err := service.FetchEmailsWithAttachments(gmailService, userID, accessToken.AuthToken)
		if err != nil {
			log.Println("Failed to fetch emails:", err)
			return
		}

		limit := 5
		if len(emailMedRecord) < limit {
			limit = len(emailMedRecord)
		}
		first5Emails := emailMedRecord[:limit]
		
		log.Println("Following email models will be saved:", len(first5Emails))

		err = c.service.SaveMedicalRecords(&first5Emails, userID)
		if err != nil {
			log.Println("Error while saving email data:", err)
			return
		}

		log.Println("Email sync completed for user:", userID)
	}(userIDInt64, token.AccessToken)
}

// This API is to Fetch Emails directly from userID and saved access token
func (c *GmailSyncController) FetchEmailsHandler(ctx *gin.Context) {
	user_id := utils.GetParamAsUInt(ctx, "user_id")

	gToken, gErr := c.gTokenService.GetSingleTblUserToken(user_id, "Gmail")
	if gErr != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "Error while syncing records", nil, gErr)
		return
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URI")
	var googleOauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"https://mail.google.com/"},
		Endpoint:     google.Endpoint,
	}

	gmailService, err := service.CreateGmailServiceClient(gToken.AuthToken, googleOauthConfig)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to create Gmail client", nil, err)
		return
	}
	accessToken, err := c.gTokenService.GetSingleTblUserToken(user_id, "DigiLocker")
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to fetch emails", nil, err)
		return
	}

	emails, err := service.FetchEmailsWithAttachments(gmailService, user_id, accessToken.AuthToken)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to fetch emails", nil, err)
		return
	}

	first5Emails := emails[:5]
	log.Println("Following email models will be saved:", len(first5Emails))

	err = c.service.SaveMedicalRecords(&first5Emails, user_id)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Error while saving data", nil, err)
	}
	c.gTokenService.DeleteTblUserToken(user_id, "")
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Email Sync completed", nil, nil, nil)
}
