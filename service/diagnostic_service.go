package service

import (
	"biostat/models"
	"biostat/repository"
)

type DiagnosticService interface {
	CreateLab(lab *models.DiagnosticLab) error
	GetAllLabs(page int, limit int) ([]models.DiagnosticLab, int64, error)
	GetLabById(diagnosticlLabId uint64) (*models.DiagnosticLab, error)
	UpdateLab(diagnosticlLab *models.DiagnosticLab, authUserId string) error
	DeleteLab(diagnosticlLabId uint64, authUserId string) error
	GetAllDiagnosticLabAuditRecords(page, limit int) ([]models.DiagnosticLabAudit, int64, error)
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
	GetAllTestRefRangeView(limit int, offset int) ([]models.DiagnosticTestReferenceRange, int64, error)
	ViewTestReferenceRange(testReferenceRangeId uint64) (*models.DiagnosticTestReferenceRange, error)
	GetTestReferenceRangeAuditRecord(testReferenceRangeId, auditId uint64, limit, offset int) ([]models.DiagnosticTestReferenceRangeAudit, int64, error)
}

type DiagnosticServiceImpl struct {
	diagnosticRepo repository.DiagnosticRepository
}

func NewDiagnosticService(repo repository.DiagnosticRepository) DiagnosticService {
	return &DiagnosticServiceImpl{diagnosticRepo: repo}
}

func (s *DiagnosticServiceImpl) GetDiagnosticTests(limit int, offset int) ([]models.DiagnosticTest, int64, error) {
	return s.diagnosticRepo.GetAllDiagnosticTests(limit, offset)
}

func (s *DiagnosticServiceImpl) CreateDiagnosticTest(diagnosticTest *models.DiagnosticTest, createdBy string) (*models.DiagnosticTest, error) {
	return s.diagnosticRepo.CreateDiagnosticTest(diagnosticTest, createdBy)
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
	return s.diagnosticRepo.CreateDiagnosticComponent(diagnosticComponent)
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
	return s.diagnosticRepo.CreateDiagnosticTestComponentMapping(diagnosticTestComponentMapping)
}

func (s *DiagnosticServiceImpl) UpdateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error) {
	return s.diagnosticRepo.UpdateDiagnosticTestComponentMapping(diagnosticTestComponentMapping)
}

func (s *DiagnosticServiceImpl) DeleteDiagnosticTestComponentMapping(diagnosticTestId uint64, diagnosticComponentId uint64) error {
	return s.diagnosticRepo.DeleteDiagnosticTestComponentMapping(diagnosticTestId, diagnosticComponentId)
}

func (s *DiagnosticServiceImpl) CreateLab(lab *models.DiagnosticLab) error {
	return s.diagnosticRepo.CreateLab(lab)
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
func (s *DiagnosticServiceImpl) GetAllDiagnosticLabAuditRecords(page, limit int) ([]models.DiagnosticLabAudit, int64, error) {
	return s.diagnosticRepo.GetAllDiagnosticLabAuditRecords(page, limit)
}

func (s *DiagnosticServiceImpl) GetDiagnosticLabAuditRecord(labId, labAuditId uint64) ([]models.DiagnosticLabAudit, error) {
	return s.diagnosticRepo.GetDiagnosticLabAuditRecord(labId, labAuditId)
}

func (s *DiagnosticServiceImpl) GetAllLabs(page, limit int) ([]models.DiagnosticLab, int64, error) {
	return s.diagnosticRepo.GetAllLabs(page, limit)
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

func (s *DiagnosticServiceImpl) GetAllTestRefRangeView(limit int, offset int) ([]models.DiagnosticTestReferenceRange, int64, error) {
	return s.diagnosticRepo.GetAllTestRefRangeView(limit, offset)
}

func (s *DiagnosticServiceImpl) GetTestReferenceRangeAuditRecord(testReferenceRangeId, auditId uint64, limit, offset int) ([]models.DiagnosticTestReferenceRangeAudit, int64, error) {
	return s.diagnosticRepo.GetTestReferenceRangeAuditRecord(testReferenceRangeId, auditId, limit, offset)
}
