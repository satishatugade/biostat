package repository

import (
	"biostat/constant"
	"biostat/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type ExerciseRepository interface {
	CreateExercise(exercise *models.Exercise) error
	GetExercises(limit, offset int) ([]models.Exercise, int64, error)
	GetExerciseById(exerciseId uint64) (*models.Exercise, error)
	UpdateExercise(authUserId string, exercise *models.Exercise) error
	DeleteExercise(exerciseId uint64, authUserId string) error
	GetExerciseAuditRecord(exerciseId, exerciseAuditId uint64) ([]models.ExerciseAudit, error)
	GetAllExerciseAuditRecord(page, limit int) ([]models.ExerciseAudit, int64, error)
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

func (e *ExerciseRepositoryImpl) GetExerciseById(exerciseId uint64) (*models.Exercise, error) {
	var exercise models.Exercise
	if err := e.db.First(&exercise, "exercise_id = ?", exerciseId).Error; err != nil {
		return &exercise, err
	}
	return &exercise, nil
}

func (repo *ExerciseRepositoryImpl) SaveExerciseAudit(existingExercise *models.Exercise, operationType string, authUserId string) error {
	auditLog := models.ExerciseAudit{
		ExerciseId:     existingExercise.ExerciseId,
		ExerciseName:   existingExercise.ExerciseName,
		Description:    existingExercise.Description,
		Category:       existingExercise.Category,
		IntensityLevel: existingExercise.IntensityLevel,
		Duration:       existingExercise.Duration,
		DurationUnit:   existingExercise.DurationUnit,
		Benefits:       existingExercise.Benefits,
		IsDeleted:      existingExercise.IsDeleted,
		OperationType:  operationType,
		CreatedAt:      existingExercise.CreatedAt,
		UpdatedAt:      time.Now(),
		CreatedBy:      existingExercise.CreatedBy,
		UpdatedBy:      authUserId,
	}

	return repo.db.Create(&auditLog).Error
}

func (repo *ExerciseRepositoryImpl) UpdateExercise(authUserId string, updatedExercise *models.Exercise) error {
	existingExercise := new(models.Exercise)
	if err := repo.db.Where("exercise_id = ?", updatedExercise.ExerciseId).First(existingExercise).Error; err != nil {
		return err
	}
	updatedExercise.UpdatedAt = time.Now()
	result := repo.db.Model(&models.Exercise{}).Where("exercise_id = ?", updatedExercise.ExerciseId).Updates(updatedExercise)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no records updated")
	}
	if err := repo.SaveExerciseAudit(existingExercise, constant.UPDATE, authUserId); err != nil {
		return err
	}
	return nil
}

func (repo *ExerciseRepositoryImpl) DeleteExercise(exerciseId uint64, authUserId string) error {
	existingExercise, err := repo.GetExerciseById(exerciseId)
	if err != nil {
		return err
	}

	if err := repo.SaveExerciseAudit(existingExercise, constant.DELETE, authUserId); err != nil {
		return err
	}

	if err := repo.db.Model(&models.Exercise{}).
		Where("exercise_id = ?", exerciseId).
		Update("is_deleted", 1).Error; err != nil {
		return err
	}

	return nil
}

func (repo *ExerciseRepositoryImpl) GetExerciseAuditRecord(exerciseId, exerciseAuditId uint64) ([]models.ExerciseAudit, error) {
	var audits []models.ExerciseAudit
	query := repo.db.Model(&models.ExerciseAudit{})

	if exerciseId != 0 {
		query = query.Where("exercise_id = ?", exerciseId)
	}
	if exerciseAuditId != 0 {
		query = query.Where("exercise_audit_id = ?", exerciseAuditId)
	}

	err := query.Order("exercise_audit_id desc").Find(&audits).Error
	if err != nil {
		return nil, err
	}
	return audits, nil
}

func (repo *ExerciseRepositoryImpl) GetAllExerciseAuditRecord(page, limit int) ([]models.ExerciseAudit, int64, error) {
	var audits []models.ExerciseAudit
	var totalRecords int64

	offset := (page - 1) * limit
	query := repo.db.Model(&models.ExerciseAudit{}).Order("exercise_audit_id desc")

	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&audits).Error; err != nil {
		return nil, 0, err
	}

	return audits, totalRecords, nil
}
