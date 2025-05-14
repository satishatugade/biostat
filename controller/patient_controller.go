package controller

import (
	"biostat/auth"
	"biostat/constant"
	"biostat/database"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	userService          service.UserService
	apiService           service.ApiService
	diseaseService       service.DiseaseService
	smsService           service.SmsService
	emailService         *service.EmailService
}

func NewPatientController(patientService service.PatientService, dietService service.DietService,
	allergyService service.AllergyService, medicalRecordService service.TblMedicalRecordService,
	medicationService service.MedicationService, appointmentService service.AppointmentService,
	diagnosticService service.DiagnosticService, userService service.UserService,
	apiService service.ApiService, diseaseService service.DiseaseService, smsService service.SmsService, emailService *service.EmailService) *PatientController {
	return &PatientController{patientService: patientService, dietService: dietService,
		allergyService: allergyService, medicalRecordService: medicalRecordService,
		medicationService: medicationService, appointmentService: appointmentService,
		diagnosticService: diagnosticService,
		userService:       userService,
		apiService:        apiService,
		diseaseService:    diseaseService,
		smsService:        smsService,
		emailService:      emailService,
	}
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
	user_id, err := utils.GetUserIDFromContext(c, pc.patientService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	diseaseProfiles, err := pc.patientService.GetPatientDiseaseProfiles(user_id)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient disease profiles not found", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient disease profiles retrieved successfully", diseaseProfiles, nil, nil)
}

func (pc *PatientController) GetPatientDiagnosticResultValues(c *gin.Context) {

	// authUserId, exists := utils.GetUserDataContext(c)
	// if !exists {
	// 	models.ErrorResponse(c, constant.Failure, http.StatusNotFound, constant.KeyCloakErrorMessage, nil, nil)
	// 	return
	// }
	// patientId, err := pc.patientService.GetUserIdByAuthUserId(authUserId)
	// if err != nil {
	// 	models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient not found", nil, err)
	// 	return
	// }
	user_id, err := utils.GetUserIDFromContext(c, pc.patientService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	patientDiagnosticReportId, ok := utils.GetQueryIntParam(c, "patient_report_id", 0)
	if !ok {
		log.Println("GetPatientDiagnosticResultValue patientDiagnosticReportId status not provided : ", patientDiagnosticReportId)
	}
	summaryParam := c.DefaultQuery("summary", "false")
	summary := false
	if parsed, err := strconv.ParseBool(summaryParam); err == nil {
		summary = parsed
	}
	if summary {
		reportSummaryData, err1 := pc.patientService.GetPatientDiagnosticReportSummary(user_id, patientDiagnosticReportId, summary)
		if err1 != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient report not found", nil, err1)
			return
		}
		models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient report summary fetch successfully", reportSummaryData, nil, nil)
	} else {
		reportValues, err := pc.patientService.GetPatientDiagnosticResultValue(user_id, patientDiagnosticReportId)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient report not found", nil, err)
			return
		}
		models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient report fetched successfully", reportValues, nil, nil)
	}
}

func (pc *PatientController) GetPatientDiagnosticTrendValue(c *gin.Context) {
	authUserId, exists := utils.GetUserDataContext(c)
	if !exists {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, constant.KeyCloakErrorMessage, nil, nil)
		return
	}
	var req models.DiagnosticResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("DiagnosticResultRequest filter result values")
	}
	patientId, err := pc.patientService.GetUserIdByAuthUserId(authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient not found", nil, err)
		return
	}
	req.PatientId = patientId
	results, err := pc.patientService.GetPatientDiagnosticTrendValue(req)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient reports not found", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient reports fetched successfully", results, nil, nil)
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
	// patientIdParam := c.Param("patient_id")
	// var patientId *uint64
	// if patientIdParam != "" {
	// 	id, err := strconv.ParseUint(patientIdParam, 10, 64)
	// 	if err != nil {
	// 		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid patient_user_id", nil, err)
	// 		return
	// 	}
	// 	patientId = &id
	// }
	user_id, err := utils.GetUserIDFromContext(c, pc.patientService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	relatives, err := pc.patientService.GetRelativeList(&user_id)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch patient relatives", nil, err)
		return
	}
	_, message := utils.GetResponseStatusMessage(
		len(relatives),
		"Patient relatives retrieved successfully",
		"Relatives not found",
	)

	models.SuccessResponse(c, constant.Success, http.StatusOK, message, relatives, nil, nil)
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
	authUserId, exists := utils.GetUserDataContext(c)
	if !exists {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, constant.KeyCloakErrorMessage, nil, nil)
		return
	}
	patientId, err := pc.patientService.GetUserIdByAuthUserId(authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient not found", nil, err)
		return
	}
	relativeIdStr := c.Param("relative_id")
	relativeId, err := strconv.ParseUint(relativeIdStr, 10, 32)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid relative ID", nil, err)
		return
	}

	relative, err := pc.patientService.GetPatientRelativeById(uint64(relativeId), patientId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Relative not found", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient relative retrieved successfully", relative, nil, nil)
}

