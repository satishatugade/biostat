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
