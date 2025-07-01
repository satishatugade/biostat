package utils

import (
	"biostat/constant"
	"biostat/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	offset := (page - 1) * limit

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

// func GetUserDataContext(c *gin.Context) (string, bool) {
// 	sub, exists := c.Get("sub")
// 	if !exists {
// 		return "Could not retrieve user id", false
// 	}

// 	subStr, ok := sub.(string)
// 	if !ok {
// 		return "user id not a valid string", false
// 	}

// 	return subStr, true
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
		MiddleName:         user.MiddleName,
		LastName:           user.LastName,
		DateOfBirth:        user.DateOfBirth,
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
		case "pharmacist":
			mappedUsers = append(mappedUsers, models.Pharmacist{
				PharmacistId:  user.UserId,
				FirstName:     user.FirstName,
				LastName:      user.LastName,
				Gender:        user.Gender,
				MobileNo:      user.MobileNo,
				Email:         user.Email,
				PharmacyName:  user.ClinicName,
				Speciality:    user.Speciality,
				LicenseNumber: user.LicenseNumber,
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
				MiddleName:  user.MiddleName,
				LastName:    user.LastName,
				DateOfBirth: user.DateOfBirth,
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
			GenderId:      user.GenderId,
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
	case "caregiver":
		return models.Caregiver{
			PatientId:  user.UserId,
			FirstName:  user.FirstName,
			MiddleName: user.MiddleName,
			LastName:   user.LastName,
			Gender:     user.Gender,
			GenderId:   user.GenderId,
			MobileNo:   user.MobileNo,
			Email:      user.Email,
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
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	case "pharmacist":
		return models.Pharmacist{
			PharmacistId:  user.UserId,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			Gender:        user.Gender,
			GenderId:      user.GenderId,
			MobileNo:      user.MobileNo,
			Email:         user.Email,
			PharmacyName:  user.ClinicName,
			Speciality:    user.Speciality,
			LicenseNumber: user.LicenseNumber,
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
			GenderId:      user.GenderId,
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
			PatientId:     user.UserId,
			FirstName:     user.FirstName,
			MiddleName:    user.MiddleName,
			LastName:      user.LastName,
			DateOfBirth:   user.DateOfBirth,
			Gender:        user.Gender,
			GenderId:      user.GenderId,
			MobileNo:      user.MobileNo,
			MaritalStatus: user.MaritalStatus,
			Address:       user.Address,
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

func GetUserIDFromContext(ctx *gin.Context, getUserIdBySubFunc func(string) (uint64, error)) (string, uint64, bool, error) {
	sub, subExists := ctx.Get("sub")
	if !subExists {
		return "", 0, false, errors.New("user not found")
	}

	delegateUserID := ctx.GetHeader("X-Delegate-User-Id")
	if delegateUserID != "" {
		id, err := strconv.ParseUint(delegateUserID, 10, 64)
		if err != nil {
			return "", 0, false, errors.New("invalid X-Delegate-User-Id")
		}
		return sub.(string), id, true, nil
	}
	userID, err := getUserIdBySubFunc(sub.(string))
	if err != nil {
		return "", 0, false, fmt.Errorf("failed to get user ID by sub: %w", err)
	}
	return sub.(string), userID, false, nil
}

func MappedRelationAccordingRelationship(userInfo *models.SystemUser_, relationId int) (int, error) {
	log.Println("Mapping relation based on gender and relationship...")

	var newRelationId int
	gender := strings.ToLower(userInfo.Gender)

	switch relationId {
	case 1, 2: // Father or Mother -> Son or Daughter
		if gender == "male" {
			newRelationId = 5 // Son
		} else if gender == "female" {
			newRelationId = 6 // Daughter
		}
	case 3, 4: // Brother or Sister -> Brother or Sister
		if gender == "male" {
			newRelationId = 3 // Brother
		} else if gender == "female" {
			newRelationId = 4 // Sister
		}
	case 5, 6: // Son or Daughter -> Father or Mother
		if gender == "male" {
			newRelationId = 1 // Father
		} else if gender == "female" {
			newRelationId = 2 // Mother
		}
	case 7, 8: // Husband or Wife -> Husband or Wife
		if gender == "male" {
			newRelationId = 7 // Husband
		} else if gender == "female" {
			newRelationId = 8 // Wife
		}
	case 9, 10: // Uncle or Aunt
		if gender == "male" {
			newRelationId = 9 // Uncle
		} else if gender == "female" {
			newRelationId = 10 // Aunt
		}
	case 11: // Cousin (gender-neutral)
		newRelationId = 11
	case 12, 13: // Grandfather or Grandmother
		if gender == "male" {
			newRelationId = 12 // Grandfather
		} else if gender == "female" {
			newRelationId = 13 // Grandmother
		}
	case 14, 15: // Nephew or Niece
		if gender == "male" {
			newRelationId = 14 // Nephew
		} else if gender == "female" {
			newRelationId = 15 // Niece
		}
	case 16: // Friend (gender-neutral)
		newRelationId = 16
	case 17: // Colleague (gender-neutral)
		newRelationId = 17
	case 18: // Guardian (gender-neutral)
		newRelationId = 18
	default:
		log.Printf("Unknown or unsupported relation ID: %d", relationId)
		return 0, fmt.Errorf("unsupported relation ID: %d", relationId)
	}

	if newRelationId == 0 {
		log.Printf("Unable to determine relationship for gender: %s", gender)
		return 0, fmt.Errorf("unable to determine relationship for gender: %s", gender)
	}

	log.Printf("Mapped relation ID: %d based on gender: %s", newRelationId, gender)
	return newRelationId, nil
}

func CalculatePatientBMI(weightKg, heightCm float64) (float64, string) {
	if heightCm <= 0 {
		log.Println("Height must be greater than 0")
	}
	heightM := heightCm / 100
	bmi := weightKg / (heightM * heightM)
	var category string
	switch {
	case bmi < 18.5:
		category = "Underweight"
	case bmi >= 18.5 && bmi < 24.9:
		category = "Normal weight"
	case bmi >= 25.0 && bmi < 29.9:
		category = "Overweight"
	case bmi >= 30.0:
		category = "Obese"
	default:
		category = "Unknown"
	}
	return bmi, category
}

func MakeRESTRequest(method, url string, body interface{}, headers map[string]string) (int, map[string]interface{}, error) {
	var requestBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return 0, nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		requestBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if errMsg, ok := responseData["error"].(string); ok && errMsg != "" {
		return resp.StatusCode, responseData, fmt.Errorf("API error: %s", errMsg)
	}

	return resp.StatusCode, responseData, nil
}

func GenerateGoogleCalendarLink(title, description, location string, start, end time.Time) string {
	base := "https://calendar.google.com/calendar/render?action=TEMPLATE"

	startStr := start.UTC().Format("20060102T150405Z")
	endStr := end.UTC().Format("20060102T150405Z")

	query := url.Values{}
	query.Set("text", title)
	query.Set("dates", fmt.Sprintf("%s/%s", startStr, endStr))
	query.Set("details", description)
	query.Set("location", location)

	return fmt.Sprintf("%s&%s", base, query.Encode())
}

func ExtractNotificationID(body map[string]interface{}) (uuid.UUID, error) {
	content, ok := body["content"].(map[string]interface{})
	if !ok {
		return uuid.Nil, errors.New("invalid or missing 'content' field")
	}
	notifIDStr, ok := content["notification_id"].(string)
	if !ok {
		return uuid.Nil, errors.New("'notification_id' not found or not a string")
	}
	notifUUID, err := uuid.Parse(notifIDStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid UUID format: " + err.Error())
	}
	return notifUUID, nil
}

func GetRefRangeAndColorCode(resultValue, normalMin, normalMax string) (string, string) {
	result, err1 := strconv.ParseFloat(resultValue, 64)
	min, err2 := strconv.ParseFloat(normalMin, 64)
	max, err3 := strconv.ParseFloat(normalMax, 64)

	if err1 != nil || err2 != nil || err3 != nil {
		log.Printf("[ColorCode] Invalid input - resultValue: %v (err: %v), normalMin: %v (err: %v), normalMax: %v (err: %v)",
			resultValue, err1, normalMin, err2, normalMax, err3)
		return "text-gray-500", "gray"
	}
	if result == 0 {
		return "text-black-500", "black"
	}
	if result < min {
		return "text-blue-500", "blue"
	} else if result > max {
		return "text-red-500", "red"
	} else {
		return "text-green-500", "green"
	}
}

func GetMappingType(roleName string, mappingType *string) string {
	if mappingType != nil && *mappingType == "PCG" {
		return "C"
	}
	switch roleName {
	case "patient":
		return "S"
	case "doctor":
		return "D"
	case "nurse":
		return "N"
	case "relative":
		return "R"
	case "caregiver":
		return "C"
	case "admin":
		return "A"
	case "pharmacist":
		return "P"
	default:
		return ""
	}
}

func GetConcurrentTaskCount() int {
	if n, err := strconv.Atoi(os.Getenv("CONCURRENT_TASK_COUNT_RUN")); err == nil && n > 0 {
		return n
	}
	return 50
}
