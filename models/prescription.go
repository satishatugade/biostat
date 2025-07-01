package models

import (
	"time"
)

type PatientPrescription struct {
	PrescriptionId   uint64     `gorm:"primaryKey;column:prescription_id" json:"prescription_id"`
	PatientId        uint64     `gorm:"column:patient_id" json:"patient_id"`
	RecordId         uint64     `json:"record_id"`
	PrescribedBy     string     `gorm:"column:prescribed_by" json:"prescribed_by"`
	PrescriptionName *string    `gorm:"column:prescription_name" json:"prescription_name"`
	Description      string     `gorm:"column:description" json:"description"`
	PrescriptionDate *time.Time `gorm:"column:prescription_date" json:"prescription_date"`
	IsDigital        bool       `gorm:"column:is_digital;default:false" json:"is_digital"`
	StartDate        *time.Time `gorm:"column:prescription_start_date" json:"prescription_start_date"`
	EndDate          *time.Time `gorm:"column:prescription_end_date" json:"prescription_end_date"`

	// Relationship to PrescriptionDetail
	PrescriptionDetails []PrescriptionDetail `gorm:"foreignKey:PrescriptionId;references:PrescriptionId" json:"prescription_details"`
	MedicalRecord       TblMedicalRecord     `gorm:"foreignKey:RecordId;references:RecordId" json:"prescription_attachment"`
}

func (PatientPrescription) TableName() string {
	return "tbl_patient_prescription"
}

type PrescriptionDetail struct {
	PrescriptionDetailId   uint64                     `gorm:"primaryKey;autoIncrement;column:prescription_detail_id" json:"prescription_detail_id"`
	PrescriptionId         uint64                     `gorm:"column:prescription_id" json:"prescription_id"`
	MedicineName           string                     `gorm:"column:medicine_name" json:"medicine_name"`
	PrescriptionType       string                     `gorm:"column:prescription_type" json:"prescription_type"`
	Duration               int                        `gorm:"column:duration" json:"duration"`
	DurationUnitType       string                     `gorm:"column:duration_unit_type" json:"duration_unit_type"`
	DoseQuantity           float64                    `gorm:"-" json:"dose_quantity,omitempty"`
	UnitValue              float64                    `gorm:"-" json:"unit_value,omitempty"`
	UnitType               string                     `gorm:"-" json:"unit_type,omitempty"`
	MedUnit                string                     `gorm:"column:med_unit" json:"med_unit"`
	MedUnitValue           float64                    `gorm:"column:med_unit_value" json:"med_unit_value"`
	Instruction            string                     `gorm:"-" json:"instruction,omitempty"`
	DosageInfo             []PrescriptionDoseSchedule `gorm:"foreignKey:PrescriptionDetailId;references:PrescriptionDetailId" json:"dosage_info"`
	PrescriptionAttachment TblMedicalRecord           `gorm:"-" json:"prescription_attachment"`
}

func (PrescriptionDetail) TableName() string {
	return "tbl_prescription_detail"
}

type PrescriptionDoseSchedule struct {
	DoseScheduleId       uint64  `gorm:"primaryKey;autoIncrement;column:dose_schedule_id" json:"dose_schedule_id"`
	PrescriptionDetailId uint64  `gorm:"column:prescription_detail_id;not null" json:"prescription_detail_id"`
	TimeOfDay            string  `gorm:"column:time_of_day;type:varchar(50)" json:"time_of_day"`
	IsGiven              int     `gorm:"column:is_given;default:0" json:"is_given"`
	DoseQuantity         float64 `gorm:"column:dose_quantity;type:numeric(10,2)" json:"dose_quantity"`
	UnitValue            float64 `gorm:"column:unit_value" json:"unit_value"`
	UnitType             string  `gorm:"column:unit_type" json:"unit_type"`
	Instruction          string  `gorm:"column:instruction;type:text" json:"instruction"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedBy string    `gorm:"column:created_by;type:varchar(100)" json:"created_by"`
	UpdatedBy string    `gorm:"column:updated_by;type:varchar(100)" json:"updated_by"`
}

func (PrescriptionDoseSchedule) TableName() string {
	return "tbl_prescription_dose_schedule"
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

type UserMedicineInfo struct {
	PrescriptionDetailID int64  `json:"prescription_detail_id"`
	PrescriptionID       int64  `json:"prescription_id"`
	MedicineName         string `json:"medicine_name"`
	PrescriptionType     string `json:"prescription_type"`
	Duration             int    `json:"duration"`
	DurationUnitType     string `json:"duration_unit_type"`
}

// PatientInfo holds patient demographic details
type PatientInfoData struct {
	Doctor     string
	Name       string
	Phone      string
	ReportDate string
	DOB        string
}

// LabInfo holds laboratory details
type LabInfoData struct {
	Name    string
	Address string
	Phone   string
	Email   string
}

// TestResult represents a single test component's result
type TestResult struct {
	TestComponentName string
	Unit              string
	RefRange          string
	TrendValues       []Cell
}

// CellData represents a single trend value with its date
type Cell struct {
	ResultDate string
	Value      string
	IsNormal   bool // true if within ref range, false otherwise
}

// ReportData combines all data needed for the PDF report
type ReportData struct {
	Patient     PatientInfoData
	Lab         LabInfoData
	TestResults []TestResult
	Dates       []string // All unique dates for trend values
}
