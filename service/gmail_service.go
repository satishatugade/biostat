package service

import (
	"biostat/config"
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
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GmailSyncService interface {
	GetGmailAuthURL(userId uint64) (string, error)
	CreateGmailServiceClient(accessToken string, googleOauthConfig *oauth2.Config) (*gmail.Service, error)
	CreateGmailServiceFromToken(ctx context.Context, accessToken string) (*gmail.Service, error)
	FetchEmailsWithAttachment(service *gmail.Service, userId uint64, filterString string, processID uuid.UUID, labNames []string) ([]*models.TblMedicalRecord, error)
	ExtractAttachment(service *gmail.Service, message *gmail.Message, bodyText string, subject string, emailDate string, userEmail string, userId uint64, processID uuid.UUID, index int) []*models.TblMedicalRecord
	CreateGmailServiceForApp(userID uint64, accessToken string) (*gmail.Service, error)
	SyncGmailWeb(userID uint64, code string) error
	SyncGmailApp(userID uint64, service *gmail.Service) error
	GetPatientNameOnDoc(patientNameOnReport string, userID uint64) (*models.PatientDocResponse, error)
}

type GmailSyncServiceImpl struct {
	processStatusService ProcessStatusService
	medRecordService     TblMedicalRecordService
	userService          UserService
	diagnosticRepo       repository.DiagnosticRepository
	apiService           ApiService
	patientService       PatientService
	recordRepo           repository.TblMedicalRecordRepository
	db                   *gorm.DB
}

func NewGmailSyncService(processStatusService ProcessStatusService, medRecordService TblMedicalRecordService, userService UserService, diagnosticRepo repository.DiagnosticRepository, apiService ApiService, patientService PatientService, recordRepo repository.TblMedicalRecordRepository, db *gorm.DB) GmailSyncService {
	return &GmailSyncServiceImpl{processStatusService: processStatusService, medRecordService: medRecordService, userService: userService, diagnosticRepo: diagnosticRepo, apiService: apiService, patientService: patientService, recordRepo: recordRepo, db: db}
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

func (s *GmailSyncServiceImpl) FetchEmailsWithAttachment(service *gmail.Service, userId uint64, filterString string, processID uuid.UUID, labNames []string) ([]*models.TblMedicalRecord, error) {
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
	log.Println("Inbox Search Query:", userId, ":", userEmail, ":", filterString)
	s.processStatusService.LogStep(processID, step, constant.Running, msg1, errorMsg, nil, nil, nil, nil, nil, nil)
	results, err := service.Users.Messages.List("me").Q(filterString).MaxResults(500).Do()
	if err != nil {
		log.Println("Error fetching user gmails for filter query:", userId, ":", err)
		s.processStatusService.LogStepAndFail(processID, step, constant.Failure, string(constant.GmailSearchMessage), err.Error(), nil, nil, nil)
		return nil, err
	}
	log.Println("Got email from initial search:", len(results.Messages))
	var allMessages []*gmail.Message
	allMessages = append(allMessages, results.Messages...)
	for results.NextPageToken != "" {
		results, err = service.Users.Messages.List("me").Q(filterString).MaxResults(500).PageToken(results.NextPageToken).Do()
		if err != nil {
			log.Println("Error fetching paginated user gmails:", userId, ":", err)
			break
		}
		log.Println("Got email from pagination search:", len(results.Messages))
		allMessages = append(allMessages, results.Messages...)
	}

	s.processStatusService.LogStep(processID, step, constant.Success, msg1, errorMsg, nil, nil, nil, nil, nil, nil)
	var emailSummaries []string
	var records []*models.TblMedicalRecord
	findMailstep := string(constant.FindingEmailWithAttachment)

	for idx, msg := range allMessages {
		indexCount := idx
		msgId := msg.Id
		message, err := service.Users.Messages.Get("me", msgId).Format("full").Do()
		if err != nil || message == nil {
			log.Println("FetchEmailsWithAttachments Error getting email for ", userId, userEmail, ":", err)
			logmsg := fmt.Sprintf("Error getting email for userID: %v, userEmail: %v", userId, userEmail)
			s.processStatusService.LogStepAndFail(processID, findMailstep, constant.Failure, logmsg, err.Error(), nil, nil, nil)
			continue
		}
		var bodyText string
		subject := getHeader(message.Payload.Headers, "Subject")
		emailDate := getHeader(message.Payload.Headers, "Date")
		from := getHeader(message.Payload.Headers, "From")
		emailSummaries = append(emailSummaries, fmt.Sprintf("%d : from %s: %s", idx+1, from, subject))
		log.Println(idx+1, ": Checking ", emailDate, "||", subject)

		bodyText = GetMessageBody(message)
		normalizeBodyText := utils.NormalizeText(bodyText)
		foundLab := false
		emailMsg := ""
		for _, lab := range labNames {
			normalizeLabName := utils.NormalizeText(lab)
			matched, detail := utils.MatchLabInBody(normalizeBodyText, normalizeLabName)
			if matched {
				summary := detail.Summary(lab)
				emailMsg = fmt.Sprintf("Lab name %s found in EmailSub %s dated %s, summary: %s", lab, subject, emailDate, summary)
				foundLab = true
				break
			}
		}
		if !foundLab {
			msg := fmt.Sprintf("No valid lab found in email body for EmailSub %s dated %s", subject, emailDate)
			log.Println(msg, "\n message body:", bodyText)
			s.processStatusService.LogStep(processID, findMailstep, constant.Running, msg, errorMsg, nil, &indexCount, nil, nil, nil, &msgId)
			continue
		}
		log.Println("FetchEmailsWithAttachments Processing Email for:", userId, ":", userEmail, ": Processing Mail", idx+1, " of ", len(results.Messages))
		attachments := s.ExtractAttachment(service, message, bodyText, subject, emailDate, userEmail, userId, processID, idx+1)
		msg := fmt.Sprintf("%s and %d attachments found", emailMsg, len(attachments))
		s.processStatusService.LogStep(processID, findMailstep, constant.Running, msg, errorMsg, nil, &indexCount, nil, nil, nil, &msgId)
		records = append(records, attachments...)
	}

	logMsg := strings.Join(emailSummaries, " | ")
	log.Println(logMsg)
	step1 := string(constant.FetchEmailsList)
	s.processStatusService.LogStep(processID, step1, constant.Success, fmt.Sprintf("%d emails found: %s", len(emailSummaries), logMsg), errorMsg, nil, nil, nil, nil, nil, nil)

	msg3 := fmt.Sprintf("Found %d email attachment in %d email", len(records), len(allMessages))
	log.Println(msg3)
	totalRecord := len(records)
	s.processStatusService.LogStep(processID, findMailstep, constant.Success, msg3, errorMsg, nil, &totalRecord, nil, nil, nil, nil)
	log.Println("@FetchEmailsWithAttachments->Gmail Records found:", len(records), "userEmail: ", userEmail)
	return records, nil
}

func (s *GmailSyncServiceImpl) ExtractAttachment(service *gmail.Service, message *gmail.Message, bodyText string, subject string, emailDate string, userEmail string, userId uint64, processID uuid.UUID, mailIdx int) []*models.TblMedicalRecord {
	var records []*models.TblMedicalRecord
	totalAttempted := 0
	successCount := 0
	errorMsg := ""
	step := string(constant.DownloadAttachment)
	for idx, part := range message.Payload.Parts {
		if part.Filename != "" {
			recordIndexCount := idx + 1
			attachmentId := part.Body.AttachmentId
			msg := fmt.Sprintf("Downloading attachment %s Dated on %s from EmailSub %s", part.Filename, emailDate, subject)
			log.Println("@ExtractAttachments Processing Record from Email:", mailIdx, "-", recordIndexCount, "/", len(message.Payload.Parts), ": ", part.Filename, "-", subject)
			s.processStatusService.LogStep(processID, step, constant.Running, msg, errorMsg, nil, &recordIndexCount, &recordIndexCount, nil, nil, &attachmentId)
			attachmentData, err := DownloadAttachment(service, message.Id, attachmentId)
			if err != nil {
				log.Printf("@ExtractAttachments->DownloadAttachment %s: %v", part.Filename, err)
				msg = fmt.Sprintf("Error Downloading attachment %s from %s Dated on %s", part.Filename, subject, emailDate)
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
			subBody := fmt.Sprintf("Subject and body of email sub : %s : Body :%+v ", subject, bodyText)
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
				UDF1:              subject,
				UDF2:              emailDate,
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

func decodeBase64Safe(data string) string {
	if data == "" {
		return ""
	}

	if decoded, err := base64.RawURLEncoding.DecodeString(data); err == nil {
		// log.Println("Retuning base64.RawURLEncoding.DecodeString")
		return string(decoded)
	} else {
		log.Printf("[decodeBase64Safe] RawURLEncoding failed: %v", err)
	}
	if decoded, err := base64.StdEncoding.DecodeString(data); err == nil {
		// log.Println("Retuning base64.StdEncoding.DecodeString")
		return string(decoded)
	} else {
		log.Printf("[decodeBase64Safe] StdEncoding failed: %v", err)
	}

	padded := data
	if m := len(padded) % 4; m != 0 {
		padded += strings.Repeat("=", 4-m)
	}
	if decoded, err := base64.URLEncoding.DecodeString(padded); err == nil {
		// log.Println("Retuning base64.URLEncoding.DecodeString")
		return string(decoded)
	} else {
		log.Printf("[decodeBase64Safe] URLEncoding with padding failed: %v", err)
	}

	log.Printf("[decodeBase64Safe] All decoding attempts failed, returning raw string")
	return data
}

func extractBodyFromParts(parts []*gmail.MessagePart) string {
	for _, part := range parts {
		if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
			// log.Println("extractBodyFromParts Detected Text/plain")
			return decodeBase64Safe(part.Body.Data)
		}
		if part.MimeType == "text/html" && part.Body != nil && part.Body.Data != "" {
			// log.Println("extractBodyFromParts Detected Text/html")
			return utils.StripHTML(decodeBase64Safe(part.Body.Data))
		}
		if len(part.Parts) > 0 {
			if result := extractBodyFromParts(part.Parts); result != "" {
				// log.Println("extractBodyFromParts this part has its own sub-parts")
				return result
			}
		}
	}
	// log.Println("Returning empty extractBodyFromParts")
	return ""
}

func GetMessageBody(msg *gmail.Message) string {
	if msg == nil || msg.Payload == nil {
		log.Println("Message is nil returning Empty string")
		return ""
	}
	if len(msg.Payload.Parts) == 0 && msg.Payload.Body != nil && msg.Payload.Body.Data != "" {
		body := decodeBase64Safe(msg.Payload.Body.Data)
		if strings.Contains(strings.ToLower(msg.Payload.MimeType), "html") {
			// log.Println("Extracting & Returning HTML")
			return utils.StripHTML(body)
		}
		log.Println("Returning decodeBase64Safe HTML")
		return body
	}

	// Multipart email
	return extractBodyFromParts(msg.Payload.Parts)
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
	var labNames []string
	for _, lab := range labs {
		labNames = append(labNames, lab.LabName)
	}
	if err != nil {
		gs.processStatusService.LogStepAndFail(processID, step, constant.Failure, string(constant.UserLabNotFound), err.Error(), nil, nil, nil)
		log.Println("@GmailSyncCore->GetPatientLabNameAndEmail:", userId, " err:", err)
		return err
	} else {
		labMsg := fmt.Sprintf("Total %d labs fetched %s ", len(labNames), strings.Join(labNames, " | "))
		gs.processStatusService.LogStep(processID, step, constant.Success, labMsg, errorMsg, nil, nil, nil, nil, nil, nil)
	}
	filterString := utils.FormatLabsForGmailFilter(labs)
	profile, err := gmailService.Users.GetProfile("me").Do()
	if err != nil {
		step := string(constant.ProcessVerifyCredentials)
		gs.processStatusService.LogStep(processID, step, constant.Success, string(constant.InvalidCredentials), errorMsg, nil, nil, nil, nil, nil, nil)
		log.Println("@GmailSyncCore->gmailService.Users.GetProfile:", userId, " err:", err)
		return err
	}
	emailMedRecord, err := gs.FetchEmailsWithAttachment(gmailService, userId, filterString, processID, labNames)
	if err != nil {
		msg := "No valid medical records were found during this current Gmail sync."
		gs.processStatusService.LogStepAndFail(processID, step, constant.Failure, msg, err.Error(), nil, nil, nil)
		log.Println("@GmailSyncCore->FetchEmailsWithAttachments:", userId, " err:", err)
		return err
	}
	totalRecord := len(emailMedRecord)
	checkDocTypeStep := string(constant.CheckDocType)
	docCompleteMsg := string(constant.CheckDocTypeCompleted)
	totalAttempted := 0
	errorCount := 0
	for idx, record := range emailMedRecord {
		recordInfo := fmt.Sprintf("%s:- %s", record.UDF2, record.UDF1)
		attachmentId, err := utils.GetAttachmentIDFromRecord(record)
		if err != nil {
			log.Println("GetAttachmentIDFromRecord Error:", err)
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
			log.Printf("PDF check failed: for %s %v", record.RecordUrl, err)
			msg := fmt.Sprintf("Processing doc %d | %s | Failed to check if PDF is password protected. Doc URL : %s", idx+1, recordInfo, record.RecordUrl)
			gs.processStatusService.LogStepAndFail(processID, checkPasswordProtectedStep, constant.Failure, msg, err.Error(), &idx, nil, &attachmentId)
			continue
		} else if pdfCheckResult.IsProtected {
			decryptMsg := fmt.Sprintf("Processing doc %d | %s | Doc password protected: %s, password we fetched: %s, Trying to decrypt, document URL: %s", idx+1, recordInfo, map[bool]string{true: "Yes", false: "No"}[pdfCheckResult.IsProtected], pdfCheckResult.Password, record.RecordUrl)
			fileData, err = DecryptPDFIfProtected(fileData, pdfCheckResult.Password)
			if err != nil {
				log.Printf("Decryption failed: %v", err)
				decryptErr := fmt.Sprintf("Processing doc %d | %s | Doc decryption failed, password we fetched: %s, document URL: %s", idx+1, recordInfo, pdfCheckResult.Password, record.RecordUrl)
				gs.processStatusService.LogStepAndFail(processID, checkPasswordProtectedStep, constant.Failure, decryptErr, err.Error(), &idx, nil, &attachmentId)
				continue
			}
			log.Println("Decryption successful.")
			gs.processStatusService.LogStep(processID, checkPasswordProtectedStep, constant.Success, decryptMsg, errorMsg, nil, nil, nil, nil, nil, &attachmentId)
		} else {
			msg := fmt.Sprintf("Processing doc %d | %s | Document is not password protected. Doc URL : %s", idx+1, recordInfo, record.RecordUrl)
			gs.processStatusService.LogStep(processID, checkPasswordProtectedStep, constant.Success, msg, errorMsg, nil, nil, nil, nil, nil, &attachmentId)
		}
		recordIndexCount := idx + 1
		msg1 := fmt.Sprintf("Processing doc %d | %s | %s", idx+1, string(constant.CheckDocTypeMessage), recordInfo)
		gs.processStatusService.LogStep(processID, checkDocTypeStep, constant.Running, msg1, errorMsg, nil, &recordIndexCount, &recordIndexCount, nil, nil, &attachmentId)
		log.Println("@GmailSyncCore: checking doc type for document ", idx+1, "/", totalRecord, record.RecordName)
		docType := ""
		keywordDocType := ""
		var docTypeLogs string
		var keywordDocTypeLogs string
		var patientNameOnReport string
		apiResponse, err := gs.medRecordService.EnqueueDocTypeCheckTask(attachmentId, record.RecordName, fileData, processID)
		// apiResponse, err := gs.apiService.CallDocumentTypeAPI(bytes.NewReader(fileData), record.RecordName)
		if err != nil {
			errorCount++
			log.Printf("@GmailServiceCode->utils.CallDocumentTypeAPI type:%s %v ", record.RecordName, err)
			docMsg := fmt.Sprintf("Processing doc %d | %s | %s", idx+1, string(constant.CheckDocTypeFailedMessage), recordInfo)
			gs.processStatusService.LogStepAndFail(processID, checkDocTypeStep, constant.Failure, docMsg, err.Error(), &idx, nil, nil)
			docType = string(constant.OTHER)
		}
		if apiResponse.Content != nil {
			patientNameOnReport = apiResponse.Content.PatientName
			if apiResponse.Content.LLMClassifier.DocumentType != "" {
				docType = apiResponse.Content.LLMClassifier.DocumentType
				docTypeLogs = apiResponse.Content.LLMClassifier.Logs
			}
			if apiResponse.Content.RegexClassifier.DocumentType != "" {
				keywordDocType = apiResponse.Content.RegexClassifier.DocumentType
				keywordDocTypeLogs = apiResponse.Content.RegexClassifier.Logs
			}
		}
		status := constant.StatusQueued
		if record.RecordCategory == string(constant.INSURANCE) || record.RecordCategory == string(constant.INVOICE) || record.RecordCategory == string(constant.DISCHARGESUMMARY) || record.RecordCategory == string(constant.VACCINATION) || record.RecordCategory == string(constant.NONMEDICAL) || record.RecordCategory == string(constant.SCANS) || record.RecordCategory == string(constant.OTHER) {
			status = constant.Success
		}
		jsonData, err := json.Marshal(apiResponse)
		if err != nil {
			return fmt.Errorf("failed to marshal apiResponse: %w", err)
		}
		record.DocTypeResponseMetaData = datatypes.JSON(jsonData)
		record.RecordCategory = keywordDocType
		record.Status = status
		record.IsPasswordProtected = pdfCheckResult.IsProtected
		record.PDFPassword = pdfCheckResult.Password
		newmsg := fmt.Sprintf("Processing doc %d Name of patient on document →:%s | Document classification : Regex → Keyword classify as →: %s | (LLM classify as →: %s ) : Document URL : %s | %s : LLMClassifier Reason :%s  : RegexClassifier Reason : %s ", idx+1, patientNameOnReport, utils.SafeString(&keywordDocType), utils.SafeString(&docType), utils.SafeString(&record.RecordUrl), utils.SafeString(&recordInfo), utils.SafeString(&docTypeLogs), utils.SafeString(&keywordDocTypeLogs))
		gs.processStatusService.LogStep(processID, checkDocTypeStep, constant.Running, newmsg, errorMsg, nil, &recordIndexCount, &recordIndexCount, nil, nil, &attachmentId)
	}
	successCount := totalAttempted - errorCount
	gs.processStatusService.LogStep(processID, checkDocTypeStep, constant.Success, docCompleteMsg, errorMsg, nil, nil, &totalAttempted, &successCount, &errorCount, nil)
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
		recordInfo := fmt.Sprintf("%s:- %s", record.UDF2, record.UDF1)
		if record.RecordCategory == string(constant.TESTREPORT) || record.RecordCategory == string(constant.MEDICATION) {
			attachmentId, err := utils.GetAttachmentIDFromRecord(record)
			if err != nil {
				log.Println("GetAttachmentIDFromRecord Error:", err)
			} else {
				log.Println("GetAttachmentIDFromRecord Attachment ID:", attachmentId)
			}
			msg := fmt.Sprintf("Processing doc %d | Starting digitization for report id :%d category: %s %s | %s", idx+1, record.RecordId, record.RecordCategory, record.RecordName, recordInfo)
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
				msg := fmt.Sprintf("Processing doc %d | Failed digitization for report id :%d category: %s %s | %s", idx+1, record.RecordId, record.RecordCategory, record.RecordName, recordInfo)
				log.Println("Error @GmailSyncCore->CreateDigitizationTask: ", taskErr)
				gs.processStatusService.LogStepAndFail(processID, step4, constant.Failure, msg, taskErr.Error(), nil, &record.RecordId, &attachmentId)
			}
			gs.processStatusService.LogStep(processID, step4, constant.Success, msg, errorMsg, &record.RecordId, nil, nil, nil, nil, &attachmentId)
		} else if record.RecordCategory == string(constant.INSURANCE) || record.RecordCategory == string(constant.INVOICE) || record.RecordCategory == string(constant.DISCHARGESUMMARY) || record.RecordCategory == string(constant.VACCINATION) {
			type DocTypeResponse struct {
				Content struct {
					PatientName string `json:"patient_name"`
				} `json:"content"`
			}
			var docResponse DocTypeResponse
			if err := json.Unmarshal(record.DocTypeResponseMetaData, &docResponse); err != nil {
				log.Println("failed to unmarshal DocTypeResponseMetadata:", err)
			}
			log.Println("DocTypeResponse ", record.DocTypeResponseMetaData)
			log.Println("DocTypeResponse patient name ", docResponse.Content.PatientName)
			_, patientDocInfo, matchNameError := gs.AssignDocToPatient(record.RecordCategory, docResponse.Content.PatientName, userId)
			if matchNameError != nil {
				log.Println("Error fetching patient info:", matchNameError)
			} else {
				// Pretty JSON print
				jsonData, err := json.MarshalIndent(patientDocInfo, "", "  ")
				if err != nil {
					log.Println("Error marshaling patientDocInfo:", err)
				} else {
					log.Println("AssignDocToPatient api response Patient doc Info:\n", string(jsonData))
				}
			}
			tx := gs.db.Begin()
			err := gs.recordRepo.UpdateMedicalRecordMappingByRecordId(tx, &record.RecordId, map[string]interface{}{"user_id": patientDocInfo.MatchedUserID, "is_unknown_record": patientDocInfo.IsFallback})
			if err != nil {
				return err
			}
			if err := tx.Commit().Error; err != nil {
				return err
			}
		}

	}
	msg5 := ""
	if len(emailMedRecord) > 0 {
		msg5 = fmt.Sprintf("Gmail Sync completed for %d records. These records are now being processed for digitization. You’ll be notified once the process is complete.", len(emailMedRecord))
	} else {
		msg5 = "Gmail Sync completed. No records found"
	}
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

func (gs *GmailSyncServiceImpl) AssignDocToPatient(keywordDocType, patientNameOnReport string, userId uint64) (constant.JobStatus, *models.PatientDocResponse, error) {
	status := constant.StatusQueued
	var patientDocInfo *models.PatientDocResponse
	var err error

	if keywordDocType == string(constant.OTHER) ||
		keywordDocType == string(constant.INSURANCE) ||
		keywordDocType == string(constant.VACCINATION) ||
		keywordDocType == string(constant.DISCHARGESUMMARY) ||
		keywordDocType == string(constant.INVOICE) ||
		keywordDocType == string(constant.NONMEDICAL) {

		status = constant.StatusSuccess

		// nested condition → fetch patient name on doc
		if keywordDocType == string(constant.INSURANCE) ||
			keywordDocType == string(constant.VACCINATION) ||
			keywordDocType == string(constant.DISCHARGESUMMARY) ||
			keywordDocType == string(constant.INVOICE) {

			patientDocInfo, err = gs.GetPatientNameOnDoc(patientNameOnReport, userId)
			if err != nil {
				return status, nil, err
			}
			return status, patientDocInfo, nil
		} else {
			userDetails, err := gs.patientService.GetUserProfileByUserId(userId)
			if err != nil {
				config.Log.Info("User not found", zap.Uint64("UserId", userId))
			}
			fullName := BuildFullName(userDetails.FirstName, userDetails.MiddleName, userDetails.LastName)
			fallback := false
			patientDocInfo = &models.PatientDocResponse{
				UserID:           userId,
				FinalPatientName: fullName,
				MatchedUserID:    userId,
				IsFallback:       fallback,
			}
			return status, patientDocInfo, nil
		}
	} else {
		userDetails, err := gs.patientService.GetUserProfileByUserId(userId)
		if err != nil {
			config.Log.Info("User not found", zap.Uint64("UserId", userId))
		}
		fullName := BuildFullName(userDetails.FirstName, userDetails.MiddleName, userDetails.LastName)
		fallback := false
		patientDocInfo = &models.PatientDocResponse{
			UserID:           userId,
			FinalPatientName: fullName,
			MatchedUserID:    userId,
			IsFallback:       fallback,
		}
		return status, patientDocInfo, nil
	}
}

func (gs *GmailSyncServiceImpl) GetPatientNameOnDoc(patientNameOnReport string, userID uint64) (*models.PatientDocResponse, error) {
	userDetails, err := gs.patientService.GetUserProfileByUserId(userID)
	if err != nil {
		config.Log.Info("User not found", zap.Uint64("UserId", userID))
	}
	relatives, _ := gs.patientService.GetRelativeList(&userID)
	var rels []models.Relative
	if len(relatives) > 0 {
		for _, r := range relatives {
			fullName := BuildFullName(r.FirstName, r.MiddleName, r.LastName)
			rels = append(rels, models.Relative{
				UserID: r.RelativeId,
				Name:   fullName,
			})
		}
		fullName := strings.TrimSpace(fmt.Sprintf("%s %s %s", userDetails.FirstName, userDetails.MiddleName, userDetails.LastName))
		req := &models.PatientDocRequest{
			UserID:            userID,
			PatientName:       fullName,
			Relatives:         rels,
			ReportPatientName: patientNameOnReport,
		}
		apiResp, err := gs.apiService.CallPatientDocInfoAPI(req)
		if err != nil {
			return &models.PatientDocResponse{}, err
		}
		return apiResp, nil
	} else {
		fullName := BuildFullName(userDetails.FirstName, userDetails.MiddleName, userDetails.LastName)
		fallback := true
		apiResp := &models.PatientDocResponse{
			UserID:           userID,
			FinalPatientName: fullName,
			MatchedUserID:    userID,
			IsFallback:       fallback,
		}
		return apiResp, nil
	}
}

func BuildFullName(first, middle, last string) string {
	var parts []string
	if first != "" {
		parts = append(parts, first)
	}
	if middle != "" {
		parts = append(parts, middle)
	}
	if last != "" {
		parts = append(parts, last)
	}
	return strings.Join(parts, " ")
}
