package service

import (
	"biostat/constant"
	"biostat/models"
	"biostat/repository"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type ProcessStatusService interface {
	StartProcess(processID uuid.UUID, userID uint64, processType, entityID, entityType string, step string) uuid.UUID
	UpdateProcess(processID uuid.UUID, status string, entityID *string, message *string, step *string, completed bool)
	GetUserRecentProcesses(userID uint64) ([]models.ProcessStatus, error)

	StartProcessRedis(processID uuid.UUID, userID uint64, processType, entityID, entityType string, step string) (string, error)
	UpdateProcessRedis(key, status string, entityID *string, message string, step string, completed bool) error
}

type ProcessStatusServiceImpl struct {
	repo        repository.ProcessStatusRepository
	redisClient *redis.Client
}

func NewProcessStatusService(repo repository.ProcessStatusRepository, redisClient *redis.Client) ProcessStatusService {
	return &ProcessStatusServiceImpl{repo: repo, redisClient: redisClient}
}

func (s *ProcessStatusServiceImpl) StartProcess(processID uuid.UUID, userID uint64, processType, entityID, entityType string, step string) uuid.UUID {
	// processID := uuid.New()
	go func() {
		defer s.suppressError()
		status := &models.ProcessStatus{
			ProcessStatusID: processID,
			UserID:          userID,
			ProcessType:     processType,
			EntityID:        entityID,
			EntityType:      entityType,
			Status:          "running",
			Step:            step,
			StartedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if err := s.repo.CreateProcess(status); err != nil {
			log.Printf("@StartProcess->CreateProcess failed to create: %v", err)
		}
	}()
	return processID
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

func (s *ProcessStatusServiceImpl) GetUserRecentProcesses(userID uint64) ([]models.ProcessStatus, error) {
	// return s.repo.GetRecentUserProcesses(userID, 60)
	var processList []models.ProcessStatus

	ctx := context.Background()
	pattern := fmt.Sprintf("process_status:%d:*", userID)

	var cursor uint64
	for {
		keys, nextCursor, err := s.redisClient.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			val, err := s.redisClient.Get(ctx, key).Result()
			if err != nil {
				continue
			}

			var proc models.ProcessStatus
			if err := json.Unmarshal([]byte(val), &proc); err != nil {
				continue
			}

			processList = append(processList, proc)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return processList, nil
}
