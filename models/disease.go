package models

import (
	"time"
)

type PatientDiseaseProfile struct {
	PatientDiseaseProfileId uint64    `gorm:"primaryKey;autoIncrement" json:"patient_disease_profile_id"`
	PatientId               uint64    `gorm:"not null" json:"patient_id"`
	DiseaseProfileId        uint64    `gorm:"not null" json:"disease_profile_id"`
	ReminderFlag            *bool     `json:"reminder_flag"`        // Nullable
	DietPlanSubscribed      *bool     `json:"diet_plan_subscribed"` // Nullable
	AttachedDate            time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"attached_date"`
	UpdatedAt               time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	AttachedFlag            int       `json:"attached_flag"`
	// Relations
	DiseaseProfile DiseaseProfile `gorm:"foreignKey:DiseaseProfileId" json:"disease_profile"`
}

type DiseaseProfile struct {
	DiseaseProfileId uint64    `json:"disease_profile_id" gorm:"primaryKey"`
	DiseaseId        uint64    `json:"disease_id"`
	IsDeleted        int       `json:"is_deleted"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	CreatedBy        string    `json:"created_by"`
	UpdatedBy        string    `json:"updated_by"`
	Disease          Disease   `json:"disease" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Disease struct {
	DiseaseId         uint64    `json:"disease_id" gorm:"primaryKey"`
	DiseaseTypeId     uint64    `json:"disease_type_id" gorm:"-"`
	DiseaseSnomedCode string    `json:"disease_snomed_code"`
	DiseaseName       string    `json:"disease_name"`
	Description       string    `json:"description"`
	ImageURL          string    `json:"image_url"`
	SlugURL           string    `json:"slug_url"`
	IsDeleted         int       `json:"is_deleted"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	CreatedBy         string    `json:"created_by"`

	DiseaseTypeMapping DiseaseTypeMapping `json:"-" gorm:"foreignKey:DiseaseId;references:DiseaseId"`
	DiseaseType        DiseaseType        `json:"disease_type" gorm:"-"`
	Severity           []Severity         `json:"severity_levels" gorm:"many2many:tbl_disease_severity_mapping;joinForeignKey:DiseaseId;joinReferences:SeverityId"`
	Symptoms           []Symptom          `json:"symptoms" gorm:"many2many:tbl_disease_symptom_mapping;joinForeignKey:DiseaseId;joinReferences:SymptomId"`
	Causes             []Cause            `json:"causes" gorm:"many2many:tbl_disease_cause_mapping;joinForeignKey:DiseaseId;joinReferences:CauseId"`
	Medications        []Medication       `json:"medications" gorm:"many2many:tbl_disease_medication_mapping;foreignKey:DiseaseId;joinForeignKey:DiseaseId;References:MedicationId;joinReferences:MedicationId"`
	Exercises          []Exercise         `json:"exercise_recommendations" gorm:"many2many:tbl_disease_exercise_mapping;foreignKey:DiseaseId;joinForeignKey:DiseaseId;References:ExerciseId;joinReferences:ExerciseId"`
	DietPlans          []DietPlanTemplate `json:"diet_recommendations" gorm:"many2many:tbl_disease_diet_mapping;foreignKey:DiseaseId;joinForeignKey:DiseaseId;References:DietPlanTemplateId;joinReferences:DietPlanTemplateId"`
	DiagnosticTests    []DiagnosticTest   `json:"diagnostic_tests" gorm:"many2many:tbl_disease_diagnostic_test_mapping;joinForeignKey:DiseaseId;joinReferences:DiagnosticTestId"`
}

type DiseaseType struct {
	DiseaseTypeId uint64 `json:"disease_type_id" gorm:"primaryKey"`
	DiseaseType   string `json:"disease_type"`
}

type DiseaseTypeMapping struct {
	DiseaseTypeMappingId uint64      `json:"-" gorm:"primaryKey"`
	DiseaseId            uint64      `json:"-" gorm:"index"`
	DiseaseTypeId        uint64      `json:"-"`
	DiseaseType          DiseaseType `json:"disease_type" gorm:"foreignKey:DiseaseTypeId;references:DiseaseTypeId"`
}

