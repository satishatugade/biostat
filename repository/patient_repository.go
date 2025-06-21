package repository

import (
	"biostat/database"
	"biostat/models"
	"biostat/utils"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type PatientRepository interface {
	GetAllGender() ([]models.GenderMaster, error)
	GetGenderById(genderId uint64) (models.GenderMaster, error)
	GetAllRelation() ([]models.PatientRelation, error)
	GetRelationById(relationId uint64) (models.PatientRelation, error)
	GetAllPatients(limit int, offset int) ([]models.Patient, int64, error)
	AddPatientPrescription(createdBy string, prescription *models.PatientPrescription) error
	UpdatePatientPrescription(authUserId string, prescription *models.PatientPrescription) error
	GetSinglePrescription(prescriptiuonId uint64, patientId uint64) (models.PatientPrescription, error)
	GetPrescriptionByPatientId(patientId uint64, limit int, offset int) ([]models.PatientPrescription, int64, error)
	GetPrescriptionDetailByPatientId(PatientId uint64, limit int, offset int) ([]models.PrescriptionDetail, int64, error)
	GetPatientDiseaseProfiles(patientId uint64, AttachedFlag int) ([]models.PatientDiseaseProfile, error)
	AddPatientDiseaseProfile(tx *gorm.DB, input *models.PatientDiseaseProfile) (*models.PatientDiseaseProfile, error)
	UpdateFlag(patientId uint64, req *models.DPRequest) error
	GetPatientDiagnosticResultValue(patientId uint64, patientDiagnosticReportId uint64) ([]models.PatientDiagnosticReport, map[uint64]uint64, error)
	UpdatePatientById(userId uint64, patientData *models.Patient) (models.SystemUser_, error)
	UpdateUserAddressByUserId(userId uint64, newaddress models.AddressMaster) (models.AddressMaster, error)
	GetDistinctMedicinesByPatientID(patientID uint64) ([]models.UserMedicineInfo, error)

	MapSystemUserToPatient(updatedPatient *models.SystemUser_, updatedAddress models.AddressMaster) *models.Patient
	AddPatientRelative(relative *models.PatientRelative) error
	AssignPrimaryCaregiver(patientId uint64, relativeId uint64, mappingType string) error
	SetCaregiverMappingDeletedStatus(patientId uint64, caregiverId uint64, isDeleted int) error
	GetPatientRelative(patientId string) ([]models.PatientRelative, error)
	GetRelativeList(relativeUserIds []uint64, userRelation []models.UserRelation, relation []models.PatientRelation) ([]models.PatientRelative, error)
	GetCaregiverList(caregiverUserIds []uint64) ([]models.Caregiver, error)
	GetDoctorList(doctorUserIds []uint64) ([]models.Doctor, error)
	GetPatientList(patientUserIds []uint64) ([]models.Patient, error)
	FetchUserIdByPatientId(patientId *uint64, mappingType []string, isSelf bool, isDeleted int) ([]models.UserRelation, error)
	FetchPatientIdByUserId(patientId *uint64, mappingType []string, isSelf bool, isDeleted int) ([]models.UserRelation, error)
	GetPatientRelativeById(relativeId uint64, relation []models.PatientRelation) (models.PatientRelative, error)
	CheckPatientRelativeMapping(relativeId uint64, patientId uint64, mappingType string) (uint64, uint64, error)
	GetRelationNameById(relationId []uint64) ([]models.PatientRelation, error)
	AddPatientClinicalRange(customeRange *models.PatientCustomRange) error
	GetNursesList(limit int, offset int) ([]models.Nurse, int64, error)

	GetUserProfileByUserId(user_id uint64) (*models.SystemUser_, error)
	GetUserDataUserId(userId []uint64, limit, offset int) ([]models.SystemUser_, int64, error)
	IsUserBasicProfileComplete(user_id uint64) (bool, error)
	IsUserFamilyDetailsComplete(user_id uint64) (bool, error)
	IsUserHealthDetailsComplete(user_id uint64) (bool, error)
	GetPatientHealthDetail(patientId uint64) (models.TblPatientHealthProfile, error)
	ExistsByUserIdAndRoleId(userId uint64, roleId uint64) (bool, error)
	FetchPatientDiagnosticTrendValue(input models.DiagnosticResultRequest) ([]map[string]interface{}, error)
	ParseDiagnosticTrendData(rawData []map[string]interface{}) ([]map[string]interface{}, error)
	GetUserSUBByID(ID uint64) (string, error)
	NoOfUpcomingAppointments(patientID uint64) (int64, error)
	NoOfMedicationsForDashboard(patientID uint64) (int64, error)
	NoOfMessagesForDashboard(patientID uint64) (int64, error)
	NoOfLabReusltsForDashboard(patientID uint64) (int64, error)
	FetchPatientDiagnosticReports(patientID uint64, filter models.DiagnosticReportFilter) ([]models.DiagnosticReportResponse, error)
	GetPatientDiagnosticReportResult(patientID uint64, filter models.DiagnosticReportFilter, limit, offset int) ([]models.ReportRow, int64, error)
	ProcessReportGridData(rows []models.ReportRow) map[string]interface{}
	RestructureDiagnosticReports(data []models.DiagnosticReportResponse) []map[string]interface{}
	GetDiagnosticReportId(patientId uint64) (uint64, error)

	SaveUserHealthProfile(tx *gorm.DB, input *models.TblPatientHealthProfile) (*models.TblPatientHealthProfile, error)
	CheckPatientHealthProfileExist(tx *gorm.DB, patientId uint64) (bool, error)
	UpdatePatientHealthDetail(req *models.TblPatientHealthProfile) error
	AddTestComponentDisplayConfig(config *models.PatientTestComponentDisplayConfig) error
	GetPinnedComponentCount(patientId uint64) (int64, error)

	GrantPermission(userID, relativeID uint64, permissionID int64, granted bool) error
	HasPermission(userID, relativeID uint64, permissionCode string) (bool, error)
	ListPermissions(userID, relativeID uint64) ([]models.PermissionResult, error)
	GetPermissionByCode(code string) (*models.PermissionMaster, error)
	CheckPermissionValue(userID, relativeID uint64, permissionID int64) (exists bool, currentValue bool, err error)
	UpdatePermissionValue(userID, relativeID uint64, permissionID int64, value bool) error
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

func (p *PatientRepositoryImpl) GetRelationById(relationId uint64) (models.PatientRelation, error) {
	var relation models.PatientRelation
	err := p.db.First(&relation, relationId).Error
	return relation, err
}

func (p *PatientRepositoryImpl) GetAllGender() ([]models.GenderMaster, error) {
	var genders []models.GenderMaster
	err := p.db.Where("is_deleted = ?", 0).Find(&genders).Error
	return genders, err
}

func (p *PatientRepositoryImpl) GetGenderById(genderId uint64) (models.GenderMaster, error) {
	var gender models.GenderMaster
	err := p.db.First(&gender, genderId).Error
	return gender, err
}

func (p *PatientRepositoryImpl) AddPatientPrescription(createdBy string, prescription *models.PatientPrescription) error {
	tx := p.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	for i := range prescription.PrescriptionDetails {
		prescription.PrescriptionDetails[i].PrescriptionDetailId = 0

		for j := range prescription.PrescriptionDetails[i].DosageInfo {
			prescription.PrescriptionDetails[i].DosageInfo[j].DoseScheduleId = 0
			prescription.PrescriptionDetails[i].DosageInfo[j].CreatedBy = createdBy
		}
	}
	if err := tx.Create(&prescription).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (ps *PatientRepositoryImpl) UpdatePatientPrescription(authUserId string, prescription *models.PatientPrescription) error {
	tx := ps.db.Begin()
	if err := tx.Model(&models.PatientPrescription{}).
		Where("prescription_id = ? AND patient_id = ?", prescription.PrescriptionId, prescription.PatientId).
		Updates(map[string]interface{}{
			"prescribed_by":     prescription.PrescribedBy,
			"prescription_name": prescription.PrescriptionName,
			"description":       prescription.Description,
			"prescription_date": prescription.PrescriptionDate,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, detail := range prescription.PrescriptionDetails {
		detail.PrescriptionId = prescription.PrescriptionId

		if detail.PrescriptionDetailId == 0 {
			if err := tx.Create(&detail).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			updateMap := map[string]interface{}{}
			if detail.MedicineName != "" {
				updateMap["medicine_name"] = detail.MedicineName
			}
			if detail.PrescriptionType != "" {
				updateMap["prescription_type"] = detail.PrescriptionType
			}
			if detail.Duration != 0 {
				updateMap["duration"] = detail.Duration
			}
			if detail.DurationUnitType != "" {
				updateMap["duration_unit_type"] = detail.DurationUnitType
			}
			if err := tx.Model(&models.PrescriptionDetail{}).
				Where("prescription_detail_id = ? AND prescription_id = ?", detail.PrescriptionDetailId, detail.PrescriptionId).
				Updates(updateMap).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
		for _, dosage := range detail.DosageInfo {
			if dosage.DoseScheduleId == 0 {
				dosage.PrescriptionDetailId = detail.PrescriptionDetailId
				dosage.CreatedBy = authUserId
				dosage.UpdatedBy = authUserId
				if err := tx.Create(&dosage).Error; err != nil {
					tx.Rollback()
					return err
				}
			} else {
				updateDosage := map[string]interface{}{
					"dose_quantity": dosage.DoseQuantity,
					"unit_value":    dosage.UnitValue,
					"unit_type":     dosage.UnitType,
					"instruction":   dosage.Instruction,
					"is_given":      dosage.IsGiven,
					"updated_by":    authUserId,
				}
				if err := tx.Model(&models.PrescriptionDoseSchedule{}).
					Where("dose_schedule_id = ?", dosage.DoseScheduleId).
					Updates(updateDosage).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	return tx.Commit().Error
}

func (p *PatientRepositoryImpl) GetPrescriptionByPatientId(patientId uint64, limit int, offset int) ([]models.PatientPrescription, int64, error) {
	var prescriptions []models.PatientPrescription
	var totalRecords int64

	query := p.db.
		Where("patient_id = ?", patientId).
		Preload("PrescriptionDetails").Preload("PrescriptionDetails.DosageInfo").Preload("MedicalRecord").Order("prescription_id DESC").
		Limit(limit).
		Offset(offset).
		Find(&prescriptions).
		Count(&totalRecords)

	if query.Error != nil {
		return nil, 0, query.Error
	}

	return prescriptions, totalRecords, nil
}

func (ps *PatientRepositoryImpl) GetPrescriptionDetailByPatientId(patientId uint64, limit int, offset int) ([]models.PrescriptionDetail, int64, error) {
	var prescriptions []models.PatientPrescription
	var total int64
	if err := ps.db.Model(&models.PatientPrescription{}).Where("patient_id = ?", patientId).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []models.PrescriptionDetail{}, 0, nil
	}

	err := ps.db.Preload("PrescriptionDetails.DosageInfo").Preload("MedicalRecord").Where("patient_id = ?", patientId).Order("prescription_id DESC").
		Limit(limit).Offset(offset).
		Find(&prescriptions).Error
	if err != nil {
		return nil, 0, err
	}

	var result []models.PrescriptionDetail
	for _, prescription := range prescriptions {
		for _, detail := range prescription.PrescriptionDetails {
			d := detail
			d.PrescriptionAttachment = prescription.MedicalRecord
			result = append(result, d)
		}
	}

	return result, total, nil
}

func (pr *PatientRepositoryImpl) GetSinglePrescription(prescriptionId uint64, patientId uint64) (models.PatientPrescription, error) {
	var prescription models.PatientPrescription

	err := pr.db.
		Preload("PrescriptionDetails").Preload("PrescriptionDetails.DosageInfo").Preload("MedicalRecord").
		Where("prescription_id = ? AND patient_id = ?", prescriptionId, patientId).
		First(&prescription).Error

	if err != nil {
		return models.PatientPrescription{}, err
	}

	return prescription, nil
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
		MiddleName:  user.MiddleName,
		LastName:    user.LastName,
		DateOfBirth: user.DateOfBirth,
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
	err = p.db.Model(&user).Select("*").Updates(patientData).Error
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

func (p *PatientRepositoryImpl) GetPatientDiagnosticResultValue(patientId uint64, patientDiagnosticReportId uint64) ([]models.PatientDiagnosticReport, map[uint64]uint64, error) {

	_, uniqueReportIds, err := p.GetPatientDiagnosticReportIds(patientId, patientDiagnosticReportId)
	if err != nil {
		log.Printf("Failed to get patient diagnostic report and lab: %v", err)
	}
	reportsWithDetails, err := p.GetPatientDiagnosticTestResult(patientId, uniqueReportIds)
	if err != nil {
		log.Printf("Failed to get patient diagnostic tests: %v", err)
	}
	recordIds, err := p.GetPatientMedicalRecordId(uniqueReportIds)
	if err != nil {
		log.Printf("Failed to get patient diagnostic tests: %v", err)
	}

	return reportsWithDetails, recordIds, nil
}

func (p *PatientRepositoryImpl) GetPatientMedicalRecordId(uniqueReportIds []uint64) (map[uint64]uint64, error) {
	var attachments []models.PatientReportAttachment
	recordMap := make(map[uint64]uint64)

	err := p.db.
		Model(&models.PatientReportAttachment{}).
		Where("patient_diagnostic_report_id IN ?", uniqueReportIds).
		Find(&attachments).Error

	if err != nil {
		return nil, err
	}

	for _, attach := range attachments {
		recordMap[attach.PatientDiagnosticReportId] = attach.RecordId
	}

	return recordMap, nil
}

func (p *PatientRepositoryImpl) RestructurePatientDiagnosticReport(reports []models.PatientDiagnosticReport, medicalRecordInfo []models.TblMedicalRecord, recordIdsMap map[uint64]uint64) []map[string]interface{} {
	restructuredResponse := make([]map[string]interface{}, len(reports))
	for i, report := range reports {
		restructured := map[string]interface{}{
			"patient_diagnostic_report_id": strconv.FormatUint(report.PatientDiagnosticReportId, 10),
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
		recordId, ok := recordIdsMap[report.PatientDiagnosticReportId]
		var matchedMedicalRecord models.TblMedicalRecord

		if ok {
			for _, rec := range medicalRecordInfo {
				if rec.RecordId == recordId {
					matchedMedicalRecord = rec
					break
				}
			}
		}

		if lab, ok := restructured["diagnostic_lab"].(models.DiagnosticLab); ok {
			lab.PatientDiagnosticTests = report.PatientDiagnosticTests
			lab.PatientReportAttachments.MedicalRecord = matchedMedicalRecord
			restructured["diagnostic_lab"] = lab
		} else if labs, ok := restructured["diagnostic_lab"].([]models.DiagnosticLab); ok && len(labs) > 0 {
			labs[0].PatientDiagnosticTests = report.PatientDiagnosticTests
			lab.PatientReportAttachments.MedicalRecord = matchedMedicalRecord
			restructured["diagnostic_lab"] = labs[0]
		}
		restructuredResponse[i] = restructured
	}
	return restructuredResponse
}

func (p *PatientRepositoryImpl) GetPatientDiagnosticReportIds(patientId uint64, reportId uint64) (map[uint64]models.PatientDiagnosticReport, []uint64, error) {
	var reports []models.PatientDiagnosticReport
	query := p.db.Debug().Joins("DiagnosticLabs").Where("tbl_patient_diagnostic_report.patient_id = ?", patientId).Where("tbl_patient_diagnostic_report.is_deleted = ?", 0)

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

func (p *PatientRepositoryImpl) GetPatientDiagnosticTestResult(patientId uint64, reportIds []uint64) ([]models.PatientDiagnosticReport, error) {
	var patientReport []models.PatientDiagnosticReport
	result := p.db.Debug().Model(&models.PatientDiagnosticReport{}).
		Preload("DiagnosticLabs").
		Preload("DiagnosticLabs.PatientReportAttachments", "patient_diagnostic_report_id IN ?", reportIds).
		// Preload("DiagnosticLabs.PatientReportAttachments.MedicalRecord").
		Preload("PatientDiagnosticTests.DiagnosticTest").
		Preload("PatientDiagnosticTests.DiagnosticTest.Components").
		Preload("PatientDiagnosticTests.DiagnosticTest.Components.TestResultValue", "patient_id = ? AND patient_diagnostic_report_id IN (?)", patientId, reportIds).
		Preload("PatientDiagnosticTests.DiagnosticTest.Components.ReferenceRange").
		Where("patient_diagnostic_report_id IN (?) AND is_deleted = ? ", reportIds, 0).Order("patient_diagnostic_report_id DESC").
		Find(&patientReport)

	if result.Error != nil {
		log.Printf("GORM error fetching patient diagnostic tests: %v", result.Error)
		return nil, fmt.Errorf("error fetching patient diagnostic tests: %w", result.Error)
	}
	return patientReport, nil
}

func (p *PatientRepositoryImpl) AddPatientRelative(relative *models.PatientRelative) error {
	return p.db.Create(relative).Error
}

func (ps *PatientRepositoryImpl) AssignPrimaryCaregiver(patientId uint64, relativeId uint64, mappingType string) error {
	tx := ps.db.Begin()
	rollbackErr := func(err error) error {
		tx.Rollback()
		return err
	}
	var relation models.SystemUserRoleMapping
	if err := tx.Where("patient_id = ? AND user_id = ?", patientId, relativeId).
		First(&relation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return rollbackErr(fmt.Errorf("relative does not exist. please add your family member first"))
		}
		return rollbackErr(err)
	}
	if relation.MappingType == mappingType {
		var role string
		if mappingType == "R" {
			role = "relative"
		} else {
			role = "primary-caregiver"
		}
		return rollbackErr(fmt.Errorf("user is already assigned as %s", role))
	}
	if mappingType == "PCG" {
		var pcgCount int64
		if err := tx.Model(&models.SystemUserRoleMapping{}).
			Where("patient_id = ? AND mapping_type = ?", patientId, "PCG").
			Count(&pcgCount).Error; err != nil {
			return rollbackErr(err)
		}
		if pcgCount >= 2 {
			return rollbackErr(fmt.Errorf("maximum 2 primary caregivers allowed"))
		}
	}
	if err := tx.Model(&models.SystemUserRoleMapping{}).Where("patient_id = ? AND user_id = ?", patientId, relativeId).Update("mapping_type", mappingType).Error; err != nil {
		return rollbackErr(err)
	}

	return tx.Commit().Error
}

func (pr *PatientRepositoryImpl) SetCaregiverMappingDeletedStatus(patientId, caregiverId uint64, isDeleted int) error {
	result := pr.db.Model(&models.SystemUserRoleMapping{}).
		Where("patient_id = ? AND user_id = ? AND mapping_type IN ?", patientId, caregiverId, []string{"C"}).
		Update("is_deleted", isDeleted)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no caregiver mapping found to update")
	}

	return nil
}

func (p *PatientRepositoryImpl) GetPatientRelative(patientId string) ([]models.PatientRelative, error) {
	var relatives []models.PatientRelative
	err := p.db.Where("patient_id = ?", patientId).Find(&relatives).Error
	return relatives, err
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
		if relativeId == *r.RelationId {
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
		relationMap[*r.RelationId] = r.RelationShip
	}

	mappingTypeMap := make(map[uint64]string)
	for _, r := range userRelations {
		mappingTypeMap[r.UserId] = r.MappingType
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
			if mapping_name, ok := mappingTypeMap[uid]; ok {
				relativeInfo[i].MappingType = mapping_name
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
		Table("tbl_system_user_ as su").
		Select(`su.user_id AS relative_id, su.first_name, su.last_name, su.gender_id,gm.gender_code AS gender, su.date_of_birth, su.mobile_no AS mobile_no, su.email, su.created_at, su.updated_at,
				(
					SELECT MAX(format_datetime(report_date))
					FROM tbl_patient_diagnostic_report AS dr
					WHERE dr.patient_id = su.user_id
				) AS latest_diganotisic`).
		Joins("LEFT JOIN tbl_gender_master gm ON gm.gender_id = su.gender_id").
		Where("su.user_id IN ?", userIds).
		Scan(&relatives).Error
	return relatives, err
}

func (p *PatientRepositoryImpl) FetchUserIdByPatientId(patientId *uint64, mappingType []string, isSelf bool, isDeleted int) ([]models.UserRelation, error) {
	var userRelations []models.UserRelation

	db := p.db.Table("tbl_system_user_role_mapping")
	if patientId != nil {
		db = db.Where("patient_id = ?", *patientId)
	}
	db = db.Where("mapping_type IN (?) AND is_self = ? AND is_deleted = ? ", mappingType, isSelf, isDeleted)
	err := db.Select("user_id,patient_id,relation_id,mapping_type").Scan(&userRelations).Error
	if err != nil {
		return nil, err
	}
	return userRelations, nil
}

func (p *PatientRepositoryImpl) FetchPatientIdByUserId(userId *uint64, mappingType []string, isSelf bool, isDeleted int) ([]models.UserRelation, error) {
	var userRelations []models.UserRelation

	db := p.db.Table("tbl_system_user_role_mapping")
	if userId != nil {
		db = db.Where("user_id = ?", *userId)
	}
	db = db.Where("mapping_type IN (?) AND is_self = ? AND is_deleted = ? ", mappingType, isSelf, isDeleted)
	err := db.Select("user_id,patient_id,relation_id,mapping_type").Scan(&userRelations).Error
	if err != nil {
		return nil, err
	}
	return userRelations, nil
}

func (p *PatientRepositoryImpl) GetCaregiverList(caregiverUserIds []uint64) ([]models.Caregiver, error) {

	var caregivers []models.Caregiver

	if len(caregiverUserIds) == 0 {
		return caregivers, nil
	}

	err := p.db.
		Table("tbl_system_user_ AS su").
		Select(`su.user_id AS caregiver_id, 
		        su.first_name, 
				su.middle_name,
		        su.last_name, 
		        su.gender_id, 
				gm.gender_code AS gender,
		        su.date_of_birth, 
		        su.mobile_no AS mobile_no, 
		        su.email, 
				su.address,
		        su.created_at, 
		        su.updated_at`).
		Joins("LEFT JOIN tbl_gender_master gm ON gm.gender_id = su.gender_id").
		Where("su.user_id IN ?", caregiverUserIds).
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
		Table("tbl_system_user_ su").
		Select(`su.user_id AS patient_id,
				su.first_name,
				su.last_name,
				su.date_of_birth,
				gm.gender_code AS gender,
				su.mobile_no,
				su.address,
				su.emergency_contact,
				su.abha_number,
				su.blood_group,
				su.nationality,
				su.citizenship_status,
				su.passport_number,
				su.country_of_residence,
				su.is_indian_origin,
				su.email,
				su.created_at,
				su.updated_at`).
		Joins("LEFT JOIN tbl_gender_master gm ON gm.gender_id = su.gender_id").
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
	gender, err := p.GetGenderById(user.GenderId)
	if err != nil {
		log.Printf("Gender not found for ID %v: %v", user.GenderId, err)
	}
	user.Gender = gender.GenderCode
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

func (p *PatientRepositoryImpl) IsUserBasicProfileComplete(user_id uint64) (bool, error) {
	var user models.SystemUser_
	isComplete := false
	err := p.db.Select("first_name", "last_name", "mobile_no", "email", "gender_id", "date_of_birth").
		Where("user_id = ?", user_id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	isComplete = user.GenderId != 0 && user.DateOfBirth != nil && !user.DateOfBirth.IsZero() && user.MobileNo != "" && user.Email != ""
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
	var profile models.TblPatientHealthProfile
	err := p.db.Table("tbl_patient_health_profile").
		Select("height_cm", "weight_kg", "blood_type", "smoking_status", "alcohol_consumption").Where("patient_id = ?", user_id).First(&profile).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	isComplete := profile.HeightCM > 0 && profile.WeightKG > 0 && profile.BloodType != "" && profile.SmokingStatus != "" && profile.AlcoholConsumption != ""
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
	selectFields := `
		pdr.patient_diagnostic_report_id,
		pdr.patient_id,
		format_datetime(pdr.collected_date) AS collected_date,
		format_datetime(pdr.report_date) AS report_date,
		pdr.report_status,
		pdt.test_note,
		format_datetime(pdt.test_date) AS test_date,
		pdtrv.diagnostic_test_id,
		pdtrv.diagnostic_test_component_id,
		tdpdtcm.test_component_name,
		pdtrv.result_value,
		dtrr.normal_min,
		dtrr.normal_max,
		dtrr.units,
		pdtrv.result_status,
		format_datetime(pdtrv.result_date) AS result_date,
		pdtrv.result_comment,
		dc.is_pinned `

	query := fmt.Sprintf(`
		SELECT %s
		FROM tbl_patient_diagnostic_report pdr
		INNER JOIN tbl_patient_diagnostic_test pdt 
			ON pdr.patient_diagnostic_report_id = pdt.patient_diagnostic_report_id
		INNER JOIN tbl_patient_diagnostic_test_result_value pdtrv 
			ON pdt.diagnostic_test_id = pdtrv.diagnostic_test_id 
			AND pdt.patient_diagnostic_report_id = pdtrv.patient_diagnostic_report_id
		LEFT JOIN tbl_diagnostic_test_reference_range dtrr 
			ON pdtrv.diagnostic_test_component_id = dtrr.diagnostic_test_component_id
		LEFT JOIN tbl_disease_profile_diagnostic_test_component_master tdpdtcm 
			ON tdpdtcm.diagnostic_test_component_id = pdtrv.diagnostic_test_component_id
		LEFT JOIN tbl_patient_test_component_display_config dc 
			ON pdtrv.diagnostic_test_component_id = dc.diagnostic_test_component_id 
			AND pdtrv.patient_id = dc.patient_id`, selectFields)

	query += ` WHERE pdr.patient_id = ?  AND pdr.is_deleted = 0 `
	args := []interface{}{input.PatientId}

	if input.DiagnosticTestComponentId != nil {
		query += " AND pdtrv.diagnostic_test_component_id = ? "
		args = append(args, *input.DiagnosticTestComponentId)
		query += " ORDER BY pdtrv.result_date DESC"
	}

	if input.PatientDiagnosticReportId != nil {
		query += " AND pdtrv.patient_diagnostic_report_id = ?"
		args = append(args, *input.PatientDiagnosticReportId)
	}

	if input.ReportDateStart != nil && input.ReportDateEnd != nil {
		query += " AND pdr.report_date BETWEEN ? AND ?"
		args = append(args, *input.ReportDateStart, *input.ReportDateEnd)
	}

	if input.ResultDateStart != nil && input.ResultDateEnd != nil {
		query += " AND pdtrv.result_date BETWEEN ? AND ?"
		args = append(args, *input.ResultDateStart, *input.ResultDateEnd)
	}

	if input.IsPinned != nil {
		query += " AND dc.is_pinned = ?"
		args = append(args, *input.IsPinned)
		query += " ORDER BY dc.created_at DESC"
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

func (p *PatientRepositoryImpl) ParseDiagnosticTrendData(rawData []map[string]interface{}) ([]map[string]interface{}, error) {
	grouped := make(map[interface{}]map[string]interface{})

	for _, row := range rawData {
		componentID := row["diagnostic_test_component_id"]

		if _, exists := grouped[componentID]; !exists {
			grouped[componentID] = map[string]interface{}{
				"diagnostic_test_component_id": componentID,
				"test_component_name":          row["test_component_name"],
				"ref_unit":                     row["units"],
				"normal_min":                   row["normal_min"],
				"normal_max":                   row["normal_max"],
				"is_pinned":                    row["is_pinned"],
				"diagnostic_test_id":           row["diagnostic_test_id"],
				"patient_id":                   row["patient_id"],
				"trend_history":                []map[string]interface{}{},
			}
		}

		historyEntry := map[string]interface{}{
			"patient_diagnostic_report_id": row["patient_diagnostic_report_id"],
			"report_date":                  row["report_date"],
			"result_date":                  row["result_date"],
			"collected_date":               row["collected_date"],
			"test_date":                    row["test_date"],
			"report_status":                row["report_status"],
			"result_status":                row["result_status"],
			"result_value":                 row["result_value"],
			"result_comment":               row["result_comment"],
			"test_note":                    row["test_note"],
		}

		group := grouped[componentID]
		group["trend_history"] = append(group["trend_history"].([]map[string]interface{}), historyEntry)
	}

	result := make([]map[string]interface{}, 0, len(grouped))
	for _, group := range grouped {
		result = append(result, group)
	}

	return result, nil
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
		relationMap[*r.RelationId] = r.RelationShip
	}

	var orderedRelations []models.PatientRelation
	for _, id := range ids {
		if relationName, ok := relationMap[id]; ok {
			orderedRelations = append(orderedRelations, models.PatientRelation{
				RelationId:   &id,
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

func (p *PatientRepositoryImpl) CheckPatientHealthProfileExist(tx *gorm.DB, patientId uint64) (bool, error) {
	var count int64
	err := tx.Model(&models.TblPatientHealthProfile{}).
		Where("patient_id = ?", patientId).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *PatientRepositoryImpl) UpdatePatientHealthDetail(req *models.TblPatientHealthProfile) error {
	return p.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.TblPatientHealthProfile{}).
			Where("patient_id = ?", req.PatientId).
			Count(&count).Error; err != nil {
			return err
		}

		data := map[string]interface{}{
			"height_cm":               req.HeightCM,
			"weight_kg":               req.WeightKG,
			"blood_type":              req.BloodType,
			"smoking_status":          req.SmokingStatus,
			"alcohol_consumption":     req.AlcoholConsumption,
			"physical_activity_level": req.PhysicalActivityLevel,
			"dietary_preferences":     req.DietaryPreferences,
			"existing_conditions":     req.ExistingConditions,
			"family_medical_history":  req.FamilyMedicalHistory,
			"menstrual_history":       req.MenstrualHistory,
			"notes":                   req.Notes,
			"updated_by":              req.UpdatedBy,
		}

		if count > 0 {
			if err := tx.Model(&models.TblPatientHealthProfile{}).
				Where("patient_id = ?", req.PatientId).
				Updates(data).Error; err != nil {
				return err
			}
			return nil
		}
		return gorm.ErrRecordNotFound
	})
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

func (p *PatientRepositoryImpl) FetchPatientDiagnosticReports(patientId uint64, filter models.DiagnosticReportFilter) ([]models.DiagnosticReportResponse, error) {
	var results []models.DiagnosticReportResponse

	query := `
		SELECT
			pdr.patient_diagnostic_report_id,
			pdr.patient_id,
			pdr.collected_date,
			pdr.report_date,
			pdr.report_status,
			pdr.report_name,
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
			pdtrv.result_comment,
			dl.diagnostic_lab_id,
			dl.lab_name,
			pdtrv.udf1 as qualifier
		FROM
			tbl_patient_diagnostic_test_result_value pdtrv
		LEFT JOIN tbl_patient_diagnostic_test pdt
			ON pdtrv.patient_diagnostic_report_id = pdt.patient_diagnostic_report_id
			AND pdtrv.diagnostic_test_id = pdt.diagnostic_test_id
		LEFT JOIN tbl_patient_diagnostic_report pdr
			ON pdtrv.patient_diagnostic_report_id = pdr.patient_diagnostic_report_id
		LEFT JOIN tbl_diagnostic_test_reference_range dtrr
			ON pdtrv.diagnostic_test_component_id = dtrr.diagnostic_test_component_id
		LEFT JOIN tbl_disease_profile_diagnostic_test_component_master tdpdtcm
			ON pdtrv.diagnostic_test_component_id = tdpdtcm.diagnostic_test_component_id
		LEFT JOIN tbl_diagnostic_lab dl
			ON pdr.diagnostic_lab_id = dl.diagnostic_lab_id
		WHERE
			pdtrv.patient_id = ?
	`

	var args []interface{}
	args = append(args, patientId)

	if filter.ReportID != nil {
		query += " AND pdtrv.patient_diagnostic_report_id = ?"
		args = append(args, *filter.ReportID)
	}

	if filter.TestName != nil {
		query += " AND pdt.test_note ILIKE ?"
		args = append(args, "%"+*filter.TestName+"%")
	}

	if filter.Qualifier != nil {
		query += " AND pdtrv.udf1 = ?"
		args = append(args, *filter.Qualifier)
	}

	if filter.TestComponentName != nil {
		query += " AND tdpdtcm.test_component_name ILIKE ?"
		args = append(args, "%"+*filter.TestComponentName+"%")
	}

	if filter.DiagnosticLabID != nil {
		query += " AND dl.diagnostic_lab_id = ?"
		args = append(args, *filter.DiagnosticLabID)
	}

	if filter.ReportName != nil {
		query += " AND pdr.report_name = ?"
		args = append(args, *filter.ReportName)
	}

	if filter.ReportStatus != nil {
		query += " AND pdr.report_status = ?"
		args = append(args, *filter.ReportStatus)
	}

	if filter.ResultDateFrom != nil && filter.ResultDateTo != nil {
		query += " AND pdtrv.result_date BETWEEN ? AND ?"
		args = append(args, *filter.ResultDateFrom, *filter.ResultDateTo)
	}

	fmt.Println("Executing Query:", query)
	fmt.Println("With Args:", args)

	if err := p.db.Debug().Raw(query, args...).Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

func (p *PatientRepositoryImpl) GetPatientDiagnosticReportResult(patientId uint64, filter models.DiagnosticReportFilter, limit int, offset int) ([]models.ReportRow, int64, error) {
	var results []models.ReportRow

	totalReports, err := p.CountPatientDiagnosticReports(patientId, filter)
	if err != nil {
		return nil, 0, err
	}

	query := `WITH paginated_reports AS (SELECT DISTINCT pdr.patient_diagnostic_report_id FROM tbl_patient_diagnostic_report pdr WHERE pdr.patient_id = ? AND pdr.is_deleted = 0 `
	args := []interface{}{patientId}

	if filter.ReportStatus != nil {
		query += " AND pdr.report_status = ?"
		args = append(args, *filter.ReportStatus)
	}
	if filter.ReportName != nil {
		query += " AND pdr.report_name = ?"
		args = append(args, *filter.ReportName)
	}

	if filter.ReportID != nil {
		query += " AND pdr.patient_diagnostic_report_id = ?"
		args = append(args, *filter.ReportID)
	}

	query += ` ORDER BY pdr.patient_diagnostic_report_id DESC LIMIT ? OFFSET ? ) `
	args = append(args, limit, offset)

	query += `
		SELECT 
			pdr.patient_diagnostic_report_id,
			pdr.patient_id,
			format_datetime(pdr.collected_date) AS collected_date,
			format_datetime(pdr.report_date) AS report_date,
			pdr.report_status,
			pdr.report_name,
			pdt.test_note,
			pdt.test_date,
			pdtrv.diagnostic_test_id,
			pdtm.test_name,
			pdtrv.diagnostic_test_component_id,
			COALESCE(orig_comp.test_component_name, tdpdtcm.test_component_name) AS test_component_name,
			COALESCE(orig_comp.units, tdpdtcm.units) AS component_unit,
			pdtrv.result_value,
			dtrr.normal_min,
			dtrr.normal_max,
			dtrr.units AS ref_units,
			pdtrv.result_status,
			format_datetime(pdtrv.result_date) AS result_date,
			pdtrv.result_comment,
			dl.diagnostic_lab_id,
			dl.lab_name,
			dc.is_pinned,
			pdtrv.udf1 AS qualifier
		FROM tbl_patient_diagnostic_test_result_value pdtrv
		LEFT JOIN tbl_patient_diagnostic_test pdt 
			ON pdtrv.patient_diagnostic_report_id = pdt.patient_diagnostic_report_id 
			AND pdtrv.diagnostic_test_id = pdt.diagnostic_test_id
		LEFT JOIN tbl_patient_diagnostic_report pdr 
			ON pdtrv.patient_diagnostic_report_id = pdr.patient_diagnostic_report_id
		LEFT JOIN tbl_diagnostic_test_reference_range dtrr 
			ON pdtrv.diagnostic_test_component_id = dtrr.diagnostic_test_component_id
		LEFT JOIN tbl_disease_profile_diagnostic_test_component_master tdpdtcm 
			ON pdtrv.diagnostic_test_component_id = tdpdtcm.diagnostic_test_component_id
		LEFT JOIN tbl_diagnostic_test_component_alias_mapping tcam 
			ON tcam.alias_test_component_id = pdtrv.diagnostic_test_component_id
		LEFT JOIN tbl_disease_profile_diagnostic_test_component_master orig_comp
   			ON orig_comp.diagnostic_test_component_id = tcam.diagnostic_test_component_id
		LEFT JOIN tbl_disease_profile_diagnostic_test_master pdtm 
			ON pdtrv.diagnostic_test_id = pdtm.diagnostic_test_id
		LEFT JOIN tbl_patient_test_component_display_config dc 
			ON pdtrv.diagnostic_test_component_id = dc.diagnostic_test_component_id 
			AND pdtrv.patient_id = dc.patient_id
		LEFT JOIN tbl_diagnostic_lab dl 
			ON pdr.diagnostic_lab_id = dl.diagnostic_lab_id
		WHERE pdtrv.patient_diagnostic_report_id IN (
			SELECT patient_diagnostic_report_id FROM paginated_reports
		)
	`

	if filter.TestName != nil {
		query += " AND pdtm.test_name ILIKE ? "
		args = append(args, "%"+*filter.TestName+"%")
	}
	if filter.TestNote != nil {
		query += " AND pdt.test_note ILIKE ? "
		args = append(args, "%"+*filter.TestNote+"%")
	}
	if filter.Qualifier != nil {
		query += " AND pdtrv.udf1 = ? "
		args = append(args, *filter.Qualifier)
	}
	if filter.TestComponentName != nil {
		query += " AND tdpdtcm.test_component_name ILIKE ? "
		args = append(args, "%"+*filter.TestComponentName+"%")
	}
	if filter.DiagnosticLabID != nil {
		query += " AND dl.diagnostic_lab_id = ? "
		args = append(args, *filter.DiagnosticLabID)
	}
	if filter.ResultDateFrom != nil && filter.ResultDateTo != nil {
		query += " AND DATE(pdtrv.result_date) BETWEEN ? AND ? "
		args = append(args, *filter.ResultDateFrom, *filter.ResultDateTo)
	}
	if filter.ReportDate != nil {
		query += " AND DATE(pdr.report_date) = ?"
		args = append(args, *filter.ReportDate)
	}
	query += p.BuildOrderByClause(filter.OrderBy, filter.OrderDir)

	fmt.Println("Executing Query:", query)
	fmt.Println("With Args:", args)
	if err := p.db.Raw(query, args...).Scan(&results).Error; err != nil {
		return nil, 0, fmt.Errorf("data query failed: %w", err)
	}

	return results, totalReports, nil
}

func (p *PatientRepositoryImpl) BuildOrderByClause(orderBy *string, orderDir *string) string {
	columnMap := map[string]string{
		"result_date":                  "pdtrv.result_date",
		"report_date":                  "pdr.report_date",
		"patient_diagnostic_report_id": "pdr.patient_diagnostic_report_id",
		"test_name":                    "pdtm.test_name",
		"lab_name":                     "dl.lab_name",
	}

	orderColumn := "pdr.report_date"
	direction := "DESC"

	if orderBy != nil && *orderBy != "" {
		if mappedCol, ok := columnMap[*orderBy]; ok {
			orderColumn = mappedCol
		} else {
			orderColumn = *orderBy
		}
	}

	if orderDir != nil && (*orderDir == "ASC" || *orderDir == "DESC") {
		direction = *orderDir
	}

	return fmt.Sprintf(" ORDER BY %s %s", orderColumn, direction)
}

func (p *PatientRepositoryImpl) CountPatientDiagnosticReports(patientId uint64, filter models.DiagnosticReportFilter) (int64, error) {
	var totalReports int64

	query := `
		SELECT COUNT(DISTINCT pdr.patient_diagnostic_report_id)
		FROM tbl_patient_diagnostic_report pdr
		WHERE pdr.patient_id = ? AND pdr.is_deleted = 0
	`
	args := []interface{}{patientId}

	if filter.ReportID != nil {
		query += " AND pdr.patient_diagnostic_report_id = ?"
		args = append(args, *filter.ReportID)
	}

	if filter.ReportStatus != nil {
		query += " AND pdr.report_status = ?"
		args = append(args, *filter.ReportStatus)
	}
	if filter.ReportName != nil {
		query += " AND pdr.report_name = ?"
		args = append(args, *filter.ReportName)
	}

	err := p.db.Raw(query, args...).Scan(&totalReports).Error
	if err != nil {
		return 0, fmt.Errorf("count query failed: %w", err)
	}

	return totalReports, nil
}

func (p *PatientRepositoryImpl) ProcessReportGridData(rows []models.ReportRow) map[string]interface{} {
	if len(rows) == 0 {
		return map[string]interface{}{}
	}
	componentMap := make(map[models.ComponentKey][]models.CellData)

	// Dates
	dateSet := make(map[string]struct{})

	for _, row := range rows {
		if row.TestComponentName == "" {
			continue
		}

		rangeStr := fmt.Sprintf("%v - %v", row.NormalMin, row.NormalMax)
		valueStr := fmt.Sprintf("%v", row.ResultValue)
		colorClass, colour := utils.GetRefRangeAndColorCode(valueStr, row.NormalMin, row.NormalMax)
		if row.ResultValue == "0" {
			valueStr = row.ResultStatus
		}
		key := models.ComponentKey{
			ComponentID: row.DiagnosticTestComponentID,
			Name:        row.TestComponentName,
			Units:       row.ComponentUnit,
			RefRange:    rangeStr,
			IsPinned:    row.IsPinned,
		}

		cell := models.CellData{
			Value:        valueStr,
			ResultStatus: row.ResultStatus,
			ColourClass:  colorClass,
			Colour:       colour,
			ReportID:     row.PatientDiagnosticReportID,
			ResultDate:   row.ResultDate,
			Qualifier:    row.Qualifier,
			ReportName:   row.ReportName,
			IsPinned:     row.IsPinned,
		}

		componentMap[key] = append(componentMap[key], cell)
		dateSet[row.ResultDate] = struct{}{}
	}

	var finalRows []map[string]interface{}
	var allDates []string

	for date := range dateSet {
		allDates = append(allDates, date)
	}

	for key, values := range componentMap {
		row := map[string]interface{}{
			"diagnostic_test_component_id": key.ComponentID,
			"test_component_name":          key.Name,
			"ref_unit":                     key.Units,
			"ref_range":                    key.RefRange,
			"report_name":                  key.ReportName,
			"is_pinned":                    key.IsPinned,
			"trend_values":                 values,
		}
		finalRows = append(finalRows, row)
	}

	return map[string]interface{}{
		"columns": []string{"TEST", "UNIT", "REF.RANGE", "DATES"},
		"dates":   allDates,
		"rows":    finalRows,
	}
}

func (p *PatientRepositoryImpl) RestructureDiagnosticReports(flatData []models.DiagnosticReportResponse) []map[string]interface{} {
	reportMap := make(map[uint64]map[string]interface{})
	for _, item := range flatData {
		reportID := item.PatientDiagnosticReportID
		if _, exists := reportMap[reportID]; !exists {
			reportMap[reportID] = map[string]interface{}{
				"patient_diagnostic_report_id": reportID,
				"patient_id":                   item.PatientID,
				"collected_date":               item.CollectedDate,
				"report_date":                  item.ReportDate,
				"report_status":                item.ReportStatus,
				"report_name":                  item.ReportName,
				"comments":                     item.ResultComment,
				"collected_at":                 item.CollectedDate,
				"diagnostic_lab": map[string]interface{}{
					"diagnostic_lab_id":       item.DiagnosticLabID,
					"lab_name":                item.LabName,
					"patient_diagnostic_test": []map[string]interface{}{},
				},
			}
		}

		report := reportMap[reportID]
		diagnosticLab := report["diagnostic_lab"].(map[string]interface{})

		testComponent := map[string]interface{}{
			"diagnostic_test_component_id": item.DiagnosticTestComponentID,
			"test_component_name":          item.TestComponentName,
			"units":                        item.Units,
			"test_result_value": []map[string]interface{}{
				{
					"diagnostic_test_id":           item.DiagnosticTestID,
					"diagnostic_test_component_id": item.DiagnosticTestComponentID,
					"result_value":                 item.ResultValue,
					"result_status":                item.ResultStatus,
					"result_date":                  item.ResultDate,
					"result_comment":               item.ResultComment,
					"qualifier":                    item.Qualifier,
				},
			},
			"test_referance_range": []map[string]interface{}{
				{
					"normal_min": item.NormalMin,
					"normal_max": item.NormalMax,
					"units":      item.Units,
				},
			},
		}

		diagnosticTest := map[string]interface{}{
			"diagnostic_test_id": item.DiagnosticTestID,
			"test_components":    []map[string]interface{}{testComponent},
		}

		patientDiagnosticTest := map[string]interface{}{
			"test_note":       item.TestNote,
			"test_date":       item.TestDate,
			"diagnostic_test": diagnosticTest,
		}

		pdtList := diagnosticLab["patient_diagnostic_test"].([]map[string]interface{})
		diagnosticLab["patient_diagnostic_test"] = append(pdtList, patientDiagnosticTest)
	}
	finalReports := make([]map[string]interface{}, 0, len(reportMap))
	for _, val := range reportMap {
		finalReports = append(finalReports, val)
	}
	return finalReports
}

func (r *PatientRepositoryImpl) GetDiagnosticReportId(patientId uint64) (uint64, error) {
	var reportId uint64
	err := r.db.Table("tbl_patient_diagnostic_report").
		Where("patient_id = ?", patientId).
		Select("MAX(patient_diagnostic_report_id)").
		Scan(&reportId).Error

	if err != nil {
		return 0, err
	}

	return reportId, nil
}

func (pr *PatientRepositoryImpl) GetPatientHealthDetail(userID uint64) (models.TblPatientHealthProfile, error) {
	var profile models.TblPatientHealthProfile
	err := pr.db.Where("patient_id = ?", userID).First(&profile).Error
	return profile, err
}

func (ps *PatientRepositoryImpl) AddTestComponentDisplayConfig(input *models.PatientTestComponentDisplayConfig) error {
	now := time.Now()

	return ps.db.Transaction(func(tx *gorm.DB) error {
		var config models.PatientTestComponentDisplayConfig
		err := tx.Where("patient_id = ? AND diagnostic_test_component_id = ?", input.PatientId, input.DiagnosticTestComponentId).First(&config).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				newConfig := models.PatientTestComponentDisplayConfig{
					PatientId:                 input.PatientId,
					DiagnosticTestComponentId: input.DiagnosticTestComponentId,
					CreatedAt:                 now,
					UpdatedAt:                 now,
					CreatedBy:                 input.CreatedBy,
					UpdatedBy:                 input.UpdatedBy,
				}

				if input.IsPinned != nil {
					newConfig.IsPinned = input.IsPinned
				} else {
					falseVal := false
					newConfig.IsPinned = &falseVal
				}

				if input.DisplayPriority != nil {
					newConfig.DisplayPriority = input.DisplayPriority
				} else {
					zeroVal := 0
					newConfig.DisplayPriority = &zeroVal
				}

				if err := tx.Create(&newConfig).Error; err != nil {
					return err
				}

				return nil
			}
			return err
		}
		if input.IsPinned != nil {
			config.IsPinned = input.IsPinned
		}
		if input.DisplayPriority != nil {
			config.DisplayPriority = input.DisplayPriority
		}
		config.UpdatedAt = now

		if err := tx.Save(&config).Error; err != nil {
			return err
		}

		return nil
	})
}

func (p *PatientRepositoryImpl) GetPinnedComponentCount(patientId uint64) (int64, error) {
	var count int64
	err := p.db.
		Model(&models.PatientTestComponentDisplayConfig{}).
		Where("patient_id = ? AND is_pinned = ?", patientId, true).
		Count(&count).Error
	return count, err
}

func (r *PatientRepositoryImpl) GrantPermission(userID, relativeID uint64, permissionID int64, granted bool) error {
	var mapping models.UserRelativePermissionMapping
	err := r.db.Where("user_id = ? AND relative_id = ? AND permission_id = ?", userID, relativeID, permissionID).First(&mapping).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		mapping = models.UserRelativePermissionMapping{
			UserID: userID, RelativeID: relativeID, PermissionID: permissionID, Granted: granted,
		}
		return r.db.Create(&mapping).Error
	}
	if err != nil {
		return err
	}
	mapping.Granted = granted
	return r.db.Save(&mapping).Error
}

func (r *PatientRepositoryImpl) HasPermission(userID, relativeID uint64, permissionCode string) (bool, error) {
	var permission models.PermissionMaster
	err := r.db.Where("code = ?", permissionCode).First(&permission).Error
	if err != nil {
		return false, err
	}

	var mapping models.UserRelativePermissionMapping
	err = r.db.Where("user_id = ? AND relative_id = ? AND permission_id = ? AND granted = true",
		userID, relativeID, permission.PermissionID).First(&mapping).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (r *PatientRepositoryImpl) ListPermissions(userID, relativeID uint64) ([]models.PermissionResult, error) {
	var results []models.PermissionResult

	err := r.db.Table("tbl_user_relative_permission_mappings").
		Select("tbl_user_relative_permission_mappings.user_id, tbl_user_relative_permission_mappings.relative_id, tbl_permissions_master.code, tbl_user_relative_permission_mappings.granted").
		Joins("JOIN tbl_permissions_master ON tbl_user_relative_permission_mappings.permission_id = tbl_permissions_master.permission_id").
		Where("tbl_user_relative_permission_mappings.user_id = ? AND tbl_user_relative_permission_mappings.relative_id = ?", userID, relativeID).
		Scan(&results).Error

	return results, err
}

func (r *PatientRepositoryImpl) GetPermissionByCode(code string) (*models.PermissionMaster, error) {
	var permission models.PermissionMaster
	err := r.db.Where("code = ?", code).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *PatientRepositoryImpl) CheckPermissionValue(userID, relativeID uint64, permissionID int64) (exists bool, currentValue bool, err error) {
	var mapping models.UserRelativePermissionMapping
	err = r.db.Where("user_id = ? AND relative_id = ? AND permission_id = ?", userID, relativeID, permissionID).
		First(&mapping).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, false, nil
	}
	return true, mapping.Granted, err
}

func (r *PatientRepositoryImpl) UpdatePermissionValue(userID, relativeID uint64, permissionID int64, value bool) error {
	return r.db.Model(&models.UserRelativePermissionMapping{}).
		Where("user_id = ? AND relative_id = ? AND permission_id = ?", userID, relativeID, permissionID).
		Update("granted", value).Error
}

func (ur *PatientRepositoryImpl) GetDistinctMedicinesByPatientID(patientID uint64) ([]models.UserMedicineInfo, error) {
	var results []models.UserMedicineInfo

	query := `
  SELECT prescription_detail_id, prescription_id, medicine_name, prescription_type, duration, duration_unit_type
  FROM (
    SELECT *, ROW_NUMBER() OVER (PARTITION BY medicine_name ORDER BY prescription_id) AS rn
    FROM tbl_prescription_detail
    WHERE prescription_id IN (
      SELECT prescription_id
      FROM tbl_patient_prescription
      WHERE patient_id = ?
    )
  ) t
  WHERE rn = 1;
  `

	err := ur.db.Raw(query, patientID).Scan(&results).Error
	return results, err
}
