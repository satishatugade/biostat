package repository

import (
	"biostat/models"
	"errors"

	"gorm.io/gorm"
)

type PatientRepository interface {
	GetAllRelation() ([]models.PatientRelation, error)
	GetRelationById(relationId int) (models.PatientRelation, error)
	GetAllPatients(limit int, offset int) ([]models.Patient, int64, error)
	AddPatientPrescription(*models.PatientPrescription) error
	GetAllPrescription(limit int, offset int) ([]models.PatientPrescription, int64, error)
	GetPrescriptionByPatientId(patientId string, limit int, offset int) ([]models.PatientPrescription, int64, error)
	GetPatientDiseaseProfiles(patientId string) ([]models.PatientDiseaseProfile, error)
	GetPatientDiagnosticResultValue(patientId uint64, patientDiagnosticReportId uint64) ([]models.PatientDiagnosticReport, error)
	GetPatientById(patientId uint) (*models.Patient, error)
	UpdatePatientById(authUserId string, patientData *models.Patient) (*models.Patient, error)
	AddPatientRelative(relative *models.PatientRelative) error
	GetPatientRelative(patientId string) ([]models.PatientRelative, error)
	GetRelativeList(relativeUserIds []uint64) ([]models.PatientRelative, error)
	GetCaregiverList(caregiverUserIds []uint64) ([]models.Caregiver, error)
	GetDoctorList(doctorUserIds []uint64) ([]models.Doctor, error)
	GetPatientList(patientUserIds []uint64) ([]models.Patient, error)
	FetchUserIdByPatientId(patientId *uint64, mappingType string, isSelf bool) ([]uint64, error)
	GetPatientRelativeById(relativeId uint) (models.PatientRelative, error)
	UpdatePatientRelative(relativeId uint, relative *models.PatientRelative) (models.PatientRelative, error)
	AddPatientClinicalRange(customeRange *models.PatientCustomRange) error
	// UpdatePrescription(*models.PatientPrescription) error

	GetUserProfile(user_id string) (*models.SystemUser_, error)
	GetUserIdBySUB(SUB string) (uint64, error)
	IsUserBasicProfileComplete(user_id uint64) (bool, error)
	IsUserFamilyDetailsComplete(user_id uint64) (bool, error)
}

type PatientRepositoryImpl struct {
	db                *gorm.DB
	diseaseRepository DiseaseRepositoryImpl
}

func NewPatientRepository(db *gorm.DB) PatientRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &PatientRepositoryImpl{db: db}
}

func (p *PatientRepositoryImpl) GetAllRelation() ([]models.PatientRelation, error) {
	var relations []models.PatientRelation
	err := p.db.Find(&relations).Error
	return relations, err
}

// GetRelationById implements PatientRepository.
func (p *PatientRepositoryImpl) GetRelationById(relationId int) (models.PatientRelation, error) {
	var relation models.PatientRelation
	err := p.db.First(&relation, relationId).Error
	return relation, err
}

func (p *PatientRepositoryImpl) AddPatientPrescription(prescription *models.PatientPrescription) error {
	if err := p.db.Create(prescription).Error; err != nil {
		return err
	}
	return nil
}

func (p *PatientRepositoryImpl) GetAllPrescription(limit int, offset int) ([]models.PatientPrescription, int64, error) {
	var prescriptions []models.PatientPrescription
	var totalRecords int64

	query := p.db.
		Preload("PrescriptionDetails").
		Find(&prescriptions).
		Count(&totalRecords)

	if query.Error != nil {
		return nil, 0, query.Error
	}

	return prescriptions, totalRecords, nil
}

func (p *PatientRepositoryImpl) GetPrescriptionByPatientId(patientID string, limit int, offset int) ([]models.PatientPrescription, int64, error) {
	var prescriptions []models.PatientPrescription
	var totalRecords int64

	query := p.db.
		Where("patient_id = ?", patientID).
		Preload("PrescriptionDetails").
		Limit(limit).
		Offset(offset).
		Find(&prescriptions).
		Count(&totalRecords)

	if query.Error != nil {
		return nil, 0, query.Error
	}

	return prescriptions, totalRecords, nil
}

