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

func GetParamAsUInt(c *gin.Context, param string) uint64 {
	valueStr := c.Param(param)
	value, err := strconv.ParseUint(valueStr, 10, 64)
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
		return "Could not retrieve user id", false
	}

	subStr, ok := sub.(string)
	if !ok {
		return "user id not a valid string", false
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

func MapSystemUserToPatient(user *models.SystemUser_) *models.Patient {
	return &models.Patient{
		PatientId:          user.UserId,
		FirstName:          user.FirstName,
		LastName:           user.LastName,
		DateOfBirth:        user.DateOfBirth.String(),
		Gender:             user.Gender,
		MobileNo:           user.MobileNo,
		Address:            user.Address,
		EmergencyContact:   user.EmergencyContact,
		AbhaNumber:         user.AbhaNumber,
		BloodGroup:         user.BloodGroup,
		Nationality:        user.Nationality,
		CitizenshipStatus:  user.CitizenshipStatus,
		PassportNumber:     user.PassportNumber,
		CountryOfResidence: user.CountryOfResidence,
		IsIndianOrigin:     user.IsIndianOrigin,
		Email:              user.Email,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
	}
}

func MapUserToRoleSchema(user models.SystemUser_, roleName string) interface{} {
	role := strings.ToLower(roleName)
	switch role {
	case "nurse":
		return models.Nurse{
			NurseId:           user.UserId,
			FirstName:         user.FirstName,
			LastName:          user.LastName,
			Gender:            user.Gender,
			MobileNo:          user.MobileNo,
			Email:             user.Email,
			Specialty:         user.Specialty,
			LicenseNumber:     user.LicenseNumber,
			ClinicName:        user.ClinicName,
			ClinicAddress:     user.ClinicAddress,
			YearsOfExperience: derefInt(user.YearsOfExperience),
			ConsultationFee:   derefFloat(user.ConsultationFee),
			WorkingHours:      user.WorkingHours,
			CreatedAt:         user.CreatedAt,
			UpdatedAt:         user.UpdatedAt,
		}
	case "doctor":
		return models.Doctor{
			DoctorId:          user.UserId,
			FirstName:         user.FirstName,
			LastName:          user.LastName,
			Specialty:         user.Specialty,
			Gender:            user.Gender,
			MobileNo:          user.MobileNo,
			LicenseNumber:     user.LicenseNumber,
			ClinicName:        user.ClinicName,
			ClinicAddress:     user.ClinicAddress,
			Email:             user.Email,
			YearsOfExperience: derefInt(user.YearsOfExperience),
			ConsultationFee:   derefFloat(user.ConsultationFee),
			WorkingHours:      user.WorkingHours,
			CreatedAt:         user.CreatedAt,
			UpdatedAt:         user.UpdatedAt,
		}
	case "admin":
		return models.Admin{
			UserId:    user.UserId,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			RoleName:  user.RoleName,
			MobileNo:  user.MobileNo,
			Email:     user.Email,
		}
	default:
		return models.Patient{
			PatientId:          user.UserId,
			FirstName:          user.FirstName,
			LastName:           user.LastName,
			DateOfBirth:        user.DateOfBirth.String(),
			Gender:             user.Gender,
			MobileNo:           user.MobileNo,
			Address:            user.Address,
			EmergencyContact:   user.EmergencyContact,
			AbhaNumber:         user.AbhaNumber,
			BloodGroup:         user.BloodGroup,
			Nationality:        user.Nationality,
			CitizenshipStatus:  user.CitizenshipStatus,
			PassportNumber:     user.PassportNumber,
			CountryOfResidence: user.CountryOfResidence,
			IsIndianOrigin:     user.IsIndianOrigin,
			Email:              user.Email,
			CreatedAt:          user.CreatedAt,
			UpdatedAt:          user.UpdatedAt,
		}
	}
}

func MapUserToPublicProviderInfo(user models.SystemUser_, roleName string) interface{} {
	role := strings.ToLower(roleName)

	switch role {
	case "doctor":
		return models.DoctorInfo{
			DoctorId:          user.UserId,
			FirstName:         user.FirstName,
			LastName:          user.LastName,
			Specialty:         user.Specialty,
			ClinicName:        user.ClinicName,
			Gender:            user.Gender,
			MobileNo:          user.MobileNo,
			ClinicAddress:     user.ClinicAddress,
			YearsOfExperience: derefInt(user.YearsOfExperience),
			WorkingHours:      user.WorkingHours,
		}
	case "nurse":
		return models.Nurse{
			NurseId:    user.UserId,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			Specialty:  user.Specialty,
			ClinicName: user.ClinicName,
		}
	case "patient":
		return models.Patient{
			PatientId:  user.UserId,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			BloodGroup: user.BloodGroup,
		}
	default:
		return map[string]string{
			"message": "unsupported provider type",
		}
	}
}

func derefInt(ptr *int) int {
	if ptr != nil {
		return *ptr
	}
	return 0
}

func derefFloat(ptr *float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return 0.0
}

func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
