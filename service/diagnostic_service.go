package service

import (
	"biostat/database"
	"biostat/models"
	"biostat/repository"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type DiagnosticService interface {
	CreateLab(lab *models.DiagnosticLab) (*models.DiagnosticLab, error)
	GetAllLabs(limit, offset int) ([]models.DiagnosticLab, int64, error)
	GetPatientDiagnosticLabs(patientid uint64, limit int, offset int) ([]models.DiagnosticLabResponse, int64, error)
	GetSinglePatientDiagnosticLab(patientId uint64, labId *uint64) (*models.DiagnosticLabResponse, error)
	GetLabById(diagnosticlLabId uint64) (*models.DiagnosticLab, error)
	UpdateLab(diagnosticlLab *models.DiagnosticLab, authUserId string) error
	DeleteLab(diagnosticlLabId uint64, authUserId string) error
	GetAllDiagnosticLabAuditRecords(limit, offset int) ([]models.DiagnosticLabAudit, int64, error)
	AddMapping(userId uint64, LabInfo *models.DiagnosticLab) error
	GetDiagnosticLabAuditRecord(labId, labAuditId uint64) ([]models.DiagnosticLabAudit, error)

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
	DigitizeDiagnosticReport(reportData models.LabReport, patientId uint64, recordId *uint64) (string, error)
	NotifyAbnormalResult(patientId uint64) error
	ArchivePatientDiagnosticReport(reportID uint64, isDeleted int) error
}

type DiagnosticServiceImpl struct {
	diagnosticRepo repository.DiagnosticRepository
	emailService   EmailService
	patientService PatientService
}

func NewDiagnosticService(repo repository.DiagnosticRepository, emailService EmailService, patientService PatientService) DiagnosticService {
	return &DiagnosticServiceImpl{diagnosticRepo: repo, emailService: emailService, patientService: patientService}
}

func (s *DiagnosticServiceImpl) GetDiagnosticTests(limit int, offset int) ([]models.DiagnosticTest, int64, error) {
	return s.diagnosticRepo.GetAllDiagnosticTests(limit, offset)
}

func (s *DiagnosticServiceImpl) AddMapping(patientId uint64, labInfo *models.DiagnosticLab) error {
	return s.diagnosticRepo.AddMapping(patientId, labInfo)
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

func (s *DiagnosticServiceImpl) CreateLab(lab *models.DiagnosticLab) (*models.DiagnosticLab, error) {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	createdLab, err := s.diagnosticRepo.CreateLab(tx, lab)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
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
func (s *DiagnosticServiceImpl) GetAllDiagnosticLabAuditRecords(limit, offset int) ([]models.DiagnosticLabAudit, int64, error) {
	return s.diagnosticRepo.GetAllDiagnosticLabAuditRecords(limit, offset)
}

func (s *DiagnosticServiceImpl) GetDiagnosticLabAuditRecord(labId, labAuditId uint64) ([]models.DiagnosticLabAudit, error) {
	return s.diagnosticRepo.GetDiagnosticLabAuditRecord(labId, labAuditId)
}

func (s *DiagnosticServiceImpl) GetAllLabs(limit, offset int) ([]models.DiagnosticLab, int64, error) {
	return s.diagnosticRepo.GetAllLabs(limit, offset)
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

func (s *DiagnosticServiceImpl) DigitizeDiagnosticReport(reportData models.LabReport, patientId uint64, recordId *uint64) (string, error) {
	testNameCache, componentNameCache := s.diagnosticRepo.LoadDiagnosticTestMasterData()
	if testNameCache == nil || componentNameCache == nil {
		log.Println("Failed to load master data")
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
			tx.Rollback()
		}
	}()

	var diagnosticLabId uint64
	labName := reportData.ReportDetails.LabName
	// if labName == "" {
	// 	log.Println("Lab name is empty, skipping lab creation/mapping.")
	// 	return "", fmt.Errorf("lab name is required to proceed")
	// }
	if val, exists := diagnosticLabs[strings.ToLower(labName)]; exists {
		diagnosticLabId = val
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
			return "", fmt.Errorf("error while saving diagnostic lab info: %w", err) // Wrap error
		}
		diagnosticLabId = labInfo.DiagnosticLabId
	}

	var parsedDate time.Time
	var err error
	if reportData.ReportDetails.ReportDate != "" {
		layouts := []string{
			time.RFC3339,
			"02-Jan-2006",
			"02-Jan-06",
			"02/01/2006",
		}
		for _, layout := range layouts {
			parsedDate, err = time.Parse(layout, reportData.ReportDetails.ReportDate)
			if err == nil {
				break
			}
		}
		if err != nil {
			log.Println("Invalid date format:", err)
			tx.Rollback()
			return "", fmt.Errorf("invalid date format: %w", err)
		}
	} else {
		log.Println("Empty report date, using default date")
		parsedDate = time.Now()
	}

	patientReport := models.PatientDiagnosticReport{
		DiagnosticLabId: diagnosticLabId,
		PatientId:       patientId,
		PaymentStatus:   "Pending",
		ReportName:      reportData.ReportDetails.ReportName,
		CollectedDate:   parsedDate,
		ReportDate:      parsedDate,
		Observation:     "",
		IsDigital:       reportData.ReportDetails.IsDigital,
		CollectedAt:     reportData.ReportDetails.LabLocation,
	}
	reportInfo, err := s.diagnosticRepo.GeneratePatientDiagnosticReport(tx, &patientReport)
	if err != nil {
		log.Println("ERROR saving PatientDiagnosticReport:", err)
		tx.Rollback()
		return "", fmt.Errorf("error while saving patient diagnostic report: %w", err)
	}
	recordmapping := models.PatientReportAttachment{
		PatientDiagnosticReportId: reportInfo.PatientDiagnosticReportId,
		RecordId:                  *recordId,
	}
	if err := s.diagnosticRepo.SavePatientReportAttachmentMapping(tx, &recordmapping); err != nil {
		log.Println("Error while creating SavePatientReportAttachmentMapping:", err)
		tx.Rollback()
		return "", fmt.Errorf("error while SavePatientReportAttachmentMapping: %w", err)
	}
	for _, testData := range reportData.Tests {
		testName := testData.TestName
		testInterpretation := testData.Interpretation

		var diagnosticTestId uint64
		if id, exists := testNameCache[strings.ToLower(testName)]; exists {
			diagnosticTestId = id
		} else {
			log.Println("DiagnosticTest not available in database creating new test")
			newTest := models.DiagnosticTest{TestName: testName}
			testInfo, err := s.diagnosticRepo.CreateDiagnosticTest(tx, &newTest, "System")
			if err != nil {
				log.Println("Error while creating DiagnosticTest:", err)
				tx.Rollback()
				return "", fmt.Errorf("error while creating diagnostic test: %w", err) // Wrap error
			}
			diagnosticTestId = testInfo.DiagnosticTestId
			testNameCache[strings.ToLower(testInfo.TestName)] = diagnosticTestId
		}

		//Save patient test
		testRecord := models.PatientDiagnosticTest{
			PatientDiagnosticReportId: reportInfo.PatientDiagnosticReportId,
			DiagnosticTestId:          diagnosticTestId,
			TestNote:                  testInterpretation,
			TestDate:                  parsedDate,
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
			if id, exists := componentNameCache[strings.ToLower(component.TestComponentName)]; exists {
				diagnosticComponentId = id
				result := models.PatientDiagnosticTestResultValue{
					DiagnosticTestId:          diagnosticTestId,
					DiagnosticTestComponentId: diagnosticComponentId,
					ResultStatus:              GetResultStatus(component.ResultValue, component.ReferenceRange.Min, component.ReferenceRange.Max),
					ResultValue:               func() float64 { v, _ := strconv.ParseFloat(component.ResultValue, 64); return v }(),
					ResultDate:                parsedDate,
					PatientId:                 patientId,
					PatientDiagnosticReportId: reportInfo.PatientDiagnosticReportId,
				}
				_, err := s.diagnosticRepo.SavePatientReportResultValue(tx, &result)
				if err != nil {
					log.Println("ERROR saving test result:", err)
					tx.Rollback()
					return "", fmt.Errorf("error while saving test result: %w", err)
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
				componentNameCache[strings.ToLower(componentInfo.TestComponentName)] = diagnosticComponentId

				mapping := models.DiagnosticTestComponentMapping{
					DiagnosticTestId:      diagnosticTestId,
					DiagnosticComponentId: diagnosticComponentId,
				}
				if _, err := s.diagnosticRepo.CreateDiagnosticTestComponentMapping(tx, &mapping); err != nil {
					log.Println("Error while creating DiagnosticTestComponentMapping:", err)
					tx.Rollback()
					return "", fmt.Errorf("error while creating diagnostic test component mapping: %w", err) // Wrap error
				}
				referenceRange := models.DiagnosticTestReferenceRange{
					DiagnosticTestId:          diagnosticTestId,
					DiagnosticTestComponentId: diagnosticComponentId,
					NormalMin:                 func() float64 { v, _ := strconv.ParseFloat(component.ReferenceRange.Min, 64); return v }(),
					NormalMax:                 func() float64 { v, _ := strconv.ParseFloat(component.ReferenceRange.Max, 64); return v }(),
					Units:                     component.Units,
				}
				refRangeErr := s.diagnosticRepo.AddTestReferenceRange(&referenceRange)
				if refRangeErr != nil {
					log.Println("ERROR saving test Ref. range:", refRangeErr)
					tx.Rollback()
					return "", fmt.Errorf("error while saving test reference range: %w", refRangeErr)
				}
				result := models.PatientDiagnosticTestResultValue{
					DiagnosticTestId:          diagnosticTestId,
					DiagnosticTestComponentId: diagnosticComponentId,
					ResultStatus:              GetResultStatus(component.ResultValue, component.ReferenceRange.Min, component.ReferenceRange.Max),
					ResultValue:               func() float64 { v, _ := strconv.ParseFloat(component.ResultValue, 64); return v }(),
					ResultDate:                parsedDate,
					PatientId:                 patientId,
					PatientDiagnosticReportId: reportInfo.PatientDiagnosticReportId,
				}
				_, resultValueErr := s.diagnosticRepo.SavePatientReportResultValue(tx, &result)
				if resultValueErr != nil {
					log.Println("ERROR saving test result:", resultValueErr)
					tx.Rollback()
					return "", fmt.Errorf("error while saving test result: %w", resultValueErr)
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

func GetResultStatus(resultVal, minStr, maxStr string) string {
	if resultVal == "" || minStr == "" || maxStr == "" {
		return "-"
	}
	result, err1 := strconv.ParseFloat(resultVal, 64)
	min, err2 := strconv.ParseFloat(minStr, 64)
	max, err3 := strconv.ParseFloat(maxStr, 64)

	if err1 != nil || err2 != nil || err3 != nil {
		return "-"
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
