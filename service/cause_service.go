package service

import (
	"biostat/models"
	"biostat/repository"
)

type CauseService interface {
	GetAllCauses(limit int, offset int) ([]models.Cause, int64, error)
	AddDiseaseCause(cause *models.Cause) error
	UpdateCause(cause *models.Cause, authUserId string) error
	DeleteCause(causeId uint64, authUserId string) error
	GetCauseAuditRecord(causeId uint64, causeAuditId uint64) ([]models.CauseAudit, error)
	GetAllCauseAuditRecord(page, limit int) ([]models.CauseAudit, int64, error)
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

func (c *CauseServiceImpl) AddDiseaseCause(cause *models.Cause) error {
	return c.causeRepo.AddDiseaseCause(cause)
}

func (c *CauseServiceImpl) UpdateCause(cause *models.Cause, authUserId string) error {
	return c.causeRepo.UpdateCause(cause, authUserId)
}
func (s *CauseServiceImpl) DeleteCause(causeId uint64, authUserId string) error {
	return s.causeRepo.DeleteCause(causeId, authUserId)
}

func (s *CauseServiceImpl) GetAllCauseAuditRecord(page, limit int) ([]models.CauseAudit, int64, error) {
	return s.causeRepo.GetAllCauseAuditRecord(page, limit)
}

func (s *CauseServiceImpl) GetCauseAuditRecord(causeId, causeAuditId uint64) ([]models.CauseAudit, error) {
	return s.causeRepo.GetCauseAuditRecord(causeId, causeAuditId)
}
