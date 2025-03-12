package controller

import (
	"biostat/constant"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DiseaseController struct {
	diseaseService service.DiseaseService
	causeService   service.CauseService
}

// func NewDiseaseController(diseaseService service.DiseaseService) *DiseaseController {
// 	return &DiseaseController{diseaseService: diseaseService}
// }

func NewDiseaseController(diseaseService service.DiseaseService, causeService service.CauseService) *DiseaseController {
	return &DiseaseController{
		diseaseService: diseaseService,
		causeService:   causeService,
	}
}

func (dc *DiseaseController) GetDiseaseInfo(c *gin.Context) {
	var diseases []models.Disease
	var totalRecords int64

	page, limit, offset := utils.GetPaginationParams(c)

	diseases, totalRecords, err := dc.diseaseService.GetDiseases(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve diseases", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	message := "Diseases info not found"
	if len(diseases) > 0 {
		message = "Diseases info retrieved successfully"
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, message, diseases, pagination, nil)
}

func (dc *DiseaseController) GetDiseaseProfile(c *gin.Context) {

	var diseaseProfiles []models.DiseaseProfile
	var totalRecords int64

	page, limit, offset := utils.GetPaginationParams(c)
	diseaseProfiles, totalRecords, err := dc.diseaseService.GetDiseaseProfiles(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve disease profile", nil, err)
		return
	}

	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	message := "Diseases profile not found"
	if len(diseaseProfiles) > 0 {
		message = "Diseases profile info retrieved successfully"
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, message, diseaseProfiles, pagination, nil)
}

// Get all causes with pagination
func (dc *DiseaseController) GetAllCauses(c *gin.Context) {
	page, limit, offset := utils.GetPaginationParams(c)

	causes, totalRecords, err := dc.causeService.GetAllCauses(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve causes", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(causes),
		"Causes retrieved successfully",
		"Disease causes not found",
	)
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	models.SuccessResponse(c, constant.Success, statusCode, message, causes, pagination, nil)
}

func (dc *DiseaseController) AddDiseaseCause(c *gin.Context) {
	var cause models.Cause

	if err := c.ShouldBindJSON(&cause); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid cause input", nil, err)
		return
	}

	err := dc.causeService.AddDiseaseCause(&cause)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add cause", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Cause added successfully", cause, nil, nil)
}

func (dc *DiseaseController) UpdateDiseaseCause(c *gin.Context) {
	var cause models.Cause

	if err := c.ShouldBindJSON(&cause); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid disease cause input", nil, err)
		return
	}

	if cause.CauseId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Cause Id is required", nil, nil)
		return
	}

	err := dc.causeService.UpdateCause(&cause)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update cause", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Cause updated successfully", cause, nil, nil)
}