func (pc *PatientController) UpdatePatientRelative(c *gin.Context) {
	_, exists := utils.GetUserDataContext(c)
	if !exists {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, constant.KeyCloakErrorMessage, nil, nil)
		return
	}
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
	page, limit, offset := utils.GetPaginationParams(c)

	doctors, totalRecords, err := pc.patientService.GetDoctorList(patientId, page, limit)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch patient doctor", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	statusCode, message := utils.GetResponseStatusMessage(
		len(doctors),
		"Patient doctor retrieved successfully",
		"Doctor not found",
	)

	models.SuccessResponse(c, constant.Success, statusCode, message, doctors, pagination, nil)
}

func (pc *PatientController) GetDoctorList(c *gin.Context) {
	page, limit, offset := utils.GetPaginationParams(c)
	doctors, totalRecords, err := pc.patientService.GetDoctorList(nil, limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch patient doctor", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	doctorProfile := utils.MapUsersToSchema(doctors, "doctor")
	statusCode, message := utils.GetResponseStatusMessage(
		len(doctors),
		"Doctor list retrieved successfully",
		"Doctor not found",
	)

	models.SuccessResponse(c, constant.Success, statusCode, message, doctorProfile, pagination, nil)
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
	tx := database.DB.Begin()
	if tx.Error != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to initiate transaction", nil, tx.Error)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Recovered in AddPatientAllergyRestriction:", r)
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to save allergy details", nil, errors.New("Failed to save allergy details"))
			return
		}
	}()

	if err := pc.allergyService.AddPatientAllergyRestriction(tx, &allergy); err != nil {
		tx.Rollback()
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add allergy", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Allergy restriction added successfully", allergy, nil, nil)
	return
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
	// userID := utils.GetParamAsInt(ctx, "user_id")
	user_id, err := utils.GetUserIDFromContext(ctx, c.patientService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	records, err := c.medicalRecordService.GetUserMedicalRecords(int64(user_id))
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
	authUserId, exists := utils.GetUserDataContext(ctx)
	if !exists {
		models.ErrorResponse(ctx, constant.Failure, http.StatusNotFound, constant.KeyCloakErrorMessage, nil, nil)
		return
	}
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "Please provid file to save", nil, errors.New("Error while uploading document"))
		return
	}
	user_id, err := c.patientService.GetUserIdBySUB(authUserId)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "User can not be authorised", nil, err)
		return
	}

	userDigiToken, err := c.userService.GetSingleTblUserToken(user_id, "DigiLocker")
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Digilocker Token not found", nil, err)
		return
	}

	// fileData, err := utils.ReadFileBytes(file)
	// if err != nil {
	// 	models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "User can not be authorised", nil, err)
	// 	return
	// }
	var fileBuf bytes.Buffer
	tee := io.TeeReader(file, &fileBuf)

	fileData, err := io.ReadAll(tee)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Failed to read file", nil, err)
		return
	}

	newRecord, err := service.SaveRecordToDigiLocker(userDigiToken.AuthToken, fileData, header.Filename, header.Header.Get("Content-Type"))
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Failed to Upload user record to Digilocker", nil, err)
		return
	}

	// payload, err := utils.ProcessFileUpload(ctx)
	// if err != nil {
	// 	models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "File processing failed", nil, err)
	// 	return
	// }
	newRecord.UploadSource = ctx.PostForm("upload_source")
	newRecord.Description = ctx.PostForm("description")
	newRecord.RecordCategory = ctx.PostForm("record_category")
	newRecord.SourceAccount = userDigiToken.ProviderId
	newRecord.UploadedBy = user_id

	data, err := c.medicalRecordService.CreateTblMedicalRecord(newRecord, user_id, &fileBuf, header.Filename)
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
	// authUserId, exists := utils.GetUserDataContext(ctx)
	// if !exists {
	// 	models.ErrorResponse(ctx, constant.Failure, http.StatusNotFound, constant.KeyCloakErrorMessage, nil, nil)
	// 	return
	// }

	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	if !utils.StringInSlice(req.User, roles.([]string)) {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, nil)
		return
	}

	user_id, err := utils.GetUserIDFromContext(ctx, pc.patientService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	user, err := pc.patientService.GetUserProfileByUserId(user_id)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "Failed to load profile", nil, err)
		return
	}
	userProfile := utils.MapUserToRoleSchema(*user, req.User)

	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "User profile retrieved successfully", userProfile, nil, nil)
}

