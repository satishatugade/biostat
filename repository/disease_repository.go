package repository

import (
	"biostat/constant"
	"biostat/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type DiseaseRepository interface {
	GetAllDiseasesInfo(limit int, offset int) ([]models.Disease, int64, error)
	GetDiseases(diseaseId uint64) (*models.Disease, error)
	GetAllDiseases(limit int, offset int) ([]models.Disease, int64, error)
	CreateDiseaseProfile(profile models.DiseaseProfile) error
	GetDiseaseProfiles(limit int, offset int) ([]models.DiseaseProfile, int64, error)
	GetDiseaseProfileById(diseaseProfileId string) (*models.DiseaseProfile, error)
	CreateDisease(disease *models.Disease) error
	UpdateDisease(updatedDisease *models.Disease, authUserId string) error
	DeleteDisease(DiseaseId uint64, authUserId string) error
	GetDiseaseAuditLogs(diseaseId uint64, diseaseAuditId uint64) ([]models.DiseaseAudit, error)
	GetAllDiseaseAuditLogs(page, limit int) ([]models.DiseaseAudit, int64, error)
	IsDiseaseProfileExists(diseaseProfileId uint) (bool, error)

	InsertMedication(medication *models.Medication) error
	InsertMedicationType(medicationType *[]models.MedicationType) error

	BulkInsert(data interface{}) error
}

type DiseaseRepositoryImpl struct {
	db *gorm.DB
}

func NewDiseaseRepository(db *gorm.DB) DiseaseRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &DiseaseRepositoryImpl{db: db}
}

// GetDiseases implements DiseaseRepository.
func (repo *DiseaseRepositoryImpl) GetDiseases(diseaseId uint64) (*models.Disease, error) {
	var disease models.Disease
	if err := repo.db.Where("disease_id = ?", diseaseId).First(&disease).Error; err != nil {
		return nil, err
	}
	return &disease, nil
}

func (repo *DiseaseRepositoryImpl) CreateDisease(disease *models.Disease) error {
	if disease == nil {
		return nil
	}
	tx := repo.db.Begin()

	if err := tx.Create(disease).Error; err != nil {
		tx.Rollback()
		return err
	}
	mapping := models.DiseaseTypeMapping{
		DiseaseId:     disease.DiseaseId,
		DiseaseTypeId: disease.DiseaseTypeId, // make sure this is set in input
	}

	if err := tx.Create(&mapping).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (repo *DiseaseRepositoryImpl) GetAllDiseases(limit, offset int) ([]models.Disease, int64, error) {
	var diseases []models.Disease
	var totalRecords int64

	// Get total count of diseases before applying pagination
	if err := repo.db.Model(&models.Disease{}).Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	// Build base query
	query := repo.db.Model(&models.Disease{}).Order("disease_id DESC")

	// Apply pagination only if limit is greater than 0
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	// Fetch diseases
	if err := query.Find(&diseases).Error; err != nil {
		return nil, 0, err
	}

	return diseases, totalRecords, nil
}

func (r *DiseaseRepositoryImpl) GetAllDiseasesInfo(limit int, offset int) ([]models.Disease, int64, error) {
	var diseases []models.Disease
	var totalRecords int64

	if err := r.db.Model(&models.Disease{}).Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.
		Preload("DiseaseTypeMapping.DiseaseType").
		Preload("Symptoms").
		Preload("Causes").
		Limit(limit).Offset(offset).
		Find(&diseases).Error

	if err != nil {
		return nil, 0, err
	}

	return diseases, totalRecords, nil
}

func (r *DiseaseRepositoryImpl) CreateDiseaseProfile(profile models.DiseaseProfile) error {
	return r.db.Create(&profile).Error
}

func (r *DiseaseRepositoryImpl) GetDiseaseProfiles(limit int, offset int) ([]models.DiseaseProfile, int64, error) {
	var diseaseProfiles []models.DiseaseProfile
	var totalRecords int64

	if err := r.db.Model(&models.DiseaseProfile{}).Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("Disease").
		Preload("Disease.Severity").
		Preload("Disease.Symptoms").
		Preload("Disease.Causes").
		Preload("Disease.DiseaseTypeMapping").
		Preload("Disease.DiseaseTypeMapping.DiseaseType").
		Preload("Disease.Medications").
		Preload("Disease.Medications.MedicationTypes").
		Preload("Disease.Exercises").
		Preload("Disease.Exercises.ExerciseArtifact").
		Preload("Disease.DietPlans").
		Preload("Disease.DietPlans.Meals").
		Preload("Disease.DietPlans.Meals.Nutrients").
		Preload("Disease.DiagnosticTests").
		Preload("Disease.DiagnosticTests.Components").
		Order("disease_profile_id DESC").
		Limit(limit).
		Offset(offset).
		Find(&diseaseProfiles).Error

	if err != nil {
		return nil, 0, err
	}
	for i := range diseaseProfiles {
		diseaseProfiles[i].Disease.DiseaseType = diseaseProfiles[i].Disease.DiseaseTypeMapping.DiseaseType
	}

	return diseaseProfiles, totalRecords, nil
}

func (r *DiseaseRepositoryImpl) GetDiseaseProfileById(diseaseProfileId string) (*models.DiseaseProfile, error) {
	var diseaseProfile models.DiseaseProfile

	err := r.db.Preload("Disease").
		Preload("Disease.Symptoms").
		Preload("Disease.Causes").
		Preload("Disease.DiseaseTypeMapping").
		Preload("Disease.DiseaseTypeMapping.DiseaseType").
		Preload("Disease.Medications").
		Preload("Disease.Medications.MedicationTypes").
		Preload("Disease.Exercises").
		Preload("Disease.Exercises.ExerciseArtifact").
		Preload("Disease.DietPlans").
		Preload("Disease.DietPlans.Meals").
		Preload("Disease.DietPlans.Meals.Nutrients").
		Preload("Disease.DiagnosticTests").
		Preload("Disease.DiagnosticTests.Components").
		Where("disease_profile_id = ?", diseaseProfileId).
		First(&diseaseProfile).Error

	if err != nil {
		return nil, err
	}

	diseaseProfile.Disease.DiseaseType = diseaseProfile.Disease.DiseaseTypeMapping.DiseaseType

	return &diseaseProfile, nil
}

func (d *DiseaseRepositoryImpl) IsDiseaseProfileExists(diseaseProfileId uint) (bool, error) {
	var count int64
	err := d.db.Model(&models.DiseaseProfile{}).Where("disease_profile_id = ?", diseaseProfileId).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repo *DiseaseRepositoryImpl) DiseaseAudit(existingDisease *models.Disease, operationType string, updatedBy string) error {
	auditLog := models.DiseaseAudit{
		DiseaseId:         existingDisease.DiseaseId,
		DiseaseSnomedCode: existingDisease.DiseaseSnomedCode,
		DiseaseName:       existingDisease.DiseaseName,
		Description:       existingDisease.Description,
		ImageURL:          existingDisease.ImageURL,
		SlugURL:           existingDisease.SlugURL,
		OperationType:     operationType,
		IsDeleted:         1,
		CreatedAt:         existingDisease.CreatedAt,
		UpdatedAt:         time.Now(),
		CreatedBy:         existingDisease.CreatedBy,
		UpdatedBy:         updatedBy,
	}

	return repo.db.Create(&auditLog).Error
}

func (repo *DiseaseRepositoryImpl) UpdateDisease(Disease *models.Disease, authUserId string) error {
	existingDisease, err := repo.GetDiseases(Disease.DiseaseId)
	if err != nil {
		return err
	}
	Disease.UpdatedAt = time.Now()
	result := repo.db.Model(&models.Disease{}).Where("disease_id = ?", Disease.DiseaseId).Updates(Disease)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no records updated")
	}
	if err := repo.DiseaseAudit(existingDisease, constant.UPDATE, authUserId); err != nil {
		return err
	}
	return nil
}

