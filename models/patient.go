package models

import (
	"time"
)

type Patient struct {
	PatientId            uint      `gorm:"primaryKey;autoIncrement" json:"patient_id"`
	FirstName            string    `gorm:"type:varchar(255)" json:"first_name"`
	LastName             string    `gorm:"type:varchar(255)" json:"last_name"`
	DateOfBirth          string    `gorm:"type:date" json:"date_of_birth"`
	Gender               string    `gorm:"type:varchar(50)" json:"gender"`
	MobileNo             string    `gorm:"type:varchar(255)" json:"mobile_no"`
	Address              string    `gorm:"type:text" json:"address"`
	EmergencyContact     string    `gorm:"type:varchar(255)" json:"emergency_contact"`
	EmergencyContactName string    `gorm:"type:varchar(255)" json:"emergency_contact_name"`
	AbhaNumber           string    `gorm:"type:varchar(255)" json:"abha_number"`
	BloodGroup           string    `gorm:"type:varchar(50)" json:"blood_group"`
	Nationality          string    `gorm:"type:varchar(255)" json:"nationality"`
	CitizenshipStatus    string    `gorm:"type:varchar(255)" json:"citizenship_status"`
	PassportNumber       string    `gorm:"type:varchar(255)" json:"passport_number"`
	CountryOfResidence   string    `gorm:"type:varchar(255)" json:"country_of_residence"`
	IsIndianOrigin       bool      `gorm:"type:boolean" json:"is_indian_origin"`
	Email                string    `gorm:"type:varchar(255)" json:"email"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type PatientRelative struct {
	RelativeId   uint      `json:"relative_id" gorm:"primaryKey"`
	PatientId    uint      `json:"patient_id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Gender       string    `json:"gender"`
	DateOfBirth  string    `json:"date_of_birth"`
	Relationship string    `json:"relationship"`
	MobileNo     string    `json:"mobile_no"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (PatientRelative) TableName() string {
	return "tbl_patient_relative"
}

func (Patient) TableName() string {
	return "tbl_patient"
}

type PatientCustomRange struct {
	PatientDpCustomRangeId    uint      `gorm:"column:pdp_custom_range_id;primaryKey;autoIncrement" json:"pdp_custom_range_id"`
	PatientId                 uint      `gorm:"column:patient_id" json:"patient_id"`
	DiseaseProfileId          uint      `gorm:"column:disease_profile_id" json:"disease_profile_id"`
	DiagnosticTestId          uint      `gorm:"column:diagnostic_test_id" json:"diagnostic_test_id"`
	DiagnosticTestComponentId uint      `gorm:"column:diagnostic_test_component_id" json:"diagnostic_test_component_id"`
	NormalMin                 float64   `gorm:"column:normal_min" json:"normal_min"`
	NormalMax                 float64   `gorm:"column:normal_max" json:"normal_max"`
	Unit                      string    `gorm:"column:unit" json:"unit"`
	CustomFrequency           string    `gorm:"column:custom_frequency" json:"custom_frequency"`
	CustomFrequencyUnit       string    `gorm:"column:custom_frequency_unit" json:"custom_frequency_unit"`
	CreatedAt                 time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                 time.Time `gorm:"column:updated_at" json:"updated_at"`

	DiagnosticTest DiagnosticTest `gorm:"foreignKey:DiagnosticTestId" json:"diagnostic_test"`
}

func (PatientCustomRange) TableName() string {
	return "tbl_patient_dp_custom_range"
}
