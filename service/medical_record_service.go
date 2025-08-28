package service

import (
	"biostat/config"
	"biostat/constant"
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type TblMedicalRecordService interface {
	GetAllMedicalRecord(patientId uint64, limit int, offset int) ([]map[string]interface{}, int64, error)
	GetUserMedicalRecords(userID uint64) ([]models.TblMedicalRecord, error)
	CreateTblMedicalRecord(createdBy uint64, authUserId string, file multipart.File, header *multipart.FileHeader, uploadSource string, description string, recordCategory, recordSubCategory string) (*models.TblMedicalRecord, error)
	CreateDigitizationTask(record *models.TblMedicalRecord, userInfo models.SystemUser_, userId uint64, file *bytes.Buffer, filename string, processID uuid.UUID, attachmentId *string) error
	EnqueueDocTypeCheckTask(attachmentId string, recordName string, fileData []byte, processID uuid.UUID) (*models.DocTypeAPIResponse, error)
	SaveMedicalRecords(data []*models.TblMedicalRecord, userId uint64) error
	UpdateTblMedicalRecord(data *models.TblMedicalRecord) (*models.TblMedicalRecord, error)
	GetMedicalRecordByRecordId(RecordId uint64) (*models.TblMedicalRecord, error)
	DeleteTblMedicalRecord(id int, updatedBy string) error
	IsRecordAccessibleToUser(userID uint64, recordID uint64) (bool, error)
	GetMedicalRecords(userID uint64, category string, limit, offset, isDeleted int) ([]models.MedicalRecordResponseRes, int64, map[string]int64, error)

	ReadMedicalRecord(ResourceId uint64, userId, reqUserId uint64) (interface{}, error)
	MovePatientRecord(patientId, targetPatientId, recordId, reportId uint64) error
}

type tblMedicalRecordServiceImpl struct {
	tblMedicalRecordRepo repository.TblMedicalRecordRepository
	apiService           ApiService
	diagnosticService    DiagnosticService
	patientService       PatientService
	userService          UserService
	taskQueue            *asynq.Client
	redisClient          *redis.Client
	processStatusService ProcessStatusService
}

func NewTblMedicalRecordService(repo repository.TblMedicalRecordRepository, apiService ApiService, diagnosticService DiagnosticService, patientService PatientService, userService UserService, taskQueue *asynq.Client,
	redisClient *redis.Client, processStatusService ProcessStatusService) TblMedicalRecordService {
	return &tblMedicalRecordServiceImpl{tblMedicalRecordRepo: repo, apiService: apiService, diagnosticService: diagnosticService, patientService: patientService, userService: userService, taskQueue: taskQueue,
		redisClient: redisClient, processStatusService: processStatusService}
}

func (s *tblMedicalRecordServiceImpl) GetUserMedicalRecords(userID uint64) ([]models.TblMedicalRecord, error) {
	return s.tblMedicalRecordRepo.GetMedicalRecordsByUserID(userID, nil)
}

func (s *tblMedicalRecordServiceImpl) GetAllMedicalRecord(patientId uint64, limit int, offset int) ([]map[string]interface{}, int64, error) {
	data, totalRecords, err := s.tblMedicalRecordRepo.GetAllMedicalRecord(patientId, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	processed := s.tblMedicalRecordRepo.ProcessMedicalRecordResponse(data)
	return processed, totalRecords, nil
}

func (s *tblMedicalRecordServiceImpl) CreateTblMedicalRecord(userId uint64, authUserId string, file multipart.File, header *multipart.FileHeader, uploadSource string, description string, recordCategory, recordSubCategory string) (*models.TblMedicalRecord, error) {
	processType := string(constant.ManualRecordUpload)
	step := string(constant.ProcessSaveRecords)
	msg := string(constant.SaveRecord)
	errorMsg := ""
	processID, _ := s.processStatusService.StartProcessInRedis(userId, processType, strconv.FormatUint(userId, 10),
		string(constant.MedicalRecordEntity),
		step,
	)
	s.processStatusService.LogStep(processID, step, constant.Running, msg, errorMsg, nil, nil, nil, nil, nil, nil)
	uploadingPerson, err := s.userService.GetUserIdBySUB(authUserId)
	if err != nil {
		log.Println("GetUserIdBySUB uploadingPerson : ", err)
		return nil, err
	}

	fileName := utils.SanitizeFileName(header.Filename)
	uniqueSuffix := time.Now().Format("20060102150405") + "-" + uuid.New().String()[:8]
	ext := filepath.Ext(fileName)
	originalName := strings.TrimSuffix(fileName, ext)
	safeFileName := fmt.Sprintf("%s_%s%s", originalName, uniqueSuffix, ext)
	destinationPath := filepath.Join("uploads", safeFileName)
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		return nil, err
	}
	if err := utils.SaveFile(header, destinationPath); err != nil {
		log.Println("save file error : ", err)
		return nil, err
	}

	var fileBuf bytes.Buffer
	tee := io.TeeReader(file, &fileBuf)
	if _, err := io.ReadAll(tee); err != nil {
		return nil, err
	}
	Status := constant.StatusQueued
	if recordCategory == string(constant.OTHER) || recordCategory == string(constant.INSURANCE) || recordCategory == string(constant.VACCINATION) || recordCategory == string(constant.DISCHARGESUMMARY) || recordCategory == string(constant.INVOICE) || recordCategory == string(constant.NONMEDICAL) || recordCategory == string(constant.SCANS) {
		Status = constant.StatusSuccess
	}
	newRecord := models.TblMedicalRecord{
		RecordName:        header.Filename,
		RecordSize:        int64(header.Size),
		FileType:          header.Header.Get("Content-Type"),
		RecordUrl:         fmt.Sprintf("%s/uploads/%s", os.Getenv("SHORT_URL_BASE"), safeFileName),
		UploadDestination: "LocalServer",
		UploadSource:      uploadSource,
		Description:       description,
		RecordCategory:    recordCategory,
		RecordSubCategory: recordSubCategory,
		FetchedAt:         time.Now(),
		UploadedBy:        uploadingPerson,
		SourceAccount:     fmt.Sprint(uploadSource),
		Status:            Status,
	}

	record, err := s.tblMedicalRecordRepo.CreateTblMedicalRecord(&newRecord)
	if err != nil {
		msg = "Failed to save record"
		log.Println("CreateTblMedicalRecord ERROR : ", err)
		s.processStatusService.LogStepAndFail(processID, step, constant.Failure, msg, err.Error(), nil, nil, nil)
		return nil, err
	}
	var mappings []models.TblMedicalRecordUserMapping
	mappings = append(mappings, models.TblMedicalRecordUserMapping{
		UserID:   userId,
		RecordID: record.RecordId,
	})
	mappingErr := s.tblMedicalRecordRepo.CreateMedicalRecordMappings(&mappings)
	if mappingErr != nil {
		log.Println("CreateMedicalRecordMappings Mapping  ERROR : ", err)
		return nil, mappingErr
	}
	if record.RecordCategory == string(constant.TESTREPORT) || record.RecordCategory == string(constant.MEDICATION) {
		userInfo, err := s.userService.GetSystemUserInfoByUserID(userId)
		if err != nil {
			log.Println("GetSystemUserInfoByUserID ERROR : ", err)
			return nil, err
		}
		if err := s.CreateDigitizationTask(record, userInfo, userId, &fileBuf, fileName, processID, nil); err != nil {
			log.Printf("Digitization task failed: %v", err)
			s.processStatusService.LogStepAndFail(processID, step, constant.Failure, msg, err.Error(), nil, nil, nil)
		}
		s.processStatusService.LogStep(processID, step, constant.Success, "Record saved, digitization is in progress", errorMsg, nil, nil, nil, nil, nil, nil)
	} else {
		s.processStatusService.LogStep(processID, step, constant.Success, "Record saved successfully", errorMsg, nil, nil, nil, nil, nil, nil)
	}
	return record, nil
}

func (s *tblMedicalRecordServiceImpl) CreateDigitizationTask(record *models.TblMedicalRecord, userInfo models.SystemUser_,
	userId uint64, fileBuf *bytes.Buffer, filename string, processID uuid.UUID, attachmentId *string) error {
	if record.RecordCategory == string(constant.TESTREPORT) || record.RecordCategory == string(constant.MEDICATION) {
		log.Println("Queue worker starts............")
		tempDir := os.TempDir()
		tempPath := filepath.Join(tempDir, fmt.Sprintf("record_%d_%s", record.RecordId, filename))
		// var fileBytes []byte
		fileBytes := fileBuf.Bytes()
		if record.IsPasswordProtected {
			log.Println("PDF is password protected, decrypting before saving...")
			decryptedBytes, err := DecryptPDFIfProtected(fileBytes, record.PDFPassword)
			if err != nil {
				log.Printf("Failed to decrypt PDF for record %d: %v", record.RecordId, err)
				return err
			}
			fileBytes = decryptedBytes
			log.Println("Decryption successful, proceeding with saving decrypted file.")
		}

		// if err := os.WriteFile(tempPath, fileBuf.Bytes(), 0644); err != nil {
		if err := os.WriteFile(tempPath, fileBytes, 0644); err != nil {
			log.Printf("Failed to write temp file for record %d: %v", record.RecordId, err)
			return err
		}
		payload := models.DigitizationPayload{
			RecordID:     record.RecordId,
			UserID:       userId,
			PatientName:  userInfo.FirstName + " " + userInfo.MiddleName + " " + userInfo.LastName,
			FilePath:     tempPath,
			Category:     record.RecordCategory,
			FileName:     filename,
			ProcessID:    processID,
			AttachmentId: attachmentId,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal digitization payload for record %d: %v", record.RecordId, err)
			return err
		}

		task := asynq.NewTask("digitize:record", payloadBytes)
		if _, err := s.taskQueue.Enqueue(task, asynq.MaxRetry(config.PropConfig.Retry.MaxAttempts), asynq.Retention(time.Duration(config.PropConfig.TaskQueue.Retention)), asynq.ProcessIn(time.Duration(config.PropConfig.TaskQueue.Delay))); err != nil {
			log.Printf("Failed to enqueue digitization task for record %d: %v", record.RecordId, err)
			return err
		}
		log.Printf("record Id : %d : status : %s", record.RecordId, "queued")
		s.redisClient.Set(context.Background(), fmt.Sprintf("record_status:%d", record.RecordId), "queued", time.Duration(config.PropConfig.TaskQueue.Expiration))
	}
	return nil
}

// Global map to store responses for doc type checks
// var DocTypeResponses = struct {
// 	sync.Mutex
// 	Data map[string]chan string
// }{Data: make(map[string]chan string)}

var DocTypeResponses = struct {
	sync.Mutex
	Data map[string]chan *models.DocTypeAPIResponse
}{Data: make(map[string]chan *models.DocTypeAPIResponse)}

func (mrs *tblMedicalRecordServiceImpl) EnqueueDocTypeCheckTask(
	attachmentId string,
	recordName string,
	fileData []byte,
	processID uuid.UUID,
) (*models.DocTypeAPIResponse, error) {
	payload := models.DocTypeCheckPayload{
		AttachmentID: attachmentId,
		FileName:     recordName,
		FileBytes:    fileData,
		ProcessID:    processID,
	}

	ch := make(chan *models.DocTypeAPIResponse, 1)

	DocTypeResponses.Lock()
	DocTypeResponses.Data[attachmentId] = ch
	DocTypeResponses.Unlock()

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		mrs.cleanupDocTypeChannel(attachmentId)
		return &models.DocTypeAPIResponse{}, fmt.Errorf("failed to marshal doc type payload: %w", err)
	}

	task := asynq.NewTask("check:doctype", payloadBytes)
	_, err = mrs.taskQueue.Enqueue(
		task,
		asynq.MaxRetry(config.PropConfig.Retry.MaxAttempts),
		asynq.Retention(time.Duration(config.PropConfig.TaskQueue.Retention)),
		asynq.ProcessIn(time.Duration(config.PropConfig.TaskQueue.Delay)),
	)
	if err != nil {
		mrs.cleanupDocTypeChannel(attachmentId)
		return &models.DocTypeAPIResponse{}, fmt.Errorf("failed to enqueue doc type check task: %w", err)
	}

	docType := <-ch
	// log.Printf("Doc type for record %s: %v", attachmentId, docType)

	mrs.cleanupDocTypeChannel(attachmentId)
	return docType, nil
}

