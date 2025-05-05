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

func GetQueryIntParam(c *gin.Context, paramName string, defaultValue int) (uint64, bool) {
	paramStr := c.DefaultQuery(paramName, strconv.Itoa(defaultValue))
	paramValue, err := strconv.ParseUint(paramStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return paramValue, true
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

func MapUsersToSchema(users []models.SystemUser_, roleName string) []interface{} {
	var mappedUsers []interface{}

	role := strings.ToLower(roleName)
	for _, user := range users {
		switch role {
		case "nurse":
			mappedUsers = append(mappedUsers, models.Nurse{
				NurseId:       user.UserId,
				FirstName:     user.FirstName,
				LastName:      user.LastName,
				Gender:        user.Gender,
				MobileNo:      user.MobileNo,
				Email:         user.Email,
				Speciality:    user.Speciality,
				LicenseNumber: user.LicenseNumber,
				ClinicName:    user.ClinicName,
				ClinicAddress: user.ClinicAddress,
				UserAddress: models.AddressMaster{
					AddressId:    user.AddressMapping.AddressId,
					AddressLine1: user.AddressMapping.Address.AddressLine1,
					AddressLine2: user.AddressMapping.Address.AddressLine2,
					Landmark:     user.AddressMapping.Address.Landmark,
					City:         user.AddressMapping.Address.City,
					State:        user.AddressMapping.Address.State,
					Country:      user.AddressMapping.Address.Country,
					PostalCode:   user.AddressMapping.Address.PostalCode,
					Latitude:     user.AddressMapping.Address.Latitude,
					Longitude:    user.AddressMapping.Address.Longitude,
				},
				YearsOfExperience: derefInt(user.YearsOfExperience),
				ConsultationFee:   derefFloat(user.ConsultationFee),
				WorkingHours:      user.WorkingHours,
				CreatedAt:         user.CreatedAt,
				UpdatedAt:         user.UpdatedAt,
			})
		case "doctor":
			mappedUsers = append(mappedUsers, models.Doctor{
				DoctorId:      user.UserId,
				FirstName:     user.FirstName,
				LastName:      user.LastName,
				Speciality:    user.Speciality,
				Gender:        user.Gender,
				MobileNo:      user.MobileNo,
				LicenseNumber: user.LicenseNumber,
				ClinicName:    user.ClinicName,
				ClinicAddress: user.ClinicAddress,
				UserAddress: models.AddressMaster{
					AddressId:    user.AddressMapping.AddressId,
					AddressLine1: user.AddressMapping.Address.AddressLine1,
					AddressLine2: user.AddressMapping.Address.AddressLine2,
					Landmark:     user.AddressMapping.Address.Landmark,
					City:         user.AddressMapping.Address.City,
					State:        user.AddressMapping.Address.State,
					Country:      user.AddressMapping.Address.Country,
					PostalCode:   user.AddressMapping.Address.PostalCode,
					Latitude:     user.AddressMapping.Address.Latitude,
					Longitude:    user.AddressMapping.Address.Longitude,
				},
				Email:             user.Email,
				YearsOfExperience: derefInt(user.YearsOfExperience),
				ConsultationFee:   derefFloat(user.ConsultationFee),
				WorkingHours:      user.WorkingHours,
				CreatedAt:         user.CreatedAt,
				UpdatedAt:         user.UpdatedAt,
			})
		case "admin":
			mappedUsers = append(mappedUsers, models.Admin{
				UserId:    user.UserId,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				RoleName:  user.RoleName,
				MobileNo:  user.MobileNo,
				Email:     user.Email,
			})
		default:
			mappedUsers = append(mappedUsers, models.Patient{
				PatientId:   user.UserId,
				FirstName:   user.FirstName,
				LastName:    user.LastName,
				DateOfBirth: user.DateOfBirth.String(),
				Gender:      user.Gender,
				MobileNo:    user.MobileNo,
				Address:     user.Address,
				UserAddress: models.AddressMaster{
					AddressId:    user.AddressMapping.AddressId,
					AddressLine1: user.AddressMapping.Address.AddressLine1,
					AddressLine2: user.AddressMapping.Address.AddressLine2,
					Landmark:     user.AddressMapping.Address.Landmark,
					City:         user.AddressMapping.Address.City,
					State:        user.AddressMapping.Address.State,
					Country:      user.AddressMapping.Address.Country,
					PostalCode:   user.AddressMapping.Address.PostalCode,
					Latitude:     user.AddressMapping.Address.Latitude,
					Longitude:    user.AddressMapping.Address.Longitude,
				},
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
			})
		}
	}

	return mappedUsers
}

func MapUserToRoleSchema(user models.SystemUser_, roleName string) interface{} {
	role := strings.ToLower(roleName)
	switch role {
	case "nurse":
		return models.Nurse{
			NurseId:       user.UserId,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			Gender:        user.Gender,
			MobileNo:      user.MobileNo,
			Email:         user.Email,
			Speciality:    user.Speciality,
			LicenseNumber: user.LicenseNumber,
			ClinicName:    user.ClinicName,
			ClinicAddress: user.ClinicAddress,
			UserAddress: models.AddressMaster{
				AddressId:    user.AddressMapping.AddressId,
				AddressLine1: user.AddressMapping.Address.AddressLine1,
				AddressLine2: user.AddressMapping.Address.AddressLine2,
				Landmark:     user.AddressMapping.Address.Landmark,
				City:         user.AddressMapping.Address.City,
				State:        user.AddressMapping.Address.State,
				Country:      user.AddressMapping.Address.Country,
				PostalCode:   user.AddressMapping.Address.PostalCode,
				Latitude:     user.AddressMapping.Address.Latitude,
				Longitude:    user.AddressMapping.Address.Longitude,
			},
			YearsOfExperience: derefInt(user.YearsOfExperience),
			ConsultationFee:   derefFloat(user.ConsultationFee),
			WorkingHours:      user.WorkingHours,
			CreatedAt:         user.CreatedAt,
			UpdatedAt:         user.UpdatedAt,
		}
	case "doctor":
		return models.Doctor{
			DoctorId:      user.UserId,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			Speciality:    user.Speciality,
			Gender:        user.Gender,
			MobileNo:      user.MobileNo,
			LicenseNumber: user.LicenseNumber,
			ClinicName:    user.ClinicName,
			ClinicAddress: user.ClinicAddress,
			UserAddress: models.AddressMaster{
				AddressId:    user.AddressMapping.AddressId,
				AddressLine1: user.AddressMapping.Address.AddressLine1,
				AddressLine2: user.AddressMapping.Address.AddressLine2,
				Landmark:     user.AddressMapping.Address.Landmark,
				City:         user.AddressMapping.Address.City,
				State:        user.AddressMapping.Address.State,
				Country:      user.AddressMapping.Address.Country,
				PostalCode:   user.AddressMapping.Address.PostalCode,
				Latitude:     user.AddressMapping.Address.Latitude,
				Longitude:    user.AddressMapping.Address.Longitude,
			},
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
			PatientId:   user.UserId,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			DateOfBirth: user.DateOfBirth.String(),
			Gender:      user.Gender,
			MobileNo:    user.MobileNo,
			Address:     user.Address,
			UserAddress: models.AddressMaster{
				AddressId:    user.AddressMapping.AddressId,
				AddressLine1: user.AddressMapping.Address.AddressLine1,
				AddressLine2: user.AddressMapping.Address.AddressLine2,
				Landmark:     user.AddressMapping.Address.Landmark,
				City:         user.AddressMapping.Address.City,
				State:        user.AddressMapping.Address.State,
				Country:      user.AddressMapping.Address.Country,
				PostalCode:   user.AddressMapping.Address.PostalCode,
				Latitude:     user.AddressMapping.Address.Latitude,
				Longitude:    user.AddressMapping.Address.Longitude,
			},
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

func MapUserToPublicProviderInfo(user interface{}, roleName string) interface{} {
	role := strings.ToLower(roleName)

	switch role {
	case "doctor":
		u := user.(models.SystemUser_)
		return models.DoctorInfo{
			DoctorId:          u.UserId,
			FirstName:         u.FirstName,
			LastName:          u.LastName,
			Speciality:        u.Speciality,
			ClinicName:        u.ClinicName,
			Gender:            u.Gender,
			MobileNo:          u.MobileNo,
			ClinicAddress:     u.ClinicAddress,
			YearsOfExperience: derefInt(u.YearsOfExperience),
			WorkingHours:      u.WorkingHours,
		}
	case "nurse":
		u := user.(models.SystemUser_)
		return models.NurseInfo{
			NurseId:           u.UserId,
			FirstName:         u.FirstName,
			LastName:          u.LastName,
			Gender:            u.Gender,
			MobileNo:          u.MobileNo,
			Speciality:        u.Speciality,
			ClinicName:        u.ClinicName,
			ClinicAddress:     u.ClinicAddress,
			YearsOfExperience: derefInt(u.YearsOfExperience),
			WorkingHours:      u.WorkingHours,
		}
	case "lab":
		lab := user.(models.DiagnosticLab)
		return models.LabInfo{
			LabId:            lab.DiagnosticLabId,
			LabNo:            lab.LabNo,
			LabName:          lab.LabName,
			LabAddress:       lab.LabAddress,
			LabContactNumber: lab.LabContactNumber,
			LabEmail:         lab.LabEmail,
		}

	case "patient":
		u := user.(models.SystemUser_)
		return models.PatientInfo{
			PatientId:  u.UserId,
			FirstName:  u.FirstName,
			LastName:   u.LastName,
			BloodGroup: u.BloodGroup,
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

func ParseIntField(val interface{}) int64 {
	switch v := val.(type) {
	case float64:
		return int64(v)
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return parsed
		}
	}
	return 0
}

func ParseDateField(val interface{}) time.Time {
	switch v := val.(type) {
	case string:
		if parsed, err := time.Parse(time.RFC3339, v); err == nil {
			return parsed
		}
		if parsed, err := time.Parse("2006-01-02", v); err == nil {
			return parsed
		}

	case float64:
		return time.Unix(int64(v), 0)
	}
	return time.Now()
}

func ConvertToZoomTime(dateStr, timeStr string) time.Time {
	combined := fmt.Sprintf("%s %s", dateStr, timeStr)
	layout := "2006-01-02 15:04:05"
	startTime, err := time.Parse(layout, combined)
	if err != nil {
		return time.Now().UTC()
	}
	return startTime
}

func ToDiseaseProfileSummaryDTO(profile models.DiseaseProfile) models.DiseaseProfileSummary {
	return models.DiseaseProfileSummary{
		DiseaseProfileId: profile.DiseaseProfileId,
		DiseaseName:      profile.Disease.DiseaseName,
		Description:      profile.Disease.Description,
	}
}

func ToDiseaseProfileSummaryDTOs(profiles []models.DiseaseProfile) []models.DiseaseProfileSummary {
	summaries := make([]models.DiseaseProfileSummary, len(profiles))
	for i, p := range profiles {
		summaries[i] = ToDiseaseProfileSummaryDTO(p)
	}
	return summaries
}
