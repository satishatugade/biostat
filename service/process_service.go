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
	GetUserRecentProcesses(userID uint64, processKey string) ([]models.ProcessStatusResponse, error)
	GetUserActivityLog(userID uint64, limit, offset int) ([]models.ProcessStatusResponse, int64, error)

	UpdateProcessStatusInRedis(processID uuid.UUID, status string, statusMsg string, step string, completed bool) error
	StartProcessInRedis(userID uint64, processType, entityID, entityType string, step string) (uuid.UUID, uuid.UUID)
	AddOrUpdateStepLogInRedis(processID uuid.UUID, newLog models.ProcessStepLog) error
	LogStep(processID uuid.UUID, step, status, message, errorMsg string, recordId *uint64, totalrecord *int, successCount *int, failedCount *int)
	LogStepAndFail(processID uuid.UUID, step, status, message, errorMsg string)
}

type ProcessStatusServiceImpl struct {
	repo        repository.ProcessStatusRepository
	redisClient *redis.Client
}

func NewProcessStatusService(repo repository.ProcessStatusRepository, redisClient *redis.Client) ProcessStatusService {
	return &ProcessStatusServiceImpl{repo: repo, redisClient: redisClient}
}

// func (s *ProcessStatusServiceImpl) AddOrUpdateStepLogInRedis(processID uuid.UUID, newLog models.ProcessStepLog) error {
// 	ctx := context.Background()
// 	key := "process_status:" + processID.String()

// 	raw, err := s.redisClient.Get(ctx, key).Bytes()
// 	if err != nil {
// 		return err
// 	}

// 	var processStatus models.ProcessStatus
// 	if err := json.Unmarshal(raw, &processStatus); err != nil {
// 		return err
// 	}

// 	updated := false
// 	for i, logEntry := range processStatus.ActivityLog {
// 		if logEntry.Step == newLog.Step {
// 			// Update existing step log fields as needed
// 			processStatus.ActivityLog[i].Status = newLog.Status
// 			processStatus.ActivityLog[i].Message = newLog.Message
// 			processStatus.ActivityLog[i].RecordIndex = newLog.RecordIndex
// 			processStatus.ActivityLog[i].TotalRecords = newLog.TotalRecords
// 			processStatus.ActivityLog[i].SuccessCount = newLog.SuccessCount
// 			processStatus.ActivityLog[i].FailedCount = newLog.FailedCount
// 			processStatus.ActivityLog[i].Error = newLog.Error
// 			processStatus.ActivityLog[i].StepUpdatedAt = time.Now()
// 			updated = true
// 			break
// 		}
// 	}

// 	if !updated {
// 		// Append log if step not found
// 		newLog.StepStartedAt = time.Now()
// 		newLog.StepUpdatedAt = time.Now()
// 		processStatus.ActivityLog = append(processStatus.ActivityLog, newLog)
// 	}

// 	// update overall step and status
// 	processStatus.Step = newLog.Step
// 	processStatus.Status = newLog.Status
// 	processStatus.StatusMessage = newLog.Message
// 	processStatus.UpdatedAt = time.Now()

// 	updatedData, err := json.Marshal(processStatus)
// 	if err != nil {
// 		return err
// 	}

// 	return s.redisClient.Set(ctx, key, updatedData, 0).Err()
// }

