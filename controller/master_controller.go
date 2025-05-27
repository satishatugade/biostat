package controller

import (
	"biostat/constant"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type MasterController struct {
	allergyService      service.AllergyService
	diseaseService      service.DiseaseService
	causeService        service.CauseService
	symptomService      service.SymptomService
	medicationService   service.MedicationService
	dietService         service.DietService
	exerciseService     service.ExerciseService
	diagnosticService   service.DiagnosticService
	roleService         service.RoleService
	supportGroupService service.SupportGroupService
	hospitalService     service.HospitalService
	userService         service.UserService
}

func NewMasterController(allergyService service.AllergyService, diseaseService service.DiseaseService,
	causeService service.CauseService, symptomService service.SymptomService, medicationService service.MedicationService,
	dietService service.DietService, exerciseService service.ExerciseService, diagnosticService service.DiagnosticService,
	roleService service.RoleService, supportGroupService service.SupportGroupService, hospitalService service.HospitalService, userService service.UserService) *MasterController {
	return &MasterController{allergyService: allergyService,
		diseaseService:      diseaseService,
		causeService:        causeService,
		symptomService:      symptomService,
		medicationService:   medicationService,
		dietService:         dietService,
		exerciseService:     exerciseService,
		diagnosticService:   diagnosticService,
		roleService:         roleService,
		supportGroupService: supportGroupService,
		hospitalService:     hospitalService,
		userService:         userService,
	}
}

func (mc *MasterController) CreateDiseaseProfile(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var req models.DiseaseProfile
	if err := c.ShouldBindJSON(&req); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request payload", nil, err)
		return
	}
	req.CreatedAt = time.Now()
	req.CreatedBy = authUserId
	if err := mc.diseaseService.CreateDiseaseProfile(req); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Disease profile already exists.", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Disease profile created successfully", nil, nil, nil)
}

func (mc *MasterController) CreateDisease(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var disease models.Disease
	if err := c.ShouldBindJSON(&disease); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request payload", nil, err)
		return
	}
	disease.CreatedBy = authUserId
	err1 := mc.diseaseService.CreateDisease(&disease)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to create disease condition.", nil, err1)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Disease condition created successfully", disease, nil, nil)
}

func (mc *MasterController) UpdateDiseaseInfo(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var Disease models.Disease
	if err := c.ShouldBindJSON(&Disease); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request input", nil, err)
		return
	}

	if Disease.DiseaseId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Disease Id Required", nil, nil)
		return
	}
	err := mc.diseaseService.UpdateDisease(&Disease, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update disease", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Disease updated successfully", Disease, nil, nil)
}

func (mc *MasterController) DeleteDisease(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	diseaseId, err := strconv.ParseUint(c.Param("disease_id"), 10, 32)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid disease Id", nil, err)
		return
	}
	err = mc.diseaseService.DeleteDisease(diseaseId, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to delete disease", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Disease deleted successfully", nil, nil, nil)
}

func (mc *MasterController) GetDiseaseAuditLogs(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var diseaseId uint64
	diseaseIdStr := c.Query("disease_id")
	if diseaseIdStr != "" {
		parsedDiseaseId, err := strconv.ParseUint(diseaseIdStr, 10, 32)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid disease Id", nil, err)
			return
		}
		diseaseId = parsedDiseaseId
	}

	var diseaseAuditId uint64
	if auditIdStr := c.Query("disease_audit_id"); auditIdStr != "" {
		auditId, err := strconv.ParseUint(auditIdStr, 10, 32)
		if err == nil {
			temp := auditId
			diseaseAuditId = temp
		}
	}
	page, limit, offset := utils.GetPaginationParams(c)
	message := "Disease audit record not found"
	if diseaseId == 0 && diseaseAuditId == 0 {
		data, totalRecords, err := mc.diseaseService.GetAllDiseaseAuditLogs(limit, offset)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit logs", nil, err)
			return
		}
		pagination := utils.GetPagination(limit, page, offset, totalRecords)
		statusCode, message := utils.GetResponseStatusMessage(
			len(data),
			"Disease audit record retrieved successfully",
			"Disease audit record not found",
		)
		models.SuccessResponse(c, constant.Success, statusCode, message, data, pagination, nil)
		return
	}

	auditRecord, err := mc.diseaseService.GetDiseaseAuditLogs(diseaseId, diseaseAuditId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit logs", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(auditRecord),
		constant.AuditSuccessMessage,
		constant.AuditErrorMessage,
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, auditRecord, nil, nil)

}

func (mc *MasterController) GetDiseaseInfo(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	DiseaseId, err := strconv.ParseUint(c.Param("disease_id"), 10, 32)
	page, limit, offset := utils.GetPaginationParams(c)
	message := "Diseases not found"
	if err != nil || DiseaseId == 0 {
		data, totalRecords, err := mc.diseaseService.GetAllDiseases(limit, offset)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve diseases", nil, err)
			return
		}
		pagination := utils.GetPagination(limit, page, offset, totalRecords)
		if len(data) > 0 {
			message = "All Diseases retrieved successfully"
		}
		models.SuccessResponse(c, constant.Success, http.StatusOK, message, data, pagination, nil)
		return
	}
	diseases, err := mc.diseaseService.GetDiseases(DiseaseId)
	if err == nil {
		message = "Diseases retrieved successfully"
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, message, diseases, nil, nil)
}

func (mc *MasterController) GetDiseaseProfile(c *gin.Context) {
	_, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var diseaseProfiles []models.DiseaseProfile
	var totalRecords int64

	page, limit, offset := utils.GetPaginationParams(c)
	diseaseProfiles, totalRecords, err := mc.diseaseService.GetDiseaseProfiles(limit, offset)
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

func (mc *MasterController) GetDiseaseProfileById(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	diseaseProfileId := c.Param("disease_profile_id")
	diseaseProfile, err := mc.diseaseService.GetDiseaseProfileById(diseaseProfileId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Disease profile not found", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Disease profile retrieved successfully", diseaseProfile, nil, nil)
}

func (mc *MasterController) GetAllCauses(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	causes, totalRecords, err := mc.causeService.GetAllCauses(limit, offset)
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

func (mc *MasterController) AddDiseaseCauseMapping(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var input models.DiseaseCauseMapping
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	input.CreatedBy = authUserId
	err := mc.causeService.AddDiseaseCauseMapping(&input)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Disease-cause mapping already exists.", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Disease-cause mapping added successfully.", nil, nil, nil)
}

func (mc *MasterController) AddDiseaseCause(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var cause models.Cause
	if err := c.ShouldBindJSON(&cause); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid cause input", nil, err)
		return
	}
	cause.CreatedBy = authUserId
	log.Println("AddDiseaseCause request data : ", cause)
	savedCause, err := mc.causeService.AddDiseaseCause(&cause)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Cause type does not exists.", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Cause added successfully", savedCause, nil, nil)
}

