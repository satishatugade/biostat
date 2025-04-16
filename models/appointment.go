package models

import (
	"time"
)

type Appointment struct {
	AppointmentID   uint64    `gorm:"column:appointment_id;primaryKey;autoIncrement" json:"appointment_id"`
	PatientID       uint64    `gorm:"column:patient_id" json:"patient_id"`
	ProviderID      uint64    `gorm:"column:provider_id" json:"provider_id"`
	ProviderType    string    `gorm:"column:provider_type;size:20;not null" json:"provider_type"`
	ScheduledBy     uint64    `gorm:"column:scheduled_by" json:"scheduled_by"`
	AppointmentType string    `gorm:"column:appointment_type;not null" json:"appointment_type"`
	AppointmentDate time.Time `gorm:"column:appointment_date;type:date;not null" json:"appointment_date"`
	AppointmentTime string    `gorm:"column:appointment_time;type:time;not null" json:"appointment_time"`
	DurationMinutes int       `gorm:"column:duration_minutes;default:30" json:"duration_minutes"`
	IsInperson      int       `gorm:"column:is_inperson;default:0" json:"is_inperson"`
	MeetingUrl      string    `gorm:"column:meeting_url;" json:"meeting_url"`
	Status          string    `gorm:"column:status;size:20;default:'pending'" json:"status"`
	PaymentStatus   string    `gorm:"column:payment_status;size:20;default:'unpaid'" json:"payment_status"`
	PaymentID       string    `gorm:"column:payment_id" json:"payment_id"`
	Notes           string    `gorm:"column:notes" json:"notes"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	IsDeleted       int       `gorm:"column:is_deleted;default:0" json:"is_deleted"`
}

func (Appointment) TableName() string {
	return "tbl_appointment_master"
}
