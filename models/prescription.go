package models

import (
	"time"
)

// PatientPrescription represents the main prescription record
type PatientPrescription struct {
	PrescriptionId            uint      `gorm:"primaryKey;column:prescription_id" json:"prescription_id"`
	PatientId                 uint      `gorm:"column:patient_id" json:"patient_id"`
	DoctorId                  uint      `gorm:"column:doctor_id" json:"doctor_id"`
	PrescriptionName          string    `gorm:"column:prescription_name" json:"prescription_name"`
	Description               string    `gorm:"column:description" json:"description"`
	PrescriptionDate          time.Time `gorm:"column:prescription_date" json:"prescription_date"`
	PrescriptionAttachmentUrl string    `gorm:"column:prescription_attachment_url" json:"prescription_attachment_url"`

	// Relationship to PrescriptionDetail
	PrescriptionDetails []PrescriptionDetail `gorm:"foreignKey:PrescriptionId;references:PrescriptionId" json:"prescription_details"`
}

// TableName for PatientPrescription
func (PatientPrescription) TableName() string {
	return "tbl_patient_prescription"
}

// PrescriptionDetail represents the details of each prescription
type PrescriptionDetail struct {
	PrescriptionDetailId uint    `gorm:"primaryKey;column:prescription_detail_id" json:"prescription_detail_id"`
	PrescriptionId       uint    `gorm:"column:prescription_id" json:"prescription_id"`
	PrescriptionType     string  `gorm:"column:prescription_type" json:"prescription_type"`
	DoseQuantity         float64 `gorm:"column:dose_quantity" json:"dose_quantity"`
	Duration             int     `gorm:"column:duration" json:"duration"`
	UnitValue            float64 `gorm:"column:unit_value" json:"unit_value"`
	UnitType             string  `gorm:"column:unit_type" json:"unit_type"`
	Frequency            int     `gorm:"column:frequency" json:"frequency"`
	TimesPerDay          int     `gorm:"column:times_per_day" json:"times_per_day"`
	IntervalHour         int     `gorm:"column:interval_hour" json:"interval_hour"`
	Instruction          string  `gorm:"column:instruction" json:"instruction"`
}

// TableName for PrescriptionDetail
func (PrescriptionDetail) TableName() string {
	return "tbl_prescription_detail"
}
