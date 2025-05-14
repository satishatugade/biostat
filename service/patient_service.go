package service

import (
	"biostat/models"
	"biostat/repository"
	"fmt"
	"log"

	"gorm.io/gorm"
)

type PatientService interface {
	GetAllRelation() ([]models.PatientRelation, error)
	GetRelationById(relationId int) (models.PatientRelation, error)
	GetPatients(limit int, offset int) ([]models.Patient, int64, error)
	GetUserIdByAuthUserId(authUserId string) (uint64, error)
	UpdatePatientById(authUserId string, patientData *models.Patient) (*models.Patient, error)
	GetPatientDiseaseProfiles(PatientId uint64) ([]models.PatientDiseaseProfile, error)
	AddPatientDiseaseProfile(tx *gorm.DB, input *models.PatientDiseaseProfile) (*models.PatientDiseaseProfile, error)
	UpdateFlag(patientId uint64, req *models.DPRequest) error
	GetPatientDiagnosticResultValue(PatientId uint64, patientDiagnosticReportId uint64) ([]map[string]interface{}, error)
	GetPatientDiagnosticReportSummary(PatientId uint64, patientDiagnosticReportId uint64, summary bool) (models.ResultSummary, error)

	AddPatientPrescription(patientPrescription *models.PatientPrescription) error
	GetAllPrescription(limit int, offset int) ([]models.PatientPrescription, int64, error)
	GetPrescriptionByPatientId(PatientDiseaseProfileId string, limit int, offset int) ([]models.PatientPrescription, int64, error)
	AddPatientRelative(relative *models.PatientRelative) error
	GetPatientRelative(patientId string) ([]models.PatientRelative, error)
	GetRelativeList(patientId *uint64) ([]models.PatientRelative, error)
	GetCaregiverList(patientId *uint64) ([]models.Caregiver, error)
	GetDoctorList(patientId *uint64, limit, offset int) ([]models.SystemUser_, int64, error)
	GetPatientList() ([]models.Patient, error)
	GetPatientRelativeById(relativeId uint64, patientId uint64) (models.PatientRelative, error)
	UpdatePatientRelative(relativeId uint, relative *models.PatientRelative) (models.PatientRelative, error)
	AddPatientClinicalRange(customeRange *models.PatientCustomRange) error
	GetUserProfileByUserId(user_id uint64) (*models.SystemUser_, error)
	GetUserOnboardingStatusByUID(uid uint64) (bool, bool, bool, int64, int64, int64, int64, error)
	GetUserSUBByID(ID uint64) (string, error)
	GetUserIdBySUB(sub string) (uint64, error)
	ExistsByUserIdAndRoleId(userId uint64, roleId uint64) (bool, error)

	GetNursesList(patientId *uint64, limit int, offset int) ([]models.SystemUser_, int64, error)
	GetPharmacistList(patientId *uint64, limit int, offset int) ([]models.SystemUser_, int64, error)
	GetPatientDiagnosticTrendValue(input models.DiagnosticResultRequest) ([]map[string]interface{}, error)

	SaveUserHealthProfile(tx *gorm.DB, input *models.TblPatientHealthProfile) (*models.TblPatientHealthProfile, error)
}

type PatientServiceImpl struct {
	patientRepo     repository.PatientRepository
	patientRepoImpl repository.PatientRepositoryImpl
	userRepo        repository.UserRepository
	apiService      ApiService
}

// Ensure patientRepo is properly initialized
func NewPatientService(repo repository.PatientRepository, apiService ApiService) PatientService {
	return &PatientServiceImpl{patientRepo: repo, apiService: apiService}
}

// GetAllRelation implements PatientService.
func (s *PatientServiceImpl) GetAllRelation() ([]models.PatientRelation, error) {
	return s.patientRepo.GetAllRelation()
}

// GetRelationById implements PatientService.
func (s *PatientServiceImpl) GetRelationById(relationId int) (models.PatientRelation, error) {
	return s.patientRepo.GetRelationById(relationId)
}

// GetAllPatientPrescription implements PatientService.
func (s *PatientServiceImpl) GetAllPrescription(limit int, offset int) ([]models.PatientPrescription, int64, error) {
	return s.patientRepo.GetAllPrescription(limit, offset)
}

func (s *PatientServiceImpl) GetPrescriptionByPatientId(patientID string, limit int, offset int) ([]models.PatientPrescription, int64, error) {
	return s.patientRepo.GetPrescriptionByPatientId(patientID, limit, offset)
}

func (s *PatientServiceImpl) GetPatients(limit int, offset int) ([]models.Patient, int64, error) {
	return s.patientRepo.GetAllPatients(limit, offset)
}

func (s *PatientServiceImpl) GetUserIdByAuthUserId(authUserId string) (uint64, error) {
	return s.patientRepo.GetUserIdByAuthUserId(authUserId)
}

