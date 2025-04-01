package controller

import (
	"biostat/constant"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MasterController struct {
	allergyService    service.AllergyService
	diseaseService    service.DiseaseService
	causeService      service.CauseService
	symptomService    service.SymptomService
	medicationService service.MedicationService
	dietService       service.DietService
	exerciseService   service.ExerciseService
	diagnosticService service.DiagnosticService
}

func NewMasterController(allergyService service.AllergyService, diseaseService service.DiseaseService,
	causeService service.CauseService, symptomService service.SymptomService, medicationService service.MedicationService,
	dietService service.DietService, exerciseService service.ExerciseService, diagnosticService service.DiagnosticService) *MasterController {
	return &MasterController{allergyService: allergyService,
		diseaseService:    diseaseService,
		causeService:      causeService,
		symptomService:    symptomService,
		medicationService: medicationService,
		dietService:       dietService,
		exerciseService:   exerciseService,
		diagnosticService: diagnosticService,
	}
}

func (mc *MasterController) CreateDisease(c *gin.Context) {
	var disease models.Disease
	if err := c.ShouldBindJSON(&disease); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request payload", nil, err)
		return
	}
	err := mc.diseaseService.CreateDisease(&disease)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to create disease", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Disease created successfully", nil, nil, nil)
}

func (mc *MasterController) GetDiseaseInfo(c *gin.Context) {
	DiseaseId, err := strconv.ParseUint(c.Param("disease_id"), 10, 32)
	page, limit, offset := utils.GetPaginationParams(c)
	message := "Diseases not found"
	fmt.Println("DiseaseId ", DiseaseId)
	if err != nil || DiseaseId == 0 {
		data, totalRecords, err := mc.diseaseService.GetAllDiseases(page, limit)
		if err != nil {
			models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve diseases", nil, err)
			return
		}
		pagination := utils.GetPagination(limit, page, offset, totalRecords)
		log.Println("Disease data ", data)
		if len(data) > 0 {
			message = "All Diseases retrieved successfully"
		}
		models.SuccessResponse(c, constant.Success, http.StatusOK, message, data, pagination, nil)
		return
	}
	diseases, err := mc.diseaseService.GetDiseases(uint(DiseaseId))
	if err == nil {
		message = "Diseases retrieved successfully"
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, message, diseases, nil, nil)
}

func (mc *MasterController) GetDiseaseProfile(c *gin.Context) {

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
	diseaseProfileId := c.Param("disease_profile_id")

	diseaseProfile, err := mc.diseaseService.GetDiseaseProfileById(diseaseProfileId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Disease profile not found", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Disease profile retrieved successfully", diseaseProfile, nil, nil)
}

// Get all causes with pagination
func (mc *MasterController) GetAllCauses(c *gin.Context) {
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

func (mc *MasterController) AddDiseaseCause(c *gin.Context) {
	var cause models.Cause

	if err := c.ShouldBindJSON(&cause); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid cause input", nil, err)
		return
	}

	err := mc.causeService.AddDiseaseCause(&cause)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add cause", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Cause added successfully", cause, nil, nil)
}

func (mc *MasterController) UpdateDiseaseCause(c *gin.Context) {
	var cause models.Cause

	if err := c.ShouldBindJSON(&cause); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid disease cause input", nil, err)
		return
	}

	if cause.CauseId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Cause Id is required", nil, nil)
		return
	}

	err := mc.causeService.UpdateCause(&cause)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update cause", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Cause updated successfully", cause, nil, nil)
}

func (mc *MasterController) GetAllSymptom(c *gin.Context) {
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
	var symptom models.Symptom

	if err := c.ShouldBindJSON(&symptom); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid symptom input", nil, err)
		return
	}

	err := mc.symptomService.AddDiseaseSymptom(&symptom)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add symptom", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "symptom added successfully", symptom, nil, nil)
}

func (mc *MasterController) UpdateDiseaseSymptom(c *gin.Context) {
	var symptom models.Symptom

	if err := c.ShouldBindJSON(&symptom); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid disease symptom input", nil, err)
		return
	}

	if symptom.SymptomId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "symptom Id is required", nil, nil)
		return
	}

	err := mc.symptomService.UpdateSymptom(&symptom)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update symptom", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "symptom updated successfully", symptom, nil, nil)
}

