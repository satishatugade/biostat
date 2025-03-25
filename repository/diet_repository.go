package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type DietRepository interface {
	CreateDietPlanTemplate(dietPlan *models.DietPlanTemplate) error
	GetDietPlanTemplates(limit, offset int) ([]models.DietPlanTemplate, int64, error)
	GetDietPlanById(dietPlanTemplateId string) (models.DietPlanTemplate, error)
	UpdateDietPlanTemplate(dietPlanTemplateId string, dietPlan *models.DietPlanTemplate) error
	GetPatientDietPlan(patientId string) ([]models.PatientDietPlan, error)
}

type DietRepositoryImpl struct {
	db *gorm.DB
}

func NewDietRepository(db *gorm.DB) DietRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &DietRepositoryImpl{db: db}
}

// CreateDietPlanTemplate implements DietRepository.
func (d *DietRepositoryImpl) CreateDietPlanTemplate(dietPlan *models.DietPlanTemplate) error {
	return d.db.Create(dietPlan).Error
}

func (d *DietRepositoryImpl) GetDietPlanTemplates(limit, offset int) ([]models.DietPlanTemplate, int64, error) {
	var dietPlans []models.DietPlanTemplate
	var totalRecords int64

	if err := d.db.Model(&models.DietPlanTemplate{}).Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	if err := d.db.Limit(limit).Offset(offset).Find(&dietPlans).Error; err != nil {
		return nil, 0, err
	}

	return dietPlans, totalRecords, nil
}

func (d *DietRepositoryImpl) GetDietPlanById(dietPlanTemplateId string) (models.DietPlanTemplate, error) {
	var dietPlan models.DietPlanTemplate
	if err := d.db.First(&dietPlan, "diet_plan_template_id = ?", dietPlanTemplateId).Error; err != nil {
		return dietPlan, err
	}
	return dietPlan, nil
}

func (d *DietRepositoryImpl) UpdateDietPlanTemplate(dietPlanTemplateId string, dietPlan *models.DietPlanTemplate) error {
	return d.db.Model(&models.DietPlanTemplate{}).Where("diet_plan_template_id = ?", dietPlanTemplateId).Updates(dietPlan).Error
}

func (d *DietRepositoryImpl) GetPatientDietPlan(patientId string) ([]models.PatientDietPlan, error) {
	var dietPlans []models.PatientDietPlan

	err := d.db.Preload("DietPlanTemplate").
		Preload("DietPlanTemplate.Meals").
		Preload("DietPlanTemplate.Meals.Nutrients").
		Preload("DietCreator").
		Where("patient_id = ?", patientId).
		Find(&dietPlans).Error

	if err != nil {
		return nil, err
	}

	return dietPlans, nil
}
