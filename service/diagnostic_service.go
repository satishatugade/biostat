package service

import (
	"biostat/config"
	"biostat/constant"
	"biostat/database"
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DiagnosticService interface {
	CreateLab(userId uint64, authUserId string, lab models.AddLabRequest) (*models.DiagnosticLab, error)
	GetAllLabs(limit, offset int) ([]models.DiagnosticLabResponse, int64, error)
	GetPatientDiagnosticLabs(patientid uint64, limit int, offset int) ([]models.DiagnosticLabResponse, int64, error)
	GetSinglePatientDiagnosticLab(patientId uint64, labId *uint64) (*models.DiagnosticLabResponse, error)
	GetLabById(diagnosticlLabId uint64) (*models.DiagnosticLab, error)
	UpdateLab(diagnosticlLab *models.DiagnosticLab, authUserId string) error
	DeleteLab(diagnosticlLabId uint64, authUserId string) error
	DeleteLabByUser(lab_id, user_id uint64, deletedBy string) error
	GetAllDiagnosticLabAuditRecords(limit, offset int) ([]models.DiagnosticLabAudit, int64, error)
	AddMapping(userId uint64, LabInfo *models.DiagnosticLab) error
	GetDiagnosticLabAuditRecord(labId, labAuditId uint64) ([]models.DiagnosticLabAudit, error)
	GeneratePatientDiagnosticReport(tx *gorm.DB, patientDiagnoReport *models.PatientDiagnosticReport) (*models.PatientDiagnosticReport, error)
	SavePatientReportAttachmentMapping(recordMapping *models.PatientReportAttachment) error
	SaveUserTag(userID uint64, tagName string, recordID, reportID *uint64) ([]*models.UserTag, error)
	GetDiagnosticTests(limit int, offset int) ([]models.DiagnosticTest, int64, error)
	CreateDiagnosticTest(diagnosticTest *models.DiagnosticTest, createdBy string) (*models.DiagnosticTest, error)
	UpdateDiagnosticTest(diagnosticTest *models.DiagnosticTest, updatedBy string) (*models.DiagnosticTest, error)
	GetSingleDiagnosticTest(diagnosticTestId int) (*models.DiagnosticTest, error)
	DeleteDiagnosticTest(diagnosticTestId int, updatedBy string) error

	GetAllDiagnosticComponents(limit int, offset int) ([]models.DiagnosticTestComponent, int64, error)
	CreateDiagnosticComponent(diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error)
	UpdateDiagnosticComponent(authUserId string, diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error)
	DeleteDiagnosticTestComponent(diagnosticTestComponetId uint64, updatedBy string) error
	GetSingleDiagnosticComponent(diagnosticComponentId int) (*models.DiagnosticTestComponent, error)

	GetAllDiagnosticTestComponentMappings(limit int, offset int) ([]models.DiagnosticTestComponentMapping, int64, error)
	CreateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error)
	UpdateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error)
	DeleteDiagnosticTestComponentMapping(diagnosticTestId uint64, diagnosticComponentId uint64) error

	AddDiseaseDiagnosticTestMapping(mapping *models.DiseaseDiagnosticTestMapping) error

	//reference range
	AddTestReferenceRange(input *models.DiagnosticTestReferenceRange) error
	UpdateTestReferenceRange(input *models.DiagnosticTestReferenceRange, updatedBy string) error
	DeleteTestReferenceRange(testReferenceRangeId uint64, deletedBy string) error
	GetAllTestRefRangeView(limit int, offset int, isDeleted uint64) ([]models.Diagnostic_Test_Component_ReferenceRange, int64, error)
	ViewTestReferenceRange(testReferenceRangeId uint64) (*models.DiagnosticTestReferenceRange, error)
	GetTestReferenceRangeAuditRecord(testReferenceRangeId, auditId uint64, limit, offset int) ([]models.Diagnostic_Test_Component_ReferenceRange, int64, error)
	DigitizeDiagnosticReport(reportData models.LabReport, patientId uint64, recordId *uint64, reportId *uint64) (string, error)
	CheckReportExistWithSampleDateTestComponent(reportData models.LabReport, patientId uint64, recordId *uint64, processId uuid.UUID, attachmentId *string, reportId *uint64) error
	AddMappingToMergeTestComponent(mapping []models.DiagnosticTestComponentAliasMapping) error
	NotifyAbnormalResult(patientId uint64) error
	ArchivePatientDiagnosticReport(reportID uint64, isDeleted int) error
	GetSources(patientId uint64, limit, offset int) ([]models.HealthVitalSourceType, int64, error)
	GetDiagnosticLabReportName(patientId uint64) ([]models.DiagnosticReport, error)
}

type DiagnosticServiceImpl struct {
	diagnosticRepo       repository.DiagnosticRepository
	emailService         EmailService
	patientService       PatientService
	medicalRecordsRepo   repository.TblMedicalRecordRepository
	processStatusService ProcessStatusService
}

