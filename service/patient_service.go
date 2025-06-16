package service

import (
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"bytes"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type PatientService interface {
	GetAllRelation() ([]models.PatientRelation, error)
	GetRelationById(relationId int) (models.PatientRelation, error)
	GetPatients(limit int, offset int) ([]models.Patient, int64, error)
	UpdatePatientById(userId uint64, patientData *models.Patient) (*models.Patient, error)
	GetPatientDiseaseProfiles(PatientId uint64) ([]models.PatientDiseaseProfile, error)
	AddPatientDiseaseProfile(tx *gorm.DB, input *models.PatientDiseaseProfile) (*models.PatientDiseaseProfile, error)
	UpdateFlag(patientId uint64, req *models.DPRequest) error
	GetPatientDiagnosticResultValue(PatientId uint64, patientDiagnosticReportId uint64) ([]map[string]interface{}, error)
	GetPatientDiagnosticReportSummary(PatientId uint64, patientDiagnosticReportId uint64, summary bool) (models.ResultSummary, error)

	AddPatientPrescription(createdBy string, prescription *models.PatientPrescription) error
	UpdatePatientPrescription(authUserId string, prescription *models.PatientPrescription) error
	GetPrescriptionByPatientId(PatientId uint64, limit int, offset int) ([]models.PatientPrescription, int64, error)
	GetPrescriptionDetailByPatientId(PatientId uint64, limit int, offset int) ([]models.PrescriptionDetail, int64, error)
	GetPrescriptionInfo(prescriptiuonId uint64, patientId uint64) (string, error)
	GetPharmacokineticsInfo(prescriptiuonId uint64, patientId uint64) (string, error)
	SummarizeHistorybyAIModel(patientId uint64) (string, error)
	AddPatientRelative(relative *models.PatientRelative) error
	GetPatientRelative(patientId string) ([]models.PatientRelative, error)
	GetRelativeList(patientId *uint64) ([]models.PatientRelative, error)
	AssignPrimaryCaregiver(patientId uint64, relativeId uint64, mappingType string) error
	SetCaregiverMappingDeletedStatus(patientId uint64, caregiverId uint64, isDeleted int) error
	GetCaregiverList(patientId *uint64) ([]models.Caregiver, error)
	GetDoctorList(patientId *uint64, User string, limit, offset int) ([]models.SystemUser_, int64, error)
	GetPatientList() ([]models.Patient, error)
	GetPatientRelativeById(relativeId uint64, patientId uint64) (models.PatientRelative, error)
	AddPatientClinicalRange(customeRange *models.PatientCustomRange) error
	GetUserProfileByUserId(user_id uint64) (*models.SystemUser_, error)
	GetUserOnboardingStatusByUID(uid uint64) (bool, bool, bool, int64, int64, int64, int64, error)
	GetUserSUBByID(ID uint64) (string, error)
	ExistsByUserIdAndRoleId(userId uint64, roleId uint64) (bool, error)
	AddRelation(tx *gorm.DB, req models.AddRelationRequest, patientId uint64) error

	GetNursesList(patientId *uint64, limit int, offset int) ([]models.SystemUser_, int64, error)
	GetPharmacistList(patientId *uint64, limit int, offset int) ([]models.SystemUser_, int64, error)
	GetPatientDiagnosticTrendValue(input models.DiagnosticResultRequest) ([]map[string]interface{}, error)
	FetchPatientDiagnosticReports(patientID uint64, filter models.DiagnosticReportFilter) ([]map[string]interface{}, error)
	GetPatientDiagnosticReportResult(patientID uint64, filter models.DiagnosticReportFilter, limit, offset int) (map[string]interface{}, int64, error)
	GenerateExcelFile(data map[string]interface{}) ([]byte, error)
	SaveUserHealthProfile(tx *gorm.DB, input *models.TblPatientHealthProfile) (*models.TblPatientHealthProfile, error)
	GetPatientHealthDetail(patientId uint64) (models.TblPatientHealthProfile, error)
	UpdatePatientHealthDetail(req *models.TblPatientHealthProfile) error
	AddTestComponentDisplayConfig(config *models.PatientTestComponentDisplayConfig) error

	AssignPermission(userID, relativeID uint64, permissionCode string, granted bool) error
	CanPerformAction(userID, relativeID uint64, permissionCode string) (bool, error)
	AssignMultiplePermissions(userID, relativeID uint64, permissions map[string]bool) error
}

type PatientServiceImpl struct {
	patientRepo       repository.PatientRepository
	patientRepoImpl   repository.PatientRepositoryImpl
	apiService        ApiService
	allergyService    AllergyService
	medicalRecordRepo repository.TblMedicalRecordRepository
	roleRepo          repository.RoleRepository
}

// Ensure patientRepo is properly initialized
func NewPatientService(repo repository.PatientRepository, apiService ApiService, allergyService AllergyService, medicalRecordRepo repository.TblMedicalRecordRepository, roleRepo repository.RoleRepository) PatientService {
	return &PatientServiceImpl{patientRepo: repo, apiService: apiService, allergyService: allergyService, medicalRecordRepo: medicalRecordRepo, roleRepo: roleRepo}
}

// GetAllRelation implements PatientService.
func (s *PatientServiceImpl) GetAllRelation() ([]models.PatientRelation, error) {
	return s.patientRepo.GetAllRelation()
}

// GetRelationById implements PatientService.
func (s *PatientServiceImpl) GetRelationById(relationId int) (models.PatientRelation, error) {
	return s.patientRepo.GetRelationById(relationId)
}

func (s *PatientServiceImpl) GetPrescriptionByPatientId(patientId uint64, limit int, offset int) ([]models.PatientPrescription, int64, error) {
	return s.patientRepo.GetPrescriptionByPatientId(patientId, limit, offset)
}

func (s *PatientServiceImpl) GetPrescriptionDetailByPatientId(patientId uint64, limit int, offset int) ([]models.PrescriptionDetail, int64, error) {
	return s.patientRepo.GetPrescriptionDetailByPatientId(patientId, limit, offset)
}

func (s *PatientServiceImpl) GetPrescriptionInfo(prescriptionId uint64, patientId uint64) (string, error) {
	data, err := s.patientRepo.GetSinglePrescription(prescriptionId, patientId)
	if err != nil {
		return "", err
	}
	summaryText, err := s.apiService.AnalyzePrescriptionWithAI(data)
	if err != nil {
		return "", err
	}
	return summaryText, nil

}

func (s *PatientServiceImpl) GetPharmacokineticsInfo(prescriptionId uint64, patientId uint64) (string, error) {
	data, err := s.patientRepo.GetSinglePrescription(prescriptionId, patientId)
	if err != nil {
		return "", err
	}
	userInfo, err := s.patientRepo.GetUserProfileByUserId(patientId)
	if err != nil {
		return "", err
	}
	condition, err1 := s.patientRepo.GetPatientDiseaseProfiles(patientId, 0)
	if err1 != nil {
		return "", err1
	}
	allergies, err := s.allergyService.GetPatientAllergyRestriction(patientId)
	if err != nil {
		return "", err
	}
	var allergyList []string
	for _, allergy := range allergies {
		allergyList = append(allergyList, allergy.Allergy.AllergyName)
	}
	healthDetails, err := s.patientRepo.GetPatientHealthDetail(patientId)
	if err != nil {
		return "", err
	}
	var diseaseName string
	var diseaseType string

	for _, diseaseCondition := range condition {
		diseaseName = diseaseCondition.DiseaseProfile.Disease.DiseaseName
		diseaseType = diseaseCondition.DiseaseProfile.Disease.DiseaseType.DiseaseType
	}

	input := models.PharmacokineticsInput{
		Prescription: models.PrescriptionData{
			PatientName:  userInfo.FirstName + " " + userInfo.LastName,
			Age:          time.Now().Year() - userInfo.DateOfBirth.Year(),
			Gender:       userInfo.Gender,
			BloodGroup:   healthDetails.BloodType,
			BMI:          fmt.Sprintf("%.2f %s", healthDetails.BMI, healthDetails.BmiCategory),
			HeightCM:     fmt.Sprintf("%.2f", healthDetails.HeightCM),
			WeightKG:     fmt.Sprintf("%.2f", healthDetails.WeightKG),
			PrescribedOn: data.PrescriptionDate.Format("2006-01-02"),
			Prescription: []models.PrescribedDrug{},
		},
		History: models.HistoryData{
			PatientName:        userInfo.FirstName + " " + userInfo.LastName,
			Conditions:         []string{diseaseName + " " + diseaseType},
			Allergies:          allergyList,
			CurrentMedications: []models.CurrentMedication{},
			Lifestyle: []map[string]interface{}{
				{
					"SmokingStatus":         healthDetails.SmokingStatus,
					"AlcoholConsumption":    healthDetails.AlcoholConsumption,
					"PhysicalActivityLevel": healthDetails.PhysicalActivityLevel,
					"DietaryPreferences":    healthDetails.DietaryPreferences,
					"ExistingConditions":    healthDetails.ExistingConditions,
					"FamilyMedicalHistory":  healthDetails.FamilyMedicalHistory,
					"MenstrualHistory":      healthDetails.MenstrualHistory,
				},
			},
		},
	}

	for _, d := range data.PrescriptionDetails {
		input.Prescription.Prescription = append(input.Prescription.Prescription, models.PrescribedDrug{
			Drug:      d.MedicineName,
			Dosage:    fmt.Sprintf("%.0f%s", d.UnitValue, d.UnitType),
			Frequency: fmt.Sprintf("%d times/day", int(d.DosageInfo[0].DoseQuantity)),
			Duration:  fmt.Sprintf("%d days", d.Duration),
		})
	}

	summaryText, err := s.apiService.AnalyzePharmacokineticsInfo(input)
	if err != nil {
		return "", err
	}
	return summaryText, nil
}

func (s *PatientServiceImpl) SummarizeHistorybyAIModel(patientId uint64) (string, error) {

	data, _, err := s.patientRepo.GetPrescriptionByPatientId(patientId, 100, 0)
	if err != nil {
		return "", err
	}
	userInfo, err := s.patientRepo.GetUserProfileByUserId(patientId)
	if err != nil {
		return "", err
	}
	condition, err1 := s.patientRepo.GetPatientDiseaseProfiles(patientId, 0)
	if err1 != nil {
		return "", err1
	}
	allergies, err := s.allergyService.GetPatientAllergyRestriction(patientId)
	if err != nil {
		return "", err
	}
	var allergyList []string
	for _, allergy := range allergies {
		allergyList = append(allergyList, allergy.Allergy.AllergyName)
	}

	results, err2 := s.FetchPatientDiagnosticReports(patientId, models.DiagnosticReportFilter{ReportID: func() *uint64 { id, _ := s.patientRepo.GetDiagnosticReportId(patientId); return &id }()})

	if err2 != nil {
		return "", err2
	}
	var diseaseName string
	var diseaseType string

	for _, diseaseCondition := range condition {
		diseaseName = diseaseCondition.DiseaseProfile.Disease.DiseaseName
		diseaseType = diseaseCondition.DiseaseProfile.Disease.DiseaseType.DiseaseType
	}

	input := models.PharmacokineticsInput{
		Prescription: models.PrescriptionData{
			PatientName:  userInfo.FirstName + " " + userInfo.LastName,
			Age:          time.Now().Year() - userInfo.DateOfBirth.Year(),
			Gender:       userInfo.Gender,
			PrescribedOn: "2025-05-28",
			Prescription: []models.PrescribedDrug{},
		},
		History: models.HistoryData{
			PatientName:        userInfo.FirstName + " " + userInfo.LastName,
			Conditions:         []string{diseaseName + " " + diseaseType},
			Allergies:          allergyList,
			CurrentMedications: []models.CurrentMedication{},
			RecentLabResults:   results,
		},
	}

	for _, d := range data[0].PrescriptionDetails {
		input.Prescription.Prescription = append(input.Prescription.Prescription, models.PrescribedDrug{
			Drug:      d.MedicineName,
			Dosage:    fmt.Sprintf("%.0f%s", d.UnitValue, d.UnitType),
			Frequency: fmt.Sprintf("%d times/day", int(d.DosageInfo[0].DoseQuantity)),
			Duration:  fmt.Sprintf("%d days", d.Duration),
		})
	}

	summaryText, err := s.apiService.SummarizeMedicalHistory(input)
	if err != nil {
		return "", err
	}
	return summaryText, nil
}

func (s *PatientServiceImpl) GetPatients(limit int, offset int) ([]models.Patient, int64, error) {
	return s.patientRepo.GetAllPatients(limit, offset)
}

func (s *PatientServiceImpl) UpdatePatientById(userId uint64, patientData *models.Patient) (*models.Patient, error) {
	updatedPatient, err := s.patientRepo.UpdatePatientById(userId, patientData)
	if err != nil {
		return &models.Patient{}, err
	}
	updatedAddress, err := s.patientRepo.UpdateUserAddressByUserId(userId, patientData.UserAddress)
	if err != nil {
		return &models.Patient{}, err
	}
	return s.patientRepo.MapSystemUserToPatient(&updatedPatient, updatedAddress), nil
}

func (s *PatientServiceImpl) AddPatientPrescription(createdBy string, prescription *models.PatientPrescription) error {
	return s.patientRepo.AddPatientPrescription(createdBy, prescription)
}

func (s *PatientServiceImpl) UpdatePatientPrescription(createdBy string, prescription *models.PatientPrescription) error {
	return s.patientRepo.UpdatePatientPrescription(createdBy, prescription)
}

func (s *PatientServiceImpl) GetPatientDiseaseProfiles(PatientId uint64) ([]models.PatientDiseaseProfile, error) {
	return s.patientRepo.GetPatientDiseaseProfiles(PatientId, 0)
}

func (s *PatientServiceImpl) AddPatientDiseaseProfile(tx *gorm.DB, input *models.PatientDiseaseProfile) (*models.PatientDiseaseProfile, error) {
	return s.patientRepo.AddPatientDiseaseProfile(tx, input)
}

func (s *PatientServiceImpl) UpdateFlag(patientId uint64, req *models.DPRequest) error {
	return s.patientRepo.UpdateFlag(patientId, req)
}

func (s *PatientServiceImpl) AddPatientRelative(relative *models.PatientRelative) error {
	return s.patientRepo.AddPatientRelative(relative)
}

func (s *PatientServiceImpl) AssignPrimaryCaregiver(patientId, relativeId uint64, mappingType string) error {
	return s.patientRepo.AssignPrimaryCaregiver(patientId, relativeId, mappingType)
}

func (s *PatientServiceImpl) SetCaregiverMappingDeletedStatus(patientId, caregiverId uint64, isDeleted int) error {
	return s.patientRepo.SetCaregiverMappingDeletedStatus(patientId, caregiverId, isDeleted)
}

// GetPatientRelatives implements PatientService.
func (s *PatientServiceImpl) GetPatientRelative(patientId string) ([]models.PatientRelative, error) {
	return s.patientRepo.GetPatientRelative(patientId)
}

// AddPatientClinicalRange implements PatientService.
func (s *PatientServiceImpl) AddPatientClinicalRange(customRange *models.PatientCustomRange) error {
	return s.patientRepo.AddPatientClinicalRange(customRange)
}

func (s *PatientServiceImpl) GetPatientRelativeById(relativeId uint64, patientId uint64) (models.PatientRelative, error) {
	patientRelativeId, relationeId, err := s.patientRepo.CheckPatientRelativeMapping(relativeId, patientId, "R")
	if err != nil {
		log.Println("CheckPatientRelativeMapping Not found :")
		return models.PatientRelative{}, err
	}
	relationeIds := []uint64{relativeId}
	var relation []models.PatientRelation
	if relationeId != 0 {
		relation, err = s.patientRepo.GetRelationNameById(relationeIds)
		if err != nil {
			log.Println("GetRelationNameById Not found :")
		}
	}
	return s.patientRepo.GetPatientRelativeById(patientRelativeId, relation)
}

func ExtractUserAndRelationIds(userRelations []models.UserRelation) ([]uint64, []uint64) {
	var userIds []uint64
	var relationIds []uint64

	for _, ur := range userRelations {
		userIds = append(userIds, ur.UserId)
		relationIds = append(relationIds, ur.RelationId)
	}

	return userIds, relationIds
}

func (s *PatientServiceImpl) GetRelativeList(patientId *uint64) ([]models.PatientRelative, error) {
	userRelationIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, []string{"R", "PCG"}, false, 0)
	if err != nil {
		return []models.PatientRelative{}, err
	}
	if len(userRelationIds) == 0 {
		return []models.PatientRelative{}, nil
	}
	relativeUserIds, relationIds := ExtractUserAndRelationIds(userRelationIds)
	var relation []models.PatientRelation
	relation, err = s.patientRepo.GetRelationNameById(relationIds)
	if err != nil {
		log.Println("GetRelationNameById Not found :")
	}
	userRelatives, err := s.patientRepo.GetRelativeList(relativeUserIds, userRelationIds, relation)
	if err != nil {
		return []models.PatientRelative{}, err
	}
	for idx, _ := range userRelatives {
		perms, err := s.patientRepo.ListPermissions(*patientId, userRelatives[idx].RelativeId)
		if err != nil {
			log.Println("@GetRelativeList -> ListPermissions,", err)
			continue
		}
		userRelatives[idx].HealthScore = s.GetUserHealthScore(userRelatives[idx].RelativeId)
		userRelatives[idx].Permissions = perms
	}

	return userRelatives, nil
}