func (pc *PatientController) GetUserOnBoardingStatus(ctx *gin.Context) {
	user_id, err := utils.GetUserIDFromContext(ctx, pc.patientService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	basicDetailsAdded, familyDetailsAdded, healthDetailsAdded, noOfUpcomingAppointments, noOfMedicationsForDashboard, noOfMessagesForDashboard, noOfLabReusltsForDashboard, err := pc.patientService.GetUserOnboardingStatusByUID(user_id)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Error while gwtting Onboarding status", nil, err)
		return
	}

	var status models.ThirdPartyTokenStatus

	gmail, _ := pc.userService.GetSingleTblUserToken(user_id, "gmail")
	if gmail != nil {
		status.GmailPresent = true
	}
	digilocker, _ := pc.userService.GetSingleTblUserToken(user_id, "DigiLocker")
	if digilocker != nil {
		status.DigiLockerPresent = true
		createdAtUTC := digilocker.CreatedAt
		nowLoc := time.Now().UTC()

		duration := createdAtUTC.Sub(nowLoc)
		hoursDiff := duration.Hours()
		fmt.Println(hoursDiff)

		if time.Since(digilocker.CreatedAt.UTC()) > time.Hour {
			status.IsDLExpired = true
		} else {
			status.IsDLExpired = false
		}
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Onboarding details retrieved successfully",
		gin.H{"basic_details": basicDetailsAdded, "family_details": familyDetailsAdded,
			"health_details": healthDetailsAdded, "DigiLocker": status.DigiLockerPresent,
			"IsDLExpired": status.IsDLExpired, "GmailPresent": status.GmailPresent,
			"no_of_upcoming_appointments": noOfUpcomingAppointments, "no_of_medications_for_dashboard": noOfMedicationsForDashboard,
			"no_of_messages_for_dashboard": noOfMessagesForDashboard, "no_of_lab_reuslts_for_dashboard": noOfLabReusltsForDashboard,
		}, nil, nil)
	return
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
	nurses, totalRecords, err := pc.patientService.GetNursesList(nil, limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch nurses", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	nursesData := utils.MapUsersToSchema(nurses, "nurse")
	statusCode, message := utils.GetResponseStatusMessage(
		len(nurses),
		"Nurses list retrieved successfully",
		"Nurses not found",
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, nursesData, pagination, nil)
}

func (pc *PatientController) GetPharmacistList(c *gin.Context) {
	page, limit, offset := utils.GetPaginationParams(c)
	pharmacists, totalRecords, err := pc.patientService.GetPharmacistList(nil, limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch pharmacists", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	pharmacistsData := utils.MapUsersToSchema(pharmacists, "pharmacist")
	statusCode, message := utils.GetResponseStatusMessage(
		len(pharmacists),
		"Pharmacists list retrieved successfully",
		"Pharmacists not found",
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, pharmacistsData, pagination, nil)
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
	} else if appointment.ProviderType == "lab" {
		_, err := pc.diagnosticService.GetLabById(appointment.ProviderID)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to check diagnostic center existence", nil, err)
			return
		}

	} else {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid provider type", nil, errors.New("Invalid provider type"))
		return
	}

	if appointment.IsInperson == 0 {
		patientUser, _ := pc.patientService.GetUserProfileByUserId(appointment.PatientID)
		providerUser, _ := pc.patientService.GetUserProfileByUserId(appointment.ProviderID)
		rawInvitees := []map[string]string{
			{"name": patientUser.FirstName, "email": patientUser.Email},
			{"name": providerUser.FirstName, "email": providerUser.Email},
		}
		startTime := utils.ConvertToZoomTime(appointment.AppointmentDate.Format("2006-01-02"), appointment.AppointmentTime)

		zToken, _ := pc.userService.GetSingleTblUserToken(0, "ZOOM")
		expiresIn := 59 * time.Minute

		if time.Since(zToken.CreatedAt.UTC()) > expiresIn {
			res, err := service.GetRefreshedZoomAccessToken(zToken.RefreshToken)
			if err != nil {
				fmt.Println("Error while getting access token", err)
				models.ErrorResponse(ctx, constant.Failure, http.StatusServiceUnavailable, "Unable to schedule meeting please try again in sometime", nil, err)
				return
			}
			pc.userService.CreateTblUserToken(&models.TblUserToken{UserId: 0, AuthToken: res["access_token"].(string), Provider: "ZOOM", ProviderId: "catseyesystems", RefreshToken: res["refresh_token"].(string), CreatedAt: time.Now().UTC(), ExpiresAt: time.Now().UTC().Add(59 * time.Minute)})

			zoomRes, err := service.CreateZoomMeeting(res["access_token"].(string), "Online Doctor Consultation", appointment.AppointmentType, startTime, 30, rawInvitees)
			if err != nil {
				fmt.Println("Error while scheduling meeting:", err)
				models.ErrorResponse(ctx, constant.Failure, http.StatusServiceUnavailable, "Unable to schedule meeting please try again in sometime", nil, err)
				return
			}
			appointment.MeetingUrl = zoomRes.JoinURL
		} else {
			zoomRes, err := service.CreateZoomMeeting(zToken.AuthToken, "Online Doctor Consultation", appointment.AppointmentType, time.Now(), 30, rawInvitees)
			if err != nil {
				fmt.Println("Error while scheduling meeting:", err)
				models.ErrorResponse(ctx, constant.Failure, http.StatusServiceUnavailable, "Unable to schedule meeting please try again in sometime", nil, err)
				return
			}
			appointment.MeetingUrl = zoomRes.JoinURL
		}
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
	var providerInfo interface{}
	if appointment.ProviderType == "lab" {
		lab, _ := pc.diagnosticService.GetLabById(appointment.ProviderID)
		providerInfo = utils.MapUserToPublicProviderInfo(*lab, "lab")
	} else {
		user, _ := pc.patientService.GetUserProfileByUserId(createdAppointment.ProviderID)
		providerInfo = utils.MapUserToPublicProviderInfo(*user, createdAppointment.ProviderType)
	}
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
	user, err := pc.patientService.GetUserProfileByUserId(user_id)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "Failed to load profile", nil, err)
		return
	}
	userProfile := utils.MapUserToRoleSchema(*user, "patient")
	err = pc.emailService.SendAppointmentMail(appointmentResponse, userProfile)
	if err != nil {
		log.Println("Error sending appointment email:", err)
	}

	models.SuccessResponse(ctx, constant.Success, http.StatusCreated, "Appointment created scheduled", appointmentResponse, nil, nil)
	return
}