func (mc *MasterController) AddDietPlanTemplate(c *gin.Context) {
	var dietPlan models.DietPlanTemplate
	if err := c.ShouldBindJSON(&dietPlan); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid diet plan input", nil, err)
		return
	}

	if err := mc.dietService.CreateDietPlanTemplate(&dietPlan); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add diet plan", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Diet plan added successfully", dietPlan, nil, nil)
}
func (mc *MasterController) GetAllDietPlanTemplates(c *gin.Context) {
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
	dietPlanTemplateId := c.Param("diet_id")
	var dietPlan models.DietPlanTemplate
	dietPlan, err := mc.dietService.GetDietPlanById(dietPlanTemplateId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Diet plan not found", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diet plan retrieved successfully", dietPlan, nil, nil)
}

func (mc *MasterController) UpdateDietPlanTemplate(c *gin.Context) {
	dietPlanTemplateId := c.Param("diet_id")
	var dietPlan models.DietPlanTemplate

	if err := c.ShouldBindJSON(&dietPlan); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid diet plan input", nil, err)
		return
	}

	if err := mc.dietService.UpdateDietPlanTemplate(dietPlanTemplateId, &dietPlan); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update diet plan", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diet plan updated successfully", dietPlan, nil, nil)
}

func (mc *MasterController) AddExercise(c *gin.Context) {
	var exercise models.Exercise
	if err := c.ShouldBindJSON(&exercise); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid exercise input", nil, err)
		return
	}

	if err := mc.exerciseService.CreateExercise(&exercise); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add exercise", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Exercise added successfully", exercise, nil, nil)
}

func (mc *MasterController) GetAllExercises(c *gin.Context) {
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

func (mc *MasterController) GetExerciseByID(c *gin.Context) {
	id := c.Param("exercise_id")
	var exercise models.Exercise

	exercise, err := mc.exerciseService.GetExerciseByID(id)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Exercise not found", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Exercise retrieved successfully", exercise, nil, nil)
}

func (mc *MasterController) UpdateExercise(c *gin.Context) {
	id := c.Param("exercise_id")
	var exercise models.Exercise

	if err := c.ShouldBindJSON(&exercise); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid exercise input", nil, err)
		return
	}

	if err := mc.exerciseService.UpdateExercise(id, &exercise); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update exercise", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Exercise updated successfully", exercise, nil, nil)
}

func (mc *MasterController) GetAllergyRestrictions(c *gin.Context) {
	allergies, err := mc.allergyService.GetAllergies()
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve allergies", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Allergies retrieved successfully", allergies, nil, nil)
}

func (mc *MasterController) GetAllMedication(c *gin.Context) {
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

func (mc *MasterController) UpdateMedication(c *gin.Context) {
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

func (dc *MasterController) GetDiagnosticTests(c *gin.Context) {
	page, limit, offset := utils.GetPaginationParams(c)
	diagnosticTest, totalRecord, err := dc.diagnosticService.GetDiagnosticTests(limit, offset)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve diagnostic tests", nil, err)
		return
	}
	pagination := utils.GetPagination(limit, page, offset, totalRecord)
	message := "Diagnostic tests not found"
	if len(diagnosticTest) > 0 {
		message = "Diagnostic tests retrieved successfully"
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, message, diagnosticTest, pagination, nil)
}

func (dc *MasterController) CreateDiagnosticTest(c *gin.Context) {
	var diagnosticTest models.DiagnosticTest
	err := c.ShouldBindJSON(&diagnosticTest)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	diagnosticTestRes, err := dc.diagnosticService.CreateDiagnosticTest(&diagnosticTest, "admin")
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to create diagnostic test", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Diagnostic test created successfully", diagnosticTestRes, nil, nil)
}

func (dc *MasterController) UpdateDiagnosticTest(c *gin.Context) {
	var diagnosticTest models.DiagnosticTest
	err := c.ShouldBindJSON(&diagnosticTest)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	if diagnosticTest.DiagnosticTestId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "DiagnosticTestId is required", nil, nil)
		return
	}
	diagnosticTestRes, err := dc.diagnosticService.UpdateDiagnosticTest(&diagnosticTest, "admin")
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update diagnostic test", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diagnostic test updated successfully", diagnosticTestRes, nil, nil)
}

