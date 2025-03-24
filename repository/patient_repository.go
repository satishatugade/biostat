package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type PatientRepository interface {
	GetAllPatients(limit int, offset int) ([]models.Patient, int64, error)
	AddPatientPrescription(*models.PatientPrescription) error
	GetAllPrescription(limit int, offset int) ([]models.PatientPrescription, int64, error)
	GetPrescriptionByPatientID(patientID string, limit int, offset int) ([]models.PatientPrescription, int64, error)
	GetPatientById(patientId string) (*models.Patient, error)
	// UpdatePrescription(*models.PatientPrescription) error
}

type PatientRepositoryImpl struct {
	db *gorm.DB
}

func NewPatientRepository(db *gorm.DB) PatientRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &PatientRepositoryImpl{db: db}
}

func (r *PatientRepositoryImpl) AddPatientPrescription(prescription *models.PatientPrescription) error {
	if err := r.db.Create(prescription).Error; err != nil {
		return err
	}
	return nil
}

func (r *PatientRepositoryImpl) GetAllPrescription(limit int, offset int) ([]models.PatientPrescription, int64, error) {
	var prescriptions []models.PatientPrescription
	var totalRecords int64

	query := r.db.
		Preload("PrescriptionDetails").
		Find(&prescriptions).
		Count(&totalRecords)

	if query.Error != nil {
		return nil, 0, query.Error
	}

	return prescriptions, totalRecords, nil
}

func (r *PatientRepositoryImpl) GetPrescriptionByPatientID(patientID string, limit int, offset int) ([]models.PatientPrescription, int64, error) {
	var prescriptions []models.PatientPrescription
	var totalRecords int64

	query := r.db.
		Where("patient_id = ?", patientID).
		Preload("PrescriptionDetails").
		Limit(limit).
		Offset(offset).
		Find(&prescriptions).
		Count(&totalRecords)

	if query.Error != nil {
		return nil, 0, query.Error
	}

	return prescriptions, totalRecords, nil
}

func (r *PatientRepositoryImpl) GetAllPatientPrescription(prescription *models.PatientPrescription) ([]models.PatientPrescription, int64, error) {
	var prescriptions []models.PatientPrescription
	var totalRecords int64

	query := r.db.
		Preload("PrescriptionDetails").
		Where("patient_id = ?", prescription.PatientId).
		Find(&prescriptions).
		Count(&totalRecords)

	if query.Error != nil {
		return nil, 0, query.Error
	}

	return prescriptions, totalRecords, nil
}

// GetPatientById implements PatientRepository.
func (r *PatientRepositoryImpl) GetPatientById(patientId string) (*models.Patient, error) {
	var patient models.Patient
	err := r.db.Where("patient_id = ?", patientId).First(&patient).Error
	if err != nil {
		return nil, err
	}
	return &patient, nil
}

func (r *PatientRepositoryImpl) GetAllPatients(limit int, offset int) ([]models.Patient, int64, error) {

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