func (pc *PatientController) GetUserAppointments(ctx *gin.Context) {
	user_id, err := utils.GetUserIDFromContext(ctx, pc.patientService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	appointments, err := pc.appointmentService.GetUserAppointments(user_id)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to fetch appointments", nil, err)
		return
	}
	var responses []models.AppointmentResponse
	for _, appointment := range appointments {
		var providerInfo interface{}
		if appointment.ProviderType == "lab" {
			lab, _ := pc.diagnosticService.GetLabById(appointment.ProviderID)
			providerInfo = utils.MapUserToPublicProviderInfo(*lab, "lab")
		} else {
			user, _ := pc.patientService.GetUserProfileByUserId(appointment.ProviderID)
			providerInfo = utils.MapUserToPublicProviderInfo(*user, appointment.ProviderType)
		}
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
	_, message := utils.GetResponseStatusMessage(
		len(responses),
		"Appointments retrieved successfully",
		"Appointments not found",
	)
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, message, responses, nil, nil)
	return
}

func (pc *PatientController) UpdateUserAppointment(ctx *gin.Context) {
	var updateReq models.UpdateAppointmentRequest
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
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Recovered in UpdateAppointment:", r)
			models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to update appointment", nil, errors.New("Failed to update appointment"))
			return
		}
	}()

	existing, err := pc.appointmentService.FindAppointmentByID(tx, updateReq.AppointmentID)
	if err != nil {
		tx.Rollback()
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Appointment could not be updated", nil, err)
		return
	}
	if updateReq.UpdateType == 1 {
		existing.AppointmentDate = updateReq.AppointmentDate
		existing.AppointmentTime = updateReq.AppointmentTime
	}

	updated, err := pc.appointmentService.UpdateAppointmentByType(tx, existing, updateReq.UpdateType, user_id)
	if err != nil {
		tx.Rollback()
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Appointment could not be updated", nil, err)
		return
	}

	var providerInfo interface{}
	if existing.ProviderType == "lab" {
		lab, _ := pc.diagnosticService.GetLabById(existing.ProviderID)
		providerInfo = utils.MapUserToPublicProviderInfo(*lab, "lab")
	} else {
		user, _ := pc.patientService.GetUserProfileByUserId(existing.ProviderID)
		providerInfo = utils.MapUserToPublicProviderInfo(*user, existing.ProviderType)
	}

	appointmentResponse := models.AppointmentResponse{
		AppointmentID:   updated.AppointmentID,
		PatientID:       updated.PatientID,
		ProviderType:    updated.ProviderType,
		ProviderInfo:    providerInfo,
		ScheduledBy:     updated.ScheduledBy,
		AppointmentType: updated.AppointmentType,
		AppointmentDate: updated.AppointmentDate,
		AppointmentTime: updated.AppointmentTime,
		DurationMinutes: updated.DurationMinutes,
		IsInperson:      updated.IsInperson,
		Status:          updated.Status,
		MeetingUrl:      updated.MeetingUrl,
		PaymentStatus:   updated.PaymentStatus,
		Notes:           updated.Notes,
		PaymentID:       updated.PaymentID,
		CreatedAt:       updated.CreatedAt,
		UpdatedAt:       updated.UpdatedAt,
	}

	if err := tx.Commit().Error; err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to commit transaction", nil, err)
		return
	}

	var message string
	if updateReq.UpdateType == 1 {
		message = "Appointment rescheduled successfully"
	} else {
		message = "Appointment cancelled"
	}

	models.SuccessResponse(ctx, constant.Success, http.StatusOK, message, appointmentResponse, nil, nil)
	return
}