// cleanupDocTypeChannel safely removes a record's channel
func (mrs *tblMedicalRecordServiceImpl) cleanupDocTypeChannel(recordID string) {
	DocTypeResponses.Lock()
	delete(DocTypeResponses.Data, recordID)
	DocTypeResponses.Unlock()
}

func MatchPatientNameWithRelative(relatives []models.PatientRelative, patientName string, fallbackUserID uint64, systemPatientName string) (uint64, string, bool, string) {
	normalizedPatientName := strings.TrimSpace(strings.ToLower(patientName))
	config.Log.Info("Patient name on report returned from AI service", zap.String("Patient name", patientName))

	highestScore := -1
	var bestMatchID uint64
	var bestMatchName string
	isUnknownReport := false
	var matchMessage string

	// Match against system patient name
	systemNameParts := strings.Fields(strings.TrimSpace(systemPatientName))
	systemPermutations := utils.GeneratePermutations(systemNameParts)

	for _, perm := range systemPermutations {
		full := strings.ToLower(strings.Join(perm, " "))
		score := utils.CalculateNameScore(normalizedPatientName, full)
		log.Printf("Matching with system patient name permutation '%s' | Score: %d", full, score)

		if score > highestScore && score >= 30 {
			highestScore = score
			bestMatchID = fallbackUserID
			bestMatchName = full
			matchMessage = fmt.Sprintf("Report matches with system patient name '%s' (self) : Score : %d", bestMatchName, highestScore)
			isUnknownReport = true
		}
	}

	// Match against relatives
	for _, relative := range relatives {
		nameParts := []string{}
		if relative.FirstName != "" {
			nameParts = append(nameParts, relative.FirstName)
		}
		if relative.MiddleName != "" {
			nameParts = append(nameParts, relative.MiddleName)
		}
		if relative.LastName != "" {
			nameParts = append(nameParts, relative.LastName)
		}

		relativePermutations := utils.GeneratePermutations(nameParts)

		for _, perm := range relativePermutations {
			full := strings.ToLower(strings.Join(perm, " "))
			score := utils.CalculateNameScore(normalizedPatientName, full)
			log.Printf("Matching with relative permutation '%s' | Score: %d", full, score)

			if score > highestScore && score >= 30 {
				highestScore = score
				bestMatchID = relative.RelativeId
				bestMatchName = full
				matchMessage = fmt.Sprintf("Report matches with relative '%s' : Score : %d", bestMatchName, highestScore)
				isUnknownReport = true
			}
		}
	}
	if !isUnknownReport {
		bestMatchID = fallbackUserID
		bestMatchName = systemPatientName
		matchMessage = fmt.Sprintf("No good match found. Falling back to system patient name '%s' (User ID: %d) (SELF) in OTHER Bucket", bestMatchName, bestMatchID)
		log.Print(matchMessage)
		isUnknownReport = true
	} else {
		isUnknownReport = false
	}

	log.Printf("Best report name match with: '%s' | User ID: %d | Score: %d | isUnknownReport: %v", bestMatchName, bestMatchID, highestScore, isUnknownReport)
	return bestMatchID, bestMatchName, isUnknownReport, matchMessage
}