func (mc *MasterController) GetAllCauseTypes(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	page, limit, offset := utils.GetPaginationParams(c)
	isDeleted, ok := utils.GetQueryIntParam(c, "is_deleted", 0)
	if !ok {
		log.Println("GetAllCauseTypes is deleted status not provided : ", isDeleted)
	}
	causeTypes, totalRecords, err := mc.causeService.GetAllCauseTypes(limit, offset, isDeleted)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve cause types", nil, err)
		return
	}

	statusCode, message := utils.GetResponseStatusMessage(
		len(causeTypes),
		"Cause types retrieved successfully",
		"Cause types not found",
	)
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	models.SuccessResponse(c, constant.Success, statusCode, message, causeTypes, pagination, nil)
}

func (mc *MasterController) AddCauseType(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var causeType models.CauseTypeMaster
	if err := c.ShouldBindJSON(&causeType); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid cause type input", nil, err)
		return
	}
	causeType.CreatedBy = authUserId
	log.Println("AddCauseType request data : ", causeType)
	savedCauseType, err := mc.causeService.AddCauseType(&causeType)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Cause type could not be added.", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Cause type added successfully", savedCauseType, nil, nil)
}

func (mc *MasterController) UpdateCauseType(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	var causeType models.CauseTypeMaster
	if err := c.ShouldBindJSON(&causeType); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid cause type input", nil, err)
		return
	}

	causeType.UpdatedBy = authUserId
	log.Println("UpdateCauseType request data:", causeType)

	updatedCauseType, err := mc.causeService.UpdateCauseType(&causeType, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, err.Error(), nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Cause type updated successfully", updatedCauseType, nil, nil)
}

func (mc *MasterController) DeleteCauseType(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	causeTypeId, ok := utils.ParseUintParam(c, "cause_type_id")
	if !ok {
		return
	}
	if err := mc.causeService.DeleteCauseType(causeTypeId, authUserId); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, err.Error(), nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Cause type deleted successfully", nil, nil, nil)
}

func (mc *MasterController) UpdateDiseaseCause(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var cause models.Cause
	if err := c.ShouldBindJSON(&cause); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid disease cause input", nil, err)
		return
	}
	if cause.CauseId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Cause Id is required", nil, nil)
		return
	}
	log.Println("UpdateCause request data : ", cause)
	updatedCause, err := mc.causeService.UpdateCause(&cause, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update cause", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Cause updated successfully", updatedCause, nil, nil)
}

func (mc *MasterController) DeleteCause(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	causeId, err := strconv.ParseUint(c.Param("cause_id"), 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid cause Id", nil, err)
		return
	}
	log.Println("DeleteCause data By Id : ", causeId)
	DeleteErr := mc.causeService.DeleteCause(causeId, authUserId)
	if DeleteErr != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, DeleteErr.Error(), nil, DeleteErr)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Cause deleted successfully", nil, nil, nil)
}

func (mc *MasterController) GetCauseAuditRecord(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var causeId uint64
	causeIdStr := c.Query("cause_id")
	if causeIdStr != "" {
		parsedCauseId, err := strconv.ParseUint(causeIdStr, 10, 32)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid cause Id", nil, err)
			return
		}
		causeId = parsedCauseId
	}
	var causeAuditId uint64
	if auditIdStr := c.Query("cause_audit_id"); auditIdStr != "" {
		auditId, err := strconv.ParseUint(auditIdStr, 10, 32)
		if err == nil {
			causeAuditId = auditId
		}
	}
	page, limit, offset := utils.GetPaginationParams(c)
	message := "Cause audit record not found"
	if causeId == 0 && causeAuditId == 0 {
		data, totalRecords, err := mc.causeService.GetAllCauseAuditRecord(limit, offset)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit logs", nil, err)
			return
		}
		pagination := utils.GetPagination(limit, page, offset, totalRecords)
		statusCode, message := utils.GetResponseStatusMessage(
			len(data),
			constant.AuditSuccessMessage,
			constant.AuditErrorMessage,
		)
		models.SuccessResponse(c, constant.Success, statusCode, message, data, pagination, nil)
		return
	}
	auditRecord, err := mc.causeService.GetCauseAuditRecord(causeId, causeAuditId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit record", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(auditRecord),
		constant.AuditSuccessMessage,
		constant.AuditErrorMessage,
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, auditRecord, nil, nil)
}

func (mc *MasterController) GetCauseTypeAuditRecord(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var causeTypeId uint64
	causeTypeIdStr := c.Query("cause_type_id")
	if causeTypeIdStr != "" {
		parsedCauseTypeId, err := strconv.ParseUint(causeTypeIdStr, 10, 64)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid cause type ID", nil, err)
			return
		}
		causeTypeId = parsedCauseTypeId
	}

	var causeTypeAuditId uint64
	if auditIdStr := c.Query("cause_type_audit_id"); auditIdStr != "" {
		auditId, err := strconv.ParseUint(auditIdStr, 10, 64)
		if err == nil {
			causeTypeAuditId = auditId
		}
	}

	page, limit, offset := utils.GetPaginationParams(c)
	message := "Cause type audit record not found"

	if causeTypeId == 0 && causeTypeAuditId == 0 {
		data, totalRecords, err := mc.causeService.GetAllCauseTypeAuditRecord(limit, offset)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve cause type audit logs", nil, err)
			return
		}
		pagination := utils.GetPagination(limit, page, offset, totalRecords)
		statusCode, message := utils.GetResponseStatusMessage(
			len(data),
			constant.AuditSuccessMessage,
			constant.AuditErrorMessage,
		)
		models.SuccessResponse(c, constant.Success, statusCode, message, data, pagination, nil)
		return
	}

	auditRecord, err := mc.causeService.GetCauseTypeAuditRecord(causeTypeId, causeTypeAuditId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve cause type audit record", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(auditRecord),
		constant.AuditSuccessMessage,
		constant.AuditErrorMessage,
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, auditRecord, nil, nil)
}

func (mc *MasterController) GetAllSymptom(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	symptom, totalRecords, err := mc.symptomService.GetAllSymptom(limit, offset)
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

func (mc *MasterController) AddSymptom(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var symptom models.Symptom
	if err := c.ShouldBindJSON(&symptom); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid symptom input", nil, err)
		return
	}
	symptom.CreatedBy = authUserId
	log.Println("AddDiseaseSymptom request data : ", symptom)
	savedSymptoms, err := mc.symptomService.AddDiseaseSymptom(&symptom)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add symptom", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "symptom added successfully", savedSymptoms, nil, nil)
}

