package service

import (
	"biostat/models"
	"biostat/repository"
	"fmt"
)

type PatientService interface {
	GetPatients(limit int, offset int) ([]models.Patient, int64, error)
	GetPatientById(patientId uint) (*models.Patient, error)
	UpdatePatientById(patientId string, patientData *models.Patient) (*models.Patient, error)
	GetPatientDiseaseProfiles(PatientId string) ([]models.PatientDiseaseProfile, error)
	AddPatientPrescription(patientPrescription *models.PatientPrescription) error
	GetAllPrescription(limit int, offset int) ([]models.PatientPrescription, int64, error)
	GetPrescriptionByPatientID(PatientDiseaseProfileId string, limit int, offset int) ([]models.PatientPrescription, int64, error)
	AddPatientRelative(relative *models.PatientRelative) error
	GetPatientRelative(patientId string) ([]models.PatientRelative, error)
	GetRelativeList(patientId *uint64) ([]models.PatientRelative, error)
	GetCaregiverList(patientId *uint64) ([]models.Caregiver, error)
	GetDoctorList(patientId *uint64) ([]models.Doctor, error)
	GetPatientList() ([]models.Patient, error)
	GetPatientRelativeById(relativeId uint) (models.PatientRelative, error)
	UpdatePatientRelative(relativeId uint, relative *models.PatientRelative) (models.PatientRelative, error)
	AddPatientClinicalRange(customeRange *models.PatientCustomRange) error
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

func (s *PatientServiceImpl) GetPatientById(patientId uint) (*models.Patient, error) {
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

// AddPatientClinicalRange implements PatientService.
func (s *PatientServiceImpl) AddPatientClinicalRange(customRange *models.PatientCustomRange) error {
	return s.patientRepo.AddPatientClinicalRange(customRange)
}

// GetPatientRelativeById implements PatientService.
func (s *PatientServiceImpl) GetPatientRelativeById(relativeId uint) (models.PatientRelative, error) {
	return s.patientRepo.GetPatientRelativeById(relativeId)
}

// GetRelativeList implements PatientService.
func (s *PatientServiceImpl) GetRelativeList(patientId *uint64) ([]models.PatientRelative, error) {
	relativeUserIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, "R", false)
	if err != nil {
		return []models.PatientRelative{}, err
	}

	return s.patientRepo.GetRelativeList(relativeUserIds)
}

// GetCaregiverList implements PatientService.
func (s *PatientServiceImpl) GetCaregiverList(patientId *uint64) ([]models.Caregiver, error) {

	caregiverUserIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, "C", false)
	if err != nil {
		return []models.Caregiver{}, err
	}

	return s.patientRepo.GetCaregiverList(caregiverUserIds)
}

// GetDoctorList implements PatientService.
func (s *PatientServiceImpl) GetDoctorList(patientId *uint64) ([]models.Doctor, error) {

	doctorUserIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, "D", false)
	if err != nil {
		return []models.Doctor{}, err
	}

	return s.patientRepo.GetDoctorList(doctorUserIds)
}

// GetPatientList implements PatientService.
func (s *PatientServiceImpl) GetPatientList() ([]models.Patient, error) {

	patientUserIds, err := s.patientRepo.FetchUserIdByPatientId(nil, "S", true)
	if err != nil {
		return []models.Patient{}, err
	}
	fmt.Println("patientUserIds ", patientUserIds)
	return s.patientRepo.GetPatientList(patientUserIds)

}
