package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type UserBioFeedback struct {
	FeedbackID   uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"feedback_id"`
	UserID       *int64         `gorm:"column:user_id" json:"user_id"`
	Name         *string        `gorm:"column:name" json:"name"`
	Email        *string        `gorm:"column:email" json:"email"`
	MobileNo     *string        `gorm:"column:mobile_no" json:"mobile_no"`
	FeedbackType string         `gorm:"column:feedback_type;not null" json:"feedback_type"`
	Message      string         `gorm:"column:message;not null" json:"message"`
	Images       datatypes.JSON `gorm:"column:images" json:"images"`
	DeviceInfo   datatypes.JSON `gorm:"column:device_info" json:"device_info"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (UserBioFeedback) TableName() string {
	return "tbl_user_bio_feedback"
}

type CreateFeedbackRequest struct {
	UserID       *int64      `json:"user_id"`
	Name         *string     `json:"name"`
	Email        *string     `json:"email"`
	MobileNo     *string     `json:"mobile_no"`
	FeedbackType string      `json:"feedback_type" binding:"required"`
	Message      string      `json:"message" binding:"required"`
	Images       interface{} `json:"images"`
	DeviceInfo   interface{} `json:"device_info"`
}