func (repo *DiseaseRepositoryImpl) DeleteDisease(diseaseId uint64, authUserId string) error {
	existingDisease, err := repo.GetDiseases(diseaseId)
	if err != nil {
		return err
	}
	if err := repo.DiseaseAudit(existingDisease, constant.DELETE, authUserId); err != nil {
		return err
	}
	if err := repo.db.Model(&models.Disease{}).Where("disease_id = ?", diseaseId).Update("is_deleted", 1).Error; err != nil {
		return err
	}
	return nil
}

func (repo *DiseaseRepositoryImpl) GetDiseaseAuditLogs(diseaseId uint64, diseaseAuditId uint64) ([]models.DiseaseAudit, error) {
	var auditLogs []models.DiseaseAudit
	query := repo.db
	if diseaseId != 0 {
		query = query.Where("disease_id = ?", diseaseId)
	}
	if diseaseAuditId != 0 {
		query = query.Where("disease_audit_id = ?", diseaseAuditId)
	}

	err := query.Find(&auditLogs).Error
	if err != nil {
		return nil, err
	}
	return auditLogs, nil
}

func (repo *DiseaseRepositoryImpl) GetAllDiseaseAuditLogs(page, limit int) ([]models.DiseaseAudit, int64, error) {
	var auditLogs []models.DiseaseAudit
	var totalRecords int64

	// Get total count
	repo.db.Model(&models.DiseaseAudit{}).Count(&totalRecords)

	// Fetch data with pagination
	err := repo.db.Limit(limit).Offset((page - 1) * limit).Find(&auditLogs).Error
	return auditLogs, totalRecords, err
}

func (r *DiseaseRepositoryImpl) BulkInsert(data interface{}) error {
	return r.db.Create(data).Error
}

// InsertMedication implements DiseaseRepository.
func (repo *DiseaseRepositoryImpl) InsertMedication(medication *models.Medication) error {
	return repo.db.Create(medication).Error
}

// InsertMedicationType implements DiseaseRepository.
func (repo *DiseaseRepositoryImpl) InsertMedicationType(medicationType *[]models.MedicationType) error {
	return repo.db.Create(medicationType).Error
}
