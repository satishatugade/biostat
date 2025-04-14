package service

import (
	"biostat/models"
	"biostat/repository"
)

type SymptomService interface {
	GetAllSymptom(limit int, offset int) ([]models.Symptom, int64, error)
	AddDiseaseSymptom(symptom *models.Symptom) error
	UpdateSymptom(symptom *models.Symptom, authUserId string) error
	DeleteSymptom(symptomId uint64, authUserId string) error
	GetSymptomAuditRecord(symptomId uint64, symptomAuditId uint64) ([]models.SymptomAudit, error)
	GetAllSymptomAuditRecord(page, limit int) ([]models.SymptomAudit, int64, error)

	AddDiseaseSymptomMapping(mapping *models.DiseaseSymptomMapping) error
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
func (s *SymptomServiceImpl) UpdateSymptom(symptom *models.Symptom, authUserId string) error {
	return s.symptomRepo.UpdateSymptom(symptom, authUserId)
}

func (s *SymptomServiceImpl) DeleteSymptom(symptomId uint64, authUserId string) error {
	return s.symptomRepo.DeleteSymptom(symptomId, authUserId)
}

// AddDiseaseSymptom implements SymptomService.
func (s *SymptomServiceImpl) AddDiseaseSymptom(symptom *models.Symptom) error {
	return s.symptomRepo.AddDiseaseSymptom(symptom)
}

func (s *SymptomServiceImpl) GetAllSymptomAuditRecord(page, limit int) ([]models.SymptomAudit, int64, error) {
	return s.symptomRepo.GetAllSymptomAuditRecord(page, limit)
}

func (s *SymptomServiceImpl) GetSymptomAuditRecord(symptomId, symptomAuditId uint64) ([]models.SymptomAudit, error) {
	return s.symptomRepo.GetSymptomAuditRecord(symptomId, symptomAuditId)
}

func (s *SymptomServiceImpl) AddDiseaseSymptomMapping(mapping *models.DiseaseSymptomMapping) error {
	return s.symptomRepo.AddDiseaseSymptomMapping(mapping)
}