// SavePatientReportAttachmentMapping implements DiagnosticService.
func (s *DiagnosticServiceImpl) SavePatientReportAttachmentMapping(recordMapping *models.PatientReportAttachment) error {
	return s.diagnosticRepo.SavePatientReportAttachmentMapping(recordMapping)
}

func (s *DiagnosticServiceImpl) SaveUserTag(userID uint64, tagName string, recordID, reportID *uint64) ([]*models.UserTag, error) {
	if recordID == nil && reportID == nil {
		return nil, fmt.Errorf("either recordID or reportID must be provided")
	}

	var tags []string
	err := json.Unmarshal([]byte(tagName), &tags)
	if err != nil {
		return nil, fmt.Errorf("invalid tags format: %v", err)
	}
	var savedTags []*models.UserTag

	for _, t := range tags {
		trimmedTag := strings.TrimSpace(t)
		if trimmedTag == "" {
			continue
		}

		tag := &models.UserTag{
			UserId:                    userID,
			TagName:                   trimmedTag,
			RecordId:                  recordID,
			PatientDiagnosticReportId: reportID,
		}

		if err := s.diagnosticRepo.CreateUserTag(tag); err != nil {
			log.Println("Error saving user tag:", err)
			return nil, err
		}

		savedTags = append(savedTags, tag)
	}

	return savedTags, nil
}

func (s *DiagnosticServiceImpl) GeneratePatientDiagnosticReport(tx *gorm.DB, patientDiagnoReport *models.PatientDiagnosticReport) (*models.PatientDiagnosticReport, error) {
	return s.diagnosticRepo.GeneratePatientDiagnosticReport(tx, patientDiagnoReport)
}

func NewDiagnosticService(repo repository.DiagnosticRepository, emailService EmailService, patientService PatientService,
	medicalRecordsRepo repository.TblMedicalRecordRepository, processStatusService ProcessStatusService) DiagnosticService {
	return &DiagnosticServiceImpl{diagnosticRepo: repo, emailService: emailService, patientService: patientService,
		medicalRecordsRepo: medicalRecordsRepo, processStatusService: processStatusService}
}

func (s *DiagnosticServiceImpl) GetSources(patientId uint64, limit, offset int) ([]models.HealthVitalSourceType, int64, error) {
	data, totalRecord, err := s.diagnosticRepo.FetchSources(limit, offset)
	if err != nil {
		return []models.HealthVitalSourceType{}, 0, err
	}
	lab, _, _ := s.diagnosticRepo.GetPatientDiagnosticLabs(patientId, limit, offset)
	return AddSourceName(data, lab, totalRecord)
}

func AddSourceName(data []models.HealthVitalSourceType, labs []models.DiagnosticLabResponse, totalRecord int64) ([]models.HealthVitalSourceType, int64, error) {
	for i, sourceType := range data {
		if sourceType.SourceTypeId == 3 {
			newSources := []models.HealthVitalSource{}
			for _, lab := range labs {
				newSources = append(newSources, models.HealthVitalSource{
					SourceId:   lab.DiagnosticLabId,
					SourceName: lab.LabName,
				})
			}
			data[i].Sources = newSources
		}
	}
	return data, totalRecord, nil
}

func (s *DiagnosticServiceImpl) GetDiagnosticTests(limit int, offset int) ([]models.DiagnosticTest, int64, error) {
	return s.diagnosticRepo.GetAllDiagnosticTests(limit, offset)
}

func (s *DiagnosticServiceImpl) AddMapping(patientId uint64, labInfo *models.DiagnosticLab) error {
	tx := database.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	if err := s.diagnosticRepo.AddMapping(tx, patientId, labInfo); err != nil {
		tx.Rollback()
		return fmt.Errorf("error while mapping diagnostic lab: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit AddMapping transaction: %w", err)
	}

	return nil
}

func (ds *DiagnosticServiceImpl) ArchivePatientDiagnosticReport(reportID uint64, isDeleted int) error {
	return ds.diagnosticRepo.ArchivePatientDiagnosticReport(reportID, isDeleted)
}

func (s *DiagnosticServiceImpl) CreateDiagnosticTest(diagnosticTest *models.DiagnosticTest, createdBy string) (*models.DiagnosticTest, error) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	createdTest, err := s.diagnosticRepo.CreateDiagnosticTest(tx, diagnosticTest, createdBy)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return createdTest, nil
}

func (s *DiagnosticServiceImpl) UpdateDiagnosticTest(diagnosticTest *models.DiagnosticTest, updatedBy string) (*models.DiagnosticTest, error) {
	return s.diagnosticRepo.UpdateDiagnosticTest(diagnosticTest, updatedBy)
}