func (mc *MasterController) AddDiseaseSymptomMapping(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var input models.DiseaseSymptomMapping
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	input.CreatedBy = authUserId
	log.Println("AddDiseaseSymptomMapping mapping Id : ", input)
	err := mc.symptomService.AddDiseaseSymptomMapping(&input)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Disease-symptoms mapping already exists.", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Disease-symptoms mapping added successfully.", nil, nil, nil)
}

func (mc *MasterController) UpdateDiseaseSymptom(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var symptom models.Symptom
	if err := c.ShouldBindJSON(&symptom); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid disease symptom data", nil, err)
		return
	}
	if symptom.SymptomId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "symptom Id is required", nil, nil)
		return
	}
	log.Println("UpdateSymptom request data : ", symptom)
	updatedSymptoms, err := mc.symptomService.UpdateSymptom(&symptom, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update symptom data", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "symptom updated successfully", updatedSymptoms, nil, nil)
}

func (mc *MasterController) DeleteSymptom(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	symptomId, err := strconv.ParseUint(c.Param("symptom_id"), 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid symptom Id", nil, err)
		return
	}
	log.Println("DeleteSymptom data by Id : ", symptomId)
	err = mc.symptomService.DeleteSymptom(symptomId, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Deletion failed", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Symptom deleted successfully", nil, nil, nil)
}

func (mc *MasterController) GetSymptomAuditRecord(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var symptomId uint64
	symptomIdStr := c.Query("symptom_id")
	if symptomIdStr != "" {
		parsedSymptomId, err := strconv.ParseUint(symptomIdStr, 10, 32)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid symptom Id", nil, err)
			return
		}
		symptomId = parsedSymptomId
	}

	var symptomAuditId uint64
	if auditIdStr := c.Query("symptom_audit_id"); auditIdStr != "" {
		auditId, err := strconv.ParseUint(auditIdStr, 10, 32)
		if err == nil {
			symptomAuditId = auditId
		}
	}
	page, limit, offset := utils.GetPaginationParams(c)
	message := "Symptom audit record not found"
	if symptomId == 0 && symptomAuditId == 0 {
		data, totalRecords, err := mc.symptomService.GetAllSymptomAuditRecord(limit, offset)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit logs", nil, err)
			return
		}
		pagination := utils.GetPagination(limit, page, offset, totalRecords)
		statusCode, message := utils.GetResponseStatusMessage(
			len(data),
			constant.AuditSuccessMessage,
			constant.AuditErrorMessage,
		)
		models.SuccessResponse(c, constant.Success, statusCode, message, data, pagination, nil)
		return
	}
	auditRecord, err := mc.symptomService.GetSymptomAuditRecord(symptomId, symptomAuditId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit record", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(auditRecord),
		constant.AuditSuccessMessage,
		constant.AuditErrorMessage,
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, auditRecord, nil, nil)
}

func (mc *MasterController) GetAllSymptomTypes(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	isDeleted, ok := utils.GetQueryIntParam(c, "is_deleted", 0)
	if !ok {
		log.Println("GetAllSymptomTypes is deleted status not provided : ", isDeleted)
	}
	symptomTypes, totalRecords, err := mc.symptomService.GetAllSymptomTypes(limit, offset, isDeleted)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve symptom types", nil, err)
		return
	}

	statusCode, message := utils.GetResponseStatusMessage(
		len(symptomTypes),
		"Symptom types retrieved successfully",
		"Symptom types not found",
	)
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	models.SuccessResponse(c, constant.Success, statusCode, message, symptomTypes, pagination, nil)
}

func (mc *MasterController) AddSymptomType(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var symptomType models.SymptomTypeMaster
	if err := c.ShouldBindJSON(&symptomType); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid symptom type input", nil, err)
		return
	}
	symptomType.CreatedBy = authUserId
	log.Println("AddSymptomType request data : ", symptomType)
	savedSymptomType, err := mc.symptomService.AddSymptomType(&symptomType)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Symptom type could not be added.", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Symptom type added successfully", savedSymptomType, nil, nil)
}

func (mc *MasterController) UpdateSymptomType(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var symptomType models.SymptomTypeMaster
	if err := c.ShouldBindJSON(&symptomType); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid symptom type input", nil, err)
		return
	}

	symptomType.UpdatedBy = authUserId
	log.Println("UpdateSymptomType request data:", symptomType)

	updatedSymptomType, err := mc.symptomService.UpdateSymptomType(&symptomType, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, err.Error(), nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Symptom type updated successfully", updatedSymptomType, nil, nil)
}

func (mc *MasterController) DeleteSymptomType(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	symptomTypeId, ok := utils.ParseUintParam(c, "symptom_type_id")
	if !ok {
		return
	}
	if err := mc.symptomService.DeleteSymptomType(symptomTypeId, authUserId); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, err.Error(), nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Symptom type deleted successfully", nil, nil, nil)
}

func (mc *MasterController) GetSymptomTypeAuditRecord(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var symptomTypeId uint64
	symptomTypeIdStr := c.Query("symptom_type_id")
	if symptomTypeIdStr != "" {
		parsedSymptomTypeId, err := strconv.ParseUint(symptomTypeIdStr, 10, 64)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid symptom type ID", nil, err)
			return
		}
		symptomTypeId = parsedSymptomTypeId
	}

	var symptomTypeAuditId uint64
	if auditIdStr := c.Query("symptom_type_audit_id"); auditIdStr != "" {
		auditId, err := strconv.ParseUint(auditIdStr, 10, 64)
		if err == nil {
			symptomTypeAuditId = auditId
		}
	}

	page, limit, offset := utils.GetPaginationParams(c)
	message := "Symptom type audit record not found"

	if symptomTypeId == 0 && symptomTypeAuditId == 0 {
		data, totalRecords, err := mc.symptomService.GetAllSymptomTypeAuditRecord(limit, offset)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve symptom type audit logs", nil, err)
			return
		}
		pagination := utils.GetPagination(limit, page, offset, totalRecords)
		statusCode, message := utils.GetResponseStatusMessage(
			len(data),
			constant.AuditSuccessMessage,
			constant.AuditErrorMessage,
		)
		models.SuccessResponse(c, constant.Success, statusCode, message, data, pagination, nil)
		return
	}
	auditRecord, err := mc.symptomService.GetSymptomTypeAuditRecord(symptomTypeId, symptomTypeAuditId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve symptom type audit record", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(auditRecord),
		constant.AuditSuccessMessage,
		constant.AuditErrorMessage,
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, auditRecord, nil, nil)
}

