package repository

import (
	"biostat/constant"
	"biostat/models"
	"biostat/utils"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SubscriptionRepository interface {
	GetSubsciptionMasterPlanBySubscriptionId(subscriptionID uint64) (*models.SubscriptionMaster, error)
	GetFamilyById(familyID uint64) (*models.PatientFamilyGroup, error)
	UpdateFamilySubscription(family *models.PatientFamilyGroup) (uint64, error)
	CreateFamily(family *models.PatientFamilyGroup) error
	UpdateSubscriptionStatus(enabled bool, updatedBy string) error
	GetSubscriptionShowStatus() (bool, error)
	GetSubscriptionWithServices() ([]models.SubscriptionMaster, error)
	GetFamilyGroupByMemberID(memberID uint64) (*models.PatientFamilyGroup, error)
	CheckActiveSubscriptionPlanByMemberId(memberId uint64) (*models.SubscriptionMaster, bool, error)
}

type SubscriptionRepositoryImpl struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &SubscriptionRepositoryImpl{db: db}
}

func (r *SubscriptionRepositoryImpl) GetSubsciptionMasterPlanBySubscriptionId(subscriptionId uint64) (*models.SubscriptionMaster, error) {
	var plan models.SubscriptionMaster
	if err := r.db.First(&plan, subscriptionId).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *SubscriptionRepositoryImpl) GetFamilyById(familyID uint64) (*models.PatientFamilyGroup, error) {
	var family models.PatientFamilyGroup
	if err := r.db.First(&family, familyID).Error; err != nil {
		return nil, err
	}
	return &family, nil
}

func (r *SubscriptionRepositoryImpl) UpdateFamilySubscription(family *models.PatientFamilyGroup) (uint64, error) {
	if err := r.db.Save(family).Error; err != nil {
		return 0, err
	}
	return family.FamilyId, nil
}

func (r *SubscriptionRepositoryImpl) CreateFamily(family *models.PatientFamilyGroup) error {
	return r.db.Create(family).Error
}

func (r *SubscriptionRepositoryImpl) GetSubscriptionShowStatus() (bool, error) {
	var setting models.SystemSetting
	err := r.db.First(&setting, "setting_key = ?", constant.SUBSCRIPTIONENABLED).Error

	if err == gorm.ErrRecordNotFound {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return setting.SettingValue, nil
}

func (r *SubscriptionRepositoryImpl) UpdateSubscriptionStatus(enabled bool, updatedBy string) error {
	setting := models.SystemSetting{
		SettingKey:   constant.SUBSCRIPTIONENABLED,
		SettingValue: enabled,
		UpdatedAt:    time.Now(),
		UpdatedBy:    updatedBy,
	}

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "setting_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"setting_value", "updated_at", "updated_by"}),
	}).Create(&setting).Error
}

func (r *SubscriptionRepositoryImpl) GetSubscriptionWithServices() ([]models.SubscriptionMaster, error) {
	var plans []models.SubscriptionMaster

	err := r.db.
		Preload("ServiceMappings.Service").
		Find(&plans).Error

	if err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *SubscriptionRepositoryImpl) GetFamilyGroupByMemberID(memberID uint64) (*models.PatientFamilyGroup, error) {
	var family models.PatientFamilyGroup

	err := r.db.
		Where("member_id = ?", memberID).
		First(&family).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no family found for member_id %d", memberID)
		}
		return nil, fmt.Errorf("failed to fetch family group: %w", err)
	}

	return &family, nil
}

func (r *SubscriptionRepositoryImpl) CheckActiveSubscriptionPlanByMemberId(memberID uint64) (*models.SubscriptionMaster, bool, error) {
	var familyGroup models.PatientFamilyGroup

	err := r.db.Where("member_id = ? AND is_active = true", memberID).First(&familyGroup).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	now := time.Now()
	if familyGroup.SubscriptionStartDate != nil && familyGroup.SubscriptionEndDate != nil {
		if now.Before(*familyGroup.SubscriptionStartDate) {
			return nil, false, nil
		}
		if now.After(*familyGroup.SubscriptionEndDate) {
			return nil, false, nil
		}
	}

	if familyGroup.CurrentSubscriptionId == nil {
		return nil, false, nil
	}

	var plan models.SubscriptionMaster
	err = r.db.First(&plan, *familyGroup.CurrentSubscriptionId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	subscriptionStartDate := utils.FormatDateTime(familyGroup.SubscriptionStartDate)
	subscriptionEndDate := utils.FormatDateTime(familyGroup.SubscriptionEndDate)
	plan.SubscriptionStartDate = &subscriptionStartDate
	plan.SubscriptionEndDate = &subscriptionEndDate
	plan.FamilyId = familyGroup.FamilyId
	plan.FamilyName = familyGroup.FamilyName
	plan.MemberId = familyGroup.MemberId
	return &plan, true, nil
}
