package models

import (
	"time"
)

// PatientAllergyRestriction represents tbl_patient_allergy_restriction
type PatientAllergyRestriction struct {
	PatientAllergyRestrictionId uint64    `gorm:"primaryKey" json:"patient_allergy_restriction_id"`
	PatientId                   uint64    `gorm:"column:patient_id" json:"patient_id"`
	AllergyId                   uint64    `gorm:"column:allergy_id" json:"allergy_id"`
	SeverityId                  uint64    `gorm:"column:severity_id" json:"severity_id"`
	Reaction                    string    `gorm:"column:reaction" json:"reaction"`
	Description                 string    `gorm:"column:description" json:"description"`
	CreatedAt                   time.Time `gorm:"column:created_at" json:"created_at"`

	Severity Severity `gorm:"foreignKey:SeverityId" json:"severity"`
	Allergy  Allergy  `gorm:"foreignKey:AllergyId" json:"allergy"`
}

func (PatientAllergyRestriction) TableName() string {
	return "tbl_patient_allergy_restriction"
}

// Allergy represents tbl_allergy
type Allergy struct {
	AllergyId     uint   `gorm:"primaryKey" json:"allergy_id"`
	AllergyName   string `gorm:"column:allergy_name" json:"allergy_name"`
	AllergyTypeId uint   `gorm:"column:allergy_type_id" json:"allergy_type_id"`

	AllergyType AllergyType `gorm:"foreignKey:AllergyTypeId" json:"allergy_type"`
}

func (Allergy) TableName() string {
	return "tbl_allergy"
}

// AllergyType represents tbl_allergy_type
type AllergyType struct {
	AllergyTypeId uint   `gorm:"primaryKey" json:"allergy_type_id"`
	AllergyType   string `gorm:"column:allergy_type" json:"allergy_type"`
}

func (AllergyType) TableName() string {
	return "tbl_allergy_type"
}
