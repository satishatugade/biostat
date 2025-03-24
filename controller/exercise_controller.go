package controller

import (
	"biostat/constant"
	"biostat/models"
	"biostat/service"
	"biostat/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ExerciseController struct {
	exerciseService service.ExerciseService
}

func NewExerciseController(exerciseService service.ExerciseService) *ExerciseController {
	return &ExerciseController{exerciseService: exerciseService}
}

func (ec *ExerciseController) AddExercise(c *gin.Context) {
	var exercise models.Exercise
	if err := c.ShouldBindJSON(&exercise); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid exercise input", nil, err)
		return
	}

	if err := ec.exerciseService.CreateExercise(&exercise); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to add exercise", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusCreated, "Exercise added successfully", exercise, nil, nil)
}

func (ec *ExerciseController) GetAllExercises(c *gin.Context) {
	page, limit, offset := utils.GetPaginationParams(c)
	exercises, totalRecords, err := ec.exerciseService.GetExercises(limit, offset)
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

func (ec *ExerciseController) GetExerciseByID(c *gin.Context) {
	id := c.Param("exercise_id")
	var exercise models.Exercise

	exercise, err := ec.exerciseService.GetExerciseByID(id)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusNotFound, "Exercise not found", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Exercise retrieved successfully", exercise, nil, nil)
}

func (ec *ExerciseController) UpdateExercise(c *gin.Context) {
	id := c.Param("exercise_id")
	var exercise models.Exercise

	if err := c.ShouldBindJSON(&exercise); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid exercise input", nil, err)
		return
	}

	if err := ec.exerciseService.UpdateExercise(id, &exercise); err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusInternalServerError, "Failed to update exercise", nil, err)
		return
	}

	models.SuccessResponse(c, constant.Success, http.StatusOK, "Exercise updated successfully", exercise, nil, nil)
}
