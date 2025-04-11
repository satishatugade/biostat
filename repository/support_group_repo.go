package repository

import (
	"biostat/constant"
	"biostat/models"
	"time"

	"gorm.io/gorm"
)

type SupportGroupRepository interface {
	AddSupportGroup(supportGroup *models.SupportGroup) error
	GetSupportGroupById(supportGrpId uint64) (*models.SupportGroup, error)
	GetAllSupportGroups(limit, offset int) ([]models.SupportGroup, int64, error)
	UpdateSupportGroup(updatedData *models.SupportGroup, authUserId string) error
	DeleteSupportGroup(supportGrpId uint64, authUserId string) error
	GetSupportGroupAuditRecord(supportGroupId, supportGroupAuditId uint64) ([]models.SupportGroupAudit, error)
	GetAllSupportGroupAuditRecord(page, limit int) ([]models.SupportGroupAudit, int64, error)
}

type SupportGroupRepositoryImpl struct {
	db *gorm.DB
}

func NewSupportGroupRepository(db *gorm.DB) SupportGroupRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &SupportGroupRepositoryImpl{db: db}
}

func (r *SupportGroupRepositoryImpl) AddSupportGroup(group *models.SupportGroup) error {
	return r.db.Create(group).Error
}

func (r *SupportGroupRepositoryImpl) GetSupportGroupById(supportGrpId uint64) (*models.SupportGroup, error) {
	var group models.SupportGroup
	if err := r.db.Where("support_group_id = ?", supportGrpId).First(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *SupportGroupRepositoryImpl) GetAllSupportGroups(limit, offset int) ([]models.SupportGroup, int64, error) {
	var groups []models.SupportGroup
	var total int64
	err := r.db.Model(&models.SupportGroup{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = r.db.Offset(offset).Limit(limit).Order("created_at desc").Find(&groups).Error
	return groups, total, err
}

func SaveSupportGroupAudit(tx *gorm.DB, group *models.SupportGroup, operationType string, performedBy string) error {
	audit := models.SupportGroupAudit{
		SupportGroupId: group.SupportGroupId,
		GroupName:      group.GroupName,
		Description:    group.Description,
		Location:       group.Location,
		IsDeleted:      group.IsDeleted,
		OperationType:  operationType,
		CreatedBy:      performedBy,
		CreatedAt:      time.Now(),
	}
	return tx.Create(&audit).Error
}

func (r *SupportGroupRepositoryImpl) UpdateSupportGroup(updatedData *models.SupportGroup, updatedBy string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing models.SupportGroup
		if err := tx.Where("support_group_id = ?", updatedData.SupportGroupId).First(&existing).Error; err != nil {
			return err
		}
		if err := SaveSupportGroupAudit(tx, &existing, constant.UPDATE, updatedBy); err != nil {
			return err
		}
		if err := tx.Model(&existing).Updates(updatedData).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *SupportGroupRepositoryImpl) DeleteSupportGroup(supportGrpId uint64, authUserId string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing models.SupportGroup
		if err := tx.Where("support_group_id = ?", supportGrpId).First(&existing).Error; err != nil {
			return err
		}

		if err := SaveSupportGroupAudit(tx, &existing, constant.DELETE, authUserId); err != nil {
			return err
		}
		if err := tx.Model(&existing).Updates(map[string]interface{}{
			"is_deleted": 1,
			"updated_by": authUserId,
			"updated_at": time.Now(),
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *SupportGroupRepositoryImpl) GetSupportGroupAuditRecord(supportGroupId, supportGroupAuditId uint64) ([]models.SupportGroupAudit, error) {
	var audits []models.SupportGroupAudit
	query := r.db.Model(&models.SupportGroupAudit{})

	if supportGroupId != 0 {
		query = query.Where("support_group_id = ?", supportGroupId)
	}
	if supportGroupAuditId != 0 {
		query = query.Where("support_group_audit_id = ?", supportGroupAuditId)
	}

	err := query.Order("created_at DESC").Find(&audits).Error
	return audits, err
}

func (r *SupportGroupRepositoryImpl) GetAllSupportGroupAuditRecord(limit, offset int) ([]models.SupportGroupAudit, int64, error) {
	var audits []models.SupportGroupAudit
	var total int64

	tx := r.db.Model(&models.SupportGroupAudit{}).Count(&total)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	err := r.db.Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&audits).Error

	return audits, total, err
}
