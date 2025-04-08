package service

import (
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"fmt"
	"io"
)

type DiseaseService interface {
	GetDiseases(diseaseId uint) (*models.Disease, error)
	GetAllDiseases(limit int, offset int) ([]models.Disease, int64, error)
	GetDiseaseProfiles(limit int, offset int) ([]models.DiseaseProfile, int64, error)
	GetDiseaseProfileById(diseaseProfileId string) (*models.DiseaseProfile, error)
	CreateDisease(disease *models.Disease) error
	UpdateDisease(Disease *models.Disease) error
	DeleteDisease(DiseaseId uint) error
	GetDiseaseAuditLogs(diseaseId uint, diseaseAuditId uint) ([]models.DiseaseAudit, error)
	GetAllDiseaseAuditLogs(page, limit int) ([]models.DiseaseAudit, int64, error)

	ProcessUploadFromStream(entity, authUserId string, reader io.Reader) (int, error)
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

func (s *DiseaseServiceImpl) UpdateDisease(Disease *models.Disease) error {
	return s.diseaseRepo.UpdateDisease(Disease)
}

func (s *DiseaseServiceImpl) DeleteDisease(diseaseId uint) error {
	return s.diseaseRepo.DeleteDisease(diseaseId)
}

func (s *DiseaseServiceImpl) GetDiseaseAuditLogs(diseaseId uint, diseaseAuditId uint) ([]models.DiseaseAudit, error) {
	return s.diseaseRepo.GetDiseaseAuditLogs(diseaseId, diseaseAuditId)
}

func (s *DiseaseServiceImpl) GetAllDiseaseAuditLogs(page, limit int) ([]models.DiseaseAudit, int64, error) {
	return s.diseaseRepo.GetAllDiseaseAuditLogs(page, limit)
}

func (s *DiseaseServiceImpl) ProcessUploadFromStream(entity, authUserId string, reader io.Reader) (int, error) {
	switch entity {
	case "DiseaseMaster":
		return ProcessAndInsert[models.Disease](s, reader, authUserId)
	case "SymptomMaster":
		return ProcessAndInsert[models.Symptom](s, reader, authUserId)
	case "CauseMaster":
		return ProcessAndInsert[models.Cause](s, reader, authUserId)
	case "ExerciseMaster":
		return ProcessAndInsert[models.Exercise](s, reader, authUserId)
	case "DietMaster":
		return ProcessAndInsert[models.DietPlanTemplate](s, reader, authUserId)
	default:
		return 0, fmt.Errorf("unsupported entity: %s", entity)
	}
}

func ProcessAndInsert[T any](s *DiseaseServiceImpl, reader io.Reader, authUserId string) (int, error) {
	data, err := utils.ParseExcelFromReader[T](reader) // returns []T
	if err != nil {
		return 0, err
	}

	var ptrList []*T
	for i := range data {
		ptrList = append(ptrList, &data[i])
	}

	SetCreatedByForAll(ptrList, authUserId)

	err = s.diseaseRepo.BulkInsert(&ptrList)
	return len(ptrList), err
}

func SetCreatedByForAll[T any](list []*T, userId string) {
	for _, item := range list {
		if setter, ok := any(item).(models.Creator); ok {
			setter.SetCreatedBy(userId)
		}
	}
}
