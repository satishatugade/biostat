package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type PatientRepository interface {
	GetAllPatients(limit int, offset int) ([]models.Patient, int64, error)
	AddPatientPrescription(*models.PatientPrescription) error
	// UpdatePrescription(*models.PatientPrescription) error
}

type patientRepositoryImpl struct {
	db *gorm.DB
}

func NewPatientRepository(db *gorm.DB) PatientRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &patientRepositoryImpl{db: db}
}

func (r *patientRepositoryImpl) AddPatientPrescription(prescription *models.PatientPrescription) error {
	if err := r.db.Create(prescription).Error; err != nil {
		return err
	}
	return nil
}

func (r *patientRepositoryImpl) GetAllPatients(limit int, offset int) ([]models.Patient, int64, error) {

	var patients []models.Patient
	var totalRecords int64

	// Count total records in the table
	err := r.db.Model(&models.Patient{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}

	// Fetch paginated data
	err = r.db.Limit(limit).Offset(offset).Find(&patients).Error
	if err != nil {
		return nil, 0, err
	}

	return patients, totalRecords, nil
}
