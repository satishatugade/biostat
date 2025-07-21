package repository

import (
	"biostat/constant"
	"biostat/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProcessStatusRepository interface {
	CreateProcess(process *models.ProcessStatus) error
	UpdateProcess(processID uuid.UUID, updates map[string]interface{}) error
	GetRecentUserProcesses(userID uint64, recentMinutes int) ([]models.ProcessStatus, error)
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

func (r *ProcessStatusRepositoryImpl) GetRecentUserProcesses(userID uint64, recentMinutes int) ([]models.ProcessStatus, error) {
	var processes []models.ProcessStatus
	cutoffTime := time.Now().Add(-time.Duration(recentMinutes) * time.Minute)

	err := r.db.Where("user_id = ? AND (status = ? OR status = ? OR (status = ? AND updated_at >= ?))",
		userID, constant.Running, constant.Failure, constant.Success, cutoffTime).
		Order("updated_at desc").
		Find(&processes).Error

	return processes, err
}
