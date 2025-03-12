package models

import "time"

type DiseaseProfileDiagnosticTestMasterAudit struct {
	ID               uint      `gorm:"primaryKey;column:diagnostic_test_audit_id"`
	DiagnosticTestID uint      `gorm:"column:diagnostic_test_id"`
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
