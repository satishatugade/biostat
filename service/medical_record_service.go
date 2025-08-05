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
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type TblMedicalRecordService interface {
	GetAllMedicalRecord(patientId uint64, limit int, offset int) ([]map[string]interface{}, int64, error)
	GetUserMedicalRecords(userID uint64) ([]models.TblMedicalRecord, error)
	CreateTblMedicalRecord(createdBy uint64, authUserId string, file multipart.File, header *multipart.FileHeader, uploadSource string, description string, recordCategory string) (*models.TblMedicalRecord, error)
	CreateDigitizationTask(record *models.TblMedicalRecord, userInfo models.SystemUser_, userId uint64, authUserId string, file *bytes.Buffer, filename string) error
	SaveMedicalRecords(data []*models.TblMedicalRecord, userId uint64) error
	UpdateTblMedicalRecord(data *models.TblMedicalRecord) (*models.TblMedicalRecord, error)
	GetMedicalRecordByRecordId(RecordId uint64) (*models.TblMedicalRecord, error)
	DeleteTblMedicalRecord(id int, updatedBy string) error
	IsRecordAccessibleToUser(userID uint64, recordID uint64) (bool, error)
	GetMedicalRecords(userID uint64, limit, offset, isDeleted int) ([]models.MedicalRecordResponseRes, int64, error)

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

func (s *tblMedicalRecordServiceImpl) CreateTblMedicalRecord(userId uint64, authUserId string, file multipart.File, header *multipart.FileHeader, uploadSource string, description string, recordCategory string) (*models.TblMedicalRecord, error) {
	processID := uuid.New()
	key, _ := s.processStatusService.StartProcessRedis(processID, userId, "record_upload", fmt.Sprintf("UserID %s", strconv.FormatUint(userId, 10)), "tbl_medical_record", "record_saving")
	uploadingPerson, err := s.userService.GetUserIdBySUB(authUserId)
	if err != nil {
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
		return nil, err
	}

	var fileBuf bytes.Buffer
	tee := io.TeeReader(file, &fileBuf)
	if _, err := io.ReadAll(tee); err != nil {
		return nil, err
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
		FetchedAt:         time.Now(),
		UploadedBy:        uploadingPerson,
		SourceAccount:     fmt.Sprint(uploadSource),
		Status:            constant.StatusQueued,
	}

	record, err := s.tblMedicalRecordRepo.CreateTblMedicalRecord(&newRecord)
	if err != nil {
		s.processStatusService.UpdateProcessRedis(key, constant.Failure, nil, "Failed to save record", "record_saving", true)
		return nil, err
	}
	var mappings []models.TblMedicalRecordUserMapping
	mappings = append(mappings, models.TblMedicalRecordUserMapping{
		UserID:   userId,
		RecordID: record.RecordId,
	})
	s.processStatusService.UpdateProcessRedis(key, constant.Running, nil, "Record saved to database", "record_saving", false)
	mappingErr := s.tblMedicalRecordRepo.CreateMedicalRecordMappings(&mappings)
	if mappingErr != nil {
		return nil, mappingErr
	}
	userInfo, err := s.userService.GetSystemUserInfoByUserID(userId)
	if err != nil {
		return nil, err
	}
	if err := s.CreateDigitizationTask(record, userInfo, userId, authUserId, &fileBuf, fileName); err != nil {
		s.processStatusService.UpdateProcessRedis(key, constant.Failure, nil, "Record saved, but digitization failed", "digitization", true)
		log.Printf("Digitization task failed: %v", err)
	}
	s.processStatusService.UpdateProcessRedis(key, constant.Success, nil, "Record saved, digitization is in progress", "digitization", true)
	return record, nil
}

func (s *tblMedicalRecordServiceImpl) CreateDigitizationTask(record *models.TblMedicalRecord, userInfo models.SystemUser_,
	userId uint64, authUserId string, fileBuf *bytes.Buffer, filename string) error {
	if record.RecordCategory == "Test Reports" || record.RecordCategory == "Prescriptions" || record.RecordCategory == "test_report" {
		log.Println("Queue worker starts............")
		tempDir := os.TempDir()
		tempPath := filepath.Join(tempDir, fmt.Sprintf("record_%d_%s", record.RecordId, filename))

		if err := os.WriteFile(tempPath, fileBuf.Bytes(), 0644); err != nil {
			log.Printf("Failed to write temp file for record %d: %v", record.RecordId, err)
			return err
		}
		payload := models.DigitizationPayload{
			RecordID:    record.RecordId,
			UserID:      userId,
			PatientName: userInfo.FirstName + " " + userInfo.MiddleName + " " + userInfo.LastName,
			FilePath:    tempPath,
			Category:    record.RecordCategory,
			FileName:    filename,
			AuthUserID:  authUserId,
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

func MatchPatientNameWithRelative(relatives []models.PatientRelative, patientName string, fallbackUserID uint64, systemPatientName string) (uint64, bool) {
	normalizedPatientName := strings.TrimSpace(strings.ToLower(patientName))
	config.Log.Info("Patient name on report returned from AI service", zap.String("Patient name", patientName))

	highestScore := -1
	var bestMatchID uint64
	var bestMatchName string
	isUnknownReport := false

	// Match against system patient name
	systemNameParts := strings.Fields(strings.TrimSpace(systemPatientName))
	systemPermutations := utils.GeneratePermutations(systemNameParts)

	for _, perm := range systemPermutations {
		full := strings.ToLower(strings.Join(perm, " "))
		score := utils.CalculateNameScore(normalizedPatientName, full)
		log.Printf("Matching with system patient name permutation '%s' | Score: %d", full, score)

		if score > highestScore && score >= 60 {
			highestScore = score
			bestMatchID = fallbackUserID
			bestMatchName = full
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

			if score > highestScore && score >= 60 {
				highestScore = score
				bestMatchID = relative.RelativeId
				bestMatchName = full
				isUnknownReport = true
			}
		}
	}
	if !isUnknownReport {
		bestMatchID = fallbackUserID
		bestMatchName = systemPatientName
		log.Printf("No good match found. Falling back to system patient name '%s' (User ID: %d)", bestMatchName, bestMatchID)
		isUnknownReport = true
	} else {
		isUnknownReport = false
	}

	log.Printf("Best report name match with: '%s' | User ID: %d | Score: %d | isUnknownReport: %v", bestMatchName, bestMatchID, highestScore, isUnknownReport)
	return bestMatchID, isUnknownReport
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
		log.Printf("Creating mapping for %d", record.RecordId)
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
	belongsTouser, err := s.tblMedicalRecordRepo.IsRecordBelongsToUser(userID, recordID)
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
	isAccessible, err := s.IsRecordAccessibleToUser(userId, ResourceId)
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

func (s *tblMedicalRecordServiceImpl) GetMedicalRecords(userID uint64, limit, offset, isDeleted int) ([]models.MedicalRecordResponseRes, int64, error) {
	records, total, err := s.tblMedicalRecordRepo.GetMedicalRecordsByUser(userID, limit, offset, isDeleted)
	if err != nil {
		return nil, 0, err
	}

	var recordIDs []uint64
	for _, r := range records {
		recordIDs = append(recordIDs, r.RecordId)
	}

	attachments, err := s.tblMedicalRecordRepo.GetDiagnosticAttachmentByRecordIDs(recordIDs)
	if err != nil {
		return nil, 0, err
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
			SourceAccount:             rec.SourceAccount,
			Status:                    string(rec.Status),
			UploadSource:              rec.UploadSource,
			ErrorMessage:              rec.ErrorMessage,
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

	return responses, total, nil
}

func (s *tblMedicalRecordServiceImpl) MovePatientRecord(patientId, targetPatientId, recordId, reportId uint64) error {
	return s.tblMedicalRecordRepo.MovePatientRecord(patientId, targetPatientId, recordId, reportId)
}
