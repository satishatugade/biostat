package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type DiagnosticRepository interface {
	// DiagnosticTest Repository
	GetAllDiagnosticTests(limit int, offset int) ([]models.DiagnosticTest, int64, error)
	CreateDiagnosticTest(diagnosticTest *models.DiagnosticTest, createdBy string) (*models.DiagnosticTest, error)
	UpdateDiagnosticTest(diagnosticTest *models.DiagnosticTest, updatedBy string) (*models.DiagnosticTest, error)
	GetSingleDiagnosticTest(diagnosticTestId int) (*models.DiagnosticTest, error)
	DeleteDiagnosticTest(diagnosticTestId int, updatedBy string) error

	// DiagnosticComponent Repository
	GetAllDiagnosticComponents(limit int, offset int) ([]models.DiagnosticTestComponent, int64, error)
	CreateDiagnosticComponent(diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error)
	UpdateDiagnosticComponent(diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error)
	GetSingleDiagnosticComponent(diagnosticComponentId int) (*models.DiagnosticTestComponent, error)

	// DiagnosticTestComponentMapping Repository
	GetAllDiagnosticTestComponentMappings(limit int, offset int) ([]models.DiagnosticTestComponentMapping, int64, error)
	CreateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error)
	UpdateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error)
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

// Diagnostic Test Repository Start
func (r *diagnosticRepositoryImpl) CreateDiagnosticTest(diagnosticTest *models.DiagnosticTest, createdBy string) (*models.DiagnosticTest, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(diagnosticTest).Error; err != nil {
			return err
		}
		auditRecord := models.DiseaseProfileDiagnosticTestMasterAudit{
			DiagnosticTestID: diagnosticTest.DiagnosticTestId,
			TestLoincCode:    diagnosticTest.LoincCode,
			TestName:         diagnosticTest.Name,
			TestDescription:  diagnosticTest.Description,
			Category:         diagnosticTest.Category,
			Units:            diagnosticTest.Units,
			Property:         diagnosticTest.Property,
			TimeAspect:       diagnosticTest.TimeAspect,
			System:           diagnosticTest.System,
			Scale:            diagnosticTest.Scale,
			Method:           "CREATE",
			OperationType:    "CREATE",
			UpdatedBy:        createdBy,
		}
		if err := tx.Create(&auditRecord).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return diagnosticTest, nil
}

func (r *diagnosticRepositoryImpl) UpdateDiagnosticTest(diagnosticTest *models.DiagnosticTest, updatedBy string) (*models.DiagnosticTest, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var diagnosticTestOld models.DiagnosticTest
		if err := tx.First(&diagnosticTestOld, diagnosticTest.DiagnosticTestId).Error; err != nil {
			return err
		}
		auditRecord := models.DiseaseProfileDiagnosticTestMasterAudit{
			DiagnosticTestID: diagnosticTestOld.DiagnosticTestId,
			TestLoincCode:    diagnosticTestOld.LoincCode,
			TestName:         diagnosticTestOld.Name,
			TestDescription:  diagnosticTestOld.Description,
			Category:         diagnosticTestOld.Category,
			Units:            diagnosticTestOld.Units,
			Property:         diagnosticTestOld.Property,
			TimeAspect:       diagnosticTestOld.TimeAspect,
			System:           diagnosticTestOld.System,
			Scale:            diagnosticTestOld.Scale,
			Method:           "UPDATE",
			OperationType:    "UPDATE",
			UpdatedBy:        updatedBy,
		}
		if err := tx.Create(&auditRecord).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.DiagnosticTest{}).Where("diagnostic_test_id=?", diagnosticTest.DiagnosticTestId).Updates(
			map[string]interface{}{
				"test_loinc_code":  diagnosticTest.LoincCode,
				"test_name":        diagnosticTest.Name,
				"test_type":        diagnosticTest.Type,
				"test_description": diagnosticTest.Description,
				"category":         diagnosticTest.Category,
				"units":            diagnosticTest.Units,
				"property":         diagnosticTest.Property,
				"time_aspect":      diagnosticTest.TimeAspect,
				"system":           diagnosticTest.System,
				"scale":            diagnosticTest.Scale,
				"updated_at":       gorm.Expr("NOW()"),
			}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return diagnosticTest, nil
}

func (r *diagnosticRepositoryImpl) GetSingleDiagnosticTest(diagnosticTestId int) (*models.DiagnosticTest, error) {
	var diagnosticTest models.DiagnosticTest
	err := r.db.Preload("Components").Where("diagnostic_test_id=?", diagnosticTestId).First(&diagnosticTest).Error
	if err != nil {
		return nil, err
	}
	return &diagnosticTest, nil
}

