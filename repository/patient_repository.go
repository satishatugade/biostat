package repository

import (
	"biostat/models"
	"errors"

	"gorm.io/gorm"
)

type PatientRepository interface {
	GetAllPatients(limit int, offset int) ([]models.Patient, int64, error)
	AddPatientPrescription(*models.PatientPrescription) error
	GetAllPrescription(limit int, offset int) ([]models.PatientPrescription, int64, error)
	GetPrescriptionByPatientID(patientID string, limit int, offset int) ([]models.PatientPrescription, int64, error)
	GetPatientDiseaseProfiles(PatientId string) ([]models.PatientDiseaseProfile, error)
	GetPatientById(patientId uint) (*models.Patient, error)
	UpdatePatientById(patientId string, patientData *models.Patient) (*models.Patient, error)
	AddPatientRelative(relative *models.PatientRelative) error
	GetPatientRelative(patientId string) ([]models.PatientRelative, error)
	GetPatientRelativeById(relativeId uint) (models.PatientRelative, error)
	UpdatePatientRelative(relativeId uint, relative *models.PatientRelative) (models.PatientRelative, error)
	AddPatientClinicalRange(customeRange *models.PatientCustomRange) error
	// UpdatePrescription(*models.PatientPrescription) error
}

type PatientRepositoryImpl struct {
	db                *gorm.DB
	diseaseRepository DiseaseRepositoryImpl
}

func NewPatientRepository(db *gorm.DB) PatientRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &PatientRepositoryImpl{db: db}
}

func (p *PatientRepositoryImpl) AddPatientPrescription(prescription *models.PatientPrescription) error {
	if err := p.db.Create(prescription).Error; err != nil {
		return err
	}
	return nil
}

func (p *PatientRepositoryImpl) GetAllPrescription(limit int, offset int) ([]models.PatientPrescription, int64, error) {
	var prescriptions []models.PatientPrescription
	var totalRecords int64

	query := p.db.
		Preload("PrescriptionDetails").
		Find(&prescriptions).
		Count(&totalRecords)

	if query.Error != nil {
		return nil, 0, query.Error
	}

	return prescriptions, totalRecords, nil
}

func (p *PatientRepositoryImpl) GetPrescriptionByPatientID(patientID string, limit int, offset int) ([]models.PatientPrescription, int64, error) {
	var prescriptions []models.PatientPrescription
	var totalRecords int64

	query := p.db.
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

func (p *PatientRepositoryImpl) GetAllPatientPrescription(prescription *models.PatientPrescription) ([]models.PatientPrescription, int64, error) {
	var prescriptions []models.PatientPrescription
	var totalRecords int64

	query := p.db.
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
func (p *PatientRepositoryImpl) GetPatientById(patientId uint) (*models.Patient, error) {
	var patient models.Patient
	err := p.db.Where("patient_id = ?", patientId).First(&patient).Error
	if err != nil {
		return nil, err
	}
	return &patient, nil
}

func (p *PatientRepositoryImpl) GetAllPatients(limit int, offset int) ([]models.Patient, int64, error) {

	var patients []models.Patient
	var totalRecords int64

	// Count total records in the table
	err := p.db.Model(&models.Patient{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}

	// Fetch paginated data
	err = p.db.Limit(limit).Offset(offset).Find(&patients).Error
	if err != nil {
		return nil, 0, err
	}

	return patients, totalRecords, nil
}

func (p *PatientRepositoryImpl) UpdatePatientById(patientId string, patientData *models.Patient) (*models.Patient, error) {
	var patient models.Patient

	// Check if patient exists
	err := p.db.Where("patient_id = ?", patientId).First(&patient).Error
	if err != nil {
		return nil, err
	}

	// Update patient info
	err = p.db.Model(&patient).Updates(patientData).Error
	if err != nil {
		return nil, err
	}

	return &patient, nil
}

func (p *PatientRepositoryImpl) GetPatientDiseaseProfiles(PatientId string) ([]models.PatientDiseaseProfile, error) {
	var patientDiseaseProfiles []models.PatientDiseaseProfile

	err := p.db.Preload("DiseaseProfile").
		Preload("DiseaseProfile.Disease").
		Preload("DiseaseProfile.Disease.Severity").
		Preload("DiseaseProfile.Disease.Symptoms").
		Preload("DiseaseProfile.Disease.Causes").
		Preload("DiseaseProfile.Disease.DiseaseTypeMapping").
		Preload("DiseaseProfile.Disease.DiseaseTypeMapping.DiseaseType").
		Preload("DiseaseProfile.Disease.Medications").
		Preload("DiseaseProfile.Disease.Medications.MedicationTypes").
		Preload("DiseaseProfile.Disease.Exercises").
		Preload("DiseaseProfile.Disease.Exercises.ExerciseArtifact").
		Preload("DiseaseProfile.Disease.DietPlans").
		Preload("DiseaseProfile.Disease.DietPlans.Meals").
		Preload("DiseaseProfile.Disease.DietPlans.Meals.Nutrients").
		Preload("DiseaseProfile.Disease.DiagnosticTests").
		Preload("DiseaseProfile.Disease.DiagnosticTests.Components").
		// Where("patient_disease_profile_id = ?", PatientDiseaseProfileId).
		Where("patient_id = ?", PatientId).
		Find(&patientDiseaseProfiles).Error

	if err != nil {
		return nil, err
	}

	return patientDiseaseProfiles, nil
}

// AddPatientRelative implements PatientRepository.
func (p *PatientRepositoryImpl) AddPatientRelative(relative *models.PatientRelative) error {
	return p.db.Create(relative).Error
}

func (p *PatientRepositoryImpl) GetPatientRelative(patientId string) ([]models.PatientRelative, error) {
	var relatives []models.PatientRelative
	err := p.db.Where("patient_id = ?", patientId).Find(&relatives).Error
	return relatives, err
}

func (p *PatientRepositoryImpl) UpdatePatientRelative(relativeId uint, updatedRelative *models.PatientRelative) (models.PatientRelative, error) {
	var relative models.PatientRelative

	// Find the existing relative
	if err := p.db.First(&relative, "relative_id = ?", relativeId).Error; err != nil {
		return models.PatientRelative{}, err
	}

	// Update the fields
	if err := p.db.Model(&relative).Updates(updatedRelative).Error; err != nil {
		return models.PatientRelative{}, err
	}

	// Return the updated relative
	return relative, nil
}

func (s *PatientRepositoryImpl) IsPatientExists(patientID uint) (bool, error) {
	var count int64
	err := s.db.Model(&models.Patient{}).Where("patient_id = ?", patientID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *PatientRepositoryImpl) AddPatientClinicalRange(customRange *models.PatientCustomRange) error {
	tx := p.db.Begin()
	exists, err := p.IsPatientExists(customRange.PatientId)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !exists {
		tx.Rollback()
		return errors.New("patient does not exist")
	}

	exists, err = p.diseaseRepository.IsDiseaseProfileExists(customRange.DiseaseProfileId)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !exists {
		tx.Rollback()
		return errors.New("disease profile does not exist")
	}

	if err := tx.Create(customRange).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// GetPatientRelativeById implements PatientRepository.
func (p *PatientRepositoryImpl) GetPatientRelativeById(relativeId uint) (models.PatientRelative, error) {

	var relative models.PatientRelative
	err := p.db.Where("relative_id = ?", relativeId).First(&relative).Error
	if err != nil {
		return relative, err
	}
	return relative, nil
}