func (s *PatientServiceImpl) GetCaregiverList(patientId *uint64) ([]models.Caregiver, error) {

	userRelationIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, []string{"C"}, false, 0)
	if err != nil {
		return []models.Caregiver{}, err
	}
	if len(userRelationIds) == 0 {
		return []models.Caregiver{}, nil
	}
	caregiverUserIds, _ := ExtractUserAndRelationIds(userRelationIds)

	return s.patientRepo.GetCaregiverList(caregiverUserIds)
}

func (s *PatientServiceImpl) GetDoctorList(patientId *uint64, User string, limit, offset int) ([]models.SystemUser_, int64, error) {

	userRelationIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, []string{"D"}, false, 0)
	if err != nil {
		return []models.SystemUser_{}, 0, err
	}
	if len(userRelationIds) == 0 {
		return []models.SystemUser_{}, 0, nil
	}
	if User != "patient" {
		filteredDoctorIds := make([]uint64, 0)
		for _, rel := range userRelationIds {
			if rel.UserId == rel.PatientId {
				filteredDoctorIds = append(filteredDoctorIds, rel.UserId)
			}
		}

		if len(filteredDoctorIds) == 0 {
			return []models.SystemUser_{}, 0, nil
		}
		return s.patientRepo.GetUserDataUserId(filteredDoctorIds, limit, offset)
	} else {
		doctorUserIds, _ := ExtractUserAndRelationIds(userRelationIds)
		return s.patientRepo.GetUserDataUserId(doctorUserIds, limit, offset)
	}

}

