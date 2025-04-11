package service

import (
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"fmt"
	"io"
	"strconv"
	"time"
)

type DiseaseService interface {
	GetDiseases(diseaseId uint64) (*models.Disease, error)
	GetAllDiseases(limit int, offset int) ([]models.Disease, int64, error)
	GetDiseaseProfiles(limit int, offset int) ([]models.DiseaseProfile, int64, error)
	GetDiseaseProfileById(diseaseProfileId string) (*models.DiseaseProfile, error)
	CreateDisease(disease *models.Disease) error
	UpdateDisease(Disease *models.Disease, authUserId string) error
	DeleteDisease(DiseaseId uint64, authUserId string) error
	GetDiseaseAuditLogs(diseaseId uint64, diseaseAuditId uint64) ([]models.DiseaseAudit, error)
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

func (s *DiseaseServiceImpl) GetDiseases(diseaseId uint64) (*models.Disease, error) {
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

func (s *DiseaseServiceImpl) UpdateDisease(Disease *models.Disease, authUserId string) error {
	return s.diseaseRepo.UpdateDisease(Disease, authUserId)
}

func (s *DiseaseServiceImpl) DeleteDisease(diseaseId uint64, authUserId string) error {
	return s.diseaseRepo.DeleteDisease(diseaseId, authUserId)
}

func (s *DiseaseServiceImpl) GetDiseaseAuditLogs(diseaseId uint64, diseaseAuditId uint64) ([]models.DiseaseAudit, error) {
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
	case "DTMaster":
		return ProcessAndInsert[models.DiagnosticTest](s, reader, authUserId)
	case "DTCMaster":
		return ProcessAndInsert[models.DiagnosticTestComponent](s, reader, authUserId)
	case "DiagnosticLab":
		return ProcessAndInsert[models.DiagnosticLab](s, reader, authUserId)
	case "MedicationMaster":
		return processMedicationInsert(s, reader, authUserId)
	default:
		return 0, fmt.Errorf("unsupported entity: %s", entity)
	}
}

func ProcessAndInsert[T any](s *DiseaseServiceImpl, reader io.Reader, authUserId string) (int, error) {
	data, err := utils.ParseExcelFromReader[T](reader)
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

func processMedicationInsert(s *DiseaseServiceImpl, reader io.Reader, authUserId string) (int, error) {

	type MedicationExcelRow struct {
		MedicationName     string `json:"medication_name"`
		MedicationCode     string `json:"medication_code"`
		Description        string `json:"description"`
		MedicationType     string `json:"medication_type"`
		UnitValue          string `json:"unit_value"`
		UnitType           string `json:"unit_type"`
		MedicationCost     string `json:"medication_cost"`
		MedicationImageURL string `json:"medication_image_url"`
	}

	data, err := utils.ParseExcelFromReader[MedicationExcelRow](reader)
	if err != nil {
		return 0, err
	}

	type medKey struct {
		Name        string
		Code        string
		Description string
	}
	groupMap := make(map[medKey][]MedicationExcelRow)

	for _, row := range data {
		key := medKey{
			Name:        row.MedicationName,
			Code:        row.MedicationCode,
			Description: row.Description,
		}
		groupMap[key] = append(groupMap[key], row)
	}

	totalInserted := 0

	for key, rows := range groupMap {
		med := models.Medication{
			MedicationName: key.Name,
			MedicationCode: key.Code,
			Description:    key.Description,
			CreatedBy:      authUserId,
			IsDeleted:      0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		err := s.diseaseRepo.InsertMedication(&med)
		if err != nil {
			return totalInserted, err
		}

		var medTypes []models.MedicationType
		for _, row := range rows {
			unitVal, _ := strconv.ParseFloat(row.UnitValue, 64)
			costVal, _ := strconv.ParseFloat(row.MedicationCost, 64)
			medTypes = append(medTypes, models.MedicationType{
				MedicationId:       med.MedicationId,
				MedicationType:     row.MedicationType,
				UnitValue:          unitVal,
				UnitType:           row.UnitType,
				MedicationCost:     costVal,
				MedicationImageURL: row.MedicationImageURL,
				CreatedBy:          authUserId,
				IsDeleted:          0,
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			})
		}

		err = s.diseaseRepo.InsertMedicationType(&medTypes)
		if err != nil {
			return totalInserted, err
		}

		totalInserted++
	}

	return totalInserted, nil
}