func (p *PatientRepositoryImpl) GetAllPatientPrescription(prescription *models.PatientPrescription) ([]models.PatientPrescription, int64, error) {
	var prescriptions []models.PatientPrescription
	var totalRecords int64

	query := p.db.
		Preload("PrescriptionDetails").
		Where("patient_id = ?", prescription.PatientId).
		Find(&prescriptions).
		Count(&totalRecords)

	if query.Error != nil {
		return nil, 0, query.Error
	}

	return prescriptions, totalRecords, nil
}

// GetPatientById implements PatientRepository.
func (p *PatientRepositoryImpl) GetPatientById(patientId uint) (*models.Patient, error) {
	var patient models.Patient
	err := p.db.Where("patient_id = ?", patientId).First(&patient).Error
	if err != nil {
		return nil, err
	}
	return &patient, nil
}

func (p *PatientRepositoryImpl) GetAllPatients(limit int, offset int) ([]models.Patient, int64, error) {

	var patients []models.Patient
	var totalRecords int64

	// Count total records in the table
	err := p.db.Model(&models.Patient{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}

	// Fetch paginated data
	err = p.db.Limit(limit).Offset(offset).Find(&patients).Error
	if err != nil {
		return nil, 0, err
	}

	return patients, totalRecords, nil
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

func (p *PatientRepositoryImpl) UpdatePatientById(authUserId string, patientData *models.Patient) (*models.Patient, error) {
	var user models.SystemUser_
	err := p.db.Where("auth_user_id = ?", authUserId).First(&user).Error
	if err != nil {
		return nil, err
	}
	err = p.db.Model(&user).Updates(patientData).Error
	if err != nil {
		return nil, err
	}
	return MapSystemUserToPatient(&user), nil
}

func (p *PatientRepositoryImpl) GetPatientDiseaseProfiles(PatientId string) ([]models.PatientDiseaseProfile, error) {
	var patientDiseaseProfiles []models.PatientDiseaseProfile

	err := p.db.Preload("DiseaseProfile").
		Preload("DiseaseProfile.Disease").
		Preload("DiseaseProfile.Disease.Severity").
		Preload("DiseaseProfile.Disease.Symptoms").
		Preload("DiseaseProfile.Disease.Causes").
		Preload("DiseaseProfile.Disease.DiseaseTypeMapping").
		Preload("DiseaseProfile.Disease.DiseaseTypeMapping.DiseaseType").
		Preload("DiseaseProfile.Disease.Medications").
		Preload("DiseaseProfile.Disease.Medications.MedicationTypes").
		Preload("DiseaseProfile.Disease.Exercises").
		Preload("DiseaseProfile.Disease.Exercises.ExerciseArtifact").
		Preload("DiseaseProfile.Disease.DietPlans").
		Preload("DiseaseProfile.Disease.DietPlans.Meals").
		Preload("DiseaseProfile.Disease.DietPlans.Meals.Nutrients").
		Preload("DiseaseProfile.Disease.DiagnosticTests").
		Preload("DiseaseProfile.Disease.DiagnosticTests.Components").
		// Where("patient_disease_profile_id = ?", PatientDiseaseProfileId).
		Where("patient_id = ?", PatientId).
		Find(&patientDiseaseProfiles).Error

	if err != nil {
		return nil, err
	}

	return patientDiseaseProfiles, nil
}

// GetPatientDiagnosticResultValues implements PatientRepository.
func (p *PatientRepositoryImpl) GetPatientDiagnosticResultValue(patientId uint64, patientDiagnosticReportId uint64) ([]models.PatientDiagnosticReport, error) {
	var reports []models.PatientDiagnosticReport

	query := p.db.
		Preload("DiagnosticLab").
		Preload("DiagnosticLab.PatientReportAttachments").
		Preload("DiagnosticLab.PatientDiagnosticTests").
		Preload("DiagnosticLab.PatientDiagnosticTests.DiagnosticTest").
		Preload("DiagnosticLab.PatientDiagnosticTests.DiagnosticTest.Components").
		Preload("DiagnosticLab.PatientDiagnosticTests.DiagnosticTest.Components.TestResultValue")

	if patientDiagnosticReportId > 0 {
		query = query.Where("patient_diagnostic_report_id = ?", patientDiagnosticReportId)
	} else {
		query = query.Where("patient_id = ?", patientId)
	}

	err := query.Find(&reports).Error
	if err != nil {
		return nil, err
	}

	return reports, nil
}

// AddPatientRelative implements PatientRepository.
func (p *PatientRepositoryImpl) AddPatientRelative(relative *models.PatientRelative) error {
	return p.db.Create(relative).Error
}

func (p *PatientRepositoryImpl) GetPatientRelative(patientId string) ([]models.PatientRelative, error) {
	var relatives []models.PatientRelative
	err := p.db.Where("patient_id = ?", patientId).Find(&relatives).Error
	return relatives, err
}

func (p *PatientRepositoryImpl) UpdatePatientRelative(relativeId uint, updatedRelative *models.PatientRelative) (models.PatientRelative, error) {
	var relative models.PatientRelative

	// Find the existing relative
	if err := p.db.First(&relative, "relative_id = ?", relativeId).Error; err != nil {
		return models.PatientRelative{}, err
	}

	// Update the fields
	if err := p.db.Model(&relative).Updates(updatedRelative).Error; err != nil {
		return models.PatientRelative{}, err
	}

	// Return the updated relative
	return relative, nil
}

func (s *PatientRepositoryImpl) IsPatientExists(patientID uint) (bool, error) {
	var count int64
	err := s.db.Model(&models.Patient{}).Where("patient_id = ?", patientID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *PatientRepositoryImpl) AddPatientClinicalRange(customRange *models.PatientCustomRange) error {
	tx := p.db.Begin()
	exists, err := p.IsPatientExists(customRange.PatientId)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !exists {
		tx.Rollback()
		return errors.New("patient does not exist")
	}

	exists, err = p.diseaseRepository.IsDiseaseProfileExists(customRange.DiseaseProfileId)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !exists {
		tx.Rollback()
		return errors.New("disease profile does not exist")
	}

	if err := tx.Create(customRange).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// GetPatientRelativeById implements PatientRepository.
func (p *PatientRepositoryImpl) GetPatientRelativeById(relativeId uint) (models.PatientRelative, error) {

	var relative models.PatientRelative
	err := p.db.Where("relative_id = ?", relativeId).First(&relative).Error
	if err != nil {
		return relative, err
	}
	return relative, nil
}

// GetRelativeList implements PatientRepository.
func (p *PatientRepositoryImpl) GetRelativeList(relativeUserIds []uint64) ([]models.PatientRelative, error) {
	var relatives []models.PatientRelative

	if len(relativeUserIds) == 0 {
		return relatives, nil
	}

	err := p.db.
		Table("tbl_system_user_").
		Select(`user_id AS relative_id, 
		        first_name, 
		        last_name, 
		        gender, 
		        date_of_birth, 
		        mobile_no AS mobile_no, 
		        email, 
		        created_at, 
		        updated_at`).
		Where("user_id IN ?", relativeUserIds).
		Scan(&relatives).Error

	if err != nil {
		return nil, err
	}

	return relatives, nil
}

func (p *PatientRepositoryImpl) FetchUserIdByPatientId(patientId *uint64, mappingType string, isSelf bool) ([]uint64, error) {
	var relativeUserIds []uint64
	db := p.db.Table("tbl_system_user_role_mapping")
	if patientId != nil {
		db = db.Where("patient_id = ?", *patientId)
	}
	db = db.Where("mapping_type = ? AND is_self = ?", mappingType, isSelf)
	err := db.Pluck("user_id", &relativeUserIds).Error
	if err != nil {
		return nil, err
	}
	return relativeUserIds, nil
}

// GetCaregiverList implements PatientRepository.
func (p *PatientRepositoryImpl) GetCaregiverList(caregiverUserIds []uint64) ([]models.Caregiver, error) {

	var caregivers []models.Caregiver

	if len(caregiverUserIds) == 0 {
		return caregivers, nil
	}

	err := p.db.
		Table("tbl_system_user_").
		Select(`user_id AS caregiver_id, 
		        first_name, 
		        last_name, 
		        gender, 
		        date_of_birth, 
		        mobile_no AS mobile_no, 
		        email, 
				address,
		        created_at, 
		        updated_at`).
		Where("user_id IN ?", caregiverUserIds).
		Scan(&caregivers).Error

	if err != nil {
		return nil, err
	}
	return caregivers, nil
}

// GetDoctorList implements PatientRepository.
func (p *PatientRepositoryImpl) GetDoctorList(doctorUserIds []uint64) ([]models.Doctor, error) {
	var doctors []models.Doctor
	if len(doctorUserIds) == 0 {
		return doctors, nil
	}
	err := p.db.
		Table("tbl_system_user_").
		Select(`user_id AS doctor_id,
	        first_name,
	        last_name,
	        specialty,
	        gender,
			mobile_no,
	        license_number,
	        clinic_name,
	        clinic_address,
	        email,
	        years_of_experience,
	        consultation_fee,
	        working_hours,
	        created_at,
	        updated_at`).
		Where("user_id IN ?", doctorUserIds).
		Scan(&doctors).Error

	if err != nil {
		return nil, err
	}
	return doctors, nil
}

// GetPatientList implements PatientRepository.
func (p *PatientRepositoryImpl) GetPatientList(patientUserIds []uint64) ([]models.Patient, error) {
	var patients []models.Patient

	if len(patientUserIds) == 0 {
		return patients, nil // Return empty slice if no patient IDs
	}

	err := p.db.
		Table("tbl_system_user_").
		Select(`user_id AS patient_id,
				first_name,
				last_name,
				date_of_birth,
				gender,
				mobile_no,
				address,
				emergency_contact,
				abha_number,
				blood_group,
				nationality,
				citizenship_status,
				passport_number,
				country_of_residence,
				is_indian_origin,
				email,
				created_at,
				updated_at`).
		Where("user_id IN ?", patientUserIds).
		Scan(&patients).Error

	if err != nil {
		return nil, err
	}

	return patients, nil
}

func (p *PatientRepositoryImpl) GetUserProfile(user_id string) (*models.SystemUser_, error) {
	var user models.SystemUser_
	err := p.db.Model(&models.SystemUser_{}).Where("auth_user_id=?", user_id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (p *PatientRepositoryImpl) GetUserIdBySUB(SUB string) (uint64, error) {
	var user models.SystemUser_
	err := p.db.Select("user_id").Where("auth_user_id=?", SUB).First(&user).Error
	if err != nil {
		return 0, err
	}
	return user.UserId, nil
}

func (p *PatientRepositoryImpl) IsUserBasicProfileComplete(user_id uint64) (bool, error) {
	var user models.SystemUser_
	isComplete := false
	err := p.db.Select("first_name", "last_name", "mobile_no", "email", "address", "abha_number", "emergency_contact", "gender", "date_of_birth").
		Where("user_id = ?", user_id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	isComplete = user.Gender != "" && !user.DateOfBirth.IsZero() && user.MobileNo != "" && user.Email != "" && user.Address != "" && user.AbhaNumber != "" && user.EmergencyContact != ""
	return isComplete, nil
}

func (p *PatientRepositoryImpl) IsUserFamilyDetailsComplete(user_id uint64) (bool, error) {
	var count int64
	err := p.db.Table("tbl_system_user_role_mapping").Where("patient_id = ? AND mapping_type != 'S'", user_id).Count(&count).Error
	if err != nil {
		return false, err
	}
	isComplete := count > 0
	return isComplete, nil
}
