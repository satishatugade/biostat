package service

import (
	"biostat/models"
	"biostat/repository"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type TblMedicalRecordService interface {
	GetAllTblMedicalRecords(limit int, offset int) ([]models.TblMedicalRecord, int64, error)
	GetUserMedicalRecords(userID uint64) ([]models.TblMedicalRecord, error)
	CreateTblMedicalRecord(data *models.TblMedicalRecord, createdBy uint64, authUserId string, file *bytes.Buffer, filename string) (*models.TblMedicalRecord, error)
	SaveMedicalRecords(data *[]models.TblMedicalRecord, userId uint64) error
	UpdateTblMedicalRecord(data *models.TblMedicalRecord, updatedBy string) (*models.TblMedicalRecord, error)
	GetSingleTblMedicalRecord(id int64) (*models.TblMedicalRecord, error)
	DeleteTblMedicalRecord(id int, updatedBy string) error
	IsRecordAccessibleToUser(userID uint64, recordID int64) (bool, error)

	ReadMedicalRecord(ResourceId int64, userId, reqUserId uint64) (interface{}, error)
}

type tblMedicalRecordServiceImpl struct {
	tblMedicalRecordRepo repository.TblMedicalRecordRepository
	apiService           ApiService
	diagnosticService    DiagnosticService
	medicalRecordService TblMedicalRecordService
	patientService       PatientService
	userService          UserService
}

func NewTblMedicalRecordService(repo repository.TblMedicalRecordRepository, apiService ApiService, diagnosticService DiagnosticService, patientService PatientService, userService UserService) TblMedicalRecordService {
	return &tblMedicalRecordServiceImpl{tblMedicalRecordRepo: repo, apiService: apiService, diagnosticService: diagnosticService, patientService: patientService, userService: userService}
}

func (s *tblMedicalRecordServiceImpl) GetUserMedicalRecords(userID uint64) ([]models.TblMedicalRecord, error) {
	return s.tblMedicalRecordRepo.GetMedicalRecordsByUserID(userID, nil)
}

func (s *tblMedicalRecordServiceImpl) GetAllTblMedicalRecords(limit int, offset int) ([]models.TblMedicalRecord, int64, error) {
	return s.tblMedicalRecordRepo.GetAllTblMedicalRecords(limit, offset)
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
	if record.RecordCategory == "Test Reports" {
		var imageCopy bytes.Buffer

		if _, err := io.Copy(&imageCopy, bytes.NewReader(fileBuf.Bytes())); err != nil {
			log.Printf("Failed to copy image buffer for async call (record %d): %v", record.RecordId, err)
			return record, nil
		}

		go func(imageBuf bytes.Buffer, recordID, userID uint64) {
			reportData, err := s.apiService.CallGeminiService(&imageBuf, filename)
			if err != nil {
				log.Printf("Gemini Service Error for record %d: %v", recordID, err)
				return
			}

			message, err := s.diagnosticService.DigitizeDiagnosticReport(reportData, userID, &record.RecordId)
			if err != nil {
				log.Printf("Digitize error for record %d: %v", recordID, err)
				return
			}
			payload := models.TblMedicalRecord{
				RecordId:     record.RecordId,
				DigitizeFlag: 1,
			}
			_, err1 := s.tblMedicalRecordRepo.UpdateTblMedicalRecord(&payload, authUserId)
			if err1 != nil {
				log.Println("Failed to update record : ", err1)
				return
			}
			if err := s.diagnosticService.NotifyAbnormalResult(userID); err != nil {
				log.Printf("NotifyAbnormalResult error: %v", err)
			}
			log.Printf("Digitization result for record %d: %v", recordID, message)
		}(imageCopy, uint64(record.RecordId), createdBy)
	} else if record.RecordCategory == "Prescriptions" {
		var imageCopy bytes.Buffer

		if _, err := io.Copy(&imageCopy, bytes.NewReader(fileBuf.Bytes())); err != nil {
			log.Printf("Failed to copy image buffer for prescription async call (record %d): %v", record.RecordId, err)
			return record, nil
		}

		go func(imageBuf bytes.Buffer, recordID, userID uint64) {
			prescriptionData, err := s.apiService.CallPrescriptionDigitizeAPI(&imageBuf, filename)
			if err != nil {
				log.Printf("Prescription Digitization API error for record %d: %+v", recordID, err)
				return
			}
			prescriptionData.PatientId = userID
			prescriptionData.RecordId = recordID
			prescriptionData.IsDigital = true
			err1 := s.patientService.AddPatientPrescription(authUserId, &prescriptionData)
			if err1 != nil {
				log.Printf("SavePrescriptionData error for record %d: %v", recordID, err1)
				return
			}
			payload := models.TblMedicalRecord{
				RecordId:     record.RecordId,
				DigitizeFlag: 1,
			}
			_, err2 := s.tblMedicalRecordRepo.UpdateTblMedicalRecord(&payload, authUserId)
			if err2 != nil {
				log.Println("Failed to update record : ", err2)
				return
			}
			log.Printf("Prescription digitization result for record %d: %v", recordID, prescriptionData)
		}(imageCopy, uint64(record.RecordId), createdBy)
	}

	return record, nil
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
