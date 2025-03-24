package controller

import (
	"biostat/constant"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DietController struct {
	dietService service.DietService
}

func NewDietController(dietService service.DietService) *DietController {
	return &DietController{dietService: dietService}
}

func (dc *DietController) AddDietPlanTemplate(c *gin.Context) {
	var dietPlan models.DietPlanTemplate
	if err := c.ShouldBindJSON(&dietPlan); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid diet plan input", nil, err)
		return
	}

	if err := dc.dietService.CreateDietPlanTemplate(&dietPlan); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add diet plan", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Diet plan added successfully", dietPlan, nil, nil)
}
func (dc *DietController) GetAllDietPlanTemplates(c *gin.Context) {
	page, limit, offset := utils.GetPaginationParams(c)
	dietPlans, totalRecords, err := dc.dietService.GetDietPlanTemplates(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve diet plans", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	statusCode, message := utils.GetResponseStatusMessage(
		len(dietPlans),
		"Diet plans retrieved successfully",
		"No diet plans found",
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, dietPlans, pagination, nil)
}

func (dc *DietController) GetDietPlanById(c *gin.Context) {
	dietPlanTemplateId := c.Param("diet_id")
	var dietPlan models.DietPlanTemplate
	dietPlan, err := dc.dietService.GetDietPlanById(dietPlanTemplateId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Diet plan not found", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diet plan retrieved successfully", dietPlan, nil, nil)
}

func (dc *DietController) UpdateDietPlanTemplate(c *gin.Context) {
	dietPlanTemplateId := c.Param("diet_id")
	var dietPlan models.DietPlanTemplate

	if err := c.ShouldBindJSON(&dietPlan); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid diet plan input", nil, err)
		return
	}

	if err := dc.dietService.UpdateDietPlanTemplate(dietPlanTemplateId, &dietPlan); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update diet plan", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diet plan updated successfully", dietPlan, nil, nil)
}
