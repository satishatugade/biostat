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

type PatientController struct {
	patientService       service.PatientService
	dietService          service.DietService
	allergyService       service.AllergyService
	medicalRecordService service.TblMedicalRecordService
}

func NewPatientController(patientService service.PatientService, dietService service.DietService, allergyService service.AllergyService, medicalRecordService service.TblMedicalRecordService) *PatientController {
	return &PatientController{patientService: patientService, dietService: dietService, allergyService: allergyService, medicalRecordService: medicalRecordService}
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

func (pc *PatientController) GetPatientByID(c *gin.Context) {
	patientIdStr := c.Param("patient_id")
	patientId, err := strconv.ParseUint(patientIdStr, 10, 32)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid patient ID", nil, err)
		return
	}

	patient, err := pc.patientService.GetPatientById(uint(patientId))
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient not found", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient info retrieved successfully", patient, nil, nil)
}

func (pc *PatientController) UpdatePatientInfoById(c *gin.Context) {
	patientId := c.Param("patient_id")

	var patientData models.Patient
	if err := c.ShouldBindJSON(&patientData); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input data", nil, err)
		return
	}

	updatedPatient, err := pc.patientService.UpdatePatientById(patientId, &patientData)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update patient info", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient info updated successfully", updatedPatient, nil, nil)
}

func (pc *PatientController) GetPatientDiseaseProfiles(c *gin.Context) {
	PatientId := c.Param("patient_id")

	diseaseProfiles, err := pc.patientService.GetPatientDiseaseProfiles(PatientId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient disease profiles not found", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient disease profiles retrieved successfully", diseaseProfiles, nil, nil)
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

func (pc *PatientController) GetAllPrescription(c *gin.Context) {
	page, limit, offset := utils.GetPaginationParams(c)
	prescription, totalRecords, err := pc.patientService.GetAllPrescription(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve prescription", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	statusCode, message := utils.GetResponseStatusMessage(
		len(prescription),
		"prescription info retrieved successfully",
		"Prescription info not found",
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, prescription, pagination, nil)
}

func (pc *PatientController) GetPrescriptionByPatientID(c *gin.Context) {
	patientID := c.Param("patient_id")
	page, limit, offset := utils.GetPaginationParams(c)

	prescriptions, totalRecords, err := pc.patientService.GetPrescriptionByPatientID(patientID, limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve prescriptions", nil, err)
		return
	}

	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	statusCode, message := utils.GetResponseStatusMessage(
		len(prescriptions),
		"Prescription info retrieved successfully",
		"Prescription info not found",
	)

	models.SuccessResponse(c, constant.Success, statusCode, message, prescriptions, pagination, nil)
}

func (pc *PatientController) GetPatientDietPlan(c *gin.Context) {
	patientId := c.Param("patient_id")

	dietPlans, err := pc.dietService.GetPatientDietPlan(patientId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch diet plans", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(dietPlans),
		"Patient Diet plans retrieved successfully",
		"Diet plan not found",
	)

	models.SuccessResponse(c, constant.Success, statusCode, message, dietPlans, nil, nil)
}

func (pc *PatientController) AddPatientRelative(c *gin.Context) {
	var relative models.PatientRelative
	if err := c.ShouldBindJSON(&relative); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}

	if err := pc.patientService.AddPatientRelative(&relative); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to create patient relative", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Patient relative created successfully", relative, nil, nil)
}

func (pc *PatientController) GetPatientRelative(c *gin.Context) {
	patientId := c.Param("patient_id")

	relatives, err := pc.patientService.GetPatientRelative(patientId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch patient relatives", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(relatives),
		"Patient relatives retrieved successfully",
		"Relatives not found",
	)

	models.SuccessResponse(c, constant.Success, statusCode, message, relatives, nil, nil)
}

func (pc *PatientController) GetPatientRelativeByRelativeId(c *gin.Context) {
	relativeIdStr := c.Param("relative_id")
	relativeId, err := strconv.ParseUint(relativeIdStr, 10, 32)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid relative ID", nil, err)
		return
	}

	relative, err := pc.patientService.GetPatientRelativeById(uint(relativeId))
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Relative not found", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient relative retrieved successfully", relative, nil, nil)
}

func (pc *PatientController) UpdatePatientRelative(c *gin.Context) {
	relativeIdStr := c.Param("relative_id")

	relativeId, err := strconv.ParseUint(relativeIdStr, 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid relative ID", nil, err)
		return
	}

	var updatedRelative models.PatientRelative
	if err := c.ShouldBindJSON(&updatedRelative); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}

	patientRelative, err := pc.patientService.UpdatePatientRelative(uint(relativeId), &updatedRelative)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update patient relative", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient relative updated successfully", patientRelative, nil, nil)
}

