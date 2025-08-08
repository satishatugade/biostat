package controller

import (
	"biostat/auth"
	"biostat/config"
	"biostat/constant"
	"biostat/database"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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
	emailService         service.EmailService
	orderService         service.OrderService
	notificationService  service.NotificationService
	authService          auth.AuthService
	roleService          service.RoleService
	permissionService    service.PermissionService
	subscriptionService  service.SubscriptionService
	processStatusService service.ProcessStatusService
	gmailSyncService     service.GmailSyncService
}

func NewPatientController(patientService service.PatientService, dietService service.DietService,
	allergyService service.AllergyService, medicalRecordService service.TblMedicalRecordService,
	medicationService service.MedicationService, appointmentService service.AppointmentService,
	diagnosticService service.DiagnosticService, userService service.UserService,
	apiService service.ApiService, diseaseService service.DiseaseService, smsService service.SmsService,
	emailService service.EmailService, orderService service.OrderService, notificationService service.NotificationService,
	authService auth.AuthService, roleService service.RoleService, permissionService service.PermissionService,
	subscriptionService service.SubscriptionService, processStatusService service.ProcessStatusService, gmailSyncService service.GmailSyncService) *PatientController {
	return &PatientController{patientService: patientService, dietService: dietService,
		allergyService: allergyService, medicalRecordService: medicalRecordService,
		medicationService: medicationService, appointmentService: appointmentService,
		diagnosticService:    diagnosticService,
		userService:          userService,
		apiService:           apiService,
		diseaseService:       diseaseService,
		smsService:           smsService,
		emailService:         emailService,
		orderService:         orderService,
		notificationService:  notificationService,
		authService:          authService,
		roleService:          roleService,
		permissionService:    permissionService,
		subscriptionService:  subscriptionService,
		processStatusService: processStatusService,
		gmailSyncService:     gmailSyncService,
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

	relation, err := pc.patientService.GetRelationById(relationId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Relation not found", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Relation fetched successfully", relation, nil, nil)
}

func (pc *PatientController) GetGenderById(c *gin.Context) {
	genderIdStr := c.Param("gender_id")
	genderId, err := strconv.ParseUint(genderIdStr, 10, 32)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid gender ID", nil, err)
		return
	}

	gender, err := pc.patientService.GetGenderById(genderId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Gender not found", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Gender fetched successfully", gender, nil, nil)
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
	sub, userId, isDelegate, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	log.Println("isDelegate : userid", isDelegate, userId)
	if isDelegate {
		reqUserID, err := pc.userService.GetUserIdBySUB(sub)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, err.Error(), nil, err)
			return
		}
		err = pc.patientService.CanContinue(userId, reqUserID, constant.PermissionEditProfile)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, string(constant.PermissionEditInfo), nil, err)
			return
		}
	}

	var patientData models.Patient
	if err := c.ShouldBindJSON(&patientData); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input data", nil, err)
		return
	}

	user, err := pc.patientService.GetUserProfileByUserId(userId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "failed to get user", nil, err)
		return
	}
	user.Email = patientData.Email
	user.MobileNo = patientData.MobileNo
	user.FirstName = patientData.FirstName
	user.LastName = patientData.LastName

	userUpdateErr := pc.authService.UpdateUserInKeycloak(*user)
	if userUpdateErr != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "failed to update user in keycloak", nil, userUpdateErr)
		return
	}
	updatedPatient, err := pc.patientService.UpdatePatientById(userId, &patientData)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update patient info", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient info updated successfully", updatedPatient, nil, nil)
}

func (pc *PatientController) GetPatientDiseaseProfiles(c *gin.Context) {
	_, user_id, _, err1 := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
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
	_, user_id, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
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
			models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Oops! The AI service is temporarily down. Please try again later or contact support team.", nil, err1)
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

func (pc *PatientController) SummarizeHistorybyAIModel(c *gin.Context) {
	_, patientId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	summarize_history, err := pc.patientService.SummarizeHistorybyAIModel(patientId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Oops! The AI service is temporarily down. Please try again later or contact support team.", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient summary fetched.", summarize_history, nil, nil)
}

func (pc *PatientController) GetPatientDiagnosticTrendValue(c *gin.Context) {
	_, userId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var req models.DiagnosticResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("DiagnosticResultRequest filter result values")
	}
	req.PatientId = userId
	results, err := pc.patientService.GetPatientDiagnosticTrendValue(req)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Patient reports not found", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Historical trend fetched successfully", results, nil, nil)
}

func (pc *PatientController) GetDiagnosticResults(c *gin.Context) {
	_, user_id, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	user, err := pc.patientService.GetUserProfileByUserId(user_id)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Failed to load profile", nil, err)
		return
	}

	var filter models.DiagnosticReportFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// return
	}

	results, err := pc.patientService.FetchPatientDiagnosticReports(user.UserId, filter)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Failed to fetch diagnostic results", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "report load successfully", results, nil, nil)
}

func (pc *PatientController) GetPatientDiagnosticReportResult(c *gin.Context) {
	_, user_id, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	user, err := pc.patientService.GetUserProfileByUserId(user_id)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Failed to load profile", nil, err)
		return
	}

	var filter models.DiagnosticReportFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		// return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	results, totalRecords, err := pc.patientService.GetPatientDiagnosticReportResult(user.UserId, filter, limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Failed to fetch diagnostic results", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Report grid view load successfully", results, pagination, nil)
}

