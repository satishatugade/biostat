package service

import (
	"biostat/models"
	"biostat/repository"
)

type DiseaseService interface {
	GetDiseases(diseaseId uint) (*models.Disease, error)
	GetAllDiseases(limit int, offset int) ([]models.Disease, int64, error)
	GetDiseaseProfiles(limit int, offset int) ([]models.DiseaseProfile, int64, error)
	GetDiseaseProfileById(diseaseProfileId string) (*models.DiseaseProfile, error)
	CreateDisease(disease *models.Disease) error
}

type DiseaseServiceImpl struct {
	diseaseRepo repository.DiseaseRepository
}

// GetAllDiseasesInfo implements DiseaseService.
func (s *DiseaseServiceImpl) GetAllDiseases(limit int, offset int) ([]models.Disease, int64, error) {
	return s.diseaseRepo.GetAllDiseases(limit, offset)
	// return s.diseaseRepo.GetAllDiseasesInfo(limit, offset)
}

func NewDiseaseService(repo repository.DiseaseRepository) DiseaseService {
	return &DiseaseServiceImpl{diseaseRepo: repo}
}

func (s *DiseaseServiceImpl) GetDiseases(diseaseId uint) (*models.Disease, error) {
	return s.diseaseRepo.GetDiseases(diseaseId)
}

func (s *DiseaseServiceImpl) GetDiseaseProfileById(diseaseProfileId string) (*models.DiseaseProfile, error) {
	return s.diseaseRepo.GetDiseaseProfileById(diseaseProfileId)
}

func (s *DiseaseServiceImpl) GetDiseaseProfiles(limit int, offset int) ([]models.DiseaseProfile, int64, error) {
	return s.diseaseRepo.GetDiseaseProfiles(limit, offset)
}

func (s *DiseaseServiceImpl) CreateDisease(disease *models.Disease) error {
	return s.diseaseRepo.CreateDisease(disease)
}
