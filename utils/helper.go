package utils

import (
	"biostat/constant"
	"biostat/models"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func GetPaginationParams(c *gin.Context) (int, int, int) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit // Calculate offset

	return page, limit, offset
}

func GetPagination(limit int, page int, offset int, totalRecords int64) *models.Pagination {
	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))

	return &models.Pagination{
		Limit:      limit,
		Page:       page,
		Offset:     offset,
		Total:      totalRecords,
		TotalPages: int64(totalPages),
	}
}

func GetResponseStatusMessage(dataLength int, successMsg, notFoundMsg string) (int, string) {
	if dataLength > 0 {
		return http.StatusOK, successMsg
	}
	return http.StatusNotFound, notFoundMsg
}
func GetParamAsInt(c *gin.Context, param string) int {
	value, err := strconv.Atoi(c.Param(param))
	if err != nil {
		return 0
	}
	return value
}

func ParseUintParam(c *gin.Context, paramName string) (uint64, bool) {
	paramValue := c.Param(paramName)
	if paramValue == "" {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, paramName+" is required", nil, nil)
		return 0, false
	}

	parsedValue, err := strconv.ParseUint(paramValue, 10, 64)
	if err != nil {
		models.ErrorResponse(c, constant.Failure, http.StatusBadRequest, "Invalid "+paramName, nil, err)
		return 0, false
	}

	return parsedValue, true
}

func GenerateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8

	rand.Seed(time.Now().UnixNano())
	password := make([]byte, length)

	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}

	return string(password)
}

func GetUserDataContext(c *gin.Context) (string, bool) {
	sub, exists := c.Get("sub")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve user id"})
		return "", false
	}

	subStr, ok := sub.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user id not a valid string"})
		return "", false
	}

	return subStr, true
}

// UserId, exists := utils.GetUserDataContext(c)
// if !exists {
// 	return
// }

// user, err := client.GetUserByID(ctx, tokenStr, utils.KeycloakRealm, UserId)
// if err != nil {
// 	fmt.Println("User data ", user)
// 	log.Println("User data ", user)
// 	log.Println("User data email ", user.Email)
// 	return
// }

// func Logger(msg string) {
// 	if log.GetLevel() == log.DebugLevel {
// 		log.Debug(msg)
// 	} else {
// 		log.Info(msg)
// 	}
// }

func ParseExcelFromReader[T any](reader io.Reader) ([]T, error) {
	var results []T

	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, err
	}

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil || len(rows) < 2 {
		return nil, fmt.Errorf("invalid or empty sheet")
	}

	headers := rows[0]
	for _, row := range rows[1:] {
		var item T
		val := reflect.ValueOf(&item).Elem()

		for i, cell := range row {
			if i >= len(headers) {
				continue
			}

			fieldName := strings.TrimSpace(headers[i])
			for j := 0; j < val.NumField(); j++ {
				field := val.Type().Field(j)
				tag := field.Tag.Get("json")
				if tag == fieldName && val.Field(j).CanSet() {
					val.Field(j).SetString(cell)
				}
			}
		}
		results = append(results, item)
	}

	return results, nil
}
