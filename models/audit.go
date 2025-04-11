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
	OperationType     string    `json:"operation_type"`
	IsDeleted         int       `json:"is_deleted"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	CreatedBy         string    `json:"created_by"`
	UpdatedBy         string    `json:"updated_by"`
}

// Table Name
func (DiseaseAudit) TableName() string {
	return "tbl_disease_master_audit"
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
	IsDeleted     int       `json:"is_deleted"`
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
	IsDeleted      int       `json:"is_deleted"`
}

func (SymptomAudit) TableName() string {
	return "tbl_symptom_master_audit"
}

type MedicationAudit struct {
	MedicationAuditId uint64    `json:"medication_audit_id" gorm:"column:medication_audit_id;primaryKey"`
	MedicationId      uint64    `json:"medication_id" gorm:"column:medication_id"`
	MedicationName    string    `json:"medication_name" gorm:"column:medication_name"`
	MedicationCode    string    `json:"medication_code" gorm:"column:medication_code"`
	Description       string    `json:"description" gorm:"column:description"`
	OperationType     string    `json:"operation_type" gorm:"column:operation_type"`
	IsDeleted         int       `json:"is_deleted" gorm:"column:is_deleted"`
	CreatedAt         time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	CreatedBy         string    `json:"created_by" gorm:"column:created_by"`
	UpdatedBy         string    `json:"updated_by" gorm:"column:updated_by"`
}

func (MedicationAudit) TableName() string {
	return "tbl_medication_master_audit"
}

type MedicationTypeAudit struct {
	MedicationTypeAuditId uint64  `json:"medication_type_audit_id" gorm:"column:medication_type_audit_id;primaryKey"`
	DosageId              uint64  `json:"dosage_id" gorm:"column:dosage_id"`
	MedicationId          uint64  `json:"medication_id" gorm:"column:medication_id"`
	MedicationType        string  `json:"medication_type" gorm:"column:medication_type"`
	UnitValue             float64 `json:"unit_value" gorm:"column:unit_value"`
	UnitType              string  `json:"unit_type" gorm:"column:unit_type"`
	MedicationCost        float64 `json:"medication_cost" gorm:"column:medication_cost"`
	MedicationImageURL    string  `json:"medication_image_url" gorm:"column:medication_image_url"`

	OperationType string `json:"operation_type" gorm:"column:operation_type"`
	IsDeleted     int    `json:"is_deleted" gorm:"column:is_deleted"`

	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	UpdatedBy string    `json:"updated_by" gorm:"column:updated_by"`
}

func (MedicationTypeAudit) TableName() string {
	return "tbl_medication_type_audit"
}

type ExerciseAudit struct {
	ExerciseAuditId uint64    `json:"exercise_audit_id" gorm:"primaryKey;column:exercise_audit_id"`
	ExerciseId      uint64    `json:"exercise_id" gorm:"column:exercise_id"`
	ExerciseName    string    `json:"exercise_name" gorm:"column:exercise_name"`
	Description     string    `json:"description" gorm:"column:description"`
	Category        string    `json:"category" gorm:"column:category"`
	IntensityLevel  string    `json:"intensity_level" gorm:"column:intensity_level"`
	Duration        int       `json:"duration" gorm:"column:duration"`
	DurationUnit    string    `json:"duration_unit" gorm:"column:duration_unit"`
	Benefits        string    `json:"benefits" gorm:"column:benefits"`
	IsDeleted       int       `json:"is_deleted" gorm:"column:is_deleted"`
	OperationType   string    `json:"operation_type" gorm:"column:operation_type"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt       time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
	CreatedBy       string    `json:"created_by" gorm:"column:created_by"`
	UpdatedBy       string    `json:"updated_by" gorm:"column:updated_by"`
}

func (ExerciseAudit) TableName() string {
	return "tbl_exercise_master_audit"
}

type DiagnosticTestAudit struct {
	DiagnosticTestAuditId uint64    `gorm:"column:diagnostic_test_audit_id;primaryKey" json:"diagnostic_test_audit_id"`
	DiagnosticTestId      uint64    `gorm:"column:diagnostic_test_id" json:"diagnostic_test_id"`
	TestLoincCode         string    `gorm:"column:test_loinc_code" json:"test_loinc_code"`
	TestName              string    `gorm:"column:test_name" json:"test_name"`
	TestType              string    `gorm:"column:test_type" json:"test_type"`
	TestDescription       string    `gorm:"column:test_description" json:"test_description"`
	Category              string    `gorm:"column:category" json:"category"`
	Units                 string    `gorm:"column:units" json:"units"`
	Property              string    `gorm:"column:property" json:"property"`
	TimeAspect            string    `gorm:"column:time_aspect" json:"time_aspect"`
	System                string    `gorm:"column:system" json:"system"`
	Scale                 string    `gorm:"column:scale" json:"scale"`
	OperationType         string    `gorm:"column:operation_type" json:"operation_type"`
	IsDeleted             int       `gorm:"column:is_deleted" json:"is_deleted"`
	CreatedAt             time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedBy             string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy             string    `gorm:"column:updated_by" json:"updated_by"`
}

func (DiagnosticTestAudit) TableName() string {
	return "tbl_disease_profile_diagnostic_test_master_audit"
}

type DiagnosticTestComponentAudit struct {
	DiagnosticTestComponentAuditId uint64    `gorm:"column:diagnostic_test_component_audit_id;primaryKey;autoIncrement" json:"diagnostic_test_component_audit_id"`
	DiagnosticTestComponentId      uint64    `gorm:"column:diagnostic_test_component_id" json:"diagnostic_test_component_id"`
	TestComponentLoincCode         string    `gorm:"column:test_component_loinc_code" json:"test_component_loinc_code"`
	TestComponentName              string    `gorm:"column:test_component_name" json:"test_component_name"`
	TestComponentType              string    `gorm:"column:test_component_type" json:"test_component_type"`
	Description                    string    `gorm:"column:description" json:"description"`
	Units                          string    `gorm:"column:units" json:"units"`
	Property                       string    `gorm:"column:property" json:"property"`
	TimeAspect                     string    `gorm:"column:time_aspect" json:"time_aspect"`
	System                         string    `gorm:"column:system" json:"system"`
	Scale                          string    `gorm:"column:scale" json:"scale"`
	TestComponentFrequency         string    `gorm:"column:test_component_frequency" json:"test_component_frequency"`
	OperationType                  string    `gorm:"column:operation_type" json:"operation_type"` // e.g., INSERT, UPDATE, DELETE
	IsDeleted                      int       `gorm:"column:is_deleted" json:"is_deleted"`
	CreatedAt                      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                      time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedBy                      string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy                      string    `gorm:"column:updated_by" json:"updated_by"`
}

func (DiagnosticTestComponentAudit) TableName() string {
	return "tbl_disease_profile_diagnostic_test_component_master_audit"
}
