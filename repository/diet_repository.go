package repository

import (
	"biostat/models"
	"time"

	"gorm.io/gorm"
)

type DietRepository interface {
	CreateDietPlanTemplate(dietPlan *models.DietPlanTemplate) error
	GetDietPlanTemplates(limit, offset int) ([]models.DietPlanTemplate, int64, error)
	GetDietPlanById(dietPlanTemplateId string) (models.DietPlanTemplate, error)
	UpdateDietPlanTemplate(dietPlanTemplateId uint64, dietPlan *models.DietPlanTemplate) error
	GetPatientDietPlan(patientId string) ([]models.PatientDietPlan, error)
	AddDiseaseDietMapping(mapping *models.DiseaseDietMapping) error
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

	if err := d.db.Order("diet_plan_template_id DESC").Limit(limit).Offset(offset).Preload("Meals").
		Preload("Meals.Nutrients").Find(&dietPlans).Error; err != nil {
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

func (d *DietRepositoryImpl) UpdateDietPlanTemplate(dietPlanTemplateId uint64, dietPlan *models.DietPlanTemplate) error {
	tx := d.db.Begin()

	// Step 1: Update DietPlanTemplate
	if err := tx.Model(&models.DietPlanTemplate{}).
		Where("diet_plan_template_id = ?", dietPlanTemplateId).
		Updates(map[string]interface{}{
			"name":        dietPlan.Name,
			"description": dietPlan.Description,
			"goal":        dietPlan.Goal,
			"notes":       dietPlan.Notes,
			"updated_at":  time.Now(),
			"created_by":  dietPlan.CreatedBy,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Step 2: Iterate over each meal
	for _, meal := range dietPlan.Meals {
		meal.DietPlanTemplateId = dietPlan.DietPlanTemplateId

		// If MealId exists, update; otherwise, create
		if meal.MealId != 0 {
			if err := tx.Model(&models.Meal{}).
				Where("meal_id = ?", meal.MealId).
				Updates(map[string]interface{}{
					"meal_type":   meal.MealType,
					"description": meal.Description,
					"updated_at":  time.Now(),
					"created_by":  meal.CreatedBy,
				}).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			meal.CreatedAt = time.Now()
			meal.UpdatedAt = time.Now()
			if err := tx.Create(&meal).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		// Step 3: Nutrients inside each meal
		for _, nutrient := range meal.Nutrients {
			nutrient.MealId = meal.MealId

			if nutrient.NutrientId != 0 {
				if err := tx.Model(&models.Nutrient{}).
					Where("nutrient_id = ?", nutrient.NutrientId).
					Updates(map[string]interface{}{
						"nutrient_name": nutrient.NutrientName,
						"amount":        nutrient.Amount,
						"unit":          nutrient.Unit,
						"updated_at":    time.Now(),
						"created_by":    nutrient.CreatedBy,
					}).Error; err != nil {
					tx.Rollback()
					return err
				}
			} else {
				nutrient.CreatedAt = time.Now()
				nutrient.UpdatedAt = time.Now()
				if err := tx.Create(&nutrient).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	return tx.Commit().Error
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

func (r *DietRepositoryImpl) AddDiseaseDietMapping(mapping *models.DiseaseDietMapping) error {
	return r.db.Create(mapping).Error
}