type Symptom struct {
	SymptomId     uint64              `json:"symptom_id" gorm:"primaryKey"`
	SymptomName   string              `json:"symptom_name"`
	SymptomTypeId []uint64            `json:"symptom_type_id,omitempty" gorm:"-"`
	Commonality   string              `json:"commonality"`
	Description   string              `json:"description"`
	IsDeleted     int                 `json:"is_deleted"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	CreatedBy     string              `json:"created_by"`
	SymptomType   []SymptomTypeMaster `gorm:"many2many:tbl_symptom_type_mapping;foreignKey:SymptomId;joinForeignKey:SymptomId;References:SymptomTypeId;joinReferences:SymptomTypeId" json:"symptom_type"`
}

type SymptomTypeMaster struct {
	SymptomTypeId          uint64    `gorm:"column:symptom_type_id;primaryKey;autoIncrement" json:"symptom_type_id"`
	SymptomType            string    `gorm:"column:symptom_type;size:255;not null" json:"symptom_type"`
	SymptomTypeDescription string    `gorm:"column:symptom_type_description" json:"symptom_type_description"`
	IsDeleted              int       `gorm:"column:is_deleted" json:"is_deleted"`
	CreatedAt              time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at,omitempty"`
	CreatedBy              string    `gorm:"column:created_by;size:255" json:"created_by"`
	UpdatedBy              string    `gorm:"column:updated_by;size:255" json:"updated_by"`
}

func (SymptomTypeMaster) TableName() string {
	return "tbl_symptom_type_master"
}

type SymptomTypeMapping struct {
	SymptomId     uint64     `gorm:"primaryKey;autoIncrement:false;not null" json:"symptom_id"`
	SymptomTypeId uint64     `gorm:"primaryKey;autoIncrement:false;not null" json:"symptom_type_id"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoUpdateTime" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at;" json:"updated_at"`
	CreatedBy     string     `gorm:"column:created_by;" json:"created_by"`
	UpdatedBy     string     `gorm:"column:updated_by;" json:"updated_by"`
}

func (SymptomTypeMapping) TableName() string {
	return "tbl_symptom_type_mapping"
}

type DiseaseSymptomMapping struct {
	DiseaseSymptomMappingId uint64    `json:"disease_symptom_mapping_id" gorm:"primaryKey"`
	DiseaseId               uint64    `json:"disease_id"`
	SymptomId               uint64    `json:"symptom_id"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	CreatedBy               string    `json:"created_by"`
	UpdatedBy               string    `json:"updated_by"`
}

func (DiseaseSymptomMapping) TableName() string {
	return "tbl_disease_symptom_mapping"
}

type Severity struct {
	SeverityId    uint64 `json:"severity_id" gorm:"primaryKey"`
	SeverityLevel string `json:"severity_level"`
	IsDeleted     int    `json:"is_deleted"`
}

type Cause struct {
	CauseId     uint64            `json:"cause_id" gorm:"primaryKey"`
	CauseName   string            `json:"cause_name"`
	CauseTypeId []uint64          `json:"cause_type_id,omitempty" gorm:"-"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CreatedBy   string            `json:"created_by"`
	IsDeleted   int               `json:"is_deleted"`
	CauseType   []CauseTypeMaster `gorm:"many2many:tbl_cause_type_mapping;foreignKey:CauseId;joinForeignKey:CauseId;References:CauseTypeId;joinReferences:CauseTypeId" json:"cause_type"`
}

type CauseTypeMaster struct {
	CauseTypeId          uint64    `gorm:"column:cause_type_id;primaryKey;autoIncrement" json:"cause_type_id"`
	CauseType            string    `gorm:"column:cause_type;not null" json:"cause_type"`
	CauseTypeDescription string    `gorm:"column:cause_type_description" json:"cause_type_description"`
	IsDeleted            int       `gorm:"column:is_deleted" json:"is_deleted"`
	CreatedAt            time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at,omitempty"`
	CreatedBy            string    `gorm:"column:created_by;" json:"created_by"`
	UpdatedBy            string    `gorm:"column:updated_by;" json:"updated_by"`
}

func (CauseTypeMaster) TableName() string {
	return "tbl_cause_type_master"
}

type CauseTypeMapping struct {
	CauseId     uint64     `gorm:"primaryKey;autoIncrement:false;not null" json:"cause_id"`
	CauseTypeId uint64     `gorm:"primaryKey;autoIncrement:false;not null" json:"cause_type_id"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoUpdateTime" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;" json:"updated_at"`
	CreatedBy   string     `gorm:"column:created_by;" json:"created_by"`
	UpdatedBy   string     `gorm:"column:updated_by;" json:"updated_by"`
}

func (CauseTypeMapping) TableName() string {
	return "tbl_cause_type_mapping"
}