func (mc *PatientController) GetAllLabs(c *gin.Context) {
	_, exists := utils.GetUserDataContext(c)
	if !exists {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, constant.KeyCloakErrorMessage, nil, nil)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	data, totalRecords, err := mc.diagnosticService.GetAllLabs(limit, offset)
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

func (pc *PatientController) DigiLockerSyncController(ctx *gin.Context) {
	type UserRequest struct {
		Code        string `json:"code"`
		OnlyRefresh int    `json:"onlyRefresh"`
	}
	var req UserRequest

	sub, subExists := ctx.Get("sub")
	if !subExists {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "User not found", nil, errors.New("Error while getting profile"))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	user_id, err := pc.patientService.GetUserIdBySUB(sub.(string))
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "User can not be authorised", nil, err)
		return
	}
	digiTokenRes, err := service.GetDigiLockerToken(req.Code)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to get digilocker token", nil, err)
		return
	}
	if !strings.Contains(digiTokenRes["scope"].(string), "files.uploadeddocs") {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Please provide drive access", nil, errors.New("Invalid scope"))
		return
	}

	pc.userService.CreateTblUserToken(&models.TblUserToken{UserId: user_id, AuthToken: digiTokenRes["access_token"].(string), Provider: "DigiLocker", ProviderId: digiTokenRes["digilockerid"].(string), CreatedAt: time.Now().UTC()})
	if req.OnlyRefresh == 1 {
		models.SuccessResponse(ctx, constant.Success, http.StatusOK, "DigiLocker token refreshed successfully", digiTokenRes, nil, nil)
		return
	}

	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "DigiLocker sunc is in process you'll be notified once done", digiTokenRes, nil, nil)

	go func(userID uint64, token string, digiLockerId string) {
		log.Println("Starting DigiLocker directory & document sync in background...")

		dirsRes, err := service.GetDigiLockerDirs(token)
		if err != nil {
			log.Println("Error fetching dirs:", err)
			return
		}

		items, ok := dirsRes["items"].([]interface{})
		if !ok {
			log.Println("Items is not a list")
			return
		}

		var allDocs []models.TblMedicalRecord
		for _, item := range items {
			record, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			if record["type"] == "file" {
				newRecord := models.TblMedicalRecord{
					RecordName:     record["name"].(string),
					RecordSize:     utils.ParseIntField(record["size"].(string)),
					FileType:       record["mime"].(string),
					UploadSource:   "DigiLocker",
					SourceAccount:  digiLockerId,
					RecordCategory: "Report",
					Description:    record["description"].(string),
					UploadedBy:     userID,
					RecordUrl:      "https://digilocker.meripehchaan.gov.in/public/oauth2/1/file/" + record["uri"].(string),
					FetchedAt:      time.Now(),
					CreatedAt:      utils.ParseDateField(record["date"]),
				}
				allDocs = append(allDocs, newRecord)
			}

			if record["type"] == "dir" {
				subDocs, err := service.FetchDirItemsRecursively(token, record["id"].(string), digiLockerId, userID)
				if err != nil {
					log.Println("Error in subdirectory:", err)
					continue
				}
				allDocs = append(allDocs, subDocs...)
			}
		}
		log.Printf("Total documents collected: %d %v", len(allDocs), allDocs)
		err = pc.medicalRecordService.SaveMedicalRecords(&allDocs, userID)
		if err != nil {
			log.Println("Error occurend while saving medical records from digilocker for", userID, digiLockerId, err)
		}
		log.Println("Successfully saved medical records from DigiLocker for user ID:", userID)

	}(user_id, digiTokenRes["access_token"].(string), digiTokenRes["digilockerid"].(string))

}