func (s *tblMedicalRecordServiceImpl) SaveMedicalRecords(records []*models.TblMedicalRecord, userId uint64) error {
	var uniqueRecords []*models.TblMedicalRecord
	for _, record := range records {
		exists, err := s.tblMedicalRecordRepo.ExistsRecordForUser(userId, record.UploadSource, record.RecordUrl)
		if err != nil {
			return err
		}
		if !exists {
			uniqueRecords = append(uniqueRecords, record)
		}
	}
	if len(uniqueRecords) == 0 {
		return nil
	}

	err := s.tblMedicalRecordRepo.CreateMultipleTblMedicalRecords(uniqueRecords)
	if err != nil {
		return err
	}
	var mappings []models.TblMedicalRecordUserMapping
	for _, record := range uniqueRecords {
		mappings = append(mappings, models.TblMedicalRecordUserMapping{
			UserID:   userId,
			RecordID: record.RecordId,
		})
	}
	return s.tblMedicalRecordRepo.CreateMedicalRecordMappings(&mappings)
}

func (s *tblMedicalRecordServiceImpl) UpdateTblMedicalRecord(data *models.TblMedicalRecord) (*models.TblMedicalRecord, error) {
	return s.tblMedicalRecordRepo.UpdateTblMedicalRecord(data)
}

