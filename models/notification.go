package models

import (
	"time"

	"github.com/google/uuid"
)

type UserNotificationMapping struct {
	ID               uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID           uint64    `gorm:"column:user_id" json:"user_id"`
	NotificationID   uuid.UUID `gorm:"column:notification_id" json:"notification_id"`
	SourceType       string    `gorm:"source_type" json:"source_type"`
	SourceID         string    `gorm:"source_id" json:"source_id"`
	Title            string    `gorm:"column:title" json:"title"`
	Message          string    `gorm:"column:message" json:"message"`
	Tags             string    `gorm:"column:tags" json:"tags"`
	NotificationType string    `gorm:"column:notification_type" json:"notification_type"`
	CreatedAt        time.Time `gorm:"column:create_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (UserNotificationMapping) TableName() string {
	return "user_notification_mapping"
}
