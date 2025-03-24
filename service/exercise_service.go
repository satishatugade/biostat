package service

import (
	"biostat/models"
	"biostat/repository"
)

type ExerciseService interface {
	CreateExercise(exercise *models.Exercise) error
	GetExercises(limit, offset int) ([]models.Exercise, int64, error)
	GetExerciseByID(id string) (models.Exercise, error)
	UpdateExercise(id string, exercise *models.Exercise) error
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

func (e *ExerciseServiceImpl) GetExerciseByID(id string) (models.Exercise, error) {
	return e.exerciseRepo.GetExerciseByID(id)
}

func (e *ExerciseServiceImpl) UpdateExercise(id string, exercise *models.Exercise) error {
	return e.exerciseRepo.UpdateExercise(id, exercise)
}
