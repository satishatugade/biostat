package controller

import (
	"biostat/constant"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MedicationController struct {
	medicationService service.MedicationService
}

func NewMedicationController(medicationService service.MedicationService) *MedicationController {
	return &MedicationController{medicationService: medicationService}
}

func (mc *MedicationController) GetAllMedication(c *gin.Context) {
	page, limit, offset := utils.GetPaginationParams(c)
	medications, totalRecords, err := mc.medicationService.GetMedications(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve medications", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	statusCode, message := utils.GetResponseStatusMessage(
		len(medications),
		"Medication info retrieved successfully",
		"Medication info not found",
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, medications, pagination, nil)
}

func (mc *MedicationController) AddMedication(c *gin.Context) {
	var medication models.Medication
	if err := c.ShouldBindJSON(&medication); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid medication input", nil, err)
		return
	}

	if err := mc.medicationService.CreateMedication(&medication); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add medication", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Medication added successfully", medication, nil, nil)
}

// func (mc *MedicationController) UpdateMedication(c *gin.Context) {
// 	medicationId, err := strconv.Atoi(c.Param("medication_id"))
// 	if err != nil {
// 		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid medication Id", nil, err)
// 		return
// 	}

// 	var medication models.Medication
// 	if err := c.ShouldBindJSON(&medication); err != nil {
// 		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid medication input", nil, err)
// 		return
// 	}

// 	medication.MedicationId = uint(medicationId)

// 	if err := mc.medicationService.UpdateMedication(&medication); err != nil {
// 		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update medication", nil, err)
// 		return
// 	}

// 	models.SuccessResponse(c, constant.Success, http.StatusOK, "Medication updated successfully", medication, nil, nil)
// }

func (mc *MedicationController) UpdateMedication(c *gin.Context) {
	var medication models.Medication
	if err := c.ShouldBindJSON(&medication); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid medication update input", nil, err)
		return
	}

	if err := mc.medicationService.UpdateMedication(&medication); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update medication", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Medication updated successfully", medication, nil, nil)
}
