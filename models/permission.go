package models

import (
	"time"

	"github.com/google/uuid"
)

type PermissionMaster struct {
	PermissionID int64     `gorm:"primaryKey;column:permission_id" json:"permission_id"`
	Code         string    `gorm:"uniqueIndex;column:code" json:"code"`
	Name         string    `gorm:"column:name" json:"name"`
	Description  string    `gorm:"column:description" json:"description"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (PermissionMaster) TableName() string {
	return "tbl_permissions_master"
}

type UserRelativePermissionMapping struct {
	MappingID    uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"mapping_id"`
	UserID       uint64    `gorm:"column:user_id" json:"user_id"`
	RelativeID   uint64    `gorm:"column:relative_id" json:"relative_id"`
	PermissionID int64     `gorm:"column:permission_id" json:"permission_id"`
	Granted      bool      `gorm:"column:granted" json:"granted"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserRelativePermissionMapping) TableName() string {
	return "tbl_user_relative_permission_mappings"
}

type PermissionResult struct {
	UserID     uint64 `json:"user_id"`
	RelativeID uint64 `json:"relative_id"`
	Code       string `json:"code"`
	Granted    bool   `json:"granted"`
}

type PermissionUpdateDto struct {
	Code      string `json:"code"`
	Value     bool   `json:"value"`
	GrantedTo uint64 `json:"granted_to"`
}

type RelativePermissionInput struct {
	RelativeID  uint64                `json:"relative_id"`
	Permissions []PermissionUpdateDto `json:"permissions"`
}

type ManagePermissionsRequest struct {
	RelativePermissions []RelativePermissionInput `json:"relative_permissions"`
}
