package repository

import (
	"biostat/database"
	"biostat/models"
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"
)

type PatientRepository interface {
	GetAllRelation() ([]models.PatientRelation, error)
	GetRelationById(relationId int) (models.PatientRelation, error)
	GetAllPatients(limit int, offset int) ([]models.Patient, int64, error)
	AddPatientPrescription(createdBy string, prescription *models.PatientPrescription) error
	GetPrescriptionByPatientId(patientId uint64, limit int, offset int) ([]models.PatientPrescription, int64, error)
	GetPatientDiseaseProfiles(patientId uint64, AttachedFlag int) ([]models.PatientDiseaseProfile, error)
	AddPatientDiseaseProfile(tx *gorm.DB, input *models.PatientDiseaseProfile) (*models.PatientDiseaseProfile, error)
	UpdateFlag(patientId uint64, req *models.DPRequest) error
	GetPatientDiagnosticResultValue(patientId uint64, patientDiagnosticReportId uint64) ([]models.PatientDiagnosticReport, error)
	GetUserIdByAuthUserId(authUserId string) (uint64, error)
	UpdatePatientById(userId uint64, patientData *models.Patient) (models.SystemUser_, error)
	UpdateUserAddressByUserId(userId uint64, newaddress models.AddressMaster) (models.AddressMaster, error)

	MapSystemUserToPatient(updatedPatient *models.SystemUser_, updatedAddress models.AddressMaster) *models.Patient
	AddPatientRelative(relative *models.PatientRelative) error
	GetPatientRelative(patientId string) ([]models.PatientRelative, error)
	GetRelativeList(relativeUserIds []uint64, userRelation []models.UserRelation, relation []models.PatientRelation) ([]models.PatientRelative, error)
	GetCaregiverList(caregiverUserIds []uint64) ([]models.Caregiver, error)
	GetDoctorList(doctorUserIds []uint64) ([]models.Doctor, error)
	GetPatientList(patientUserIds []uint64) ([]models.Patient, error)
	FetchUserIdByPatientId(patientId *uint64, mappingType string, isSelf bool) ([]models.UserRelation, error)
	GetPatientRelativeById(relativeId uint64, relation []models.PatientRelation) (models.PatientRelative, error)
	CheckPatientRelativeMapping(relativeId uint64, patientId uint64, mappingType string) (uint64, uint64, error)
	GetRelationNameById(relationId []uint64) ([]models.PatientRelation, error)
	UpdatePatientRelative(relativeId uint, relative *models.PatientRelative) (models.PatientRelative, error)
	AddPatientClinicalRange(customeRange *models.PatientCustomRange) error
	// UpdatePrescription(*models.PatientPrescription) error
	GetNursesList(limit int, offset int) ([]models.Nurse, int64, error)

	GetUserProfileByUserId(user_id uint64) (*models.SystemUser_, error)
	GetUserDataUserId(userId []uint64, limit, offset int) ([]models.SystemUser_, int64, error)
	GetUserIdBySUB(SUB string) (uint64, error)
	IsUserBasicProfileComplete(user_id uint64) (bool, error)
	IsUserFamilyDetailsComplete(user_id uint64) (bool, error)
	IsUserHealthDetailsComplete(user_id uint64) (bool, error)
	ExistsByUserIdAndRoleId(userId uint64, roleId uint64) (bool, error)
	FetchPatientDiagnosticTrendValue(input models.DiagnosticResultRequest) ([]map[string]interface{}, error)
	GetUserSUBByID(ID uint64) (string, error)
	NoOfUpcomingAppointments(patientID uint64) (int64, error)
	NoOfMedicationsForDashboard(patientID uint64) (int64, error)
	NoOfMessagesForDashboard(patientID uint64) (int64, error)
	NoOfLabReusltsForDashboard(patientID uint64) (int64, error)

	SaveUserHealthProfile(tx *gorm.DB, input *models.TblPatientHealthProfile) (*models.TblPatientHealthProfile, error)
}

