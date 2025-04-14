package repository

import (
	"biostat/constant"
	"biostat/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type CauseRepository interface {
	GetAllCauses(limit int, offset int) ([]models.Cause, int64, error)
	AddDiseaseCause(cause *models.Cause) error
	UpdateCause(cause *models.Cause, authUserId string) error
	GetDiseaseCauseById(causeId uint64) (*models.Cause, error)
	DeleteCause(causeId uint64, authUserId string) error
	GetCauseAuditRecord(causeId uint64, causeAuditId uint64) ([]models.CauseAudit, error)
	GetAllCauseAuditRecord(page, limit int) ([]models.CauseAudit, int64, error)
	AddDiseaseCauseMapping(DCMapping *models.DiseaseCauseMapping) error
}

type CauseRepositoryImpl struct {
	db *gorm.DB
}

func NewCauseRepository(db *gorm.DB) CauseRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &CauseRepositoryImpl{db: db}
}

// GetAllCauses implements CauseRepository.
func (c *CauseRepositoryImpl) GetAllCauses(limit int, offset int) ([]models.Cause, int64, error) {

	var causes []models.Cause
	var totalRecords int64

	query := c.db.Model(&models.Cause{})
	query.Count(&totalRecords)

	err := query.Limit(limit).Offset(offset).Find(&causes).Error
	if err != nil {
		return nil, 0, err
	}
	return causes, totalRecords, nil
}

func (c *CauseRepositoryImpl) AddDiseaseCause(cause *models.Cause) error {
	return c.db.Create(cause).Error
}

func (repo *CauseRepositoryImpl) GetDiseaseCauseById(causeId uint64) (*models.Cause, error) {
	var cause models.Cause
	if err := repo.db.Where("cause_id = ?", causeId).First(&cause).Error; err != nil {
		return nil, err
	}
	return &cause, nil
}

func (repo *CauseRepositoryImpl) UpdateCause(updatedCause *models.Cause, authUserId string) error {
	existingCause, err := repo.GetDiseaseCauseById(updatedCause.CauseId)
	if err != nil {
		return err
	}

	updatedCause.UpdatedAt = time.Now()

	result := repo.db.Model(&models.Cause{}).Where("cause_id = ?", updatedCause.CauseId).Updates(updatedCause)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no records updated")
	}

	if err := repo.SaveCauseAudit(existingCause, constant.UPDATE, authUserId); err != nil {
		return err
	}
	return nil
}

func (repo *CauseRepositoryImpl) SaveCauseAudit(existingCause *models.Cause, operationType string, updatedBy string) error {
	auditLog := models.CauseAudit{
		CauseId:       existingCause.CauseId,
		CauseName:     existingCause.CauseName,
		CauseType:     existingCause.CauseType,
		Description:   existingCause.Description,
		OperationType: operationType,
		CreatedAt:     existingCause.CreatedAt,
		UpdatedAt:     time.Now(),
		CreatedBy:     existingCause.CreatedBy,
		UpdatedBy:     updatedBy,
	}

	return repo.db.Create(&auditLog).Error
}

func (repo *CauseRepositoryImpl) DeleteCause(causeId uint64, deletedBy string) error {
	cause, err := repo.GetDiseaseCauseById(causeId)
	if err != nil {
		return err
	}
	if err := repo.SaveCauseAudit(cause, constant.DELETE, deletedBy); err != nil {
		return err
	}
	result := repo.db.Model(&models.Cause{}).Where("cause_id = ?", causeId).Update("is_deleted", 1)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records deleted")
	}

	return nil
}

func (repo *CauseRepositoryImpl) GetAllCauseAuditRecord(page, limit int) ([]models.CauseAudit, int64, error) {
	var auditLogs []models.CauseAudit
	var totalRecords int64

	// Count total records
	repo.db.Model(&models.CauseAudit{}).Count(&totalRecords)

	// Fetch paginated data
	err := repo.db.
		Limit(limit).
		Offset((page - 1) * limit).
		Order("cause_audit_id DESC").
		Find(&auditLogs).Error

	return auditLogs, totalRecords, err
}

func (repo *CauseRepositoryImpl) GetCauseAuditRecord(causeId, causeAuditId uint64) ([]models.CauseAudit, error) {
	var auditLogs []models.CauseAudit
	query := repo.db

	if causeId != 0 {
		query = query.Where("cause_id = ?", causeId)
	}
	if causeAuditId != 0 {
		query = query.Where("cause_audit_id = ?", causeAuditId)
	}

	err := query.Order("cause_audit_id DESC").Find(&auditLogs).Error
	if err != nil {
		return nil, err
	}
	return auditLogs, nil
}

func (r *CauseRepositoryImpl) AddDiseaseCauseMapping(mapping *models.DiseaseCauseMapping) error {
	return r.db.Create(mapping).Error
}
