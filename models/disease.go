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
	// Medications      []Medication `json:"medications" gorm:"many2many:tbl_disease_medication_mapping;foreignKey:DiseaseId;joinForeignKey:DiseaseId;References:MedicationId;joinReferences:MedicationId"`
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
	Medications        []Medication       `json:"medications" gorm:"many2many:tbl_disease_medication_mapping;foreignKey:DiseaseId;joinForeignKey:DiseaseId;References:MedicationId;joinReferences:MedicationId"`
	Exercises          []Exercise         `json:"exercise_recommendations" gorm:"many2many:tbl_disease_exercise_mapping;foreignKey:DiseaseId;joinForeignKey:DiseaseId;References:ExerciseId;joinReferences:ExerciseId"`
	DietPlans          []DietPlanTemplate `json:"diet_recommendations" gorm:"many2many:tbl_disease_diet_mapping;foreignKey:DiseaseId;joinForeignKey:DiseaseId;References:DietPlanTemplateId;joinReferences:DietPlanTemplateId"`
	DiagnosticTests    []DiagnosticTest   `json:"diagnostic_tests" gorm:"many2many:tbl_disease_diagnostic_test_mapping;joinForeignKey:DiseaseId;joinReferences:DiagnosticTestId"`
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

func (DiseaseProfile) TableName() string { return "tbl_disease_profile" }
func (Cause) TableName() string          { return "tbl_cause_master" }
func (Severity) TableName() string       { return "tbl_severity_master" }
func (Symptom) TableName() string        { return "tbl_symptom_master" }
func (Disease) TableName() string        { return "tbl_disease_master" }
func (DiseaseType) TableName() string    { return "tbl_disease_type_master" }
func (DiseaseTypeMapping) TableName() string {
	return "tbl_disease_type_mapping"
}

type Medication struct {
	MedicationId    uint             `json:"medication_id" gorm:"column:medication_id;primaryKey"`
	MedicationName  string           `json:"medication_name" gorm:"column:medication_name"`
	MedicationCode  string           `json:"medication_code" gorm:"column:medication_code"`
	Description     string           `json:"description" gorm:"column:description"`
	CreatedAt       time.Time        `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time        `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	MedicationTypes []MedicationType `json:"medication_type" gorm:"foreignKey:MedicationId;references:MedicationId"`
}

// Table name override
func (Medication) TableName() string {
	return "tbl_medication_master"
}

type MedicationType struct {
	DosageId           uint    `json:"dosage_id" gorm:"column:dosage_id;primaryKey"`
	MedicationId       uint    `json:"medication_id" gorm:"column:medication_id"`
	MedicationType     string  `json:"medication_type" gorm:"column:medication_type"`
	UnitValue          float64 `json:"unit_value" gorm:"column:unit_value"`
	UnitType           string  `json:"unit_type" gorm:"column:unit_type"`
	MedicationCost     float64 `json:"medication_cost" gorm:"column:medication_cost"`
	MedicationImageURL string  `json:"medication_image_url" gorm:"column:medication_image_url"`
}

// Table name override
func (MedicationType) TableName() string {
	return "tbl_medication_type"
}

type DiseaseMedicationMapping struct {
	DiseaseMedicationMappingId uint `json:"-" gorm:"primaryKey"`
	DiseaseId                  uint `json:"disease_id"`
	MedicationId               uint `json:"medication_id"`
}

func (DiseaseMedicationMapping) TableName() string {
	return "tbl_disease_medication_mapping"
}

