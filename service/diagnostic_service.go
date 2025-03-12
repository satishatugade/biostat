package service

import (
	"biostat/models"
	"biostat/repository"
)

type DiagnosticService interface {
	GetDiagnosticTests(limit int, offset int) ([]models.DiagnosticTest, int64, error)
	CreateDiagnosticTest(diagnosticTest *models.DiagnosticTest,createdBy string) (*models.DiagnosticTest, error)
	UpdateDiagnosticTest(diagnosticTest *models.DiagnosticTest,updatedBy string) (*models.DiagnosticTest, error)
	GetSingleDiagnosticTest(diagnosticTestId int) (*models.DiagnosticTest, error)
	DeleteDiagnosticTest(diagnosticTestId int,updatedBy string) error

	GetAllDiagnosticComponents(limit int, offset int) ([]models.DiagnosticTestComponent, int64, error)
	CreateDiagnosticComponent(diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error)
	UpdateDiagnosticComponent(diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error)
	GetSingleDiagnosticComponent(diagnosticComponentId int) (*models.DiagnosticTestComponent, error)

	GetAllDiagnosticTestComponentMappings(limit int, offset int) ([]models.DiagnosticTestComponentMapping, int64, error)
	CreateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error)
	UpdateDiagnosticTestComponentMapping(diagnosticTestComponentMapping *models.DiagnosticTestComponentMapping) (*models.DiagnosticTestComponentMapping, error)
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

func (s *diagnosticServiceImpl) CreateDiagnosticTest(diagnosticTest *models.DiagnosticTest,createdBy string) (*models.DiagnosticTest, error) {
	return s.diagnosticRepo.CreateDiagnosticTest(diagnosticTest,createdBy)
}

func (s *diagnosticServiceImpl) UpdateDiagnosticTest(diagnosticTest *models.DiagnosticTest,updatedBy string) (*models.DiagnosticTest, error) {
	return s.diagnosticRepo.UpdateDiagnosticTest(diagnosticTest,updatedBy)
}

func (s *diagnosticServiceImpl) GetSingleDiagnosticTest(diagnosticTestId int) (*models.DiagnosticTest, error) {
	return s.diagnosticRepo.GetSingleDiagnosticTest(diagnosticTestId)
}

func (s *diagnosticServiceImpl) DeleteDiagnosticTest(diagnosticTestId int,updatedBy string) error {
	return s.diagnosticRepo.DeleteDiagnosticTest(diagnosticTestId,updatedBy)
}

func (s *diagnosticServiceImpl) GetAllDiagnosticComponents(limit int, offset int) ([]models.DiagnosticTestComponent, int64, error) {
	return s.diagnosticRepo.GetAllDiagnosticComponents(limit, offset)
}

func (s *diagnosticServiceImpl) CreateDiagnosticComponent(diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error) {
	return s.diagnosticRepo.CreateDiagnosticComponent(diagnosticComponent)
}

func (s *diagnosticServiceImpl) UpdateDiagnosticComponent(diagnosticComponent *models.DiagnosticTestComponent) (*models.DiagnosticTestComponent, error) {
	return s.diagnosticRepo.UpdateDiagnosticComponent(diagnosticComponent)
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

