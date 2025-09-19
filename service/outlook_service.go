package service

import (
	"biostat/constant"
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type OutLookService interface {
	GetOutLookAuthURL(userID uint64) (string, error)
	GetOutLookToken(ctx context.Context, code string) (*oauth2.Token, error)
	SyncOutLookWeb(ctx context.Context, userId uint64, token *oauth2.Token) error
}

type OutLookServiceImpl struct {
	userService          UserService
	apiService           ApiService
	processStatusService ProcessStatusService
	gmailSyncService     GmailSyncService
	diagnosticRepo       repository.DiagnosticRepository
}

func NewOutLookService(userService UserService, apiService ApiService, processStatusService ProcessStatusService, gmailSyncService GmailSyncService, diagnosticRepo repository.DiagnosticRepository) OutLookService {
	return &OutLookServiceImpl{userService: userService, apiService: apiService, processStatusService: processStatusService, gmailSyncService: gmailSyncService, diagnosticRepo: diagnosticRepo}
}

func (ols *OutLookServiceImpl) GetOutLookAuthURL(userID uint64) (string, error) {
	_, err := ols.diagnosticRepo.GetPatientLabNameAndEmail(userID)
	if err != nil {
		return "", err
	}
	var oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("AZURE_REDIRECT_URL"),
		Scopes:       []string{"openid", "profile", "offline_access", "User.Read", "Mail.Read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		},
	}

	authURL := oauthConfig.AuthCodeURL(strconv.FormatUint(userID, 10),
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)
	return authURL, nil
}

func (ols *OutLookServiceImpl) GetOutLookToken(ctx context.Context, code string) (*oauth2.Token, error) {
	var oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("AZURE_REDIRECT_URL"),
		Scopes:       []string{"openid", "profile", "offline_access", "User.Read", "Mail.Read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		},
	}
	log.Println("GetOutLookToken:Code:", code)
	token, err := oauthConfig.Exchange(ctx, code)
	return token, err
}

func GetOutlookEmailFromToken(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch user info, status: %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if email, ok := data["mail"].(string); ok && email != "" {
		return email, nil
	}
	if upn, ok := data["userPrincipalName"].(string); ok && upn != "" {
		return upn, nil
	}

	return "", fmt.Errorf("email not found in token response")
}

func (ols *OutLookServiceImpl) SyncOutLookWeb(ctx context.Context, userID uint64, token *oauth2.Token) error {
	processIdKey, _ := ols.processStatusService.StartProcessInRedis(
		userID,
		string(constant.GmailSync),
		strconv.FormatUint(userID, 10),
		string(constant.MedicalRecordEntity),
		string(constant.ProcessTokenExchange),
	)
	step := string(constant.ProcessTokenExchange)
	email, err := GetOutlookEmailFromToken(token.AccessToken)
	if err != nil {
		ols.processStatusService.LogStepAndFail(processIdKey, step, constant.Failure, string(constant.TokenExchangeFailed), err.Error(), nil, nil, nil)
		return fmt.Errorf("failed to fetch email: %w", err)
	}
	_, tokenErr := ols.userService.CreateTblUserToken(&models.TblUserToken{
		UserId:       userID,
		AuthToken:    token.AccessToken,
		RefreshToken: token.RefreshToken,
		Provider:     "OutLook",
		ProviderId:   email,
		ExpiresAt:    token.Expiry,
	})
	if tokenErr != nil {
		ols.processStatusService.LogStepAndFail(processIdKey, step, constant.Failure, "Failed to save user token: ", tokenErr.Error(), nil, nil, nil)
	} else {
		ols.processStatusService.LogStep(processIdKey, step, constant.Success, "User token fetch successfully", "", nil, nil, nil, nil, nil, nil)
	}
	allEmailRecords, err := ols.GetOutLookRecords(userID, processIdKey, ctx, token.AccessToken)
	if err != nil {
		step := string(constant.ProcessFetchLabs)
		msg := "No valid medical records were found during this current Gmail sync."
		ols.processStatusService.LogStepAndFail(processIdKey, step, constant.Failure, msg, err.Error(), nil, nil, nil)
		log.Println("@GmailSyncCore->FetchEmailsWithAttachments:", userID, " err:", err)
		return err
	}
	log.Println("Records fetched:", len(allEmailRecords))
	return ols.gmailSyncService.GmailSyncCore(userID, processIdKey, allEmailRecords)
}

func (ols *OutLookServiceImpl) GetOutLookRecords(userId uint64, processID uuid.UUID, ctx context.Context, accessToken string) ([]*models.TblMedicalRecord, error) {
	msg := string(constant.FetchUserLab)
	step := string(constant.ProcessFetchLabs)
	errorMsg := ""
	ols.processStatusService.LogStep(processID, step, constant.Success, msg, errorMsg, nil, nil, nil, nil, nil, nil)
	log.Println("Fetching Labss")
	labs, err := ols.diagnosticRepo.GetPatientLabNameAndEmail(userId)
	var labNames []string
	for _, lab := range labs {
		labNames = append(labNames, lab.LabName)
	}
	if err != nil {
		ols.processStatusService.LogStepAndFail(processID, step, constant.Failure, string(constant.UserLabNotFound), err.Error(), nil, nil, nil)
		log.Println("@GmailSyncCore->GetPatientLabNameAndEmail:", userId, " err:", err)
		return nil, err
	} else {
		labMsg := fmt.Sprintf("Total %d labs fetched %s ", len(labNames), strings.Join(labNames, " | "))
		ols.processStatusService.LogStep(processID, step, constant.Success, labMsg, errorMsg, nil, nil, nil, nil, nil, nil)
	}
	filterString := utils.FormatLabsForOutlookFilter(labs)
	log.Println("Filter string:", filterString)
	return ols.FetchEmailsWithFilter(ctx, accessToken, filterString, userId, processID)
}

func (s *OutLookServiceImpl) FetchEmailsWithFilter(ctx context.Context, accessToken, filterString string, userId uint64, processID uuid.UUID) ([]*models.TblMedicalRecord, error) {
	errorMsg := ""
	stepSearch := string(constant.ProcessGmailSearch)
	stepFindMail := string(constant.FindingEmailWithAttachment)
	stepFetchList := string(constant.FetchEmailsList)
	stepDownload := string(constant.DownloadAttachment)

	msg := string(constant.GmailSearchMessage)
	msg1 := fmt.Sprintf("%s Inbox Search Query : %s", msg, filterString)
	log.Println("Inbox Search Query:", userId, ":", filterString)
	s.processStatusService.LogStep(processID, stepSearch, constant.Running, msg1, errorMsg, nil, nil, nil, nil, nil, nil)

	baseURL := "https://graph.microsoft.com/v1.0/me/messages"
	query := url.Values{}
	query.Set("$top", "50")
	if filterString != "" {
		query.Set("$filter", filterString)
	}
	requestURL := fmt.Sprintf("%s?%s", baseURL, query.Encode())
	log.Println("request URL", requestURL)
	var allMessages []models.OutlookMessage
	for requestURL != "" {
		req, _ := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Prefer", "outlook.body-content-type=\"text\"")
		req.Header.Set("ConsistencyLevel", "eventual")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			s.processStatusService.LogStepAndFail(processID, stepSearch, constant.Failure, msg1, err.Error(), nil, nil, nil)
			return nil, fmt.Errorf("failed request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			apiErr := fmt.Sprintf("graph API error: %s", string(body))
			s.processStatusService.LogStepAndFail(processID, stepSearch, constant.Failure, msg1, apiErr, nil, nil, nil)
			return nil, fmt.Errorf("graph API error: %s", string(body))
		}

		var result models.OutlookMessagesResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			s.processStatusService.LogStepAndFail(processID, stepSearch, constant.Failure, msg1, err.Error(), nil, nil, nil)
			return nil, err
		}

		allMessages = append(allMessages, result.Value...)
		requestURL = result.NextLink
	}
	s.processStatusService.LogStep(processID, stepSearch, constant.Success, msg1, errorMsg, nil, nil, nil, nil, nil, nil)
	log.Println("Got emails from Outlook:", len(allMessages))
	var emailSummaries []string
	var allRecords []*models.TblMedicalRecord
	for idx, msg := range allMessages {
		indexCount := idx

		emailSummaries = append(emailSummaries,
			fmt.Sprintf("%d : from %s: %s", idx+1, msg.From.EmailAddress.Address, msg.Subject))

		log.Println(idx+1, ": Checking ", msg.Received, "||", msg.Subject)

		if !msg.HasAttachments {
			s.processStatusService.LogStep(processID, stepFindMail, constant.Running,
				fmt.Sprintf("No attachments found in email subject %s dated %s", msg.Subject, msg.Received),
				errorMsg, nil, &indexCount, nil, nil, nil, &msg.ID)
			continue
		}
		attachments, err := s.FetchAttachments(ctx, accessToken, msg.ID)
		if err != nil {
			log.Printf("Error fetching attachments for msg %s %s: %v ", msg.Received, msg.Subject, err)
			s.processStatusService.LogStepAndFail(processID, stepFindMail, constant.Failure,
				fmt.Sprintf("Error fetching attachments for email subject %s dated %s", msg.Subject, msg.Received),
				err.Error(), nil, nil, nil)
			continue
		}
		log.Println("Processing Outlook Email for:", userId, ": Processing Mail", idx+1, " of ", len(allMessages))
		msgInfo := fmt.Sprintf("Subject %s dated %s and %d attachments found",
			msg.Subject, msg.Received, len(attachments))
		s.processStatusService.LogStep(processID, stepFindMail, constant.Running, msgInfo, errorMsg, nil, &indexCount, nil, nil, nil, &msg.ID)
		totalAttempted := 0
		successCount := 0
		for attIdx, att := range attachments {
			recordIndexCount := attIdx + 1
			log.Println("@ExtractAttachments Processing Record from Email:", idx+1, "-", recordIndexCount, "/", len(attachments), ": ", att.Name, "-", msg.Subject)

			record, err := s.saveAttachment(att, msg, 1)
			if err != nil {
				log.Printf("Error saving attachment %s: %v", att.Name, err)
				s.processStatusService.LogStep(processID, stepDownload, constant.Failure, fmt.Sprintf("Error Downloading attachment %s from %s dated %s", att.Name, msg.Subject, msg.Received), err.Error(), nil, &recordIndexCount, &recordIndexCount, nil, nil, &att.ID)
				continue
			}
			totalAttempted++
			successCount++
			allRecords = append(allRecords, record)
			s.processStatusService.LogStep(processID, stepDownload, constant.Running,
				fmt.Sprintf("Downloaded attachment %s Dated on %s from EmailSub %s", att.Name, msg.Received, msg.Subject),
				errorMsg, nil, &recordIndexCount, &recordIndexCount, nil, nil, &att.ID)
		}
		count := successCount
		failedCount := totalAttempted - successCount
		s.processStatusService.LogStep(processID, stepDownload, constant.Success, string(constant.DownloadAttachmentComplete),
			errorMsg, nil, nil, &count, &successCount, &failedCount, nil)
	}
	logMsg := strings.Join(emailSummaries, " | ")
	log.Println(logMsg)
	s.processStatusService.LogStep(processID, stepFetchList, constant.Success,
		fmt.Sprintf("%d emails found: %s", len(emailSummaries), logMsg),
		errorMsg, nil, nil, nil, nil, nil, nil)
	msg3 := fmt.Sprintf("Found %d email attachments in %d emails", len(allRecords), len(allMessages))
	totalRecord := len(allRecords)
	s.processStatusService.LogStep(processID, stepFindMail, constant.Success, msg3,
		errorMsg, nil, &totalRecord, nil, nil, nil, nil)

	log.Println("@FetchEmailsWithFilter->Outlook Records found:", len(allRecords))
	return allRecords, nil
}

