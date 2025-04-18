package controller

import (
	"biostat/constant"
	"biostat/database"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PatientController struct {
	patientService       service.PatientService
	dietService          service.DietService
	allergyService       service.AllergyService
	medicalRecordService service.TblMedicalRecordService
	medicationService    service.MedicationService
	appointmentService   service.AppointmentService
	diagnosticService    service.DiagnosticService
}

func NewPatientController(patientService service.PatientService, dietService service.DietService, allergyService service.AllergyService, medicalRecordService service.TblMedicalRecordService,
	medicationService service.MedicationService, appointmentService service.AppointmentService, diagnosticService service.DiagnosticService) *PatientController {
	return &PatientController{patientService: patientService, dietService: dietService,
		allergyService: allergyService, medicalRecordService: medicalRecordService,
		medicationService: medicationService, appointmentService: appointmentService,
		diagnosticService: diagnosticService}
}

func (pc *PatientController) GetAllRelation(c *gin.Context) {
	relations, err := pc.patientService.GetAllRelation()
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch relations", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Relations fetched successfully", relations, nil, nil)
}

func (pc *PatientController) GetRelationById(c *gin.Context) {
	relationIdStr := c.Param("relation_id")
	relationId, err := strconv.ParseUint(relationIdStr, 10, 32)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid relation ID", nil, err)
		return
	}

	relation, err := pc.patientService.GetRelationById(int(relationId))
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Relation not found", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Relation fetched successfully", relation, nil, nil)
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

	patient, err := pc.patientService.GetPatientById(&patientId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient not found", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient info retrieved successfully", patient, nil, nil)
}

func (pc *PatientController) UpdatePatientInfoById(c *gin.Context) {
	authUserId, exists := utils.GetUserDataContext(c)
	if !exists {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, constant.KeyCloakErrorMessage, nil, nil)
		return
	}

	var patientData models.Patient
	if err := c.ShouldBindJSON(&patientData); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input data", nil, err)
		return
	}

	updatedPatient, err := pc.patientService.UpdatePatientById(authUserId, &patientData)
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

func (pc *PatientController) GetPatientDiagnosticResultValues(c *gin.Context) {
	var req models.DiagnosticResultRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	diseaseProfiles, err := pc.patientService.GetPatientDiagnosticResultValue(req.PatientId, req.PatientDiagnosticReportId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient report not found", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient report fetch successfully", diseaseProfiles, nil, nil)
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

func (pc *PatientController) GetPrescriptionByPatientId(c *gin.Context) {
	patientID := c.Param("patient_id")
	page, limit, offset := utils.GetPaginationParams(c)

	prescriptions, totalRecords, err := pc.patientService.GetPrescriptionByPatientId(patientID, limit, offset)
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

func (pc *PatientController) GetPatientRelativeList(c *gin.Context) {
	patientIdParam := c.Param("patient_id")
	var patientId *uint64
	if patientIdParam != "" {
		id, err := strconv.ParseUint(patientIdParam, 10, 64)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid patient_user_id", nil, err)
			return
		}
		patientId = &id
	}
	relatives, err := pc.patientService.GetRelativeList(patientId)
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

func (pc *PatientController) GetRelativeList(c *gin.Context) {

	relatives, err := pc.patientService.GetRelativeList(nil)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch relatives", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(relatives),
		"Relatives retrieved successfully",
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

func (pc *PatientController) GetPatientCaregiverList(c *gin.Context) {
	patientIdParam := c.Param("patient_id")
	var patientId *uint64
	if patientIdParam != "" {
		id, err := strconv.ParseUint(patientIdParam, 10, 64)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid patient_user_id", nil, err)
			return
		}
		patientId = &id
	}

	caregivers, err := pc.patientService.GetCaregiverList(patientId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch patient caregivers", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(caregivers),
		"Patient caregiver retrieved successfully",
		"Caregiver not found",
	)

	models.SuccessResponse(c, constant.Success, statusCode, message, caregivers, nil, nil)
}

func (pc *PatientController) GetCaregiverList(c *gin.Context) {

	caregivers, err := pc.patientService.GetCaregiverList(nil)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch patient caregivers", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(caregivers),
		"Caregiver list retrieved successfully",
		"Caregiver not found",
	)

	models.SuccessResponse(c, constant.Success, statusCode, message, caregivers, nil, nil)
}

func (pc *PatientController) GetPatientDoctorList(c *gin.Context) {
	patientIdParam := c.Param("patient_id")
	var patientId *uint64
	if patientIdParam != "" {
		id, err := strconv.ParseUint(patientIdParam, 10, 64)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid patient_user_id", nil, err)
			return
		}
		patientId = &id
	}

	doctors, err := pc.patientService.GetDoctorList(patientId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch patient doctor", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(doctors),
		"Patient doctor retrieved successfully",
		"Doctor not found",
	)

	models.SuccessResponse(c, constant.Success, statusCode, message, doctors, nil, nil)
}

