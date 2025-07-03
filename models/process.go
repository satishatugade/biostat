package models

import (
	"time"

	"github.com/google/uuid"
)

type ProcessStatus struct {
	ProcessStatusID uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey;column:process_status_id" json:"process_status_id"`
	UserID          uint64     `gorm:"column:user_id" json:"user_id"`
	ProcessType     string     `gorm:"column:process_type" json:"process_type"`
	EntityID        string     `gorm:"column:entity_id" json:"entity_id"`
	EntityType      string     `gorm:"column:entity_type" json:"entity_type"`
	Status          string     `gorm:"column:status" json:"status"`
	StatusMessage   string     `gorm:"column:status_message" json:"status_message"`
	Step            string     `gorm:"column:step" json:"step"`
	StartedAt       time.Time  `gorm:"column:started_at;autoCreateTime" json:"started_at"`
	UpdatedAt       time.Time  `gorm:"column:started_at;autoCreateTime" json:"updated_at"`
	CompletedAt     *time.Time `gorm:"column:completed_at" json:"completed_at"`
}

func (ProcessStatus) TableName() string {
	return "tbl_process_status"
}
