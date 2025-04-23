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
	GetPatientDiagnosticResultValue(patientId uint64) ([]models.PatientDiagnosticReport, error)
	GetPatientById(patientId *uint64) (*models.Patient, error)
	GetUserIdByAuthUserId(authUserId string) (uint64, error)
	UpdatePatientById(authUserId string, patientData *models.Patient) (*models.Patient, error)
	AddPatientRelative(relative *models.PatientRelative) error
	GetPatientRelative(patientId string) ([]models.PatientRelative, error)
	GetRelativeList(relativeUserIds []uint64) ([]models.PatientRelative, error)
	GetCaregiverList(caregiverUserIds []uint64) ([]models.Caregiver, error)
	GetDoctorList(doctorUserIds []uint64) ([]models.Doctor, error)
	GetPatientList(patientUserIds []uint64) ([]models.Patient, error)
	FetchUserIdByPatientId(patientId *uint64, mappingType string, isSelf bool) ([]uint64, error)
	GetPatientRelativeById(relativeId uint64, relationName string) (models.PatientRelative, error)
	CheckPatientRelativeMapping(relativeId uint64, patientId uint64, mappingType string) (uint64, uint64, error)
	GetRelationNameById(relationId uint64) (models.PatientRelation, error)
	UpdatePatientRelative(relativeId uint, relative *models.PatientRelative) (models.PatientRelative, error)
	AddPatientClinicalRange(customeRange *models.PatientCustomRange) error
	// UpdatePrescription(*models.PatientPrescription) error
	GetNursesList(limit int, offset int) ([]models.Nurse, int64, error)

	GetUserProfileByUserId(user_id uint64) (*models.SystemUser_, error)
	GetUserIdBySUB(SUB string) (uint64, error)
	IsUserBasicProfileComplete(user_id uint64) (bool, error)
	IsUserFamilyDetailsComplete(user_id uint64) (bool, error)
	ExistsByUserIdAndRoleId(userId uint64, roleId uint64) (bool, error)
	FetchPatientDiagnosticTrendValue(input models.DiagnosticResultRequest) ([]map[string]interface{}, error)
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
func (p *PatientRepositoryImpl) GetPatientById(patientId *uint64) (*models.Patient, error) {
	var patient models.Patient
	err := p.db.Where("patient_id = ?", &patientId).First(&patient).Error
	if err != nil {
		return nil, err
	}
	return &patient, nil
}

