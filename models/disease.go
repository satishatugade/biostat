// package models

// import "time"

// type DiseaseProfile struct {
// 	DiseaseProfileId uint      `json:"disease_profile_id" gorm:"primaryKey"`
// 	DiseaseId        uint      `json:"disease_id"`
// 	CreatedAt        time.Time `json:"created_at"`
// 	UpdatedAt        time.Time `json:"updated_at"`
// 	Disease          Disease   `json:"disease" gorm:"foreignKey:DiseaseId;references:DiseaseId"`
// }

// type Disease struct {
// 	DiseaseId          uint               `json:"disease_id" gorm:"primaryKey"`
// 	DiseaseSnomedCode  string             `json:"disease_snomed_code"`
// 	DiseaseName        string             `json:"disease_name"`
// 	Description        string             `json:"description"`
// 	ImageURL           string             `json:"image_url"`
// 	SlugURL            string             `json:"slug_url"`
// 	CreatedAt          time.Time          `json:"created_at"`
// 	UpdatedAt          time.Time          `json:"updated_at"`
// 	DiseaseType        DiseaseType        `json:"disease_type" gorm:"-"`
// 	Severity           Severity           `json:"severity_levels" gorm:"many2many:tbl_disease_severity_mapping;foreignKey:DiseaseId;joinForeignKey:DiseaseId;References:SeverityId;joinReferences:SeverityId"`
// 	Symptoms           []Symptom          `json:"symptoms" gorm:"many2many:tbl_disease_symptom_mapping;foreignKey:DiseaseId;joinForeignKey:DiseaseId;References:SymptomId;joinReferences:SymptomId"`
// 	Causes             []Cause            `json:"causes" gorm:"many2many:tbl_disease_cause_mapping;foreignKey:DiseaseId;joinForeignKey:DiseaseId;References:CauseId;joinReferences:CauseId"`
// 	DiseaseTypeMapping DiseaseTypeMapping `json:"-" gorm:"foreignKey:DiseaseId;references:DiseaseId"`
// }

// type DiseaseType struct {
// 	DiseaseTypeId uint   `json:"disease_type_id" gorm:"primaryKey"`
// 	DiseaseType   string `json:"disease_type"`
// }

// type DiseaseTypeMapping struct {
// 	DiseaseTypeMappingId uint        `json:"-" gorm:"primaryKey"`
// 	DiseaseId            uint        `json:"-"`
// 	DiseaseTypeId        uint        `json:"-"`
// 	DiseaseType          DiseaseType `json:"disease_type" gorm:"foreignKey:DiseaseTypeId;references:DiseaseTypeId"`
// }

// type Symptom struct {
// 	SymptomId   uint   `json:"symptom_id" gorm:"primaryKey"`
// 	SymptomName string `json:"symptom_name"`
// 	SymptomType string `json:"symptom_type"`
// 	Commonality string `json:"commonality"`
// 	Description string `json:"description"`
// }

// type Severity struct {
// 	SeverityId    uint   `json:"severity_id" gorm:"primaryKey"`
// 	SeverityLevel string `json:"severity_level"`
// }

// type Cause struct {
// 	CauseId     uint   `json:"cause_id" gorm:"primaryKey"`
// 	CauseName   string `json:"cause_name"`
// 	CauseType   string `json:"cause_type"`
// 	Description string `json:"description"`
// }

// func (DiseaseProfile) TableName() string { return "tbl_disease_profile" }

// func (Cause) TableName() string {
// 	return "tbl_cause_master"
// }

// func (Severity) TableName() string {
// 	return "tbl_severity_master"
// }

// func (Symptom) TableName() string {
// 	return "tbl_symptom_master"
// }

// func (Disease) TableName() string {
// 	return "tbl_disease_master"
// }

// func (DiseaseType) TableName() string {
// 	return "tbl_disease_type_master"
// }

// func (DiseaseTypeMapping) TableName() string {
// 	return "tbl_disease_type_mapping"
// }

package models

import "time"

type DiseaseProfile struct {
	DiseaseProfileId uint      `json:"disease_profile_id" gorm:"primaryKey"`
	DiseaseId        uint      `json:"disease_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Disease          Disease   `json:"disease" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Disease struct {
	DiseaseId          uint               `json:"disease_id" gorm:"primaryKey"`
	DiseaseSnomedCode  string             `json:"disease_snomed_code"`
	DiseaseName        string             `json:"disease_name"`
	Description        string             `json:"description"`
	ImageURL           string             `json:"image_url"`
	SlugURL            string             `json:"slug_url"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	DiseaseType        *DiseaseType       `json:"disease_type" gorm:"-"`
	Severity           []Severity         `json:"severity_levels" gorm:"many2many:tbl_disease_severity_mapping;joinForeignKey:DiseaseId;joinReferences:SeverityId"`
	Symptoms           []Symptom          `json:"symptoms" gorm:"many2many:tbl_disease_symptom_mapping;joinForeignKey:DiseaseId;joinReferences:SymptomId"`
	Causes             []Cause            `json:"causes" gorm:"many2many:tbl_disease_cause_mapping;joinForeignKey:DiseaseId;joinReferences:CauseId"`
	DiseaseTypeMapping DiseaseTypeMapping `json:"-" gorm:"foreignKey:DiseaseId;references:DiseaseId"`
}

type DiseaseType struct {
	DiseaseTypeId uint   `json:"disease_type_id" gorm:"primaryKey"`
	DiseaseType   string `json:"disease_type"`
}

type DiseaseTypeMapping struct {
	DiseaseTypeMappingId uint        `json:"-" gorm:"primaryKey"`
	DiseaseId            uint        `json:"disease_id" gorm:"index"`
	DiseaseTypeId        uint        `json:"disease_type_id"`
	DiseaseType          DiseaseType `json:"disease_type" gorm:"foreignKey:DiseaseTypeId;references:DiseaseTypeId"`
}

type Symptom struct {
	SymptomId   uint   `json:"symptom_id" gorm:"primaryKey"`
	SymptomName string `json:"symptom_name"`
	SymptomType string `json:"symptom_type"`
	Commonality string `json:"commonality"`
	Description string `json:"description"`
}

type Severity struct {
	SeverityId    uint   `json:"severity_id" gorm:"primaryKey"`
	SeverityLevel string `json:"severity_level"`
}

type Cause struct {
	CauseId     uint   `json:"cause_id" gorm:"primaryKey"`
	CauseName   string `json:"cause_name"`
	CauseType   string `json:"cause_type"`
	Description string `json:"description"`
}

// Table Names
func (DiseaseProfile) TableName() string { return "tbl_disease_profile" }
func (Cause) TableName() string          { return "tbl_cause_master" }
func (Severity) TableName() string       { return "tbl_severity_master" }
func (Symptom) TableName() string        { return "tbl_symptom_master" }
func (Disease) TableName() string        { return "tbl_disease_master" }
func (DiseaseType) TableName() string    { return "tbl_disease_type_master" }
func (DiseaseTypeMapping) TableName() string {
	return "tbl_disease_type_mapping"
}