func (mc *MasterController) AddDietPlanTemplate(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var dietPlan models.DietPlanTemplate
	if err := c.ShouldBindJSON(&dietPlan); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid diet plan data input", nil, err)
		return
	}
	dietPlan.CreatedBy = authUserId
	log.Println("CreateDietPlanTemplate request data : ", dietPlan)
	if err := mc.dietService.CreateDietPlanTemplate(&dietPlan); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add diet plan", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Diet plan added successfully", dietPlan, nil, nil)
}
func (mc *MasterController) GetAllDietPlanTemplates(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	dietPlans, totalRecords, err := mc.dietService.GetDietPlanTemplates(limit, offset)
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

func (mc *MasterController) GetDietPlanById(c *gin.Context) {
	_, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	dietPlanTemplateId := c.Param("diet_id")
	var dietPlan models.DietPlanTemplate
	log.Println("GetDietPlanById dietPlanTemplateId : ", dietPlanTemplateId)
	dietPlan, err := mc.dietService.GetDietPlanById(dietPlanTemplateId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Diet plan not found", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diet plan retrieved successfully", dietPlan, nil, nil)
}

func (mc *MasterController) UpdateDietPlanTemplate(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	dietPlanTemplateIdStr := c.Param("diet_id")
	dietPlanTemplateId, err := strconv.ParseUint(dietPlanTemplateIdStr, 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid diet_id parameter", nil, err)
		return
	}
	var dietPlan models.DietPlanTemplate
	if err := c.ShouldBindJSON(&dietPlan); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid diet plan input", nil, err)
		return
	}
	dietPlan.CreatedBy = authUserId
	log.Println("UpdateDietPlanTemplate by template Id : ", dietPlan)
	if err := mc.dietService.UpdateDietPlanTemplate(dietPlanTemplateId, &dietPlan); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update diet plan", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diet plan updated successfully", dietPlan, nil, nil)
}

func (mc *MasterController) AddDiseaseDietMapping(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var input models.DiseaseDietMapping
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	input.CreatedBy = authUserId
	log.Println("AddDiseaseDietMapping request data : ", input)
	err := mc.dietService.AddDiseaseDietMapping(&input)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Disease-diet mapping already exists.", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Disease-diet mapping added successfully.", nil, nil, nil)
}

func (mc *MasterController) AddDiseaseExerciseMapping(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var input models.DiseaseExerciseMapping
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	input.CreatedBy = authUserId
	log.Println("AddDiseaseExerciseMapping request data : ", input)
	err := mc.exerciseService.AddDiseaseExerciseMapping(&input)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Disease-exercise mapping already exists.", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Disease-exercise mapping added successfully.", nil, nil, nil)
}

func (mc *MasterController) AddExercise(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var exercise models.Exercise
	if err := c.ShouldBindJSON(&exercise); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid exercise input", nil, err)
		return
	}
	exercise.CreatedBy = authUserId
	log.Println("CreateExercise request data : ", exercise)
	if err := mc.exerciseService.CreateExercise(&exercise); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add exercise", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Exercise added successfully", exercise, nil, nil)
}

func (mc *MasterController) GetAllExercises(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	exercises, totalRecords, err := mc.exerciseService.GetExercises(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve exercises", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	statusCode, message := utils.GetResponseStatusMessage(
		len(exercises),
		"Exercise info retrieved successfully",
		"Exercise info not found",
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, exercises, pagination, nil)
}

func (mc *MasterController) GetExerciseById(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	exerciseIdStr := c.Param("exercise_id")
	exerciseId, err := strconv.ParseUint(exerciseIdStr, 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid exercise Id", nil, err)
		return
	}
	exercise, err := mc.exerciseService.GetExerciseById(uint64(exerciseId))
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Exercise data not found", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Exercise retrieved successfully", exercise, nil, nil)
}

func (mc *MasterController) UpdateExercise(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var exercise models.Exercise
	if err := c.ShouldBindJSON(&exercise); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid exercise input", nil, err)
		return
	}
	if err := mc.exerciseService.UpdateExercise(authUserId, &exercise); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update exercise", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Exercise updated successfully", exercise, nil, nil)
}

func (mc *MasterController) DeleteExercise(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	exerciseId, err := strconv.ParseUint(c.Param("exercise_id"), 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid exercise ID", nil, err)
		return
	}
	err = mc.exerciseService.DeleteExercise(exerciseId, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Deletion failed", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Exercise deleted", nil, nil, nil)
}

func (mc *MasterController) GetExerciseAuditRecord(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var exerciseId uint64
	exerciseIdStr := c.Query("exercise_id")
	if exerciseIdStr != "" {
		parsedExerciseId, err := strconv.ParseUint(exerciseIdStr, 10, 64)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid exercise ID", nil, err)
			return
		}
		exerciseId = parsedExerciseId
	}
	var exerciseAuditId uint64
	if auditIdStr := c.Query("exercise_audit_id"); auditIdStr != "" {
		auditId, err := strconv.ParseUint(auditIdStr, 10, 64)
		if err == nil {
			exerciseAuditId = auditId
		}
	}
	page, limit, offset := utils.GetPaginationParams(c)
	message := "Exercise audit record not found"
	if exerciseId == 0 && exerciseAuditId == 0 {
		data, totalRecords, err := mc.exerciseService.GetAllExerciseAuditRecord(limit, offset)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit record", nil, err)
			return
		}
		pagination := utils.GetPagination(limit, page, offset, totalRecords)
		statusCode, message := utils.GetResponseStatusMessage(
			len(data),
			constant.AuditSuccessMessage,
			constant.AuditErrorMessage,
		)
		models.SuccessResponse(c, constant.Success, statusCode, message, data, pagination, nil)
		return
	}
	auditRecord, err := mc.exerciseService.GetExerciseAuditRecord(exerciseId, exerciseAuditId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit logs", nil, err)
		return
	}

	statusCode, message := utils.GetResponseStatusMessage(
		len(auditRecord),
		constant.AuditSuccessMessage,
		constant.AuditErrorMessage,
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, auditRecord, nil, nil)
}

func (mc *MasterController) GetAllergyRestrictions(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	allergies, err := mc.allergyService.GetAllergies()
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve allergies", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Allergies retrieved successfully", allergies, nil, nil)
}

func (mc *MasterController) GetAllMedication(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
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

func (mc *MasterController) AddMedication(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var medication models.Medication
	if err := c.ShouldBindJSON(&medication); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid medication input", nil, err)
		return
	}
	medication.CreatedBy = authUserId
	if err := mc.medicationService.CreateMedication(&medication); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add medication", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Medication added successfully", medication, nil, nil)
}

func (mc *MasterController) UpdateMedication(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var medication models.Medication
	if err := c.ShouldBindJSON(&medication); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid medication update input", nil, err)
		return
	}
	if err := mc.medicationService.UpdateMedication(&medication, authUserId); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update medication", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Medication updated successfully", medication, nil, nil)
}

func (mc *MasterController) DeleteMedication(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	medicationId, err := strconv.ParseUint(c.Param("medication_id"), 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid medication Id", nil, err)
		return
	}
	err = mc.medicationService.DeleteMedication(medicationId, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Medication deletion failed", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Medication deleted", nil, nil, nil)
}

