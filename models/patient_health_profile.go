package models

import "time"

type TblPatientHealthProfile struct {
	PatientHealthProfileID uint64    `gorm:"column:patient_health_profile_id;primaryKey;autoIncrement" json:"patient_health_profile_id"`
	PatientId              uint64    `gorm:"column:patient_id;not null;unique" json:"patient_id"`
	HeightCM               float64   `gorm:"column:height_cm" json:"height_cm"`
	WeightKG               float64   `gorm:"column:weight_kg" json:"weight_kg"`
	BMI                    float64   `gorm:"column:bmi" json:"bmi"`
	BmiCategory            string    `gorm:"column:bmi_category" json:"bmi_category"`
	BloodType              string    `gorm:"column:blood_type;size:3" json:"blood_type"`
	SmokingStatus          string    `gorm:"column:smoking_status;size:50" json:"smoking_status"`
	AlcoholConsumption     string    `gorm:"column:alcohol_consumption;size:50" json:"alcohol_consumption"`
	PhysicalActivityLevel  string    `gorm:"column:physical_activity_level;size:50" json:"physical_activity_level"`
	DietaryPreferences     string    `gorm:"column:dietary_preferences;size:100" json:"dietary_preferences"`
	ExistingConditions     string    `gorm:"column:existing_conditions;type:text" json:"existing_conditions"`
	FamilyMedicalHistory   string    `gorm:"column:family_medical_history;type:text" json:"family_medical_history"`
	MenstrualHistory       string    `gorm:"column:menstrual_history;type:text" json:"menstrual_history"`
	Notes                  string    `gorm:"column:notes;type:text" json:"notes"`
	CreatedAt              time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedBy              string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy              string    `gorm:"column:updated_by" json:"updated_by"`
}

func (TblPatientHealthProfile) TableName() string {
	return "tbl_patient_health_profile"
}

type AllergyReq struct {
	AllergyID   uint64 `json:"allergy_id"`
	AllergyName string `json:"allergy_name"`
	SeverityId  uint64 `json:"severity_id"`
}

type PatientHealthProfileRequest struct {
	HeightCm              float64      `json:"height"`
	WeightKg              float64      `json:"weight"`
	BloodType             string       `json:"blood_group"`
	SmokingStatus         string       `json:"smoking_status"`
	AlcoholConsumption    string       `json:"alcohol_consumption"`
	PhysicalActivityLevel string       `json:"physical_activity_level"`
	DietaryPreferences    string       `json:"diet_preference"`
	ExistingConditions    string       `json:"conditions"`
	FamilyMedicalHistory  string       `json:"family_medical_history"`
	MenstrualHistory      string       `json:"menstrual_history"`
	Notes                 string       `json:"notes"`
	Allergies             []AllergyReq `json:"allergies"`
	DiseaseProfileID      uint64       `json:"disease_profiles"`
}

type PharmacokineticsInput struct {
	Prescription PrescriptionData `json:"prescription"`
	History      HistoryData      `json:"history"`
}

type PrescriptionData struct {
	PatientName  string           `json:"patient_name"`
	Age          int              `json:"age"`
	Gender       string           `json:"gender"`
	BloodGroup   string           `json:"blood_group"`
	BMI          string           `json:"bmi"`
	HeightCM     string           `json:"height"`
	WeightKG     string           `json:"weight"`
	PrescribedOn string           `json:"prescribed_on"`
	Prescription []PrescribedDrug `json:"prescription"`
}

type PrescribedDrug struct {
	Drug      string `json:"drug"`
	Dosage    string `json:"dosage"`
	Frequency string `json:"frequency"`
	Duration  string `json:"duration"`
}

type HistoryData struct {
	PatientName        string                   `json:"patient_name"`
	Conditions         []string                 `json:"conditions"`
	Allergies          []string                 `json:"allergies"`
	CurrentMedications []CurrentMedication      `json:"current_medications"`
	RecentLabResults   []map[string]interface{} `json:"recent_lab_results"`
	Lifestyle          []map[string]interface{} `json:"lifestyle"`
}

type CurrentMedication struct {
	Drug      string `json:"drug"`
	Dosage    string `json:"dosage"`
	Frequency string `json:"frequency"`
}
