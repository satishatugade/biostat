package service

import (
	"biostat/models"
	"biostat/repository"
)

type PatientService interface {
	GetPatients(limit int, offset int) ([]models.Patient, int64, error)
	AddPatientPrescription(*models.PatientPrescription) error
	// UpdatePrescription(*models.PatientPrescription) error
}

type patientServiceImpl struct {
	patientRepo repository.PatientRepository
}

// Ensure patientRepo is properly initialized
func NewPatientService(repo repository.PatientRepository) PatientService {
	return &patientServiceImpl{patientRepo: repo}
}

func (s *patientServiceImpl) GetPatients(limit int, offset int) ([]models.Patient, int64, error) {
	return s.patientRepo.GetAllPatients(limit, offset)
}

func (s *patientServiceImpl) AddPatientPrescription(prescription *models.PatientPrescription) error {
	return s.patientRepo.AddPatientPrescription(prescription)
}

// func (s *patientServiceImpl) UpdatePrescription(prescription *models.PatientPrescription) error {
// 	return s.patientRepo.UpdatePrescription(prescription)
// }
