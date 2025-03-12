package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type DiagnosticRepository interface {
	GetAllDiagnosticTests(limit int, offset int) ([]models.DiagnosticTest, int64, error)
}

type diagnosticRepositoryImpl struct {
	db *gorm.DB
}

func NewDiagnosticRepository(db *gorm.DB) DiagnosticRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &diagnosticRepositoryImpl{db: db}
}

func (r *diagnosticRepositoryImpl) GetAllDiagnosticTests(limit int, offset int) ([]models.DiagnosticTest, int64, error) {

	var diagnosticTests []models.DiagnosticTest
	var totalRecords int64

	// Count total records in the table
	err := r.db.Model(&models.DiagnosticTest{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}

	// Fetch paginated data
	err = r.db.Preload("Components").Limit(limit).Offset(offset).Find(&diagnosticTests).Error
	if err != nil {
		return nil, 0, err
	}

	return diagnosticTests, totalRecords, nil
}
