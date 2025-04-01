package repository

import (
	"biostat/models"
	"fmt"

	"gorm.io/gorm"
)

type TblMedicalRecordRepository interface {
	GetAllTblMedicalRecords(limit int, offset int) ([]models.TblMedicalRecord, int64, error)
	GetMedicalRecordsByUserID(userID int64) ([]models.TblMedicalRecord, error)
	CreateTblMedicalRecord(data *models.TblMedicalRecord) (*models.TblMedicalRecord, error)
	CreateMultipleTblMedicalRecords(data *[]models.TblMedicalRecord) error
	UpdateTblMedicalRecord(data *models.TblMedicalRecord, updatedBy string) (*models.TblMedicalRecord, error)
	GetSingleTblMedicalRecord(id int) (*models.TblMedicalRecord, error)
	DeleteTblMedicalRecord(id int, updatedBy string) error

	CreateMedicalRecordMappings(mappings *[]models.TblMedicalRecordUserMapping) error
	DeleteMecationRecordMappings(id int,) error

	DeleteTblMedicalRecordWithMappings(id int, user_id string) error
}

type tblMedicalRecordRepositoryImpl struct {
	db *gorm.DB
}

func NewTblMedicalRecordRepository(db *gorm.DB) TblMedicalRecordRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &tblMedicalRecordRepositoryImpl{db: db}
}

func (r *tblMedicalRecordRepositoryImpl) GetMedicalRecordsByUserID(userID int64) ([]models.TblMedicalRecord, error) {
	var records []models.TblMedicalRecord

	err := r.db.Table("tbl_medical_record").
		Select("tbl_medical_record.*").
		Joins("INNER JOIN tbl_medical_record_user_mapping ON tbl_medical_record.record_id = tbl_medical_record_user_mapping.record_id").
		Where("tbl_medical_record_user_mapping.user_id = ? and is_active=true", userID).
		Find(&records).Error

	if err != nil {
		return nil, err
	}
	return records, nil
}

func (r *tblMedicalRecordRepositoryImpl) GetAllTblMedicalRecords(limit int, offset int) ([]models.TblMedicalRecord, int64, error) {
	var objs []models.TblMedicalRecord
	var totalRecords int64
	err := r.db.Model(&models.TblMedicalRecord{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}
	err = r.db.Limit(limit).Offset(offset).Find(&objs).Error
	if err != nil {
		return nil, 0, err
	}
	return objs, totalRecords, nil
}

func (r *tblMedicalRecordRepositoryImpl) CreateTblMedicalRecord(data *models.TblMedicalRecord) (*models.TblMedicalRecord, error) {
	err := r.db.Create(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *tblMedicalRecordRepositoryImpl) CreateMultipleTblMedicalRecords(records *[]models.TblMedicalRecord) error {
	return r.db.Create(records).Error
}

func (r *tblMedicalRecordRepositoryImpl) UpdateTblMedicalRecord(data *models.TblMedicalRecord, updatedBy string) (*models.TblMedicalRecord, error) {
	err := r.db.Model(&models.TblMedicalRecord{}).Where("record_id = ?", data.RecordId).Updates(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *tblMedicalRecordRepositoryImpl) GetSingleTblMedicalRecord(id int) (*models.TblMedicalRecord, error) {
	var obj models.TblMedicalRecord
	err := r.db.First(&obj, id).Error
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func (r *tblMedicalRecordRepositoryImpl) DeleteTblMedicalRecord(id int, updatedBy string) error {
	return r.db.Where("record_id = ?", id).Delete(&models.TblMedicalRecord{}).Error
}

func (r *tblMedicalRecordRepositoryImpl) CreateMedicalRecordMappings(mappings *[]models.TblMedicalRecordUserMapping) error {
	return r.db.Create(mappings).Error
}

func (r *tblMedicalRecordRepositoryImpl) DeleteMecationRecordMappings(id int) error {
	return r.db.Where("record_id = ?", id).Delete(&models.TblMedicalRecordUserMapping{}).Error
}

func (r *tblMedicalRecordRepositoryImpl) DeleteTblMedicalRecordWithMappings(id int, updatedBy string) error {
	tx := r.db.Begin()

	if tx.Error != nil {
		return tx.Error
	}

	result := tx.Where("record_id = ?", id).Delete(&models.TblMedicalRecord{})
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("record with id %d not found", id)
	}

	result = tx.Where("record_id = ?", id).Delete(&models.TblMedicalRecordUserMapping{})
	if result.Error != nil {
		tx.Rollback() 
		return result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("no mappings found for record with id %d", id)
	}

	return tx.Commit().Error
}
