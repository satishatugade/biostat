package repository

import (
	"biostat/models"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TblMedicalRecordRepository interface {
	GetAllTblMedicalRecords(limit int, offset int) ([]models.TblMedicalRecord, int64, error)
	GetMedicalRecordsByUserID(userID uint64, recordIdsMap map[uint64]uint64) ([]models.TblMedicalRecord, error)
	CreateTblMedicalRecord(data *models.TblMedicalRecord) (*models.TblMedicalRecord, error)
	CreateMultipleTblMedicalRecords(data *[]models.TblMedicalRecord) error
	UpdateTblMedicalRecord(data *models.TblMedicalRecord, updatedBy string) (*models.TblMedicalRecord, error)
	GetSingleTblMedicalRecord(id int64) (*models.TblMedicalRecord, error)
	DeleteTblMedicalRecord(id int, updatedBy string) error
	IsRecordBelongsToUser(userID uint64, recordID int64) (bool, error)
	ExistsRecordForUser(userId uint64, source, url string) (bool, error)

	CreateMedicalRecordMappings(mappings *[]models.TblMedicalRecordUserMapping) error
	GetMedicalRecordMappings(recordID int64) (*models.TblMedicalRecordUserMapping, error)
	DeleteMecationRecordMappings(id int) error

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

func (r *tblMedicalRecordRepositoryImpl) GetMedicalRecordsByUserID(userID uint64, recordIdMap map[uint64]uint64) ([]models.TblMedicalRecord, error) {
	var records []models.TblMedicalRecord

	query := r.db.Table("tbl_medical_record").
		Select("tbl_medical_record.*").
		Joins("INNER JOIN tbl_medical_record_user_mapping ON tbl_medical_record.record_id = tbl_medical_record_user_mapping.record_id").
		Where("tbl_medical_record_user_mapping.user_id = ? and is_deleted=0", userID)

	if recordIdMap != nil && len(recordIdMap) > 0 {
		var recordIds []uint64
		for _, id := range recordIdMap {
			recordIds = append(recordIds, id)
		}
		query = query.Where("tbl_medical_record.record_id IN ?", recordIds)
	}
	err := query.Order("tbl_medical_record.updated_at DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}
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
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	updateFields := map[string]interface{}{}
	if data.RecordName != "" {
		updateFields["record_name"] = data.RecordName
	}
	if data.RecordSize != 0 {
		updateFields["record_size"] = data.RecordSize
	}
	if data.FileType != "" {
		updateFields["file_type"] = data.FileType
	}
	if data.UploadSource != "" {
		updateFields["upload_source"] = data.UploadSource
	}
	if data.UploadDestination != "" {
		updateFields["upload_destination"] = data.UploadDestination
	}
	if data.SourceAccount != "" {
		updateFields["source_account"] = data.SourceAccount
	}
	if data.RecordCategory != "" {
		updateFields["record_category"] = data.RecordCategory
	}
	if data.Description != "" {
		updateFields["description"] = data.Description
	}
	if data.RecordUrl != "" {
		updateFields["record_url"] = data.RecordUrl
	}
	if data.FileData != nil {
		updateFields["file_data"] = data.FileData
	}
	if data.DigitizeFlag > 0 {
		updateFields["digitize_flag"] = data.DigitizeFlag
	}
	if len(data.Metadata) != 0 {
		updateFields["metadata"] = data.Metadata
	}
	updateFields["is_verified"] = data.IsVerified
	updateFields["is_deleted"] = data.IsDeleted
	updateFields["updated_at"] = time.Now()

	err := tx.Model(&models.TblMedicalRecord{}).
		Where("record_id = ?", data.RecordId).
		Updates(updateFields).Error

	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return data, nil
}

func (r *tblMedicalRecordRepositoryImpl) GetSingleTblMedicalRecord(id int64) (*models.TblMedicalRecord, error) {
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

func (r *tblMedicalRecordRepositoryImpl) GetMedicalRecordMappings(recordID int64) (*models.TblMedicalRecordUserMapping, error) {
	var mapping models.TblMedicalRecordUserMapping
	err := r.db.Where("record_id=?", recordID).Find(&mapping).Error
	return &mapping, err
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

func (r *tblMedicalRecordRepositoryImpl) ExistsRecordForUser(userId uint64, source, url string) (bool, error) {
	var count int64
	err := r.db.
		Table("tbl_medical_record").
		Joins("INNER JOIN tbl_medical_record_user_mapping ON tbl_medical_record.record_id = tbl_medical_record_user_mapping.record_id").
		Where("tbl_medical_record_user_mapping.user_id = ? AND tbl_medical_record.upload_source = ? AND tbl_medical_record.record_url = ?", userId, source, url).
		Count(&count).Error

	return count > 0, err
}

func (r *tblMedicalRecordRepositoryImpl) IsRecordBelongsToUser(userID uint64, recordID int64) (bool, error) {
	var mapping models.TblMedicalRecordUserMapping
	err := r.db.Where("user_id = ? AND record_id = ?", userID, recordID).First(&mapping).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
