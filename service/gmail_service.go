package service

import (
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
	FetchEmailsWithAttachments(service *gmail.Service, userId uint64, filterString string) ([]*models.TblMedicalRecord, error)
	GmailServiceApp(userID uint64, accessToken string) (*gmail.Service, error)
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

func (s *GmailSyncServiceImpl) FetchEmailsWithAttachments(service *gmail.Service, userId uint64, filterString string) ([]*models.TblMedicalRecord, error) {
	profile, err := service.Users.GetProfile("me").Do()
	if err != nil {
		return nil, err
	}
	userEmail := profile.EmailAddress
	query := fmt.Sprintf("(%s) has:attachment", filterString)
	log.Println(query)
	results, err := service.Users.Messages.List("me").Q(query).Do()
	if err != nil {
		return nil, err
	}

	var records []*models.TblMedicalRecord

	for _, msg := range results.Messages {
		message, err := service.Users.Messages.Get("me", msg.Id).Do()
		if err != nil {
			continue
		}

		attachments := ExtractAttachments(service, message, userEmail, userId)
		for i, record := range attachments {
			log.Printf("Attachment [%d]: %+v", i, record)
		}
		records = append(records, attachments...)
	}
	log.Println("Gmail Records found:", len(records), "userEmail: ", userEmail)
	return records, nil
}

func ExtractAttachments(service *gmail.Service, message *gmail.Message, userEmail string, userId uint64) []*models.TblMedicalRecord {
	var records []*models.TblMedicalRecord
	for _, part := range message.Payload.Parts {
		if part.Filename != "" {
			attachmentData, err := DownloadAttachment(service, message.Id, part.Body.AttachmentId)
			if err != nil {
				log.Println("Failed to download attachment %s: %v", part.Filename, err)
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
				log.Printf("Failed to detect document type:%s %v ", part.Filename, err)
				continue
			} else {
				log.Printf("Doc Type Response:%s %v", part.Filename, docTypeResp)
			}

			destinationPath := filepath.Join("uploads", safeFileName)

			if err := os.WriteFile(destinationPath, attachmentData, 0644); err != nil {
				log.Printf("Failed to save attachment locally %s: %v", part.Filename, err)
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
				Status:            "success",
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

func (s *GmailSyncServiceImpl) SyncGmailWeb(userID uint64, code string) error {
	processID := s.processStatusService.StartProcess(userID, "gmail_sync", strconv.FormatUint(userID, 10), "tbl_medical_record", "token_exchange")
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
		s.processStatusService.UpdateProcess(processID, "failed", nil, &msg, &step, true)
		return err
	}

	s.userService.CreateTblUserToken(&models.TblUserToken{UserId: userID, AuthToken: token.AccessToken, Provider: "Gmail"})
	labs, err := s.diagnosticRepo.GetPatientLabNameAndEmail(userID)
	if err != nil {
		msg = "query generation failed"
		s.processStatusService.UpdateProcess(processID, "failed", nil, &msg, &step, true)
		return err
	}
	filterString := utils.FormatLabsForGmailFilter(labs)
	gmailService, err := s.CreateGmailServiceClient(token.AccessToken, googleOauthConfig)
	if err != nil {
		msg = "Failed to create Gmail client"
		step = "gmail_client"
		s.processStatusService.UpdateProcess(processID, "failed", nil, &msg, &step, true)
		return err
	}
	profile, err := gmailService.Users.GetProfile("me").Do()
	if err != nil {
		return err
	}

	s.processStatusService.UpdateProcess(processID, "running", &profile.EmailAddress, nil, nil, false)

	emailMedRecord, err := s.FetchEmailsWithAttachments(gmailService, userID, filterString)
	if err != nil {
		msg = "Failed to fetch emails"
		step = "fetch_emails"
		s.processStatusService.UpdateProcess(processID, "failed", nil, &msg, nil, true)
		log.Println("Failed to fetch emails:", err)
		return err
	}

	msg = "Email records fetched"
	step = "save_records"
	s.processStatusService.UpdateProcess(processID, "running", nil, &msg, &step, false)

	err = s.medRecordService.SaveMedicalRecords(emailMedRecord, userID)
	if err != nil {
		msg = "Failed to save medical records"
		log.Println("Error while saving email data:", err)
		s.processStatusService.UpdateProcess(processID, "failed", nil, &msg, nil, true)
		return err
	}
	msg = "Sync completed"

	s.processStatusService.UpdateProcess(processID, "completed", nil, &msg, nil, true)
	log.Println("Email sync completed for user:", userID)
	userInfo, err := s.userService.GetSystemUserInfoByUserID(userID)
	if err != nil {
		return err
	}
	for _, record := range emailMedRecord {
		log.Println("Starting to Digitize saved record : %s", record.RecordId)
		resp, err := http.Get(record.RecordUrl)
		if err != nil {
			log.Println("Error @SyncGmailWeb->Read File From URL:%v", err)
			continue
		}
		defer resp.Body.Close()

		fileBuf := new(bytes.Buffer)
		_, err = io.Copy(fileBuf, resp.Body)
		filename := filepath.Base(record.RecordUrl)
		err = s.medRecordService.CreateDigitizationTask(record, userInfo, userID, userInfo.AuthUserId, fileBuf, filename)
		if err != nil {
			log.Println("Error @SyncGmailWeb->CreateDigitizationTask:%v", err)
		}
	}
	return nil
}

func (s *GmailSyncServiceImpl) GmailServiceApp(userID uint64, accessToken string) (*gmail.Service, error) {
	context := context.Background()
	gmailService, err := s.CreateGmailServiceFromToken(context, accessToken)
	if err != nil {
		return nil, err
	}
	_, err = s.diagnosticRepo.GetPatientLabNameAndEmail(userID)
	if err != nil {
		return nil, err
	}

	return gmailService, nil
}

func (s *GmailSyncServiceImpl) SyncGmailApp(userID uint64, gmailService *gmail.Service) error {
	processID := s.processStatusService.StartProcess(userID, "gmail_sync", strconv.FormatUint(userID, 10), "tbl_medical_record", "token_exchange")
	msg := ""
	step := ""
	labs, err := s.diagnosticRepo.GetPatientLabNameAndEmail(userID)
	if err != nil {
		msg = "query generation failed"
		step = "fetch_labs"
		s.processStatusService.UpdateProcess(processID, "failed", nil, &msg, &step, true)
		return err
	}
	filterString := utils.FormatLabsForGmailFilter(labs)
	profile, err := gmailService.Users.GetProfile("me").Do()
	if err != nil {

		return err
	}
	s.processStatusService.UpdateProcess(processID, "running", &profile.EmailAddress, nil, nil, false)

	emailMedRecord, err := s.FetchEmailsWithAttachments(gmailService, userID, filterString)
	if err != nil {
		msg = "Failed to fetch emails"
		step = "fetch_emails"
		s.processStatusService.UpdateProcess(processID, "failed", nil, &msg, nil, true)
		log.Println("Failed to fetch emails:", err)
		return err
	}

	msg = "Email records fetched"
	step = "save_records"
	s.processStatusService.UpdateProcess(processID, "running", nil, &msg, &step, false)

	err = s.medRecordService.SaveMedicalRecords(emailMedRecord, userID)
	if err != nil {
		msg = "Failed to save medical records"
		log.Println("Error while saving email data:", err)
		s.processStatusService.UpdateProcess(processID, "failed", nil, &msg, nil, true)
		return err
	}
	msg = "Sync completed"
	s.processStatusService.UpdateProcess(processID, "completed", nil, &msg, nil, true)
	log.Println("Email sync completed for user:", userID)
	return nil
}
