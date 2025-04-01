package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type AllergyRepository interface {
	AddPatientAllergyRestriction(allergy *models.PatientAllergyRestriction) error
	GetPatientAllergyRestriction(patientId string) ([]models.PatientAllergyRestriction, error)
	UpdatePatientAllergyRestriction(allergyUpdate *models.PatientAllergyRestriction) error
}

type AllergyRepositoryImpl struct {
	db *gorm.DB
}

func NewAllergyRepository(db *gorm.DB) *AllergyRepositoryImpl {
	return &AllergyRepositoryImpl{db}
}

func (a *AllergyRepositoryImpl) AddPatientAllergyRestriction(allergy *models.PatientAllergyRestriction) error {
	return a.db.Create(allergy).Error
}

func (a *AllergyRepositoryImpl) GetPatientAllergyRestriction(patientId string) ([]models.PatientAllergyRestriction, error) {
	var allergies []models.PatientAllergyRestriction
	err := a.db.Preload("Allergy.AllergyType").Preload("Severity").
		Where("patient_id = ?", patientId).
		Find(&allergies).Error
	return allergies, err
}

func (a *AllergyRepositoryImpl) UpdatePatientAllergyRestriction(allergyUpdate *models.PatientAllergyRestriction) error {
	var existingAllergy models.PatientAllergyRestriction

	if err := a.db.First(&existingAllergy, "patient_allergy_restriction_id = ? AND patient_id = ?",
		allergyUpdate.PatientAllergyRestrictionId, allergyUpdate.PatientId).Error; err != nil {
		return err
	}

	return a.db.Model(&existingAllergy).Updates(allergyUpdate).Error
}
