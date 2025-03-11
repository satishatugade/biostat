package service

import (
	"biostat/models"
	"biostat/repository"
)

type DiseaseService interface {
	GetDiseases(limit int, offset int) ([]models.Disease, int64, error)
	GetDiseaseProfiles(limit int, offset int) ([]models.DiseaseProfile, int64, error)
}

type diseaseServiceImpl struct {
	diseaseRepo repository.DiseaseRepository
}

// Ensure patientRepo is properly initialized
func NewDiseaseService(repo repository.DiseaseRepository) DiseaseService {
	return &diseaseServiceImpl{diseaseRepo: repo}
}

func (s *diseaseServiceImpl) GetDiseases(limit int, offset int) ([]models.Disease, int64, error) {
	return s.diseaseRepo.GetAllDiseases(limit, offset)
}

func (s *diseaseServiceImpl) GetDiseaseProfiles(limit int, offset int) ([]models.DiseaseProfile, int64, error) {
	return s.diseaseRepo.GetDiseaseProfiles(limit, offset)
}