func (s *tblMedicalRecordServiceImpl) GetMedicalRecordByRecordId(RecordId uint64) (*models.TblMedicalRecord, error) {
	return s.tblMedicalRecordRepo.GetMedicalRecordByRecordId(RecordId)
}

func (s *tblMedicalRecordServiceImpl) DeleteTblMedicalRecord(id int, updatedBy string) error {
	return s.tblMedicalRecordRepo.DeleteTblMedicalRecordWithMappings(id, updatedBy)
}

func (s *tblMedicalRecordServiceImpl) IsRecordAccessibleToUser(userID uint64, recordID uint64) (bool, error) {
	belongsTouser, _ := s.tblMedicalRecordRepo.IsRecordBelongsToUser(userID, recordID)
	if belongsTouser == true {
		return true, nil
	}
	mapping, err := s.tblMedicalRecordRepo.GetMedicalRecordMappings(recordID)
	if err != nil || mapping == nil {
		return false, err
	}
	relative, err := s.patientService.GetPatientRelativeById(mapping.UserID, userID)
	log.Println(relative)
	if err != nil || relative.RelativeId == 0 {
		return false, err
	}
	return true, err
}

func (s *tblMedicalRecordServiceImpl) ReadUserDigiLockerFile(userId uint64, digiLockerFileUrl string) (*models.DigiLockerFile, error) {
	userDigiToken, err := s.userService.GetSingleTblUserToken(userId, "DigiLocker")
	if err != nil {
		return nil, err
	}
	digiResp, err := ReadDigiLockerFile(userDigiToken.AuthToken, digiLockerFileUrl)
	if err != nil {
		return nil, err
	}
	return digiResp, nil
}

