package service

import (
	"biostat/models"
	"biostat/repository"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/xrash/smetrics"
)

type TblMedicalRecordService interface {
	GetAllMedicalRecord(patientId uint64, limit int, offset int) ([]map[string]interface{}, int64, error)
	GetUserMedicalRecords(userID uint64) ([]models.TblMedicalRecord, error)
	CreateTblMedicalRecord(data *models.TblMedicalRecord, createdBy uint64, authUserId string, file *bytes.Buffer, filename string) (*models.TblMedicalRecord, error)
	CreateDigitizationTask(record *models.TblMedicalRecord, userInfo models.SystemUser_, userId uint64, authUserId string, file *bytes.Buffer, filename string) error
	SaveMedicalRecords(data []*models.TblMedicalRecord, userId uint64) error
	UpdateTblMedicalRecord(data *models.TblMedicalRecord) (*models.TblMedicalRecord, error)
	GetMedicalRecordByRecordId(RecordId uint64) (*models.TblMedicalRecord, error)
	DeleteTblMedicalRecord(id int, updatedBy string) error
	IsRecordAccessibleToUser(userID uint64, recordID uint64) (bool, error)
	GetMedicalRecords(userID uint64, limit, offset int) ([]models.MedicalRecordResponseRes, int64, error)

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
}

func NewTblMedicalRecordService(repo repository.TblMedicalRecordRepository, apiService ApiService, diagnosticService DiagnosticService, patientService PatientService, userService UserService, taskQueue *asynq.Client,
	redisClient *redis.Client) TblMedicalRecordService {
	return &tblMedicalRecordServiceImpl{tblMedicalRecordRepo: repo, apiService: apiService, diagnosticService: diagnosticService, patientService: patientService, userService: userService, taskQueue: taskQueue,
		redisClient: redisClient}
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

func (s *tblMedicalRecordServiceImpl) CreateTblMedicalRecord(data *models.TblMedicalRecord, userId uint64, authUserId string, fileBuf *bytes.Buffer, filename string) (*models.TblMedicalRecord, error) {
	record, err := s.tblMedicalRecordRepo.CreateTblMedicalRecord(data)
	if err != nil {
		return nil, err
	}
	var mappings []models.TblMedicalRecordUserMapping
	mappings = append(mappings, models.TblMedicalRecordUserMapping{
		UserID:   userId,
		RecordID: record.RecordId,
	})
	mappingErr := s.tblMedicalRecordRepo.CreateMedicalRecordMappings(&mappings)
	if mappingErr != nil {
		return nil, mappingErr
	}
	userInfo, err := s.userService.GetSystemUserInfo(authUserId)
	if err != nil {
		return nil, err
	}
	if err := s.CreateDigitizationTask(record, userInfo, userId, authUserId, fileBuf, filename); err != nil {
		log.Printf("Digitization task failed: %v", err)
	}
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
		if _, err := s.taskQueue.Enqueue(task, asynq.MaxRetry(1), asynq.Retention(24*time.Hour), asynq.ProcessIn(5*time.Minute)); err != nil {
			log.Printf("Failed to enqueue digitization task for record %d: %v", record.RecordId, err)
			return err
		}
		log.Printf("record Id : %d : status : %s", record.RecordId, "queued")
		s.redisClient.Set(context.Background(), fmt.Sprintf("record_status:%d", record.RecordId), "queued", 0)
	}
	return nil
}

