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
	FetchEmailsWithAttachments(service *gmail.Service, userId uint64, filterString, key string) ([]*models.TblMedicalRecord, error)
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

func (s *GmailSyncServiceImpl) FetchEmailsWithAttachments(service *gmail.Service, userId uint64, filterString, key string) ([]*models.TblMedicalRecord, error) {
	log.Println("@FetchEmailsWithAttachments:", userId)
	profile, err := service.Users.GetProfile("me").Do()
	if err != nil {
		log.Println("@FetchEmailsWithAttachments->service.Users.GetProfile:", userId, ":", err)
		return nil, err
	}
	step := "gmail_search"
	msg := "Searching for health records"
	userEmail := profile.EmailAddress
	query := fmt.Sprintf("(%s) has:attachment", filterString)
	log.Println("Inbox Search Query:", userId, ":", query)
	s.processStatusService.UpdateProcessRedis(key, constant.Running, nil, msg, step, false)
	results, err := service.Users.Messages.List("me").Q(query).Do()
	if err != nil {
		log.Println("@FetchEmailsWithAttachments->service.Users.Messages:", userId, ":", err)
		return nil, err
	}

	var records []*models.TblMedicalRecord

	for idx, msg := range results.Messages {
		message, err := service.Users.Messages.Get("me", msg.Id).Do()
		if err != nil {
			continue
		}
		s.processStatusService.UpdateProcessRedis(key, constant.Running, nil, fmt.Sprintf("Processing Mail: %d / %d", idx+1, len(results.Messages)), step, false)
		log.Println("@FetchEmailsWithAttachments->ExtractAttachments:", userId, ":", userEmail, ": Processing Mail", idx+1, "/", len(results.Messages))
		attachments := ExtractAttachments(service, message, userEmail, userId)
		records = append(records, attachments...)
	}
	log.Println("@FetchEmailsWithAttachments->Gmail Records found:", len(records), "userEmail: ", userEmail)
	return records, nil
}