func (pc *PatientController) AddPatientAllergyRestriction(c *gin.Context) {
	var allergy models.PatientAllergyRestriction
	if err := c.ShouldBindJSON(&allergy); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid allergy input", nil, err)
		return
	}

	if err := pc.allergyService.AddPatientAllergyRestriction(&allergy); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add allergy", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Allergy restriction added successfully", allergy, nil, nil)
}

func (pc *PatientController) GetPatientAllergyRestriction(c *gin.Context) {
	patientId := c.Param("patient_id")
	allergies, err := pc.allergyService.GetPatientAllergyRestriction(patientId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch allergies", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Allergy restrictions fetched successfully", allergies, nil, nil)
}

func (pc *PatientController) UpdatePatientAllergyRestriction(c *gin.Context) {
	var allergyUpdate models.PatientAllergyRestriction
	if err := c.ShouldBindJSON(&allergyUpdate); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid allergy update input", nil, err)
		return
	}

	if allergyUpdate.PatientAllergyRestrictionId == 0 || allergyUpdate.PatientId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Patient ID and Allergy Restriction ID are required", nil, nil)
		return
	}

	if err := pc.allergyService.UpdatePatientAllergyRestriction(&allergyUpdate); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update allergy", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Allergy updated successfully", allergyUpdate, nil, nil)
}

func (pc *PatientController) AddPatientClinicalRange(c *gin.Context) {
	var customeRange models.PatientCustomRange
	if err := c.ShouldBindJSON(&customeRange); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input data", nil, err)
		return
	}
	if err := pc.patientService.AddPatientClinicalRange(&customeRange); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add clinical range", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Clinical range added successfully", customeRange, nil, nil)
}

func (c *PatientController) GetUserMedicalRecords(ctx *gin.Context) {
	userID := utils.GetParamAsInt(ctx, "user_id")

	records, err := c.medicalRecordService.GetUserMedicalRecords(int64(userID))
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to retrieve records", nil, err)
		return
	}
	message := "Data not found"
	if len(records) > 0 {
		message = "User records retrieved successfully"
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, message, records, nil, nil)
}

func (c *PatientController) GetAllTblMedicalRecords(ctx *gin.Context) {
	page, limit, offset := utils.GetPaginationParams(ctx)

	data, total, err := c.medicalRecordService.GetAllTblMedicalRecords(limit, offset)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to retrieve records", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, total)
	message := "Data not found"
	if len(data) > 0 {
		message = "Data retrieved successfully"
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, message, data, pagination, nil)
}

func (c *PatientController) CreateTblMedicalRecord(ctx *gin.Context) {

	payload, err := utils.ProcessFileUpload(ctx)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "File processing failed", nil, err)
		return
	}
	payload.UploadSource = ctx.PostForm("upload_source")
	payload.Description = ctx.PostForm("description")
	payload.RecordType = ctx.PostForm("record_type")

	createdBy := 124

	data, err := c.medicalRecordService.CreateTblMedicalRecord(payload, int64(createdBy))
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to create record", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Record created successfully", data, nil, nil)
}

func (c *PatientController) UpdateTblMedicalRecord(ctx *gin.Context) {
	var payload models.TblMedicalRecord
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	if payload.RecordId == 0 {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Param id is required", nil, nil)
		return
	}
	updatedBy := ctx.GetString("user")
	data, err := c.medicalRecordService.UpdateTblMedicalRecord(&payload, updatedBy)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to update record", nil, nil)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Record updated successfully", data, nil, nil)
}

func (c *PatientController) GetSingleTblMedicalRecord(ctx *gin.Context) {
	id := utils.GetParamAsInt(ctx, "id")
	if id == 0 {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Param id is required", nil, nil)
		return
	}
	data, err := c.medicalRecordService.GetSingleTblMedicalRecord(id)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to retrieve record", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Record retrieved successfully", data, nil, nil)
}

func (c *PatientController) DeleteTblMedicalRecord(ctx *gin.Context) {
	id := utils.GetParamAsInt(ctx, "id")
	if id == 0 {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Param id is required", nil, nil)
		return
	}

	updatedBy := ctx.GetString("user")
	err := c.medicalRecordService.DeleteTblMedicalRecord(id, updatedBy)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to delete record", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Record deleted successfully", nil, nil, nil)
}