func (s *DiagnosticServiceImpl) GetSingleDiagnosticTest(diagnosticTestId int) (*models.DiagnosticTest, error) {
	return s.diagnosticRepo.GetSingleDiagnosticTest(diagnosticTestId)
}

func (s *DiagnosticServiceImpl) DeleteDiagnosticTest(diagnosticTestId int, updatedBy string) error {
	return s.diagnosticRepo.DeleteDiagnosticTest(diagnosticTestId, updatedBy)
}

func (s *DiagnosticServiceImpl) GetAllDiagnosticComponents(limit int, offset int) ([]models.DiagnosticTestComponent, int64, error) {
	return s.diagnosticRepo.GetAllDiagnosticComponents(limit, offset)
}

func (s *DiagnosticServiceImpl) CreateDiagnosticComponent(diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	createdComponent, err := s.diagnosticRepo.CreateDiagnosticComponent(tx, diagnosticComponent)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return createdComponent, nil
}

func (s *DiagnosticServiceImpl) UpdateDiagnosticComponent(authUserId string, diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error) {
	return s.diagnosticRepo.UpdateDiagnosticComponent(authUserId, diagnosticComponent)
}

func (s *DiagnosticServiceImpl) DeleteDiagnosticTestComponent(diagnosticTestComponetId uint64, authUserId string) error {
	return s.diagnosticRepo.DeleteDiagnosticTestComponent(diagnosticTestComponetId, authUserId)
}

func (s *DiagnosticServiceImpl) GetSingleDiagnosticComponent(diagnosticComponentId int) (*models.DiagnosticTestComponent, error) {
	return s.diagnosticRepo.GetSingleDiagnosticComponent(diagnosticComponentId)
}

func (s *DiagnosticServiceImpl) GetAllDiagnosticTestComponentMappings(limit int, offset int) ([]models.DiagnosticTestComponentMapping, int64, error) {
	return s.diagnosticRepo.GetAllDiagnosticTestComponentMappings(limit, offset)
}

func (s *DiagnosticServiceImpl) CreateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	createdMapping, err := s.diagnosticRepo.CreateDiagnosticTestComponentMapping(tx, diagnosticTestComponentMapping)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return createdMapping, nil
}

func (s *DiagnosticServiceImpl) UpdateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error) {
	return s.diagnosticRepo.UpdateDiagnosticTestComponentMapping(diagnosticTestComponentMapping)
}

func (s *DiagnosticServiceImpl) DeleteDiagnosticTestComponentMapping(diagnosticTestId uint64, diagnosticComponentId uint64) error {
	return s.diagnosticRepo.DeleteDiagnosticTestComponentMapping(diagnosticTestId, diagnosticComponentId)
}

func (s *DiagnosticServiceImpl) CreateLab(userId uint64, authUserId string, req models.AddLabRequest) (*models.DiagnosticLab, error) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if req.IsSystemLab {
		for _, labId := range req.LabID {
			existingLab, err := s.diagnosticRepo.GetLabById(labId)
			if err != nil || existingLab == nil {
				tx.Rollback()
				return nil, errors.New("Invalid Lab ID")
			}
			lab := models.DiagnosticLab{
				LabName:   existingLab.LabName,
				CreatedBy: authUserId,
			}

			createdLab, err := s.diagnosticRepo.CreateLab(tx, &lab)
			if err != nil {
				tx.Rollback()
				return nil, err
			}

			if err := s.diagnosticRepo.AddMapping(tx, userId, createdLab); err != nil {
				tx.Rollback()
				return nil, err
			}

			for _, relative_id := range req.RelativeIds {
				err = s.diagnosticRepo.AddMapping(tx, relative_id, createdLab)
				if err != nil {
					tx.Rollback()
					return nil, err
				}
			}
		}

		tx.Commit()
		return nil, nil
	}
	lab := models.DiagnosticLab{
		LabName:          req.LabName,
		LabNo:            req.LabNo,
		LabAddress:       req.LabAddress,
		City:             req.City,
		State:            req.State,
		PostalCode:       req.PostalCode,
		LabContactNumber: req.LabContactNumber,
		LabEmail:         req.LabEmail,
		CreatedBy:        authUserId,
	}
	createdLab, err := s.diagnosticRepo.CreateLab(tx, &lab)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	err = s.diagnosticRepo.AddMapping(tx, userId, createdLab)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, relative_id := range req.RelativeIds {
		err = s.diagnosticRepo.AddMapping(tx, relative_id, createdLab)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	return createdLab, nil
}

func (s *DiagnosticServiceImpl) GetLabById(id uint64) (*models.DiagnosticLab, error) {
	return s.diagnosticRepo.GetLabById(id)
}

func (s *DiagnosticServiceImpl) UpdateLab(diagnosticlLabId *models.DiagnosticLab, authUserId string) error {
	return s.diagnosticRepo.UpdateLab(diagnosticlLabId, authUserId)
}

