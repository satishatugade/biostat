package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type ExerciseRepository interface {
	CreateExercise(exercise *models.Exercise) error
	GetExercises(limit, offset int) ([]models.Exercise, int64, error)
	GetExerciseByID(id string) (models.Exercise, error)
	UpdateExercise(id string, exercise *models.Exercise) error
}

type ExerciseRepositoryImpl struct {
	db *gorm.DB
}

func NewExerciseRepository(db *gorm.DB) ExerciseRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &ExerciseRepositoryImpl{db: db}
}

// CreateExercise implements ExerciseRepository.
func (e *ExerciseRepositoryImpl) CreateExercise(exercise *models.Exercise) error {
	return e.db.Create(exercise).Error
}
func (e *ExerciseRepositoryImpl) GetExercises(limit, offset int) ([]models.Exercise, int64, error) {
	var exercises []models.Exercise
	var totalRecords int64

	if err := e.db.Model(&models.Exercise{}).Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	if err := e.db.Limit(limit).Offset(offset).Find(&exercises).Error; err != nil {
		return nil, 0, err
	}

	return exercises, totalRecords, nil
}

func (e *ExerciseRepositoryImpl) GetExerciseByID(id string) (models.Exercise, error) {
	var exercise models.Exercise
	if err := e.db.First(&exercise, "exercise_id = ?", id).Error; err != nil {
		return exercise, err
	}
	return exercise, nil
}

func (e *ExerciseRepositoryImpl) UpdateExercise(id string, exercise *models.Exercise) error {
	return e.db.Model(&models.Exercise{}).Where("exercise_id = ?", id).Updates(exercise).Error
}
