package service

import (
	"biostat/models"
	"biostat/repository"
)

type ExerciseService interface {
	CreateExercise(exercise *models.Exercise) error
	GetExercises(limit, offset int) ([]models.Exercise, int64, error)
	GetExerciseById(exerciseId uint64) (*models.Exercise, error)
	UpdateExercise(authUserId string, exercise *models.Exercise) error
	DeleteExercise(exerciseId uint64, authUserId string) error
	GetExerciseAuditRecord(exerciseId, exerciseAuditId uint64) ([]models.ExerciseAudit, error)
	GetAllExerciseAuditRecord(limit, offset int) ([]models.ExerciseAudit, int64, error)
	AddDiseaseExerciseMapping(mapping *models.DiseaseExerciseMapping) error
}

type ExerciseServiceImpl struct {
	exerciseRepo repository.ExerciseRepository
}

func NewExerciseService(repo repository.ExerciseRepository) ExerciseService {
	return &ExerciseServiceImpl{exerciseRepo: repo}
}

func (e *ExerciseServiceImpl) CreateExercise(exercise *models.Exercise) error {
	return e.exerciseRepo.CreateExercise(exercise)
}

func (e *ExerciseServiceImpl) GetExercises(limit, offset int) ([]models.Exercise, int64, error) {
	return e.exerciseRepo.GetExercises(limit, offset)
}

func (e *ExerciseServiceImpl) GetExerciseById(exerciseId uint64) (*models.Exercise, error) {
	return e.exerciseRepo.GetExerciseById(exerciseId)
}

func (e *ExerciseServiceImpl) UpdateExercise(authUserId string, exercise *models.Exercise) error {
	return e.exerciseRepo.UpdateExercise(authUserId, exercise)
}

func (e *ExerciseServiceImpl) DeleteExercise(exerciseId uint64, authUserId string) error {
	return e.exerciseRepo.DeleteExercise(exerciseId, authUserId)
}

func (s *ExerciseServiceImpl) GetExerciseAuditRecord(exerciseId, exerciseAuditId uint64) ([]models.ExerciseAudit, error) {
	return s.exerciseRepo.GetExerciseAuditRecord(exerciseId, exerciseAuditId)
}

func (s *ExerciseServiceImpl) GetAllExerciseAuditRecord(limit, offset int) ([]models.ExerciseAudit, int64, error) {
	return s.exerciseRepo.GetAllExerciseAuditRecord(limit, offset)
}
func (s *ExerciseServiceImpl) AddDiseaseExerciseMapping(mapping *models.DiseaseExerciseMapping) error {
	return s.exerciseRepo.AddDiseaseExerciseMapping(mapping)
}