func (s *DiagnosticServiceImpl) DeleteLab(diagnosticlLabId uint64, authUserId string) error {
	return s.diagnosticRepo.DeleteLab(diagnosticlLabId, authUserId)
}

func (s *DiagnosticServiceImpl) DeleteLabByUser(lab_id, user_id uint64, deletedBy string) error {
	return s.diagnosticRepo.DeleteLabByUser(lab_id, user_id, deletedBy)
}
func (s *DiagnosticServiceImpl) GetAllDiagnosticLabAuditRecords(limit, offset int) ([]models.DiagnosticLabAudit, int64, error) {
	return s.diagnosticRepo.GetAllDiagnosticLabAuditRecords(limit, offset)
}

func (s *DiagnosticServiceImpl) GetDiagnosticLabAuditRecord(labId, labAuditId uint64) ([]models.DiagnosticLabAudit, error) {
	return s.diagnosticRepo.GetDiagnosticLabAuditRecord(labId, labAuditId)
}

func (s *DiagnosticServiceImpl) GetAllLabs(limit, offset int) ([]models.DiagnosticLabResponse, int64, error) {
	var res []models.DiagnosticLabResponse
	labs, totalRecords, err := s.diagnosticRepo.GetAllLabs(limit, offset)
	if err != nil {
		return res, totalRecords, err
	}
	for _, lab := range labs {
		res = append(res, toDiagnosticLabResponse(lab))
	}
	return res, totalRecords, nil
}

func toDiagnosticLabResponse(lab models.DiagnosticLab) models.DiagnosticLabResponse {
	return models.DiagnosticLabResponse{
		DiagnosticLabId:  lab.DiagnosticLabId,
		LabNo:            lab.LabNo,
		LabName:          lab.LabName,
		LabAddress:       lab.LabAddress,
		City:             lab.City,
		State:            lab.State,
		PostalCode:       lab.PostalCode,
		LabContactNumber: lab.LabContactNumber,
		LabEmail:         lab.LabEmail,
		IsDeleted:        lab.IsDeleted,
		CreatedAt:        lab.CreatedAt,
		UpdatedAt:        lab.UpdatedAt,
		CreatedBy:        lab.CreatedBy,
		UpdatedBy:        lab.UpdatedBy,
	}
}

func (s *DiagnosticServiceImpl) GetPatientDiagnosticLabs(patientId uint64, limit, offset int) ([]models.DiagnosticLabResponse, int64, error) {
	return s.diagnosticRepo.GetPatientDiagnosticLabs(patientId, limit, offset)
}

func (s *DiagnosticServiceImpl) GetSinglePatientDiagnosticLab(patientId uint64, labId *uint64) (*models.DiagnosticLabResponse, error) {
	return s.diagnosticRepo.GetSinglePatientDiagnosticLab(patientId, labId)
}

func (s *DiagnosticServiceImpl) AddDiseaseDiagnosticTestMapping(mapping *models.DiseaseDiagnosticTestMapping) error {
	return s.diagnosticRepo.AddDiseaseDiagnosticTestMapping(mapping)
}

func (s *DiagnosticServiceImpl) AddTestReferenceRange(input *models.DiagnosticTestReferenceRange) error {
	return s.diagnosticRepo.AddTestReferenceRange(input)
}

func (s *DiagnosticServiceImpl) UpdateTestReferenceRange(input *models.DiagnosticTestReferenceRange, updatedBy string) error {
	return s.diagnosticRepo.UpdateTestReferenceRange(input, updatedBy)
}

func (s *DiagnosticServiceImpl) DeleteTestReferenceRange(testReferenceRangeId uint64, deletedBy string) error {
	return s.diagnosticRepo.DeleteTestReferenceRange(testReferenceRangeId, deletedBy)
}

func (s *DiagnosticServiceImpl) ViewTestReferenceRange(testReferenceRangeId uint64) (*models.DiagnosticTestReferenceRange, error) {
	return s.diagnosticRepo.ViewTestReferenceRange(testReferenceRangeId)
}

func (s *DiagnosticServiceImpl) GetAllTestRefRangeView(limit int, offset int, isDeleted uint64) ([]models.Diagnostic_Test_Component_ReferenceRange, int64, error) {
	return s.diagnosticRepo.GetAllTestRefRangeView(limit, offset, isDeleted)
}

func (s *DiagnosticServiceImpl) GetTestReferenceRangeAuditRecord(testReferenceRangeId, auditId uint64, limit, offset int) ([]models.Diagnostic_Test_Component_ReferenceRange, int64, error) {
	return s.diagnosticRepo.GetTestReferenceRangeAuditRecord(testReferenceRangeId, auditId, limit, offset)
}

