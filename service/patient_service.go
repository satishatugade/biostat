package service

import (
	"biostat/models"
	"biostat/repository"
)

type PatientService interface {
	GetPatients(limit int, offset int) ([]models.Patient, int64, error)
	GetPatientById(patientId string) (*models.Patient, error)
	AddPatientPrescription(*models.PatientPrescription) error
	GetAllPrescription(limit int, offset int) ([]models.PatientPrescription, int64, error)
	GetPrescriptionByPatientID(patientID string, limit int, offset int) ([]models.PatientPrescription, int64, error)
	// UpdatePrescription(*models.PatientPrescription) error
}

type PatientServiceImpl struct {
	patientRepo repository.PatientRepository
}

// GetAllPatientPrescription implements PatientService.
func (s *PatientServiceImpl) GetAllPrescription(limit int, offset int) ([]models.PatientPrescription, int64, error) {
	return s.patientRepo.GetAllPrescription(limit, offset)
}
func (s *PatientServiceImpl) GetPrescriptionByPatientID(patientID string, limit int, offset int) ([]models.PatientPrescription, int64, error) {
	return s.patientRepo.GetPrescriptionByPatientID(patientID, limit, offset)
}

// Ensure patientRepo is properly initialized
func NewPatientService(repo repository.PatientRepository) PatientService {
	return &PatientServiceImpl{patientRepo: repo}
}

func (s *PatientServiceImpl) GetPatients(limit int, offset int) ([]models.Patient, int64, error) {
	return s.patientRepo.GetAllPatients(limit, offset)
}

func (s *PatientServiceImpl) GetPatientById(patientId string) (*models.Patient, error) {
	return s.patientRepo.GetPatientById(patientId)
}

func (s *PatientServiceImpl) AddPatientPrescription(prescription *models.PatientPrescription) error {
	return s.patientRepo.AddPatientPrescription(prescription)
}

// func (s *PatientServiceImpl) UpdatePrescription(prescription *models.PatientPrescription) error {
// 	return s.patientRepo.UpdatePrescription(prescription)
// }