func (pc *PatientController) ReadDigiLockerFile(ctx *gin.Context) {
	sub, subExists := ctx.Get("sub")
	if !subExists {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "User not found", nil, errors.New("Error while reading document"))
		return
	}

	user_id, err := pc.patientService.GetUserIdBySUB(sub.(string))
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "User can not be authorised", nil, err)
		return
	}

	userDigiToken, err := pc.userService.GetSingleTblUserToken(user_id, "DigiLocker")
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Digilocker Token not found", nil, err)
		return
	}

	type UserRequest struct {
		ResourceUrl string `json:"resource_url"`
	}
	var req UserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}

	response, err := service.ReadDigiLockerFile(userDigiToken.AuthToken, req.ResourceUrl)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Error while reading file", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Resource loaded", response, nil, nil)
}

func (pc *PatientController) SaveReport(ctx *gin.Context) {
	authUserId, exists := utils.GetUserDataContext(ctx)
	if !exists {
		models.ErrorResponse(ctx, constant.Failure, http.StatusNotFound, constant.KeyCloakErrorMessage, nil, nil)
		return
	}
	patientId, err := pc.patientService.GetUserIdByAuthUserId(authUserId)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusNotFound, "Patient not found", nil, err)
		return
	}
	log.Printf("Request Content-Type: %v", ctx.Request.Header.Get("Content-Type"))
	log.Printf("Request Headers: %v", ctx.Request.Header)
	file, fileHeader, err := ctx.Request.FormFile("file")
	if err != nil {
		log.Printf("No image attached or failed to read image: %v", err)
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Failed to get image from request", nil, err)
		return
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}

	if !allowedTypes[contentType] {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Unsupported file type. Only JPEG, PNG, and PDF are allowed", nil, nil)
		return
	}

	reportData, err := pc.apiService.CallGeminiService(file, fileHeader.Filename)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to call ai service", nil, err)
		return
	}
	message, err := pc.diagnosticService.DigitizeDiagnosticReport(reportData, patientId)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, message, nil, err)
		return
	}
	go func(pid uint64) {
		if err := pc.diagnosticService.NotifyAbnormalResult(pid); err != nil {
			log.Printf("NotifyAbnormalResult error: %v", err)
		}
	}(patientId)

	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Report created successfully", nil, nil, nil)
}

