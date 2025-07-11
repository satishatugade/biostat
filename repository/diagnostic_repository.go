package repository

import (
	"biostat/constant"
	"biostat/models"
	"errors"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

type DiagnosticRepository interface {

	//Diagnostic labs
	CreateLab(tx *gorm.DB, lab *models.DiagnosticLab) (*models.DiagnosticLab, error)
	GetAllLabs(limit, offset int) ([]models.DiagnosticLab, int64, error)
	GetSinglePatientDiagnosticLab(patientId uint64, labId *uint64) (*models.DiagnosticLabResponse, error)
	GetPatientDiagnosticLabs(patientid uint64, limit int, offset int) ([]models.DiagnosticLabResponse, int64, error)
	GetLabById(diagnosticlLabId uint64) (*models.DiagnosticLab, error)
	UpdateLab(diagnosticlLab *models.DiagnosticLab, authUserId string) error
	DeleteLab(diagnosticlLabId uint64, authUserId string) error
	AddMapping(userId uint64, LabInfo *models.DiagnosticLab) error
	GetAllDiagnosticLabAuditRecords(limit, offset int) ([]models.DiagnosticLabAudit, int64, error)
	GetDiagnosticLabAuditRecord(labId, labAuditId uint64) ([]models.DiagnosticLabAudit, error)

	// DiagnosticTest Repository
	GetAllDiagnosticTests(limit int, offset int) ([]models.DiagnosticTest, int64, error)
	CreateDiagnosticTest(tx *gorm.DB, diagnosticTest *models.DiagnosticTest, createdBy string) (*models.DiagnosticTest, error)
	UpdateDiagnosticTest(diagnosticTest *models.DiagnosticTest, updatedBy string) (*models.DiagnosticTest, error)
	GetSingleDiagnosticTest(diagnosticTestId int) (*models.DiagnosticTest, error)
	DeleteDiagnosticTest(diagnosticTestId int, updatedBy string) error

	// DiagnosticComponent Repository
	GetAllDiagnosticComponents(limit int, offset int) ([]models.DiagnosticTestComponent, int64, error)
	CreateDiagnosticComponent(tx *gorm.DB, diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error)
	UpdateDiagnosticComponent(authUserId string, diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error)
	GetSingleDiagnosticComponent(diagnosticComponentId int) (*models.DiagnosticTestComponent, error)
	DeleteDiagnosticTestComponent(diagnosticTestComponetId uint64, updatedBy string) error

	// DiagnosticTestComponentMapping Repository
	GetAllDiagnosticTestComponentMappings(limit int, offset int) ([]models.DiagnosticTestComponentMapping, int64, error)
	CreateDiagnosticTestComponentMapping(tx *gorm.DB, diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error)
	UpdateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error)
	DeleteDiagnosticTestComponentMapping(diagnosticTestId uint64, diagnosticComponentId uint64) error
	AddDiseaseDiagnosticTestMapping(mapping *models.DiseaseDiagnosticTestMapping) error

	AddTestReferenceRange(input *models.DiagnosticTestReferenceRange) error
	UpdateTestReferenceRange(input *models.DiagnosticTestReferenceRange, updatedBy string) error
	DeleteTestReferenceRange(testReferenceRangeId uint64, deletedBy string) error
	GetAllTestRefRangeView(limit int, offset int, isDeleted uint64) ([]models.Diagnostic_Test_Component_ReferenceRange, int64, error)
	ViewTestReferenceRange(testReferenceRangeId uint64) (*models.DiagnosticTestReferenceRange, error)
	GetTestReferenceRangeAuditRecord(testReferenceRangeId, auditId uint64, limit, offset int) ([]models.Diagnostic_Test_Component_ReferenceRange, int64, error)
	LoadDiagnosticTestMasterData() (map[string]uint64, map[string]uint64)
	LoadDiagnosticLabData() map[string]uint64
	GeneratePatientDiagnosticReport(tx *gorm.DB, patientDiagnoReport *models.PatientDiagnosticReport) (*models.PatientDiagnosticReport, error)
	SavePatientDiagnosticTestInterpretation(tx *gorm.DB, patientDiagnoTest *models.PatientDiagnosticTest) (*models.PatientDiagnosticTest, error)
	SavePatientReportResultValue(tx *gorm.DB, resultValues *models.PatientDiagnosticTestResultValue) (*models.PatientDiagnosticTestResultValue, error)
	SavePatientReportAttachmentMapping(tx *gorm.DB, recordMapping *models.PatientReportAttachment) error
	GetAbnormalValue(patientId uint64) ([]models.TestResultAlert, error)
	ArchivePatientDiagnosticReport(reportID uint64, isDeleted int) error
	AddMappingToMergeTestComponent(mapping []models.DiagnosticTestComponentAliasMapping) error
	FetchSources(limit, offset int) ([]models.HealthVitalSourceType, int64, error)
	GetSourceById(sourceId int) (models.HealthVitalSource, error)
	GetDiagnosticLabReportName(patientId uint64) ([]models.DiagnosticReport, error)
	GetPatientLabNameAndEmail(userId uint64) ([]models.DiagnosticLabResponse, error)
}

