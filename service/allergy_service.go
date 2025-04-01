package service

import (
	"biostat/models"
	"biostat/repository"
)

type AllergyService interface {
	GetAllergies() ([]models.Allergy, error)
	AddPatientAllergyRestriction(allergy *models.PatientAllergyRestriction) error
	GetPatientAllergyRestriction(patientId string) ([]models.PatientAllergyRestriction, error)
	UpdatePatientAllergyRestriction(allergyUpdate *models.PatientAllergyRestriction) error
}

type AllergyServiceImpl struct {
	allergyRepo repository.AllergyRepository
}

func NewAllergyService(repo repository.AllergyRepository) *AllergyServiceImpl {
	return &AllergyServiceImpl{repo}
}

func (a *AllergyServiceImpl) GetAllergies() ([]models.Allergy, error) {
	return a.allergyRepo.GetAllergies()
}

func (a *AllergyServiceImpl) AddPatientAllergyRestriction(allergy *models.PatientAllergyRestriction) error {
	return a.allergyRepo.AddPatientAllergyRestriction(allergy)
}

func (a *AllergyServiceImpl) GetPatientAllergyRestriction(patientId string) ([]models.PatientAllergyRestriction, error) {
	return a.allergyRepo.GetPatientAllergyRestriction(patientId)
}

func (a *AllergyServiceImpl) UpdatePatientAllergyRestriction(allergyUpdate *models.PatientAllergyRestriction) error {
	return a.allergyRepo.UpdatePatientAllergyRestriction(allergyUpdate)
}