func (s *ProcessStatusServiceImpl) AddOrUpdateStepLogInRedis(ProcessID uuid.UUID, input models.ProcessStepLog) error {
	ctx := context.Background()
	key := "process_status:" + ProcessID.String()

	raw, err := s.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	var processStatus models.ProcessStatus
	if err := json.Unmarshal(raw, &processStatus); err != nil {
		return err
	}

	now := time.Now()
	foundStep := false
	foundRecord := false

	for i, stepLog := range processStatus.ActivityLog {
		if stepLog.Step == input.Step {
			foundStep = true

			// Search for matching record_id
			for j, recordLog := range stepLog.RecordLogs {
				if recordLog.RecordID == input.RecordIndex {
					// Update existing record log
					processStatus.ActivityLog[i].RecordLogs[j].Status = input.Status
					processStatus.ActivityLog[i].RecordLogs[j].Message = input.Message
					processStatus.ActivityLog[i].RecordLogs[j].Error = input.Error
					processStatus.ActivityLog[i].RecordLogs[j].CompletedAt = &now
					foundRecord = true
					break
				}
			}

			if !foundRecord {
				// Append new record log
				recordLog := models.ProcessStepRecordLog{
					ProcessStepRecordLogId: uuid.New(),
					ProcessStepLogID:       stepLog.ProcessStepLogId,
					RecordID:               input.RecordIndex,
					RecordIndex:            input.RecordIndex,
					Status:                 input.Status,
					Message:                input.Message,
					Error:                  input.Error,
					StartedAt:              now,
				}
				processStatus.ActivityLog[i].RecordLogs = append(processStatus.ActivityLog[i].RecordLogs, recordLog)
			}

			// Update step level metadata
			processStatus.ActivityLog[i].Status = input.Status
			processStatus.ActivityLog[i].Message = input.Message
			processStatus.ActivityLog[i].StepUpdatedAt = now
			break
		}
	}

	if !foundStep {
		// Create a new step log with the record log
		newStepLogID := uuid.New()
		newStepLog := models.ProcessStepLog{
			ProcessStepLogId: newStepLogID,
			ProcessStatusID:  ProcessID,
			Step:             input.Step,
			Status:           input.Status,
			Message:          input.Message,
			StepStartedAt:    now,
			StepUpdatedAt:    now,
			RecordLogs: []models.ProcessStepRecordLog{
				{
					ProcessStepRecordLogId: uuid.New(),
					ProcessStepLogID:       newStepLogID,
					RecordID:               input.RecordIndex,
					RecordIndex:            input.RecordIndex,
					Status:                 input.Status,
					Message:                input.Message,
					Error:                  input.Error,
					StartedAt:              now,
				},
			},
		}
		processStatus.ActivityLog = append(processStatus.ActivityLog, newStepLog)
	}

	// Update process-level metadata
	processStatus.Step = input.Step
	processStatus.Status = input.Status
	processStatus.StatusMessage = input.Message
	processStatus.UpdatedAt = now

	updated, err := json.Marshal(processStatus)
	if err != nil {
		return err
	}

	return s.redisClient.Set(ctx, key, updated, 0).Err()
}

// func (s *ProcessStatusServiceImpl) AddOrUpdateStepLogInRedis(
// 	processID uuid.UUID,
// 	step string,
// 	recordID uint64,
// 	recordIndex int64,
// 	status, message string,
// 	errorMsg *string,
// ) error {
// 	ctx := context.Background()
// 	key := "process_status:" + processID.String()

// 	raw, err := s.redisClient.Get(ctx, key).Bytes()
// 	if err != nil {
// 		return err
// 	}

// 	var processStatus models.ProcessStatus
// 	if err := json.Unmarshal(raw, &processStatus); err != nil {
// 		return err
// 	}

// 	now := time.Now()
// 	foundStep := false
// 	foundRecord := false

// 	for i, stepLog := range processStatus.ActivityLog {
// 		if stepLog.Step == step {
// 			foundStep = true
// 			// Search for matching record_id
// 			for j, recordLog := range stepLog.RecordLogs {
// 				if recordLog.RecordID == recordID {
// 					// Update existing record log
// 					processStatus.ActivityLog[i].RecordLogs[j].Status = status
// 					processStatus.ActivityLog[i].RecordLogs[j].Message = message
// 					processStatus.ActivityLog[i].RecordLogs[j].Error = errorMsg
// 					processStatus.ActivityLog[i].RecordLogs[j].CompletedAt = &now
// 					foundRecord = true
// 					break
// 				}
// 			}

// 			if !foundRecord {
// 				// Append new record log
// 				recordLog := models.ProcessStepRecordLog{
// 					ProcessStepRecordLogId: uuid.New(),
// 					ProcessStepLogID:       stepLog.ProcessStepLogId,
// 					RecordID:               recordID,
// 					RecordIndex:            recordIndex,
// 					Status:                 status,
// 					Message:                message,
// 					Error:                  errorMsg,
// 					StartedAt:              now,
// 				}
// 				processStatus.ActivityLog[i].RecordLogs = append(processStatus.ActivityLog[i].RecordLogs, recordLog)
// 			}

// 			// Update step level metadata
// 			processStatus.ActivityLog[i].Status = status
// 			processStatus.ActivityLog[i].Message = message
// 			processStatus.ActivityLog[i].StepUpdatedAt = now
// 			break
// 		}
// 	}

