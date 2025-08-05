package service

import (
	"biostat/constant"
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"bytes"
	"context"
	"encoding/base64"
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
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type GmailSyncService interface {
	GetGmailAuthURL(userId uint64) (string, error)
	CreateGmailServiceClient(accessToken string, googleOauthConfig *oauth2.Config) (*gmail.Service, error)
	CreateGmailServiceFromToken(ctx context.Context, accessToken string) (*gmail.Service, error)
	FetchEmailsWithAttachment(service *gmail.Service, userId uint64, filterString string, processID uuid.UUID) ([]*models.TblMedicalRecord, error)
	ExtractAttachments(service *gmail.Service, message *gmail.Message, userEmail string, userId uint64, processID uuid.UUID, index int) []*models.TblMedicalRecord
	CreateGmailServiceForApp(userID uint64, accessToken string) (*gmail.Service, error)
	SyncGmailWeb(userID uint64, code string) error
	SyncGmailApp(userID uint64, service *gmail.Service) error
}

type GmailSyncServiceImpl struct {
	processStatusService ProcessStatusService
	medRecordService     TblMedicalRecordService
	userService          UserService
	diagnosticRepo       repository.DiagnosticRepository
}

func NewGmailSyncService(processStatusService ProcessStatusService, medRecordService TblMedicalRecordService, userService UserService, diagnosticRepo repository.DiagnosticRepository) GmailSyncService {
	return &GmailSyncServiceImpl{processStatusService: processStatusService, medRecordService: medRecordService, userService: userService, diagnosticRepo: diagnosticRepo}
}

func (gs *GmailSyncServiceImpl) GetGmailAuthURL(userId uint64) (string, error) {
	_, err := gs.diagnosticRepo.GetPatientLabNameAndEmail(userId)
	if err != nil {
		return "", err
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
	authURL := googleOauthConfig.AuthCodeURL(strconv.FormatUint(userId, 10), oauth2.AccessTypeOffline)
	return authURL, nil
}

func (s *GmailSyncServiceImpl) CreateGmailServiceClient(accessToken string, googleOauthConfig *oauth2.Config) (*gmail.Service, error) {
	creds := oauth2.Token{AccessToken: accessToken}
	client := googleOauthConfig.Client(context.Background(), &creds)
	return gmail.New(client)
}

func (s *GmailSyncServiceImpl) CreateGmailServiceFromToken(ctx context.Context, accessToken string) (*gmail.Service, error) {
	token := &oauth2.Token{AccessToken: accessToken}
	config := &oauth2.Config{}
	client := config.Client(ctx, token)
	return gmail.New(client)
}

func (s *GmailSyncServiceImpl) CreateGmailServiceForApp(userID uint64, accessToken string) (*gmail.Service, error) {
	context := context.Background()
	log.Println("UserID:", userID, " accessToken:", accessToken)
	gmailService, err := s.CreateGmailServiceFromToken(context, accessToken)
	if err != nil {
		log.Println("@GmailServiceApp->CreateGmailServiceFromToken: ", userID, " : ", err)
		return nil, err
	}
	_, err = s.diagnosticRepo.GetPatientLabNameAndEmail(userID)
	if err != nil {
		log.Println("@GmailServiceApp->GetPatientLabNameAndEmail: ", userID, " : ", err)
		return nil, err
	}
	profile, err := gmailService.Users.GetProfile("me").Do()
	if err != nil {
		return nil, err
	}
	log.Println("@GmailServiceApp->Starting Sync For:", profile.EmailAddress)
	return gmailService, nil
}

func (s *GmailSyncServiceImpl) FetchEmailsWithAttachment(service *gmail.Service, userId uint64, filterString string, processID uuid.UUID) ([]*models.TblMedicalRecord, error) {
	log.Println("@FetchEmailsWithAttachments:", userId)
	errorMsg := ""
	profile, err := service.Users.GetProfile("me").Do()
	if err != nil {
		log.Println("@FetchEmailsWithAttachments->service.Users.GetProfile:", userId, ":", err)
		return nil, err
	}
	msg := string(constant.GmailSearchMessage)
	step := string(constant.ProcessGmailSearch)
	userEmail := profile.EmailAddress
	msg1 := fmt.Sprintf("%s Inbox Search Query : %s", msg, filterString)
	log.Println("Inbox Search Query:", userId, ":", filterString)
	s.logStep(processID, step, constant.Running, msg1, errorMsg)
	results, err := service.Users.Messages.List("me").Q(filterString).Do()
	if err != nil {
		log.Println("@FetchEmailsWithAttachments->service.Users.Messages:", userId, ":", err)
		s.logStepAndFail(processID, step, constant.Failure, string(constant.GmailSearchMessage), err.Error())
		return nil, err
	}
	s.logStep(processID, step, constant.Success, msg1, errorMsg)
	var records []*models.TblMedicalRecord
	findMailstep := string(constant.FindingEmailWithAttachment)
	for idx, msg := range results.Messages {
		message, err := service.Users.Messages.Get("me", msg.Id).Do()
		if err != nil {
			continue
		}
		msg := fmt.Sprintf("Processing Mail: %d / %d", idx+1, len(results.Messages))
		s.logStep(processID, findMailstep, constant.Running, msg, "")
		log.Println("@FetchEmailsWithAttachments->ExtractAttachments:", userId, ":", userEmail, ": Processing Mail", idx+1, "/", len(results.Messages))
		attachments := s.ExtractAttachments(service, message, userEmail, userId, processID, idx+1)
		records = append(records, attachments...)
	}
	msg3 := fmt.Sprintf("Processed Mail: %d", len(records))
	// count := len(records)
	s.logStep(processID, findMailstep, constant.Success, msg3, errorMsg)
	log.Println("@FetchEmailsWithAttachments->Gmail Records found:", len(records), "userEmail: ", userEmail)
	return records, nil
}

func (s *GmailSyncServiceImpl) ExtractAttachments(service *gmail.Service, message *gmail.Message, userEmail string, userId uint64, processID uuid.UUID, mailIdx int) []*models.TblMedicalRecord {

	var records []*models.TblMedicalRecord
	totalAttempted := 0
	successCount := 0
	docCompleteMsg := string(constant.CheckDocTypeCompleted)
	step := string(constant.DownloadAttachment)
	step1 := string(constant.CheckDocType)
	for idx, part := range message.Payload.Parts {
		if part.Filename != "" {
			msg := string(constant.DownloadAttachmentMessage)
			msg1 := string(constant.CheckDocTypeMessage)
			log.Println("@ExtractAttachments Processing Record from Email:", idx+1, "/", len(message.Payload.Parts), ": ", part.Filename, "-", getHeader(message.Payload.Headers, "Subject"))
			s.logStep(processID, step, constant.Running, msg, "")
			s.logStep(processID, step1, constant.Running, msg1, "")
			attachmentData, err := DownloadAttachment(service, message.Id, part.Body.AttachmentId)
			if err != nil {
				log.Printf("@ExtractAttachments->DownloadAttachment %s: %v", part.Filename, err)
				continue
			}
			totalAttempted++
			decodedName, err := url.QueryUnescape(part.Filename)
			if err != nil {
				decodedName = part.Filename
			}
			decodedName = strings.ReplaceAll(decodedName, " ", "_")
			re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
			safeName := re.ReplaceAllString(decodedName, "_")
			originalName := strings.TrimSuffix(safeName, filepath.Ext(safeName))
			extension := filepath.Ext(part.Filename)
			uniqueSuffix := time.Now().Format("20060102150405") + "-" + uuid.New().String()[:8]
			safeFileName := fmt.Sprintf("%s_%s%s", originalName, uniqueSuffix, extension)
			docTypeResp, err := utils.CallDocumentTypeAPI(bytes.NewReader(attachmentData), safeFileName)
			if err != nil {
				log.Printf("@ExtractAttachments->utils.CallDocumentTypeAPI type:%s %v ", part.Filename, err)
				s.logStepAndFail(processID, step1, constant.Failure, string(constant.CheckDocTypeFailedMessage), err.Error())
				continue
			}

			destinationPath := filepath.Join("uploads", safeFileName)

			if err := os.WriteFile(destinationPath, attachmentData, 0644); err != nil {
				log.Printf("@ExtractAttachments->Failed to save attachment locally %s: %v", part.Filename, err)
				continue
			}

			newRecord := &models.TblMedicalRecord{
				RecordName:        safeFileName,
				RecordSize:        int64(len(attachmentData)),
				FileType:          part.MimeType,
				RecordUrl:         fmt.Sprintf("%s/uploads/%s", os.Getenv("SHORT_URL_BASE"), safeFileName),
				UploadDestination: "LocalServer",
				Description:       getHeader(message.Payload.Headers, "Subject"),
				UploadSource:      "Gmail",
				RecordCategory:    docTypeResp,
				SourceAccount:     userEmail,
				Status:            constant.StatusQueued,
				UploadedBy:        userId,
				FetchedAt:         time.Now(),
			}
			successCount++
			records = append(records, newRecord)
			newmsg := fmt.Sprintf("%s : Docs type found :> %s", docCompleteMsg, docTypeResp)
			s.logStep(processID, step, constant.Running, msg, "")
			s.logStep(processID, step1, constant.Running, newmsg, "")
		}
	}
	// count := len(records)
	// failedCount := totalAttempted - successCount
	s.logStep(processID, step, constant.Success, string(constant.DownloadAttachmentComplete), "")
	s.logStep(processID, step1, constant.Success, docCompleteMsg, "")
	return records
}

func DownloadAttachment(service *gmail.Service, messageID, attachmentID string) ([]byte, error) {
	attachment, err := service.Users.Messages.Attachments.Get("me", messageID, attachmentID).Do()
	if err != nil {
		return nil, err
	}

	decodedData, err := base64.URLEncoding.DecodeString(attachment.Data)
	if err != nil {
		return nil, err
	}

	return decodedData, nil
}

func getHeader(headers []*gmail.MessagePartHeader, name string) string {
	for _, h := range headers {
		if h.Name == name {
			return h.Value
		}
	}
	return ""
}

func (gs *GmailSyncServiceImpl) GmailSyncCore(userId uint64, processID uuid.UUID, gmailService *gmail.Service) error {
	msg := string(constant.FetchUserLab)
	step := string(constant.ProcessFetchLabs)
	errorMsg := ""
	gs.logStep(processID, step, constant.Success, msg, errorMsg)
	labs, err := gs.diagnosticRepo.GetPatientLabNameAndEmail(userId)
	if err != nil {
		gs.logStepAndFail(processID, step, constant.Failure, string(constant.UserLabNotFound), err.Error())
		log.Println("@GmailSyncCore->GetPatientLabNameAndEmail:", userId, " err:", err)
		return err
	} else {
		gs.logStep(processID, step, constant.Success, string(constant.UserLabFetched), errorMsg)
	}
	filterString := utils.FormatLabsForGmailFilter(labs)
	profile, err := gmailService.Users.GetProfile("me").Do()
	if err != nil {
		step := string(constant.ProcessVerifyCredentials)
		gs.logStep(processID, step, constant.Success, string(constant.InvalidCredentials), errorMsg)
		log.Println("@GmailSyncCore->gmailService.Users.GetProfile:", userId, " err:", err)
		return err
	}
	step1 := string(constant.ProcessFetchEmails)
	gs.logStep(processID, step1, constant.Success, string(constant.FetchEmailAttachment), errorMsg)
	emailMedRecord, err := gs.FetchEmailsWithAttachment(gmailService, userId, filterString, processID)
	if err != nil {
		msg := "No valid medical records were found during this current Gmail sync."
		gs.logStepAndFail(processID, step, constant.Failure, msg, err.Error())
		log.Println("@GmailSyncCore->FetchEmailsWithAttachments:", userId, " err:", err)
		return err
	}
	gs.logStep(processID, step1, constant.Success, string(constant.EmailAttachmentFetch), errorMsg)
	step3 := string(constant.ProcessSaveRecords)
	msg3 := string(constant.SaveRecord)
	// totalRecord := len(emailMedRecord)
	gs.logStep(processID, step3, constant.Running, msg3, errorMsg)

	saveRecordErr := gs.medRecordService.SaveMedicalRecords(emailMedRecord, userId)
	if saveRecordErr != nil {
		log.Println("@GmailSyncCore->SaveMedicalRecords:", userId, " : ", saveRecordErr)
		gs.logStepAndFail(processID, step3, constant.Failure, string(constant.FailedSaveRecords), saveRecordErr.Error())
		return saveRecordErr
	}
	gs.logStep(processID, step3, constant.Success, msg, errorMsg)
	step4 := string(constant.ProcessDigitization)
	msg4 := string(constant.DigitizationTaskQueue)
	gs.logStep(processID, step4, constant.Running, msg4, errorMsg)
	log.Println("Email sync completed for user:", userId, " : ", profile.EmailAddress)

	userInfo, err := gs.userService.GetSystemUserInfoByUserID(userId)
	if err != nil {
		gs.logStepAndFail(processID, step, constant.Failure, string(constant.UserProfileNotFound), err.Error())
		log.Println("@GmailSyncCore->GetSystemUserInfoByUserID:", userId, " : ", err)
		return err
	}
	for idx, record := range emailMedRecord {
		gs.logStep(processID, step4, constant.Running, fmt.Sprintf("Starting to Digitize saved record: %d / %d", idx+1, len(emailMedRecord)), errorMsg)
		log.Println("Starting to Digitize saved record :", record.RecordId)
		resp, err := http.Get(record.RecordUrl)
		if err != nil {
			log.Printf("Error @GmailSyncCore->Read File From URL:%v", err)
			continue
		}
		defer resp.Body.Close()

		fileBuf := new(bytes.Buffer)
		_, _ = io.Copy(fileBuf, resp.Body)
		filename := filepath.Base(record.RecordUrl)
		taskErr := gs.medRecordService.CreateDigitizationTask(record, userInfo, userId, userInfo.AuthUserId, fileBuf, filename)
		if taskErr != nil {
			log.Println("Error @GmailSyncCore->CreateDigitizationTask: ", taskErr)
		}
	}
	msg5 := fmt.Sprintf("Gmail Sync completed for %d records. These records are now being processed for digitization. You’ll be notified once the process is complete.", len(emailMedRecord))
	gs.logStep(processID, step4, constant.Success, msg5, errorMsg)
	return nil
}

func (s *GmailSyncServiceImpl) SyncGmailWeb(userID uint64, code string) error {
	// Start the process in Redis
	errorMsg := ""
	processIdKey, _ := s.processStatusService.StartProcessInRedis(
		userID,
		string(constant.GmailSync),
		strconv.FormatUint(userID, 10),
		string(constant.MedicalRecordEntity),
		string(constant.ProcessTokenExchange),
	)

	// Step: Token Exchange
	token, err := s.exchangeGoogleToken(code)
	step := string(constant.ProcessTokenExchange)
	if err != nil {
		s.logStepAndFail(processIdKey, step, constant.Failure, string(constant.TokenExchangeFailed), err.Error())
		return err
	}
	s.logStep(processIdKey, step, constant.Success, string(constant.TokenExchangeSuccess), errorMsg)

	// Step: Save User Token
	_, tokenErr := s.userService.CreateTblUserToken(&models.TblUserToken{
		UserId:    userID,
		AuthToken: token.AccessToken,
		Provider:  "Gmail",
	})
	if tokenErr != nil {
		s.logStep(processIdKey, step, constant.Failure, "Failed to save user token: ", tokenErr.Error())
	} else {
		s.logStep(processIdKey, step, constant.Success, "User token fetch successfully", "")
	}

	// Step: Create Gmail client
	step = string(constant.ProcessGmailClient)
	gmailService, err := s.CreateGmailServiceClient(token.AccessToken, s.googleOauthConfig())
	if err != nil {
		s.logStepAndFail(processIdKey, step, constant.Failure, string(constant.GmailClientCreateFailed), err.Error())
		return err
	}
	s.logStep(processIdKey, step, constant.Success, string(constant.GmailClientCreated), errorMsg)
	return s.GmailSyncCore(userID, processIdKey, gmailService)
}

// Helper: Exchange Google OAuth token
func (s *GmailSyncServiceImpl) exchangeGoogleToken(code string) (*oauth2.Token, error) {
	return s.googleOauthConfig().Exchange(context.Background(), code)
}

// Helper: Returns configured OAuth2 config
func (s *GmailSyncServiceImpl) googleOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
		Scopes:       []string{"https://mail.google.com/"},
		Endpoint:     google.Endpoint,
	}
}

