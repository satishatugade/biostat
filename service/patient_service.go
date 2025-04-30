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
	GetPatientById(patientId *uint64) (*models.Patient, error)
	GetUserIdByAuthUserId(authUserId string) (uint64, error)
	UpdatePatientById(authUserId string, patientData *models.Patient) (*models.Patient, error)
	GetPatientDiseaseProfiles(PatientId string) ([]models.PatientDiseaseProfile, error)
	AddPatientDiseaseProfile(tx *gorm.DB, input *models.PatientDiseaseProfile) (*models.PatientDiseaseProfile, error)
	GetPatientDiagnosticResultValue(PatientId uint64, patientDiagnosticReportId uint64) ([]models.PatientDiagnosticReport, error)
	AddPatientPrescription(patientPrescription *models.PatientPrescription) error
	GetAllPrescription(limit int, offset int) ([]models.PatientPrescription, int64, error)
	GetPrescriptionByPatientId(PatientDiseaseProfileId string, limit int, offset int) ([]models.PatientPrescription, int64, error)
	AddPatientRelative(relative *models.PatientRelative) error
	GetPatientRelative(patientId string) ([]models.PatientRelative, error)
	GetRelativeList(patientId *uint64) ([]models.PatientRelative, error)
	GetCaregiverList(patientId *uint64) ([]models.Caregiver, error)
	GetDoctorList(patientId *uint64) ([]models.Doctor, error)
	GetPatientList() ([]models.Patient, error)
	GetPatientRelativeById(relativeId uint64, patientId uint64) (models.PatientRelative, error)
	UpdatePatientRelative(relativeId uint, relative *models.PatientRelative) (models.PatientRelative, error)
	AddPatientClinicalRange(customeRange *models.PatientCustomRange) error
	GetUserProfileByUserId(user_id uint64) (*models.SystemUser_, error)
	GetUserOnboardingStatusByUID(SUB string) (bool, bool, bool, error)
	GetUserIdBySUB(sub string) (uint64, error)
	ExistsByUserIdAndRoleId(userId uint64, roleId uint64) (bool, error)

	GetNursesList(limit int, offset int) ([]models.Nurse, int64, error)
	GetPatientDiagnosticTrendValue(input models.DiagnosticResultRequest) ([]map[string]interface{}, error)

	SaveUserHealthProfile(tx *gorm.DB, input *models.TblPatientHealthProfile) (*models.TblPatientHealthProfile, error)
}

type PatientServiceImpl struct {
	patientRepo repository.PatientRepository
	userRepo    repository.UserRepository
}

// Ensure patientRepo is properly initialized
func NewPatientService(repo repository.PatientRepository) PatientService {
	return &PatientServiceImpl{patientRepo: repo}
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

func (s *PatientServiceImpl) GetPatientById(patientId *uint64) (*models.Patient, error) {
	return s.patientRepo.GetPatientById(patientId)
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

func (s *PatientServiceImpl) GetPatientDiseaseProfiles(PatientId string) ([]models.PatientDiseaseProfile, error) {
	return s.patientRepo.GetPatientDiseaseProfiles(PatientId)
}

func (s *PatientServiceImpl) AddPatientDiseaseProfile(tx *gorm.DB, input *models.PatientDiseaseProfile) (*models.PatientDiseaseProfile, error) {
	return s.patientRepo.AddPatientDiseaseProfile(tx, input)
}

// AddPatientRelative implements PatientService.
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
func (s *PatientServiceImpl) GetDoctorList(patientId *uint64) ([]models.Doctor, error) {

	userRelationIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, "D", false)
	if err != nil {
		return []models.Doctor{}, err
	}
	if len(userRelationIds) == 0 {
		return []models.Doctor{}, nil
	}
	doctorUserIds, _ := ExtractUserAndRelationIds(userRelationIds)
	return s.patientRepo.GetDoctorList(doctorUserIds)
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

func (s *PatientServiceImpl) GetPatientDiagnosticResultValue(PatientId uint64, patientDiagnosticReportId uint64) ([]models.PatientDiagnosticReport, error) {
	return s.patientRepo.GetPatientDiagnosticResultValue(PatientId, patientDiagnosticReportId)
	// reportData, err := s.patientRepo.GetPatientDiagnosticResultValue(PatientId, patientDiagnosticReportId)
	// if err != nil {
	// 	return []models.PatientDiagnosticReport{}, err
	// }
	// if summary == true {
	// 	user, err := s.patientRepo.GetUserProfileByUserId(PatientId)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	data = models.PatientBasicInfo{
	// 		PatientName: user.FirstName + user.LastName,
	// 		DateOfBirth: *user.DateOfBirth,
	// 		Gender:      user.Gender,
	// 		BloodGroup:  user.BloodGroup,
	// 	}

	// }

}

func (s *PatientServiceImpl) GetUserProfileByUserId(userId uint64) (*models.SystemUser_, error) {
	user, err := s.patientRepo.GetUserProfileByUserId(userId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *PatientServiceImpl) GetUserOnboardingStatusByUID(SUB string) (bool, bool, bool, error) {
	uid, err := s.patientRepo.GetUserIdBySUB(SUB)
	if err != nil {
		return false, false, false, err
	}

	basicDetailsAdded, err := s.patientRepo.IsUserBasicProfileComplete(uid)
	if err != nil {
		return false, false, false, err
	}

	familyDetailsAdded, err := s.patientRepo.IsUserFamilyDetailsComplete(uid)
	if err != nil {
		return false, false, false, err
	}

	healthDetailsAdded, err := s.patientRepo.IsUserHealthDetailsComplete(uid)
	if err != nil {
		return false, false, false, err
	}
	return basicDetailsAdded, familyDetailsAdded, healthDetailsAdded, nil
}

func (s *PatientServiceImpl) GetNursesList(limit int, offset int) ([]models.Nurse, int64, error) {
	return s.patientRepo.GetNursesList(limit, offset)
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