func (pc *PatientController) GetDoctorList(c *gin.Context) {

	doctors, err := pc.patientService.GetDoctorList(nil)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch patient doctor", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(doctors),
		"Doctor list retrieved successfully",
		"Doctor not found",
	)

	models.SuccessResponse(c, constant.Success, statusCode, message, doctors, nil, nil)
}

func (pc *PatientController) GetPatientList(c *gin.Context) {

	patients, err := pc.patientService.GetPatientList()
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch patient list", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(patients),
		"Patient retrieved successfully",
		"Patient not found",
	)

	models.SuccessResponse(c, constant.Success, statusCode, message, patients, nil, nil)
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

func (pc *PatientController) GetUserProfile(ctx *gin.Context) {
	type UserRequest struct {
		User string `json:"user"`
	}
	var req UserRequest
	roles, rolesExists := ctx.Get("userRoles")
	if !rolesExists {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "User not found", nil, errors.New("Error while getting profile"))
		return
	}
	sub, subExists := ctx.Get("sub")
	if !subExists {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "User not found", nil, errors.New("Error while getting profile"))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	if !utils.StringInSlice(req.User, roles.([]string)) {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, nil)
		return
	}

	user_id, err := pc.patientService.GetUserIdBySUB(sub.(string))

	user, err := pc.patientService.GetUserProfileByUserId(user_id)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "Failed to load profile", nil, err)
		return
	}
	userProfile := utils.MapUserToRoleSchema(*user, req.User)

	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "User profile retrieved successfully", userProfile, nil, nil)
}

func (pc *PatientController) GetUserOnBoardingStatus(ctx *gin.Context) {
	sub, subExists := ctx.Get("sub")
	if !subExists {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "User not found", nil, errors.New("Error while getting profile"))
		return
	}
	basicDetailsAdded, familyDetailsAdded, healthDetailsAdded, err := pc.patientService.GetUserOnboardingStatusByUID(sub.(string))
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Error while gwtting Onboarding status", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Onboarding details retrieved successfully", gin.H{"basic_details": basicDetailsAdded, "family_details": familyDetailsAdded, "health_details": healthDetailsAdded}, nil, nil)
}
func (mc *PatientController) GetMedication(c *gin.Context) {
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

func (pc *PatientController) GetNursesList(c *gin.Context) {
	page, limit, offset := utils.GetPaginationParams(c)
	nurses, totalRecords, err := pc.patientService.GetNursesList(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch nurses", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	statusCode, message := utils.GetResponseStatusMessage(
		len(nurses),
		"Nurses list retrieved successfully",
		"Nurses not found",
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, nurses, pagination, nil)
}

func (pc *PatientController) ScheduleAppointment(ctx *gin.Context) {
	sub, subExists := ctx.Get("sub")
	if !subExists {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "User not found", nil, errors.New("Error while getting "))
		return
	}
	var appointment models.Appointment
	if err := ctx.ShouldBindJSON(&appointment); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}

	user_id, err := pc.patientService.GetUserIdBySUB(sub.(string))
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "User can not be authorised", nil, err)
		return
	}

	appointment.ScheduledBy = user_id
	if appointment.ProviderType == "doctor" {
		isDocPresent, err := pc.patientService.ExistsByUserIdAndRoleId(appointment.ProviderID, 2)
		if err != nil || !isDocPresent {
			models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to check doctor existence", nil, err)
			return
		}
	} else if appointment.ProviderType == "nurse" {
		isPresent, err := pc.patientService.ExistsByUserIdAndRoleId(appointment.ProviderID, 6)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to check nurse existence", nil, err)
			return
		}
		if !isPresent {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Nurse not found", nil, err)
			return
		}
	} else {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid provider type", nil, errors.New("Invalid provider type"))
		return
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Recovered in CreateAppointment:", r)
			models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to schedule appointment", nil, errors.New("Failed to schedule appointment"))
			return
		}
	}()

	createdAppointment, err := pc.appointmentService.CreateAppointment(tx, &appointment)
	if err != nil {
		tx.Rollback()
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Appointment could not be scheduled", nil, err)
		return
	}
	if err := tx.Commit().Error; err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to schedule appointment", nil, err)
		return
	}
	user, _ := pc.patientService.GetUserProfileByUserId(createdAppointment.ProviderID)
	providerInfo := utils.MapUserToPublicProviderInfo(*user, createdAppointment.ProviderType)
	appointmentResponse := models.AppointmentResponse{
		AppointmentID:   appointment.AppointmentID,
		PatientID:       appointment.PatientID,
		ProviderType:    appointment.ProviderType,
		ProviderInfo:    providerInfo,
		ScheduledBy:     appointment.ScheduledBy,
		AppointmentType: appointment.AppointmentType,
		AppointmentDate: appointment.AppointmentDate,
		AppointmentTime: appointment.AppointmentTime,
		DurationMinutes: appointment.DurationMinutes,
		IsInperson:      appointment.IsInperson,
		Status:          appointment.Status,
		MeetingUrl:      appointment.MeetingUrl,
		PaymentStatus:   appointment.PaymentStatus,
		Notes:           appointment.Notes,
		PaymentID:       appointment.PaymentID,
		CreatedAt:       appointment.CreatedAt,
		UpdatedAt:       appointment.UpdatedAt,
	}

	models.SuccessResponse(ctx, constant.Success, http.StatusCreated, "Appointment created scheduled", appointmentResponse, nil, nil)
	return
}

