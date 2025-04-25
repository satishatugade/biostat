package service

import (
	"biostat/models"
	"biostat/repository"
)

type DietService interface {
	CreateDietPlanTemplate(dietPlan *models.DietPlanTemplate) error
	GetDietPlanTemplates(limit, offset int) ([]models.DietPlanTemplate, int64, error)
	GetDietPlanById(dietPlanTemplateId string) (models.DietPlanTemplate, error)
	UpdateDietPlanTemplate(dietPlanTemplateId uint64, dietPlan *models.DietPlanTemplate) error
	GetPatientDietPlan(patientId string) ([]models.PatientDietPlan, error)
	AddDiseaseDietMapping(mapping *models.DiseaseDietMapping) error
}

type DietServiceImpl struct {
	dietRepo repository.DietRepository
}

func NewDietService(repo repository.DietRepository) DietService {
	return &DietServiceImpl{dietRepo: repo}
}

func (d *DietServiceImpl) CreateDietPlanTemplate(exercise *models.DietPlanTemplate) error {
	return d.dietRepo.CreateDietPlanTemplate(exercise)
}

func (d *DietServiceImpl) GetDietPlanTemplates(limit, offset int) ([]models.DietPlanTemplate, int64, error) {
	return d.dietRepo.GetDietPlanTemplates(limit, offset)
}

func (d *DietServiceImpl) GetDietPlanById(id string) (models.DietPlanTemplate, error) {
	return d.dietRepo.GetDietPlanById(id)
}

func (d *DietServiceImpl) UpdateDietPlanTemplate(dietPlanTemplateId uint64, dietPlan *models.DietPlanTemplate) error {
	return d.dietRepo.UpdateDietPlanTemplate(dietPlanTemplateId, dietPlan)
}

func (s *DietServiceImpl) GetPatientDietPlan(patientId string) ([]models.PatientDietPlan, error) {
	return s.dietRepo.GetPatientDietPlan(patientId)
}

func (s *DietServiceImpl) AddDiseaseDietMapping(mapping *models.DiseaseDietMapping) error {
	return s.dietRepo.AddDiseaseDietMapping(mapping)
}