func MatchPatientNameWithRelative(relatives []models.PatientRelative, patientName string, fallbackUserID uint64, systemPatientName string) uint64 {
	normalizedPatientName := strings.TrimSpace(strings.ToLower(patientName))

	highestScore := 0
	bestMatchID := fallbackUserID
	bestMatchName := systemPatientName

	normalizedSystemName := strings.TrimSpace(strings.ToLower(systemPatientName))
	soundexPatient := smetrics.Soundex(normalizedPatientName)
	soundexSystem := smetrics.Soundex(normalizedSystemName)
	sysLevDist := smetrics.WagnerFischer(normalizedPatientName, normalizedSystemName, 1, 1, 2)
	sysMaxLen := max(len(normalizedPatientName), len(normalizedSystemName))
	sysSimilarity := 100 - (sysLevDist * 100 / sysMaxLen)

	sysScore := sysSimilarity
	if soundexPatient == soundexSystem {
		sysScore += 20
	}

	log.Printf("Matching with system patient name '%s' | Similarity: %d%% | Score: %d", normalizedSystemName, sysSimilarity, sysScore)

	if sysScore > highestScore {
		highestScore = sysScore
		bestMatchID = fallbackUserID
		bestMatchName = systemPatientName
	}

	for _, relative := range relatives {
		fullName := strings.TrimSpace(strings.ToLower(relative.FirstName + " " + relative.MiddleName + " " + relative.LastName))
		soundexRelative := smetrics.Soundex(fullName)
		levDist := smetrics.WagnerFischer(normalizedPatientName, fullName, 1, 1, 2)
		maxLen := max(len(normalizedPatientName), len(fullName))
		similarity := 100 - (levDist * 100 / maxLen)

		score := similarity
		if soundexPatient == soundexRelative {
			score += 20
		}

		log.Printf("Matching with relative '%s' | Similarity: %d%% | Score: %d", fullName, similarity, score)

		if score > highestScore && score >= 40 {
			highestScore = score
			bestMatchID = relative.RelativeId
			bestMatchName = relative.FirstName + " " + relative.MiddleName + " " + relative.LastName
		}
	}

	log.Printf("Best report name match with : '%s' | User ID: %d | Score: %d", bestMatchName, bestMatchID, highestScore)
	return bestMatchID
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
		log.Println("Creating mapping for %s", record.RecordId)
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
		return nil, errors.New("you do not have access to this resource")
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
	if record.UploadDestination == "DigiLocker" {
		digiFile, err := s.ReadUserDigiLockerFile(reqUserId, record.RecordUrl)
		if err != nil {
			return nil, err
		}
		response.ContentType = digiFile.ContentType
		response.Data = digiFile.Data
		response.HMAC = digiFile.HMAC
	} else if record.UploadDestination == "LocalServer" {
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

func (s *tblMedicalRecordServiceImpl) GetMedicalRecords(userID uint64, limit, offset int) ([]models.MedicalRecordResponseRes, int64, error) {
	records, total, err := s.tblMedicalRecordRepo.GetMedicalRecordsByUser(userID, limit, offset)
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

	reportMap := make(map[uint64]int64)
	for _, a := range attachments {
		reportMap[a.RecordId] = int64(a.PatientDiagnosticReportId) //a.PatientDiagnosticReportID
	}

	var responses []models.MedicalRecordResponseRes
	for _, rec := range records {
		resp := models.MedicalRecordResponseRes{
			DigitizeFlag:              rec.DigitizeFlag, //rec.DigitizeFlag,
			FileType:                  rec.FileType,
			PatientID:                 userID,
			RecordCategory:            rec.RecordCategory,
			RecordID:                  rec.RecordId,
			RecordName:                rec.RecordName,
			RecordSize:                rec.RecordSize,
			RecordURL:                 rec.RecordUrl,
			SourceAccount:             rec.SourceAccount,
			Status:                    string(rec.Status),
			UploadSource:              rec.UploadSource,
			PatientDiagnosticReportID: "0",
		}

		if reportID, ok := reportMap[rec.RecordId]; ok {
			resp.PatientDiagnosticReportID = fmt.Sprintf("%d", reportID)
			report, err := s.tblMedicalRecordRepo.GetDiagnosticReport(reportID)
			if err != nil {
				log.Println("@GetMedicalRecords->GetDiagnosticReport:", err)
				continue
			}

			diagnostic := &models.UploadedDiagnosticRes{
				CollectedAt:     report.CollectedAt,
				CollectedDate:   report.CollectedDate.Format("02 Jan 2006 15:04:05"),
				Comments:        report.Comments,
				DiagnosticLabID: report.DiagnosticLabId,
				LabName:         "", // can fetch if needed
				ReportDate:      report.ReportDate.Format("02 Jan 2006 15:04:05"),
				ReportName:      report.ReportName,
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
						ResultDate:    result.ResultDate.Format("02 Jan 2006 15:04:05"),
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
