package models

import "time"

type DiseaseAudit struct {
	DiseaseAuditId    uint64    `json:"disease_audit_id" gorm:"primaryKey"`
	DiseaseId         uint64    `json:"disease_id"`
	DiseaseSnomedCode string    `json:"disease_snomed_code"`
	DiseaseName       string    `json:"disease_name"`
	Description       string    `json:"description"`
	ImageURL          string    `json:"image_url"`
	SlugURL           string    `json:"slug_url"`
	OperationType     string    `json:"operation_type"` // "Update" or "Delete"
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	CreatedBy         string    `json:"created_by"`
	UpdatedBy         string    `json:"updated_by"`
}

// Table Name
func (DiseaseAudit) TableName() string {
	return "tbl_disease_master_audit"
}

type DiseaseProfileDiagnosticTestMasterAudit struct {
	ID               uint64    `gorm:"primaryKey;column:diagnostic_test_audit_id"`
	DiagnosticTestID uint64    `gorm:"column:diagnostic_test_id"`
	TestLoincCode    string    `gorm:"column:test_loinc_code"`
	TestName         string    `gorm:"column:test_name"`
	TestDescription  string    `gorm:"column:test_description"`
	Category         string    `gorm:"column:category"`
	Units            string    `gorm:"column:units"`
	Property         string    `gorm:"column:property"`
	TimeAspect       string    `gorm:"column:time_aspect"`
	System           string    `gorm:"column:system"`
	Scale            string    `gorm:"column:scale"`
	Method           string    `gorm:"column:method"`
	OperationType    string    `gorm:"column:operation_type"`
	UpdatedBy        string    `gorm:"column:updated_by"`
	ModifiedOn       time.Time `gorm:"column:modified_on;autoCreateTime"`
}

func (DiseaseProfileDiagnosticTestMasterAudit) TableName() string {
	return "tbl_disease_profile_diagnostic_test_master_audit"
}

type CauseAudit struct {
	CauseAuditId  uint64    `json:"cause_audit_id" gorm:"primaryKey;autoIncrement"`
	CauseId       uint64    `json:"cause_id" gorm:"primaryKey"`
	CauseName     string    `json:"cause_name"`
	CauseType     string    `json:"cause_type"`
	Description   string    `json:"description"`
	OperationType string    `json:"operation_type"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	CreatedBy     string    `json:"created_by"`
	UpdatedBy     string    `json:"updated_by"`
}

func (CauseAudit) TableName() string {
	return "tbl_cause_master_audit"
}

type SymptomAudit struct {
	SymptomAuditId uint64    `json:"symptom_audit_id" gorm:"primaryKey"`
	SymptomId      uint64    `json:"symptom_id"`
	SymptomName    string    `json:"symptom_name"`
	SymptomType    string    `json:"symptom_type"`
	Commonality    string    `json:"commonality"`
	Description    string    `json:"description"`
	OperationType  string    `json:"operation_type"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	CreatedBy      string    `json:"created_by"`
	UpdatedBy      string    `json:"updated_by"`
}

func (SymptomAudit) TableName() string {
	return "tbl_symptom_master_audit"
}