func (s *DiagnosticServiceImpl) DigitizeDiagnosticReport(reportData models.LabReport, patientId uint64, recordId *uint64, reportId *uint64) (string, error) {
	testNameCache, componentNameCache := s.diagnosticRepo.LoadDiagnosticTestMasterData()
	if testNameCache == nil || componentNameCache == nil {
		log.Println("Failed to load diagnostic test master data")
		return "", errors.New("test and component master data not available")
	}

	diagnosticLabs := s.diagnosticRepo.LoadDiagnosticLabData()
	if diagnosticLabs == nil {
		log.Println("Failed to load lab data")
		return "", errors.New("diagnostic lab data not available")
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered in DigitizeDiagnosticReport:", r)
			log.Println("Stack trace:\n" + string(debug.Stack()))
			tx.Rollback()
		}
	}()

	var diagnosticLabId uint64
	if reportData.ReportDetails.SourceId != nil {
		diagnosticLabId = *reportData.ReportDetails.SourceId
	} else {
		labName := reportData.ReportDetails.LabName
		if labName == "" {
			log.Println("Lab name is empty, lab creating unknown lab name.")
			labName = "UnknownLab"
		}
		if val, exists := diagnosticLabs[strings.ToLower(labName)]; exists {
			diagnosticLabId = val
			reportData.ReportDetails.IsLabReport = true
		} else {
			newLab := models.DiagnosticLab{
				LabNo:            reportData.ReportDetails.LabId,
				LabName:          labName,
				LabAddress:       reportData.ReportDetails.LabLocation,
				LabContactNumber: reportData.ReportDetails.LabContactNumber,
				LabEmail:         reportData.ReportDetails.LabEmail,
			}
			labInfo, err := s.diagnosticRepo.CreateLab(tx, &newLab)
			if err != nil {
				log.Println("ERROR saving DiagnosticLab info:", err)
				tx.Rollback()
				return "", fmt.Errorf("error while saving diagnostic lab info: %w", err)
			}
			if err := s.diagnosticRepo.AddMapping(tx, patientId, labInfo); err != nil {
				log.Println("ERROR while AddMapping for diagnostic lab:", err)
			}
			diagnosticLabId = labInfo.DiagnosticLabId
			reportData.ReportDetails.IsLabReport = true
		}
	}
	reportDate, err := utils.ParseDate(reportData.ReportDetails.ReportDate)
	if err != nil {
		log.Println("ReportDate parsing failed:", err)
		tx.Rollback()
		return "", err
	}
	collectionDate := reportDate
	if reportData.ReportDetails.CollectionDate != "" {
		collectionDate, err = utils.ParseDate(reportData.ReportDetails.CollectionDate)
		if err != nil {
			log.Println("collectionDate parsing failed set report date as colection date :", err)
			collectionDate = reportDate
		}
	}
	var reportInfo *models.PatientDiagnosticReport
	var reportErr error
	if reportId == nil {
		reporIdNew := uint64(time.Now().UnixNano() + int64(rand.Intn(1000)))
		log.Println("PatientDiagnosticReport Id 19 digit : ", reporIdNew)
		patientReport := models.PatientDiagnosticReport{
			PatientDiagnosticReportId: reporIdNew,
			DiagnosticLabId:           diagnosticLabId,
			PatientId:                 patientId,
			PaymentStatus:             constant.Success,
			ReportName:                reportData.ReportDetails.ReportName,
			CollectedDate:             collectionDate,
			ReportDate:                reportDate,
			Observation:               "",
			IsDigital:                 reportData.ReportDetails.IsDigital,
			IsDeleted:                 reportData.ReportDetails.IsDeleted,
			IsLabReport:               reportData.ReportDetails.IsLabReport,
			IsHealthVital:             reportData.ReportDetails.IsHealthVital,
			CollectedAt:               reportData.ReportDetails.LabLocation,
		}
		reportInfo, reportErr = s.diagnosticRepo.GeneratePatientDiagnosticReport(tx, &patientReport)
		if reportErr != nil {
			log.Println("ERROR saving PatientDiagnosticReport:", reportErr)
			tx.Rollback()
			return "", fmt.Errorf("error while saving patient diagnostic report: %w", reportErr)
		}
		recordmapping := models.PatientReportAttachment{
			PatientDiagnosticReportId: reportInfo.PatientDiagnosticReportId,
			RecordId:                  *recordId,
			PatientId:                 patientId,
		}
		if err := s.diagnosticRepo.SavePatientReportAttachmentMapping(&recordmapping); err != nil {
			log.Println("Error while creating SavePatientReportAttachmentMapping:", err)
			tx.Rollback()
			return "", fmt.Errorf("error while SavePatientReportAttachmentMapping: %w", err)
		}
	} else {

		updateData := map[string]interface{}{
			"diagnostic_lab_id": diagnosticLabId,
			"patient_id":        patientId,
			"payment_status":    constant.Success,
			"report_name":       reportData.ReportDetails.ReportName,
			"collected_date":    collectionDate,
			"report_date":       reportDate,
			"observation":       "",
			"is_digital":        reportData.ReportDetails.IsDigital,
			"is_deleted":        reportData.ReportDetails.IsDeleted,
			"is_lab_report":     reportData.ReportDetails.IsLabReport,
			"is_health_vital":   reportData.ReportDetails.IsHealthVital,
			"collected_at":      reportData.ReportDetails.LabLocation,
		}

		reportInfo, reportErr = s.diagnosticRepo.UpdatePatientDiagnosticReport(tx, *reportId, updateData)
		if reportErr != nil {
			log.Println("ERROR updating PatientDiagnosticReport:", reportErr)
			tx.Rollback()
			return "", fmt.Errorf("error while updating patient diagnostic report: %w", reportErr)
		}

	}
	for _, testData := range reportData.Tests {
		testName := testData.TestName
		testInterpretation := testData.Interpretation

		var diagnosticTestId uint64
		if id, exists := testNameCache[strings.ToLower(strings.TrimSpace(testName))]; exists {
			diagnosticTestId = id
		} else {
			newTest := models.DiagnosticTest{TestName: testName}
			testInfo, err := s.diagnosticRepo.CreateDiagnosticTest(tx, &newTest, "System")
			if err != nil {
				log.Println("Error while creating DiagnosticTest:", err)
				tx.Rollback()
				return "", fmt.Errorf("error while creating diagnostic test: %w", err) // Wrap error
			}
			diagnosticTestId = testInfo.DiagnosticTestId
			testNameCache[strings.ToLower(strings.TrimSpace(testInfo.TestName))] = diagnosticTestId
		}

		//Save patient test
		testRecord := models.PatientDiagnosticTest{
			PatientDiagnosticReportId: reportInfo.PatientDiagnosticReportId,
			DiagnosticTestId:          diagnosticTestId,
			TestNote:                  testInterpretation,
			TestDate:                  reportDate,
		}
		_, err := s.diagnosticRepo.SavePatientDiagnosticTestInterpretation(tx, &testRecord)
		if err != nil {
			log.Println("ERROR saving test interpretation:", err)
			tx.Rollback()
			return "", fmt.Errorf("error while saving test interpretation: %w", err)
		}

		// Save all component values
		for _, component := range testData.Components {
			var diagnosticComponentId uint64
			parsedResultValue, _ := strconv.ParseFloat(component.ResultValue, 64)
			resultStatus := component.Status
			var Qualifier string
			if component.Qualifier != nil {
				Qualifier = *component.Qualifier
			}
			if parsedResultValue == 0 {
				resultStatus = component.ResultValue
			}
			if id, exists := componentNameCache[strings.ToLower(strings.TrimSpace(component.TestComponentName))]; exists {
				diagnosticComponentId = id
				if err := s.SaveDiagnosticResultValue(tx, diagnosticTestId, diagnosticComponentId, resultStatus, parsedResultValue, reportDate, patientId, reportInfo.PatientDiagnosticReportId, Qualifier); err != nil {
					tx.Rollback()
					return "", err
				}
			} else {
				newComponent := models.DiagnosticTestComponent{
					TestComponentName:      component.TestComponentName,
					Units:                  component.Units,
					TestComponentFrequency: "0",
				}
				componentInfo, err := s.diagnosticRepo.CreateDiagnosticComponent(tx, &newComponent)
				if err != nil {
					log.Println("Error while creating Diagnostic Component:", err)
					tx.Rollback()
					return "", fmt.Errorf("error while creating diagnostic test component: %w", err)
				}
				diagnosticComponentId = componentInfo.DiagnosticTestComponentId
				componentNameCache[strings.ToLower(strings.TrimSpace(componentInfo.TestComponentName))] = diagnosticComponentId

				mapping := models.DiagnosticTestComponentMapping{
					DiagnosticTestId:      diagnosticTestId,
					DiagnosticComponentId: diagnosticComponentId,
				}
				if _, err := s.diagnosticRepo.CreateDiagnosticTestComponentMapping(tx, &mapping); err != nil {
					log.Println("Error while creating DiagnosticTestComponentMapping:", err)
					tx.Rollback()
					return "", fmt.Errorf("error while creating diagnostic test component mapping: %w", err) // Wrap error
				}
				normalMin := func() float64 { v, _ := strconv.ParseFloat(component.ReferenceRange.Min, 64); return v }()
				normalMax := func() float64 { v, _ := strconv.ParseFloat(component.ReferenceRange.Max, 64); return v }()
				referenceMap, err := s.diagnosticRepo.GetAllDiagnosticReferenceRange()
				if err != nil {
					log.Println("Error while GetAllDiagnosticReferenceRange :", err)
				}
				lookupKey := fmt.Sprintf("%d-%d", diagnosticTestId, diagnosticComponentId)
				existingRef, exists := referenceMap[lookupKey]
				if !exists {
					referenceRange := models.DiagnosticTestReferenceRange{
						DiagnosticTestId:               diagnosticTestId,
						DiagnosticTestComponentId:      diagnosticComponentId,
						NormalMin:                      normalMin,
						NormalMax:                      normalMax,
						BiologicalReferenceDescription: component.BiologicalReferenceDescription,
						Units:                          component.Units,
					}
					refRangeErr := s.diagnosticRepo.AddTestReferenceRange(&referenceRange)
					if refRangeErr != nil {
						log.Println("ERROR saving test Ref. range:", refRangeErr)
						tx.Rollback()
						return "", fmt.Errorf("error while saving test reference range: %w", refRangeErr)
					}
					if err := s.SaveDiagnosticResultValue(tx, diagnosticTestId, diagnosticComponentId, resultStatus, parsedResultValue, reportDate, patientId, reportInfo.PatientDiagnosticReportId, Qualifier); err != nil {
						tx.Rollback()
						return "", err
					}
				} else {
					if existingRef.NormalMin != normalMin || existingRef.NormalMax != normalMax {
						log.Println("Mismatch detected → inserting updated value against patient Id")
						patientRefRange := models.PatientTestReferenceRange{
							PatientId:                 patientId,
							DiagnosticTestID:          diagnosticTestId,
							DiagnosticTestComponentId: diagnosticComponentId,
							NormalMin:                 normalMin,
							NormalMax:                 normalMax,
							BiologicalReferenceDesc:   *component.BiologicalReferenceDescription,
							Units:                     component.Units,
						}

						if err := s.diagnosticRepo.AddPatientTestReferenceRange(&patientRefRange); err != nil {
							log.Println("ERROR saving Patient-specific Ref. range:", err)
							tx.Rollback()
							return "", fmt.Errorf("error while saving patient-specific reference range: %w", err)
						}
						if err := s.SaveDiagnosticResultValue(tx, diagnosticTestId, diagnosticComponentId, resultStatus, parsedResultValue, reportDate, patientId, reportInfo.PatientDiagnosticReportId, Qualifier); err != nil {
							tx.Rollback()
							return "", err
						}
					} else {
						log.Println("Master reference matches → no patient-specific range needed")
					}
				}
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		log.Printf("ERROR committing transaction: err : %v", err)
		return "", err
	}
	return "Diagnostic report created!", nil
}

func (s *DiagnosticServiceImpl) SaveDiagnosticResultValue(
	tx *gorm.DB,
	diagnosticTestId, diagnosticComponentId uint64,
	resultStatus string,
	parsedResultValue float64,
	reportDate time.Time,
	patientId uint64,
	reportId uint64,
	Qualifier string,
) error {
	result := models.PatientDiagnosticTestResultValue{
		DiagnosticTestId:          diagnosticTestId,
		DiagnosticTestComponentId: diagnosticComponentId,
		ResultStatus:              resultStatus,
		ResultValue:               parsedResultValue,
		ResultDate:                reportDate,
		PatientId:                 patientId,
		UDF1:                      Qualifier,
		PatientDiagnosticReportId: reportId,
	}

	_, err := s.diagnosticRepo.SavePatientReportResultValue(tx, &result)
	if err != nil {
		log.Println("ERROR saving test result:", err)
		return fmt.Errorf("error while saving test result: %w", err)
	}
	return nil
}

func GetResultStatus(resultVal, minStr, maxStr, status string) string {
	if resultVal == "" {
		return "-"
	}
	result, _ := strconv.ParseFloat(resultVal, 64)
	min, _ := strconv.ParseFloat(minStr, 64)
	max, _ := strconv.ParseFloat(maxStr, 64)

	// if err1 != nil || err2 != nil || err3 != nil {
	// 	return "-"
	// }

	if result == 0 {
		fmt.Println("result string : ", result)
		fmt.Println("status val : ", status)
		fmt.Println("result val : ", resultVal)
		return resultVal
	}
	switch {
	case result < min:
		return "Low"
	case result > max:
		return "High"
	default:
		return "Normal"
	}
}

func (ds *DiagnosticServiceImpl) NotifyAbnormalResult(patientId uint64) error {
	patient, err := ds.patientService.GetUserProfileByUserId(patientId)
	if err != nil {
		log.Printf("GetUserProfileByUserId failed: %v", err)
	}
	alerts, err := ds.diagnosticRepo.GetAbnormalValue(patientId)
	if err != nil {
		log.Printf("failed to get abnormal values: %v", err)
	}
	if len(alerts) == 0 {
		return nil
	}
	err = ds.emailService.SendReportResultsEmail(patient, alerts)
	if err != nil {
		log.Printf("failed to send alert email: %v", err)
	}
	log.Println("SendReportResultsEmail success")
	return nil
}

func (s *DiagnosticServiceImpl) AddMappingToMergeTestComponent(mapping []models.DiagnosticTestComponentAliasMapping) error {
	return s.diagnosticRepo.AddMappingToMergeTestComponent(mapping)
}

func (s *DiagnosticServiceImpl) GetDiagnosticLabReportName(patientId uint64) ([]models.DiagnosticReport, error) {
	return s.diagnosticRepo.GetDiagnosticLabReportName(patientId)
}

func (s *DiagnosticServiceImpl) CheckReportExistWithSampleDateTestComponent(reportData models.LabReport, patientId uint64, recordId *uint64, processID uuid.UUID, attachmentId *string, reportId *uint64) error {
	log.Println("reportData.ReportDetails.CollectionDate ", reportData.ReportDetails.CollectionDate)
	step := string(constant.CheckReportDuplication)
	msg := fmt.Sprintf("Processing report id %d %s %s", *recordId, reportData.ReportDetails.ReportName, string(constant.CheckReportDuplicationMsg))
	errorMsg := ""
	s.processStatusService.LogStep(processID, step, constant.Running, msg, errorMsg, nil, nil, nil, nil, nil, attachmentId)
	CollectionDate, err := utils.ParseDate(reportData.ReportDetails.CollectionDate)
	if err != nil {
		log.Println("Collection date parsing failed:", err)
		s.processStatusService.LogStepAndFail(processID, step, constant.Failure, fmt.Sprintf("Collection date parsing failed : %s", reportData.ReportDetails.CollectionDate), err.Error(), nil, nil, attachmentId)
		return err
	}
	existingMap, err := s.diagnosticRepo.GetSampleCollectionDateTestComponentMap(patientId, CollectionDate)
	if err != nil {
		log.Println("Error fetching existing components:", err)
		s.processStatusService.LogStepAndFail(processID, step, constant.Failure, "Error while fetching existing componed based on collection date", err.Error(), nil, nil, attachmentId)
	}
	var allComponentNames []string
	for _, testData := range reportData.Tests {
		for _, component := range testData.Components {
			allComponentNames = append(allComponentNames, strings.ToLower(component.TestComponentName))
		}
	}
	log.Printf("Components being checked: %+v", allComponentNames)
	reportData.ReportDetails.IsDeleted = 0
	if ShouldSkipReport(CollectionDate, allComponentNames, existingMap) {
		log.Println("Save report in duplicate bucket and marked is_deleted as True for patient : ", patientId)
		msg = fmt.Sprintf("Report id %d saved in duplicate bucket and marked deleted for logged in UserID :%d  || sample collection date :%s || Fetched test component list : %s || Existing test component list : %+v ", *recordId, patientId, utils.FormatDateTime(&CollectionDate), allComponentNames, existingMap)
		s.processStatusService.LogStep(processID, step, constant.Success, msg, errorMsg, nil, nil, nil, nil, nil, attachmentId)
		reportData.ReportDetails.IsDeleted = 0
		_, updateErr := s.medicalRecordsRepo.UpdateTblMedicalRecord(&models.TblMedicalRecord{RecordId: *recordId, IsDeleted: 0, RecordCategory: string(constant.DUPLICATE)})
		if updateErr != nil {
			log.Println("failed to update medicalRecordService.UpdateTblMedicalRecord:", updateErr)
		}
	}
	if _, err := s.DigitizeDiagnosticReport(reportData, patientId, recordId, reportId); err != nil {
		return err
	}
	s.processStatusService.LogStep(processID, step, constant.Success, string(constant.ReportDuplicationSuccess), errorMsg, nil, nil, nil, nil, nil, nil)
	return nil
}

func ShouldSkipReport(collectionDate time.Time, componentNames []string, existingMap map[string]bool) bool {
	formattedCollectionDate := utils.FormatDateTime(&collectionDate)
	var matched int
	var missing []string
	total := len(componentNames)
	if total == 0 {
		log.Println("No components in report — cannot evaluate duplication.")
		return false
	}
	for _, name := range componentNames {
		key := fmt.Sprintf("%s|%s", formattedCollectionDate, strings.ToLower(name))
		if existingMap[key] {
			matched++
		} else {
			missing = append(missing, key)
		}
	}
	matchPercentage := (float64(matched) / float64(total)) * 100
	percentage := float64(config.PropConfig.SystemVaribale.Score)
	log.Printf("Match percentage to check duplicate report or not : %.2f%% (matched: %d / total: %d)", matchPercentage, matched, total)

	if matchPercentage >= percentage {
		log.Printf("Match ≥ %.2f%% → Treating report as duplicate. Skipping save.", percentage)
		return true
	}

	log.Printf("Match < %.2f%% → Treating report as new. Missing keys: %+v", percentage, missing)
	return false
}