type DiseaseCauseMapping struct {
	DiseaseCauseMappingId uint64    `gorm:"column:disease_cause_mapping_id;primaryKey;autoIncrement" json:"disease_cause_mapping_id"`
	DiseaseId             uint64    `gorm:"column:disease_id" json:"disease_id"`
	CauseId               uint64    `gorm:"column:cause_id" json:"cause_id"`
	CreatedAt             time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedBy             string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy             string    `gorm:"column:updated_by" json:"updated_by"`
}

func (DiseaseCauseMapping) TableName() string {
	return "tbl_disease_cause_mapping"
}

func (PatientDiseaseProfile) TableName() string { return "tbl_patient_disease_profile" }
func (DiseaseProfile) TableName() string        { return "tbl_disease_profile" }
func (Cause) TableName() string                 { return "tbl_cause_master" }
func (Severity) TableName() string              { return "tbl_severity_master" }
func (Symptom) TableName() string               { return "tbl_symptom_master" }
func (Disease) TableName() string               { return "tbl_disease_master" }
func (DiseaseType) TableName() string           { return "tbl_disease_type_master" }
func (DiseaseTypeMapping) TableName() string {
	return "tbl_disease_type_mapping"
}

type DiseaseProfileSummary struct {
	DiseaseProfileId uint64 `json:"disease_profile_id"`
	DiseaseName      string `json:"disease_name"`
	Description      string `json:"description"`
}

type Medication struct {
	MedicationId    uint64           `json:"medication_id" gorm:"column:medication_id;primaryKey"`
	MedicationName  string           `json:"medication_name" gorm:"column:medication_name"`
	MedicationCode  string           `json:"medication_code" gorm:"column:medication_code"`
	Description     string           `json:"description" gorm:"column:description"`
	CreatedAt       time.Time        `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time        `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	IsDeleted       int              `json:"is_deleted"`
	CreatedBy       string           `json:"created_by" gorm:"column:created_by"`
	MedicationTypes []MedicationType `json:"medication_type" gorm:"foreignKey:MedicationId;references:MedicationId"`
}

// Table name override
func (Medication) TableName() string {
	return "tbl_medication_master"
}

type MedicationType struct {
	DosageId           uint64    `json:"dosage_id" gorm:"column:dosage_id;primaryKey"`
	MedicationId       uint64    `json:"medication_id" gorm:"column:medication_id"`
	MedicationType     string    `json:"medication_type" gorm:"column:medication_type"`
	UnitValue          float64   `json:"unit_value" gorm:"column:unit_value"`
	UnitType           string    `json:"unit_type" gorm:"column:unit_type"`
	MedicationCost     float64   `json:"medication_cost" gorm:"column:medication_cost"`
	MedicationImageURL string    `json:"medication_image_url" gorm:"column:medication_image_url"`
	IsDeleted          int       `json:"is_deleted"`
	CreatedAt          time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
	CreatedBy          string    `json:"created_by" gorm:"column:created_by"`
}

// Table name override
func (MedicationType) TableName() string {
	return "tbl_medication_type"
}

type DiseaseMedicationMapping struct {
	DiseaseMedicationMappingId uint64    `json:"-" gorm:"primaryKey"`
	DiseaseId                  uint64    `json:"disease_id"`
	MedicationId               uint64    `json:"medication_id"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
	CreatedBy                  string    `json:"created_by"`
	UpdatedBy                  string    `json:"updated_by"`
}

func (DiseaseMedicationMapping) TableName() string {
	return "tbl_disease_medication_mapping"
}

type Exercise struct {
	ExerciseId       uint64             `json:"exercise_id" gorm:"primaryKey"`
	ExerciseName     string             `json:"exercise_name"`
	Description      string             `json:"description"`
	Category         string             `json:"category"`
	IntensityLevel   string             `json:"intensity_level"`
	Duration         string             `json:"duration"`
	DurationUnit     string             `json:"duration_unit"`
	Benefits         string             `json:"benefits"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	CreatedBy        string             `json:"created_by"`
	IsDeleted        int                `json:"is_deleted"`
	ExerciseArtifact []ExerciseArtifact `json:"artifact" gorm:"foreignKey:ExerciseId;references:ExerciseId"`
}