func (s *PatientServiceImpl) GetPatientList() ([]models.Patient, error) {

	userRelationIds, err := s.patientRepo.FetchUserIdByPatientId(nil, []string{"S"}, true, 0)
	if err != nil {
		return []models.Patient{}, err
	}
	if len(userRelationIds) == 0 {
		return []models.Patient{}, nil
	}
	patientUserIds, _ := ExtractUserAndRelationIds(userRelationIds)
	return s.patientRepo.GetPatientList(patientUserIds)
}

func (s *PatientServiceImpl) GetPatientDiagnosticResultValue(PatientId uint64, patientDiagnosticReportId uint64) ([]map[string]interface{}, error) {
	reportData, recordIds, err := s.patientRepo.GetPatientDiagnosticResultValue(PatientId, patientDiagnosticReportId)
	if err != nil {
		return nil, err
	}
	medicalRecordInfo, _ := s.medicalRecordRepo.GetMedicalRecordsByUserID(PatientId, recordIds)

	restructuredResponse := s.patientRepoImpl.RestructurePatientDiagnosticReport(reportData, medicalRecordInfo, recordIds)
	return restructuredResponse, nil
}

func (s *PatientServiceImpl) GetUserProfileByUserId(userId uint64) (*models.SystemUser_, error) {
	user, err := s.patientRepo.GetUserProfileByUserId(userId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *PatientServiceImpl) GetUserOnboardingStatusByUID(uid uint64) (bool, bool, bool, int64, int64, int64, int64, error) {
	basicDetailsAdded, err := s.patientRepo.IsUserBasicProfileComplete(uid)
	if err != nil {
		return false, false, false, 0, 0, 0, 0, err
	}

	familyDetailsAdded, err := s.patientRepo.IsUserFamilyDetailsComplete(uid)
	if err != nil {
		return false, false, false, 0, 0, 0, 0, err
	}

	healthDetailsAdded, err := s.patientRepo.IsUserHealthDetailsComplete(uid)
	if err != nil {
		return false, false, false, 0, 0, 0, 0, err
	}

	noOfUpcomingAppointments, err := s.patientRepo.NoOfUpcomingAppointments(uid)
	if err != nil {
		return false, false, false, 0, 0, 0, 0, err
	}

	noOfMedicationsForDashboard, err := s.patientRepo.NoOfMedicationsForDashboard(uid)
	if err != nil {
		return false, false, false, 0, 0, 0, 0, err
	}

	noOfMessagesForDashboard, err := s.patientRepo.NoOfMessagesForDashboard(uid)
	if err != nil {
		return false, false, false, 0, 0, 0, 0, err
	}

	noOfLabReusltsForDashboard, err := s.patientRepo.NoOfLabReusltsForDashboard(uid)
	if err != nil {
		return false, false, false, 0, 0, 0, 0, err
	}

	return basicDetailsAdded, familyDetailsAdded, healthDetailsAdded, noOfUpcomingAppointments, noOfMedicationsForDashboard, noOfMessagesForDashboard, noOfLabReusltsForDashboard, nil
}

func (s *PatientServiceImpl) GetUserSUBByID(id uint64) (string, error) {
	sub, err := s.patientRepo.GetUserSUBByID(id)
	if err != nil {
		return "", err
	}
	return sub, nil
}

func (s *PatientServiceImpl) GetNursesList(patientId *uint64, limit int, offset int) ([]models.SystemUser_, int64, error) {
	//return s.patientRepo.GetNursesList(limit, offset)
	userRelationIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, []string{"N"}, false, 0)
	if err != nil {
		return []models.SystemUser_{}, 0, err
	}
	if len(userRelationIds) == 0 {
		return []models.SystemUser_{}, 0, nil
	}
	nurseUserIds, _ := ExtractUserAndRelationIds(userRelationIds)
	return s.patientRepo.GetUserDataUserId(nurseUserIds, limit, offset)
}

