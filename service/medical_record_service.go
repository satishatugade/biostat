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
	SaveMedicalRecords(data *[]models.TblMedicalRecord, userId uint64) error
	UpdateTblMedicalRecord(data *models.TblMedicalRecord, updatedBy string) (*models.TblMedicalRecord, error)
	GetSingleTblMedicalRecord(id int64) (*models.TblMedicalRecord, error)
	DeleteTblMedicalRecord(id int, updatedBy string) error
	IsRecordAccessibleToUser(userID uint64, recordID int64) (bool, error)
	GetMedicalRecords(userID uint64, limit, offset int) ([]models.MedicalRecordResponseRes, int64, error)

	ReadMedicalRecord(ResourceId int64, userId, reqUserId uint64) (interface{}, error)
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

type DigitizationPayload struct {
	RecordID   uint64 `json:"record_id"`
	UserID     uint64 `json:"user_id"`
	FilePath   string `json:"file_path"`
	Category   string `json:"category"`
	FileName   string `json:"file_name"`
	AuthUserID string `json:"auth_user_id"`
}

func (s *tblMedicalRecordServiceImpl) CreateTblMedicalRecord(data *models.TblMedicalRecord, createdBy uint64, authUserId string, fileBuf *bytes.Buffer, filename string) (*models.TblMedicalRecord, error) {
	record, err := s.tblMedicalRecordRepo.CreateTblMedicalRecord(data)
	if err != nil {
		return nil, err
	}
	var mappings []models.TblMedicalRecordUserMapping
	mappings = append(mappings, models.TblMedicalRecordUserMapping{
		UserID:   createdBy,
		RecordID: record.RecordId,
	})
	err = s.tblMedicalRecordRepo.CreateMedicalRecordMappings(&mappings)
	if err != nil {
		return nil, err
	}
	if record.RecordCategory == "Test Reports" || record.RecordCategory == "Prescriptions" {
		// Save file temporarily to disk
		tempDir := os.TempDir()
		tempPath := filepath.Join(tempDir, fmt.Sprintf("record_%d_%s", record.RecordId, filename))
		if err := os.WriteFile(tempPath, fileBuf.Bytes(), 0644); err != nil {
			log.Printf("Failed to write temp file for record %d: %v", record.RecordId, err)
			return record, nil
		}
		log.Println("Queue worker starts............")
		// Construct payload
		payload := DigitizationPayload{
			RecordID:   record.RecordId,
			UserID:     createdBy,
			FilePath:   tempPath,
			Category:   record.RecordCategory,
			FileName:   filename,
			AuthUserID: authUserId,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal digitization payload for record %d: %v", record.RecordId, err)
			return record, nil
		}

		// Enqueue task
		task := asynq.NewTask("digitize:record", payloadBytes)
		if _, err := s.taskQueue.Enqueue(task, asynq.MaxRetry(3)); err != nil {
			log.Printf("Failed to enqueue digitization task for record %d: %v", record.RecordId, err)
			return record, nil
		}
		log.Printf("record Id : %d : status : %s ", record.RecordId, "queued")
		// Optionally set Redis status
		s.redisClient.Set(context.Background(), fmt.Sprintf("record_status:%d", record.RecordId), "queued", 10*time.Minute)
	}

	return record, nil
}

// func (s *tblMedicalRecordServiceImpl) CreateTblMedicalRecord(data *models.TblMedicalRecord, createdBy uint64, authUserId string, fileBuf *bytes.Buffer, filename string) (*models.TblMedicalRecord, error) {
// 	record, err := s.tblMedicalRecordRepo.CreateTblMedicalRecord(data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var mappings []models.TblMedicalRecordUserMapping
// 	mappings = append(mappings, models.TblMedicalRecordUserMapping{
// 		UserID:   createdBy,
// 		RecordID: record.RecordId,
// 	})
// 	err = s.tblMedicalRecordRepo.CreateMedicalRecordMappings(&mappings)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if record.RecordCategory == "Test Reports" {
// 		var imageCopy bytes.Buffer

// 		if _, err := io.Copy(&imageCopy, bytes.NewReader(fileBuf.Bytes())); err != nil {
// 			log.Printf("Failed to copy image buffer for async call (record %d): %v", record.RecordId, err)
// 			return record, nil
// 		}

// 		go func(imageBuf bytes.Buffer, recordID, userID uint64) {
// 			reportData, err := s.apiService.CallGeminiService(&imageBuf, filename)
// 			if err != nil {
// 				log.Printf("Gemini Service Error for record %d: %v", recordID, err)
// 				return
// 			}

// 			relatives, err := s.patientService.GetRelativeList(&userID)
// 			if err != nil {
// 				log.Printf("Relative not added or not found  %d: %v", recordID, err)

