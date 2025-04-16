package service

import (
	"biostat/models"
	"biostat/repository"
	"fmt"
)

type PatientService interface {
	GetAllRelation() ([]models.PatientRelation, error)
	GetRelationById(relationId int) (models.PatientRelation, error)
	GetPatients(limit int, offset int) ([]models.Patient, int64, error)
	GetPatientById(patientId *uint64) (*models.Patient, error)
	UpdatePatientById(authUserId string, patientData *models.Patient) (*models.Patient, error)
	GetPatientDiseaseProfiles(PatientId string) ([]models.PatientDiseaseProfile, error)
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
	GetPatientRelativeById(relativeId uint) (models.PatientRelative, error)
	UpdatePatientRelative(relativeId uint, relative *models.PatientRelative) (models.PatientRelative, error)
	AddPatientClinicalRange(customeRange *models.PatientCustomRange) error
	// UpdatePrescription(*models.PatientPrescription) error
	GetUserProfile(user_id string, roles []string) (*models.Patient, error)
	GetUserOnboardingStatusByUID(SUB string) (bool, bool, bool, error)
	GetUserIdBySUB(sub string) (uint64, error)
	ExistsByUserIdAndRoleId(userId uint64, roleId uint64) (bool, error)

	GetNursesList(limit int, offset int) ([]models.Nurse, int64, error)
}

type PatientServiceImpl struct {
	patientRepo repository.PatientRepository
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
func (s *PatientServiceImpl) UpdatePatientById(authUserId string, patientData *models.Patient) (*models.Patient, error) {
	return s.patientRepo.UpdatePatientById(authUserId, patientData)
}

func (s *PatientServiceImpl) AddPatientPrescription(prescription *models.PatientPrescription) error {
	return s.patientRepo.AddPatientPrescription(prescription)
}

func (s *PatientServiceImpl) GetPatientDiseaseProfiles(PatientId string) ([]models.PatientDiseaseProfile, error) {
	return s.patientRepo.GetPatientDiseaseProfiles(PatientId)
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

// GetPatientRelativeById implements PatientService.
func (s *PatientServiceImpl) GetPatientRelativeById(relativeId uint) (models.PatientRelative, error) {
	return s.patientRepo.GetPatientRelativeById(relativeId)
}

// GetRelativeList implements PatientService.
func (s *PatientServiceImpl) GetRelativeList(patientId *uint64) ([]models.PatientRelative, error) {
	relativeUserIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, "R", false)
	if err != nil {
		return []models.PatientRelative{}, err
	}

	return s.patientRepo.GetRelativeList(relativeUserIds)
}

// GetCaregiverList implements PatientService.
func (s *PatientServiceImpl) GetCaregiverList(patientId *uint64) ([]models.Caregiver, error) {

	caregiverUserIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, "C", false)
	if err != nil {
		return []models.Caregiver{}, err
	}

	return s.patientRepo.GetCaregiverList(caregiverUserIds)
}

// GetDoctorList implements PatientService.
func (s *PatientServiceImpl) GetDoctorList(patientId *uint64) ([]models.Doctor, error) {

	doctorUserIds, err := s.patientRepo.FetchUserIdByPatientId(patientId, "D", false)
	if err != nil {
		return []models.Doctor{}, err
	}

	return s.patientRepo.GetDoctorList(doctorUserIds)
}

// GetPatientList implements PatientService.
func (s *PatientServiceImpl) GetPatientList() ([]models.Patient, error) {

	patientUserIds, err := s.patientRepo.FetchUserIdByPatientId(nil, "S", true)
	if err != nil {
		return []models.Patient{}, err
	}
	fmt.Println("patientUserIds ", patientUserIds)
	return s.patientRepo.GetPatientList(patientUserIds)
}

func (s *PatientServiceImpl) GetPatientDiagnosticResultValue(PatientId uint64, patientDiagnosticReportId uint64) ([]models.PatientDiagnosticReport, error) {
	return s.patientRepo.GetPatientDiagnosticResultValue(PatientId, patientDiagnosticReportId)
}

func (s *PatientServiceImpl) GetUserProfile(user_id string, roles []string) (*models.Patient, error) {
	user, err := s.patientRepo.GetUserProfile(user_id)
	if err != nil {
		return nil, err
	}
	patient := &models.Patient{
		PatientId:          user.UserId,
		FirstName:          user.FirstName,
		LastName:           user.LastName,
		Gender:             user.Gender,
		DateOfBirth:        user.DateOfBirth.String(),
		MobileNo:           user.MobileNo,
		Address:            user.Address,
		BloodGroup:         user.BloodGroup,
		AbhaNumber:         user.AbhaNumber,
		EmergencyContact:   user.EmergencyContact,
		Email:              user.Email,
		Nationality:        user.Nationality,
		CitizenshipStatus:  user.CitizenshipStatus,
		PassportNumber:     user.PassportNumber,
		CountryOfResidence: user.CountryOfResidence,
		IsIndianOrigin:     user.IsIndianOrigin,
	}
	return patient, nil
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

	return basicDetailsAdded, familyDetailsAdded, false, nil

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