func (s *PatientServiceImpl) GetPharmacistList(patientId *uint64, limit int, offset int) ([]models.SystemUser_, int64, error) {
	userRelationIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, []string{"P"}, false, 0)
	if err != nil {
		return []models.SystemUser_{}, 0, err
	}
	if len(userRelationIds) == 0 {
		return []models.SystemUser_{}, 0, nil
	}
	chemistUserIds, _ := ExtractUserAndRelationIds(userRelationIds)
	return s.patientRepo.GetUserDataUserId(chemistUserIds, limit, offset)
}

func (s *PatientServiceImpl) ExistsByUserIdAndRoleId(userId uint64, roleId uint64) (bool, error) {
	exists, err := s.patientRepo.ExistsByUserIdAndRoleId(userId, roleId)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (ps *PatientServiceImpl) GetPatientDiagnosticTrendValue(input models.DiagnosticResultRequest) ([]map[string]interface{}, error) {
	data, err := ps.patientRepo.FetchPatientDiagnosticTrendValue(input)
	if err != nil {
		return nil, err
	}
	return ps.patientRepo.ParseDiagnosticTrendData(data)
}

func (ps *PatientServiceImpl) SaveUserHealthProfile(tx *gorm.DB, input *models.TblPatientHealthProfile) (*models.TblPatientHealthProfile, error) {
	exists, err := ps.patientRepo.CheckPatientHealthProfileExist(tx, input.PatientId)
	if err != nil {
		return nil, err
	}
	if exists {
		err := ps.patientRepo.UpdatePatientHealthDetail(input)
		if err != nil {
			return nil, err
		}
		return input, nil
	}
	return ps.patientRepo.SaveUserHealthProfile(tx, input)
}

func (s *PatientServiceImpl) GetPatientHealthDetail(patientId uint64) (models.TblPatientHealthProfile, error) {
	return s.patientRepo.GetPatientHealthDetail(patientId)
}

func (s *PatientServiceImpl) UpdatePatientHealthDetail(req *models.TblPatientHealthProfile) error {
	return s.patientRepo.UpdatePatientHealthDetail(req)
}

func (s *PatientServiceImpl) GetPatientDiagnosticReportSummary(PatientId uint64, patientDiagnosticReportId uint64, summary bool) (models.ResultSummary, error) {
	reportData, _, err := s.patientRepo.GetPatientDiagnosticResultValue(PatientId, patientDiagnosticReportId)
	if err != nil {
		return models.ResultSummary{}, err
	}

	if !summary {
		return models.ResultSummary{}, nil
	}

	user, err := s.patientRepo.GetUserProfileByUserId(PatientId)
	if err != nil {
		return models.ResultSummary{}, err
	}

	reports := s.processReportData(reportData)

	data := models.PatientBasicInfo{
		PatientName: user.FirstName + " " + user.LastName,
		DateOfBirth: *user.DateOfBirth,
		Gender:      user.Gender,
		BloodGroup:  user.BloodGroup,
		Reports:     reports,
	}
	summaryText, err := s.apiService.CallSummarizeReportService(data)
	if err != nil {
		return models.ResultSummary{}, err
	}
	return summaryText, nil
}

func (s *PatientServiceImpl) processReportData(reportData []models.PatientDiagnosticReport) []models.Report {
	var reports []models.Report
	for _, r := range reportData {
		report := s.createReport(r)
		reports = append(reports, report)
	}
	return reports
}

func (s *PatientServiceImpl) createReport(r models.PatientDiagnosticReport) models.Report {
	testDetails := s.processPatientDiagnosticTests(r.PatientDiagnosticTests)

	report := models.Report{
		PaymentStatus:     r.PaymentStatus,
		CollectedDate:     r.CollectedDate,
		CollectedAt:       r.CollectedAt,
		ProcessedAt:       r.ProcessedAt,
		ReportDate:        r.ReportDate,
		ReportStatus:      r.ReportStatus,
		Observation:       r.Observation,
		Comments:          r.Comments,
		DiagnosticLabInfo: s.createDiagnosticLabInfo(r.DiagnosticLabs, testDetails),
	}
	return report
}

func (s *PatientServiceImpl) createDiagnosticLabInfo(dl models.DiagnosticLab, testDetails []models.PatientDiagnosticTestInput) models.DiagnosticLabCenter {
	return models.DiagnosticLabCenter{
		LabName:               dl.LabName,
		LabAddress:            dl.LabAddress,
		PatientDiagnosticTest: testDetails,
	}
}

func (s *PatientServiceImpl) processPatientDiagnosticTests(patientDiagnosticTests []models.PatientDiagnosticTest) []models.PatientDiagnosticTestInput {
	var testDetails []models.PatientDiagnosticTestInput
	for _, test := range patientDiagnosticTests {
		testInput := s.createPatientDiagnosticTestInput(test)
		testDetails = append(testDetails, testInput)
	}
	return testDetails
}

func (s *PatientServiceImpl) createPatientDiagnosticTestInput(test models.PatientDiagnosticTest) models.PatientDiagnosticTestInput {
	components := s.processTestComponents(test.PatientDiagnosticReportId, test.DiagnosticTest.Components)

	testInput := models.PatientDiagnosticTestInput{
		TestNote:       test.TestNote,
		TestDate:       test.TestDate,
		TestName:       test.DiagnosticTest.TestName,
		TestComponents: components,
	}
	return testInput
}

func (s *PatientServiceImpl) processTestComponents(PatientDiagnosticReportId uint64, componentsData []models.DiagnosticTestComponent) []models.TestComponent {
	var components []models.TestComponent
	for _, c := range componentsData {
		component := s.createTestComponent(c, PatientDiagnosticReportId)
		components = append(components, component)
	}
	return components
}

func (s *PatientServiceImpl) createTestComponent(c models.DiagnosticTestComponent, PatientDiagnosticReportId uint64) models.TestComponent {
	resultValues := s.processTestResultValues(PatientDiagnosticReportId, c.TestResultValue)

	component := models.TestComponent{
		TestComponentName: c.TestComponentName,
		TestComponentType: c.TestComponentType,
		Units:             c.Units,
		TestResultValues:  resultValues,
	}
	return component
}

func (s *PatientServiceImpl) processTestResultValues(PatientDiagnosticReportId uint64, resultValuesData []models.PatientDiagnosticTestResultValue) []models.TestResultValue {
	var resultValues []models.TestResultValue
	for _, rv := range resultValuesData {
		if PatientDiagnosticReportId == rv.PatientDiagnosticReportId {
			resultValue := s.createTestResultValue(rv)
			resultValues = append(resultValues, resultValue)
		}
	}
	return resultValues
}

func (s *PatientServiceImpl) createTestResultValue(rv models.PatientDiagnosticTestResultValue) models.TestResultValue {
	return models.TestResultValue{
		ResultValue:   rv.ResultValue,
		ResultStatus:  rv.ResultStatus,
		ResultDate:    rv.ResultDate,
		ResultComment: rv.ResultComment,
		Udf1:          rv.UDF1,
	}
}

func (ps *PatientServiceImpl) FetchPatientDiagnosticReports(patientId uint64, filter models.DiagnosticReportFilter) ([]map[string]interface{}, error) {
	data, err := ps.patientRepo.FetchPatientDiagnosticReports(patientId, filter)
	if err != nil {
		return nil, err
	}

	nestedResults := ps.patientRepo.RestructureDiagnosticReports(data)
	return nestedResults, nil
}

func (ps *PatientServiceImpl) GetPatientDiagnosticReportResult(patientId uint64, filter models.DiagnosticReportFilter, limit, offset int) (map[string]interface{}, int64, error) {
	data, totalReports, err := ps.patientRepo.GetPatientDiagnosticReportResult(patientId, filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	response := ps.patientRepo.ProcessReportGridData(data)
	return response, totalReports, nil
}

func (ps *PatientServiceImpl) GenerateExcelFile(data map[string]interface{}) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Report"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"Test Component Name", "Unit", "Ref. Range"}

	dates, ok := data["dates"].([]string)
	if !ok || len(dates) == 0 {
		return nil, fmt.Errorf("invalid or no dates found")
	}
	headers = append(headers, dates...)

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
	}

	rows, ok := data["rows"].([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid rows")
	}
	for rowIndex, row := range rows {
		r := rowIndex + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), row["test_component_name"])
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), row["ref_unit"])
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), row["ref_range"])

		rawTrendValues := row["trend_values"]

		trendValues, ok := rawTrendValues.([]models.CellData)
		if !ok {
			fmt.Println("trend_values not []interface{} but type:", reflect.TypeOf(rawTrendValues))
			continue
		}
		dateToValue := make(map[string]models.CellData)
		for _, tv := range trendValues {
			dateToValue[tv.ResultDate] = tv
		}
		for colIndex, date := range dates {
			val := "-"
			styleID := 0

			if cellData, found := dateToValue[date]; found {
				val = cellData.Value
				styleID = applyColorStyle(f, cellData.ColourClass)
			}

			cell, _ := excelize.CoordinatesToCellName(colIndex+4, r)
			f.SetCellValue(sheet, cell, val)
			if styleID != 0 {
				f.SetCellStyle(sheet, cell, cell, styleID)
			}
		}
	}
	endCol, _ := excelize.ColumnNumberToName(len(headers))
	rangeRef := fmt.Sprintf("A1:%s1", endCol)
	_ = f.AutoFilter(sheet, rangeRef, nil)
	f.SetColWidth(sheet, "A", endCol, 20)

	// Write to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func applyColorStyle(f *excelize.File, className string) int {
	colorMap := map[string]string{
		"text-green-500":  "#22c55e",
		"text-blue-500":   "#3b82f6",
		"text-red-500":    "#ef4444",
		"text-yellow-500": "#eab308",
	}

	colorHex, ok := colorMap[className]
	if !ok {
		colorHex = "#000000" // default black
	}

	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: colorHex, Bold: true},
	})
	return style
}

