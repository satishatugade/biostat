package repository

import (
	"biostat/constant"
	"biostat/models"
	"biostat/utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProcessStatusRepository interface {
	CreateProcess(process *models.ProcessStatus) error
	CreateProcessStepLog(stepLog *models.ProcessStepLog) error
	UpdateProcess(processID uuid.UUID, updates map[string]interface{}) error
	UpdateLatestProcessStepLog(processStepID uuid.UUID, updates map[string]interface{}) error
	GetRecentUserProcesses(userID uint64, recentMinutes int) ([]models.ProcessStatus, error)
	FetchActivityLogsByUserID(userID uint64, limit, offset int) ([]models.ProcessStatusResponse, int64, error)
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

func (r *ProcessStatusRepositoryImpl) CreateProcessStepLog(stepLog *models.ProcessStepLog) error {
	return r.db.Create(stepLog).Error
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
	err := r.db.Preload("ActivityLog").
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

func (r *ProcessStatusRepositoryImpl) UpdateLatestProcessStepLog(processStepID uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&models.ProcessStepLog{}).
		Where("process_step_log_id = ?", processStepID).
		Updates(updates).Error
}