func (pc *PatientController) ExportDiagnosticResultsExcel(c *gin.Context) {
	_, user_id, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	user, err := pc.patientService.GetUserProfileByUserId(user_id)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Failed to load profile", nil, err)
		return
	}

	var filter models.DiagnosticReportFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		// models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid filter input", nil, err)
		// return
	}

	data, _, err := pc.patientService.GetPatientDiagnosticReportResult(user.UserId, filter, 1000, 0)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to generate report data", nil, err)
		return
	}
	fileBytes, err := pc.patientService.GenerateExcelFile(data)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to generate Excel", nil, err)
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="diagnostic_report.xlsx"`)
	c.Header("File-Name", "diagnostic_report.xlsx")
	c.Header("Access-Control-Expose-Headers", "File-Name")
	c.Data(http.StatusOK, "application/octet-stream", fileBytes)
}

func (pc *PatientController) ExportDiagnosticResultsPDF(c *gin.Context) {
	_, user_id, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	user, err := pc.patientService.GetUserProfileByUserId(user_id)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Failed to load profile", nil, err)
		return
	}
	fmt.Println("user ", user)
	var filter models.DiagnosticReportFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid filter input", nil, err)
		return
	}

	// data, _, err := pc.patientService.GetPatientDiagnosticReportResult(user.UserId, filter, 1000, 0)
	// if err != nil {
	// 	models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to generate report data", nil, err)
	// 	return
	// }

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	reportPath := filepath.Join(dir, "sample.json")

	reportBytes, err := ioutil.ReadFile(reportPath)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, err.Error(), nil, err)
		return
	}

	// Unmarshal the JSON content into the appropriate struct (assuming it's map[string]interface{})
	var reportData models.ReportData
	err = json.Unmarshal(reportBytes, &reportData)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, err.Error(), nil, err)
		return
	}

	pdfBytes, err := pc.patientService.GeneratePDF(reportData)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to generate PDF", nil, err)
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `attachment; filename="diagnostic_report.pdf"`)
	c.Header("File-Name", "diagnostic_report.pdf")
	c.Header("Access-Control-Expose-Headers", "File-Name")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func (pc *PatientController) AddPrescription(c *gin.Context) {
	authUserId, userId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var prescription models.PatientPrescription
	prescription.PatientId = userId
	if err := c.ShouldBindJSON(&prescription); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid patient prescription input data", nil, err)
		return
	}
	canAccess := pc.patientService.CanAccessAPI(userId, []string{string(constant.MappingTypeR), string(constant.MappingTypeHOF), string(constant.MappingTypeS)})
	if !canAccess {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, string(constant.PermissionUploadMedicalRecord), nil, errors.New("You need subscription for uploading own records"))
		return
	}
	err1 := pc.patientService.AddPatientPrescription(authUserId, &prescription)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add patient prescription", nil, err1)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient prescription added.", prescription, nil, nil)
}

func (pc *PatientController) ArchivePrescription(c *gin.Context) {
	_, userId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	type UserRequest struct {
		PrescriptionID uint64 `json:"prescription_id"`
	}
	fmt.Println("Got Request")
	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid prescription input data", nil, err)
		return
	}
	err1 := pc.patientService.ArchivePatientPrescription(userId, req.PrescriptionID)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to archive prescription", nil, err1)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "prescription archived", nil, nil, nil)
	return
}

func (pc *PatientController) UpdatePrescription(c *gin.Context) {
	authUserId, user_id, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	var prescription models.PatientPrescription
	if err := c.ShouldBindJSON(&prescription); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input data", nil, err)
		return
	}

	if prescription.PrescriptionId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Prescription ID is required", nil, nil)
		return
	}

	prescription.PatientId = user_id
	err = pc.patientService.UpdatePatientPrescription(authUserId, &prescription)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update prescription", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient prescription updated.", prescription, nil, nil)
}

func (pc *PatientController) GetPrescriptionByPatientId(c *gin.Context) {
	_, patientId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)

	prescriptions, totalRecords, err := pc.patientService.GetPrescriptionByPatientId(patientId, limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve prescriptions", nil, err)
		return
	}

	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	_, message := utils.GetResponseStatusMessage(
		len(prescriptions),
		"Prescription info retrieved successfully",
		"Prescription info not found",
	)

	models.SuccessResponse(c, constant.Success, http.StatusOK, message, prescriptions, pagination, nil)
}

func (pc *PatientController) GetPrescriptionDetailByPatientId(c *gin.Context) {
	_, patientId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)

	prescription, totalRecords, err := pc.patientService.GetPrescriptionDetailByPatientId(patientId, limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Prescription detail not found", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	statusCode, message := utils.GetResponseStatusMessage(
		1,
		"Prescription detail retrieved successfully",
		"Prescription detail not found",
	)

	models.SuccessResponse(c, constant.Success, statusCode, message, prescription, pagination, nil)
}

func (pc *PatientController) PrescriptionInfobyAIModel(c *gin.Context) {
	_, patientId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var request struct {
		PrescriptionId uint64 `json:"prescription_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input data", nil, err)
		return
	}
	prescriptionSummary, err := pc.patientService.GetPrescriptionInfo(request.PrescriptionId, patientId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Oops! The AI service is temporarily down. Please try again later or contact support team.", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Prescription explanation fetched.", prescriptionSummary, nil, nil)
}

