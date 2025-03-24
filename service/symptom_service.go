package service

import (
	"biostat/models"
	"biostat/repository"
)

type SymptomService interface {
	GetAllSymptom(limit int, offset int) ([]models.Symptom, int64, error)
	AddDiseaseSymptom(symptom *models.Symptom) error
	UpdateSymptom(symptom *models.Symptom) error
}

type SymptomServiceImpl struct {
	symptomRepo repository.SymptomRepository
}

func NewSymptomService(repo repository.SymptomRepository) SymptomService {
	return &SymptomServiceImpl{symptomRepo: repo}
}

// GetAllSymptom implements SymptomService.
func (s *SymptomServiceImpl) GetAllSymptom(limit int, offset int) ([]models.Symptom, int64, error) {
	return s.symptomRepo.GetAllSymptom(limit, offset)
}

// UpdateSymptom implements SymptomService.
func (s *SymptomServiceImpl) UpdateSymptom(symptom *models.Symptom) error {
	return s.symptomRepo.UpdateSymptom(symptom)
}

// AddDiseaseSymptom implements SymptomService.
func (s *SymptomServiceImpl) AddDiseaseSymptom(symptom *models.Symptom) error {
	return s.symptomRepo.AddDiseaseSymptom(symptom)
}
