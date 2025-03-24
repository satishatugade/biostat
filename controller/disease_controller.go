package controller

import (
	"biostat/constant"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DiseaseController struct {
	diseaseService service.DiseaseService
	causeService   service.CauseService
	symptomService service.SymptomService
}

func NewDiseaseController(diseaseService service.DiseaseService, causeService service.CauseService, symptomService service.SymptomService) *DiseaseController {
	return &DiseaseController{
		diseaseService: diseaseService,
		causeService:   causeService,
		symptomService: symptomService,
	}
}

func (dc *DiseaseController) CreateDisease(c *gin.Context) {
	var disease models.Disease
	if err := c.ShouldBindJSON(&disease); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request payload", nil, err)
		return
	}
	err := dc.diseaseService.CreateDisease(&disease)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to create disease", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Disease created successfully", nil, nil, nil)
}

func (dc *DiseaseController) GetDiseaseInfo(c *gin.Context) {
	var totalRecords int64
	diseaseId, err := strconv.ParseUint(c.Param("disease_id"), 10, 32)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid disease ID", nil, err)
		return
	}

	page, limit, offset := utils.GetPaginationParams(c)

	diseases, err := dc.diseaseService.GetDiseases(uint(diseaseId))
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve diseases", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	message := "Diseases info not found"
	if diseases.DiseaseId != 0 {
		message = "Diseases info retrieved successfully"
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, message, diseases, pagination, nil)
}

// func (dc *DiseaseController) GetDiseaseInfo(c *gin.Context) {
// 	// var totalRecords int64
// 	var diseases interface{}
// 	// var err error
// 	var message string
// 	diseaseIdParam := c.Param("disease_id")
// 	if diseaseIdParam != "" {
// 		// Convert disease_id to uint
// 		diseaseId, err := strconv.ParseUint(diseaseIdParam, 10, 32)
// 		if err != nil {
// 			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid disease ID", nil, err)
// 			return
// 		}

// 		// Fetch a single disease
// 		disease, err := dc.diseaseService.GetDiseases(uint(diseaseId))
// 		if err != nil {
// 			models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Disease not found", nil, err)
// 			return
// 		}
// 		message = "Disease info retrieved successfully"
// 		models.SuccessResponse(c, constant.Success, http.StatusOK, message, disease, nil, nil)
// 	} else {
// 		// Fetch all diseases with pagination
// 		// Convert disease_id to uint
// 		diseaseId, err := strconv.ParseUint(diseaseIdParam, 10, 32)
// 		if err != nil {
// 			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid disease ID", nil, err)
// 			return
// 		}

// 		page, limit, offset := utils.GetPaginationParams(c)
// 		allDiseases, totalRecords, err := dc.diseaseService.GetAllDiseasesInfo(uint(diseaseId), limit, offset)
// 		if err != nil {
// 			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve diseases", nil, err)
// 			return
// 		}

// 		message = "Diseases info not found"
// 		if len(allDiseases) > 0 {
// 			message = "Diseases info retrieved successfully"
// 		}

// 		diseases = allDiseases
// 		pagination := utils.GetPagination(limit, page, offset, totalRecords)
// 		models.SuccessResponse(c, constant.Success, http.StatusOK, message, diseases, pagination, nil)
// 	}
// }

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

func (dc *DiseaseController) GetAllSymptom(c *gin.Context) {
	page, limit, offset := utils.GetPaginationParams(c)

	symptom, totalRecords, err := dc.symptomService.GetAllSymptom(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve causes", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(symptom),
		"Symptom retrieved successfully",
		"Disease Symptom not found",
	)
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	models.SuccessResponse(c, constant.Success, statusCode, message, symptom, pagination, nil)
}

func (dc *DiseaseController) AddSymptom(c *gin.Context) {
	var symptom models.Symptom

	if err := c.ShouldBindJSON(&symptom); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid symptom input", nil, err)
		return
	}

	err := dc.symptomService.AddDiseaseSymptom(&symptom)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add symptom", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "symptom added successfully", symptom, nil, nil)
}

func (dc *DiseaseController) UpdateDiseaseSymptom(c *gin.Context) {
	var symptom models.Symptom

	if err := c.ShouldBindJSON(&symptom); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid disease symptom input", nil, err)
		return
	}

	if symptom.SymptomId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "symptom Id is required", nil, nil)
		return
	}

	err := dc.symptomService.UpdateSymptom(&symptom)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update symptom", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "symptom updated successfully", symptom, nil, nil)
}
