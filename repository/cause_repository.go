package repository

import (
	"biostat/constant"
	"biostat/models"
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
)

var em = constant.NewErrorManager()

type CauseRepository interface {
	GetAllCauses(limit int, offset int) ([]models.Cause, int64, error)
	AddDiseaseCause(cause *models.Cause) (*models.Cause, error)
	UpdateCause(cause *models.Cause, authUserId string) (*models.Cause, error)
	DeleteCause(causeId uint64, authUserId string) error
	GetCauseAuditRecord(causeId uint64, causeAuditId uint64) ([]models.CauseAudit, error)
	GetAllCauseAuditRecord(limit, offset int) ([]models.CauseAudit, int64, error)
	AddDiseaseCauseMapping(DCMapping *models.DiseaseCauseMapping) error

	//cause type
	GetAllCauseTypes(limit int, offset int) ([]models.CauseTypeMaster, int64, error)
	AddCauseType(causeType *models.CauseTypeMaster) (*models.CauseTypeMaster, error)
	UpdateCauseType(causeType *models.CauseTypeMaster, authUserId string) (*models.CauseTypeMaster, error)
	DeleteCauseType(causeTypeId uint64, authUserId string) error

	GetCauseTypeAuditRecord(causeTypeId uint64, causeTypeAuditId uint64) ([]models.CauseTypeAudit, error)
	GetAllCauseTypeAuditRecord(limit int, offset int) ([]models.CauseTypeAudit, int64, error)
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

	query := c.db.Model(&models.Cause{}).Preload("CauseType").Order("cause_id DESC")
	query.Count(&totalRecords)

	err := query.Limit(limit).Offset(offset).Find(&causes).Error
	if err != nil {
		return nil, 0, err
	}
	return causes, totalRecords, nil
}