func (pc *PatientController) SaveUserHealthProfile(ctx *gin.Context) {
	sub, subExists := ctx.Get("sub")
	if !subExists {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "Authentication token is missing or invalid", nil, errors.New("unable to retrieve user 'sub' from context"))
		return
	}
	var userReq models.PatientHealthProfileRequest
	if err := ctx.ShouldBindJSON(&userReq); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request format: please check the data you're sending", nil, err)
		return
	}

	user_id, err := pc.patientService.GetUserIdBySUB(sub.(string))
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Failed to authenticate user: user not found or unauthorized", nil, err)
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to initiate transaction", nil, tx.Error)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Recovered in SaveHealthProfile:", r)
			models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to save health details", nil, errors.New("Failed to save health details"))
			return
		}
	}()

	healthData := &models.TblPatientHealthProfile{
		PatientID:             user_id,
		HeightCM:              userReq.HeightCm,
		WeightKG:              userReq.WeightKg,
		BloodType:             userReq.BloodType,
		SmokingStatus:         userReq.SmokingStatus,
		AlcoholConsumption:    userReq.AlcoholConsumption,
		PhysicalActivityLevel: userReq.PhysicalActivityLevel,
		DietaryPreferences:    userReq.DietaryPreferences,
		ExistingConditions:    userReq.ExistingConditions,
		FamilyMedicalHistory:  userReq.FamilyMedicalHistory,
		MenstrualHistory:      userReq.MenstrualHistory,
		Notes:                 userReq.Notes,
	}

	savedHealthProfile, err := pc.patientService.SaveUserHealthProfile(tx, healthData)
	if err != nil {
		tx.Rollback()
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Failed to save health profile", nil, err)
		return
	}
	for _, allergy := range userReq.Allergies {
		err := pc.allergyService.AddPatientAllergyRestriction(tx, &models.PatientAllergyRestriction{
			PatientId:  uint(user_id),
			AllergyId:  uint(allergy.AllergyID),
			SeverityId: uint(allergy.SeverityId),
		})
		if err != nil {
			tx.Rollback()
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Failed to save allergies", nil, err)
			return
		}
	}
	_, err = pc.patientService.AddPatientDiseaseProfile(tx, &models.PatientDiseaseProfile{PatientId: user_id, DiseaseProfileId: userReq.DiseaseProfileID})
	if err != nil {
		tx.Rollback()
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Failed to save disease profile", nil, err)
		return
	}
	if err := tx.Commit().Error; err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to commit transaction", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Health data saved successfully", savedHealthProfile, nil, nil)
	return
}

func (pc *PatientController) GetDiseaseProfiles(ctx *gin.Context) {
	var diseaseProfiles []models.DiseaseProfile
	var totalRecords int64
	page, limit, offset := utils.GetPaginationParams(ctx)

	diseaseProfiles, totalRecords, err := pc.diseaseService.GetDynamicDiseaseProfiles(limit, offset, []string{"Disease"})
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to retrieve disease profile", nil, err)
		return
	}

	summaryDiseaseProfiles := utils.ToDiseaseProfileSummaryDTOs(diseaseProfiles)
	pagination := utils.GetPagination(limit, page, offset, totalRecords)

	message := "Diseases profile not found"
	if len(diseaseProfiles) > 0 {
		message = "Diseases profile info retrieved successfully"
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, message, summaryDiseaseProfiles, pagination, nil)
}