func ExtractAttachments(service *gmail.Service, message *gmail.Message, userEmail string, userId uint64) []*models.TblMedicalRecord {
	var records []*models.TblMedicalRecord
	for idx, part := range message.Payload.Parts {
		log.Println("@ExtractAttachments Processing Record from Email:", idx+1, "/", len(message.Payload.Parts), ": ", part.Filename)
		if part.Filename != "" {
			attachmentData, err := DownloadAttachment(service, message.Id, part.Body.AttachmentId)
			if err != nil {
				log.Println("@ExtractAttachments->DownloadAttachment %s: %v", part.Filename, err)
				continue
			}
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
			records = append(records, newRecord)
		}
	}
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

func (gs *GmailSyncServiceImpl) GmailSyncCore(userId uint64, processID uuid.UUID, key string, gmailService *gmail.Service) error {
	msg := "getting user labs"
	step := "fetch_labs"
	// gs.processStatusService.UpdateProcess(processID, "running", nil, &msg, &step, false)
	gs.processStatusService.UpdateProcessRedis(key, constant.Running, nil, msg, step, false)
	labs, err := gs.diagnosticRepo.GetPatientLabNameAndEmail(userId)
	if err != nil {
		msg = "query generation failed"
		gs.processStatusService.UpdateProcess(processID, constant.Failure, nil, &msg, &step, true)
		gs.processStatusService.UpdateProcessRedis(key, constant.Failure, nil, msg, step, true)
		log.Println("@GmailSyncCore->GetPatientLabNameAndEmail:", userId, " err:", err)
		return err
	}

	filterString := utils.FormatLabsForGmailFilter(labs)
	profile, err := gmailService.Users.GetProfile("me").Do()
	if err != nil {
		msg = "Invalid credentials"
		step = "verify_credentials"
		gs.processStatusService.UpdateProcess(processID, constant.Failure, nil, &msg, &step, true)
		gs.processStatusService.UpdateProcessRedis(key, constant.Failure, nil, msg, step, true)
		log.Println("@GmailSyncCore->gmailService.Users.GetProfile:", userId, " err:", err)
		return err
	}
	step = "fetch_emails"
	msg = "fetching email attachments"
	// gs.processStatusService.UpdateProcess(processID, constant.Running, &profile.EmailAddress, &msg, &step, false)
	gs.processStatusService.UpdateProcessRedis(key, constant.Running, &profile.EmailAddress, msg, step, false)

	emailMedRecord, err := gs.FetchEmailsWithAttachments(gmailService, userId, filterString, key)
	if err != nil {
		msg = "Failed to fetch emails"
		gs.processStatusService.UpdateProcess(processID, constant.Failure, nil, &msg, nil, true)
		gs.processStatusService.UpdateProcessRedis(key, constant.Failure, nil, msg, step, true)
		log.Println("@GmailSyncCore->FetchEmailsWithAttachments:", userId, " err:", err)
		return err
	}
	step = "save_records"
	msg = "Saving medical records to the database"
	// gs.processStatusService.UpdateProcess(processID, constant.Running, nil, &msg, &step, false)
	gs.processStatusService.UpdateProcessRedis(key, constant.Running, nil, msg, step, false)

	err = gs.medRecordService.SaveMedicalRecords(emailMedRecord, userId)
	if err != nil {
		msg = "Failed to save medical records"
		log.Println("@GmailSyncCore->SaveMedicalRecords:", userId, " : ", err)
		gs.processStatusService.UpdateProcess(processID, constant.Failure, nil, &msg, nil, true)
		gs.processStatusService.UpdateProcessRedis(key, constant.Failure, nil, msg, step, true)
		return err
	}
	step = "digitization"
	msg = "Creating digitization tasks for saved records"
	// gs.processStatusService.UpdateProcess(processID, constant.Running, nil, &msg, &step, false)
	gs.processStatusService.UpdateProcessRedis(key, constant.Running, nil, msg, step, false)
	log.Println("Email sync completed for user:", userId, " : ", profile.EmailAddress)

	userInfo, err := gs.userService.GetSystemUserInfoByUserID(userId)
	if err != nil {
		msg = "failed to load profile for digitization"
		gs.processStatusService.UpdateProcess(processID, constant.Failure, nil, &msg, nil, true)
		gs.processStatusService.UpdateProcessRedis(key, constant.Failure, nil, msg, step, true)
		log.Println("@GmailSyncCore->GetSystemUserInfoByUserID:", userId, " : ", err)
		return err
	}

	for idx, record := range emailMedRecord {
		gs.processStatusService.UpdateProcessRedis(key, constant.Running, nil, fmt.Sprintf("Starting to Digitize saved record: %d / %d", idx+1, len(emailMedRecord)), step, false)
		log.Println("Starting to Digitize saved record :", record.RecordId)
		resp, err := http.Get(record.RecordUrl)
		if err != nil {
			log.Println("Error @GmailSyncCore->Read File From URL:%v", err)
			continue
		}
		defer resp.Body.Close()

		fileBuf := new(bytes.Buffer)
		_, err = io.Copy(fileBuf, resp.Body)
		filename := filepath.Base(record.RecordUrl)
		err = gs.medRecordService.CreateDigitizationTask(record, userInfo, userId, userInfo.AuthUserId, fileBuf, filename)
		if err != nil {
			log.Println("Error @GmailSyncCore->CreateDigitizationTask: ", err)
		}
	}
	msg = fmt.Sprintf("Sync completed for %d records,they're in digitization process", len(emailMedRecord))
	gs.processStatusService.UpdateProcess(processID, constant.Success, nil, &msg, nil, true)
	gs.processStatusService.UpdateProcessRedis(key, constant.Success, nil, msg, step, true)
	return nil
}

func (s *GmailSyncServiceImpl) SyncGmailWeb(userID uint64, code string) error {
	processID := uuid.New()
	_ = s.processStatusService.StartProcess(processID, userID, "gmail_sync", strconv.FormatUint(userID, 10), "tbl_medical_record", "token_exchange")
	key, err := s.processStatusService.StartProcessRedis(processID, userID, "gmail_sync", strconv.FormatUint(userID, 10), "tbl_medical_record", "token_exchange")
	if err != nil {
		log.Println("@SyncGmailWeb->StartProcessRedis:", err)
		return err
	}
	msg := ""
	step := "token_exchange"
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
		msg = "Token exchange failed"
		s.processStatusService.UpdateProcess(processID, constant.Failure, nil, &msg, &step, true)
		s.processStatusService.UpdateProcessRedis(key, constant.Failure, nil, msg, step, true)
		return err
	}

	s.userService.CreateTblUserToken(&models.TblUserToken{UserId: userID, AuthToken: token.AccessToken, Provider: "Gmail"})

	gmailService, err := s.CreateGmailServiceClient(token.AccessToken, googleOauthConfig)
	if err != nil {
		msg = "Failed to create Gmail client"
		step = "gmail_client"
		s.processStatusService.UpdateProcess(processID, constant.Failure, nil, &msg, &step, true)
		s.processStatusService.UpdateProcessRedis(key, constant.Failure, nil, msg, step, true)
		return err
	}
	return s.GmailSyncCore(userID, processID, key, gmailService)
}

func (s *GmailSyncServiceImpl) SyncGmailApp(userID uint64, gmailService *gmail.Service) error {
	processID := uuid.New()
	_ = s.processStatusService.StartProcess(processID, userID, "gmail_sync", strconv.FormatUint(userID, 10), "tbl_medical_record", "token_exchange")
	key, err := s.processStatusService.StartProcessRedis(processID, userID, "gmail_sync", strconv.FormatUint(userID, 10), "tbl_medical_record", "token_exchange")
	if err != nil {
		log.Println("@SyncGmailApp->StartProcessRedis:", err)
		return err
	}
	return s.GmailSyncCore(userID, processID, key, gmailService)
}
