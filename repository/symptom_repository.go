package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type SymptomRepository interface {
	GetAllSymptom(limit int, offset int) ([]models.Symptom, int64, error)
	AddDiseaseSymptom(symptom *models.Symptom) error
	UpdateSymptom(symptom *models.Symptom) error
}

type SymptomRepositoryImpl struct {
	db *gorm.DB
}

func NewSymptomRepository(db *gorm.DB) SymptomRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &SymptomRepositoryImpl{db: db}
}

// GetAllSymptom implements SymptomRepository.
func (s *SymptomRepositoryImpl) GetAllSymptom(limit int, offset int) ([]models.Symptom, int64, error) {

	var symptom []models.Symptom
	var totalRecords int64

	query := s.db.Model(&models.Symptom{})
	query.Count(&totalRecords)

	err := query.Limit(limit).Offset(offset).Find(&symptom).Error
	if err != nil {
		return nil, 0, err
	}
	return symptom, totalRecords, nil
}

// AddDiseaseSymptom implements SymptomRepository.
func (s *SymptomRepositoryImpl) AddDiseaseSymptom(symptom *models.Symptom) error {
	return s.db.Create(symptom).Error
}

// UpdateSymptom implements SymptomRepository.
func (s *SymptomRepositoryImpl) UpdateSymptom(symptom *models.Symptom) error {
	return s.db.Model(&models.Symptom{}).Where("symptom_id = ?", symptom.SymptomId).Updates(symptom).Error
}