func (s *PatientServiceImpl) AddTestComponentDisplayConfig(config *models.PatientTestComponentDisplayConfig) error {
	return s.patientRepo.AddTestComponentDisplayConfig(config)
}

func (s *PatientServiceImpl) AssignPermission(userID, relativeID uint64, code string, granted bool) error {
	perm, err := s.patientRepo.GetPermissionByCode(code)
	if err != nil {
		return err
	}
	return s.patientRepo.GrantPermission(userID, relativeID, perm.PermissionID, granted)
}

func (s *PatientServiceImpl) CanPerformAction(userID, relativeID uint64, code string) (bool, error) {
	return s.patientRepo.HasPermission(userID, relativeID, code)
}

func (s *PatientServiceImpl) AssignMultiplePermissions(userID, relativeID uint64, permissions map[string]bool) error {
	for code, value := range permissions {
		perm, err := s.patientRepo.GetPermissionByCode(code)
		if err != nil {
			log.Printf("Permission code '%s' not found", code)
			continue
		}

		exists, currentValue, _ := s.patientRepo.CheckPermissionValue(userID, relativeID, perm.PermissionID)
		if !exists {
			err := s.patientRepo.GrantPermission(userID, relativeID, perm.PermissionID, value)
			if err != nil {
				log.Printf("Failed to create mapping for permission %s: %v", code, err)
			}
		} else if currentValue != value {
			err := s.patientRepo.UpdatePermissionValue(userID, relativeID, perm.PermissionID, value)
			if err != nil {
				log.Printf("Failed to update permission %s: %v", code, err)
			}
		}
	}
	return nil
}

