package repository

import (
	"biostat/constant"
	"biostat/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type SymptomRepository interface {
	GetAllSymptom(limit int, offset int) ([]models.Symptom, int64, error)
	AddDiseaseSymptom(symptom *models.Symptom) error
	UpdateSymptom(symptom *models.Symptom, authUserId string) error
	DeleteSymptom(symptomId uint64, authUserId string) error
	GetSymptomAuditRecord(symptomId uint64, symptomAuditId uint64) ([]models.SymptomAudit, error)
	GetAllSymptomAuditRecord(page, limit int) ([]models.SymptomAudit, int64, error)
}

type SymptomRepositoryImpl struct {
	db *gorm.DB
}

func NewSymptomRepository(db *gorm.DB) SymptomRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &SymptomRepositoryImpl{db: db}
}

// GetAllSymptom implements SymptomRepository.
func (s *SymptomRepositoryImpl) GetAllSymptom(limit int, offset int) ([]models.Symptom, int64, error) {

	var symptom []models.Symptom
	var totalRecords int64

	query := s.db.Model(&models.Symptom{})
	query.Count(&totalRecords)

	err := query.Limit(limit).Offset(offset).Find(&symptom).Error
	if err != nil {
		return nil, 0, err
	}
	return symptom, totalRecords, nil
}

// AddDiseaseSymptom implements SymptomRepository.
func (s *SymptomRepositoryImpl) AddDiseaseSymptom(symptom *models.Symptom) error {
	return s.db.Create(symptom).Error
}

func (repo *SymptomRepositoryImpl) GetSymptomById(symptomId uint64) (*models.Symptom, error) {
	var symptom models.Symptom
	if err := repo.db.Where("symptom_id = ?", symptomId).First(&symptom).Error; err != nil {
		return nil, err
	}
	return &symptom, nil
}

func (repo *SymptomRepositoryImpl) UpdateSymptom(symptom *models.Symptom, updatedBy string) error {
	existing, err := repo.GetSymptomById(symptom.SymptomId)
	if err != nil {
		return err
	}

	symptom.UpdatedAt = time.Now()
	result := repo.db.Model(&models.Symptom{}).Where("symptom_id = ?", symptom.SymptomId).Updates(symptom)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no records updated")
	}

	return repo.SaveSymptomAudit(existing, constant.UPDATE, updatedBy)
}

func (repo *SymptomRepositoryImpl) DeleteSymptom(symptomId uint64, deletedBy string) error {
	symptom, err := repo.GetSymptomById(symptomId)
	if err != nil {
		return err
	}

	err = repo.SaveSymptomAudit(symptom, constant.DELETE, deletedBy)
	if err != nil {
		return err
	}

	return repo.db.Delete(&models.Symptom{}, "symptom_id = ?", symptomId).Error
}

func (repo *SymptomRepositoryImpl) SaveSymptomAudit(symptom *models.Symptom, operationType string, user string) error {
	audit := models.SymptomAudit{
		SymptomId:     symptom.SymptomId,
		SymptomName:   symptom.SymptomName,
		SymptomType:   symptom.SymptomType,
		Commonality:   symptom.Commonality,
		Description:   symptom.Description,
		CreatedAt:     symptom.CreatedAt,
		UpdatedAt:     symptom.UpdatedAt,
		CreatedBy:     symptom.CreatedBy,
		UpdatedBy:     user,
		OperationType: operationType,
	}
	return repo.db.Create(&audit).Error
}

func (repo *SymptomRepositoryImpl) GetAllSymptomAuditRecord(page, limit int) ([]models.SymptomAudit, int64, error) {
	var auditLogs []models.SymptomAudit
	var totalRecords int64

	repo.db.Model(&models.SymptomAudit{}).Count(&totalRecords)

	err := repo.db.
		Limit(limit).
		Offset((page - 1) * limit).
		Order("symptom_audit_id DESC").
		Find(&auditLogs).Error

	return auditLogs, totalRecords, err
}

func (repo *SymptomRepositoryImpl) GetSymptomAuditRecord(symptomId, symptomAuditId uint64) ([]models.SymptomAudit, error) {
	var auditLogs []models.SymptomAudit
	query := repo.db

	if symptomId != 0 {
		query = query.Where("symptom_id = ?", symptomId)
	}
	if symptomAuditId != 0 {
		query = query.Where("symptom_audit_id = ?", symptomAuditId)
	}

	err := query.Order("symptom_audit_id DESC").Find(&auditLogs).Error
	if err != nil {
		return nil, err
	}
	return auditLogs, nil
}
