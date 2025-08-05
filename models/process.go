package models

import (
	"time"

	"github.com/google/uuid"
)

type ProcessStatus struct {
	ProcessStatusID uuid.UUID        `gorm:"type:uuid;default:gen_random_uuid();primaryKey;column:process_status_id" json:"process_status_id"`
	UserID          uint64           `gorm:"column:user_id" json:"user_id"`
	ProcessType     string           `gorm:"column:process_type" json:"process_type"`
	EntityID        string           `gorm:"column:entity_id" json:"entity_id"`
	EntityType      string           `gorm:"column:entity_type" json:"entity_type"`
	Status          string           `gorm:"column:status" json:"status"`
	StatusMessage   string           `gorm:"column:status_message" json:"status_message"`
	Step            string           `gorm:"column:step" json:"step"`
	StartedAt       time.Time        `gorm:"column:started_at;autoCreateTime" json:"started_at"`
	UpdatedAt       time.Time        `gorm:"column:updated_at;autoCreateTime" json:"updated_at"`
	CompletedAt     *time.Time       `gorm:"column:completed_at" json:"completed_at"`
	ActivityLog     []ProcessStepLog `gorm:"foreignKey:ProcessStatusID" json:"activity_log"`
}

func (ProcessStatus) TableName() string {
	return "tbl_process_status"
}

type ProcessStepLog struct {
	ProcessStepLogId uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey;column:process_step_log_id" json:"process_step_log_id"`
	ProcessStatusID  uuid.UUID `gorm:"type:uuid;column:process_status_id" json:"process_status_id"`
	Step             string    `gorm:"column:step" json:"step"`
	Status           string    `gorm:"column:status" json:"status"`
	Message          string    `gorm:"column:message" json:"message"`
	RecordIndex      *int64    `gorm:"column:record_index" json:"record_index,omitempty"`
	TotalRecords     *int      `gorm:"column:total_records" json:"total_records,omitempty"`
	SuccessCount     *int      `gorm:"column:success_count" json:"success_count"`
	FailedCount      *int      `gorm:"column:failed_count" json:"failed_count"`
	Error            *string   `gorm:"column:error" json:"error,omitempty"`
	StepStartedAt    time.Time `gorm:"column:step_started_at;autoCreateTime" json:"step_started_at"`
	StepUpdatedAt    time.Time `gorm:"column:step_updated_at;autoCreateTime" json:"step_updated_at"`
}

func (ProcessStepLog) TableName() string {
	return "tbl_process_step_log"
}

type ProcessStatusResponse struct {
	ProcessStatusID string                   `json:"process_status_id"`
	UserID          uint64                   `json:"user_id"`
	ProcessType     string                   `json:"process_type"`
	EntityID        string                   `json:"entity_id"`
	EntityType      string                   `json:"entity_type"`
	StartedAt       string                   `json:"started_at"`
	CompletedAt     string                   `json:"completed_at"`
	EndedAt         string                   `json:"ended_at"`
	Status          string                   `json:"status"`
	ActivityLog     []ProcessStepLogResponse `json:"activity_log"`
}

type ProcessStepLogResponse struct {
	ProcessStepLogId string  `json:"process_step_log_id"`
	StepName         string  `json:"step_name"`
	StartedAt        string  `json:"step_started_at"`
	CompletedAt      string  `json:"completed_at"`
	StepStatus       string  `json:"step_status"`
	Message          string  `json:"message"`
	RecordIndex      *int64  `json:"record_index"`
	TotalRecords     *int    `json:"total_records"`
	SuccessCount     *int    `json:"success_count"`
	FailedCount      *int    `json:"failed_count"`
	Error            *string `json:"error"`
}
