package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type CauseRepository interface {
	GetAllCauses(limit int, offset int) ([]models.Cause, int64, error)
	AddDiseaseCause(cause *models.Cause) error
	UpdateCause(cause *models.Cause) error
}

type CauseRepositoryImpl struct {
	db *gorm.DB
}

func NewCauseRepository(db *gorm.DB) CauseRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &CauseRepositoryImpl{db: db}
}

// GetAllCauses implements CauseRepository.
func (c *CauseRepositoryImpl) GetAllCauses(limit int, offset int) ([]models.Cause, int64, error) {

	var causes []models.Cause
	var totalRecords int64

	query := c.db.Model(&models.Cause{})
	query.Count(&totalRecords)

	err := query.Limit(limit).Offset(offset).Find(&causes).Error
	if err != nil {
		return nil, 0, err
	}
	return causes, totalRecords, nil
}

func (c *CauseRepositoryImpl) AddDiseaseCause(cause *models.Cause) error {
	return c.db.Create(cause).Error
}

func (c *CauseRepositoryImpl) UpdateCause(cause *models.Cause) error {
	return c.db.Model(&models.Cause{}).Where("cause_id = ?", cause.CauseId).Updates(cause).Error
}