func (c *CauseRepositoryImpl) AddDiseaseCause(cause *models.Cause) (*models.Cause, error) {
	tx := c.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := tx.Create(cause).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(cause.CauseTypeId) > 0 {
		var mappings []models.CauseTypeMapping
		for _, causeTypeId := range cause.CauseTypeId {
			mapping := models.CauseTypeMapping{
				CauseId:     cause.CauseId,
				CauseTypeId: causeTypeId,
				CreatedBy:   cause.CreatedBy,
			}
			mappings = append(mappings, mapping)
		}

		if err := tx.Create(&mappings).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	var insertedCause models.Cause
	if err := tx.Preload("CauseType").First(&insertedCause, cause.CauseId).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	return &insertedCause, tx.Commit().Error
}

func (repo *CauseRepositoryImpl) GetDiseaseCauseById(causeId uint64, isDeleted int) (*models.Cause, error) {
	var cause models.Cause
	if err := repo.db.Where("cause_id = ? AND is_deleted = ?", causeId, isDeleted).First(&cause).Error; err != nil {
		return nil, err
	}
	return &cause, nil
}

func (repo *CauseRepositoryImpl) UpdateCause(updatedCause *models.Cause, authUserId string) (*models.Cause, error) {
	tx := repo.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	existingTypeIds := make(map[uint64]bool)
	existingCause, err := repo.GetDiseaseCauseById(updatedCause.CauseId, 1)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	updatedCause.UpdatedAt = time.Now()
	result := tx.Model(&models.Cause{}).Where("cause_id = ?", updatedCause.CauseId).Updates(updatedCause)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, errors.New("no records updated")
	}
	if len(updatedCause.CauseTypeId) > 0 {
		var existingMappings []models.CauseTypeMapping
		if err := tx.Where("cause_id = ?", updatedCause.CauseId).Find(&existingMappings).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		for _, mapping := range existingMappings {
			existingTypeIds[mapping.CauseTypeId] = true
		}
		var newMappings []models.CauseTypeMapping
		for _, causeTypeId := range updatedCause.CauseTypeId {
			if !existingTypeIds[causeTypeId] {
				newMappings = append(newMappings, models.CauseTypeMapping{
					CauseId:     updatedCause.CauseId,
					CauseTypeId: causeTypeId,
					CreatedBy:   authUserId,
				})
			}
		}
		if len(newMappings) > 0 {
			if err := tx.Create(&newMappings).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}
	for causeTypeId, Exist := range existingTypeIds {
		if Exist {
			log.Println("Before saving data into cause audit table cause type exist : ", Exist)
			if err := repo.SaveCauseAudit(existingCause, &causeTypeId, constant.UPDATE, authUserId); err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}
	var finalUpdatedCause models.Cause
	err = tx.Preload("CauseType").Where("cause_id = ?", updatedCause.CauseId).First(&finalUpdatedCause).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	return &finalUpdatedCause, tx.Commit().Error
}

func (repo *CauseRepositoryImpl) SaveCauseAudit(existingCause *models.Cause, CauseTypeId *uint64, operationType string, updatedBy string) error {

	var causeTypeId uint64
	if CauseTypeId != nil {
		causeTypeId = *CauseTypeId
	}

	auditLog := models.CauseAudit{
		CauseId:       existingCause.CauseId,
		CauseName:     existingCause.CauseName,
		CauseTypeId:   causeTypeId,
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
	cause, err := repo.GetDiseaseCauseById(causeId, 0)
	if err != nil {
		return em.ErrorMessage("NotFound", "Cause", causeId)
	}
	if err := repo.SaveCauseAudit(cause, nil, constant.DELETE, deletedBy); err != nil {
		return em.ErrorMessage("AuditError", "Cause", err)
	}
	result := repo.db.Model(&models.Cause{}).Where("cause_id = ?", causeId).Update("is_deleted", 1)
	if result.Error != nil {
		return em.ErrorMessage("DeleteError", "Cause", causeId)
	}
	if result.RowsAffected == 0 {
		return em.ErrorMessage("NoRowsAffected", "Cause", causeId)
	}

	return em.ErrorMessage("Success", "Cause", causeId)
}

func (repo *CauseRepositoryImpl) GetAllCauseAuditRecord(limit, offset int) ([]models.CauseAudit, int64, error) {
	var auditLogs []models.CauseAudit
	var totalRecords int64

	repo.db.Model(&models.CauseAudit{}).Count(&totalRecords)
	err := repo.db.
		Limit(limit).
		Offset(offset).
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

func (repo *CauseRepositoryImpl) GetAllCauseTypes(limit int, offset int) ([]models.CauseTypeMaster, int64, error) {
	var causeTypes []models.CauseTypeMaster
	var totalCount int64

	query := repo.db.Model(&models.CauseTypeMaster{})

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, em.ErrorMessage("CountError", "CauseType", err)
	}

	if err := query.Limit(limit).Offset(offset).Find(&causeTypes).Error; err != nil {
		return nil, 0, em.ErrorMessage("NotFound", "CauseType", err)
	}

	return causeTypes, totalCount, nil
}

func (repo *CauseRepositoryImpl) AddCauseType(causeType *models.CauseTypeMaster) (*models.CauseTypeMaster, error) {
	tx := repo.db.Begin()
	if tx.Error != nil {
		return nil, em.ErrorMessage("TransactionStartError", "CauseType", tx.Error)
	}
	if err := tx.Create(&causeType).Error; err != nil {
		tx.Rollback()
		return nil, em.ErrorMessage("CreateError", "CauseType", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, em.ErrorMessage("TransactionCommitError", "CauseType", err)
	}
	return causeType, nil
}

func (repo *CauseRepositoryImpl) UpdateCauseType(updatedCauseType *models.CauseTypeMaster, authUserId string) (*models.CauseTypeMaster, error) {
	tx := repo.db.Begin()
	if tx.Error != nil {
		return nil, em.ErrorMessage("TransactionError", "CauseType", nil)
	}

	existingCauseType, err := repo.GetCauseTypeById(updatedCauseType.CauseTypeId, 1)
	if err != nil {
		tx.Rollback()
		return nil, em.ErrorMessage("NotFound", "CauseType", updatedCauseType.CauseTypeId)
	}

	updatedCauseType.UpdatedAt = time.Now()
	result := tx.Model(&models.CauseTypeMaster{}).Where("cause_type_id = ?", updatedCauseType.CauseTypeId).Updates(updatedCauseType)
	if result.Error != nil {
		tx.Rollback()
		return nil, em.ErrorMessage("UpdateError", "CauseType", updatedCauseType.CauseTypeId)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, em.ErrorMessage("NoRowsAffected", "CauseType", updatedCauseType.CauseTypeId)
	}

	if err := repo.SaveCauseTypeAudit(existingCauseType, constant.UPDATE, authUserId); err != nil {
		tx.Rollback()
		return nil, em.ErrorMessage("AuditError", "CauseType", err)
	}

	var finalUpdatedCauseType models.CauseTypeMaster
	err = tx.Where("cause_type_id = ?", updatedCauseType.CauseTypeId).First(&finalUpdatedCauseType).Error
	if err != nil {
		tx.Rollback()
		return nil, em.ErrorMessage("NotFound", "CauseType", updatedCauseType.CauseTypeId)
	}

	return &finalUpdatedCauseType, tx.Commit().Error
}

func (repo *CauseRepositoryImpl) DeleteCauseType(causeTypeId uint64, deletedBy string) error {
	causeType, err := repo.GetCauseTypeById(causeTypeId, 0)
	if err != nil {
		return em.ErrorMessage("NotFound", "CauseType", causeTypeId)
	}
	if err := repo.SaveCauseTypeAudit(causeType, constant.DELETE, deletedBy); err != nil {
		return em.ErrorMessage("AuditError", "CauseType", err)
	}
	result := repo.db.Model(&models.CauseTypeMaster{}).Where("cause_type_id = ?", causeTypeId).Update("is_deleted", 1)
	if result.Error != nil {
		return em.ErrorMessage("DeleteError", "CauseType", causeTypeId)
	}

	if result.RowsAffected == 0 {
		return em.ErrorMessage("NoRowsAffected", "CauseType", causeTypeId)
	}

	return em.ErrorMessage("Success", "CauseType", causeTypeId)
}

func (repo *CauseRepositoryImpl) GetCauseTypeById(causeTypeId uint64, isDeleted int) (*models.CauseTypeMaster, error) {
	var causeType models.CauseTypeMaster

	err := repo.db.Where("cause_type_id = ?", causeTypeId).First(&causeType).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, em.ErrorMessage("NotFound", "CauseType", causeTypeId)
		}
		return nil, em.ErrorMessage("NotFound", "CauseType", causeTypeId)
	}

	return &causeType, nil
}

func (repo *CauseRepositoryImpl) SaveCauseTypeAudit(causeType *models.CauseTypeMaster, operationType string, updatedBy string) error {
	audit := models.CauseTypeAudit{
		CauseTypeId:          causeType.CauseTypeId,
		CauseType:            causeType.CauseType,
		CauseTypeDescription: causeType.CauseTypeDescription,
		IsDeleted:            causeType.IsDeleted,
		OperationType:        operationType,
		CreatedAt:            causeType.CreatedAt,
		UpdatedAt:            causeType.UpdatedAt,
		CreatedBy:            causeType.CreatedBy,
		UpdatedBy:            updatedBy,
	}

	if err := repo.db.Create(&audit).Error; err != nil {
		return em.ErrorMessage("AuditError", "CauseType", err)
	}

	return nil
}

func (r *CauseRepositoryImpl) GetAllCauseTypeAuditRecord(limit int, offset int) ([]models.CauseTypeAudit, int64, error) {
	var causeTypeAudits []models.CauseTypeAudit
	var totalRecords int64

	query := r.db.Model(&models.CauseTypeAudit{})

	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&causeTypeAudits).Error; err != nil {
		return nil, 0, err
	}
	return causeTypeAudits, totalRecords, nil
}

func (r *CauseRepositoryImpl) GetCauseTypeAuditRecord(causeTypeId, causeTypeAuditId uint64) ([]models.CauseTypeAudit, error) {
	var causeTypeAudit []models.CauseTypeAudit
	query := r.db.Model(&models.CauseTypeAudit{})

	if causeTypeId != 0 {
		query = query.Where("cause_type_id = ?", causeTypeId)
	}
	if causeTypeAuditId != 0 {
		query = query.Where("cause_type_audit_id = ?", causeTypeAuditId)
	}
	err := query.Order("cause_type_audit_id DESC").Find(&causeTypeAudit).Error

	if err != nil {
		return nil, err
	}

	return causeTypeAudit, nil
}
