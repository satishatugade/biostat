package service

import (
	"biostat/models"
	"biostat/repository"
)

type PatientService interface {
	GetPatients(limit int, offset int) ([]models.Patient, int64, error)
	GetPatientById(patientId string) (*models.Patient, error)
	UpdatePatientById(patientId string, patientData *models.Patient) (*models.Patient, error)
	GetPatientDiseaseProfiles(PatientId string) ([]models.PatientDiseaseProfile, error)
	AddPatientPrescription(patientPrescription *models.PatientPrescription) error
	GetAllPrescription(limit int, offset int) ([]models.PatientPrescription, int64, error)
	GetPrescriptionByPatientID(PatientDiseaseProfileId string, limit int, offset int) ([]models.PatientPrescription, int64, error)
	AddPatientRelative(relative *models.PatientRelative) error
	GetPatientRelative(patientId string) ([]models.PatientRelative, error)
	UpdatePatientRelative(relativeId uint, relative *models.PatientRelative) (models.PatientRelative, error)
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
func (s *PatientServiceImpl) UpdatePatientById(patientId string, patientData *models.Patient) (*models.Patient, error) {
	return s.patientRepo.UpdatePatientById(patientId, patientData)
}

func (s *PatientServiceImpl) AddPatientPrescription(prescription *models.PatientPrescription) error {
	return s.patientRepo.AddPatientPrescription(prescription)
}

func (s *PatientServiceImpl) GetPatientDiseaseProfiles(PatientId string) ([]models.PatientDiseaseProfile, error) {
	return s.patientRepo.GetPatientDiseaseProfiles(PatientId)
}

// AddPatientRelative implements PatientService.
func (s *PatientServiceImpl) AddPatientRelative(relative *models.PatientRelative) error {
	return s.patientRepo.AddPatientRelative(relative)

}

// GetPatientRelatives implements PatientService.
func (s *PatientServiceImpl) GetPatientRelative(patientId string) ([]models.PatientRelative, error) {
	return s.patientRepo.GetPatientRelative(patientId)
}

func (s *PatientServiceImpl) UpdatePatientRelative(relativeId uint, relative *models.PatientRelative) (models.PatientRelative, error) {
	return s.patientRepo.UpdatePatientRelative(relativeId, relative)
}

// func (s *PatientServiceImpl) UpdatePrescription(prescription *models.PatientPrescription) error {
// 	return s.patientRepo.UpdatePrescription(prescription)
// }