func (p *PatientRepositoryImpl) GetUserIdByAuthUserId(authUserId string) (uint64, error) {
	var userId uint64
	query := `SELECT user_id FROM tbl_system_user_ WHERE auth_user_id = ? LIMIT 1`
	err := p.db.Raw(query, authUserId).Scan(&userId).Error
	if err != nil {
		return 0, err
	}

	return userId, nil
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

func (p *PatientRepositoryImpl) GetPatientDiagnosticResultValue(patientId uint64) ([]models.PatientDiagnosticReport, error) {
	var reports []models.PatientDiagnosticReport

	query := p.db.
		Preload("DiagnosticLab").
		Preload("DiagnosticLab.PatientReportAttachments").
		Preload("DiagnosticLab.PatientDiagnosticTests").
		Preload("DiagnosticLab.PatientDiagnosticTests.DiagnosticTest").
		Preload("DiagnosticLab.PatientDiagnosticTests.DiagnosticTest.Components").
		Preload("DiagnosticLab.PatientDiagnosticTests.DiagnosticTest.Components.ReferenceRange").
		Preload("DiagnosticLab.PatientDiagnosticTests.DiagnosticTest.Components.TestResultValue")

	if patientId > 0 {
		query = query.Where("patient_id = ?", patientId)
		// query = query.Where("patient_diagnostic_report_id = ?", patientDiagnosticReportId)
	}

	err := query.Order("patient_diagnostic_report_id ASC").Find(&reports).Error
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

func (p *PatientRepositoryImpl) CheckPatientRelativeMapping(relativeId uint64, patientId uint64, mappingType string) (uint64, uint64, error) {
	var userId uint64
	var relationId uint64

	row := p.db.Raw(`SELECT user_id, relation_id FROM tbl_system_user_role_mapping 
	WHERE user_id = ? AND patient_id = ? AND mapping_type = ? LIMIT 1`, relativeId, patientId, mappingType).Row()

	err := row.Scan(&userId, &relationId)
	if err != nil {
		return 0, 0, err
	}

	return userId, relationId, nil
}

func (p *PatientRepositoryImpl) GetPatientRelativeById(relativeId uint64, relationName string) (models.PatientRelative, error) {
	var relative models.PatientRelative
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
		Where("user_id = ?", relativeId).
		Scan(&relative).Error
	relative.Relationship = relationName
	return relative, err
}

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

func (p *PatientRepositoryImpl) GetUserProfileByUserId(user_id uint64) (*models.SystemUser_, error) {
	var user models.SystemUser_
	err := p.db.Model(&models.SystemUser_{}).Where("user_id=?", user_id).First(&user).Error
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

func (p *PatientRepositoryImpl) GetNursesList(limit int, offset int) ([]models.Nurse, int64, error) {
	var nurses []models.Nurse
	var totalRecords int64
	if err := p.db.
		Table("tbl_system_user_ AS u").
		Joins("JOIN tbl_system_user_role_mapping AS m ON u.user_id = m.user_id").
		Where("m.mapping_type = ?", "N").
		Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}
	err := p.db.
		Table("tbl_system_user_").
		Select(`tbl_system_user_.user_id as nurse_id, 
				tbl_system_user_.first_name, 
				tbl_system_user_.last_name, 
				tbl_system_user_.specialty,
				tbl_system_user_.gender,
				tbl_system_user_.mobile_no,
				tbl_system_user_.license_number,
				tbl_system_user_.clinic_name,
				tbl_system_user_.clinic_address,
				tbl_system_user_.email,
				tbl_system_user_.years_of_experience,
				tbl_system_user_.consultation_fee,
				tbl_system_user_.working_hours,
				tbl_system_user_.created_at,
				tbl_system_user_.updated_at`).
		Joins("JOIN tbl_system_user_role_mapping ON tbl_system_user_.user_id = tbl_system_user_role_mapping.user_id").
		Where("tbl_system_user_role_mapping.mapping_type = ?", "N").
		Limit(limit).
		Offset(offset).
		Scan(&nurses).Error
	if err != nil {
		return nil, 0, err
	}
	return nurses, totalRecords, nil
}

func (p *PatientRepositoryImpl) ExistsByUserIdAndRoleId(userId uint64, roleId uint64) (bool, error) {
	var count int64
	err := p.db.Table("tbl_system_user_role_mapping").
		Where("user_id = ? AND role_id = ?", userId, roleId).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (pr *PatientRepositoryImpl) FetchPatientDiagnosticTrendValue(input models.DiagnosticResultRequest) ([]map[string]interface{}, error) {

	query := `
	SELECT
		pdr.patient_diagnostic_report_id,
		pdr.patient_id,
		pdr.collected_date,
		pdr.report_date,
		pdr.report_status,
		pdt.patient_test_id,
		pdt.test_note,
		pdt.test_date,
		pdtrv.diagnostic_test_id,
		pdtrv.diagnostic_test_component_id,
		tdpdtcm.test_component_name,
		pdtrv.result_value,
		dtrr.normal_min,
		dtrr.normal_max,
		dtrr.units,
		pdtrv.result_status,
		pdtrv.result_date,
		pdtrv.result_comment
	FROM
		tbl_patient_diagnostic_report pdr
	INNER JOIN tbl_patient_diagnostic_test pdt 
		ON pdr.patient_diagnostic_report_id = pdt.patient_diagnostic_report_id
	INNER JOIN tbl_patient_diagnostic_test_result_value pdtrv 
		ON pdt.patient_diagnostic_report_id = pdtrv.patient_diagnostic_report_id
	LEFT JOIN tbl_diagnostic_test_reference_range dtrr 
		ON pdtrv.diagnostic_test_component_id = dtrr.diagnostic_test_component_id
	LEFT JOIN tbl_disease_profile_diagnostic_test_component_master tdpdtcm 
		ON tdpdtcm.diagnostic_test_component_id = pdtrv.diagnostic_test_component_id
	WHERE
		pdr.patient_id = ?
	`

	args := []interface{}{input.PatientId}

	if input.DiagnosticTestComponentId != nil {
		query += " AND pdtrv.diagnostic_test_component_id = ?"
		args = append(args, *input.DiagnosticTestComponentId)
	}

	if input.PatientDiagnosticReportId != nil {
		query += " AND pdtrv.patient_diagnostic_report_id = ?"
		args = append(args, *input.PatientDiagnosticReportId)
	}

	if input.ReportDateStart != nil && input.ReportDateEnd != nil {
		query += " AND pdr.report_date BETWEEN ? AND ?"
		args = append(args, *input.ReportDateStart, *input.ReportDateStart)
	}

	if input.ResultDateStart != nil && input.ResultDateEnd != nil {
		query += " AND pdtrv.result_date BETWEEN ? AND ?"
		args = append(args, *input.ResultDateStart, *input.ResultDateEnd)
	}

	rows, err := pr.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		results = append(results, rowMap)
	}

	return results, nil
}

// GetRelationNameById implements PatientRepository.
func (p *PatientRepositoryImpl) GetRelationNameById(relationId uint64) (models.PatientRelation, error) {
	var relation models.PatientRelation
	if err := p.db.Where("relation_id = ?", relationId).First(&relation).Error; err != nil {
		return models.PatientRelation{}, err
	}
	return relation, nil
}
