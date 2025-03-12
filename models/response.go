package models

import (
	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Limit      int   `json:"limit"`
	Page       int   `json:"page"`
	Offset     int   `json:"offset"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"totalPages"`
}

type APIResponse struct {
	Status     string      `json:"status"`
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Error      string      `json:"error"`
	Content    interface{} `json:"content"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

func SuccessResponse(c *gin.Context, status string, statusCode int, message string, content interface{}, pagination *Pagination, err error) {
	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
	}
	c.JSON(statusCode, APIResponse{
		Status:     status,
		StatusCode: statusCode,
		Message:    message,
		Content:    content,
		Pagination: pagination,
		Error:      errorMessage,
	})
}

func ErrorResponse(c *gin.Context, status string, statusCode int, message string, content interface{}, err error) {
	var errorMessage string
	if err != nil {
		errorMessage = err.Error()
	}
	c.JSON(statusCode, APIResponse{
		Status:     status,
		StatusCode: statusCode,
		Message:    message,
		Content:    nil,
		Error:      errorMessage,
	})
}
