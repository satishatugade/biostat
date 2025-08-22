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
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (UserNotificationMapping) TableName() string {
	return "tbl_user_notification_mapping"
}

type ReminderConfig struct {
	TimeSlot     string `json:"time_slot"`
	Time         string `json:"time"`      // "HH:MM"
	Frequency    string `json:"frequency"` // e.g., "daily"
	DurationDays int    `json:"duration_days"`
	Medicines    []struct {
		Name string `json:"name"`
		Dose int    `json:"dose"`
		Unit string `json:"unit"`
	} `json:"medicines"`
}

type UpdateReminderRequest struct {
	NotificationID uuid.UUID  `json:"notification_id" binding:"required"`
	RepeatInterval *uint64    `json:"repeat_interval,omitempty"`
	RepeatTimes    *uint64    `json:"repeat_times,omitempty"`
	RepeatType     *string    `json:"repeat_type,omitempty"`
	RepeatUntil    *time.Time `json:"repeat_until,omitempty"`
	NextSendAt     *time.Time `json:"next_send_at,omitempty"`
	IsActive       *bool      `json:"is_active,omitempty"`
	Medicines      []struct {
		Name string `json:"name"`
		Dose int    `json:"dose"`
		Unit string `json:"unit"`
	} `json:"medicines"`
}

type UserReminder struct {
	ReminderID         uuid.UUID  `json:"notification_id"`
	Title              string     `json:"title"`
	Message            string     `json:"message"`
	ScheduleTime       *time.Time `json:"schedule_time"`
	NotificationStatus string     `json:"notification_status"`
	RetryCount         uint64     `json:"retry_count"`
	IsRecurring        bool       `json:"is_recurring"`
	RepeatType         string     `json:"repeat_type"`
	RepeatTimes        uint64     `json:"repeat_times"`
	RepeatInterval     uint64     `json:"repeat_interval"`
	SentTimes          uint64     `json:"sent_times"`
	RepeatUntil        *time.Time `json:"repeat_until"`
	NextSendAt         *time.Time `json:"next_send_at"`
	NotificationType   string     `json:"notification_type"`
	IsActive           bool       `json:"is_active"`
}

type RecipientRequest struct {
	FCMToken *string `json:"fcm_token"`
}
