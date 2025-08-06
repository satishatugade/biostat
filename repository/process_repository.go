package repository

import (
	"biostat/constant"
	"biostat/models"
	"biostat/utils"
	"time"

	"gorm.io/gorm"
)

type ProcessStatusRepository interface {
	GetRecentUserProcesses(userID uint64, recentMinutes int) ([]models.ProcessStatus, error)
	FetchActivityLogsByUserID(userID uint64, limit, offset int) ([]models.ProcessStatusResponse, int64, error)
}

type ProcessStatusRepositoryImpl struct {
	db *gorm.DB
}

func NewProcessStatusRepository(db *gorm.DB) ProcessStatusRepository {
	return &ProcessStatusRepositoryImpl{db}
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

func (r *ProcessStatusRepositoryImpl) FetchActivityLogsByUserID(userID uint64, limit, offset int) ([]models.ProcessStatusResponse, int64, error) {
	var logs []models.ProcessStatus
	var total int64

	// Count total
	if err := r.db.Model(&models.ProcessStatus{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch logs with steps
	err := r.db.Preload("ActivityLog", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_started_at ASC")
	}).
		Where("user_id = ?", userID).
		Order("started_at DESC").
		Limit(limit).Offset(offset).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}
	var response []models.ProcessStatusResponse
	for _, log := range logs {
		var steps []models.ProcessStepLogResponse
		for _, step := range log.ActivityLog {
			steps = append(steps, models.ProcessStepLogResponse{
				ProcessStepLogId: step.ProcessStepLogId.String(),
				StepName:         step.Step,
				StepStatus:       step.Status,
				RecordIndex:      step.RecordIndex,
				TotalRecords:     step.TotalRecords,
				SuccessCount:     step.SuccessCount,
				FailedCount:      step.FailedCount,
				Message:          step.Message,
				Error:            step.Error,
				StartedAt:        utils.FormatDateTime(&step.StepStartedAt),
				CompletedAt:      utils.FormatDateTime(&step.StepUpdatedAt),
			})
		}

		response = append(response, models.ProcessStatusResponse{
			ProcessStatusID: log.ProcessStatusID.String(),
			UserID:          log.UserID,
			ProcessType:     log.ProcessType,
			EntityID:        log.EntityID,
			EntityType:      log.EntityType,
			StartedAt:       utils.FormatDateTime(&log.StartedAt),
			CompletedAt:     utils.FormatDateTime(log.CompletedAt),
			Status:          log.Status,
			ActivityLog:     steps,
		})
	}
	return response, total, nil
}
