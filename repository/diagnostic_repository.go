package repository

import (
	"biostat/constant"
	"biostat/models"
	"time"

	"gorm.io/gorm"
)

type DiagnosticRepository interface {

	//Diagnostic labs
	CreateLab(lab *models.DiagnosticLab) error
	GetLabById(diagnosticlLabId uint64) (*models.DiagnosticLab, error)
	UpdateLab(diagnosticlLab *models.DiagnosticLab, authUserId string) error
	DeleteLab(diagnosticlLabId uint64, authUserId string) error
	GetAllDiagnosticLabAuditRecords(page, limit int) ([]models.DiagnosticLabAudit, int64, error)
	GetDiagnosticLabAuditRecord(labId, labAuditId uint64) ([]models.DiagnosticLabAudit, error)

	// DiagnosticTest Repository
	GetAllDiagnosticTests(limit int, offset int) ([]models.DiagnosticTest, int64, error)
	CreateDiagnosticTest(diagnosticTest *models.DiagnosticTest, createdBy string) (*models.DiagnosticTest, error)
	UpdateDiagnosticTest(diagnosticTest *models.DiagnosticTest, updatedBy string) (*models.DiagnosticTest, error)
	GetSingleDiagnosticTest(diagnosticTestId int) (*models.DiagnosticTest, error)
	DeleteDiagnosticTest(diagnosticTestId int, updatedBy string) error

	// DiagnosticComponent Repository
	GetAllDiagnosticComponents(limit int, offset int) ([]models.DiagnosticTestComponent, int64, error)
	CreateDiagnosticComponent(diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error)
	UpdateDiagnosticComponent(authUserId string, diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error)
	GetSingleDiagnosticComponent(diagnosticComponentId int) (*models.DiagnosticTestComponent, error)
	DeleteDiagnosticTestComponent(diagnosticTestComponetId uint64, updatedBy string) error

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
		auditRecord := models.DiagnosticTestAudit{
			DiagnosticTestId: diagnosticTest.DiagnosticTestId,
			TestLoincCode:    diagnosticTest.LoincCode,
			TestName:         diagnosticTest.TestName,
			TestDescription:  diagnosticTest.Description,
			Category:         diagnosticTest.Category,
			Units:            diagnosticTest.Units,
			Property:         diagnosticTest.Property,
			TimeAspect:       diagnosticTest.TimeAspect,
			System:           diagnosticTest.System,
			Scale:            diagnosticTest.Scale,
			OperationType:    constant.CREATE,
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

func SaveDiagnosticTestAudit(tx *gorm.DB, test *models.DiagnosticTest, operation string, updatedBy string) error {
	audit := models.DiagnosticTestAudit{
		DiagnosticTestId: test.DiagnosticTestId,
		TestLoincCode:    test.LoincCode,
		TestName:         test.TestName,
		TestDescription:  test.Description,
		Category:         test.Category,
		Units:            test.Units,
		Property:         test.Property,
		TimeAspect:       test.TimeAspect,
		System:           test.System,
		Scale:            test.Scale,
		OperationType:    operation,
		UpdatedBy:        updatedBy,
	}
	return tx.Create(&audit).Error
}

func (r *diagnosticRepositoryImpl) UpdateDiagnosticTest(diagnosticTest *models.DiagnosticTest, updatedBy string) (*models.DiagnosticTest, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var diagnosticTestOld models.DiagnosticTest
		if err := tx.First(&diagnosticTestOld, diagnosticTest.DiagnosticTestId).Error; err != nil {
			return err
		}
		if err := SaveDiagnosticTestAudit(tx, &diagnosticTestOld, constant.UPDATE, updatedBy); err != nil {
			return err
		}
		if err := tx.Model(&models.DiagnosticTest{}).Where("diagnostic_test_id=?", diagnosticTest.DiagnosticTestId).Updates(
			map[string]interface{}{
				"test_loinc_code":  diagnosticTest.LoincCode,
				"test_name":        diagnosticTest.TestName,
				"test_type":        diagnosticTest.TestType,
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
		if err := SaveDiagnosticTestAudit(tx, &diagnosticTest, constant.UPDATE, updatedBy); err != nil {
			return err
		}
		// if err := tx.Where("diagnostic_test_id = ?", diagnosticTestId).Delete(&models.DiagnosticTestComponentMapping{}).Error; err != nil {
		// 	return err
		// }
		if err := tx.Model(&models.DiagnosticTest{}).
			Where("diagnostic_test_id = ?", diagnosticTestId).
			Updates(map[string]interface{}{
				"is_deleted": 1,
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *diagnosticRepositoryImpl) GetAllDiagnosticComponents(limit int, offset int) ([]models.DiagnosticTestComponent, int64, error) {
	var diagnosticComponents []models.DiagnosticTestComponent
	var totalRecords int64
	err := r.db.Model(&models.DiagnosticTestComponent{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}
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

func SaveDiagnosticTestComponentAudit(tx *gorm.DB, component *models.DiagnosticTestComponent, operationType string, updatedBy string) error {
	audit := models.DiagnosticTestComponentAudit{
		DiagnosticTestComponentId: component.DiagnosticTestComponentId,
		TestComponentLoincCode:    component.LoincCode,
		TestComponentName:         component.TestComponetName,
		TestComponentType:         component.TestComponentType,
		Description:               component.Description,
		Units:                     component.Units,
		Property:                  component.Property,
		TimeAspect:                component.TimeAspect,
		System:                    component.System,
		Scale:                     component.Scale,
		TestComponentFrequency:    component.TestComponentFrequency,
		OperationType:             operationType,
		CreatedBy:                 component.CreatedBy,
		UpdatedBy:                 updatedBy,
		IsDeleted:                 0,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}

	return tx.Create(&audit).Error
}

func (r *diagnosticRepositoryImpl) UpdateDiagnosticComponent(authUserId string, diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var existingComponent models.DiagnosticTestComponent
		if err := tx.First(&existingComponent, "diagnostic_test_component_id = ?", diagnosticComponent.DiagnosticTestComponentId).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.DiagnosticTestComponent{}).
			Where("diagnostic_test_component_id = ?", diagnosticComponent.DiagnosticTestComponentId).
			Updates(map[string]interface{}{
				"test_component_loinc_code": diagnosticComponent.LoincCode,
				"test_component_name":       diagnosticComponent.TestComponetName,
				"description":               diagnosticComponent.Description,
				"test_component_type":       diagnosticComponent.TestComponentType,
				"units":                     diagnosticComponent.Units,
				"property":                  diagnosticComponent.Property,
				"time_aspect":               diagnosticComponent.TimeAspect,
				"system":                    diagnosticComponent.System,
				"scale":                     diagnosticComponent.Scale,
				"test_component_frequency":  diagnosticComponent.TestComponentFrequency,
				"updated_at":                gorm.Expr("NOW()"),
			}).Error; err != nil {
			return err
		}

		if err := SaveDiagnosticTestComponentAudit(tx, &existingComponent, constant.UPDATE, authUserId); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return diagnosticComponent, nil
}

func (r *diagnosticRepositoryImpl) DeleteDiagnosticTestComponent(diagnosticTestComponentId uint64, updatedBy string) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var existingComponent models.DiagnosticTestComponent

		if err := tx.First(&existingComponent, diagnosticTestComponentId).Error; err != nil {
			return err
		}

		// Save to audit before marking as deleted
		if err := SaveDiagnosticTestComponentAudit(tx, &existingComponent, constant.DELETE, updatedBy); err != nil {
			return err
		}

		// Soft delete: Set is_deleted = 1
		if err := tx.Model(&models.DiagnosticTestComponent{}).
			Where("diagnostic_test_component_id = ?", diagnosticTestComponentId).
			Update("is_deleted", 1).Error; err != nil {
			return err
		}

		return nil
	})
	return err
}

func (r *diagnosticRepositoryImpl) GetSingleDiagnosticComponent(diagnosticComponentId int) (*models.DiagnosticTestComponent, error) {
	var diagnosticComponent models.DiagnosticTestComponent
	err := r.db.Where("diagnostic_test_component_id=?", diagnosticComponentId).First(&diagnosticComponent).Error
	if err != nil {
		return nil, err
	}
	return &diagnosticComponent, nil
}

func (r *diagnosticRepositoryImpl) GetAllDiagnosticTestComponentMappings(limit int, offset int) ([]models.DiagnosticTestComponentMapping, int64, error) {
	var diagnosticTestComponentMappings []models.DiagnosticTestComponentMapping
	var totalRecords int64

	err := r.db.Model(&models.DiagnosticTestComponentMapping{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}
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

func SaveDiagnosticLabAudit(tx *gorm.DB, lab *models.DiagnosticLab, actionType string, updatedBy string) error {
	audit := models.DiagnosticLabAudit{
		DiagnosticLabId:  lab.DiagnosticLabId,
		LabNo:            lab.LabNo,
		LabName:          lab.LabName,
		LabAddress:       lab.LabAddress,
		LabContactNumber: lab.LabContactNumber,
		LabEmail:         lab.LabEmail,
		OperationType:    actionType,
		CreatedBy:        updatedBy,
		UpdatedAt:        time.Now(),
	}
	return tx.Create(&audit).Error
}

func (r *diagnosticRepositoryImpl) CreateLab(lab *models.DiagnosticLab) error {
	return r.db.Create(lab).Error
}

func (r *diagnosticRepositoryImpl) GetLabById(diagnosticlLabId uint64) (*models.DiagnosticLab, error) {
	var lab models.DiagnosticLab
	err := r.db.Where("diagnostic_lab_id = ? AND is_deleted = 0", diagnosticlLabId).First(&lab).Error
	return &lab, err
}

func (r *diagnosticRepositoryImpl) UpdateLab(lab *models.DiagnosticLab, deletedBy string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing models.DiagnosticLab
		if err := tx.First(&existing, lab.DiagnosticLabId).Error; err != nil {
			return err
		}

		if err := SaveDiagnosticLabAudit(tx, &existing, constant.UPDATE, deletedBy); err != nil {
			return err
		}

		lab.UpdatedAt = time.Now()

		return tx.Model(&models.DiagnosticLab{}).
			Where("diagnostic_lab_id = ?", lab.DiagnosticLabId).
			Updates(map[string]interface{}{
				"lab_no":             lab.LabNo,
				"lab_name":           lab.LabName,
				"lab_address":        lab.LabAddress,
				"lab_contact_number": lab.LabContactNumber,
				"lab_email":          lab.LabEmail,
				"updated_at":         lab.UpdatedAt,
				"updated_by":         lab.UpdatedBy,
			}).Error
	})
}

func (r *diagnosticRepositoryImpl) DeleteLab(id uint64, deletedBy string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var lab models.DiagnosticLab
		if err := tx.First(&lab, id).Error; err != nil {
			return err
		}

		if err := SaveDiagnosticLabAudit(tx, &lab, constant.DELETE, deletedBy); err != nil {
			return err
		}

		lab.IsDeleted = 1
		lab.UpdatedBy = deletedBy
		lab.UpdatedAt = time.Now()

		return tx.Model(&models.DiagnosticLab{}).
			Where("diagnostic_lab_id = ?", id).
			Updates(map[string]interface{}{
				"is_deleted": lab.IsDeleted,
				"updated_by": lab.UpdatedBy,
				"updated_at": lab.UpdatedAt,
			}).Error
	})
}

func (r *diagnosticRepositoryImpl) GetAllDiagnosticLabAuditRecords(page, limit int) ([]models.DiagnosticLabAudit, int64, error) {
	var records []models.DiagnosticLabAudit
	var total int64

	offset := (page - 1) * limit
	query := r.db.Model(&models.DiagnosticLabAudit{}).Order("diagnostic_lab_audit_id desc")

	err := query.Count(&total).Limit(limit).Offset(offset).Find(&records).Error
	return records, total, err
}

func (r *diagnosticRepositoryImpl) GetDiagnosticLabAuditRecord(labId, labAuditId uint64) ([]models.DiagnosticLabAudit, error) {
	var records []models.DiagnosticLabAudit
	query := r.db.Model(&models.DiagnosticLabAudit{})

	if labId != 0 {
		query = query.Where("diagnostic_lab_id = ?", labId)
	}
	if labAuditId != 0 {
		query = query.Where("diagnostic_lab_audit_id = ?", labAuditId)
	}

	err := query.Order("diagnostic_lab_audit_id desc").Find(&records).Error
	return records, err
}