func (pc *PatientController) PharmacokineticsInfobyAIModel(c *gin.Context) {
	_, patientId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	var request struct {
		PrescriptionId uint64 `json:"prescription_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input data", nil, err)
		return
	}

	pkSummary, err := pc.patientService.GetPharmacokineticsInfo(request.PrescriptionId, patientId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Oops! The AI service is temporarily down. Please try again later or contact support team.", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Pharmacokinetics data fetched.", pkSummary, nil, nil)
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

//TODO DEL V
// func (pc *PatientController) GetPatientRelative(c *gin.Context) {
// 	patientId := c.Param("patient_id")

// 	relatives, err := pc.patientService.GetPatientRelative(patientId)
// 	if err != nil {
// 		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch patient relatives", nil, err)
// 		return
// 	}
// 	statusCode, message := utils.GetResponseStatusMessage(
// 		len(relatives),
// 		"Patient relatives retrieved successfully",
// 		"Relatives not found",
// 	)

// 	models.SuccessResponse(c, constant.Success, statusCode, message, relatives, nil, nil)
// }

func (pc *PatientController) GetPatientRelativeList(c *gin.Context) {
	_, user_id, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
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

func (pc *PatientController) AssignPrimaryCaregiver(c *gin.Context) {
	sub, patientId, isDelegate, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	log.Printf("sub : %s patientId: %d isDelegate : %t", sub, patientId, isDelegate)
	if isDelegate {
		log.Printf(" Inside isDelegate true sub : %s patientId: %d isDelegate : %t", sub, patientId, isDelegate)
		reqUserID, err := pc.userService.GetUserIdBySUB(sub)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, err.Error(), nil, err)
			return
		}
		err = pc.patientService.CanContinue(patientId, reqUserID, constant.PermissionViewHealth)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, string(constant.PermissionHOFAssignUnassign), nil, err)
			return
		}
	}
	// hasHOF, err := pc.roleService.HasHOFMapping(patientId, string(constant.MappingTypeHOF))
	// if !hasHOF {
	// 	models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, " Only the Head of Family can assign or unassign a new Head of Family", nil, err)
	// 	return
	// }
	relativeIdStr := c.Query("relative_id")
	mappingType := c.Query("mapping_type")
	relativeId, err := strconv.ParseUint(relativeIdStr, 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid relative ID", nil, err)
		return
	}
	if mappingType != string(constant.MappingTypeR) && mappingType != string(constant.MappingTypeHOF) {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Only 'Relative' or 'Primary Caregiver' roles can be assigned.", nil, nil)
		return
	}
	err = pc.patientService.AssignPrimaryCaregiver(patientId, relativeId, mappingType)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, err.Error(), nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "caregiver role assign successfully", nil, nil, nil)
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
	_, patientId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
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

func (pc *PatientController) GetPatientCaregiverList(c *gin.Context) {
	_, patientId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	caregivers, err := pc.patientService.GetCaregiverList(&patientId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch patient caregivers", nil, err)
		return
	}
	_, message := utils.GetResponseStatusMessage(
		len(caregivers),
		"Patient caregiver retrieved successfully",
		"Caregiver not found",
	)

	models.SuccessResponse(c, constant.Success, http.StatusOK, message, caregivers, nil, nil)
}

func (pc *PatientController) GetAssignedPatientList(c *gin.Context) {
	_, caregiverId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	patients, err := pc.patientService.GetAssignedPatientList(&caregiverId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch assigned patients", nil, err)
		return
	}

	_, message := utils.GetResponseStatusMessage(
		len(patients),
		"Assigned patient list retrieved successfully",
		"No patients assigned",
	)

	models.SuccessResponse(c, constant.Success, http.StatusOK, message, patients, nil, nil)
}

func (pc *PatientController) SetPatientUserDeletedMappingStatus(c *gin.Context) {
	_, patientId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	userIdStr := c.Query("user_id")
	if userIdStr == "" {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Missing user_id in query parameter", nil, nil)
		return
	}
	mapping_type := c.Query("mapping_type")
	if mapping_type == "" {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Missing mapping_type in query parameter", nil, nil)
		return
	}

	if !utils.IsValidMappingType(mapping_type) {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid mapping_type", nil, nil)
		return
	}

	isDeletedStr := c.DefaultQuery("is_deleted", "1")
	isDeletedInt, err := strconv.Atoi(isDeletedStr)
	if err != nil || (isDeletedInt != 0 && isDeletedInt != 1) {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid is_deleted value. Use 1 or 0.", nil, err)
		return
	}
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid user id ", nil, err)
		return
	}

	err = pc.patientService.SetPatientUserDeletedMappingStatus(patientId, userId, isDeletedInt, mapping_type)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "User not found", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "User removed successfully", nil, nil, nil)
}

func (pc *PatientController) GetCaregiverList(c *gin.Context) {

	caregivers, err := pc.patientService.GetCaregiverList(nil)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch patient caregivers", nil, err)
		return
	}
	_, message := utils.GetResponseStatusMessage(
		len(caregivers),
		"Caregiver list retrieved successfully",
		"Caregiver not found",
	)

	models.SuccessResponse(c, constant.Success, http.StatusNoContent, message, caregivers, nil, nil)
}

func (pc *PatientController) GetDoctorList(c *gin.Context) {
	User := c.Param("user")
	_, user_id, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var filterUserId *uint64
	if User == "patient" {
		filterUserId = &user_id
	}
	page, limit, offset := utils.GetPaginationParams(c)
	doctors, totalRecords, err := pc.patientService.GetDoctorList(filterUserId, User, limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch patient doctor", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	doctorProfile := utils.MapUsersToSchema(doctors, "doctor")
	_, message := utils.GetResponseStatusMessage(
		len(doctors),
		"Doctor list retrieved successfully",
		"Doctor not found",
	)

	models.SuccessResponse(c, constant.Success, http.StatusOK, message, doctorProfile, pagination, nil)
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
	_, patientId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
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
	_, user_id, _, err := utils.GetUserIDFromContext(ctx, c.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	records, err := c.medicalRecordService.GetUserMedicalRecords(user_id)
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

func (c *PatientController) GetAllMedicalRecord(ctx *gin.Context) {
	sub, patientId, isDelegate, err := utils.GetUserIDFromContext(ctx, c.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	if isDelegate {
		reqUserID, err := c.userService.GetUserIdBySUB(sub)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, err.Error(), nil, err)
			return
		}
		err = c.patientService.CanContinue(patientId, reqUserID, constant.PermissionViewHealth)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, string(constant.PermissionViewMedicalRecord), nil, err)
			return
		}
	}
	page, limit, offset := utils.GetPaginationParams(ctx)
	isDeleted, queryParamErr := strconv.Atoi(ctx.DefaultQuery("is_deleted", "0"))
	if queryParamErr != nil {
		isDeleted = 0
	}
	data, total, err := c.medicalRecordService.GetMedicalRecords(patientId, limit, offset, isDeleted)
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

func (c *PatientController) GetDiagnosticLabReportName(ctx *gin.Context) {
	_, patientId, _, err := utils.GetUserIDFromContext(ctx, c.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	data, err := c.diagnosticService.GetDiagnosticLabReportName(patientId)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to retrieve records", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Lab report name fetched.", data, nil, nil)
}

func (pc *PatientController) CreateTblMedicalRecord(ctx *gin.Context) {
	authUserId, userId, isDelegate, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	if isDelegate {
		reqUserID, err := pc.userService.GetUserIdBySUB(authUserId)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, err.Error(), nil, err)
			return
		}
		err = pc.patientService.CanContinue(userId, reqUserID, constant.PermissionUploadReport)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, string(constant.PermissionUploadMedicalRecord), nil, err)
			return
		}
	} else {
		canAccess := pc.patientService.CanAccessAPI(userId, []string{string(constant.MappingTypeR), string(constant.MappingTypeHOF), string(constant.MappingTypeS)})
		if !canAccess {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, string(constant.PermissionUploadMedicalRecord), nil, errors.New("You need subscription for uploading own records"))
			return
		}
	}
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "Please provid file to save", nil, errors.New("Error while uploading document"))
		return
	}
	uploadSource := ctx.PostForm("upload_source")
	description := ctx.PostForm("description")
	recordCategory := ctx.PostForm("record_category")
	log.Println("record category : ", recordCategory)
	data, err := pc.medicalRecordService.CreateTblMedicalRecord(userId, authUserId, file, header, uploadSource, description, recordCategory)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to create record", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Your record has been created successfully. Digitization is in progress and should complete within 4–5 minutes.", data, nil, nil)
}

func (pc *PatientController) SaveReport(ctx *gin.Context) {
	authUserId, patientId, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	idParam := ctx.Param("record_id")
	recordId, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid record Id", nil, err)
		return
	}

	record, err := pc.medicalRecordService.GetMedicalRecordByRecordId(recordId)
	if err != nil || record == nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusNotFound, "Medical record not found", nil, err)
		return
	}
	userInfo, err := pc.userService.GetSystemUserInfoByAuthUserId(authUserId)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "User info fetch failed", nil, err)
		return
	}
	if record.RecordUrl == "" {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "No file found for this record", nil, nil)
		return
	}
	resp, err := http.Get(record.RecordUrl)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to fetch file from URL", nil, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "File not accessible or missing", nil, nil)
		return
	}

	fileBuf := new(bytes.Buffer)
	_, err = io.Copy(fileBuf, resp.Body)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Error reading file content", nil, err)
		return
	}

	_, updateErr := pc.medicalRecordService.UpdateTblMedicalRecord(&models.TblMedicalRecord{RecordId: recordId, Status: constant.StatusRetrying})
	if updateErr != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to update record", nil, updateErr)
		return
	}
	filename := filepath.Base(record.RecordUrl)
	record.RecordCategory = string(constant.TESTREPORT)
	if record.RecordCategory == "report" {
		record.RecordCategory = string(constant.TESTREPORT)
	}
	processID := uuid.New()
	pc.processStatusService.LogStep(processID, string(constant.RetryStep), constant.Running, "Trying again to digitize", "", nil, nil, nil, nil, nil, nil)
	err = pc.medicalRecordService.CreateDigitizationTask(record, userInfo, patientId, fileBuf, filename, processID, nil)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to queue digitization", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Digitization queued successfully", nil, nil, nil)
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
	// updatedBy := ctx.GetString("user")
	data, err := c.medicalRecordService.UpdateTblMedicalRecord(&payload)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to update record", nil, nil)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Record updated successfully", data, nil, nil)
}

func (c *PatientController) GetMedicalRecordByRecordId(ctx *gin.Context) {
	recordId := utils.GetParamAsInt(ctx, "id")
	if recordId == 0 {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Param id is required", nil, nil)
		return
	}
	data, err := c.medicalRecordService.GetMedicalRecordByRecordId(uint64(recordId))
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
	sub, user_id, isDelegate, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	type UserRequest struct {
		User string `json:"user"`
	}
	var req UserRequest
	roles, rolesExists := ctx.Get("userRoles")
	if !rolesExists {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "User not found", nil, errors.New("error while getting profile"))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	if !utils.StringInSlice(req.User, roles.([]string)) {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid role", nil, nil)
		return
	}
	if isDelegate {
		reqUserID, err := pc.userService.GetUserIdBySUB(sub)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, err.Error(), nil, err)
			return
		}
		err = pc.patientService.CanContinue(user_id, reqUserID, constant.PermissionViewHealth)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, string(constant.PermissionViewProfile), nil, err)
			return
		}
		log.Println("Its Delegated Request with UserId:", user_id, " by User with Id:", reqUserID)
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
	_, user_id, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
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

func (pc *PatientController) GetSources(c *gin.Context) {
	_, user_id, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)

	sources, totalRecords, err := pc.diagnosticService.GetSources(user_id, limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve health vital sources", nil, err)
		return
	}

	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	statusCode, message := utils.GetResponseStatusMessage(
		len(sources),
		"Health vital sources retrieved successfully",
		"Health vital sources not found",
	)

	models.SuccessResponse(c, constant.Success, statusCode, message, sources, pagination, nil)
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
	sub, user_id, isDelegate, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var appointment models.Appointment
	if err := ctx.ShouldBindJSON(&appointment); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}

	if isDelegate {
		reqUserID, err := pc.userService.GetUserIdBySUB(sub)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, err.Error(), nil, err)
			return
		}
		err = pc.patientService.CanContinue(user_id, reqUserID, constant.PermissionScheduleAppointments)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, string(constant.PermissionScheduleAppointment), nil, err)
			return
		}
		log.Println("Its Delegated Request with UserId:", user_id, " by User with Id:", reqUserID)
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
		// providerUser, _ := pc.patientService.GetUserProfileByUserId(appointment.ProviderID)
		rawInvitees := []map[string]string{
			{"name": patientUser.FirstName, "email": patientUser.Email},
			// {"name": providerUser.FirstName, "email": providerUser.Email},
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
			log.Printf("Recovered in CreateAppointment: %v\nStack Trace:\n%s", r, debug.Stack())
			models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to schedule appointment", nil, errors.New("Failed to schedule appointment"))
			return
		}
	}()

	createdAppointment, err := pc.appointmentService.CreateAppointment(tx, &appointment)
	if err != nil {
		log.Println("@Sch:", err)
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
	user, err := pc.patientService.GetUserProfileByUserId(appointment.PatientID)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, "Failed to load profile", nil, err)
		return
	}
	log.Println("Scheduled")
	userProfile := utils.MapSystemUserToPatient(user)
	mailErr := pc.emailService.SendAppointmentMail(appointmentResponse, *userProfile, providerInfo)
	if mailErr != nil {
		log.Println("Error sending appointment email:", mailErr)
	}

	models.SuccessResponse(ctx, constant.Success, http.StatusCreated, "Appointment created scheduled", appointmentResponse, nil, nil)
	return
}

func (pc *PatientController) GetUserAppointments(ctx *gin.Context) {
	sub, user_id, isDelegate, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	if isDelegate {
		reqUserID, err := pc.userService.GetUserIdBySUB(sub)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, err.Error(), nil, err)
			return
		}
		err = pc.patientService.CanContinue(user_id, reqUserID, constant.PermissionScheduleAppointments)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, string(constant.PermissionViewAppointment), nil, err)
			return
		}
		log.Println("Its Delegated Request with UserId:", user_id, " by User with Id:", reqUserID)
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
			lab, err := pc.diagnosticService.GetLabById(appointment.ProviderID)
			if err != nil || lab == nil {
				log.Println("Error @GetUserAppointments->GetLabById err:", err, " lab:", lab)
				continue
			}
			providerInfo = utils.MapUserToPublicProviderInfo(*lab, "lab")
		} else {
			user, err := pc.patientService.GetUserProfileByUserId(appointment.ProviderID)
			if err != nil || user == nil {
				log.Println("Error @GetUserAppointments->GetUserProfileByUserId err:", err, " user:", user)
				continue
			}
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
	sub, user_id, isDelegate, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}

	if isDelegate {
		reqUserID, err := pc.userService.GetUserIdBySUB(sub)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, err.Error(), nil, err)
			return
		}
		err = pc.patientService.CanContinue(user_id, reqUserID, constant.PermissionScheduleAppointments)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, string(constant.PermissionRescheduleAppointment), nil, err)
			return
		}
		log.Println("Its Delegated Request with UserId:", user_id, " by User with Id:", reqUserID)
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

func (mc *PatientController) AddLab(c *gin.Context) {
	authUserId, userId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var lab models.AddLabRequest
	if err := c.ShouldBindJSON(&lab); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input", nil, err)
		return
	}
	canAccess := mc.patientService.CanAccessAPI(userId, []string{string(constant.MappingTypeR), string(constant.MappingTypeHOF), string(constant.MappingTypeS)})
	if !canAccess {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Please get subscription inorder to perform this action", nil, errors.New("You need subscription for adding labs"))
		return
	}
	labInfo, err := mc.diagnosticService.CreateLab(userId, authUserId, lab)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to create lab", nil, err)
		return
	}
	// err1 := mc.diagnosticService.AddMapping(userId, labInfo)
	// if err1 != nil {
	// 	models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add mapping", nil, err1)
	// 	return
	// }
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Lab created successfully", labInfo, nil, nil)
}

func (pc *PatientController) GetPatientDiagnosticLabs(c *gin.Context) {
	_, userId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	labs, totalRecords, err := pc.diagnosticService.GetPatientDiagnosticLabs(userId, limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch patient labs", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	_, message := utils.GetResponseStatusMessage(
		len(labs),
		"Patient diagnostic labs retrieved successfully",
		"Labs not found",
	)
	models.SuccessResponse(c, constant.Success, http.StatusOK, message, labs, pagination, nil)
}

func (pc *PatientController) GetAllLabs(c *gin.Context) {
	_, _, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	data, totalRecords, err := pc.diagnosticService.GetAllLabs(limit, offset)
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

func (pc *PatientController) DeleteLab(c *gin.Context) {
	authUserId, user_id, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	diagnosticlLabId, err := strconv.ParseUint(c.Param("lab_id"), 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid lab ID", nil, err)
		return
	}
	if err := pc.diagnosticService.DeleteLabByUser(diagnosticlLabId, user_id, authUserId); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to delete lab", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Lab deleted successfully", nil, nil, nil)
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
	user_id, err := pc.userService.GetUserIdBySUB(sub.(string))
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

		var allDocs []*models.TblMedicalRecord
		for _, item := range items {
			record, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			if record["type"] == "file" {
				newRecord := models.TblMedicalRecord{
					RecordName:        record["name"].(string),
					RecordSize:        utils.ParseIntField(record["size"].(string)),
					FileType:          record["mime"].(string),
					UploadSource:      "DigiLocker",
					UploadDestination: "DigiLocker",
					SourceAccount:     digiLockerId,
					RecordCategory:    "Report",
					Description:       record["description"].(string),
					UploadedBy:        userID,
					Status:            constant.StatusSuccess,
					RecordUrl:         "https://digilocker.meripehchaan.gov.in/public/oauth2/1/file/" + record["uri"].(string),
					FetchedAt:         time.Now(),
					CreatedAt:         utils.ParseDateField(record["date"]),
				}
				allDocs = append(allDocs, &newRecord)
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
		err = pc.medicalRecordService.SaveMedicalRecords(allDocs, userID)
		if err != nil {
			log.Println("Error occurend while saving medical records from digilocker for", userID, digiLockerId, err)
		}
		log.Println("Successfully saved medical records from DigiLocker for user ID:", userID)

	}(user_id, digiTokenRes["access_token"].(string), digiTokenRes["digilockerid"].(string))

}

func (pc *PatientController) ReadUserUploadedMedicalFile(ctx *gin.Context) {
	sub, reqUserId, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	user_id, err := pc.userService.GetUserIdBySUB(sub)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "User can not be authorised", nil, err)
		return
	}
	type UserRequest struct {
		ResourceId uint64 `json:"resource_id"`
	}
	var req UserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	response, err := pc.medicalRecordService.ReadMedicalRecord(req.ResourceId, user_id, reqUserId)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, string(constant.PermissionViewMedicalRecord), nil, err)
		return
	}

	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Resource loaded", response, nil, nil)
	return
}

func (pc *PatientController) AddMappingToMergeTestComponent(c *gin.Context) {
	authUserId, _, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	var req models.MergeComponentMapppingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input", nil, err)
		return
	}

	var mappings []models.DiagnosticTestComponentAliasMapping
	for _, aliasID := range req.AliasTestComponentIds {
		mappings = append(mappings, models.DiagnosticTestComponentAliasMapping{
			DiagnosticTestComponentId: req.DiagnosticTestComponentId,
			AliasTestComponentId:      aliasID,
			CreatedBy:                 authUserId,
			UpdatedBy:                 authUserId,
		})
	}

	if err := pc.diagnosticService.AddMappingToMergeTestComponent(mappings); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add alias mappings", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Test Component merge successfully", nil, nil, nil)
}

func (pc *PatientController) AddHealthStats(ctx *gin.Context) {
	_, userId, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var reportData models.LabReport
	if err := ctx.ShouldBindJSON(&reportData); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid health stats report", nil, err)
		return
	}
	recordedAt := utils.CombineDateTimeString(reportData.ReportDetails.ReportDate, reportData.ReportDetails.ReportTime)
	response, err := pc.diagnosticService.GetSinglePatientDiagnosticLab(userId, reportData.ReportDetails.DiagnosticLabId)
	if err != nil {
		log.Printf("Lab not found  %d: %v", userId, err)
		// models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Lab not found", nil, err)
		// return
	}
	reportData.ReportDetails.LabName = response.LabName
	reportData.ReportDetails.ReportDate = recordedAt
	_, err1 := pc.diagnosticService.DigitizeDiagnosticReport(reportData, userId, func() *uint64 { v := uint64(0); return &v }())
	if err1 != nil {
		log.Printf("Health stats update error : %d: %v", userId, err1)
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, err1.Error(), nil, err1)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Health stats updated successfully.", nil, nil, nil)
}

func (pc *PatientController) ArchivePatientDiagnosticReport(c *gin.Context) {
	_, _, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	reportIDStr := c.Query("report_id")
	isDeletedStr := c.Query("is_deleted")

	if reportIDStr == "" || isDeletedStr == "" {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Missing required query parameters", nil, nil)
		return
	}

	reportID, err := strconv.ParseUint(reportIDStr, 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid report ID", nil, err)
		return
	}

	isDeleted, err := strconv.Atoi(isDeletedStr)
	if err != nil || isDeleted != 1 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Delete flag should be only 1", nil, nil)
		return
	}

	err = pc.diagnosticService.ArchivePatientDiagnosticReport(reportID, isDeleted)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to archive patient diagnostic report", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient diagnostic report archived successfully", nil, nil, nil)
}

func (pc *PatientController) AddPatientReportNote(ctx *gin.Context) {
	type UpdateReportCommentRequest struct {
		PatientReportId uint64 `json:"patient_diagnostic_report_id" binding:"required"`
		PatientId       uint64 `json:"patient_id" binding:"required"`
		Comment         string `json:"comments" binding:"required"`
	}
	var req UpdateReportCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request", nil, err)
		return
	}

	err := pc.diseaseService.AddPatientReportNote(req.PatientReportId, req.PatientId, req.Comment)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to update report comment", nil, err)
		return
	}

	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Comment updated successfully", nil, nil, nil)
}

func (pc *PatientController) SaveUserHealthProfile(ctx *gin.Context) {
	authUserId, user_id, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var userReq models.PatientHealthProfileRequest
	if err := ctx.ShouldBindJSON(&userReq); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request format: please check the data you're sending", nil, err)
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
	bmi, category := utils.CalculatePatientBMI(userReq.WeightKg, userReq.HeightCm)

	healthData := &models.TblPatientHealthProfile{
		PatientId:             user_id,
		HeightCM:              userReq.HeightCm,
		WeightKG:              userReq.WeightKg,
		BMI:                   bmi,
		BmiCategory:           category,
		BloodType:             userReq.BloodType,
		SmokingStatus:         userReq.SmokingStatus,
		AlcoholConsumption:    userReq.AlcoholConsumption,
		PhysicalActivityLevel: userReq.PhysicalActivityLevel,
		DietaryPreferences:    userReq.DietaryPreferences,
		ExistingConditions:    userReq.ExistingConditions,
		FamilyMedicalHistory:  userReq.FamilyMedicalHistory,
		MenstrualHistory:      userReq.MenstrualHistory,
		Notes:                 userReq.Notes,
		CreatedBy:             authUserId,
		UpdatedBy:             authUserId,
	}

	savedHealthProfile, err := pc.patientService.SaveUserHealthProfile(tx, healthData)
	if err != nil {
		tx.Rollback()
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Failed to save health profile", nil, err)
		return
	}
	for _, allergy := range userReq.Allergies {
		err := pc.allergyService.AddPatientAllergyRestriction(tx, &models.PatientAllergyRestriction{
			PatientId:  user_id,
			AllergyId:  allergy.AllergyID,
			SeverityId: allergy.SeverityId,
		})
		if err != nil {
			tx.Rollback()
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Failed to save allergies", nil, err)
			return
		}
	}
	_, err = pc.patientService.AddPatientDiseaseProfile(tx, &models.PatientDiseaseProfile{PatientId: user_id, DiseaseProfileId: userReq.DiseaseProfileID, AttachedFlag: 0})
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

func (pc *PatientController) GetPatientHealthProfileInfo(ctx *gin.Context) {
	_, patientId, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	healthDetails, err := pc.patientService.GetPatientHealthDetail(patientId)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusNotFound, "Health detail not found", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Health Detail fetch successfully", healthDetails, nil, nil)
	return
}

func (pc *PatientController) UpdatePatientHealthDetail(ctx *gin.Context) {
	authUserId, patientId, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var req models.TblPatientHealthProfile
	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid input", nil, err)
		return
	}
	req.PatientId = patientId
	req.UpdatedBy = authUserId
	err = pc.patientService.UpdatePatientHealthDetail(&req)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to update health detail", nil, err)
		return
	}

	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Health Detail updated successfully", nil, nil, nil)
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
	authUserId, _, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	user_id, err := pc.userService.GetUserIdBySUB(authUserId)
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
	authUserId, _, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	user_id, err := pc.userService.GetUserIdBySUB(authUserId)
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
	_, _, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var req SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	authHeader := c.GetHeader("Authorization")
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
	_, patientId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var req SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
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
	// log.Println("exchanged ", exchanged)
	baseURL := os.Getenv("SHARE_REPORT_BASE_URL")
	if baseURL == "" {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Missing base URL in environment variables", nil, nil)
		return
	}

	longLink := fmt.Sprintf("%s/shared-report?token=%s", baseURL, token)
	shortCode := generateShortCode()
	shortURLMap[shortCode] = longLink
	shortURL := fmt.Sprintf("%s/v1/user/r/%s", os.Getenv("SHORT_URL_BASE"), shortCode)

	err1 := pc.emailService.ShareReportEmail(req.Email, userDetails, shortURL)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to send email", nil, err1)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Email sent successfully", nil, nil, nil)
}

func (pc *PatientController) GetUserOrders(ctx *gin.Context) {
	_, user_id, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	orders, err := pc.orderService.GetOrdersByUserID(user_id)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to fetch orders", nil, err)
		return
	}

	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Orders fetched successfully", orders, nil, nil)
	return
}

func (pc *PatientController) CreateOrder(ctx *gin.Context) {
	var req models.CreateOrderRequest
	_, user_id, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	res, err := pc.orderService.CreateOrder(&req, user_id)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "failed to create order", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Order placed successfully", res, nil, nil)
	return
}

func (pc *PatientController) AddTestComponentDisplayConfig(ctx *gin.Context) {
	authUserId, user_id, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var input models.PatientTestComponentDisplayConfig
	if err := ctx.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid input", nil, err)
		return
	}
	message := "Test component pinned successfully"
	if input.IsPinned != nil && !*input.IsPinned {
		message = "Test component unpinned successfully"
	}
	if input.IsPinned != nil && *input.IsPinned {
		count, err2 := pc.patientService.GetPinnedComponentCount(user_id)
		if err2 != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed get pinned component count", nil, err2)
			return
		}
		if count >= 6 {
			models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Oops! You can’t pin more than 6 test components. Please unpin one to proceed", nil, nil, nil)
			return
		}
	}
	if input.IsPinned == nil && input.DisplayPriority == nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "At least one of is_pinned or display_priority must be provided", nil, nil)
		return
	}
	input.PatientId = user_id
	input.CreatedBy = authUserId
	err1 := pc.patientService.AddTestComponentDisplayConfig(&input)
	if err1 != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to upsert display config", nil, err1)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, message, nil, nil, nil)
}

func (pc *PatientController) GetUserMessages(ctx *gin.Context) {
	_, user_id, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	notifications, err := pc.notificationService.GetUserNotifications(user_id)
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "notifications loaded", notifications, nil, nil)
	return
}

func (pc *PatientController) SetUserReminder(ctx *gin.Context) {
	_, user_id, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	user, err := pc.patientService.GetUserProfileByUserId(user_id)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	if user.Email == "" {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "please add email id for scheduling reminders", nil, err)
		return
	}
	var reminderReq []models.ReminderConfig
	if err := ctx.ShouldBindJSON(&reminderReq); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "invalid request body", nil, err)
		return
	}

	err = pc.notificationService.ScheduleReminders(user.NotifyId, user.FirstName, user_id, reminderReq)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "reminder could not be added", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "reminder saved successfully", nil, nil, nil)
	return
}

func (pc *PatientController) GetUserReminders(ctx *gin.Context) {
	_, user_id, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	notifications, _ := pc.notificationService.GetUserReminders(user_id)
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "reminders loaded", notifications, nil, nil)
	return
}

func (pc *PatientController) AssignPermissionHandler(ctx *gin.Context) {
	_, user_id, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	// hasHOF, err := pc.roleService.HasHOFMapping(user_id, string(constant.MappingTypeHOF))
	// if !hasHOF {
	// 	models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Only the Head of Family can manage all permission", nil, err)
	// 	return
	// }
	type PermissionEntry struct {
		Code  string `json:"code"`
		Value bool   `json:"value"`
	}
	type AssignPermissionRequest struct {
		RelativeID  uint64            `json:"relative_id" binding:"required"`
		Permissions []PermissionEntry `json:"permissions" binding:"required"`
	}
	var req AssignPermissionRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadGateway, err.Error(), nil, err)
		return
	}
	permMap := make(map[string]bool)
	for _, p := range req.Permissions {
		permMap[p.Code] = p.Value
	}

	err = pc.permissionService.AssignMultiplePermissions(user_id, req.RelativeID, permMap)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to assign permission", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Permission assigned successfully", nil, nil, nil)
	return
}

func (pc *PatientController) ManagePermission(c *gin.Context) {
	sub, patientId, isDelegate, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	log.Printf("[ManagePermission] sub: %s, patientId: %d, isDelegate: %t", sub, patientId, isDelegate)

	// If the user is a delegate, check if they have permission
	if isDelegate {
		log.Printf("[ManagePermission] Delegate check: sub: %s, patientId: %d", sub, patientId)

		reqUserID, err := pc.userService.GetUserIdBySUB(sub)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "unable to fetch user ID", nil, err)
			return
		}

		// Validate if delegate has permission to manage
		err = pc.patientService.CanContinue(patientId, reqUserID, constant.PermissionViewHealth)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusForbidden, string(constant.PermissionManage), nil, err)
			return
		}
	}

	// Ensure patientId has HOF mapping
	hasHOF, err := pc.roleService.HasHOFMapping(patientId, string(constant.MappingTypeHOF))
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "error while checking HOF", nil, err)
		return
	}

	if !hasHOF {
		models.ErrorResponse(c, constant.Failure, http.StatusForbidden, "only the Head of Family can manage permissions", nil, nil)
		return
	}

	var req models.ManagePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "invalid request format", nil, err)
		return
	}

	err = pc.permissionService.ManagePermission(req)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusForbidden, err.Error(), nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Permissions updated successfully", nil, nil, nil)
}

func (pc *PatientController) SendSOSHandler(ctx *gin.Context) {
	ip := ctx.ClientIP()
	userAgent := ctx.Request.UserAgent()
	_, patientID, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	err = pc.patientService.SendSOS(patientID, ip, userAgent)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, err.Error(), nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Emergency SOS sent to all relatives", nil, nil, nil)
	return
}

func (pc *PatientController) GetDigitizationStatus(ctx *gin.Context) {
	recordID := ctx.Param("record_id")
	statusKey := fmt.Sprintf("record_status:%s", recordID)
	status, err := config.RedisClient.Get(context.Background(), statusKey).Result()
	if err == redis.Nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusNotFound, "No status found for this record", nil, nil)
		return
	} else if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Redis error", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Record status fetched", map[string]string{"status": status}, nil, nil)
}

func (pc *PatientController) GetUserMedications(ctx *gin.Context) {
	_, patientID, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	medicines, err := pc.patientService.GetPatientMedicines(patientID)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, err.Error(), nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "medicines loaded", medicines, nil, nil)
	return
}

func (c *PatientController) GetUserShareList(ctx *gin.Context) {
	_, patientId, _, err := utils.GetUserIDFromContext(ctx, c.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	shareList, err := c.patientService.GetUserShares(patientId)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to fetch share list", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Share list fetched", shareList, nil, nil)
}

func (c *PatientController) GetAllPermissions(ctx *gin.Context) {
	permissions, err := c.permissionService.GetAllPermissions()
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to fetch permission list", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "permissions list fetched", permissions, nil, nil)
}

func (pc *PatientController) GetMappedUserAddress(c *gin.Context) {
	_, userId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	allowedMappingTypes := map[string]bool{
		string(constant.MappingTypeC): true,
		string(constant.MappingTypeR): true,
	}

	mappingType := c.Query("mapping_type")
	if !allowedMappingTypes[mappingType] {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid mapping_type. Allowed values are: C, R", nil, nil)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	mappingTypeCR := []string{string(constant.MappingTypeR), string(constant.MappingTypeC)}

	data, totalRecords, err := pc.userService.GetAllMappedUserAddress(userId, limit, offset, mappingTypeCR)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve mapped user addresses", nil, err)
		return
	}

	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	statusCode, message := utils.GetResponseStatusMessage(
		len(data),
		"Mapped user addresses retrieved successfully",
		"Mapped user addresses not found",
	)

	models.SuccessResponse(c, constant.Success, statusCode, message, data, pagination, nil)
}

func (pc *PatientController) MovePatientRecord(ctx *gin.Context) {
	sub, userId, isDelegate, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	if isDelegate {
		reqUserID, err := pc.userService.GetUserIdBySUB(sub)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, err.Error(), nil, err)
			return
		}
		err = pc.patientService.CanContinue(userId, reqUserID, constant.PermissionEditProfile)
		if err != nil {
			models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, string(constant.PermissionChangeOwner), nil, err)
			return
		}
	}
	targetPatientIdStr := ctx.Query("user_id")
	recordIdStr := ctx.Query("record_id")
	reportIdStr := ctx.Query("report_id")

	if targetPatientIdStr == "" || recordIdStr == "" || reportIdStr == "" {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Missing query parameters", nil, nil)
		return
	}

	targetPatientId, err1 := strconv.ParseUint(targetPatientIdStr, 10, 64)
	recordId, err2 := strconv.ParseUint(recordIdStr, 10, 64)
	reportId, err3 := strconv.ParseUint(reportIdStr, 10, 64)

	if err1 != nil || err2 != nil || err3 != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid query parameter values", nil, fmt.Errorf("parse errors: %v %v %v", err1, err2, err3))
		return
	}

	updateErr := pc.medicalRecordService.MovePatientRecord(userId, targetPatientId, recordId, reportId)
	if updateErr != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to move patient records", nil, updateErr)
		return
	}

	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Patient records moved successfully", nil, nil, nil)

}

func (pc *PatientController) UpdateLab(ctx *gin.Context) {
	authUserId, _, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var lab models.DiagnosticLab
	if err := ctx.ShouldBindJSON(&lab); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid input", nil, err)
		return
	}
	if err := pc.diagnosticService.UpdateLab(&lab, authUserId); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, "Failed to update lab", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Lab updated successfully", lab, nil, nil)
	return
}

func (pc *PatientController) UpdateRelativeInfoById(c *gin.Context) {
	_, userId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	var relativeData models.UpdateRelativeRequest
	if err := c.ShouldBindJSON(&relativeData); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input data", nil, err)
		return
	}
	canAccess := pc.patientService.CanAccessAPI(userId, []string{string(constant.MappingTypeR), string(constant.MappingTypeHOF), string(constant.MappingTypeS)})
	if !canAccess {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Please get subscription inorder to perform this action", nil, errors.New("You need subscription for updating information"))
		return
	}
	err = pc.patientService.CheckPatientRelativeMapping(relativeData.RelativeID, userId, relativeData.MappingType)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusForbidden, "failed to update user", nil, err)
		return
	}

	user, err := pc.patientService.GetUserProfileByUserId(relativeData.RelativeID)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "failed to get user", nil, err)
		return
	}
	user.Email = relativeData.Email
	user.MobileNo = relativeData.MobileNo
	user.FirstName = relativeData.FirstName
	user.LastName = relativeData.LastName
	exist, _, _ := pc.userService.CheckUserEmailMobileExist(&models.CheckUserMobileEmail{Email: relativeData.Email})
	if !exist {
		err := pc.notificationService.UpadateUserInNotify(user.NotifyId, nil, &user.Email, &user.MobileNo)
		if err != nil {
			log.Println("Error while updating user on notify:", err)
		}
		log.Println("Going to Keycloak Update user id : ", user)
		userUpdateErr := pc.authService.UpdateUserInKeycloak(*user)
		if userUpdateErr != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "failed to update user in keycloak", nil, userUpdateErr)
			return
		}
	}
	err = pc.patientService.UpdateRelativeInfo(userId, &relativeData)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update patient info", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Patient info updated successfully", nil, nil, nil)
	return
}

func (pc *PatientController) SubscribeFamilyPlan(ctx *gin.Context) {
	_, userId, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	hasHOF, err := pc.roleService.HasHOFMapping(userId, string(constant.MappingTypeHOF))
	if !hasHOF {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Only the Head of Family can buy subscription for family.", nil, err)
		return
	}

	var req models.SubscribeFamilyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "Invalid input", nil, err)
		return
	}
	userInfo, err := pc.userService.GetSystemUserInfoByUserID(userId)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusBadRequest, "User not found", nil, err)
		return
	}
	message := "Subscription upgraded successfully"
	resp, exist, err := pc.subscriptionService.SubscribePlan(req, userId, userInfo)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, err.Error(), nil, err)
		return
	}
	if exist {
		message = "You already have active plan, Do you want to upgrade the plan"
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, message, resp, nil, nil)
}

func (pc *PatientController) GetSubscriptionPlanService(ctx *gin.Context) {
	_, _, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	result, err := pc.subscriptionService.FetchSubscriptionPlanService()
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusNotFound, "Plan not found", nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "Plan fetched successfully", result, nil, nil)
}

func (pc *PatientController) GetActiveSubscription(ctx *gin.Context) {
	_, userId, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	resp, exist, status, message, err := pc.subscriptionService.GetActiveSubscriptionPlanByMemberId(userId)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, message, nil, err)
		return
	}
	if exist && status == constant.SUBSCRIPTIONACTIVE {
		message = "Subscription plan fetched successfully."
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, message, resp, nil, nil)
}

func (pc *PatientController) AskAIHandler(c *gin.Context) {
	_, userId, _, err := utils.GetUserIDFromContext(c, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	userInfo, err := pc.userService.GetSystemUserInfoByUserID(userId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "User not found", nil, err)
		return
	}
	var req models.AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}

	response, err := pc.patientService.StartConversation(req.Message, userInfo)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to get AI response", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "AI response fetched", response.Response, nil, nil)
}

func (pc *PatientController) GetRecentUserProcesses(ctx *gin.Context) {
	_, userId, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	processKey := ctx.DefaultQuery("process_key", "")
	processes, err := pc.processStatusService.GetUserRecentProcesses(userId, processKey)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInsufficientStorage, err.Error(), nil, err)
		return
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, "process status loaded", processes, nil, nil)
	return
}

func (pc *PatientController) GetUserActivityLog(ctx *gin.Context) {
	_, userId, _, err := utils.GetUserIDFromContext(ctx, pc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(ctx)
	logs, totalRecords, err := pc.processStatusService.GetUserActivityLog(userId, limit, offset)
	if err != nil {
		models.ErrorResponse(ctx, constant.Failure, http.StatusInternalServerError, err.Error(), nil, err)
		return
	}

	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	message := "No user activity logs found"
	if len(logs) > 0 {
		message = "User activity logs fetched"
	}
	models.SuccessResponse(ctx, constant.Success, http.StatusOK, message, logs, pagination, nil)
}
