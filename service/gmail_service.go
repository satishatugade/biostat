package service

import (
	"biostat/constant"
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"bytes"
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
	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
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
	apiService           ApiService
}

func NewGmailSyncService(processStatusService ProcessStatusService, medRecordService TblMedicalRecordService, userService UserService, diagnosticRepo repository.DiagnosticRepository, apiService ApiService) GmailSyncService {
	return &GmailSyncServiceImpl{processStatusService: processStatusService, medRecordService: medRecordService, userService: userService, diagnosticRepo: diagnosticRepo, apiService: apiService}
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
	s.processStatusService.LogStep(processID, step, constant.Running, msg1, errorMsg, nil, nil, nil, nil, nil, nil)
	results, err := service.Users.Messages.List("me").Q(filterString).Do()
	if err != nil {
		log.Println("@FetchEmailsWithAttachments->service.Users.Messages:", userId, ":", err)
		s.processStatusService.LogStepAndFail(processID, step, constant.Failure, string(constant.GmailSearchMessage), err.Error(), nil, nil, nil)
		return nil, err
	}
	s.processStatusService.LogStep(processID, step, constant.Success, msg1, errorMsg, nil, nil, nil, nil, nil, nil)
	var records []*models.TblMedicalRecord
	findMailstep := string(constant.FindingEmailWithAttachment)
	for idx, msg := range results.Messages {
		indexCount := idx
		message, err := service.Users.Messages.Get("me", msg.Id).Do()
		if err != nil || message == nil {
			log.Println("Error getting email for ", userId, userEmail, ":", err)
			logmsg := fmt.Sprintf("Error getting email for userID: %v, userEmail: %v", userId, userEmail)
			s.processStatusService.LogStepAndFail(processID, findMailstep, constant.Failure, logmsg, err.Error(), nil, nil, nil)
			continue
		}
		msg := fmt.Sprintf("Processing email attachment: %d / %d", indexCount, len(results.Messages))
		s.processStatusService.LogStep(processID, findMailstep, constant.Running, msg, errorMsg, nil, &indexCount, nil, nil, nil, nil)
		log.Println("@FetchEmailsWithAttachments->ExtractAttachments:", userId, ":", userEmail, ": Processing Mail", idx+1, "/", len(results.Messages))
		attachments := s.ExtractAttachment(service, message, userEmail, userId, processID, idx+1)
		records = append(records, attachments...)
	}
	msg3 := fmt.Sprintf("Processed email attachment: %d", len(records)-1)
	totalRecord := len(records) - 1
	s.processStatusService.LogStep(processID, findMailstep, constant.Success, msg3, errorMsg, nil, &totalRecord, nil, nil, nil, nil)
	log.Println("@FetchEmailsWithAttachments->Gmail Records found:", len(records), "userEmail: ", userEmail)
	return records, nil
}

