package service

import (
	"biostat/models"
	"biostat/repository"
)

type CauseService interface {
	GetAllCauses(limit int, offset int) ([]models.Cause, int64, error)
	AddDiseaseCause(cause *models.Cause) error
	UpdateCause(cause *models.Cause) error
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

func (c *CauseServiceImpl) UpdateCause(cause *models.Cause) error {
	return c.causeRepo.UpdateCause(cause)
}
