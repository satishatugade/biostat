package repository

import (
	"biostat/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProcessStatusRepository interface {
	CreateProcess(process *models.ProcessStatus) error
	UpdateProcess(processID uuid.UUID, updates map[string]interface{}) error
}

type ProcessStatusRepositoryImpl struct {
	db *gorm.DB
}

func NewProcessStatusRepository(db *gorm.DB) ProcessStatusRepository {
	return &ProcessStatusRepositoryImpl{db}
}

func (r *ProcessStatusRepositoryImpl) CreateProcess(process *models.ProcessStatus) error {
	return r.db.Create(process).Error
}

func (r *ProcessStatusRepositoryImpl) UpdateProcess(processID uuid.UUID, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.Model(&models.ProcessStatus{}).
		Where("process_status_id = ?", processID).
		Updates(updates).Error
}
