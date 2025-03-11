package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type DiseaseRepository interface {
	GetAllDiseases(limit int, offset int) ([]models.Disease, int64, error)
	GetDiseaseProfiles(limit int, offset int) ([]models.DiseaseProfile, int64, error)
}

type diseaseRepositoryImpl struct {
	db *gorm.DB
}

func NewDiseaseRepository(db *gorm.DB) DiseaseRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &diseaseRepositoryImpl{db: db}
}

func (r *diseaseRepositoryImpl) GetAllDiseases(limit int, offset int) ([]models.Disease, int64, error) {
	var diseases []models.Disease
	var totalRecords int64

	if err := r.db.Model(&models.Disease{}).Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.
		Preload("DiseaseTypeMapping.DiseaseType").
		Preload("Symptoms").
		Preload("Causes").
		// Preload("SeverityLevels").
		Limit(limit).Offset(offset).
		Find(&diseases).Error

	if err != nil {
		return nil, 0, err
	}

	return diseases, totalRecords, nil
}

func (r *diseaseRepositoryImpl) GetDiseaseProfiles(limit int, offset int) ([]models.DiseaseProfile, int64, error) {
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
		Find(&diseaseProfiles).Error

	if err != nil {
		return nil, 0, err
	}

	return diseaseProfiles, totalRecords, nil
}