func (mc *MasterController) AddDiseaseMedicationMapping(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var input models.DiseaseMedicationMapping
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	input.CreatedBy = authUserId
	err := mc.medicationService.AddDiseaseMedicationMapping(&input)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add disease-medication mapping", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Disease-medication mapping added successfully", nil, nil, nil)
}

func (mc *MasterController) GetMedicationAuditRecord(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var medicationId uint64
	medicationIdStr := c.Query("medication_id")
	if medicationIdStr != "" {
		parsedMedicationId, err := strconv.ParseUint(medicationIdStr, 10, 32)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid medication Id", nil, err)
			return
		}
		medicationId = parsedMedicationId
	}
	var medicationAuditId uint64
	if auditIdStr := c.Query("medication_audit_id"); auditIdStr != "" {
		auditId, err := strconv.ParseUint(auditIdStr, 10, 32)
		if err == nil {
			medicationAuditId = auditId
		}
	}
	page, limit, offset := utils.GetPaginationParams(c)
	message := "Medication audit record not found"
	if medicationId == 0 && medicationAuditId == 0 {
		data, totalRecords, err := mc.medicationService.GetAllMedicationAuditRecord(limit, offset)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit record", nil, err)
			return
		}
		pagination := utils.GetPagination(limit, page, offset, totalRecords)
		statusCode, message := utils.GetResponseStatusMessage(
			len(data),
			constant.AuditSuccessMessage,
			constant.AuditErrorMessage,
		)
		models.SuccessResponse(c, constant.Success, statusCode, message, data, pagination, nil)
		return
	}
	auditRecord, err := mc.medicationService.GetMedicationAuditRecord(medicationId, medicationAuditId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit logs", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(auditRecord),
		constant.AuditSuccessMessage,
		constant.AuditErrorMessage,
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, auditRecord, nil, nil)
}

func (mc *MasterController) AddDiseaseDiagnosticTestMapping(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var input models.DiseaseDiagnosticTestMapping
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	input.CreatedBy = authUserId
	err := mc.diagnosticService.AddDiseaseDiagnosticTestMapping(&input)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Disease-Diagnostic-Test mapping already exists.", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Disease-Diagnostic-Test mapping added successfully..", nil, nil, nil)
}

func (mc *MasterController) GetDiagnosticTests(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	diagnosticTest, totalRecord, err := mc.diagnosticService.GetDiagnosticTests(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve diagnostic tests", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecord)
	message := "Diagnostic test not found"
	if len(diagnosticTest) > 0 {
		message = "Diagnostic test retrieved successfully"
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, message, diagnosticTest, pagination, nil)
}

func (mc *MasterController) CreateDiagnosticTest(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var diagnosticTest models.DiagnosticTest
	err := c.ShouldBindJSON(&diagnosticTest)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	diagnosticTestRes, err := mc.diagnosticService.CreateDiagnosticTest(&diagnosticTest, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to create diagnostic test", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Diagnostic test created successfully", diagnosticTestRes, nil, nil)
}

func (mc *MasterController) UpdateDiagnosticTest(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var diagnosticTest models.DiagnosticTest
	err := c.ShouldBindJSON(&diagnosticTest)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	if diagnosticTest.DiagnosticTestId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Diagnostic Test Id is required", nil, nil)
		return
	}
	diagnosticTestRes, err := mc.diagnosticService.UpdateDiagnosticTest(&diagnosticTest, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update diagnostic test", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diagnostic test updated successfully", diagnosticTestRes, nil, nil)
}

func (mc *MasterController) GetSingleDiagnosticTest(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	diagnosticTestId := utils.GetParamAsInt(c, "diagnosticTestId")
	if diagnosticTestId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "DiagnosticTestId is required", nil, nil)
		return
	}
	diagnosticTest, err := mc.diagnosticService.GetSingleDiagnosticTest(diagnosticTestId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve diagnostic test", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diagnostic test retrieved successfully", diagnosticTest, nil, nil)
}

func (mc *MasterController) DeleteDiagnosticTest(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	diagnosticTestId := utils.GetParamAsInt(c, "diagnosticTestId")
	if diagnosticTestId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "DiagnosticTestId is required", nil, nil)
		return
	}
	err := mc.diagnosticService.DeleteDiagnosticTest(diagnosticTestId, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to delete diagnostic test", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diagnostic test deleted successfully", nil, nil, nil)
}

func (mc *MasterController) GetAllDiagnosticComponents(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	diagnosticTestComponent, totalRecord, err := mc.diagnosticService.GetAllDiagnosticComponents(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve diagnostic components", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecord)
	message := "Diagnostic components not found"
	if len(diagnosticTestComponent) > 0 {
		message = "Diagnostic components retrieved successfully"
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, message, diagnosticTestComponent, pagination, nil)
}

func (mc *MasterController) CreateDiagnosticComponent(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var diagnosticComponent models.DiagnosticTestComponent
	err := c.ShouldBindJSON(&diagnosticComponent)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	diagnosticComponent.CreatedBy = authUserId
	diagnosticComponentRes, err := mc.diagnosticService.CreateDiagnosticComponent(&diagnosticComponent)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to create diagnostic component", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Diagnostic component created successfully", diagnosticComponentRes, nil, nil)
}

func (mc *MasterController) UpdateDiagnosticComponent(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var diagnosticComponent models.DiagnosticTestComponent
	err := c.ShouldBindJSON(&diagnosticComponent)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	if diagnosticComponent.DiagnosticTestComponentId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Diagnostic Test ComponentId is required", nil, nil)
		return
	}
	diagnosticComponentRes, err := mc.diagnosticService.UpdateDiagnosticComponent(authUserId, &diagnosticComponent)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update diagnostic component", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diagnostic component updated successfully", diagnosticComponentRes, nil, nil)
}

func (mc *MasterController) GetSingleDiagnosticComponent(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	diagnosticComponentId := utils.GetParamAsInt(c, "diagnosticComponentId")
	if diagnosticComponentId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "DiagnosticComponentId is required", nil, nil)
		return
	}
	diagnosticComponent, err := mc.diagnosticService.GetSingleDiagnosticComponent(diagnosticComponentId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve diagnostic component", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diagnostic component retrieved successfully", diagnosticComponent, nil, nil)
}