type Exercise struct {
	ExerciseId     uint      `json:"exercise_id" gorm:"primaryKey"`
	ExerciseName   string    `json:"exercise_name"`
	Description    string    `json:"description"`
	Category       string    `json:"category"`
	IntensityLevel string    `json:"intensity_level"`
	Duration       int       `json:"duration"`
	DurationUnit   string    `json:"duration_unit"`
	Benefits       string    `json:"benefits"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DiseaseExerciseMapping struct {
	DiseaseExerciseMappingId uint      `json:"-" gorm:"primaryKey"`
	DiseaseId                uint      `json:"-"`
	ExerciseId               uint      `json:"-"`
	Exercise                 Exercise  `json:"exercise" gorm:"foreignKey:ExerciseId;references:ExerciseId"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

func (Exercise) TableName() string {
	return "tbl_exercise_master"
}

func (DiseaseExerciseMapping) TableName() string {
	return "tbl_disease_exercise_mapping"
}

type DietPlanTemplate struct {
	DietPlanTemplateId uint      `json:"diet_plan_template_id" gorm:"primaryKey"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Goal               string    `json:"goal"`
	Notes              string    `json:"notes"`
	DietCreatorId      uint      `json:"-"`
	Cost               float64   `json:"-"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (DietPlanTemplate) TableName() string {
	return "tbl_diet_plan_template"
}

type DiagnosticTest struct {
	DiagnosticTestId uint                      `gorm:"column:diagnostic_test_id;primaryKey" json:"diagnostic_test_id"`
	LoincCode        string                    `gorm:"column:test_loinc_code" json:"test_loinc_code"`
	Name             string                    `gorm:"column:test_name" json:"test_name"`
	Type             string                    `gorm:"column:test_type" json:"test_type"`
	Description      string                    `gorm:"column:test_description" json:"test_description"`
	Category         string                    `gorm:"column:category" json:"category"`
	Units            string                    `gorm:"column:units" json:"units"`
	Property         string                    `gorm:"column:property" json:"property"`
	TimeAspect       string                    `gorm:"column:time_aspect" json:"time_aspect"`
	System           string                    `gorm:"column:system" json:"system"`
	Scale            string                    `gorm:"column:scale" json:"scale"`
	CreatedAt        time.Time                 `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time                 `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	Components       []DiagnosticTestComponent `gorm:"many2many:tbl_disease_profile_diagnostic_test_component_mapping;foreignKey:DiagnosticTestId;joinForeignKey:DiagnosticTestId;References:DiagnosticTestComponentId;joinReferences:DiagnosticTestComponentId" json:"test_components"`
}

type DiagnosticTestComponent struct {
	DiagnosticTestComponentId uint      `gorm:"column:diagnostic_test_component_id;primaryKey" json:"diagnostic_test_component_id"`
	LoincCode                 string    `gorm:"column:test_component_loinc_code" json:"test_component_loinc_code"`
	Name                      string    `gorm:"column:test_component_name" json:"test_component_name"`
	Type                      string    `gorm:"column:test_component_type" json:"test_component_type"`
	Description               string    `gorm:"column:description" json:"description"`
	Units                     string    `gorm:"column:units" json:"units"`
	Property                  string    `gorm:"column:property" json:"property"`
	TimeAspect                string    `gorm:"column:time_aspect" json:"time_aspect"`
	System                    string    `gorm:"column:system" json:"system"`
	Scale                     string    `gorm:"column:scale" json:"scale"`
	TestComponentFrequency    int       `gorm:"column:test_component_frequency" json:"test_component_frequency"`
	CreatedAt                 time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt                 time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

type DiagnosticTestComponentMapping struct {
	DiagnosticTestComponentMappingId uint `gorm:"column:diagnostic_test_component_mapping_id;primaryKey" json:"diagnostic_test_component_mapping_id"`
	DiagnosticTestId                 uint `gorm:"column:diagnostic_test_id" json:"diagnostic_test_id"`
	DiagnosticComponentId            uint `gorm:"column:diagnostic_test_component_id" json:"diagnostic_test_component_id"`
}

func (DiagnosticTestComponent) TableName() string {
	return "tbl_disease_profile_diagnostic_test_component_master"
}

func (DiagnosticTest) TableName() string {
	return "tbl_disease_profile_diagnostic_test_master"
}

func (DiagnosticTestComponentMapping) TableName() string {
	return "tbl_disease_profile_diagnostic_test_component_mapping"
}

type DiseaseDiagnosticTestMapping struct {
	DiseaseDiagnosticTestMapping uint `gorm:"column:disease_diagnostic_test_mapping_id;primaryKey"`
	DiseaseId                    uint `gorm:"column:disease_id"`
	DiagnosticTestId             uint `gorm:"column:diagnostic_test_id"`
}

func (DiseaseDiagnosticTestMapping) TableName() string {
	return "tbl_disease_diagnostic_test_mapping"
}
