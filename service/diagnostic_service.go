package service

import (
	"biostat/models"
	"biostat/repository"
)

type DiagnosticService interface {
	GetDiagnosticTests(limit int, offset int) ([]models.DiagnosticTest, int64, error)
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