func (s *tblMedicalRecordServiceImpl) ReadUserLocalServerFile(localFileUrl string) (*models.LocalServerFile, error) {
	log.Println("Read Local")
	var res models.LocalServerFile
	urlParts := strings.Split(localFileUrl, "/uploads/")
	if len(urlParts) < 2 {
		return nil, errors.New("invalid file url")
	}
	filename := urlParts[1]
	localPath := fmt.Sprintf("uploads/%s", filename)
	fileBytes, err := os.ReadFile(localPath)
	if err != nil {
		return nil, err
	}
	res.ContentType = http.DetectContentType(fileBytes)
	res.Data = fileBytes
	return &res, nil
}

func (s *tblMedicalRecordServiceImpl) ReadMedicalRecord(ResourceId uint64, userId, reqUserId uint64) (interface{}, error) {
	isAccessible, err := s.IsRecordAccessibleToUser(reqUserId, ResourceId)
	if err != nil {
		return nil, err
	}
	if !isAccessible {
		return nil, errors.New("you do not have access to view report")
	}

	record, err := s.GetMedicalRecordByRecordId(ResourceId)
	if err != nil {
		return nil, err
	}
	var response struct {
		Data        []byte `json:"data"`
		ContentType string `json:"content-type"`
		HMAC        string `json:"hmac,omitempty"`
	}
	switch record.UploadDestination {
	case "DigiLocker":
		digiFile, err := s.ReadUserDigiLockerFile(reqUserId, record.RecordUrl)
		if err != nil {
			return nil, err
		}
		response.ContentType = digiFile.ContentType
		response.Data = digiFile.Data
		response.HMAC = digiFile.HMAC
	case "LocalServer":
		localFile, err := s.ReadUserLocalServerFile(record.RecordUrl)
		if err != nil {
			return nil, err
		}
		response.Data = localFile.Data
		response.ContentType = localFile.ContentType
		response.HMAC = ""
	}
	return response, nil
}

