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
}

func NewDiseaseController(diseaseService service.DiseaseService) *DiseaseController {
	return &DiseaseController{diseaseService: diseaseService}
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