func (s *PatientServiceImpl) UpdatePatientById(authUserId string, patientData *models.Patient) (*models.Patient, error) {
	userId, err := s.patientRepo.GetUserIdByAuthUserId(authUserId)
	if err != nil {
		return &models.Patient{}, err
	}
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

func (s *PatientServiceImpl) AddPatientPrescription(prescription *models.PatientPrescription) error {
	return s.patientRepo.AddPatientPrescription(prescription)
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

// GetPatientRelatives implements PatientService.
func (s *PatientServiceImpl) GetPatientRelative(patientId string) ([]models.PatientRelative, error) {
	return s.patientRepo.GetPatientRelative(patientId)
}

func (s *PatientServiceImpl) UpdatePatientRelative(relativeId uint, relative *models.PatientRelative) (models.PatientRelative, error) {
	return s.patientRepo.UpdatePatientRelative(relativeId, relative)
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
	userRelationIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, "R", false)
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
	return s.patientRepo.GetRelativeList(relativeUserIds, userRelationIds, relation)
}

// GetCaregiverList implements PatientService.
func (s *PatientServiceImpl) GetCaregiverList(patientId *uint64) ([]models.Caregiver, error) {

	userRelationIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, "C", false)
	if err != nil {
		return []models.Caregiver{}, err
	}
	if len(userRelationIds) == 0 {
		return []models.Caregiver{}, nil
	}
	fmt.Println("relativeUserIds ", userRelationIds)
	caregiverUserIds, _ := ExtractUserAndRelationIds(userRelationIds)

	return s.patientRepo.GetCaregiverList(caregiverUserIds)
}

// GetDoctorList implements PatientService.
func (s *PatientServiceImpl) GetDoctorList(patientId *uint64, limit, offset int) ([]models.SystemUser_, int64, error) {

	userRelationIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, "D", false)
	if err != nil {
		return []models.SystemUser_{}, 0, err
	}
	if len(userRelationIds) == 0 {
		return []models.SystemUser_{}, 0, nil
	}
	doctorUserIds, _ := ExtractUserAndRelationIds(userRelationIds)
	// return s.patientRepo.GetDoctorList(doctorUserIds)
	return s.patientRepo.GetUserDataUserId(doctorUserIds, limit, offset)
}

// GetPatientList implements PatientService.
func (s *PatientServiceImpl) GetPatientList() ([]models.Patient, error) {

	userRelationIds, err := s.patientRepo.FetchUserIdByPatientId(nil, "S", true)
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
	reportData, err := s.patientRepo.GetPatientDiagnosticResultValue(PatientId, patientDiagnosticReportId)
	if err != nil {
		return nil, err
	}
	restructuredResponse := s.patientRepoImpl.RestructurePatientDiagnosticReport(reportData)
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
	userRelationIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, "N", false)
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
	userRelationIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, "P", false)
	if err != nil {
		return []models.SystemUser_{}, 0, err
	}
	if len(userRelationIds) == 0 {
		return []models.SystemUser_{}, 0, nil
	}
	chemistUserIds, _ := ExtractUserAndRelationIds(userRelationIds)
	return s.patientRepo.GetUserDataUserId(chemistUserIds, limit, offset)
}

func (s *PatientServiceImpl) GetUserIdBySUB(sub string) (uint64, error) {
	userId, err := s.patientRepo.GetUserIdBySUB(sub)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func (s *PatientServiceImpl) ExistsByUserIdAndRoleId(userId uint64, roleId uint64) (bool, error) {
	exists, err := s.patientRepo.ExistsByUserIdAndRoleId(userId, roleId)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (ps *PatientServiceImpl) GetPatientDiagnosticTrendValue(input models.DiagnosticResultRequest) ([]map[string]interface{}, error) {
	return ps.patientRepo.FetchPatientDiagnosticTrendValue(input)
}

func (ps *PatientServiceImpl) SaveUserHealthProfile(tx *gorm.DB, input *models.TblPatientHealthProfile) (*models.TblPatientHealthProfile, error) {
	return ps.patientRepo.SaveUserHealthProfile(tx, input)
}

func (s *PatientServiceImpl) GetPatientDiagnosticReportSummary(PatientId uint64, patientDiagnosticReportId uint64, summary bool) (models.ResultSummary, error) {
	reportData, err := s.patientRepo.GetPatientDiagnosticResultValue(PatientId, patientDiagnosticReportId)
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
	components := s.processTestComponents(test.DiagnosticTest.Components)

	testInput := models.PatientDiagnosticTestInput{
		TestNote:       test.TestNote,
		TestDate:       test.TestDate,
		TestName:       test.DiagnosticTest.TestName,
		TestComponents: components,
	}
	return testInput
}

func (s *PatientServiceImpl) processTestComponents(componentsData []models.DiagnosticTestComponent) []models.TestComponent {
	var components []models.TestComponent
	for _, c := range componentsData {
		component := s.createTestComponent(c)
		components = append(components, component)
	}
	return components
}

func (s *PatientServiceImpl) createTestComponent(c models.DiagnosticTestComponent) models.TestComponent {
	resultValues := s.processTestResultValues(c.TestResultValue)

	component := models.TestComponent{
		TestComponentName: c.TestComponentName,
		TestComponentType: c.TestComponentType,
		Units:             c.Units,
		TestResultValues:  resultValues,
	}
	return component
}

func (s *PatientServiceImpl) processTestResultValues(resultValuesData []models.PatientDiagnosticTestResultValue) []models.TestResultValue {
	var resultValues []models.TestResultValue
	for _, rv := range resultValuesData {
		// if rv.DiagnosticTestId == targetTestID && rv.DiagnosticTestComponentId == targetComponentID {
		resultValue := s.createTestResultValue(rv)
		resultValues = append(resultValues, resultValue)
		// }
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