func (r *diagnosticRepositoryImpl) DeleteDiagnosticTest(diagnosticTestId int, updatedBy string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var diagnosticTest models.DiagnosticTest
		if err := tx.First(&diagnosticTest, diagnosticTestId).Error; err != nil {
			return err
		}
		auditRecord := models.DiseaseProfileDiagnosticTestMasterAudit{
			DiagnosticTestID: diagnosticTest.DiagnosticTestId,
			TestLoincCode:    diagnosticTest.LoincCode,
			TestName:         diagnosticTest.Name,
			TestDescription:  diagnosticTest.Description,
			Category:         diagnosticTest.Category,
			Units:            diagnosticTest.Units,
			Property:         diagnosticTest.Property,
			TimeAspect:       diagnosticTest.TimeAspect,
			System:           diagnosticTest.System,
			Scale:            diagnosticTest.Scale,
			Method:           "DELETE",
			OperationType:    "DELETE",
			UpdatedBy:        updatedBy,
		}
		if err := tx.Create(&auditRecord).Error; err != nil {
			return err
		}
		if err := tx.Where("diagnostic_test_id = ?", diagnosticTestId).Delete(&models.DiagnosticTestComponentMapping{}).Error; err != nil {
			return err
		}
		if err := tx.Where("diagnostic_test_id = ?", diagnosticTestId).Delete(&models.DiagnosticTest{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// Diagnostic Test Repository End

// Diagnostic Component Repository Start
func (r *diagnosticRepositoryImpl) GetAllDiagnosticComponents(limit int, offset int) ([]models.DiagnosticTestComponent, int64, error) {
	var diagnosticComponents []models.DiagnosticTestComponent
	var totalRecords int64

	// Count total records in the table
	err := r.db.Model(&models.DiagnosticTestComponent{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}

	// Fetch paginated data
	err = r.db.Limit(limit).Offset(offset).Find(&diagnosticComponents).Error
	if err != nil {
		return nil, 0, err
	}

	return diagnosticComponents, totalRecords, nil
}

func (r *diagnosticRepositoryImpl) CreateDiagnosticComponent(diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error) {
	err := r.db.Create(diagnosticComponent).Error
	if err != nil {
		return nil, err
	}
	return diagnosticComponent, nil
}

func (r *diagnosticRepositoryImpl) UpdateDiagnosticComponent(diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error) {
	err := r.db.Model(&models.DiagnosticTestComponent{}).Where("diagnostic_test_component_id=?", diagnosticComponent.DiagnosticTestComponentId).
		Updates(map[string]interface{}{
			"test_component_loinc_code": diagnosticComponent.LoincCode,
			"test_component_name":       diagnosticComponent.Name,
			"description":               diagnosticComponent.Description,
			"test_component_type":       diagnosticComponent.Type,
			"units":                     diagnosticComponent.Units,
			"property":                  diagnosticComponent.Property,
			"time_aspect":               diagnosticComponent.TimeAspect,
			"system":                    diagnosticComponent.System,
			"scale":                     diagnosticComponent.Scale,
			"test_component_frequency":  diagnosticComponent.TestComponentFrequency,
			"updated_at":                gorm.Expr("NOW()"),
		}).Error
	if err != nil {
		return nil, err
	}
	return diagnosticComponent, nil
}

func (r *diagnosticRepositoryImpl) GetSingleDiagnosticComponent(diagnosticComponentId int) (*models.DiagnosticTestComponent, error) {
	var diagnosticComponent models.DiagnosticTestComponent
	err := r.db.Where("diagnostic_test_component_id=?", diagnosticComponentId).First(&diagnosticComponent).Error
	if err != nil {
		return nil, err
	}
	return &diagnosticComponent, nil
}

// Diagnostic Component Repository End

// Diagnostic Test Component Mapping Repository Start

func (r *diagnosticRepositoryImpl) GetAllDiagnosticTestComponentMappings(limit int, offset int) ([]models.DiagnosticTestComponentMapping, int64, error) {
	var diagnosticTestComponentMappings []models.DiagnosticTestComponentMapping
	var totalRecords int64

	// Count total records in the table
	err := r.db.Model(&models.DiagnosticTestComponentMapping{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}

	// Fetch paginated data
	err = r.db.Limit(limit).Offset(offset).Find(&diagnosticTestComponentMappings).Error
	if err != nil {
		return nil, 0, err
	}

	return diagnosticTestComponentMappings, totalRecords, nil
}

func (r *diagnosticRepositoryImpl) CreateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error) {
	err := r.db.Create(diagnosticTestComponentMapping).Error
	if err != nil {
		return nil, err
	}
	return diagnosticTestComponentMapping, nil
}

func (r *diagnosticRepositoryImpl) UpdateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error) {
	err := r.db.Model(&models.DiagnosticTestComponentMapping{}).Where("diagnostic_test_component_mapping_id=?", diagnosticTestComponentMapping.DiagnosticTestComponentMappingId).
		Updates(map[string]interface{}{
			"diagnostic_test_id":           diagnosticTestComponentMapping.DiagnosticTestId,
			"diagnostic_test_component_id": diagnosticTestComponentMapping.DiagnosticComponentId,
		}).Error
	if err != nil {
		return nil, err
	}
	return diagnosticTestComponentMapping, nil
}

// Dignostic Test Compoent Mapping Repository End
