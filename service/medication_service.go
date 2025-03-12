package service

import (
	"biostat/models"
	"biostat/repository"
)

type MedicationService interface {
	GetMedications(limit int, offset int) ([]models.Medication, int64, error)
	CreateMedication(medication *models.Medication) error
	UpdateMedication(medication *models.Medication) error
}

type MedicationServiceImpl struct {
	medicineRepo repository.MedicationRepository
}

func NewMedicationService(repo repository.MedicationRepository) MedicationService {
	return &MedicationServiceImpl{medicineRepo: repo}
}

// GetMedications implements MedicationService.
func (m *MedicationServiceImpl) GetMedications(limit int, offset int) ([]models.Medication, int64, error) {
	return m.medicineRepo.GetMedications(limit, offset)
}

// CreateMedication implements MedicationService.
func (m *MedicationServiceImpl) CreateMedication(medication *models.Medication) error {
	return m.medicineRepo.CreateMedication(medication)
}

// UpdateMedication implements MedicationService.
func (m *MedicationServiceImpl) UpdateMedication(medication *models.Medication) error {
	return m.medicineRepo.UpdateMedication(medication)
}