func (dc *MasterController) GetSingleDiagnosticTest(c *gin.Context) {
	diagnosticTestId := utils.GetParamAsInt(c, "diagnosticTestId")
	if diagnosticTestId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "DiagnosticTestId is required", nil, nil)
		return
	}
	diagnosticTest, err := dc.diagnosticService.GetSingleDiagnosticTest(diagnosticTestId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve diagnostic test", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diagnostic test retrieved successfully", diagnosticTest, nil, nil)
}

func (dc *MasterController) DeleteDiagnosticTest(c *gin.Context) {
	diagnosticTestId := utils.GetParamAsInt(c, "diagnosticTestId")
	if diagnosticTestId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "DiagnosticTestId is required", nil, nil)
		return
	}
	err := dc.diagnosticService.DeleteDiagnosticTest(diagnosticTestId, "admin")
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to delete diagnostic test", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diagnostic test deleted successfully", nil, nil, nil)

}

func (dc *MasterController) GetAllDiagnosticComponents(c *gin.Context) {
	page, limit, offset := utils.GetPaginationParams(c)
	diagnosticTestComponent, totalRecord, err := dc.diagnosticService.GetAllDiagnosticComponents(limit, offset)
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

func (dc *MasterController) CreateDiagnosticComponent(c *gin.Context) {
	var diagnosticComponent models.DiagnosticTestComponent
	err := c.ShouldBindJSON(&diagnosticComponent)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	diagnosticComponentRes, err := dc.diagnosticService.CreateDiagnosticComponent(&diagnosticComponent)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to create diagnostic component", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Diagnostic component created successfully", diagnosticComponentRes, nil, nil)
}

func (dc *MasterController) UpdateDiagnosticComponent(c *gin.Context) {
	var diagnosticComponent models.DiagnosticTestComponent
	err := c.ShouldBindJSON(&diagnosticComponent)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	if diagnosticComponent.DiagnosticTestComponentId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "DiagnosticTestComponentId is required", nil, nil)
		return
	}
	diagnosticComponentRes, err := dc.diagnosticService.UpdateDiagnosticComponent(&diagnosticComponent)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update diagnostic component", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diagnostic component updated successfully", diagnosticComponentRes, nil, nil)
}

func (dc *MasterController) GetSingleDiagnosticComponent(c *gin.Context) {
	diagnosticComponentId := utils.GetParamAsInt(c, "diagnosticComponentId")
	if diagnosticComponentId == 0 {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "DiagnosticComponentId is required", nil, nil)
		return
	}
	diagnosticComponent, err := dc.diagnosticService.GetSingleDiagnosticComponent(diagnosticComponentId)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to retrieve diagnostic component", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diagnostic component retrieved successfully", diagnosticComponent, nil, nil)
}

func (dc *MasterController) GetAllDiagnosticTestComponentMappings(c *gin.Context) {
	page, limit, offset := utils.GetPaginationParams(c)
	diagnosticTestComponentMapping, totalRecord, err := dc.diagnosticService.GetAllDiagnosticTestComponentMappings(limit, offset)
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

func (dc *MasterController) CreateDiagnosticTestComponentMapping(c *gin.Context) {
	var diagnosticTestComponentMapping models.DiagnosticTestComponentMapping
	err := c.ShouldBindJSON(&diagnosticTestComponentMapping)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid request body", nil, err)
		return
	}
	diagnosticTestComponentMappingRes, err := dc.diagnosticService.CreateDiagnosticTestComponentMapping(&diagnosticTestComponentMapping)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to create diagnostic test component mapping", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Diagnostic test component mapping created successfully", diagnosticTestComponentMappingRes, nil, nil)
}

func (dc *MasterController) UpdateDiagnosticTestComponentMapping(c *gin.Context) {
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
	diagnosticTestComponentMappingRes, err := dc.diagnosticService.UpdateDiagnosticTestComponentMapping(&diagnosticTestComponentMapping)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update diagnostic test component mapping", nil, err)
		return
	}
	models.SuccessResponse(c, constant.Success, http.StatusOK, "Diagnostic test component mapping updated successfully", diagnosticTestComponentMappingRes, nil, nil)
}
