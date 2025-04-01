package service

import (
	"biostat/models"
	"biostat/repository"
)

type TblMedicalRecordService interface {
	GetAllTblMedicalRecords(limit int, offset int) ([]models.TblMedicalRecord, int64, error)
	GetUserMedicalRecords(userID int64) ([]models.TblMedicalRecord, error)
	CreateTblMedicalRecord(data *models.TblMedicalRecord, createdBy int64) (*models.TblMedicalRecord, error)
	SaveMedicalRecords(data *[]models.TblMedicalRecord, userId int64) error
	UpdateTblMedicalRecord(data *models.TblMedicalRecord, updatedBy string) (*models.TblMedicalRecord, error)
	GetSingleTblMedicalRecord(id int) (*models.TblMedicalRecord, error)
	DeleteTblMedicalRecord(id int, updatedBy string) error
}

type tblMedicalRecordServiceImpl struct {
	tblMedicalRecordRepo repository.TblMedicalRecordRepository
}

func NewTblMedicalRecordService(repo repository.TblMedicalRecordRepository) TblMedicalRecordService {
	return &tblMedicalRecordServiceImpl{tblMedicalRecordRepo: repo}
}

func (s *tblMedicalRecordServiceImpl) GetUserMedicalRecords(userID int64) ([]models.TblMedicalRecord, error) {
	return s.tblMedicalRecordRepo.GetMedicalRecordsByUserID(userID)
}

func (s *tblMedicalRecordServiceImpl) GetAllTblMedicalRecords(limit int, offset int) ([]models.TblMedicalRecord, int64, error) {
	return s.tblMedicalRecordRepo.GetAllTblMedicalRecords(limit, offset)
}

func (s *tblMedicalRecordServiceImpl) CreateTblMedicalRecord(data *models.TblMedicalRecord, createdBy int64) (*models.TblMedicalRecord, error) {
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
	return record, nil

}

func (s *tblMedicalRecordServiceImpl) SaveMedicalRecords(records *[]models.TblMedicalRecord, userId int64) error {
	err := s.tblMedicalRecordRepo.CreateMultipleTblMedicalRecords(records)
	if err != nil {
		return err
	}
	var mappings []models.TblMedicalRecordUserMapping
	for _, record := range *records {
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

func (s *tblMedicalRecordServiceImpl) GetSingleTblMedicalRecord(id int) (*models.TblMedicalRecord, error) {
	return s.tblMedicalRecordRepo.GetSingleTblMedicalRecord(id)
}

func (s *tblMedicalRecordServiceImpl) DeleteTblMedicalRecord(id int, updatedBy string) error {
	return s.tblMedicalRecordRepo.DeleteTblMedicalRecordWithMappings(id, updatedBy)
}
