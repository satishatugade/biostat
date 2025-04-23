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
	PaymentID       uint64    `gorm:"column:payment_id" json:"payment_id"`
	Notes           string    `gorm:"column:notes" json:"notes"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	IsDeleted       int       `gorm:"column:is_deleted;default:0" json:"is_deleted"`
}

func (Appointment) TableName() string {
	return "tbl_appointment_master"
}

type AppointmentResponse struct {
	AppointmentID   uint64      `json:"appointment_id"`
	PatientID       uint64      `json:"patient_id"`
	ProviderType    string      `json:"provider_type"`
	ProviderInfo    interface{} `json:"provider_info"`
	ScheduledBy     uint64      `json:"scheduled_by"`
	AppointmentType string      `json:"appointment_type"`
	AppointmentDate time.Time   `json:"appointment_date"`
	AppointmentTime string      `json:"appointment_time"`
	DurationMinutes int         `json:"duration_minutes"`
	IsInperson      int         `json:"is_inperson"`
	MeetingUrl      string      `json:"meeting_url"`
	Status          string      `json:"status"`
	PaymentStatus   string      `json:"payment_status"`
	PaymentID       uint64      `json:"payment_id"`
	Notes           string      `json:"notes"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	IsDeleted       int         `json:"is_deleted"`
}

type DoctorInfo struct {
	DoctorId          uint64 `json:"doctor_id"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Specialty         string `json:"specialty"`
	Gender            string `json:"gender"`
	MobileNo          string `json:"mobile_no"`
	ClinicName        string `json:"clinic_name"`
	ClinicAddress     string `json:"clinic_address"`
	YearsOfExperience int    `json:"years_of_experience"`
	WorkingHours      string `json:"working_hours"`
}

type NurseInfo struct {
	NurseId           uint64 `json:"nurse_id"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	Gender            string `json:"gender"`
	MobileNo          string `json:"mobile_no"`
	Specialty         string `json:"specialty"`
	ClinicName        string `json:"clinic_name"`
	ClinicAddress     string `json:"clinic_address"`
	YearsOfExperience int    `json:"years_of_experience"`
	WorkingHours      string `json:"working_hours"`
}

type LabInfo struct {
	LabId            uint64 `json:"lab_id"`
	LabNo            string `json:"lab_no"`
	LabName          string `json:"lab_name"`
	LabAddress       string `json:"lab_address"`
	LabContactNumber string `json:"lab_contact_number"`
	LabEmail         string `json:"lab_email"`
}

type PatientInfo struct {
	PatientId  uint64 `json:"patient_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	BloodGroup string `json:"blood_group"`
}
