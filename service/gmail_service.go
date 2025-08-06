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
	ExtractAttachment(service *gmail.Service, message *gmail.Message, userEmail string, userId uint64, processID uuid.UUID, index int) []*models.TblMedicalRecord
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
	s.processStatusService.LogStep(processID, step, constant.Running, msg1, errorMsg, nil, nil, nil, nil, nil)
	results, err := service.Users.Messages.List("me").Q(filterString).Do()
	if err != nil {
		log.Println("@FetchEmailsWithAttachments->service.Users.Messages:", userId, ":", err)
		s.processStatusService.LogStepAndFail(processID, step, constant.Failure, string(constant.GmailSearchMessage), err.Error(), nil, nil)
		return nil, err
	}
	s.processStatusService.LogStep(processID, step, constant.Success, msg1, errorMsg, nil, nil, nil, nil, nil)
	var records []*models.TblMedicalRecord
	findMailstep := string(constant.FindingEmailWithAttachment)
	for idx, msg := range results.Messages {
		indexCount := idx + 1
		message, err := service.Users.Messages.Get("me", msg.Id).Do()
		if err != nil {
			continue
		}
		msg := fmt.Sprintf("Processing email attachment: %d / %d", indexCount, len(results.Messages))
		s.processStatusService.LogStep(processID, findMailstep, constant.Running, msg, errorMsg, nil, &indexCount, nil, nil, nil)
		log.Println("@FetchEmailsWithAttachments->ExtractAttachments:", userId, ":", userEmail, ": Processing Mail", idx+1, "/", len(results.Messages))
		attachments := s.ExtractAttachment(service, message, userEmail, userId, processID, idx+1)
		records = append(records, attachments...)
	}
	msg3 := fmt.Sprintf("Processed email attachment: %d", len(records))
	totalRecord := len(records)
	s.processStatusService.LogStep(processID, findMailstep, constant.Success, msg3, errorMsg, nil, &totalRecord, nil, nil, nil)
	log.Println("@FetchEmailsWithAttachments->Gmail Records found:", len(records), "userEmail: ", userEmail)
	return records, nil
}