type PatientRepositoryImpl struct {
	db                *gorm.DB
	diseaseRepository DiseaseRepositoryImpl
	userRepo          UserRepositoryImpl
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

func (p *PatientRepositoryImpl) AddPatientPrescription(createdBy string, prescription *models.PatientPrescription) error {
	tx := p.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	for i := range prescription.PrescriptionDetails {
		prescription.PrescriptionDetails[i].PrescriptionDetailId = 0
		prescription.PrescriptionDetails[i].CreatedBy = createdBy
	}
	if err := tx.Create(&prescription).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (p *PatientRepositoryImpl) GetPrescriptionByPatientId(patientId uint64, limit int, offset int) ([]models.PatientPrescription, int64, error) {
	var prescriptions []models.PatientPrescription
	var totalRecords int64

	query := p.db.
		Where("patient_id = ?", patientId).
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

func (p *PatientRepositoryImpl) MapSystemUserToPatient(user *models.SystemUser_, updatedAddress models.AddressMaster) *models.Patient {
	return &models.Patient{
		PatientId:   user.UserId,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DateOfBirth: user.DateOfBirth.String(),
		Gender:      user.Gender,
		MobileNo:    user.MobileNo,
		Address:     user.Address,
		UserAddress: models.AddressMaster{
			AddressId:    updatedAddress.AddressId,
			AddressLine1: updatedAddress.AddressLine1,
			AddressLine2: updatedAddress.AddressLine2,
			Landmark:     updatedAddress.Landmark,
			City:         updatedAddress.City,
			State:        updatedAddress.State,
			Country:      updatedAddress.Country,
			PostalCode:   updatedAddress.PostalCode,
			Latitude:     updatedAddress.Latitude,
			Longitude:    updatedAddress.Longitude,
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

func (p *PatientRepositoryImpl) UpdatePatientById(userId uint64, patientData *models.Patient) (models.SystemUser_, error) {
	var user models.SystemUser_
	err := p.db.Where("user_id = ?", userId).First(&user).Error
	if err != nil {
		return models.SystemUser_{}, err
	}
	err = p.db.Model(&user).Updates(patientData).Error
	if err != nil {
		return models.SystemUser_{}, err
	}
	return user, nil
}

func (p *PatientRepositoryImpl) UpdateUserAddressByUserId(userId uint64, newAddress models.AddressMaster) (models.AddressMaster, error) {
	var user models.SystemUser_
	err := p.db.Where("user_id = ?", userId).First(&user).Error
	if err != nil {
		return models.AddressMaster{}, err
	}
	var addressMapping models.SystemUserAddressMapping
	err = p.db.Where("user_id = ?", user.UserId).First(&addressMapping).Error
	if err != nil {
		newAddress, err = p.userRepo.CreateSystemUserAddress(p.db, newAddress)
		if err != nil {
			return models.AddressMaster{}, err
		}
		addressMapping = models.SystemUserAddressMapping{
			UserId:    user.UserId,
			AddressId: newAddress.AddressId,
		}
		MappingErr := p.userRepo.CreateSystemUserAddressMapping(p.db, addressMapping)
		if MappingErr != nil {
			return models.AddressMaster{}, err
		}
	} else {
		err = p.db.Model(&newAddress).Where("address_id = ?", addressMapping.AddressId).Updates(newAddress).Error
		if err != nil {
			return models.AddressMaster{}, err
		}
	}
	return newAddress, nil
}

func (p *PatientRepositoryImpl) GetPatientDiseaseProfiles(PatientId uint64, AttachedFlag int) ([]models.PatientDiseaseProfile, error) {
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
		Where("patient_id = ? AND attached_flag = ?", PatientId, AttachedFlag).
		Find(&patientDiseaseProfiles).Error

	if err != nil {
		return nil, err
	}

	return patientDiseaseProfiles, nil
}

func (p *PatientRepositoryImpl) AddPatientDiseaseProfile(tx *gorm.DB, input *models.PatientDiseaseProfile) (*models.PatientDiseaseProfile, error) {
	var existingProfile models.PatientDiseaseProfile
	err := tx.Where("patient_id = ? AND disease_profile_id = ?", input.PatientId, input.DiseaseProfileId).First(&existingProfile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(input).Error; err != nil {
				return nil, err
			}
			return input, nil
		}
		return nil, err
	}
	err = tx.Model(&existingProfile).Update("attached_flag", 0).Error
	if err != nil {
		return nil, err
	}

	return &existingProfile, nil
}

func (ps *PatientRepositoryImpl) UpdateFlag(patientId uint64, req *models.DPRequest) error {
	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Recovered in UpdateFlag:", r)
		}
	}()

	query := tx.Model(&models.PatientDiseaseProfile{}).
		Where("patient_id = ? AND disease_profile_id = ?", patientId, req.DiseaseProfileId)

	switch req.Flag {
	case "profile":
		if err := query.Update("attached_flag", req.AttachedFlag).Error; err != nil {
			tx.Rollback()
			return err
		}
	case "diet":
		if err := query.Update("diet_plan_subscribed", req.DietPlanSubscibed).Error; err != nil {
			tx.Rollback()
			return err
		}
	case "reminder":
		if err := query.Update("reminder_flag", req.ReminderFlag).Error; err != nil {
			tx.Rollback()
			return err
		}
	default:
		tx.Rollback()
		return fmt.Errorf("invalid flag: %s", req.Flag)
	}

	return tx.Commit().Error
}

// func (p *PatientRepositoryImpl) GetPatientDiagnosticResultValue(patientId uint64, patientDiagnosticReportId uint64) ([]models.PatientDiagnosticReport, error) {
// 	var reports []models.PatientDiagnosticReport

// 	query := p.db.Debug().
// 		Model(&models.PatientDiagnosticReport{}).Where("patient_id = ?", patientId).
// 		Preload("DiagnosticLab").
// 		Preload("DiagnosticLab.PatientDiagnosticTests").
// 		Preload("DiagnosticLab.PatientReportAttachments").
// 		Preload("DiagnosticLab.PatientDiagnosticTests.DiagnosticTest").
// 		Preload("DiagnosticLab.PatientDiagnosticTests.DiagnosticTest.Components").
// 		Preload("DiagnosticLab.PatientDiagnosticTests.DiagnosticTest.Components.ReferenceRange")

// 	if patientId > 0 && patientDiagnosticReportId > 0 {
// 		query = query.Preload("DiagnosticLab.PatientDiagnosticTests.DiagnosticTest.Components.TestResultValue",
// 			"patient_id = ? AND patient_diagnostic_report_id = ?", patientId, patientDiagnosticReportId)
// 	} else {
// 		query = query.Preload("DiagnosticLab.PatientDiagnosticTests.DiagnosticTest.Components.TestResultValue",
// 			"patient_id = ?", patientId)
// 	}

// 	err := query.Order("patient_diagnostic_report_id ASC").Find(&reports).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return reports, nil
// }

func (p *PatientRepositoryImpl) GetPatientDiagnosticResultValue(patientId uint64, patientDiagnosticReportId uint64) ([]models.PatientDiagnosticReport, error) {

	_, uniqueReportIds, err := p.GetPatientDiagnosticReportIds(patientId, patientDiagnosticReportId)
	if err != nil {
		log.Printf("Failed to get patient diagnostic report and lab: %v", err)
	}
	reportsWithDetails, err := p.GetPatientDiagnosticTestResult(uniqueReportIds)
	if err != nil {
		log.Printf("Failed to get patient diagnostic tests: %v", err)
	}
	return reportsWithDetails, nil
}

func (p *PatientRepositoryImpl) RestructurePatientDiagnosticReport(reports []models.PatientDiagnosticReport) []map[string]interface{} {
	restructuredResponse := make([]map[string]interface{}, len(reports))
	for i, report := range reports {
		restructured := map[string]interface{}{
			"patient_diagnostic_report_id": report.PatientDiagnosticReportId,
			"patient_id":                   report.PatientId,
			"payment_status":               report.PaymentStatus,
			"report_name":                  report.ReportName,
			"collected_date":               report.CollectedDate,
			"collected_at":                 report.CollectedAt,
			"processed_at":                 report.ProcessedAt,
			"report_date":                  report.ReportDate,
			"report_status":                report.ReportStatus,
			"observation":                  report.Observation,
			"comments":                     report.Comments,
			"review_by":                    report.ReviewBy,
			"review_date":                  report.ReviewDate,
			"shared_flag":                  report.SharedFlag,
			"shared_with":                  report.SharedWith,
			"diagnostic_lab":               report.DiagnosticLabs,
		}
		if lab, ok := restructured["diagnostic_lab"].(models.DiagnosticLab); ok {
			lab.PatientDiagnosticTests = report.PatientDiagnosticTests
			restructured["diagnostic_lab"] = lab
		} else if labs, ok := restructured["diagnostic_lab"].([]models.DiagnosticLab); ok && len(labs) > 0 {
			labs[0].PatientDiagnosticTests = report.PatientDiagnosticTests
			restructured["diagnostic_lab"] = labs[0]
		}
		restructuredResponse[i] = restructured
	}
	return restructuredResponse
}

func (p *PatientRepositoryImpl) GetPatientDiagnosticReportIds(patientId uint64, reportId uint64) (map[uint64]models.PatientDiagnosticReport, []uint64, error) {
	var reports []models.PatientDiagnosticReport
	query := p.db.Debug().Joins("DiagnosticLabs").Where("patient_id = ?", patientId)

	if reportId > 0 {
		query = query.Where("patient_diagnostic_report_id = ?", reportId)
	}

	result := query.Find(&reports)

	if result.Error != nil {
		log.Printf("GORM error fetching report and lab: %v", result.Error)
		return nil, nil, fmt.Errorf("error fetching report and lab: %w", result.Error)
	}

	reportsMap := make(map[uint64]models.PatientDiagnosticReport)
	uniqueReportIDs := make([]uint64, 0)
	reportIDSet := make(map[uint64]bool)

	for _, report := range reports {
		reportsMap[report.PatientDiagnosticReportId] = report
		if !reportIDSet[report.PatientDiagnosticReportId] {
			uniqueReportIDs = append(uniqueReportIDs, report.PatientDiagnosticReportId)
			reportIDSet[report.PatientDiagnosticReportId] = true
		}
	}

	return reportsMap, uniqueReportIDs, nil
}

func (p *PatientRepositoryImpl) GetPatientDiagnosticTestResult(reportIds []uint64) ([]models.PatientDiagnosticReport, error) {
	var patientReport []models.PatientDiagnosticReport
	result := p.db.Debug().Model(&models.PatientDiagnosticReport{}).
		Preload("DiagnosticLabs").
		Preload("DiagnosticLabs.PatientReportAttachments").
		Preload("PatientDiagnosticTests.DiagnosticTest").
		Preload("PatientDiagnosticTests.DiagnosticTest.Components").
		Preload("PatientDiagnosticTests.DiagnosticTest.Components.TestResultValue").
		Preload("PatientDiagnosticTests.DiagnosticTest.Components.ReferenceRange").
		Where("patient_diagnostic_report_id IN (?)", reportIds).
		Find(&patientReport)

	if result.Error != nil {
		log.Printf("GORM error fetching patient diagnostic tests: %v", result.Error)
		return nil, fmt.Errorf("error fetching patient diagnostic tests: %w", result.Error)
	}

	// reportMap := make(map[uint64]models.PatientDiagnosticReport)
	// for _, test := range patientReport {
	// 	reportMap[test.PatientDiagnosticReportId] = test
	// }

	return patientReport, nil
}

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

func (p *PatientRepositoryImpl) GetPatientRelativeById(relativeId uint64, relations []models.PatientRelation) (models.PatientRelative, error) {
	relatives, err := p.fetchRelatives([]uint64{relativeId})
	if err != nil || len(relatives) == 0 {
		return models.PatientRelative{}, err
	}

	relative := relatives[0]
	for _, r := range relations {
		if relativeId == r.RelationId {
			relative.Relationship = r.RelationShip
			break
		}
	}

	return relative, nil
}

func (p *PatientRepositoryImpl) GetRelativeList(relativeUserIds []uint64, userRelations []models.UserRelation, relationData []models.PatientRelation) ([]models.PatientRelative, error) {
	relativeInfo, err := p.fetchRelatives(relativeUserIds)
	if err != nil {
		return nil, err
	}
	relationMap := make(map[uint64]string)
	for _, r := range relationData {
		relationMap[r.RelationId] = r.RelationShip
	}

	userToRelationIdMap := make(map[uint64]uint64)
	for _, ur := range userRelations {
		userToRelationIdMap[ur.UserId] = ur.RelationId
	}

	for i := range relativeInfo {
		uid := relativeInfo[i].RelativeId
		if relId, ok := userToRelationIdMap[uid]; ok {
			if relationName, ok := relationMap[relId]; ok {
				relativeInfo[i].RelationId = relId
				relativeInfo[i].Relationship = relationName
			}
		}
	}

	return relativeInfo, nil
}

func (p *PatientRepositoryImpl) fetchRelatives(userIds []uint64) ([]models.PatientRelative, error) {
	var relatives []models.PatientRelative

	if len(userIds) == 0 {
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
		Where("user_id IN ?", userIds).
		Scan(&relatives).Error

	return relatives, err
}

func (p *PatientRepositoryImpl) FetchUserIdByPatientId(patientId *uint64, mappingType string, isSelf bool) ([]models.UserRelation, error) {
	var userRelations []models.UserRelation

	db := p.db.Table("tbl_system_user_role_mapping")
	if patientId != nil {
		db = db.Where("patient_id = ?", *patientId)
	}
	db = db.Where("mapping_type = ? AND is_self = ?", mappingType, isSelf)
	err := db.Select("user_id,relation_id").Scan(&userRelations).Error
	if err != nil {
		return nil, err
	}
	return userRelations, nil
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
	        speciality,
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

	// Fetch address for each doctor separately
	for i, doc := range doctors {
		var address models.AddressMaster
		err := p.db.Where("user_id = ?", doc.DoctorId).First(&address).Error
		if err == nil {
			doctors[i].UserAddress = address
		}
	}

	return doctors, nil
}

func (p *PatientRepositoryImpl) GetPatientList(patientUserIds []uint64) ([]models.Patient, error) {
	var patients []models.Patient

	if len(patientUserIds) == 0 {
		return patients, nil
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
	err := p.db.Model(&models.SystemUser_{}).Preload("AddressMapping.Address").Where("user_id=?", user_id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (p *PatientRepositoryImpl) GetUserDataUserId(user_ids []uint64, limit, offset int) ([]models.SystemUser_, int64, error) {
	var users []models.SystemUser_
	var total int64
	query := p.db.Debug().Model(&models.SystemUser_{}).
		Preload("AddressMapping.Address").
		Where("user_id IN ?", user_ids)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
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
	err := p.db.Select("first_name", "last_name", "mobile_no", "email", "abha_number", "gender", "date_of_birth").
		Where("user_id = ?", user_id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	isComplete = user.Gender != "" && !user.DateOfBirth.IsZero() && user.MobileNo != "" && user.Email != "" && user.AbhaNumber != ""
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

func (p *PatientRepositoryImpl) IsUserHealthDetailsComplete(user_id uint64) (bool, error) {
	var count int64
	err := p.db.Table("tbl_patient_health_profile").Where("patient_id = ?", user_id).Count(&count).Error
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
				tbl_system_user_.speciality,
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
		ON pdt.diagnostic_test_id = pdtrv.diagnostic_test_id
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

func (p *PatientRepositoryImpl) GetRelationNameById(ids []uint64) ([]models.PatientRelation, error) {
	uniqueIds := make(map[uint64]struct{})
	for _, id := range ids {
		uniqueIds[id] = struct{}{}
	}

	var relationIds []uint64
	for id := range uniqueIds {
		relationIds = append(relationIds, id)
	}

	relationMap := make(map[uint64]string)
	var relations []models.PatientRelation
	err := p.db.Where("relation_id IN ?", relationIds).Find(&relations).Error
	if err != nil {
		return nil, err
	}
	for _, r := range relations {
		relationMap[r.RelationId] = r.RelationShip
	}

	var orderedRelations []models.PatientRelation
	for _, id := range ids {
		if relationName, ok := relationMap[id]; ok {
			orderedRelations = append(orderedRelations, models.PatientRelation{
				RelationId:   id,
				RelationShip: relationName,
			})
		}
	}

	return orderedRelations, nil
}

func (p *PatientRepositoryImpl) SaveUserHealthProfile(tx *gorm.DB, input *models.TblPatientHealthProfile) (*models.TblPatientHealthProfile, error) {
	err := tx.Create(input).Error
	if err != nil {
		return nil, err
	}
	return input, nil
}

func (p *PatientRepositoryImpl) NoOfUpcomingAppointments(patientID uint64) (int64, error) {
	var count int64
	err := p.db.Table("tbl_appointment_master").
		Where("patient_id = ? AND appointment_date > CURRENT_DATE AND is_deleted = 0 AND status NOT IN ?",
			patientID, []string{"cancelled", "completed"}).
		Count(&count).Error

	if err != nil {
		return 0, err
	}
	return count, nil
}

func (p *PatientRepositoryImpl) NoOfMedicationsForDashboard(patientID uint64) (int64, error) {
	return 0, nil
}

func (p *PatientRepositoryImpl) NoOfMessagesForDashboard(patientID uint64) (int64, error) {
	return 0, nil
}

func (p *PatientRepositoryImpl) NoOfLabReusltsForDashboard(patientID uint64) (int64, error) {
	var count int64
	err := p.db.Table("tbl_patient_diagnostic_report").
		Where("patient_id = ?", patientID).
		Count(&count).Error

	if err != nil {
		return 0, err
	}
	return count, nil
}

func (p *PatientRepositoryImpl) GetUserSUBByID(ID uint64) (string, error) {
	var user models.SystemUser_
	err := p.db.Select("auth_user_id").Where("user_id=?", ID).First(&user).Error
	if err != nil {
		return "", err
	}
	return user.AuthUserId, nil
}