// 			}
// 			matchedUserID := MatchPatientNameWithRelative(relatives, reportData.ReportDetails.PatientName, userID)
// 			fmt.Println("matchedUserID ", matchedUserID)
// 			reportData.ReportDetails.IsDigital = true
// 			message, err := s.diagnosticService.DigitizeDiagnosticReport(reportData, matchedUserID, &record.RecordId)
// 			if err != nil {
// 				log.Printf("Digitize error for record %d: %v", recordID, err)
// 				return
// 			}
// 			payload := models.TblMedicalRecord{
// 				RecordId:     record.RecordId,
// 				DigitizeFlag: 1,
// 			}
// 			_, err1 := s.tblMedicalRecordRepo.UpdateTblMedicalRecord(&payload, authUserId)
// 			if err1 != nil {
// 				log.Println("Failed to update record : ", err1)
// 				return
// 			}
// 			if err := s.diagnosticService.NotifyAbnormalResult(matchedUserID); err != nil {
// 				log.Printf("NotifyAbnormalResult error: %v", err)
// 			}
// 			log.Printf("Digitization result for record %d: %v", recordID, message)
// 		}(imageCopy, uint64(record.RecordId), createdBy)
// 	} else if record.RecordCategory == "Prescriptions" {
// 		var imageCopy bytes.Buffer

// 		if _, err := io.Copy(&imageCopy, bytes.NewReader(fileBuf.Bytes())); err != nil {
// 			log.Printf("Failed to copy image buffer for prescription async call (record %d): %v", record.RecordId, err)
// 			return record, nil
// 		}

// 		go func(imageBuf bytes.Buffer, recordID, userID uint64) {
// 			prescriptionData, err := s.apiService.CallPrescriptionDigitizeAPI(&imageBuf, filename)
// 			if err != nil {
// 				log.Printf("Prescription Digitization API error for record %d: %+v", recordID, err)
// 				return
// 			}
// 			prescriptionData.PatientId = userID
// 			prescriptionData.RecordId = recordID
// 			prescriptionData.IsDigital = true
// 			err1 := s.patientService.AddPatientPrescription(authUserId, &prescriptionData)
// 			if err1 != nil {
// 				log.Printf("SavePrescriptionData error for record %d: %v", recordID, err1)
// 				return
// 			}
// 			payload := models.TblMedicalRecord{
// 				RecordId:     record.RecordId,
// 				DigitizeFlag: 1,
// 			}
// 			_, err2 := s.tblMedicalRecordRepo.UpdateTblMedicalRecord(&payload, authUserId)
// 			if err2 != nil {
// 				log.Println("Failed to update record : ", err2)
// 				return
// 			}
// 			log.Printf("Prescription digitization result for record %d: %v", recordID, prescriptionData)
// 		}(imageCopy, uint64(record.RecordId), createdBy)
// 	}

// 	return record, nil
// }

func MatchPatientNameWithRelative(relatives []models.PatientRelative, patientName string, fallbackUserID uint64) uint64 {
	normalizedPatientName := strings.TrimSpace(strings.ToLower(patientName))

	highestScore := 0
	bestMatchID := fallbackUserID

	for _, relative := range relatives {
		fullName := strings.TrimSpace(strings.ToLower(relative.FirstName + " " + relative.LastName))

		// Soundex for phonetic similarity
		soundexPatient := smetrics.Soundex(normalizedPatientName)
		soundexRelative := smetrics.Soundex(fullName)

		// Levenshtein-based similarity score
		levDistance := smetrics.WagnerFischer(normalizedPatientName, fullName, 1, 1, 2)
		maxLen := max(len(normalizedPatientName), len(fullName))
		similarity := 100 - (levDistance * 100 / maxLen)

		// Increase score if soundex matches
		score := similarity
		if soundexPatient == soundexRelative {
			score += 20
		}

		log.Printf("Matching '%s' <-> '%s' | Similarity: %d%% | Score: %d", normalizedPatientName, fullName, similarity, score)

		if score > highestScore && score >= 40 {
			highestScore = score
			bestMatchID = relative.RelativeId
		}
	}

	if bestMatchID != fallbackUserID {
		log.Printf("Best fuzzy match found. Relative ID: %d", bestMatchID)
	} else {
		log.Printf("No match found. Falling back to default user ID: %d", fallbackUserID)
	}

	return bestMatchID
}

func (s *tblMedicalRecordServiceImpl) SaveMedicalRecords(records *[]models.TblMedicalRecord, userId uint64) error {
	var uniqueRecords []models.TblMedicalRecord

	for _, record := range *records {
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

	err := s.tblMedicalRecordRepo.CreateMultipleTblMedicalRecords(&uniqueRecords)
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

func (s *tblMedicalRecordServiceImpl) UpdateTblMedicalRecord(data *models.TblMedicalRecord, updatedBy string) (*models.TblMedicalRecord, error) {
	return s.tblMedicalRecordRepo.UpdateTblMedicalRecord(data, updatedBy)
}

func (s *tblMedicalRecordServiceImpl) GetSingleTblMedicalRecord(id int64) (*models.TblMedicalRecord, error) {
	return s.tblMedicalRecordRepo.GetSingleTblMedicalRecord(id)
}

func (s *tblMedicalRecordServiceImpl) DeleteTblMedicalRecord(id int, updatedBy string) error {
	return s.tblMedicalRecordRepo.DeleteTblMedicalRecordWithMappings(id, updatedBy)
}

func (s *tblMedicalRecordServiceImpl) IsRecordAccessibleToUser(userID uint64, recordID int64) (bool, error) {
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

func (s *tblMedicalRecordServiceImpl) ReadMedicalRecord(ResourceId int64, userId, reqUserId uint64) (interface{}, error) {
	isAccessible, err := s.IsRecordAccessibleToUser(userId, ResourceId)
	if err != nil {
		return nil, err
	}
	if !isAccessible {
		return nil, errors.New("you do not have access to this resource")
	}

	record, err := s.GetSingleTblMedicalRecord(ResourceId)
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
