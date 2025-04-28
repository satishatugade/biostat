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
	AddDiseaseSymptom(symptom *models.Symptom) (*models.Symptom, error)
	UpdateSymptom(symptom *models.Symptom, authUserId string) (*models.Symptom, error)
	DeleteSymptom(symptomId uint64, authUserId string) error
	GetSymptomAuditRecord(symptomId uint64, symptomAuditId uint64) ([]models.SymptomAudit, error)
	GetAllSymptomAuditRecord(limit, offset int) ([]models.SymptomAudit, int64, error)

	AddDiseaseSymptomMapping(mapping *models.DiseaseSymptomMapping) error

	GetAllSymptomTypes(limit int, offset int, isDeleted int) ([]models.SymptomTypeMaster, int64, error)
	AddSymptomType(symptomType *models.SymptomTypeMaster) (*models.SymptomTypeMaster, error)
	UpdateSymptomType(symptomType *models.SymptomTypeMaster, userId string) (*models.SymptomTypeMaster, error)
	DeleteSymptomType(symptomTypeId uint64, userId string) error

	GetAllSymptomTypeAuditRecord(limit, offset int) ([]models.SymptomTypeAudit, int64, error)
	GetSymptomTypeAuditRecord(symptomTypeId, symptomTypeAuditId uint64) ([]models.SymptomTypeAudit, error)
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

	query := s.db.Model(&models.Symptom{}).Preload("SymptomType").Order("symptom_id DESC")
	query.Count(&totalRecords)

	err := query.Limit(limit).Offset(offset).Find(&symptom).Error
	if err != nil {
		return nil, 0, err
	}
	return symptom, totalRecords, nil
}

func (s *SymptomRepositoryImpl) AddDiseaseSymptom(symptom *models.Symptom) (*models.Symptom, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	if err := tx.Create(symptom).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if len(symptom.SymptomTypeId) > 0 {
		var mappings []models.SymptomTypeMapping
		for _, symptomTypeId := range symptom.SymptomTypeId {
			mapping := models.SymptomTypeMapping{
				SymptomId:     symptom.SymptomId,
				SymptomTypeId: symptomTypeId,
				CreatedBy:     symptom.CreatedBy,
			}
			mappings = append(mappings, mapping)
		}

		if err := tx.Create(&mappings).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	var insertedSymptom models.Symptom
	if err := tx.Preload("SymptomType").First(&insertedSymptom, symptom.SymptomId).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	return &insertedSymptom, tx.Commit().Error
}

func (repo *SymptomRepositoryImpl) GetSymptomById(symptomId uint64) (*models.Symptom, error) {
	var symptom models.Symptom
	if err := repo.db.Where("symptom_id = ?", symptomId).First(&symptom).Error; err != nil {
		return nil, err
	}
	return &symptom, nil
}

func (repo *SymptomRepositoryImpl) UpdateSymptom(updatedSymptom *models.Symptom, updatedBy string) (*models.Symptom, error) {
	tx := repo.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	existingSymptom, err := repo.GetSymptomById(updatedSymptom.SymptomId)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	updatedSymptom.UpdatedAt = time.Now()

	result := tx.Model(&models.Symptom{}).Where("symptom_id = ?", updatedSymptom.SymptomId).Updates(updatedSymptom)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, errors.New("no records updated")
	}

	existingTypeIds := make(map[uint64]bool)

	var existingMappings []models.SymptomTypeMapping
	if err := tx.Where("symptom_id = ?", updatedSymptom.SymptomId).Find(&existingMappings).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, mapping := range existingMappings {
		existingTypeIds[mapping.SymptomTypeId] = true
	}

	var newMappings []models.SymptomTypeMapping
	for _, symptomTypeId := range updatedSymptom.SymptomTypeId {
		if !existingTypeIds[symptomTypeId] {
			newMappings = append(newMappings, models.SymptomTypeMapping{
				SymptomId:     updatedSymptom.SymptomId,
				SymptomTypeId: symptomTypeId,
				CreatedBy:     updatedBy,
			})
		}
	}

	if len(newMappings) > 0 {
		if err := tx.Create(&newMappings).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := repo.SaveSymptomAudit(existingSymptom, constant.UPDATE, updatedBy); err != nil {
		tx.Rollback()
		return nil, err
	}

	var finalUpdatedSymptom models.Symptom
	err = tx.Preload("SymptomType").Where("symptom_id = ?", updatedSymptom.SymptomId).First(&finalUpdatedSymptom).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return &finalUpdatedSymptom, tx.Commit().Error
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
	result := repo.db.Model(&models.Symptom{}).Where("symptom_id = ?", symptomId).Update("is_deleted", 1)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no records updated (cause not found or already deleted)")
	}
	return nil
}

