package controller

import (
	"biostat/constant"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PatientController struct {
	patientService service.PatientService
}

func NewPatientController(patientService service.PatientService) *PatientController {
	return &PatientController{patientService: patientService}
}

func (pc *PatientController) GetPatientInfo(c *gin.Context) {
	page, limit, offset := utils.GetPaginationParams(c)
	patients, totalRecords, err := pc.patientService.GetPatients(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve patients", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	message := "Patient info not found"
	if len(patients) > 0 {
		message = "Patient info retrieved successfully"
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, message, patients, pagination, nil)
}

func (pc *PatientController) AddPrescription(c *gin.Context) {
	var prescription models.PatientPrescription

	if err := c.ShouldBindJSON(&prescription); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid patient prescription input data", nil, err)
		return
	}
	err := pc.patientService.AddPatientPrescription(&prescription)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add patient prescription", nil, err)
		return
	}
	message := "Patient prescription added."
	models.SuccessResponse(c, constant.Success, http.StatusOK, message, prescription, nil, nil)
}

// func (pc *PatientController) UpdatePrescription(c *gin.Context) {

// 	prescriptionId := c.Param("prescription_id")
// 	id, err := strconv.ParseUint(prescriptionId, 10, 64)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid prescription ID"})
// 		return
// 	}
// 	var prescription models.PatientPrescription

// 	if err := c.ShouldBindJSON(&prescription); err != nil {
// 		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid patient prescription input data", nil, err)
// 		return
// 	}
// 	prescription.PrescriptionId = uint(id)
// 	err1 := pc.patientService.UpdatePrescription(&prescription)
// 	if err != nil {
// 		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update patient prescription", nil, err1)
// 		return
// 	}

// 	message := "Patient prescription updated successfully."
// 	models.SuccessResponse(c, constant.Success, http.StatusOK, message, prescription, nil, nil)
// }