func (s *GmailSyncServiceImpl) ExtractAttachment(service *gmail.Service, message *gmail.Message, userEmail string, userId uint64, processID uuid.UUID, mailIdx int) []*models.TblMedicalRecord {

	var records []*models.TblMedicalRecord
	totalAttempted := 0
	successCount := 0
	errorMsg := ""
	docCompleteMsg := string(constant.CheckDocTypeCompleted)
	step := string(constant.DownloadAttachment)
	step1 := string(constant.CheckDocType)
	for idx, part := range message.Payload.Parts {
		if part.Filename != "" {
			recordIndexCount := idx + 1
			msg1 := string(constant.CheckDocTypeMessage)
			body := GetMessageBody(message.Payload)
			msg := fmt.Sprintf(string(constant.DownloadAttachmentMessage)+" | Subject and Body of email %d - %d / %d: %s - %s - %+v", mailIdx, recordIndexCount, len(message.Payload.Parts), part.Filename, getHeader(message.Payload.Headers, "Subject"), body)
			log.Println("@ExtractAttachments Processing Record from Email:", mailIdx, "-", recordIndexCount, "/", len(message.Payload.Parts), ": ", part.Filename, "-", getHeader(message.Payload.Headers, "Subject"))
			s.processStatusService.LogStep(processID, step, constant.Running, msg, errorMsg, nil, &recordIndexCount, &recordIndexCount, nil, nil)
			s.processStatusService.LogStep(processID, step1, constant.Running, msg1, errorMsg, nil, &recordIndexCount, &recordIndexCount, nil, nil)
			attachmentData, err := DownloadAttachment(service, message.Id, part.Body.AttachmentId)
			if err != nil {
				log.Printf("@ExtractAttachments->DownloadAttachment %s: %v", part.Filename, err)
				s.processStatusService.LogStep(processID, step, constant.Failure, msg, errorMsg, nil, &recordIndexCount, &recordIndexCount, nil, nil)
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
			docTypeResp := ""
			docTypeResp, err = utils.CallDocumentTypeAPI(bytes.NewReader(attachmentData), safeFileName)
			if err != nil {
				log.Printf("@ExtractAttachments->utils.CallDocumentTypeAPI type:%s %v ", part.Filename, err)
				s.processStatusService.LogStepAndFail(processID, step1, constant.Failure, string(constant.CheckDocTypeFailedMessage), err.Error(), &recordIndexCount, nil)
				docTypeResp = string(constant.OTHER)
			}

			destinationPath := filepath.Join("uploads", safeFileName)

			if err := os.WriteFile(destinationPath, attachmentData, 0644); err != nil {
				log.Printf("@ExtractAttachments->Failed to save attachment locally %s: %v", part.Filename, err)
				continue
			}
			subBody := fmt.Sprintf("Subject and body of email sub : %s : Body :%+v ", getHeader(message.Payload.Headers, "Subject"), body)
			newRecord := &models.TblMedicalRecord{
				RecordName:        safeFileName,
				RecordSize:        int64(len(attachmentData)),
				FileType:          part.MimeType,
				RecordUrl:         fmt.Sprintf("%s/uploads/%s", os.Getenv("SHORT_URL_BASE"), safeFileName),
				UploadDestination: "LocalServer",
				Description:       subBody,
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
			s.processStatusService.LogStep(processID, step, constant.Running, msg, errorMsg, nil, &recordIndexCount, &recordIndexCount, nil, nil)
			s.processStatusService.LogStep(processID, step1, constant.Running, newmsg, errorMsg, nil, &recordIndexCount, &recordIndexCount, nil, nil)
		}
	}
	count := len(records)
	failedCount := totalAttempted - successCount
	s.processStatusService.LogStep(processID, step, constant.Success, string(constant.DownloadAttachmentComplete), errorMsg, nil, nil, &count, &successCount, &failedCount)
	s.processStatusService.LogStep(processID, step1, constant.Success, docCompleteMsg, errorMsg, nil, nil, &totalAttempted, &successCount, &failedCount)
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

func GetMessageBody(payload *gmail.MessagePart) string {
	if payload.Body != nil && payload.Body.Data != "" {
		body, _ := base64.URLEncoding.DecodeString(payload.Body.Data)
		return string(body)
	}

	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
			body, _ := base64.URLEncoding.DecodeString(part.Body.Data)
			return string(body)
		}
		if part.MimeType == "text/html" && part.Body != nil && part.Body.Data != "" {
			body, _ := base64.URLEncoding.DecodeString(part.Body.Data)
			return stripHTML(string(body))
		}
	}

	return ""
}

func stripHTML(input string) string {
	replacer := strings.NewReplacer("<br>", "\n", "<br/>", "\n", "<p>", "\n", "</p>", "", "<div>", "\n", "</div>", "")
	return replacer.Replace(input)
}

func (gs *GmailSyncServiceImpl) GmailSyncCore(userId uint64, processID uuid.UUID, gmailService *gmail.Service) error {
	msg := string(constant.FetchUserLab)
	step := string(constant.ProcessFetchLabs)
	errorMsg := ""
	gs.processStatusService.LogStep(processID, step, constant.Success, msg, errorMsg, nil, nil, nil, nil, nil)
	labs, err := gs.diagnosticRepo.GetPatientLabNameAndEmail(userId)
	if err != nil {
		gs.processStatusService.LogStepAndFail(processID, step, constant.Failure, string(constant.UserLabNotFound), err.Error(), nil, nil)
		log.Println("@GmailSyncCore->GetPatientLabNameAndEmail:", userId, " err:", err)
		return err
	} else {
		gs.processStatusService.LogStep(processID, step, constant.Success, string(constant.UserLabFetched), errorMsg, nil, nil, nil, nil, nil)
	}
	filterString := utils.FormatLabsForGmailFilter(labs)
	profile, err := gmailService.Users.GetProfile("me").Do()
	if err != nil {
		step := string(constant.ProcessVerifyCredentials)
		gs.processStatusService.LogStep(processID, step, constant.Success, string(constant.InvalidCredentials), errorMsg, nil, nil, nil, nil, nil)
		log.Println("@GmailSyncCore->gmailService.Users.GetProfile:", userId, " err:", err)
		return err
	}
	step1 := string(constant.ProcessFetchEmails)
	gs.processStatusService.LogStep(processID, step1, constant.Success, string(constant.FetchEmailAttachment), errorMsg, nil, nil, nil, nil, nil)
	emailMedRecord, err := gs.FetchEmailsWithAttachment(gmailService, userId, filterString, processID)
	if err != nil {
		msg := "No valid medical records were found during this current Gmail sync."
		gs.processStatusService.LogStepAndFail(processID, step, constant.Failure, msg, err.Error(), nil, nil)
		log.Println("@GmailSyncCore->FetchEmailsWithAttachments:", userId, " err:", err)
		return err
	}
	gs.processStatusService.LogStep(processID, step1, constant.Success, string(constant.EmailAttachmentFetch), errorMsg, nil, nil, nil, nil, nil)
	step3 := string(constant.ProcessSaveRecords)
	msg3 := string(constant.SaveRecord)
	totalRecord := len(emailMedRecord)
	gs.processStatusService.LogStep(processID, step3, constant.Running, msg3, errorMsg, nil, nil, &totalRecord, nil, nil)

	saveRecordErr := gs.medRecordService.SaveMedicalRecords(emailMedRecord, userId)
	if saveRecordErr != nil {
		log.Println("@GmailSyncCore->SaveMedicalRecords:", userId, " : ", saveRecordErr)
		gs.processStatusService.LogStepAndFail(processID, step3, constant.Failure, string(constant.FailedSaveRecords), saveRecordErr.Error(), nil, nil)
		return saveRecordErr
	}
	gs.processStatusService.LogStep(processID, step3, constant.Success, msg, errorMsg, nil, nil, nil, nil, nil)
	step4 := string(constant.ProcessDigitization)
	msg4 := string(constant.DigitizationTaskQueue)
	gs.processStatusService.LogStep(processID, step4, constant.Running, msg4, errorMsg, nil, nil, nil, nil, nil)
	log.Println("Email sync completed for user:", userId, " : ", profile.EmailAddress)

	userInfo, err := gs.userService.GetSystemUserInfoByUserID(userId)
	if err != nil {
		gs.processStatusService.LogStepAndFail(processID, step, constant.Failure, string(constant.UserProfileNotFound), err.Error(), nil, nil)
		log.Println("@GmailSyncCore->GetSystemUserInfoByUserID:", userId, " : ", err)
		return err
	}
	for idx, record := range emailMedRecord {
		if record.RecordCategory == string(constant.TESTREPORT) || record.RecordCategory == string(constant.PRESCRIPTION) {
			msg := fmt.Sprintf("Starting digitization for report recordId :%d (category: test report or medication document)%d _ %d", record.RecordId, idx+1, len(emailMedRecord))
			gs.processStatusService.LogStep(processID, step4, constant.Running, msg, errorMsg, &record.RecordId, nil, nil, nil, nil)
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
			taskErr := gs.medRecordService.CreateDigitizationTask(record, userInfo, userId, fileBuf, filename, processID)
			if taskErr != nil {
				log.Println("Error @GmailSyncCore->CreateDigitizationTask: ", taskErr)
				gs.processStatusService.LogStepAndFail(processID, step4, constant.Failure, "Record digitization failed", taskErr.Error(), nil, &record.RecordId)
			}
			// gs.processStatusService.LogStep(processID, step4, constant.Success, msg, errorMsg, &record.RecordId, nil, nil, nil, nil)
		}
	}
	msg5 := fmt.Sprintf("Gmail Sync completed for %d records. These records are now being processed for digitization. You’ll be notified once the process is complete.", len(emailMedRecord))
	gs.processStatusService.LogStep(processID, step4, constant.Success, msg5, errorMsg, nil, nil, nil, nil, nil)
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
		s.processStatusService.LogStepAndFail(processIdKey, step, constant.Failure, string(constant.TokenExchangeFailed), err.Error(), nil, nil)
		return err
	}
	s.processStatusService.LogStep(processIdKey, step, constant.Success, string(constant.TokenExchangeSuccess), errorMsg, nil, nil, nil, nil, nil)

	// Step: Save User Token
	_, tokenErr := s.userService.CreateTblUserToken(&models.TblUserToken{
		UserId:    userID,
		AuthToken: token.AccessToken,
		Provider:  "Gmail",
	})
	if tokenErr != nil {
		s.processStatusService.LogStepAndFail(processIdKey, step, constant.Failure, "Failed to save user token: ", tokenErr.Error(), nil, nil)
	} else {
		s.processStatusService.LogStep(processIdKey, step, constant.Success, "User token fetch successfully", "", nil, nil, nil, nil, nil)
	}

	// Step: Create Gmail client
	step = string(constant.ProcessGmailClient)
	gmailService, err := s.CreateGmailServiceClient(token.AccessToken, s.googleOauthConfig())
	if err != nil {
		s.processStatusService.LogStepAndFail(processIdKey, step, constant.Failure, string(constant.GmailClientCreateFailed), err.Error(), nil, nil)
		return err
	}
	s.processStatusService.LogStep(processIdKey, step, constant.Success, string(constant.GmailClientCreated), errorMsg, nil, nil,
		nil, nil, nil)
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

func (s *GmailSyncServiceImpl) SyncGmailApp(userID uint64, gmailService *gmail.Service) error {
	processID := uuid.New()
	msg := string(constant.ProcessStarted)
	errorMsg := ""
	s.processStatusService.LogStep(processID, string(constant.GmailSync), constant.Running, msg, errorMsg, nil, nil, nil, nil, nil)
	return s.GmailSyncCore(userID, processID, gmailService)
}