func (s *OutLookServiceImpl) FetchAttachments(ctx context.Context, accessToken, messageID string) ([]models.OutlookAttachment, error) {
	requestURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/messages/%s/attachments", messageID)

	req, _ := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

		return nil, fmt.Errorf("fetch attachments error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("graph API error: %s", string(body))
	}

	var result struct {
		Value []models.OutlookAttachment `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Value, nil
}

func (s *OutLookServiceImpl) saveAttachment(att models.OutlookAttachment, msg models.OutlookMessage, userId uint64) (*models.TblMedicalRecord, error) {
	data, err := base64.StdEncoding.DecodeString(att.ContentBytes)
	if err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	decodedName := strings.ReplaceAll(att.Name, " ", "_")
	extension := filepath.Ext(decodedName)
	nameOnly := strings.TrimSuffix(decodedName, extension)
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	safeName := re.ReplaceAllString(nameOnly, "_")
	uniqueSuffix := time.Now().Format("20060102150405") + "-" + uuid.New().String()[:8]
	safeFileName := fmt.Sprintf("%s_%s%s", safeName, uniqueSuffix, extension)
	destPath := filepath.Join("uploads", safeFileName)

	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return nil, fmt.Errorf("save file error: %w", err)
	}

	recordURL := fmt.Sprintf("%s/uploads/%s", os.Getenv("SHORT_URL_BASE"), safeFileName)

	metadata := map[string]interface{}{
		"attachment_id": att.ID,
		"message_id":    msg.ID,
	}
	metadataJSON, _ := json.Marshal(metadata)

	return &models.TblMedicalRecord{
		RecordName:        safeFileName,
		RecordSize:        int64(len(data)),
		FileType:          att.ContentType,
		UploadSource:      "Outlook",
		UploadDestination: "LocalServer",
		SourceAccount:     msg.From.EmailAddress.Address,
		RecordCategory:    string(constant.OTHER),
		Description:       fmt.Sprintf("Subject: %s | BodyPreview: %s", msg.Subject, msg.BodyPreview),
		UDF1:              msg.Subject,
		UDF2:              msg.Received,
		RecordUrl:         recordURL,
		FetchedAt:         time.Now(),
		UploadedBy:        userId,
		Status:            constant.StatusProcessing,
		Metadata:          metadataJSON,
	}, nil
}
