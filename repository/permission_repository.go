package repository

import (
	"biostat/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type PermissionRepository interface {
	UpsertUserRelativePermission(grantedToUserId, relativeId uint64, permissionId int64, granted bool) error
	GetAllPermissions() ([]models.PermissionMaster, error)
	GetPermissionByCode(code string) (*models.PermissionMaster, error)
	CheckPermissionValue(userID, relativeID uint64, permissionID int64) (exists bool, currentValue bool, err error)

	UpdatePermissionValue(userID, relativeID uint64, permissionID int64, value bool) error
	GrantPermission(userID, relativeID uint64, permissionID int64, granted bool) error
	HasPermission(userID, relativeID uint64, permissionCode string) error
	ListPermissions(userID, relativeID uint64) ([]models.PermissionResult, error)
}

type PermissionRepositoryImpl struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &PermissionRepositoryImpl{db: db}
}

func (r *PermissionRepositoryImpl) UpsertUserRelativePermission(grantedToUserID, relativeID uint64, permissionID int64, granted bool) error {
	query := `
		INSERT INTO tbl_user_relative_permission_mappings 
		(user_id, relative_id, permission_id, granted, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
		ON CONFLICT (user_id, relative_id, permission_id)
		DO UPDATE SET granted = EXCLUDED.granted, updated_at = EXCLUDED.updated_at;
	`
	return r.db.Exec(query, grantedToUserID, relativeID, permissionID, granted).Error
}

func (r *PermissionRepositoryImpl) GetPermissionByCode(code string) (*models.PermissionMaster, error) {
	var permission models.PermissionMaster
	err := r.db.Where("code = ?", code).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *PermissionRepositoryImpl) CheckPermissionValue(userID, relativeID uint64, permissionID int64) (exists bool, currentValue bool, err error) {
	var mapping models.UserRelativePermissionMapping
	err = r.db.Where("user_id = ? AND relative_id = ? AND permission_id = ?", userID, relativeID, permissionID).
		First(&mapping).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, false, nil
	}
	return true, mapping.Granted, err
}

func (r *PermissionRepositoryImpl) UpdatePermissionValue(userID, relativeID uint64, permissionID int64, value bool) error {
	return r.db.Model(&models.UserRelativePermissionMapping{}).
		Where("user_id = ? AND relative_id = ? AND permission_id = ?", userID, relativeID, permissionID).
		Update("granted", value).Error
}

func (r *PermissionRepositoryImpl) GetAllPermissions() ([]models.PermissionMaster, error) {
	var permissions []models.PermissionMaster
	result := r.db.Order("permission_id").Find(&permissions)
	if result.Error != nil {
		return nil, result.Error
	}
	return permissions, nil
}

func (r *PermissionRepositoryImpl) GrantPermission(userID, relativeID uint64, permissionID int64, granted bool) error {
	var mapping models.UserRelativePermissionMapping
	err := r.db.Where("user_id = ? AND relative_id = ? AND permission_id = ?", userID, relativeID, permissionID).First(&mapping).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		mapping = models.UserRelativePermissionMapping{
			UserID: userID, RelativeID: relativeID, PermissionID: permissionID, Granted: granted,
		}
		return r.db.Create(&mapping).Error
	}
	if err != nil {
		return err
	}
	mapping.Granted = granted
	return r.db.Save(&mapping).Error
}

func (r *PermissionRepositoryImpl) HasPermission(userID, relativeID uint64, permissionCode string) error {
	var permission models.PermissionMaster
	if err := r.db.Where("code = ?", permissionCode).First(&permission).Error; err != nil {
		return fmt.Errorf("permission code not found: %w", err)
	}
	var mapping models.UserRelativePermissionMapping
	err := r.db.Where("user_id = ? AND relative_id = ? AND permission_id = ? AND granted = true", userID, relativeID, permission.PermissionID).First(&mapping).Error

	return err
}

func (r *PermissionRepositoryImpl) ListPermissions(userID, relativeID uint64) ([]models.PermissionResult, error) {
	var results []models.PermissionResult

	err := r.db.Table("tbl_user_relative_permission_mappings").
		Select("tbl_user_relative_permission_mappings.user_id, tbl_user_relative_permission_mappings.relative_id, tbl_permissions_master.code, tbl_user_relative_permission_mappings.granted").
		Joins("JOIN tbl_permissions_master ON tbl_user_relative_permission_mappings.permission_id = tbl_permissions_master.permission_id").
		Where("tbl_user_relative_permission_mappings.user_id = ? AND tbl_user_relative_permission_mappings.relative_id = ?", userID, relativeID).
		Scan(&results).Error

	return results, err
}