type ExerciseArtifact struct {
	ExerciseArtifactId int64     `gorm:"primaryKey;column:exercise_artifact_id" json:"exercise_artifact_id"`
	ExerciseId         int64     `gorm:"column:exercise_id" json:"exercise_id"`
	ArtifactType       string    `gorm:"column:artifact_type;size:50" json:"artifact_type"`
	ArtifactURL        string    `gorm:"column:artifact_url" json:"artifact_url"`
	CreatedAt          time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
	CreatedBy          string    `json:"created_by"`
}

type DiseaseExerciseMapping struct {
	DiseaseExerciseMappingId uint64    `json:"-" gorm:"primaryKey"`
	DiseaseId                uint64    `json:"disease_id"`
	ExerciseId               uint64    `json:"exercise_id"`
	Exercise                 Exercise  `json:"exercise" gorm:"foreignKey:ExerciseId;references:ExerciseId"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
	CreatedBy                string    `json:"created_by"`
	UpdatedBy                string    `json:"updated_by"`
}

func (Exercise) TableName() string {
	return "tbl_exercise_master"
}

func (ExerciseArtifact) TableName() string {
	return "tbl_exercise_artifact"
}

func (DiseaseExerciseMapping) TableName() string {
	return "tbl_disease_exercise_mapping"
}

type DietPlanTemplate struct {
	DietPlanTemplateId uint64    `json:"diet_plan_template_id" gorm:"primaryKey"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Goal               string    `json:"goal"`
	Notes              string    `json:"notes"`
	DietCreatorId      uint      `json:"-"`
	Cost               float64   `json:"-"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	CreatedBy          string    `json:"created_by"`
	IsDeleted          int       `json:"is_deleted"`
	Meals              []Meal    `json:"meals" gorm:"foreignKey:DietPlanTemplateId;constraint:OnDelete:CASCADE;"`
}

type Meal struct {
	MealId             uint64     `json:"meal_id" gorm:"primaryKey"`
	DietPlanTemplateId uint64     `json:"diet_plan_template_id"`
	MealType           string     `json:"meal_type"`
	Description        string     `json:"description"`
	Nutrients          []Nutrient `json:"nutrients" gorm:"foreignKey:MealId;constraint:OnDelete:CASCADE;"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	CreatedBy          string     `json:"created_by"`
	IsDeleted          int        `json:"is_deleted"`
}

type Nutrient struct {
	NutrientId   uint64    `json:"nutrient_id" gorm:"primaryKey"`
	MealId       uint64    `json:"meal_id"`
	NutrientName string    `json:"nutrient_name"`
	Amount       string    `json:"amount"`
	Unit         string    `json:"unit"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedBy    string    `json:"created_by"`
	IsDeleted    int       `json:"is_deleted"`
}

func (DietPlanTemplate) TableName() string {
	return "tbl_diet_plan_template"
}

func (Meal) TableName() string {
	return "tbl_meal"
}

func (Nutrient) TableName() string {
	return "tbl_nutrient"
}

type DiseaseDietMapping struct {
	DiseaseDietMappingId uint64    `json:"disease_diet_mapping_id" gorm:"primaryKey"`
	DiseaseId            uint64    `json:"disease_id"`
	DietPlanTemplateId   uint64    `json:"diet_plan_template_id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	CreatedBy            string    `json:"created_by"`
	UpdatedBy            string    `json:"updated_by"`
}

func (DiseaseDietMapping) TableName() string {
	return "tbl_disease_diet_mapping"
}

type DiagnosticTest struct {
	DiagnosticTestId uint64                    `gorm:"column:diagnostic_test_id;primaryKey" json:"diagnostic_test_id"`
	LoincCode        string                    `gorm:"column:test_loinc_code" json:"test_loinc_code"`
	TestName         string                    `gorm:"column:test_name" json:"test_name"`
	TestType         string                    `gorm:"column:test_type" json:"test_type"`
	Description      string                    `gorm:"column:test_description" json:"test_description"`
	Category         string                    `gorm:"column:category" json:"category"`
	Units            string                    `gorm:"column:units" json:"units"`
	Property         string                    `gorm:"column:property" json:"property"`
	TimeAspect       string                    `gorm:"column:time_aspect" json:"time_aspect"`
	System           string                    `gorm:"column:system" json:"system"`
	Scale            string                    `gorm:"column:scale" json:"scale"`
	CreatedAt        time.Time                 `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time                 `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedBy        string                    `gorm:"column:created_by" json:"created_by"`
	IsDeleted        int                       `gorm:"column:is_deleted" json:"is_deleted"`
	Components       []DiagnosticTestComponent `gorm:"many2many:tbl_disease_profile_diagnostic_test_component_mapping;foreignKey:DiagnosticTestId;joinForeignKey:DiagnosticTestId;References:DiagnosticTestComponentId;joinReferences:DiagnosticTestComponentId" json:"test_components"`
}