func (mc *MasterController) GetAllDiagnosticTestComponentMappings(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	diagnosticTestComponentMapping, totalRecord, err := mc.diagnosticService.GetAllDiagnosticTestComponentMappings(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve diagnostic test component mappings", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecord)
	message := "Diagnostic test component mappings not found"
	if len(diagnosticTestComponentMapping) > 0 {
		message = "Diagnostic test component mappings retrieved successfully"
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, message, diagnosticTestComponentMapping, pagination, nil)
}

func (mc *MasterController) CreateDiagnosticTestComponentMapping(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var diagnosticTestComponentMapping models.DiagnosticTestComponentMapping
	err := c.ShouldBindJSON(&diagnosticTestComponentMapping)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	diagnosticTestComponentMapping.CreatedBy = authUserId
	diagnosticTestComponentMappingRes, err := mc.diagnosticService.CreateDiagnosticTestComponentMapping(&diagnosticTestComponentMapping)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Diagnostic test-component mapping already exists.", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Diagnostic test component mapping created successfully", diagnosticTestComponentMappingRes, nil, nil)
}

func (mc *MasterController) UpdateDiagnosticTestComponentMapping(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var diagnosticTestComponentMapping models.DiagnosticTestComponentMapping
	err := c.ShouldBindJSON(&diagnosticTestComponentMapping)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	if diagnosticTestComponentMapping.DiagnosticTestComponentMappingId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "DiagnosticTestComponentMappingId is required", nil, nil)
		return
	}
	diagnosticTestComponentMapping.UpdatedBy = authUserId
	diagnosticTestComponentMappingRes, err := mc.diagnosticService.UpdateDiagnosticTestComponentMapping(&diagnosticTestComponentMapping)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update diagnostic test component mapping", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diagnostic test component mapping updated successfully", diagnosticTestComponentMappingRes, nil, nil)
}

func (mc *MasterController) DeleteDiagnosticTestComponentMapping(c *gin.Context) {
	_, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var diagnosticTestComponentMapping models.DiagnosticTestComponentMapping
	err := c.ShouldBindJSON(&diagnosticTestComponentMapping)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	if diagnosticTestComponentMapping.DiagnosticTestId == 0 || diagnosticTestComponentMapping.DiagnosticComponentId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "DiagnosticTestId and DiagnosticComponentId is required", nil, nil)
		return
	}
	err = mc.diagnosticService.DeleteDiagnosticTestComponentMapping(diagnosticTestComponentMapping.DiagnosticTestId, diagnosticTestComponentMapping.DiagnosticComponentId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to delete diagnostic test component mapping", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diagnostic test component mapping deleted successfully", nil, nil, nil)
}

func (mc *MasterController) DeleteDiagnosticTestComponent(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	componentId, err := strconv.ParseUint(c.Param("diagnostic_test_component_id"), 10, 32)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid diagnostic test component ID", nil, err)
		return
	}

	err = mc.diagnosticService.DeleteDiagnosticTestComponent(uint64(componentId), authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to delete diagnostic test component", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diagnostic test component deleted successfully", nil, nil, nil)
}

func (mc *MasterController) GetRoleById(c *gin.Context) {
	roleIdStr := c.Param("role_id")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 32)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid role ID", nil, err)
		return
	}
	role, err := mc.roleService.GetRoleById(roleId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Role not found", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Role retrieved successfully", role, nil, nil)
}

func (mc *MasterController) UploadMasterData(c *gin.Context) {
	entity := c.Param("entity")

	file, err := c.FormFile("file")
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "file is required", nil, err)
		return
	}
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	fileReader, err := file.Open()
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "failed to open file", nil, err)
		return
	}
	defer fileReader.Close()

	count, err := mc.diseaseService.ProcessUploadFromStream(entity, authUserId, fileReader)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "failed to upload file", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Bulk upload successful", count, nil, nil)

}

func (mc *MasterController) CreateLab(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var lab models.DiagnosticLab
	lab.CreatedBy = authUserId
	if err := c.ShouldBindJSON(&lab); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input", nil, err)
		return
	}
	labInfo, err := mc.diagnosticService.CreateLab(&lab)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to create lab", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Lab created successfully", labInfo, nil, nil)
}

func (mc *MasterController) GetAllLabs(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
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

func (mc *MasterController) GetLabById(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	diagnosticlLabId, err := strconv.ParseUint(c.Param("lab_id"), 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid lab ID", nil, err)
		return
	}
	lab, err := mc.diagnosticService.GetLabById(diagnosticlLabId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Lab not found", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Lab fetched successfully", lab, nil, nil)
}

func (mc *MasterController) UpdateLab(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var lab models.DiagnosticLab
	if err := c.ShouldBindJSON(&lab); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input", nil, err)
		return
	}
	if err := mc.diagnosticService.UpdateLab(&lab, authUserId); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update lab", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Lab updated successfully", lab, nil, nil)
}

func (mc *MasterController) DeleteLab(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	diagnosticlLabId, err := strconv.ParseUint(c.Param("lab_id"), 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid lab ID", nil, err)
		return
	}
	if err := mc.diagnosticService.DeleteLab(diagnosticlLabId, authUserId); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to delete lab", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Lab deleted successfully", nil, nil, nil)
}

func (mc *MasterController) GetDiagnosticLabAuditRecord(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var labId uint64
	labIdStr := c.Query("diagnostic_lab_id")
	if labIdStr != "" {
		parsedLabId, err := strconv.ParseUint(labIdStr, 10, 32)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid diagnostic lab ID", nil, err)
			return
		}
		labId = parsedLabId
	}

	var labAuditId uint64
	if auditIdStr := c.Query("diagnostic_lab_audit_id"); auditIdStr != "" {
		auditId, err := strconv.ParseUint(auditIdStr, 10, 32)
		if err == nil {
			labAuditId = auditId
		}
	}
	page, limit, offset := utils.GetPaginationParams(c)
	message := "Diagnostic lab audit record not found"
	if labId == 0 && labAuditId == 0 {
		data, totalRecords, err := mc.diagnosticService.GetAllDiagnosticLabAuditRecords(limit, offset)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit logs", nil, err)
			return
		}
		pagination := utils.GetPagination(limit, page, offset, totalRecords)
		statusCode, message := utils.GetResponseStatusMessage(
			len(data),
			constant.AuditSuccessMessage,
			constant.AuditErrorMessage,
		)
		models.SuccessResponse(c, constant.Success, statusCode, message, data, pagination, nil)
		return
	}
	auditRecord, err := mc.diagnosticService.GetDiagnosticLabAuditRecord(labId, labAuditId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit logs", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(auditRecord),
		constant.AuditSuccessMessage,
		constant.AuditErrorMessage,
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, auditRecord, nil, nil)
}

func (mc *MasterController) AddSupportGroup(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var input models.SupportGroup
	input.CreatedBy = authUserId
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input", nil, err)
		return
	}

	err := mc.supportGroupService.AddSupportGroup(&input)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add support group", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Support group added successfully", input, nil, nil)
}