func (repo *SymptomRepositoryImpl) SaveSymptomAudit(symptom *models.Symptom, operationType string, user string) error {
	audit := models.SymptomAudit{
		SymptomId:   symptom.SymptomId,
		SymptomName: symptom.SymptomName,
		// SymptomType:   symptom.SymptomType.SymptomType,
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

func (repo *SymptomRepositoryImpl) GetAllSymptomAuditRecord(limit, offset int) ([]models.SymptomAudit, int64, error) {
	var auditLogs []models.SymptomAudit
	var totalRecords int64

	repo.db.Model(&models.SymptomAudit{}).Count(&totalRecords)

	err := repo.db.
		Limit(limit).
		Offset(offset).
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

func (r *SymptomRepositoryImpl) AddDiseaseSymptomMapping(mapping *models.DiseaseSymptomMapping) error {
	return r.db.Create(mapping).Error
}

func (repo *SymptomRepositoryImpl) GetAllSymptomTypes(limit int, offset int, isDeleted int) ([]models.SymptomTypeMaster, int64, error) {
	var symptomTypes []models.SymptomTypeMaster
	var totalRecords int64
	query := repo.db.Model(&models.SymptomTypeMaster{})
	if isDeleted >= 0 {
		query = query.Where("is_deleted = ?", isDeleted)
	}
	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Limit(limit).Offset(offset).Find(&symptomTypes).Error; err != nil {
		return nil, 0, err
	}
	return symptomTypes, totalRecords, nil
}

func (repo *SymptomRepositoryImpl) AddSymptomType(symptomType *models.SymptomTypeMaster) (*models.SymptomTypeMaster, error) {
	tx := repo.db.Begin()
	if tx.Error != nil {
		return nil, em.ErrorMessage("TransactionError", "SymptomType", tx.Error)
	}

	if err := tx.Create(symptomType).Error; err != nil {
		tx.Rollback()
		return nil, em.ErrorMessage("CreateError", "SymptomType", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, em.ErrorMessage("TransactionError", "SymptomType", err)
	}
	return symptomType, nil
}

func (repo *SymptomRepositoryImpl) UpdateSymptomType(updatedSymptomType *models.SymptomTypeMaster, authUserId string) (*models.SymptomTypeMaster, error) {
	tx := repo.db.Begin()
	if tx.Error != nil {
		return nil, em.ErrorMessage("TransactionError", "SymptomType", nil)
	}

	// Fetch the existing symptom type
	existingSymptomType, err := repo.GetSymptomTypeById(updatedSymptomType.SymptomTypeId, 1)
	if err != nil {
		tx.Rollback()
		return nil, em.ErrorMessage("NotFound", "SymptomType", updatedSymptomType.SymptomTypeId)
	}

	// Update the symptom type
	updatedSymptomType.UpdatedAt = time.Now()
	result := tx.Model(&models.SymptomTypeMaster{}).Where("symptom_type_id = ?", updatedSymptomType.SymptomTypeId).Updates(updatedSymptomType)
	if result.Error != nil {
		tx.Rollback()
		return nil, em.ErrorMessage("UpdateError", "SymptomType", updatedSymptomType.SymptomTypeId)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, em.ErrorMessage("NoRowsAffected", "SymptomType", updatedSymptomType.SymptomTypeId)
	}

	// Save the audit log for update
	if err := repo.SaveSymptomTypeAudit(existingSymptomType, constant.UPDATE, authUserId); err != nil {
		tx.Rollback()
		return nil, em.ErrorMessage("AuditError", "SymptomType", err)
	}

	// Fetch the final updated symptom type
	var finalUpdatedSymptomType models.SymptomTypeMaster
	err = tx.Where("symptom_type_id = ?", updatedSymptomType.SymptomTypeId).First(&finalUpdatedSymptomType).Error
	if err != nil {
		tx.Rollback()
		return nil, em.ErrorMessage("NotFound", "SymptomType", updatedSymptomType.SymptomTypeId)
	}

	return &finalUpdatedSymptomType, tx.Commit().Error
}

func (repo *SymptomRepositoryImpl) DeleteSymptomType(symptomTypeId uint64, deletedBy string) error {
	tx := repo.db.Begin()
	if tx.Error != nil {
		return em.ErrorMessage("TransactionError", "SymptomType", tx.Error)
	}
	symptomType, err := repo.GetSymptomTypeById(symptomTypeId, 0)
	if err != nil {
		tx.Rollback()
		return em.ErrorMessage("NotFound", "SymptomType", symptomTypeId)
	}

	if err := repo.SaveSymptomTypeAudit(symptomType, constant.DELETE, deletedBy); err != nil {
		tx.Rollback()
		return em.ErrorMessage("AuditError", "SymptomType", err)
	}

	result := tx.Model(&models.SymptomTypeMaster{}).Where("symptom_type_id = ?", symptomTypeId).Update("is_deleted", 1)
	if result.Error != nil {
		tx.Rollback()
		return em.ErrorMessage("DeleteError", "SymptomType", symptomTypeId)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return em.ErrorMessage("NoRowsAffected", "SymptomType", symptomTypeId)
	}

	return tx.Commit().Error
}
func (repo *SymptomRepositoryImpl) GetSymptomTypeById(symptomTypeId uint64, isDeleted int) (*models.SymptomTypeMaster, error) {
	var symptomType models.SymptomTypeMaster
	err := repo.db.Where("symptom_type_id = ? AND is_deleted = ? ", symptomTypeId, isDeleted).First(&symptomType).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, em.ErrorMessage("NotFound", "SymptomType", symptomTypeId)
		}
		return nil, em.ErrorMessage("NotFound", "SymptomType", symptomTypeId)
	}
	return &symptomType, nil
}

func (repo *SymptomRepositoryImpl) SaveSymptomTypeAudit(symptomType *models.SymptomTypeMaster, operationType string, updatedBy string) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		audit := models.SymptomTypeAudit{
			SymptomTypeId:          symptomType.SymptomTypeId,
			SymptomType:            symptomType.SymptomType,
			SymptomTypeDescription: symptomType.SymptomTypeDescription,
			IsDeleted:              symptomType.IsDeleted,
			OperationType:          operationType,
			CreatedAt:              symptomType.CreatedAt,
			UpdatedAt:              symptomType.UpdatedAt,
			CreatedBy:              symptomType.CreatedBy,
			UpdatedBy:              updatedBy,
		}

		if err := tx.Create(&audit).Error; err != nil {
			return em.ErrorMessage("AuditError", "SymptomType", err)
		}

		return nil
	})
}

func (repo *SymptomRepositoryImpl) GetAllSymptomTypeAuditRecord(limit, offset int) ([]models.SymptomTypeAudit, int64, error) {
	var auditRecords []models.SymptomTypeAudit
	var totalRecords int64

	if err := repo.db.Model(&models.SymptomTypeAudit{}).Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	if err := repo.db.Limit(limit).Offset(offset).Find(&auditRecords).Error; err != nil {
		return nil, 0, err
	}

	return auditRecords, totalRecords, nil
}

func (repo *SymptomRepositoryImpl) GetSymptomTypeAuditRecord(symptomTypeId, symptomTypeAuditId uint64) ([]models.SymptomTypeAudit, error) {
	var symptomTypeAudit []models.SymptomTypeAudit
	query := repo.db.Model(&models.SymptomTypeAudit{})

	if symptomTypeId != 0 {
		query = query.Where("symptom_type_id = ?", symptomTypeId)
	}

	if symptomTypeAuditId != 0 {
		query = query.Where("symptom_type_audit_id = ?", symptomTypeAuditId)
	}

	err := query.Order("symptom_type_audit_id DESC").Find(&symptomTypeAudit).Error
	if err != nil {
		return nil, err
	}

	return symptomTypeAudit, nil
}
