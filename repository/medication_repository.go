package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type MedicationRepository interface {
	GetMedications(limit int, offset int) ([]models.Medication, int64, error)
	CreateMedication(medication *models.Medication) error
	UpdateMedication(medication *models.Medication) error
}

type MedicationRepositoryImpl struct {
	db *gorm.DB
}

func NewMedicationRepository(db *gorm.DB) MedicationRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &MedicationRepositoryImpl{db: db}
}

// GetMedications implements MedicationRepository.

func (m *MedicationRepositoryImpl) GetMedications(limit int, offset int) ([]models.Medication, int64, error) {
	var medications []models.Medication
	var totalRecords int64

	// Count total records for pagination
	if err := m.db.Model(&models.Medication{}).Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	// Fetch medications with their associated types
	if err := m.db.Preload("MedicationTypes").Limit(limit).Offset(offset).Find(&medications).Error; err != nil {
		return nil, 0, err
	}

	return medications, totalRecords, nil
}

// CreateMedication implements MedicationRepository.
func (m *MedicationRepositoryImpl) CreateMedication(medication *models.Medication) error {
	return m.db.Create(medication).Error
}

func (r *MedicationRepositoryImpl) UpdateMedication(medication *models.Medication) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Medication{}).
			Where("medication_id = ?", medication.MedicationId).
			Updates(medication).Error; err != nil {
			return err
		}

		for _, medType := range medication.MedicationTypes {
			if medType.DosageId > 0 {
				if err := tx.Model(&models.MedicationType{}).
					Where("dosage_id = ? AND medication_id = ?", medType.DosageId, medication.MedicationId).
					Updates(medType).Error; err != nil {
					return err
				}
			} else {
				medType.MedicationId = medication.MedicationId
				if err := tx.Create(&medType).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}