func (mc *MasterController) GetAllSupportGroups(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	data, totalRecords, err := mc.supportGroupService.GetAllSupportGroups(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch support groups", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	statusCode, msg := utils.GetResponseStatusMessage(len(data), "Support groups found", "No support groups found")
	models.SuccessResponse(c, constant.Success, statusCode, msg, data, pagination, nil)
}

func (mc *MasterController) GetSupportGroupById(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	supportGroupId, err := strconv.ParseUint(c.Param("support_group_id"), 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid support group ID", nil, err)
		return
	}
	group, err := mc.supportGroupService.GetSupportGroupById(supportGroupId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch support group", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Support group fetched successfully", group, nil, nil)
}

func (mc *MasterController) UpdateSupportGroup(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var input models.SupportGroup
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input data", nil, err)
		return
	}

	err := mc.supportGroupService.UpdateSupportGroup(&input, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update support group", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Support group updated successfully", nil, nil, nil)
}

func (mc *MasterController) DeleteSupportGroup(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	supportGroupId, err := strconv.ParseUint(c.Param("support_group_id"), 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid support group ID", nil, err)
		return
	}
	if err := mc.supportGroupService.DeleteSupportGroup(supportGroupId, authUserId); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to delete support group", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Support group deleted successfully", nil, nil, nil)
}

func (mc *MasterController) GetSupportGroupAuditRecord(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var supportGroupId uint64
	supportGroupIdStr := c.Query("support_group_id")
	if supportGroupIdStr != "" {
		parsedID, err := strconv.ParseUint(supportGroupIdStr, 10, 32)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid support group ID", nil, err)
			return
		}
		supportGroupId = parsedID
	}
	var supportGroupAuditId uint64
	if auditIdStr := c.Query("support_group_audit_id"); auditIdStr != "" {
		parsedAuditId, err := strconv.ParseUint(auditIdStr, 10, 32)
		if err == nil {
			supportGroupAuditId = parsedAuditId
		}
	}
	page, limit, offset := utils.GetPaginationParams(c)

	if supportGroupId == 0 && supportGroupAuditId == 0 {
		data, totalRecords, err := mc.supportGroupService.GetAllSupportGroupAuditRecord(limit, offset)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit logs", nil, err)
			return
		}
		pagination := utils.GetPagination(limit, page, offset, totalRecords)
		statusCode, message := utils.GetResponseStatusMessage(
			len(data),
			constant.AuditSuccessMessage,
			constant.AuditErrorMessage,
		)
		models.SuccessResponse(c, constant.Success, statusCode, message, data, pagination, nil)
		return
	}
	auditRecords, err := mc.supportGroupService.GetSupportGroupAuditRecord(supportGroupId, supportGroupAuditId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit logs", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(auditRecords),
		constant.AuditSuccessMessage,
		constant.AuditErrorMessage,
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, auditRecords, nil, nil)
}

func (mc *MasterController) AddHospital(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var input models.Hospital
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input", nil, err)
		return
	}
	input.CreatedBy = authUserId
	if err := mc.hospitalService.AddHospital(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add hospital", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Hospital added successfully", input, nil, nil)
}

func (mc *MasterController) UpdateHospital(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var input models.Hospital
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid input", nil, err)
		return
	}
	input.UpdatedBy = authUserId
	if err := mc.hospitalService.UpdateHospital(&input, authUserId); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update hospital", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Hospital updated successfully", input, nil, nil)
}

func (mc *MasterController) GetAllHospitals(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	hospitals, totalRecords, err := mc.hospitalService.GetAllHospitals(nil, limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch hospitals", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Hospitals fetched successfully", hospitals, pagination, nil)
}

func (mc *MasterController) GetHospitalById(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	hospitalIdParam := c.Param("hospital_id")
	hospitalId, err := strconv.ParseUint(hospitalIdParam, 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid Hospital ID", nil, err)
		return
	}
	hospital, err := mc.hospitalService.GetHospitalById(hospitalId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch hospital", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Hospital fetched successfully", hospital, nil, nil)
}

func (mc *MasterController) DeleteHospital(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	hospitalIDStr := c.Param("hospital_id")
	hospitalID, err := strconv.ParseInt(hospitalIDStr, 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid hospital ID", nil, err)
		return
	}
	if err := mc.hospitalService.DeleteHospitalById(hospitalID, authUserId); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to delete hospital", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Hospital deleted successfully", nil, nil, nil)
}

func (mc *MasterController) CreateService(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var service models.Service
	if err := c.ShouldBindJSON(&service); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request", nil, err)
		return
	}
	service.CreatedBy = authUserId
	if err := mc.hospitalService.AddService(&service); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to create service", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Service created successfully", service, nil, nil)
}

func (mc *MasterController) GetAllServices(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	services, err := mc.hospitalService.GetAllServices()
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to fetch services", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Services fetched successfully", services, nil, nil)
}

func (mc *MasterController) GetServiceById(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	serviceId := c.Param("service_id")
	service, err := mc.hospitalService.GetServiceById(serviceId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Service not found", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Service fetched successfully", service, nil, nil)
}

func (mc *MasterController) UpdateService(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var service models.Service
	if err := c.ShouldBindJSON(&service); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request", nil, err)
		return
	}
	if err := mc.hospitalService.UpdateService(&service, authUserId); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update service", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Service updated successfully", service, nil, nil)
}

func (mc *MasterController) DeleteService(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	serviceIdStr := c.Param("service_id")
	serviceId, err := strconv.ParseUint(serviceIdStr, 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid service ID", nil, err)
		return
	}
	if err := mc.hospitalService.DeleteService(serviceId, authUserId); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to delete service", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Service deleted successfully", nil, nil, nil)
}

func (mc *MasterController) GetHospitalAuditRecord(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var hospitalId uint64
	hospitalIdStr := c.Query("hospital_id")
	if hospitalIdStr != "" {
		parsedHospitalId, err := strconv.ParseUint(hospitalIdStr, 10, 32)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid hospital ID", nil, err)
			return
		}
		hospitalId = parsedHospitalId
	}
	var hospitalAuditId uint64
	if auditIdStr := c.Query("hospital_audit_id"); auditIdStr != "" {
		auditId, err := strconv.ParseUint(auditIdStr, 10, 32)
		if err == nil {
			hospitalAuditId = auditId
		}
	}
	page, limit, offset := utils.GetPaginationParams(c)
	message := "Hospital audit record not found"
	if hospitalId == 0 && hospitalAuditId == 0 {
		data, totalRecords, err := mc.hospitalService.GetAllHospitalAuditRecord(limit, offset)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit logs", nil, err)
			return
		}
		pagination := utils.GetPagination(limit, page, offset, totalRecords)
		statusCode, message := utils.GetResponseStatusMessage(
			len(data),
			constant.AuditSuccessMessage,
			constant.AuditErrorMessage,
		)
		models.SuccessResponse(c, constant.Success, statusCode, message, data, pagination, nil)
		return
	}
	auditRecord, err := mc.hospitalService.GetHospitalAuditRecord(hospitalId, hospitalAuditId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit logs", nil, err)
		return
	}

	statusCode, message := utils.GetResponseStatusMessage(
		len(auditRecord),
		constant.AuditSuccessMessage,
		constant.AuditErrorMessage,
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, auditRecord, nil, nil)
}

