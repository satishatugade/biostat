package models

import (
	"time"
)

type PatientPrescription struct {
	PrescriptionId            uint64     `gorm:"primaryKey;column:prescription_id" json:"prescription_id"`
	PatientId                 uint64     `gorm:"column:patient_id" json:"patient_id"`
	PrescribedBy              string     `gorm:"column:prescribed_by" json:"prescribed_by"`
	PrescriptionName          *string    `gorm:"column:prescription_name" json:"prescription_name"`
	Description               string     `gorm:"column:description" json:"description"`
	PrescriptionDate          *time.Time `gorm:"column:prescription_date" json:"prescription_date"`
	PrescriptionAttachmentUrl string     `gorm:"column:prescription_attachment_url" json:"prescription_attachment_url"`

	// Relationship to PrescriptionDetail
	PrescriptionDetails []PrescriptionDetail `gorm:"foreignKey:PrescriptionId;references:PrescriptionId" json:"prescription_details"`
}

// TableName for PatientPrescription
func (PatientPrescription) TableName() string {
	return "tbl_patient_prescription"
}

// PrescriptionDetail represents the details of each prescription
type PrescriptionDetail struct {
	PrescriptionDetailId uint64    `gorm:"primaryKey;autoIncrement;column:prescription_detail_id" json:"prescription_detail_id"`
	PrescriptionId       uint64    `gorm:"column:prescription_id" json:"prescription_id"`
	MedicineName         string    `gorm:"column:medicine_name" json:"medicine_name"`
	PrescriptionType     string    `gorm:"column:prescription_type" json:"prescription_type"`
	DoseQuantity         float64   `gorm:"column:dose_quantity" json:"dose_quantity"`
	Duration             int       `gorm:"column:duration" json:"duration"`
	UnitValue            float64   `gorm:"column:unit_value" json:"unit_value"`
	UnitType             string    `gorm:"column:unit_type" json:"unit_type"`
	Frequency            int       `gorm:"column:frequency" json:"frequency"`
	TimesPerDay          int       `gorm:"column:times_per_day" json:"times_per_day"`
	IntervalHour         int       `gorm:"column:interval_hour" json:"interval_hour"`
	Instruction          string    `gorm:"column:instruction" json:"instruction"`
	CreatedAt            time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at" json:"-"`
	CreatedBy            string    `gorm:"column:created_by" json:"-"`
	UpdatedBy            string    `gorm:"column:updated_by" json:"-"`
}

// TableName for PrescriptionDetail
func (PrescriptionDetail) TableName() string {
	return "tbl_prescription_detail"
}

type PatientPrescriptionData struct {
	PrescriptionId            uint64               `json:"prescription_id"`
	PatientId                 uint64               `json:"patient_id"`
	PrescribedBy              string               `json:"prescribed_by"`
	PrescriptionName          string               `json:"prescription_name"`
	Description               string               `json:"description"`
	PrescriptionDate          string               `json:"prescription_date"`
	PrescriptionAttachmentUrl string               `json:"prescription_attachment_url"`
	PrescriptionDetails       []PrescriptionDetail `json:"prescription_details"`
}

type PrescriptionDetailData struct {
	PrescriptionDetailId uint64    `json:"prescription_detail_id"`
	PrescriptionId       uint64    `json:"prescription_id"`
	MedicineName         string    `json:"medicine_name"`
	PrescriptionType     string    `json:"prescription_type"`
	DoseQuantity         float64   `json:"dose_quantity"`
	Duration             int       `json:"duration"`
	UnitValue            float64   `json:"unit_value"`
	UnitType             string    `json:"unit_type"`
	Frequency            int       `json:"frequency"`
	TimesPerDay          int       `json:"times_per_day"`
	IntervalHour         int       `json:"interval_hour"`
	Instruction          string    `json:"instruction"`
	CreatedAt            time.Time `json:"created_at"`
}