func (pc *PatientController) AttachDiseaseProfileTOPatient(ctx *gin.Context) {
	authUserId, exists := utils.GetUserDataContext(ctx)
	if !exists {
		models.ErrorResponse(ctx, constant.Failure, http.StatusNotFound, constant.KeyCloakErrorMessage, nil, nil)
		return
	}
	user_id, err := pc.patientService.GetUserIdBySUB(authUserId)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Failed to authenticate user: user not found or unauthorized", nil, err)
		return
	}

	type UserRequest struct {
		DiseaseProfileID uint64 `json:"disease_profile_id"`
	}
	var userReq UserRequest
	if err := ctx.ShouldBindJSON(&userReq); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to initiate transaction", nil, tx.Error)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Recovered in SaveHealthProfile:", r)
			models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to save health details", nil, errors.New("Failed to save health details"))
			return
		}
	}()

	_, err = pc.diseaseService.GetDiseaseProfileById(strconv.FormatUint(userReq.DiseaseProfileID, 10))
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusNotFound, "Disease profile not found", nil, err)
		return
	}

	diseaseProfile, err := pc.patientService.AddPatientDiseaseProfile(tx, &models.PatientDiseaseProfile{PatientId: user_id, DiseaseProfileId: userReq.DiseaseProfileID, AttachedFlag: 0})
	if err != nil {
		tx.Rollback()
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Failed to save disease profile", nil, err)
		return
	}
	if err := tx.Commit().Error; err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to commit transaction", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Disease Profile attached", diseaseProfile, nil, nil)
}

func (pc *PatientController) UpdateDiseaseProfile(ctx *gin.Context) {
	authUserId, exists := utils.GetUserDataContext(ctx)
	if !exists {
		models.ErrorResponse(ctx, constant.Failure, http.StatusNotFound, constant.KeyCloakErrorMessage, nil, nil)
		return
	}

	user_id, err := pc.patientService.GetUserIdBySUB(authUserId)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Failed to authenticate user: user not found or unauthorized", nil, err)
		return
	}
	var req models.DPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	err = pc.patientService.UpdateFlag(user_id, &req)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Failed to detach disease profile", nil, err)
		return
	}

	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Disease profile updated", nil, nil, nil)
}

var shortURLMap = make(map[string]string)

func generateShortCode() string {
	b := make([]byte, 6)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:6]
}

type SendSMSRequest struct {
	MobileNo string `json:"mobile_no" binding:"required"`
}

func (pc *PatientController) SendSMS(c *gin.Context) {
	_, exists := utils.GetUserDataContext(c)
	if !exists {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, constant.KeyCloakErrorMessage, nil, nil)
		return
	}
	var req SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	authHeader := c.GetHeader("Authorization") // Bearer <token>
	token := strings.TrimPrefix(authHeader, "Bearer ")
	exchanged, err := auth.ExchangeToken(token)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Token exchange failed", nil, err)
		return
	}
	longLink := fmt.Sprintf("http://localhost:3000/shared-report?token=%s", exchanged.AccessToken)
	shortCode := generateShortCode()
	shortURLMap[shortCode] = longLink
	shortURL := fmt.Sprintf("http://localhost:5500/v1/user/r/%s", shortCode)
	err1 := pc.smsService.SendSMS(req.MobileNo, shortURL)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to send SMS", nil, err1)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "SMS sent successfully", exchanged, nil, nil)
}

type SendEmailRequest struct {
	Email []string `json:"email" binding:"required"`
}

func (pc *PatientController) ShareReport(c *gin.Context) {
	authUserId, exists := utils.GetUserDataContext(c)
	if !exists {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, constant.KeyCloakErrorMessage, nil, nil)
		return
	}
	var req SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	patientId, err := pc.patientService.GetUserIdByAuthUserId(authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient not found", nil, err)
		return
	}
	userDetails, err := pc.patientService.GetUserProfileByUserId(patientId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Failed to load profile", nil, err)
		return
	}

	authHeader := c.GetHeader("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	// exchanged, err := auth.ExchangeToken(token)
	// if err != nil {
	// 	models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Token exchange failed", nil, err)
	// 	return
	// }

	baseURL := os.Getenv("SHARE_REPORT_BASE_URL")
	if baseURL == "" {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Missing base URL in environment variables", nil, nil)
		return
	}

	longLink := fmt.Sprintf("%s/shared-report?token=%s", baseURL, token)
	shortCode := generateShortCode()
	shortURLMap[shortCode] = longLink
	shortURL := fmt.Sprintf("%s/v1/user/r/%s", os.Getenv("SHORT_URL_BASE"), shortCode)

	err1 := pc.emailService.ShareReportEmail(req.Email, userDetails.FirstName+" "+userDetails.LastName, shortURL)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to send email", nil, err1)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Email sent successfully", nil, nil, nil)
}