type DiagnosticTestComponent struct {
	DiagnosticTestComponentId uint64                             `gorm:"column:diagnostic_test_component_id;primaryKey" json:"diagnostic_test_component_id"`
	LoincCode                 string                             `gorm:"column:test_component_loinc_code" json:"test_component_loinc_code"`
	TestComponentName         string                             `gorm:"column:test_component_name" json:"test_component_name"`
	TestComponentType         string                             `gorm:"column:test_component_type" json:"test_component_type"`
	Description               string                             `gorm:"column:description" json:"description"`
	Units                     string                             `gorm:"column:units" json:"units"`
	Property                  string                             `gorm:"column:property" json:"property"`
	TimeAspect                string                             `gorm:"column:time_aspect" json:"time_aspect"`
	System                    string                             `gorm:"column:system" json:"system"`
	Scale                     string                             `gorm:"column:scale" json:"scale"`
	TestComponentFrequency    string                             `gorm:"column:test_component_frequency" json:"test_component_frequency"`
	CreatedAt                 time.Time                          `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt                 time.Time                          `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedBy                 string                             `gorm:"column:created_by" json:"created_by"`
	IsDeleted                 int                                `gorm:"column:is_deleted" json:"is_deleted"`
	TestResultValue           []PatientDiagnosticTestResultValue `gorm:"foreignKey:DiagnosticTestComponentId;references:DiagnosticTestComponentId" json:"test_result_value"`
	ReferenceRange            []DiagnosticTestReferenceRange     `gorm:"foreignKey:DiagnosticTestComponentId;references:DiagnosticTestComponentId" json:"test_referance_range"`
}

type DiagnosticTestComponentMapping struct {
	DiagnosticTestComponentMappingId uint64    `gorm:"column:diagnostic_test_component_mapping_id;primaryKey" json:"diagnostic_test_component_mapping_id"`
	DiagnosticTestId                 uint64    `gorm:"column:diagnostic_test_id" json:"diagnostic_test_id"`
	DiagnosticComponentId            uint64    `gorm:"column:diagnostic_test_component_id" json:"diagnostic_test_component_id"`
	CreatedAt                        time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt                        time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedBy                        string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy                        string    `gorm:"column:updated_by" json:"updated_by"`
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
	DiseaseDiagnosticTestMapping uint64    `gorm:"column:disease_diagnostic_test_mapping_id;primaryKey"`
	DiseaseId                    uint64    `json:"disease_id"`
	DiagnosticTestId             uint64    `json:"diagnostic_test_id"`
	CreatedAt                    time.Time `json:"created_at"`
	UpdatedAt                    time.Time `json:"updated_at"`
	CreatedBy                    string    `json:"created_by"`
	UpdatedBy                    string    `json:"updated_by"`
}

func (DiseaseDiagnosticTestMapping) TableName() string {
	return "tbl_disease_diagnostic_test_mapping"
}

func (d *Disease) SetCreatedBy(userId string) {
	d.CreatedBy = userId
}

func (s *Symptom) SetCreatedBy(userId string) {
	s.CreatedBy = userId
}

func (c *Cause) SetCreatedBy(userId string) {
	c.CreatedBy = userId
}

func (c *Medication) SetCreatedBy(userId string) {
	c.CreatedBy = userId
}

func (c *Exercise) SetCreatedBy(userId string) {
	c.CreatedBy = userId
}

func (c *DietPlanTemplate) SetCreatedBy(userId string) {
	c.CreatedBy = userId
}

func (c *DiagnosticTest) SetCreatedBy(userId string) {
	c.CreatedBy = userId
}

func (c *DiagnosticTestComponent) SetCreatedBy(userId string) {
	c.CreatedBy = userId
}

func (c *DiagnosticLab) SetCreatedBy(userId string) {
	c.CreatedBy = userId
}

func (c *SupportGroup) SetCreatedBy(userId string) {
	c.CreatedBy = userId
}

func (c *Hospital) SetCreatedBy(userId string) {
	c.CreatedBy = userId
}
func (c *Service) SetCreatedBy(userId string) {
	c.CreatedBy = userId
}

type Creator interface {
	SetCreatedBy(string)
}
