package service

import (
	"biostat/models"
	"biostat/repository"
)

type SymptomService interface {
	GetAllSymptom(limit int, offset int) ([]models.Symptom, int64, error)
	AddDiseaseSymptom(symptom *models.Symptom) (*models.Symptom, error)
	UpdateSymptom(symptom *models.Symptom, authUserId string) (*models.Symptom, error)
	DeleteSymptom(symptomId uint64, authUserId string) error
	GetSymptomAuditRecord(symptomId uint64, symptomAuditId uint64) ([]models.SymptomAudit, error)
	GetAllSymptomAuditRecord(limit, offset int) ([]models.SymptomAudit, int64, error)

	AddDiseaseSymptomMapping(mapping *models.DiseaseSymptomMapping) error

	GetAllSymptomTypes(limit int, offset int) ([]models.SymptomTypeMaster, int64, error)
	AddSymptomType(symptomType *models.SymptomTypeMaster) (*models.SymptomTypeMaster, error)
	UpdateSymptomType(symptomType *models.SymptomTypeMaster, userId string) (*models.SymptomTypeMaster, error)
	DeleteSymptomType(symptomTypeId uint64, userId string) error

	GetAllSymptomTypeAuditRecord(limit, offset int) ([]models.SymptomTypeAudit, int64, error)
	GetSymptomTypeAuditRecord(symptomTypeId, symptomTypeAuditId uint64) ([]models.SymptomTypeAudit, error)
}

type SymptomServiceImpl struct {
	symptomRepo repository.SymptomRepository
}

func NewSymptomService(repo repository.SymptomRepository) SymptomService {
	return &SymptomServiceImpl{symptomRepo: repo}
}

func (s *SymptomServiceImpl) GetAllSymptom(limit int, offset int) ([]models.Symptom, int64, error) {
	return s.symptomRepo.GetAllSymptom(limit, offset)
}

func (s *SymptomServiceImpl) UpdateSymptom(symptom *models.Symptom, authUserId string) (*models.Symptom, error) {
	return s.symptomRepo.UpdateSymptom(symptom, authUserId)
}

func (s *SymptomServiceImpl) DeleteSymptom(symptomId uint64, authUserId string) error {
	return s.symptomRepo.DeleteSymptom(symptomId, authUserId)
}

func (s *SymptomServiceImpl) AddDiseaseSymptom(symptom *models.Symptom) (*models.Symptom, error) {
	return s.symptomRepo.AddDiseaseSymptom(symptom)
}

func (s *SymptomServiceImpl) GetAllSymptomAuditRecord(limit, offset int) ([]models.SymptomAudit, int64, error) {
	return s.symptomRepo.GetAllSymptomAuditRecord(limit, offset)
}

func (s *SymptomServiceImpl) GetSymptomAuditRecord(symptomId, symptomAuditId uint64) ([]models.SymptomAudit, error) {
	return s.symptomRepo.GetSymptomAuditRecord(symptomId, symptomAuditId)
}

func (s *SymptomServiceImpl) AddDiseaseSymptomMapping(mapping *models.DiseaseSymptomMapping) error {
	return s.symptomRepo.AddDiseaseSymptomMapping(mapping)
}

func (ss *SymptomServiceImpl) GetAllSymptomTypes(limit int, offset int) ([]models.SymptomTypeMaster, int64, error) {
	return ss.symptomRepo.GetAllSymptomTypes(limit, offset)
}

func (ss *SymptomServiceImpl) AddSymptomType(symptomType *models.SymptomTypeMaster) (*models.SymptomTypeMaster, error) {
	return ss.symptomRepo.AddSymptomType(symptomType)
}

func (ss *SymptomServiceImpl) UpdateSymptomType(symptomType *models.SymptomTypeMaster, userId string) (*models.SymptomTypeMaster, error) {
	return ss.symptomRepo.UpdateSymptomType(symptomType, userId)
}

func (ss *SymptomServiceImpl) DeleteSymptomType(symptomTypeId uint64, userId string) error {
	return ss.symptomRepo.DeleteSymptomType(symptomTypeId, userId)
}

func (service *SymptomServiceImpl) GetAllSymptomTypeAuditRecord(limit, offset int) ([]models.SymptomTypeAudit, int64, error) {
	return service.symptomRepo.GetAllSymptomTypeAuditRecord(limit, offset)
}

func (service *SymptomServiceImpl) GetSymptomTypeAuditRecord(symptomTypeId, symptomTypeAuditId uint64) ([]models.SymptomTypeAudit, error) {
	return service.symptomRepo.GetSymptomTypeAuditRecord(symptomTypeId, symptomTypeAuditId)
}
