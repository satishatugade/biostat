package service

import (
	"biostat/models"
	"biostat/repository"
	"log"
	"time"

	"github.com/google/uuid"
)

type ProcessStatusService interface {
	StartProcess(userID uint64, processType, entityID, entityType string, step string) uuid.UUID
	UpdateProcess(processID uuid.UUID, status string, entityID *string, message *string, step *string, completed bool)
}

type ProcessStatusServiceImpl struct {
	repo repository.ProcessStatusRepository
}

func NewProcessStatusService(repo repository.ProcessStatusRepository) ProcessStatusService {
	return &ProcessStatusServiceImpl{repo: repo}
}

func (s *ProcessStatusServiceImpl) StartProcess(userID uint64, processType, entityID, entityType string, step string) uuid.UUID {
	processID := uuid.New()
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
			log.Println("@StartProcess->CreateProcess failed to create: %v", err)
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
