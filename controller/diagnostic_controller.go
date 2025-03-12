package controller

import (
	"biostat/constant"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DiagnosticController struct {
	diagnosticService service.DiagnosticService
}

func NewDiagnosticController(diagnosticService service.DiagnosticService) *DiagnosticController {
	return &DiagnosticController{diagnosticService: diagnosticService}
}

func (dc *DiagnosticController) GetDiagnosticTests(c *gin.Context) {
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

func (dc *DiagnosticController) CreateDiagnosticTest(c *gin.Context) {
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

func (dc *DiagnosticController) UpdateDiagnosticTest(c *gin.Context) {
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

func (dc *DiagnosticController) GetSingleDiagnosticTest(c *gin.Context) {
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

func (dc *DiagnosticController) DeleteDiagnosticTest(c *gin.Context) {
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

func (dc *DiagnosticController) GetAllDiagnosticComponents(c *gin.Context) {
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

func (dc *DiagnosticController) CreateDiagnosticComponent(c *gin.Context) {
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

func (dc *DiagnosticController) UpdateDiagnosticComponent(c *gin.Context) {
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

func (dc *DiagnosticController) GetSingleDiagnosticComponent(c *gin.Context) {
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

func (dc *DiagnosticController) GetAllDiagnosticTestComponentMappings(c *gin.Context) {
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

func (dc *DiagnosticController) CreateDiagnosticTestComponentMapping(c *gin.Context) {
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

func (dc *DiagnosticController) UpdateDiagnosticTestComponentMapping(c *gin.Context) {
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