func (s *PatientServiceImpl) GetUserHealthScore(userID uint64) int {
	healthScore := 0
	basicDetailsAdded, _ := s.patientRepo.IsUserBasicProfileComplete(userID)

	if basicDetailsAdded {
		healthScore += 10
	}
	healthDetailsAdded, _ := s.patientRepo.IsUserHealthDetailsComplete(userID)
	if healthDetailsAdded {
		healthScore += 10
	}

	noOfLabReusltsForDashboard, _ := s.patientRepo.NoOfLabReusltsForDashboard(userID)
	if noOfLabReusltsForDashboard > 0 {
		healthScore += 10
	}
	healthProfile, err := s.patientRepo.GetPatientHealthDetail(userID)
	if err == nil {
		if healthProfile.BmiCategory == "Normal weight" {
			healthScore += 5
		}
		if healthProfile.SmokingStatus != "Current" {
			healthScore += 5
		}
		if healthProfile.AlcoholConsumption == "Never" {
			healthScore += 5
		}
		if healthProfile.PhysicalActivityLevel == "High" {
			healthScore += 5
		}
	}

	return healthScore
}

func (s *PatientServiceImpl) AddRelation(tx *gorm.DB, req models.AddRelationRequest, patientId uint64) error {
	var checkType string
	switch req.CurrentRole {
	case "caregiver":
		checkType = "C"
	case "relative":
		checkType = "R"
	default:
		return fmt.Errorf("unsupported role: %s", req.CurrentRole)
	}

	_, relationId, err := s.patientRepo.CheckPatientRelativeMapping(req.UserID, patientId, checkType)
	if err != nil {
		return err
	}
	role, err := s.roleRepo.GetRoleIdByRoleName(req.RoleName)
	if err != nil {
		tx.Rollback()
		return err
	}
	mappingType := utils.GetMappingType(role.RoleName)
	if mappingType == "" {
		tx.Rollback()
		return fmt.Errorf("invalid mapping type for role: %s", role.RoleName)
	}
	mapping := models.SystemUserRoleMapping{
		UserId:      req.UserID,
		PatientId:   patientId,
		RelationId:  int(relationId),
		MappingType: mappingType,
		IsSelf:      false,
		RoleId:      role.RoleId,
	}
	return s.roleRepo.AddSystemUserMapping(tx, &mapping)
}