func (s *tblMedicalRecordServiceImpl) GetMedicalRecords(userID uint64, category string, limit, offset, isDeleted int) ([]models.MedicalRecordResponseRes, int64, map[string]int64, error) {
	records, total, counts, err := s.tblMedicalRecordRepo.GetMedicalRecordsByUser(userID, category, limit, offset, isDeleted)
	if err != nil {
		return nil, 0, nil, err
	}

	var recordIDs []uint64
	for _, r := range records {
		recordIDs = append(recordIDs, r.RecordId)
	}

	attachments, err := s.tblMedicalRecordRepo.GetDiagnosticAttachmentByRecordIDs(recordIDs)
	if err != nil {
		return nil, 0, nil, err
	}

	reportMap := make(map[uint64]uint64)
	for _, a := range attachments {
		reportMap[a.RecordId] = uint64(a.PatientDiagnosticReportId) //a.PatientDiagnosticReportID
	}

	var responses []models.MedicalRecordResponseRes
	for _, rec := range records {
		resp := models.MedicalRecordResponseRes{
			DigitizeFlag:              rec.DigitizeFlag, //rec.DigitizeFlag,
			FileType:                  rec.FileType,
			PatientID:                 userID,
			RecordCategory:            rec.RecordCategory,
			RecordID:                  rec.RecordId,
			IsDeleted:                 rec.IsDeleted,
			RecordName:                rec.RecordName,
			RecordSize:                rec.RecordSize,
			RecordURL:                 rec.RecordUrl,
			IsVerified:                rec.IsVerified,
			RecordDescription:         rec.Description,
			SourceAccount:             rec.SourceAccount,
			Status:                    string(rec.Status),
			UploadSource:              rec.UploadSource,
			ErrorMessage:              rec.ErrorMessage,
			CreatedAt:                 utils.FormatDateTime(&rec.CreatedAt),
			PatientDiagnosticReportID: "0",
		}

		if reportID, ok := reportMap[rec.RecordId]; ok {
			resp.PatientDiagnosticReportID = fmt.Sprintf("%d", reportID)
			report, err := s.tblMedicalRecordRepo.GetDiagnosticReport(reportID, isDeleted)
			if err != nil {
				log.Println("@GetMedicalRecords->GetDiagnosticReport:", err)
				continue
			}

			diagnostic := &models.UploadedDiagnosticRes{
				CollectedAt:     report.CollectedAt,
				CollectedDate:   utils.FormatDateTime(&report.CollectedDate),
				Comments:        report.Comments,
				DiagnosticLabID: report.DiagnosticLabId,
				LabName:         "", // can fetch if needed
				ReportDate:      utils.FormatDateTime(&report.ReportDate),
				ReportName:      report.ReportName,
				IsDeleted:       report.IsDeleted,
				ReportStatus:    report.ReportStatus,
			}

			tests, _ := s.tblMedicalRecordRepo.GetDiagnosticTests(reportID)
			for _, test := range tests {
				testMaster, _ := s.tblMedicalRecordRepo.GetDiagnosticTestMaster(test.DiagnosticTestId) //(test.DiagnosticTestID)
				dt := models.DiagnosticTestRes{
					DiagnosticTestID: test.DiagnosticTestId,
					TestName:         testMaster.TestName,
					TestNote:         test.TestNote,
					TestDate:         test.TestDate, // if available
				}

				results, _ := s.tblMedicalRecordRepo.GetTestComponents(reportID, test.DiagnosticTestId)
				for _, result := range results {
					comp, _ := s.tblMedicalRecordRepo.GetComponentDetails(result.DiagnosticTestComponentId) //(result.DiagnosticTestComponentID)
					ranges, _ := s.tblMedicalRecordRepo.GetReferenceRanges(result.DiagnosticTestComponentId)

					dtc := models.TestComponentRes{
						DiagnosticTestComponentID: result.DiagnosticTestComponentId,
						TestComponentName:         comp.TestComponentName,
						Units:                     comp.Units,
					}

					for _, r := range ranges {
						dtc.TestReferenceRange = append(dtc.TestReferenceRange, models.DiagnosticReferenceRangeRes{
							Age:       r.Age,
							AgeGroup:  r.AgeGroup,
							Gender:    r.Gender,
							NormalMin: fmt.Sprintf("%.2f", r.NormalMin),
							NormalMax: fmt.Sprintf("%.2f", r.NormalMax),
							Units:     r.Units,
						})
					}

					dtc.TestResultValue = append(dtc.TestResultValue, models.TestResultValueRes{
						ResultValue:   fmt.Sprintf("%.2f", result.ResultValue),
						Qualifier:     "",
						ResultComment: result.ResultComment,
						ResultDate:    utils.FormatDateTime(&result.ResultDate),
						ResultStatus:  result.ResultStatus,
					})

					dt.TestComponents = append(dt.TestComponents, dtc)
				}

				diagnostic.PatientDiagnosticTest = append(diagnostic.PatientDiagnosticTest, dt)
			}

			resp.UploadedDiagnostic = diagnostic
		}

		responses = append(responses, resp)
	}

	return responses, total, counts, nil
}

func (s *tblMedicalRecordServiceImpl) MovePatientRecord(patientId, targetPatientId, recordId, reportId uint64) error {
	return s.tblMedicalRecordRepo.MovePatientRecord(patientId, targetPatientId, recordId, reportId)
}
