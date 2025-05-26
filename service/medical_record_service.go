package service

import (
	"biostat/models"
	"biostat/repository"
	"bytes"
	"fmt"
	"io"
	"log"
)

type TblMedicalRecordService interface {
	GetAllTblMedicalRecords(limit int, offset int) ([]models.TblMedicalRecord, int64, error)
	GetUserMedicalRecords(userID int64) ([]models.TblMedicalRecord, error)
	CreateTblMedicalRecord(data *models.TblMedicalRecord, createdBy uint64, authUserId string, file *bytes.Buffer, filename string) (*models.TblMedicalRecord, error)
	SaveMedicalRecords(data *[]models.TblMedicalRecord, userId uint64) error
	UpdateTblMedicalRecord(data *models.TblMedicalRecord, updatedBy string) (*models.TblMedicalRecord, error)
	GetSingleTblMedicalRecord(id int64) (*models.TblMedicalRecord, error)
	DeleteTblMedicalRecord(id int, updatedBy string) error
	IsRecordBelongsToUser(userID uint64, recordID int64) (bool, error)
}

type tblMedicalRecordServiceImpl struct {
	tblMedicalRecordRepo repository.TblMedicalRecordRepository
	apiService           ApiService
	diagnosticService    DiagnosticService
	patientService       PatientService
}

func NewTblMedicalRecordService(repo repository.TblMedicalRecordRepository, apiService ApiService, diagnosticService DiagnosticService, patientService PatientService) TblMedicalRecordService {
	return &tblMedicalRecordServiceImpl{tblMedicalRecordRepo: repo, apiService: apiService, diagnosticService: diagnosticService, patientService: patientService}
}

func (s *tblMedicalRecordServiceImpl) GetUserMedicalRecords(userID int64) ([]models.TblMedicalRecord, error) {
	return s.tblMedicalRecordRepo.GetMedicalRecordsByUserID(userID)
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
		RecordID: int64(record.RecordId),
	})
	err = s.tblMedicalRecordRepo.CreateMedicalRecordMappings(&mappings)
	if err != nil {
		return nil, err
	}
	fmt.Println("RecordCategory ", record.RecordCategory)

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

			message, err := s.diagnosticService.DigitizeDiagnosticReport(reportData, userID)
			if err != nil {
				log.Printf("Digitize error for record %d: %v", recordID, err)
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
			err1 := s.patientService.AddPatientPrescription(authUserId, &prescriptionData)
			if err1 != nil {
				log.Printf("SavePrescriptionData error for record %d: %v", recordID, err1)
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
			RecordID: int64(record.RecordId),
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

func (s *tblMedicalRecordServiceImpl) IsRecordBelongsToUser(userID uint64, recordID int64) (bool, error) {
	return s.tblMedicalRecordRepo.IsRecordBelongsToUser(userID, recordID)
}