// Helper: Log step success or failure (non-terminal)
func (s *GmailSyncServiceImpl) logStep(processID uuid.UUID, step, status, message, errorMsg string) {
	var errPtr *string
	if errorMsg != "" {
		errPtr = &errorMsg
	}
	logEntry := models.ProcessStepLog{
		ProcessStepLogId: uuid.New(),
		ProcessStatusID:  processID,
		Step:             step,
		Status:           status,
		Message:          message,
		Error:            errPtr,
		StepStartedAt:    time.Now(),
		StepUpdatedAt:    time.Now(),
	}
	if err := s.processStatusService.AddOrUpdateStepLogInRedis(processID, logEntry); err != nil {
		log.Printf("[Warning] failed to add/update step log in Redis: %v", err)
	}
	if err := s.processStatusService.UpdateProcessStatusInRedis(processID, status, message, step, false); err != nil {
		log.Printf("[Warning] failed to update process status in Redis: %v", err)
	}
}

// Helper: Log step and mark process as failed (terminal)
func (s *GmailSyncServiceImpl) logStepAndFail(processID uuid.UUID, step, status, message, errorMsg string) {
	s.logStep(processID, step, status, message, errorMsg)
	if err := s.processStatusService.UpdateProcessStatusInRedis(processID, constant.Failure, message, step, true); err != nil {
		log.Printf("[Error] failed to mark process failure in Redis: %v", err)
	}
}

func (s *GmailSyncServiceImpl) SyncGmailApp(userID uint64, gmailService *gmail.Service) error {
	processID := uuid.New()
	_, _ = s.processStatusService.StartProcess(processID, userID, string(constant.GmailSync), strconv.FormatUint(userID, 10), string(constant.MedicalRecordEntity), string(constant.ProcessTokenExchange))
	key, err := s.processStatusService.StartProcessRedis(processID, userID, string(constant.GmailSync), strconv.FormatUint(userID, 10), string(constant.MedicalRecordEntity), string(constant.ProcessTokenExchange))
	if err != nil {
		log.Println("@SyncGmailApp->StartProcessRedis:", err)
		return err
	}
	log.Panicln("key ", key)
	return s.GmailSyncCore(userID, processID, gmailService)
}
