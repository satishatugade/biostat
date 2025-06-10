package models

import (
	"time"
)

type Patient struct {
	PatientId            uint64        `json:"patient_id"`
	FirstName            string        `json:"first_name"`
	LastName             string        `json:"last_name"`
	DateOfBirth          *time.Time    `json:"date_of_birth"`
	Gender               string        `json:"gender"`
	MobileNo             string        `json:"mobile_no"`
	MaritalStatus        string        `json:"marital_status"`
	Address              string        `json:"address"`
	UserAddress          AddressMaster `gorm:"-" json:"user_address"`
	EmergencyContact     string        `json:"emergency_contact"`
	EmergencyContactName string        `json:"emergency_contact_name"`
	AbhaNumber           string        `json:"abha_number"`
	BloodGroup           string        `json:"blood_group"`
	Nationality          string        `json:"nationality"`
	CitizenshipStatus    string        `json:"citizenship_status"`
	PassportNumber       string        `json:"passport_number"`
	CountryOfResidence   string        `json:"country_of_residence"`
	IsIndianOrigin       bool          `json:"is_indian_origin"`
	Email                string        `json:"email"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
}

type PatientRelation struct {
	RelationId   uint64 `json:"relation_id" gorm:"column:relation_id"`
	RelationShip string `json:"relationship" gorm:"column:relationship"`
}

func (PatientRelation) TableName() string {
	return "tbl_relation_master"
}

type PatientRelative struct {
	RelativeId   uint64             `json:"relative_id"`
	PatientId    *uint              `json:"patient_id,omitempty"`
	FirstName    string             `json:"first_name"`
	LastName     string             `json:"last_name"`
	Gender       string             `json:"gender"`
	MappingType  string             `json:"mapping_type"`
	DateOfBirth  string             `json:"date_of_birth"`
	RelationId   uint64             `json:"relation_id"`
	Relationship string             `json:"relationship"`
	MobileNo     string             `json:"mobile_no"`
	Email        string             `json:"email"`
	Permissions  []PermissionResult `json:"permissions" gorm:"-"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
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
