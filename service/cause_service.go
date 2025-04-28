package service

import (
	"biostat/models"
	"biostat/repository"
)

type CauseService interface {
	GetAllCauses(limit int, offset int) ([]models.Cause, int64, error)
	AddDiseaseCause(cause *models.Cause) (*models.Cause, error)
	UpdateCause(cause *models.Cause, authUserId string) (*models.Cause, error)
	DeleteCause(causeId uint64, authUserId string) error
	GetCauseAuditRecord(causeId uint64, causeAuditId uint64) ([]models.CauseAudit, error)
	GetAllCauseAuditRecord(limit, offset int) ([]models.CauseAudit, int64, error)
	AddDiseaseCauseMapping(DCMapping *models.DiseaseCauseMapping) error

	//Cause Type
	GetAllCauseTypes(limit int, offset int, isDeleted int) ([]models.CauseTypeMaster, int64, error)
	AddCauseType(causeType *models.CauseTypeMaster) (*models.CauseTypeMaster, error)
	UpdateCauseType(causeType *models.CauseTypeMaster, authUserId string) (*models.CauseTypeMaster, error)
	DeleteCauseType(causeTypeId uint64, authUserId string) error
	GetCauseTypeAuditRecord(causeTypeId uint64, causeTypeAuditId uint64) ([]models.CauseTypeAudit, error)
	GetAllCauseTypeAuditRecord(limit int, offset int) ([]models.CauseTypeAudit, int64, error)
}

type CauseServiceImpl struct {
	causeRepo repository.CauseRepository
}

func NewCauseService(repo repository.CauseRepository) CauseService {
	return &CauseServiceImpl{causeRepo: repo}
}

// GetAllCauses implements CauseService.
func (c *CauseServiceImpl) GetAllCauses(limit int, offset int) ([]models.Cause, int64, error) {
	return c.causeRepo.GetAllCauses(limit, offset)
}

func (c *CauseServiceImpl) AddDiseaseCause(cause *models.Cause) (*models.Cause, error) {
	return c.causeRepo.AddDiseaseCause(cause)
}

func (c *CauseServiceImpl) UpdateCause(cause *models.Cause, authUserId string) (*models.Cause, error) {
	return c.causeRepo.UpdateCause(cause, authUserId)
}
func (s *CauseServiceImpl) DeleteCause(causeId uint64, authUserId string) error {
	return s.causeRepo.DeleteCause(causeId, authUserId)
}

func (s *CauseServiceImpl) GetAllCauseAuditRecord(limit, offset int) ([]models.CauseAudit, int64, error) {
	return s.causeRepo.GetAllCauseAuditRecord(limit, offset)
}

func (s *CauseServiceImpl) GetCauseAuditRecord(causeId, causeAuditId uint64) ([]models.CauseAudit, error) {
	return s.causeRepo.GetCauseAuditRecord(causeId, causeAuditId)
}

func (c *CauseServiceImpl) AddDiseaseCauseMapping(DCMapping *models.DiseaseCauseMapping) error {
	return c.causeRepo.AddDiseaseCauseMapping(DCMapping)
}

func (c *CauseServiceImpl) AddCauseType(causeType *models.CauseTypeMaster) (*models.CauseTypeMaster, error) {
	return c.causeRepo.AddCauseType(causeType)
}

func (c *CauseServiceImpl) UpdateCauseType(causeType *models.CauseTypeMaster, authUserId string) (*models.CauseTypeMaster, error) {
	return c.causeRepo.UpdateCauseType(causeType, authUserId)
}

func (s *CauseServiceImpl) DeleteCauseType(causeTypeId uint64, authUserId string) error {
	return s.causeRepo.DeleteCauseType(causeTypeId, authUserId)
}

func (s *CauseServiceImpl) GetAllCauseTypes(limit int, offset int, isDeleted int) ([]models.CauseTypeMaster, int64, error) {
	return s.causeRepo.GetAllCauseTypes(limit, offset, isDeleted)
}

func (s *CauseServiceImpl) GetCauseTypeAuditRecord(causeTypeId uint64, causeTypeAuditId uint64) ([]models.CauseTypeAudit, error) {
	return s.causeRepo.GetCauseTypeAuditRecord(causeTypeId, causeTypeAuditId)
}

func (s *CauseServiceImpl) GetAllCauseTypeAuditRecord(limit int, offset int) ([]models.CauseTypeAudit, int64, error) {
	return s.causeRepo.GetAllCauseTypeAuditRecord(limit, offset)
}