type DiagnosticRepositoryImpl struct {
	db *gorm.DB
}

func NewDiagnosticRepository(db *gorm.DB) DiagnosticRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &DiagnosticRepositoryImpl{db: db}
}

// GetSourceById implements DiagnosticRepository.
func (r *DiagnosticRepositoryImpl) GetSourceById(sourceId int) (models.HealthVitalSource, error) {
	panic("unimplemented")
}

func (r *DiagnosticRepositoryImpl) LoadDiagnosticLabData() map[string]uint64 {
	labMap := make(map[string]uint64)

	var labs []models.DiagnosticLab
	if err := r.db.Select("diagnostic_lab_id", "lab_name").Find(&labs).Error; err != nil {
		log.Println("Error loading diagnostic lab data:", err)
		return nil
	}
	for _, lab := range labs {
		labMap[strings.ToLower(lab.LabName)] = lab.DiagnosticLabId
	}
	return labMap
}

func (r *DiagnosticRepositoryImpl) GetAllDiagnosticTests(limit int, offset int) ([]models.DiagnosticTest, int64, error) {

	var diagnosticTests []models.DiagnosticTest
	var totalRecords int64
	err := r.db.Model(&models.DiagnosticTest{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Preload("Components").Order("diagnostic_test_id DESC").Limit(limit).Offset(offset).Find(&diagnosticTests).Error
	if err != nil {
		return nil, 0, err
	}
	return diagnosticTests, totalRecords, nil
}

func (r *DiagnosticRepositoryImpl) CreateDiagnosticTest(tx *gorm.DB, diagnosticTest *models.DiagnosticTest, createdBy string) (*models.DiagnosticTest, error) {
	if err := tx.Create(diagnosticTest).Error; err != nil {
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

func (r *DiagnosticRepositoryImpl) UpdateDiagnosticTest(diagnosticTest *models.DiagnosticTest, updatedBy string) (*models.DiagnosticTest, error) {
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

func (r *DiagnosticRepositoryImpl) GetSingleDiagnosticTest(diagnosticTestId int) (*models.DiagnosticTest, error) {
	var diagnosticTest models.DiagnosticTest
	err := r.db.Preload("Components").Where("diagnostic_test_id=?", diagnosticTestId).First(&diagnosticTest).Error
	if err != nil {
		return nil, err
	}
	return &diagnosticTest, nil
}

func (r *DiagnosticRepositoryImpl) DeleteDiagnosticTest(diagnosticTestId int, updatedBy string) error {
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

func (r *DiagnosticRepositoryImpl) GetAllDiagnosticComponents(limit int, offset int) ([]models.DiagnosticTestComponent, int64, error) {
	var diagnosticComponents []models.DiagnosticTestComponent
	var totalRecords int64
	err := r.db.Model(&models.DiagnosticTestComponent{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}
	err = r.db.Order("diagnostic_test_component_id DESC").Limit(limit).Offset(offset).Find(&diagnosticComponents).Error
	if err != nil {
		return nil, 0, err
	}
	return diagnosticComponents, totalRecords, nil
}

func (r *DiagnosticRepositoryImpl) CreateDiagnosticComponent(tx *gorm.DB, diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error) {
	if err := tx.Create(diagnosticComponent).Error; err != nil {
		return nil, err
	}
	return diagnosticComponent, nil
}

func SaveDiagnosticTestComponentAudit(tx *gorm.DB, component *models.DiagnosticTestComponent, operationType string, updatedBy string) error {
	audit := models.DiagnosticTestComponentAudit{
		DiagnosticTestComponentId: component.DiagnosticTestComponentId,
		TestComponentLoincCode:    component.LoincCode,
		TestComponentName:         component.TestComponentName,
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

func (r *DiagnosticRepositoryImpl) UpdateDiagnosticComponent(authUserId string, diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var existingComponent models.DiagnosticTestComponent
		if err := tx.First(&existingComponent, "diagnostic_test_component_id = ?", diagnosticComponent.DiagnosticTestComponentId).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.DiagnosticTestComponent{}).
			Where("diagnostic_test_component_id = ?", diagnosticComponent.DiagnosticTestComponentId).
			Updates(map[string]interface{}{
				"test_component_loinc_code": diagnosticComponent.LoincCode,
				"test_component_name":       diagnosticComponent.TestComponentName,
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

func (r *DiagnosticRepositoryImpl) DeleteDiagnosticTestComponent(diagnosticTestComponentId uint64, updatedBy string) error {
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

func (r *DiagnosticRepositoryImpl) GetSingleDiagnosticComponent(diagnosticComponentId int) (*models.DiagnosticTestComponent, error) {
	var diagnosticComponent models.DiagnosticTestComponent
	err := r.db.Where("diagnostic_test_component_id=?", diagnosticComponentId).First(&diagnosticComponent).Error
	if err != nil {
		return nil, err
	}
	return &diagnosticComponent, nil
}

func (r *DiagnosticRepositoryImpl) GetAllDiagnosticTestComponentMappings(limit int, offset int) ([]models.DiagnosticTestComponentMapping, int64, error) {
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

func (r *DiagnosticRepositoryImpl) CreateDiagnosticTestComponentMapping(tx *gorm.DB, diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error) {
	if err := tx.Create(diagnosticTestComponentMapping).Error; err != nil {
		return nil, err
	}
	return diagnosticTestComponentMapping, nil
}

func (r *DiagnosticRepositoryImpl) UpdateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error) {
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

func (r *DiagnosticRepositoryImpl) DeleteDiagnosticTestComponentMapping(diagnosticTestId uint64, diagnosticComponentId uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var mapping models.DiagnosticTestComponentMapping

		if err := tx.Where("diagnostic_test_id = ? AND diagnostic_test_component_id = ?", diagnosticTestId, diagnosticComponentId).
			First(&mapping).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return gorm.ErrRecordNotFound
			}
			return err
		}
		if err := tx.Delete(&mapping).Error; err != nil {
			return err
		}

		return nil
	})
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

func (r *DiagnosticRepositoryImpl) CreateLab(tx *gorm.DB, lab *models.DiagnosticLab) (*models.DiagnosticLab, error) {
	if err := tx.Create(&lab).Error; err != nil {
		return nil, err
	}
	return lab, nil
}

func (r *DiagnosticRepositoryImpl) GetAllLabs(limit, offset int) ([]models.DiagnosticLab, int64, error) {
	var labs []models.DiagnosticLab
	var total int64
	query := r.db.Model(&models.DiagnosticLab{})

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Order("diagnostic_lab_id desc").Find(&labs).Error
	return labs, total, err
}

func (dr *DiagnosticRepositoryImpl) GetPatientLabNameAndEmail(userId uint64) ([]models.DiagnosticLabResponse, error) {
	var labs []models.DiagnosticLabResponse

	err := dr.db.
		Table("tbl_diagnostic_lab AS dl").
		Select("dl.lab_name, dl.lab_email").
		Joins("JOIN tbl_patient_diagnostic_lab_mapping AS dlm ON dl.diagnostic_lab_id = dlm.diagnostic_lab_id").
		Where("dlm.patient_id = ?", userId).
		Scan(&labs).Error

	if err != nil {
		return nil, err
	}

	if len(labs) == 0 {
		return nil, errors.New("no labs found please add labs first")
	}

	return labs, nil
}

func (dr *DiagnosticRepositoryImpl) GetPatientDiagnosticLabs(patientId uint64, limit int, offset int) ([]models.DiagnosticLabResponse, int64, error) {
	var labs []models.DiagnosticLabResponse
	var count int64

	if err := dr.db.Model(&models.PatientDiagnosticLabMapping{}).
		Where("patient_id = ? AND is_deleted = 0", patientId).
		Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := dr.db.Table("tbl_patient_diagnostic_lab_mapping AS mapping").
		Select(`labs.diagnostic_lab_id, labs.lab_no, labs.lab_name, labs.lab_address, labs.city, labs.state,
			labs.postal_code, labs.lab_contact_number, labs.lab_email, labs.is_deleted,
			labs.created_at, labs.updated_at, labs.created_by, labs.updated_by`).
		Joins("JOIN tbl_diagnostic_lab AS labs ON labs.diagnostic_lab_id = mapping.diagnostic_lab_id").
		Where("mapping.patient_id = ? AND mapping.is_deleted = 0", patientId).
		Order("labs.diagnostic_lab_id DESC").
		Limit(limit).
		Offset(offset).
		Scan(&labs).Error

	return labs, count, err
}

func (dr *DiagnosticRepositoryImpl) GetSinglePatientDiagnosticLab(patientId uint64, diagnosticLabId *uint64) (*models.DiagnosticLabResponse, error) {
	var lab models.DiagnosticLabResponse

	query := dr.db.Table("tbl_patient_diagnostic_lab_mapping AS mapping").
		Select(`labs.diagnostic_lab_id, labs.lab_no, labs.lab_name, labs.lab_address, labs.city, labs.state,
			labs.postal_code, labs.lab_contact_number, labs.lab_email, labs.is_deleted,
			labs.created_at, labs.updated_at, labs.created_by, labs.updated_by`).
		Joins("JOIN tbl_diagnostic_lab AS labs ON labs.diagnostic_lab_id = mapping.diagnostic_lab_id").
		Where("mapping.patient_id = ? AND mapping.is_deleted = 0", patientId)

	if diagnosticLabId != nil {
		query = query.Where("mapping.diagnostic_lab_id = ?", *diagnosticLabId)
	} else {
		query = query.Order("mapping.created_at DESC").Limit(1)
	}

	err := query.Scan(&lab).Error
	if err != nil {
		return nil, err
	}

	return &lab, nil
}

func (r *DiagnosticRepositoryImpl) GetLabById(diagnosticlLabId uint64) (*models.DiagnosticLab, error) {
	var lab models.DiagnosticLab
	err := r.db.Where("diagnostic_lab_id = ? AND is_deleted = 0", diagnosticlLabId).First(&lab).Error
	return &lab, err
}

func (r *DiagnosticRepositoryImpl) UpdateLab(lab *models.DiagnosticLab, deletedBy string) error {
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

func (r *DiagnosticRepositoryImpl) DeleteLab(id uint64, deletedBy string) error {
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

func (repo *DiagnosticRepositoryImpl) AddMapping(patientId uint64, labInfo *models.DiagnosticLab) error {
	newLabMapping := models.PatientDiagnosticLabMapping{
		PatientId:       patientId,
		DiagnosticLabId: labInfo.DiagnosticLabId,
		CreatedBy:       labInfo.CreatedBy,
		UpdatedBy:       labInfo.CreatedBy,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		IsDeleted:       0,
	}
	if err := repo.db.Create(&newLabMapping).Error; err != nil {
		return err
	}
	return nil
}

func (r *DiagnosticRepositoryImpl) GetAllDiagnosticLabAuditRecords(limit, offset int) ([]models.DiagnosticLabAudit, int64, error) {
	var records []models.DiagnosticLabAudit
	var total int64
	query := r.db.Model(&models.DiagnosticLabAudit{}).Order("diagnostic_lab_audit_id desc")

	err := query.Count(&total).Limit(limit).Offset(offset).Find(&records).Error
	return records, total, err
}

func (r *DiagnosticRepositoryImpl) GetDiagnosticLabAuditRecord(labId, labAuditId uint64) ([]models.DiagnosticLabAudit, error) {
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

func (r *DiagnosticRepositoryImpl) AddDiseaseDiagnosticTestMapping(mapping *models.DiseaseDiagnosticTestMapping) error {
	return r.db.Create(mapping).Error
}

func (r *DiagnosticRepositoryImpl) AddTestReferenceRange(input *models.DiagnosticTestReferenceRange) error {
	return r.db.Create(&input).Error
}

func (r *DiagnosticRepositoryImpl) SaveTestReferenceRangeAudit(tx *gorm.DB, oldRecord *models.DiagnosticTestReferenceRange, updatedBy string, operationType string) error {
	auditEntry := models.DiagnosticTestReferenceRangeAudit{
		TestReferenceRangeId:      oldRecord.TestReferenceRangeId,
		DiagnosticTestId:          oldRecord.DiagnosticTestId,
		DiagnosticTestComponentId: oldRecord.DiagnosticTestComponentId,
		Age:                       oldRecord.Age,
		AgeGroup:                  oldRecord.AgeGroup,
		Gender:                    oldRecord.Gender,
		NormalMin:                 oldRecord.NormalMin,
		NormalMax:                 oldRecord.NormalMax,
		Units:                     oldRecord.Units,
		OperationType:             operationType,
		CreatedAt:                 time.Now(),
		UpdatedBy:                 updatedBy,
	}

	return tx.Create(&auditEntry).Error
}

func (r *DiagnosticRepositoryImpl) UpdateTestReferenceRange(input *models.DiagnosticTestReferenceRange, updatedBy string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var oldRecord models.DiagnosticTestReferenceRange
		if err := tx.Where("test_reference_range_id = ?", input.TestReferenceRangeId).First(&oldRecord).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.DiagnosticTestReferenceRange{}).
			Where("test_reference_range_id = ?", input.TestReferenceRangeId).
			Updates(input).Error; err != nil {
			return err
		}
		if err := r.SaveTestReferenceRangeAudit(tx, &oldRecord, updatedBy, constant.UPDATE); err != nil {
			return err
		}
		return nil
	})
}

func (r *DiagnosticRepositoryImpl) DeleteTestReferenceRange(testReferenceRangeId uint64, deletedBy string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var oldRecord models.DiagnosticTestReferenceRange
		if err := tx.Where("test_reference_range_id = ?", testReferenceRangeId).First(&oldRecord).Error; err != nil {
			return err
		}
		oldRecord.IsDeleted = 1
		if err := r.SaveTestReferenceRangeAudit(tx, &oldRecord, deletedBy, constant.DELETE); err != nil {
			return err
		}
		if err := tx.Model(&models.DiagnosticTestReferenceRange{}).
			Where("test_reference_range_id = ?", testReferenceRangeId).
			Updates(map[string]interface{}{
				"is_deleted": 1,
				"updated_at": time.Now(),
			}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *DiagnosticRepositoryImpl) ViewTestReferenceRange(testReferenceRangeId uint64) (*models.DiagnosticTestReferenceRange, error) {
	var ranges *models.DiagnosticTestReferenceRange
	err := r.db.Where("diagnostic_test_id = ? ", testReferenceRangeId).Find(&ranges).Error
	return ranges, err
}

func (r *DiagnosticRepositoryImpl) GetAllTestRefRangeView(limit int, offset int, isDeleted uint64) ([]models.Diagnostic_Test_Component_ReferenceRange, int64, error) {
	var results []models.Diagnostic_Test_Component_ReferenceRange
	var total int64

	query := r.db.Model(&models.DiagnosticTestReferenceRange{}).
		Joins("JOIN tbl_disease_profile_diagnostic_test_master AS dt ON dt.diagnostic_test_id = tbl_diagnostic_test_reference_range.diagnostic_test_id").
		Joins("JOIN tbl_disease_profile_diagnostic_test_component_master AS dtc ON dtc.diagnostic_test_component_id = tbl_diagnostic_test_reference_range.diagnostic_test_component_id").
		Select("tbl_diagnostic_test_reference_range.*, dt.test_name, dtc.test_component_name")

	if isDeleted >= 0 && isDeleted <= 1 {
		query = query.Where("tbl_diagnostic_test_reference_range.is_deleted = ?", isDeleted)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("tbl_diagnostic_test_reference_range.test_reference_range_id DESC").
		Limit(limit).
		Offset(offset).
		Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *DiagnosticRepositoryImpl) GetTestReferenceRangeAuditRecord(testReferenceRangeId, auditId uint64, limit, offset int) ([]models.Diagnostic_Test_Component_ReferenceRange, int64, error) {
	var audits []models.Diagnostic_Test_Component_ReferenceRange
	var totalRecords int64

	query := r.db.Table("tbl_diagnostic_test_reference_range_audit AS dtra").
		Joins("JOIN tbl_disease_profile_diagnostic_test_master AS dt ON dt.diagnostic_test_id = dtra.diagnostic_test_id").
		Joins("JOIN tbl_disease_profile_diagnostic_test_component_master AS dtc ON dtc.diagnostic_test_component_id = dtra.diagnostic_test_component_id").
		Select("dtra.*, dt.test_name, dtc.test_component_name")

	if testReferenceRangeId != 0 {
		query = query.Where("dtra.test_reference_range_id = ?", testReferenceRangeId)
	}

	if auditId != 0 {
		query = query.Where("dtra.test_reference_range_audit_id = ?", auditId)
	}

	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Order("dtra.test_reference_range_audit_id DESC").Find(&audits).Error; err != nil {
		return nil, 0, err
	}

	return audits, totalRecords, nil
}

func (s *DiagnosticRepositoryImpl) LoadDiagnosticTestMasterData() (map[string]uint64, map[string]uint64) {
	testNameCache := make(map[string]uint64)
	componentNameCache := make(map[string]uint64)
	duplicateComponents := make([]string, 0)

	var tests []models.DiagnosticTest
	if err := s.db.Find(&tests).Error; err != nil {
		log.Printf("Error loading test master data: %v", err)
		return nil, nil
	}
	for _, test := range tests {
		testNameCache[strings.ToLower(strings.TrimSpace(test.TestName))] = test.DiagnosticTestId
	}
	subQuery := s.db.
		Table("tbl_diagnostic_test_component_alias_mapping").
		Select("alias_test_component_id").
		Where("diagnostic_test_component_id != alias_test_component_id").
		Where("is_deleted = 0")

	var components []models.DiagnosticTestComponent
	if err := s.db.
		Where("diagnostic_test_component_id NOT IN (?)", subQuery).
		Find(&components).Error; err != nil {
		log.Printf("Error loading component master data: %v", err)
		return nil, nil
	}

	for _, component := range components {
		key := strings.ToLower(strings.TrimSpace(component.TestComponentName))
		if _, exists := componentNameCache[key]; exists {
			duplicateComponents = append(duplicateComponents, component.TestComponentName)
			continue
		}
		componentNameCache[key] = component.DiagnosticTestComponentId
	}
	log.Printf("Duplicate Test Component found: %s", duplicateComponents)
	return testNameCache, componentNameCache
}

func (r *DiagnosticRepositoryImpl) GeneratePatientDiagnosticReport(tx *gorm.DB, report *models.PatientDiagnosticReport) (*models.PatientDiagnosticReport, error) {
	if err := tx.Create(report).Error; err != nil {
		return nil, err
	}
	return report, nil
}

func (r *DiagnosticRepositoryImpl) SavePatientDiagnosticTestInterpretation(tx *gorm.DB, interpretation *models.PatientDiagnosticTest) (*models.PatientDiagnosticTest, error) {
	if err := tx.Create(interpretation).Error; err != nil {
		return nil, err
	}
	return interpretation, nil
}

func (r *DiagnosticRepositoryImpl) SavePatientReportResultValue(tx *gorm.DB, resultValues *models.PatientDiagnosticTestResultValue) (*models.PatientDiagnosticTestResultValue, error) {
	if err := tx.Create(resultValues).Error; err != nil {
		return nil, err
	}
	return resultValues, nil
}

func (r *DiagnosticRepositoryImpl) SavePatientReportAttachmentMapping(tx *gorm.DB, recordMapping *models.PatientReportAttachment) error {
	if err := tx.Create(recordMapping).Error; err != nil {
		return err
	}
	return nil
}

func (ds *DiagnosticRepositoryImpl) GetAbnormalValue(patientId uint64) ([]models.TestResultAlert, error) {

	var alerts []models.TestResultAlert
	err := ds.db.Raw(`
		SELECT
			pdtrv.result_value,
			dtrr.normal_min,
			dtrr.normal_max,
			pdtrv.result_status,
			dpdtm.test_name,
			dpdtcm.test_component_name,
			format_datetime(pdtrv.result_date) AS result_date,
			CASE
				WHEN pdtrv.result_value < dtrr.normal_min THEN 'Below Range'
				WHEN pdtrv.result_value > dtrr.normal_max THEN 'Above Range'
				ELSE 'Within Range'
			END AS result_comment
		FROM tbl_patient_diagnostic_test_result_value pdtrv
		JOIN tbl_diagnostic_test_reference_range dtrr
			ON pdtrv.diagnostic_test_id = dtrr.diagnostic_test_id
			AND pdtrv.diagnostic_test_component_id = dtrr.diagnostic_test_component_id
		JOIN tbl_disease_profile_diagnostic_test_master dpdtm
			ON pdtrv.diagnostic_test_id = dpdtm.diagnostic_test_id
		JOIN tbl_disease_profile_diagnostic_test_component_master dpdtcm
			ON pdtrv.diagnostic_test_component_id = dpdtcm.diagnostic_test_component_id
		WHERE (
			pdtrv.result_value < dtrr.normal_min OR
			pdtrv.result_value > dtrr.normal_max
		)
		AND pdtrv.result_value IS NOT NULL
		AND pdtrv.patient_id = ?
		AND pdtrv.patient_diagnostic_report_id = (
			SELECT patient_diagnostic_report_id
			FROM tbl_patient_diagnostic_report
			WHERE patient_id = ?
			ORDER BY patient_diagnostic_report_id DESC
		LIMIT 1 );
	`, patientId, patientId).Scan(&alerts).Error

	if err != nil {
		return nil, err
	}

	return alerts, nil
}

func (dr *DiagnosticRepositoryImpl) ArchivePatientDiagnosticReport(recordId uint64, isDeleted int) error {
	tx := dr.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	err := tx.Model(&models.TblMedicalRecord{}).
		Where("record_id = ?", recordId).
		Update("is_deleted", isDeleted).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	var reportId uint64
	err = tx.Model(&models.PatientReportAttachment{}).
		Select("patient_diagnostic_report_id").
		Where("record_id = ?", recordId).
		Scan(&reportId).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Model(&models.PatientDiagnosticReport{}).
		Where("patient_diagnostic_report_id = ?", reportId).
		Update("is_deleted", isDeleted).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *DiagnosticRepositoryImpl) AddMappingToMergeTestComponent(mapping []models.DiagnosticTestComponentAliasMapping) error {
	return r.db.Create(mapping).Error
}

func (r *DiagnosticRepositoryImpl) FetchSources(limit, offset int) ([]models.HealthVitalSourceType, int64, error) {
	var results []models.HealthVitalSourceType
	var total int64

	err := r.db.Model(&models.HealthVitalSourceType{}).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.
		Preload("Sources", "is_deleted = ?", 0).
		Order("source_type_id").
		Limit(limit).Offset(offset).
		Find(&results).Error

	return results, total, err
}

func (r *DiagnosticRepositoryImpl) GetDiagnosticLabReportName(patientId uint64) ([]models.DiagnosticReport, error) {
	var reportNameInfo []models.DiagnosticReport
	if err := r.db.Model(&models.PatientDiagnosticReport{}).Select(`patient_diagnostic_report_id, report_name || ' - ' || format_datetime(report_date) AS report_name `).
		Where("patient_id = ? AND is_deleted = ?", patientId, 0).Scan(&reportNameInfo).Error; err != nil {
		log.Println("Error loading diagnostic lab report name:", err)
		return []models.DiagnosticReport{}, nil
	}
	return reportNameInfo, nil
}