// 	if !foundStep {
// 		// Create a new step log with record log
// 		newStepLogID := uuid.New()
// 		newStepLog := models.ProcessStepLog{
// 			ProcessStepLogId: newStepLogID,
// 			ProcessStatusID:  processID,
// 			Step:             step,
// 			Status:           status,
// 			Message:          message,
// 			StepStartedAt:    now,
// 			StepUpdatedAt:    now,
// 			RecordLogs: []models.ProcessStepRecordLog{
// 				{
// 					ProcessStepRecordLogId: uuid.New(),
// 					ProcessStepLogID:       newStepLogID,
// 					RecordID:               recordID,
// 					RecordIndex:            recordIndex,
// 					Status:                 status,
// 					Message:                message,
// 					Error:                  errorMsg,
// 					StartedAt:              now,
// 				},
// 			},
// 		}
// 		processStatus.ActivityLog = append(processStatus.ActivityLog, newStepLog)
// 	}

// 	// Update main process status as well
// 	processStatus.Step = step
// 	processStatus.Status = status
// 	processStatus.StatusMessage = message
// 	processStatus.UpdatedAt = now

// 	updated, err := json.Marshal(processStatus)
// 	if err != nil {
// 		return err
// 	}

// 	return s.redisClient.Set(ctx, key, updated, 0).Err()
// }

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

// func (s *ProcessStatusServiceImpl) UpdateProcessStatusInRedis(
// 	processID uuid.UUID,
// 	status string,
// 	statusMsg string,
// 	step string,
// 	completed bool,
// ) error {
// 	ctx := context.Background()
// 	key := "process_status:" + processID.String()

// 	raw, err := s.redisClient.Get(ctx, key).Bytes()
// 	if err != nil {
// 		return err
// 	}

// 	var processStatus models.ProcessStatus
// 	if err := json.Unmarshal(raw, &processStatus); err != nil {
// 		return err
// 	}

// 	processStatus.Status = status
// 	processStatus.StatusMessage = statusMsg
// 	processStatus.Step = step
// 	processStatus.UpdatedAt = time.Now()
// 	if completed {
// 		now := time.Now()
// 		processStatus.CompletedAt = &now
// 	}

// 	updated, err := json.Marshal(processStatus)
// 	if err != nil {
// 		return err
// 	}

// 	return s.redisClient.Set(ctx, key, updated, 0).Err()
// }

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
	if err == redis.Nil {
		return fmt.Errorf("process status not found in Redis for ID %s", processID)
	} else if err != nil {
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

	// Set with no expiration (or define a TTL if needed)
	return s.redisClient.Set(ctx, key, updated, 0).Err()
}

func (s *ProcessStatusServiceImpl) LogStep(processID uuid.UUID, step, status, message, errorMsg string, recordId *uint64, totalRecord *int, successCount *int, failedCount *int) {
	var errPtr *string
	if errorMsg != "" {
		errPtr = &errorMsg
	}
	logEntry := models.ProcessStepLog{
		ProcessStepLogId: uuid.New(),
		ProcessStatusID:  processID,
		Step:             step,
		Status:           status,
		Message:          message,
		RecordIndex:      recordId,
		TotalRecords:     totalRecord,
		SuccessCount:     successCount,
		FailedCount:      failedCount,
		Error:            errPtr,
		StepStartedAt:    time.Now(),
		StepUpdatedAt:    time.Now(),
	}
	if err := s.AddOrUpdateStepLogInRedis(processID, logEntry); err != nil {
		log.Printf("[Warning] failed to add/update step log in Redis: %v", err)
	}
	if err := s.UpdateProcessStatusInRedis(processID, status, message, step, false); err != nil {
		log.Printf("[Warning] failed to update process status in Redis: %v", err)
	}
}

func (s *ProcessStatusServiceImpl) LogStepAndFail(processID uuid.UUID, step, status, message, errorMsg string) {
	s.LogStep(processID, step, status, message, errorMsg, nil, nil, nil, nil)
	if err := s.UpdateProcessStatusInRedis(processID, constant.Failure, message, step, true); err != nil {
		log.Printf("[Error] failed to mark process failure in Redis: %v", err)
	}
}