func (pc *PatientController) GetUserAppointments(ctx *gin.Context) {
	sub, subExists := ctx.Get("sub")
	if !subExists {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "User not found", nil, errors.New("Error while getting "))
		return
	}
	user_id, err := pc.patientService.GetUserIdBySUB(sub.(string))
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "User can not be authorised", nil, err)
		return
	}
	appointments, err := pc.appointmentService.GetUserAppointments(user_id)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to fetch appointments", nil, err)
		return
	}
	var responses []models.AppointmentResponse
	for _, appointment := range appointments {
		user, _ := pc.patientService.GetUserProfileByUserId(appointment.ProviderID)
		providerInfo := utils.MapUserToPublicProviderInfo(*user, appointment.ProviderType)
		appointmentResponse := models.AppointmentResponse{
			AppointmentID:   appointment.AppointmentID,
			PatientID:       appointment.PatientID,
			ProviderType:    appointment.ProviderType,
			ProviderInfo:    providerInfo,
			ScheduledBy:     appointment.ScheduledBy,
			AppointmentType: appointment.AppointmentType,
			AppointmentDate: appointment.AppointmentDate,
			AppointmentTime: appointment.AppointmentTime,
			DurationMinutes: appointment.DurationMinutes,
			IsInperson:      appointment.IsInperson,
			Status:          appointment.Status,
			MeetingUrl:      appointment.MeetingUrl,
			PaymentStatus:   appointment.PaymentStatus,
			Notes:           appointment.Notes,
			PaymentID:       appointment.PaymentID,
			CreatedAt:       appointment.CreatedAt,
			UpdatedAt:       appointment.UpdatedAt,
		}
		responses = append(responses, appointmentResponse)
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(responses),
		"Appointments retrieved successfully",
		"Appointments not found",
	)
	models.SuccessResponse(ctx, constant.Success, statusCode, message, responses, nil, nil)
	return
}

func (mc *PatientController) GetAllLabs(c *gin.Context) {
	_, exists := utils.GetUserDataContext(c)
	if !exists {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, constant.KeyCloakErrorMessage, nil, nil)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	data, totalRecords, err := mc.diagnosticService.GetAllLabs(page, limit)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve labs", nil, err)
		return
	}

	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	statusCode, message := utils.GetResponseStatusMessage(
		len(data),
		"Diagnostic labs retrieved successfully",
		"Diagnostic labs not found",
	)

	models.SuccessResponse(c, constant.Success, statusCode, message, data, pagination, nil)
}
