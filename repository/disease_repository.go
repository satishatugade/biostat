package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type DiseaseRepository interface {
	GetAllDiseasesInfo(limit int, offset int) ([]models.Disease, int64, error)
	GetDiseases(diseaseId uint) (*models.Disease, error)
	GetAllDiseases(limit int, offset int) ([]models.Disease, int64, error)
	GetDiseaseProfiles(limit int, offset int) ([]models.DiseaseProfile, int64, error)
	GetDiseaseProfileById(diseaseProfileId string) (*models.DiseaseProfile, error)
	CreateDisease(disease *models.Disease) error
	IsDiseaseProfileExists(diseaseProfileId uint) (bool, error)
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
func (repo *DiseaseRepositoryImpl) GetDiseases(diseaseId uint) (*models.Disease, error) {
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
	return repo.db.Create(disease).Error
}

func (repo *DiseaseRepositoryImpl) GetAllDiseases(limit, offset int) ([]models.Disease, int64, error) {
	var diseases []models.Disease
	var totalRecords int64

	// Get total count of diseases before applying pagination
	if err := repo.db.Model(&models.Disease{}).Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	// Build base query
	query := repo.db.Model(&models.Disease{}).Order("disease_id ASC")

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
		Order("disease_profile_id ASC").
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