func (mc *MasterController) GetServiceAuditRecord(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var serviceId uint64
	serviceIdStr := c.Query("service_id")
	if serviceIdStr != "" {
		parsedServiceId, err := strconv.ParseUint(serviceIdStr, 10, 32)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid service ID", nil, err)
			return
		}
		serviceId = parsedServiceId
	}

	var serviceAuditId uint64
	if auditIdStr := c.Query("service_audit_id"); auditIdStr != "" {
		auditId, err := strconv.ParseUint(auditIdStr, 10, 32)
		if err == nil {
			serviceAuditId = auditId
		}
	}
	page, limit, offset := utils.GetPaginationParams(c)
	message := "Service audit record not found"
	if serviceId == 0 && serviceAuditId == 0 {
		data, totalRecords, err := mc.hospitalService.GetAllServiceAuditRecord(limit, offset)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit logs", nil, err)
			return
		}
		pagination := utils.GetPagination(limit, page, offset, totalRecords)
		statusCode, message := utils.GetResponseStatusMessage(
			len(data),
			constant.AuditSuccessMessage,
			constant.AuditErrorMessage,
		)
		models.SuccessResponse(c, constant.Success, statusCode, message, data, pagination, nil)
		return
	}
	auditRecord, err := mc.hospitalService.GetServiceAuditRecord(serviceId, serviceAuditId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve audit logs", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(auditRecord),
		constant.AuditSuccessMessage,
		constant.AuditErrorMessage,
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, auditRecord, nil, nil)
}

func (mc *MasterController) AddServiceMapping(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var serviceMapping models.ServiceMapping
	if err := c.ShouldBindJSON(&serviceMapping); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	if serviceMapping.ServiceProviderId == 0 || serviceMapping.ServiceId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "ServiceProviderId and ServiceId are required", nil, nil)
		return
	}
	serviceMapping.CreatedBy = authUserId
	err := mc.hospitalService.AddServiceMapping(serviceMapping)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Hospital-service mapping already exists.", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Hospital-service mapping added successfully.", nil, nil, nil)
}

func (mc *MasterController) AddTestReferenceRange(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var input models.DiagnosticTestReferenceRange
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}

	input.CreatedBy = authUserId
	err := mc.diagnosticService.AddTestReferenceRange(&input)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Unable to add reference range", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Diagnostic test reference range added successfully.", nil, nil, nil)
}

func (mc *MasterController) UpdateTestReferenceRange(c *gin.Context) {
	authUserId, _, err1 := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err1 != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err1.Error(), nil, err1)
		return
	}
	var input models.DiagnosticTestReferenceRange
	if err := c.ShouldBindJSON(&input); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	err := mc.diagnosticService.UpdateTestReferenceRange(&input, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update reference range", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Reference range updated successfully.", nil, nil, nil)
}

func (mc *MasterController) DeleteTestReferenceRange(c *gin.Context) {
	authUserId, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}

	idParam := c.Param("test_reference_range_id")
	testReferenceRangeId, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid ID", nil, err)
		return
	}

	err = mc.diagnosticService.DeleteTestReferenceRange(testReferenceRangeId, authUserId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Error deleting reference range", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Reference range deleted successfully.", nil, nil, nil)
}

func (mc *MasterController) ViewTestReferenceRange(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	idParam := c.Param("test_reference_range_id")
	testReferenceRangeId, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid refernece range Id", nil, err)
		return
	}

	ranges, err := mc.diagnosticService.ViewTestReferenceRange(testReferenceRangeId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Error occurs while fetching reference ranges", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Reference ranges fetched successfully.", ranges, nil, nil)
}

func (mc *MasterController) GetAllTestReferenceRange(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	page, limit, offset := utils.GetPaginationParams(c)
	isDeleted, ok := utils.GetQueryIntParam(c, "is_deleted", 0)
	if !ok {
		log.Println("GetAllTestRefRangeView is deleted status not provided : ", isDeleted)
	}
	referenceRanges, totalRecords, err := mc.diagnosticService.GetAllTestRefRangeView(limit, offset, isDeleted)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Unable to fetch reference range.", nil, err)
		return
	}
	statusCode, message := utils.GetResponseStatusMessage(
		len(referenceRanges),
		"Test reference ranges retrieved successfully",
		"Test reference ranges not found",
	)
	pagination := utils.GetPagination(limit, page, offset, totalRecords)
	models.SuccessResponse(c, constant.Success, statusCode, message, referenceRanges, pagination, nil)
}

func (mc *MasterController) GetTestReferenceRangeAuditRecord(c *gin.Context) {
	_, _, err := utils.GetUserIDFromContext(c, mc.userService.GetUserIdBySUB)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, err.Error(), nil, err)
		return
	}
	var testReferenceRangeId uint64
	testReferenceRangeStr := c.Query("test_reference_range_id")
	if testReferenceRangeStr != "" {
		parsedTestReferenceRangeId, err := strconv.ParseUint(testReferenceRangeStr, 10, 64)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid test reference range ID", nil, err)
			return
		}
		testReferenceRangeId = parsedTestReferenceRangeId
	}

	var testReferenceAuditId uint64
	if auditIdStr := c.Query("test_reference_range_audit_id"); auditIdStr != "" {
		auditId, err := strconv.ParseUint(auditIdStr, 10, 64)
		if err == nil {
			testReferenceAuditId = auditId
		}
	}
	page, limit, offset := utils.GetPaginationParams(c)
	if testReferenceRangeId == 0 && testReferenceAuditId == 0 {
		data, totalRecords, err := mc.diagnosticService.GetTestReferenceRangeAuditRecord(0, 0, limit, offset)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve test reference range audit logs", nil, err)
			return
		}
		pagination := utils.GetPagination(limit, page, offset, totalRecords)
		statusCode, message := utils.GetResponseStatusMessage(
			len(data),
			"Test reference range audit records retrieved successfully",
			"Test reference range audit records not found",
		)
		models.SuccessResponse(c, constant.Success, statusCode, message, data, pagination, nil)
		return
	}

	auditRecord, _, err := mc.diagnosticService.GetTestReferenceRangeAuditRecord(testReferenceRangeId, testReferenceAuditId, limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve test reference range audit logs", nil, err)
		return
	}

	statusCode, message := utils.GetResponseStatusMessage(
		len(auditRecord),
		constant.AuditSuccessMessage,
		constant.AuditErrorMessage,
	)
	models.SuccessResponse(c, constant.Success, statusCode, message, auditRecord, nil, nil)
}