func (s *GmailSyncServiceImpl) ExtractAttachment(service *gmail.Service, message *gmail.Message, userEmail string, userId uint64, processID uuid.UUID, mailIdx int) []*models.TblMedicalRecord {

	var records []*models.TblMedicalRecord
	totalAttempted := 0
	successCount := 0
	errorMsg := ""
	step := string(constant.DownloadAttachment)
	for idx, part := range message.Payload.Parts {
		if part.Filename != "" {
			recordIndexCount := idx + 1
			attachmentId := part.Body.AttachmentId
			body := GetMessageBody(message.Payload)
			msg := fmt.Sprintf(string(constant.DownloadAttachmentMessage)+" | Subject and Body of email %d - %d / %d: %s - %s - %+v", mailIdx, recordIndexCount, len(message.Payload.Parts), part.Filename, getHeader(message.Payload.Headers, "Subject"), body)
			log.Println("@ExtractAttachments Processing Record from Email:", mailIdx, "-", recordIndexCount, "/", len(message.Payload.Parts), ": ", part.Filename, "-", getHeader(message.Payload.Headers, "Subject"))
			s.processStatusService.LogStep(processID, step, constant.Running, msg, errorMsg, nil, &recordIndexCount, &recordIndexCount, nil, nil, &attachmentId)
			attachmentData, err := DownloadAttachment(service, message.Id, attachmentId)
			if err != nil {
				log.Printf("@ExtractAttachments->DownloadAttachment %s: %v", part.Filename, err)
				s.processStatusService.LogStep(processID, step, constant.Failure, msg, errorMsg, nil, &recordIndexCount, &recordIndexCount, nil, nil, &attachmentId)
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
			destinationPath := filepath.Join("uploads", safeFileName)

			if err := os.WriteFile(destinationPath, attachmentData, 0644); err != nil {
				log.Printf("@ExtractAttachments->Failed to save attachment locally %s: %v", part.Filename, err)
				continue
			}
			initialMetadata := map[string]interface{}{
				"attachment_id": attachmentId,
			}
			metadataJSON, _ := json.Marshal(initialMetadata)
			recordURL := fmt.Sprintf("%s/uploads/%s", os.Getenv("SHORT_URL_BASE"), safeFileName)
			subBody := fmt.Sprintf("Subject and body of email sub : %s : Body :%+v ", getHeader(message.Payload.Headers, "Subject"), body)
			newRecord := &models.TblMedicalRecord{
				RecordName:        safeFileName,
				RecordSize:        int64(len(attachmentData)),
				FileType:          part.MimeType,
				RecordUrl:         recordURL,
				UploadDestination: "LocalServer",
				Description:       subBody,
				UploadSource:      "Gmail",
				RecordCategory:    string(constant.OTHER),
				SourceAccount:     userEmail,
				Status:            constant.StatusProcessing,
				Metadata:          metadataJSON,
				UploadedBy:        userId,
				FetchedAt:         time.Now(),
			}
			successCount++
			records = append(records, newRecord)
			s.processStatusService.LogStep(processID, step, constant.Running, msg, errorMsg, nil, &recordIndexCount, &recordIndexCount, nil, nil, &attachmentId)
		}
	}
	count := len(records)
	failedCount := totalAttempted - successCount
	s.processStatusService.LogStep(processID, step, constant.Success, string(constant.DownloadAttachmentComplete), errorMsg, nil, nil, &count, &successCount, &failedCount, nil)
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

// func GetMessageBody(payload *gmail.MessagePart) string {
// 	if payload.Body != nil && payload.Body.Data != "" {
// 		// body, _ := base64.URLEncoding.DecodeString(payload.Body.Data)
// 		body, _ := base64.RawURLEncoding.DecodeString(payload.Body.Data)
// 		return string(body)
// 	}

// 	for _, part := range payload.Parts {
// 		if part.MimeType == "multipart/alternative" || part.MimeType == "multipart/mixed" {
// 			result := GetMessageBody(part)
// 			if result != "" {
// 				return result
// 			}
// 		}
// 		if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
// 			// body, _ := base64.URLEncoding.DecodeString(part.Body.Data)
// 			body, _ := base64.RawURLEncoding.DecodeString(payload.Body.Data)
// 			return string(body)
// 		}
// 		if part.MimeType == "text/html" && part.Body != nil && part.Body.Data != "" {
// 			// body, _ := base64.URLEncoding.DecodeString(part.Body.Data)
// 			body, _ := base64.RawURLEncoding.DecodeString(payload.Body.Data)
// 			return utils.StripHTML(string(body))
// 		}
// 	}

// 	return ""
// }

func GetMessageBody(payload *gmail.MessagePart) string {
	log.Println("[GetMessageBody] Starting function call")

	// Case 1: Direct body data at root level
	if payload.Body != nil && payload.Body.Data != "" {
		log.Println("[GetMessageBody] Found root-level body data")
		decoded, err := base64.RawURLEncoding.DecodeString(payload.Body.Data)
		if err != nil {
			log.Println("[GetMessageBody] Error decoding root-level body:", err)
			return ""
		}
		preview := string(decoded)
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		log.Printf("[GetMessageBody] Message body (root-level preview): %q", preview)
		return string(decoded)
	}

	log.Println("[GetMessageBody] No root-level body, checking parts... total parts:", len(payload.Parts))

	// Loop through all parts
	for i, part := range payload.Parts {
		log.Printf("[GetMessageBody] Checking part %d, MimeType: %s", i, part.MimeType)

		// Recursive check for multipart messages
		if part.MimeType == "multipart/alternative" || part.MimeType == "multipart/mixed" {
			log.Printf("[GetMessageBody] Found multipart type in part %d, diving deeper...", i)
			result := GetMessageBody(part)
			if result != "" {
				log.Printf("[GetMessageBody] Got body from nested part %d", i)
				return result
			}
			log.Printf("[GetMessageBody] No body found in nested part %d", i)
		}

		// Plain text case
		if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
			log.Printf("[GetMessageBody] Found text/plain body in part %d", i)
			decoded, err := decodeGmailBase64(part.Body.Data)
			if err != nil {
				log.Printf("[GetMessageBody] Error decoding text/plain in part %d: %v", i, err)
				continue
			}
			preview := string(decoded)
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			log.Printf("[GetMessageBody] Message body (text/plain preview): %q", preview)
			return string(decoded)
		}

		// HTML case
		if part.MimeType == "text/html" && part.Body != nil && part.Body.Data != "" {
			log.Printf("[GetMessageBody] Found text/html body in part %d", i)
			decoded, err := base64.RawURLEncoding.DecodeString(part.Body.Data)
			if err != nil {
				log.Printf("[GetMessageBody] Error decoding text/html in part %d: %v", i, err)
				continue
			}
			stripped := utils.StripHTML(string(decoded))
			preview := stripped
			if len(preview) > 100 {
				preview = preview[:100] + "..."
			}
			log.Printf("[GetMessageBody] Message body (HTML preview after strip): %q", preview)
			return stripped
		}
	}

	log.Println("[GetMessageBody] No body found in any parts")
	return ""
}

func decodeGmailBase64(data string) ([]byte, error) {
	// Gmail sometimes sends with spaces or line breaks — clean them
	cleanData := strings.ReplaceAll(data, "-", "+")
	cleanData = strings.ReplaceAll(cleanData, "_", "/")
	cleanData = strings.ReplaceAll(cleanData, "\n", "")
	cleanData = strings.ReplaceAll(cleanData, "\r", "")
	cleanData = strings.ReplaceAll(cleanData, " ", "")

	// Add padding if missing
	if m := len(cleanData) % 4; m != 0 {
		cleanData += strings.Repeat("=", 4-m)
	}

	return base64.StdEncoding.DecodeString(cleanData)
}

func (gs *GmailSyncServiceImpl) GmailSyncCore(userId uint64, processID uuid.UUID, gmailService *gmail.Service) error {
	msg := string(constant.FetchUserLab)
	step := string(constant.ProcessFetchLabs)
	errorMsg := ""
	gs.processStatusService.LogStep(processID, step, constant.Success, msg, errorMsg, nil, nil, nil, nil, nil, nil)
	labs, err := gs.diagnosticRepo.GetPatientLabNameAndEmail(userId)
	if err != nil {
		gs.processStatusService.LogStepAndFail(processID, step, constant.Failure, string(constant.UserLabNotFound), err.Error(), nil, nil, nil)
		log.Println("@GmailSyncCore->GetPatientLabNameAndEmail:", userId, " err:", err)
		return err
	} else {
		gs.processStatusService.LogStep(processID, step, constant.Success, string(constant.UserLabFetched), errorMsg, nil, nil, nil, nil, nil, nil)
	}
	filterString := utils.FormatLabsForGmailFilter(labs)
	profile, err := gmailService.Users.GetProfile("me").Do()
	if err != nil {
		step := string(constant.ProcessVerifyCredentials)
		gs.processStatusService.LogStep(processID, step, constant.Success, string(constant.InvalidCredentials), errorMsg, nil, nil, nil, nil, nil, nil)
		log.Println("@GmailSyncCore->gmailService.Users.GetProfile:", userId, " err:", err)
		return err
	}
	step1 := string(constant.ProcessFetchEmails)
	gs.processStatusService.LogStep(processID, step1, constant.Success, string(constant.FetchEmailAttachment), errorMsg, nil, nil, nil, nil, nil, nil)
	emailMedRecord, err := gs.FetchEmailsWithAttachment(gmailService, userId, filterString, processID)
	if err != nil {
		msg := "No valid medical records were found during this current Gmail sync."
		gs.processStatusService.LogStepAndFail(processID, step, constant.Failure, msg, err.Error(), nil, nil, nil)
		log.Println("@GmailSyncCore->FetchEmailsWithAttachments:", userId, " err:", err)
		return err
	}
	gs.processStatusService.LogStep(processID, step1, constant.Success, string(constant.EmailAttachmentFetch), errorMsg, nil, nil, nil, nil, nil, nil)
	totalRecord := len(emailMedRecord)
	checkDocTypeStep := string(constant.CheckDocType)
	docCompleteMsg := string(constant.CheckDocTypeCompleted)
	totalAttempted := 0
	successCount := 0
	for idx, record := range emailMedRecord {
		attachmentId, err := utils.GetAttachmentIDFromRecord(record)
		if err != nil {
			log.Println("GetAttachmentIDFromRecord Error:", err)
		} else {
			log.Println("GetAttachmentIDFromRecord Attachment ID:", attachmentId)
		}
		fileName := filepath.Base(record.RecordUrl)
		localPath := filepath.Join("uploads", fileName)
		fileData, err := os.ReadFile(localPath)
		if err != nil {
			log.Println("error while reading file while getting Doctype", record.RecordName)
			continue
		}
		checkPasswordProtectedStep := string(constant.CheckPasswordProtectedStep)
		pwsProtectedMsg := string(constant.CheckPasswordProtectedStepMsg)
		gs.processStatusService.LogStep(processID, checkPasswordProtectedStep, constant.Running, pwsProtectedMsg, errorMsg, nil, nil, nil, nil, nil, &attachmentId)
		pdfCheckResult, err := gs.apiService.CheckPDFAndGetPassword(bytes.NewReader(fileData), record.RecordName, record.Description)
		if err != nil {
			log.Printf("PDF check failed: %v", err)
			gs.processStatusService.LogStepAndFail(processID, checkPasswordProtectedStep, constant.Failure, "Failed to check if PDF is password protected", err.Error(), &idx, nil, &attachmentId)
			continue
		} else if pdfCheckResult.IsProtected {
			log.Printf("PDF is password protected. Password: %s", pdfCheckResult.Password)
			decryptMsg := fmt.Sprintf("PDF is password protected: %s, password we fetched: %s, Trying to decrypt, document URL: %s", map[bool]string{true: "Yes", false: "No"}[pdfCheckResult.IsProtected], pdfCheckResult.Password, record.RecordUrl)
			fileData, err = DecryptPDFIfProtected(fileData, pdfCheckResult.Password)
			if err != nil {
				log.Printf("Decryption failed: %v", err)
				gs.processStatusService.LogStepAndFail(processID, checkPasswordProtectedStep, constant.Failure, "PDF decryption failed", err.Error(), &idx, nil, &attachmentId)
				continue
			}
			log.Println("Decryption successful.")
			gs.processStatusService.LogStep(processID, checkPasswordProtectedStep, constant.Success, decryptMsg, errorMsg, nil, nil, nil, nil, nil, &attachmentId)
		} else {
			gs.processStatusService.LogStep(processID, checkPasswordProtectedStep, constant.Success, "Password protection check successful.", errorMsg, nil, nil, nil, nil, nil, &attachmentId)
		}
		recordIndexCount := idx + 1
		msg1 := string(constant.CheckDocTypeMessage)
		gs.processStatusService.LogStep(processID, checkDocTypeStep, constant.Running, msg1, errorMsg, nil, &recordIndexCount, &recordIndexCount, nil, nil, &attachmentId)
		log.Println("@GmailSyncCore: checking doc type for document ", idx+1, "/", totalRecord, record.RecordName)
		docTypeResp := ""
		docTypeResp, err = utils.CallDocumentTypeAPI(bytes.NewReader(fileData), record.RecordName)
		if err != nil {
			log.Printf("@GmailServiceCode->utils.CallDocumentTypeAPI type:%s %v ", record.RecordName, err)
			gs.processStatusService.LogStepAndFail(processID, checkDocTypeStep, constant.Failure, string(constant.CheckDocTypeFailedMessage), err.Error(), &idx, nil, nil)
			docTypeResp = string(constant.OTHER)
		}
		status := constant.StatusQueued
		if docTypeResp == string(constant.OTHER) {
			status = constant.StatusSuccess
		}
		record.RecordCategory = docTypeResp
		record.Status = status
		record.IsPasswordProtected = pdfCheckResult.IsProtected
		record.PDFPassword = pdfCheckResult.Password
		newmsg := fmt.Sprintf("%s : Docs type found %s : Document URL : %s", docCompleteMsg, docTypeResp, record.RecordUrl)
		gs.processStatusService.LogStep(processID, checkDocTypeStep, constant.Running, newmsg, errorMsg, nil, &recordIndexCount, &recordIndexCount, nil, nil, &attachmentId)
	}
	failedCount := totalAttempted - successCount
	gs.processStatusService.LogStep(processID, checkDocTypeStep, constant.Success, docCompleteMsg, errorMsg, nil, nil, &totalAttempted, &successCount, &failedCount, nil)
	step3 := string(constant.ProcessSaveRecords)
	msg3 := string(constant.SaveRecord)
	gs.processStatusService.LogStep(processID, step3, constant.Running, msg3, errorMsg, nil, nil, &totalRecord, nil, nil, nil)
	saveRecordErr := gs.medRecordService.SaveMedicalRecords(emailMedRecord, userId)
	if saveRecordErr != nil {
		log.Println("@GmailSyncCore->SaveMedicalRecords:", userId, " : ", saveRecordErr)
		gs.processStatusService.LogStepAndFail(processID, step3, constant.Failure, string(constant.FailedSaveRecords), saveRecordErr.Error(), nil, nil, nil)
		return saveRecordErr
	}
	gs.processStatusService.LogStep(processID, step3, constant.Success, string(constant.RecordSaveSuccess), errorMsg, nil, nil, nil, nil, nil, nil)
	step4 := string(constant.ProcessDigitization)
	msg4 := string(constant.DigitizationTaskQueue)
	gs.processStatusService.LogStep(processID, step4, constant.Running, msg4, errorMsg, nil, nil, nil, nil, nil, nil)
	log.Println("Email sync completed for user:", userId, " : ", profile.EmailAddress)

	userInfo, err := gs.userService.GetSystemUserInfoByUserID(userId)
	if err != nil {
		gs.processStatusService.LogStepAndFail(processID, step, constant.Failure, string(constant.UserProfileNotFound), err.Error(), nil, nil, nil)
		log.Println("@GmailSyncCore->GetSystemUserInfoByUserID:", userId, " : ", err)
		return err
	}
	for idx, record := range emailMedRecord {
		if record.RecordCategory == string(constant.TESTREPORT) || record.RecordCategory == string(constant.PRESCRIPTION) {
			attachmentId, err := utils.GetAttachmentIDFromRecord(record)
			if err != nil {
				log.Println("GetAttachmentIDFromRecord Error:", err)
			} else {
				log.Println("GetAttachmentIDFromRecord Attachment ID:", attachmentId)
			}
			msg := fmt.Sprintf("Starting digitization for report recordId :%d (category: test report or medication document)%d _ %d", record.RecordId, idx+1, len(emailMedRecord))
			gs.processStatusService.LogStep(processID, step4, constant.Running, msg, errorMsg, &record.RecordId, nil, nil, nil, nil, &attachmentId)
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
			taskErr := gs.medRecordService.CreateDigitizationTask(record, userInfo, userId, fileBuf, filename, processID, &attachmentId)
			if taskErr != nil {
				log.Println("Error @GmailSyncCore->CreateDigitizationTask: ", taskErr)
				gs.processStatusService.LogStepAndFail(processID, step4, constant.Failure, "Record digitization failed", taskErr.Error(), nil, &record.RecordId, &attachmentId)
			}
			gs.processStatusService.LogStep(processID, step4, constant.Success, msg, errorMsg, &record.RecordId, nil, nil, nil, nil, &attachmentId)
		}
	}
	msg5 := fmt.Sprintf("Gmail Sync completed for %d records. These records are now being processed for digitization. You’ll be notified once the process is complete.", len(emailMedRecord))
	gs.processStatusService.LogStep(processID, step4, constant.Success, msg5, errorMsg, nil, nil, nil, nil, nil, nil)
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
		s.processStatusService.LogStepAndFail(processIdKey, step, constant.Failure, string(constant.TokenExchangeFailed), err.Error(), nil, nil, nil)
		return err
	}
	s.processStatusService.LogStep(processIdKey, step, constant.Success, string(constant.TokenExchangeSuccess), errorMsg, nil, nil, nil, nil, nil, nil)

	// Step: Save User Token
	_, tokenErr := s.userService.CreateTblUserToken(&models.TblUserToken{
		UserId:    userID,
		AuthToken: token.AccessToken,
		Provider:  "Gmail",
	})
	if tokenErr != nil {
		s.processStatusService.LogStepAndFail(processIdKey, step, constant.Failure, "Failed to save user token: ", tokenErr.Error(), nil, nil, nil)
	} else {
		s.processStatusService.LogStep(processIdKey, step, constant.Success, "User token fetch successfully", "", nil, nil, nil, nil, nil, nil)
	}

	// Step: Create Gmail client
	step = string(constant.ProcessGmailClient)
	gmailService, err := s.CreateGmailServiceClient(token.AccessToken, s.googleOauthConfig())
	if err != nil {
		s.processStatusService.LogStepAndFail(processIdKey, step, constant.Failure, string(constant.GmailClientCreateFailed), err.Error(), nil, nil, nil)
		return err
	}
	s.processStatusService.LogStep(processIdKey, step, constant.Success, string(constant.GmailClientCreated), errorMsg, nil, nil,
		nil, nil, nil, nil)
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
	s.processStatusService.LogStep(processID, string(constant.GmailSync), constant.Running, msg, errorMsg, nil, nil, nil, nil, nil, nil)
	return s.GmailSyncCore(userID, processID, gmailService)
}
func DecryptPDFIfProtected(fileData []byte, password string) ([]byte, error) {
	buf := &bytes.Buffer{}

	conf := model.NewDefaultConfiguration()
	conf.UserPW = password // set password here
	// conf := model.NewAESConfiguration(password, "", 256)

	err := pdfcpuapi.Decrypt(bytes.NewReader(fileData), buf, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt PDF: %w", err)
	}
	return buf.Bytes(), nil
}
