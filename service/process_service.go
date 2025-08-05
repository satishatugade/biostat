package service

import (
	"biostat/constant"
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type ProcessStatusService interface {
	StartProcess(processID uuid.UUID, userID uint64, processType, entityID, entityType string, step string) (uuid.UUID, uuid.UUID)
	CreateProcessStepLog(processID uuid.UUID, status string, message, step *string, totalRecordFound *int) uuid.UUID
	UpdateProcess(processID uuid.UUID, status string, entityID *string, message *string, step *string, completed bool)
	UpdateProcessStepLog(processStepID uuid.UUID, status string, message, step *string, recordCount *int, completed bool, err *error, successCount, failedCount *int) error
	GetUserRecentProcesses(userID uint64, processKey string) ([]models.ProcessStatusResponse, error)
	GetUserActivityLog(userID uint64, limit, offset int) ([]models.ProcessStatusResponse, int64, error)

	UpdateProcessStatusInRedis(processID uuid.UUID, status string, statusMsg string, step string, completed bool) error
	StartProcessInRedis(userID uint64, processType, entityID, entityType string, step string) (uuid.UUID, uuid.UUID)
	StartProcessRedis(processID uuid.UUID, userID uint64, processType, entityID, entityType string, step string) (string, error)
	UpdateProcessRedis(key, status string, entityID *string, message string, step string, completed bool) error
	AddOrUpdateStepLogInRedis(processID uuid.UUID, newLog models.ProcessStepLog) error
}

type ProcessStatusServiceImpl struct {
	repo        repository.ProcessStatusRepository
	redisClient *redis.Client
}

func NewProcessStatusService(repo repository.ProcessStatusRepository, redisClient *redis.Client) ProcessStatusService {
	return &ProcessStatusServiceImpl{repo: repo, redisClient: redisClient}
}

func (s *ProcessStatusServiceImpl) StartProcess(processID uuid.UUID, userID uint64, processType, entityID, entityType string, step string) (uuid.UUID, uuid.UUID) {
	processStepId := uuid.New()
	// go func() {
	// 	defer s.suppressError()
	status := &models.ProcessStatus{
		ProcessStatusID: processID,
		UserID:          userID,
		ProcessType:     processType,
		EntityID:        entityID,
		EntityType:      entityType,
		Status:          constant.Running,
		Step:            step,
		StartedAt:       time.Now(),
	}

	if err := s.repo.CreateProcess(status); err != nil {
		log.Printf("@StartProcess->CreateProcess failed to create: %v", err)
	}
	stepLog := &models.ProcessStepLog{
		ProcessStepLogId: processStepId,
		ProcessStatusID:  processID,
		Step:             step,
		Status:           constant.Running,
		Message:          string(constant.ProcessStarted),
		StepStartedAt:    time.Now(),
	}

	if err := s.repo.CreateProcessStepLog(stepLog); err != nil {
		log.Printf("@StartProcess->CreateProcessStepLog failed: %v", err)
	}
	// }()
	return processID, processStepId
}

func (s *ProcessStatusServiceImpl) CreateProcessStepLog(processID uuid.UUID, status string, message, step *string, totalRecordFound *int) uuid.UUID {
	processStepId := uuid.New()
	go func() {
		defer s.suppressError()

		logEntry := &models.ProcessStepLog{
			ProcessStepLogId: processStepId,
			ProcessStatusID:  processID,
			TotalRecords:     totalRecordFound,
			Step:             utils.SafeDeref(step),
			Status:           status,
			Message:          utils.SafeDeref(message),
			StepStartedAt:    time.Now(),
		}

		if err := s.repo.CreateProcessStepLog(logEntry); err != nil {
			log.Printf("@CreateProcessStepLog -> failed to create: %v", err)
		}
	}()
	return processStepId
}

func (s *ProcessStatusServiceImpl) UpdateProcessStepLog(processStepID uuid.UUID, status string, message, step *string, count *int, completed bool, err *error, succeCount, failedCount *int) error {
	// go func() {
	// 	defer s.suppressError()

	updates := map[string]interface{}{
		"status": status,
	}

	if message != nil {
		updates["message"] = *message
	}
	if step != nil {
		updates["step"] = *step
	}
	if completed {
		updates["step_updated_at"] = time.Now()
	}
	if count != nil {
		updates["total_records"] = count
	}
	if err != nil && *err != nil {
		updates["error"] = (*err).Error()
	}
	if succeCount != nil {
		updates["success_count"] = succeCount
	}
	if failedCount != nil {
		updates["failed_count"] = failedCount
	}
	err1 := s.repo.UpdateLatestProcessStepLog(processStepID, updates)
	if err1 != nil {
		log.Printf("@UpdateProcessStepLog -> failed to update: %v", err)
		return err1
	}
	return nil
	// }()
}

func (s *ProcessStatusServiceImpl) AddOrUpdateStepLogInRedis(processID uuid.UUID, newLog models.ProcessStepLog) error {
	ctx := context.Background()
	key := "process_status:" + processID.String()

	raw, err := s.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	var processStatus models.ProcessStatus
	if err := json.Unmarshal(raw, &processStatus); err != nil {
		return err
	}

	updated := false
	for i, logEntry := range processStatus.ActivityLog {
		if logEntry.Step == newLog.Step {
			// Update existing step log fields as needed
			processStatus.ActivityLog[i].Status = newLog.Status
			processStatus.ActivityLog[i].Message = newLog.Message
			processStatus.ActivityLog[i].RecordIndex = newLog.RecordIndex
			processStatus.ActivityLog[i].TotalRecords = newLog.TotalRecords
			processStatus.ActivityLog[i].SuccessCount = newLog.SuccessCount
			processStatus.ActivityLog[i].FailedCount = newLog.FailedCount
			processStatus.ActivityLog[i].Error = newLog.Error
			processStatus.ActivityLog[i].StepUpdatedAt = time.Now()
			updated = true
			break
		}
	}

	if !updated {
		// Append log if step not found
		newLog.StepStartedAt = time.Now()
		newLog.StepUpdatedAt = time.Now()
		processStatus.ActivityLog = append(processStatus.ActivityLog, newLog)
	}

	// Optionally update overall step and status if needed, e.g.:
	processStatus.Step = newLog.Step
	processStatus.Status = newLog.Status
	processStatus.StatusMessage = newLog.Message
	processStatus.UpdatedAt = time.Now()

	updatedData, err := json.Marshal(processStatus)
	if err != nil {
		return err
	}

	return s.redisClient.Set(ctx, key, updatedData, 0).Err()
}

func (s *ProcessStatusServiceImpl) UpdateProcess(processID uuid.UUID, status string, entityID *string, message *string, step *string, completed bool) {
	defer s.suppressError()

	updates := map[string]interface{}{
		"status": status,
	}
	if entityID != nil {
		updates["entity_id"] = *entityID
	}
	if message != nil {
		updates["status_message"] = *message
	}
	if step != nil {
		updates["step"] = *step
	}
	if completed {
		updates["completed_at"] = time.Now()
	}

	if err := s.repo.UpdateProcess(processID, updates); err != nil {
		log.Printf("@UpdateProcess->UpdateProcess failed to update: %v", err)
	}
}

func (s *ProcessStatusServiceImpl) suppressError() {
	if r := recover(); r != nil {
		log.Printf("@suppressError [ProcessStatus] panic recovered: %v", r)
	}
}

func (s *ProcessStatusServiceImpl) StartProcessRedis(processID uuid.UUID, userID uint64, processType, entityID, entityType string, step string) (string, error) {
	status := &models.ProcessStatus{
		ProcessStatusID: processID,
		UserID:          userID,
		ProcessType:     processType,
		EntityID:        entityID,
		EntityType:      entityType,
		Status:          constant.Running,
		Step:            step,
		StatusMessage:   "Initializing...",
		StartedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	key := fmt.Sprintf("process_status:%d:%s:%s", userID, processType, processID)
	data, err := json.Marshal(status)
	if err != nil {
		log.Println("@StartProcessRedis->Marshal", err)
		return "", err
	}
	err = s.redisClient.Set(context.Background(), key, data, 2*time.Hour).Err()
	if err != nil {
		log.Println("@StartProcessRedis->redisClient.Set", err)
		return "", err
	}
	return key, nil
}

func (s *ProcessStatusServiceImpl) StartProcessInRedis(userID uint64, processType, entityID, entityType string, step string) (uuid.UUID, uuid.UUID) {
	processID := uuid.New()
	processStepId := uuid.New()
	status := &models.ProcessStatus{
		ProcessStatusID: processID,
		UserID:          userID,
		ProcessType:     processType,
		EntityID:        entityID,
		EntityType:      entityType,
		Status:          constant.Running,
		Step:            step,
		StartedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ActivityLog:     []models.ProcessStepLog{},
	}

	stepLog := models.ProcessStepLog{
		ProcessStepLogId: processStepId,
		ProcessStatusID:  processID,
		Step:             step,
		Status:           constant.Running,
		Message:          string(constant.ProcessStarted),
		StepStartedAt:    time.Now(),
		StepUpdatedAt:    time.Now(),
	}
	status.ActivityLog = append(status.ActivityLog, stepLog)

	jsonData, err := json.Marshal(status)
	if err != nil {
		log.Printf("@StartProcessInRedis marshal failed: %v", err)
		return uuid.Nil, uuid.Nil
	}

	// Store the entire status struct (with first log) into Redis as a JSON blob
	redisSetErr := s.redisClient.Set(context.Background(), "process_status:"+processID.String(), jsonData, 0).Err()
	if redisSetErr != nil {
		log.Printf("@StartProcessInRedis->Redis SET failed: %v", redisSetErr)
	}
	log.Println("StartProcessInRedis processID ", processID)
	return processID, processStepId
}

func (s *ProcessStatusServiceImpl) UpdateProcessRedis(key, status string, entityID *string, message string, step string, completed bool) error {
	log.Println("Updating Redis :", key)
	ctx := context.Background()
	val, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		log.Println("@UpdateProcessRedis->redisClient.Get", err)
		return err
	}

	var process models.ProcessStatus
	err = json.Unmarshal([]byte(val), &process)
	if err != nil {
		log.Println("@UpdateProcessRedis->Unmarshal", err)
		return err
	}
	process.Status = status
	process.Step = step
	process.StatusMessage = message
	process.UpdatedAt = time.Now()
	if completed {
		now := time.Now()
		process.CompletedAt = &now
	}
	if entityID != nil {
		process.EntityID = *entityID
	}
	data, err := json.Marshal(process)
	if err != nil {
		log.Println("@UpdateProcessRedis->Marshal", err)
		return err
	}
	err = s.redisClient.Set(ctx, key, data, 2*time.Hour).Err()
	if err != nil {
		log.Println("@UpdateProcessRedis->redisClient.Set", err)
		return err
	}
	return nil
}

func (s *ProcessStatusServiceImpl) GetUserRecentProcesses(userID uint64, processKey string) ([]models.ProcessStatusResponse, error) {
	ctx := context.Background()
	var processList []models.ProcessStatus
	log.Println("GetUserRecentProcesses process key : ", processKey)
	pattern := "process_status:*"
	if processKey != "" {
		pattern = fmt.Sprintf("process_status:*%s*", processKey)
	}

	var cursor uint64
	for {
		keys, nextCursor, err := s.redisClient.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			log.Printf("Scanning key: %s", key)

			val, err := s.redisClient.Get(ctx, key).Result()
			if err != nil {
				log.Printf("Redis GET error: %v", err)
				continue
			}

			var proc models.ProcessStatus
			if err := json.Unmarshal([]byte(val), &proc); err != nil {
				log.Printf("Unmarshal failed for key %s: %v", key, err)
				continue
			}

			log.Printf("UserID in process: %d", proc.UserID)

			if proc.UserID == userID {
				processList = append(processList, proc)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	var response []models.ProcessStatusResponse
	for _, log := range processList {
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
	log.Println("GetUserRecentProcesses ProcessList response: ", response)
	return response, nil
}

func (s *ProcessStatusServiceImpl) GetUserActivityLog(userID uint64, limit, offset int) ([]models.ProcessStatusResponse, int64, error) {
	return s.repo.FetchActivityLogsByUserID(userID, limit, offset)
}

func (s *ProcessStatusServiceImpl) UpdateProcessStatusInRedis(
	processID uuid.UUID,
	status string,
	statusMsg string,
	step string,
	completed bool,
) error {
	ctx := context.Background()
	key := "process_status:" + processID.String()

	raw, err := s.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	var processStatus models.ProcessStatus
	if err := json.Unmarshal(raw, &processStatus); err != nil {
		return err
	}

	processStatus.Status = status
	processStatus.StatusMessage = statusMsg
	processStatus.Step = step
	processStatus.UpdatedAt = time.Now()
	if completed {
		now := time.Now()
		processStatus.CompletedAt = &now
	}

	updated, err := json.Marshal(processStatus)
	if err != nil {
		return err
	}

	return s.redisClient.Set(ctx, key, updated, 0).Err()
}
