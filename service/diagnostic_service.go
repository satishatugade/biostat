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
}

type diagnosticServiceImpl struct {
	diagnosticRepo repository.DiagnosticRepository
}

func NewDiagnosticService(repo repository.DiagnosticRepository) DiagnosticService {
	return &diagnosticServiceImpl{diagnosticRepo: repo}
}

func (s *diagnosticServiceImpl) GetDiagnosticTests(limit int, offset int) ([]models.DiagnosticTest, int64, error) {
	return s.diagnosticRepo.GetAllDiagnosticTests(limit, offset)
}

func (s *diagnosticServiceImpl) CreateDiagnosticTest(diagnosticTest *models.DiagnosticTest, createdBy string) (*models.DiagnosticTest, error) {
	return s.diagnosticRepo.CreateDiagnosticTest(diagnosticTest, createdBy)
}

func (s *diagnosticServiceImpl) UpdateDiagnosticTest(diagnosticTest *models.DiagnosticTest, updatedBy string) (*models.DiagnosticTest, error) {
	return s.diagnosticRepo.UpdateDiagnosticTest(diagnosticTest, updatedBy)
}

func (s *diagnosticServiceImpl) GetSingleDiagnosticTest(diagnosticTestId int) (*models.DiagnosticTest, error) {
	return s.diagnosticRepo.GetSingleDiagnosticTest(diagnosticTestId)
}

func (s *diagnosticServiceImpl) DeleteDiagnosticTest(diagnosticTestId int, updatedBy string) error {
	return s.diagnosticRepo.DeleteDiagnosticTest(diagnosticTestId, updatedBy)
}

func (s *diagnosticServiceImpl) GetAllDiagnosticComponents(limit int, offset int) ([]models.DiagnosticTestComponent, int64, error) {
	return s.diagnosticRepo.GetAllDiagnosticComponents(limit, offset)
}

func (s *diagnosticServiceImpl) CreateDiagnosticComponent(diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error) {
	return s.diagnosticRepo.CreateDiagnosticComponent(diagnosticComponent)
}

func (s *diagnosticServiceImpl) UpdateDiagnosticComponent(authUserId string, diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error) {
	return s.diagnosticRepo.UpdateDiagnosticComponent(authUserId, diagnosticComponent)
}

func (s *diagnosticServiceImpl) DeleteDiagnosticTestComponent(diagnosticTestComponetId uint64, authUserId string) error {
	return s.diagnosticRepo.DeleteDiagnosticTestComponent(diagnosticTestComponetId, authUserId)
}

func (s *diagnosticServiceImpl) GetSingleDiagnosticComponent(diagnosticComponentId int) (*models.DiagnosticTestComponent, error) {
	return s.diagnosticRepo.GetSingleDiagnosticComponent(diagnosticComponentId)
}

func (s *diagnosticServiceImpl) GetAllDiagnosticTestComponentMappings(limit int, offset int) ([]models.DiagnosticTestComponentMapping, int64, error) {
	return s.diagnosticRepo.GetAllDiagnosticTestComponentMappings(limit, offset)
}

func (s *diagnosticServiceImpl) CreateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error) {
	return s.diagnosticRepo.CreateDiagnosticTestComponentMapping(diagnosticTestComponentMapping)
}

func (s *diagnosticServiceImpl) UpdateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error) {
	return s.diagnosticRepo.UpdateDiagnosticTestComponentMapping(diagnosticTestComponentMapping)
}

func (s *diagnosticServiceImpl) DeleteDiagnosticTestComponentMapping(diagnosticTestId uint64, diagnosticComponentId uint64) error {
	return s.diagnosticRepo.DeleteDiagnosticTestComponentMapping(diagnosticTestId, diagnosticComponentId)
}

func (s *diagnosticServiceImpl) CreateLab(lab *models.DiagnosticLab) error {
	return s.diagnosticRepo.CreateLab(lab)
}

func (s *diagnosticServiceImpl) GetLabById(id uint64) (*models.DiagnosticLab, error) {
	return s.diagnosticRepo.GetLabById(id)
}

func (s *diagnosticServiceImpl) UpdateLab(diagnosticlLabId *models.DiagnosticLab, authUserId string) error {
	return s.diagnosticRepo.UpdateLab(diagnosticlLabId, authUserId)
}

func (s *diagnosticServiceImpl) DeleteLab(diagnosticlLabId uint64, authUserId string) error {
	return s.diagnosticRepo.DeleteLab(diagnosticlLabId, authUserId)
}
func (s *diagnosticServiceImpl) GetAllDiagnosticLabAuditRecords(page, limit int) ([]models.DiagnosticLabAudit, int64, error) {
	return s.diagnosticRepo.GetAllDiagnosticLabAuditRecords(page, limit)
}

func (s *diagnosticServiceImpl) GetDiagnosticLabAuditRecord(labId, labAuditId uint64) ([]models.DiagnosticLabAudit, error) {
	return s.diagnosticRepo.GetDiagnosticLabAuditRecord(labId, labAuditId)
}

func (s *diagnosticServiceImpl) GetAllLabs(page, limit int) ([]models.DiagnosticLab, int64, error) {
	return s.diagnosticRepo.GetAllLabs(page, limit)
}

func (s *diagnosticServiceImpl) AddDiseaseDiagnosticTestMapping(mapping *models.DiseaseDiagnosticTestMapping) error {
	return s.diagnosticRepo.AddDiseaseDiagnosticTestMapping(mapping)
}
