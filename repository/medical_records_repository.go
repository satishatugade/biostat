package repository

import (
	"biostat/config"
	"biostat/constant"
	"biostat/models"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TblMedicalRecordRepository interface {
	GetAllMedicalRecord(patientId uint64, limit int, offset int) ([]models.ReportRow, int64, error)
	ProcessMedicalRecordResponse(data []models.ReportRow) []map[string]interface{}
	GetMedicalRecordsByUserID(userID uint64, recordIdsMap map[uint64]uint64) ([]models.TblMedicalRecord, error)
	CreateTblMedicalRecord(tx *gorm.DB, data *models.TblMedicalRecord) (*models.TblMedicalRecord, error)
	CreateMultipleTblMedicalRecords(tx *gorm.DB, data []*models.TblMedicalRecord) error
	UpdateTblMedicalRecord(data *models.TblMedicalRecord) (*models.TblMedicalRecord, error)
	GetMedicalRecordByRecordId(RecordId uint64) (*models.TblMedicalRecord, error)
	DeleteTblMedicalRecord(id int, updatedBy string) error
	IsRecordBelongsToUser(userID uint64, recordID uint64) (bool, error)
	ExistsRecordForUser(userId uint64, source, url string) (bool, error)

	CreateMedicalRecordMappings(tx *gorm.DB, mappings *[]models.TblMedicalRecordUserMapping) error
	UpdateMedicalRecordMappingByRecordId(tx *gorm.DB, RecordId *uint64, mapping map[string]interface{}) error
	GetMedicalRecordMappings(recordID uint64) (*models.TblMedicalRecordUserMapping, error)
	DeleteMecationRecordMappings(id int) error

	DeleteTblMedicalRecordWithMappings(id int, user_id string) error

	GetMedicalRecordsByUser(userID uint64, category, tag string, limit, offset, isDeleted int) ([]models.TblMedicalRecord, int64, map[string]int64, error)
	GetDiagnosticAttachmentByRecordIDs(recordIDs []uint64) ([]models.PatientReportAttachment, error)
	GetDiagnosticReport(reportID uint64, isDeleted int) (*models.PatientDiagnosticReport, error)
	GetDiagnosticTests(reportID uint64) ([]models.PatientDiagnosticTest, error)
	GetTestComponents(reportID uint64, testID uint64) ([]models.PatientDiagnosticTestResultValue, error)
	GetComponentDetails(componentID uint64) (*models.DiagnosticTestComponent, error)
	GetReferenceRanges(componentID uint64) ([]models.DiagnosticTestReferenceRange, error) //separate table
	GetDiagnosticTestMaster(testID uint64) (*models.DiagnosticTest, error)
	MovePatientRecord(patientId, targetPatientId, recordId, reportId uint64) error
	GetReportRecordMapping(userID uint64, category, tag string, isDeleted int) (map[uint64][]uint64, int64, map[string]int64, error)
	GetRecordsByIDs(recordIDs []uint64, isDeleted int) ([]models.TblMedicalRecord, error)
	GetAllReportTag(userId uint64, limit, offset int) ([]models.UserTag, int64, error)
	AddTag(tag *models.UserTag) error
}

type tblMedicalRecordRepositoryImpl struct {
	db *gorm.DB
}

func NewTblMedicalRecordRepository(db *gorm.DB) TblMedicalRecordRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &tblMedicalRecordRepositoryImpl{db: db}
}

func (r *tblMedicalRecordRepositoryImpl) GetMedicalRecordsByUser(userID uint64, category, tag string, limit, offset, isDeleted int) ([]models.TblMedicalRecord, int64, map[string]int64, error) {
	var records []models.TblMedicalRecord
	var total int64
	counts := make(map[string]int64)

	status := config.PropConfig.SystemVaribale.Status
	statuses := strings.Split(status, ",")

	baseQuery := r.db.
		Table("tbl_medical_record AS mr").
		Joins("JOIN tbl_medical_record_user_mapping AS mrum ON mr.record_id = mrum.record_id").
		Where("mrum.user_id = ? AND mr.is_deleted = ? AND mr.status IN (?) ", userID, isDeleted, statuses)

	if tag != "" {
		baseQuery = baseQuery.
			Joins("JOIN tbl_user_tag AS ut ON ut.record_id = mr.record_id").
			Where("ut.tag_name ILIKE ? AND ut.user_id = ?", "%"+tag+"%", userID)
	}

	var categoryCounts []struct {
		RecordCategory string
		Count          int64
	}
	countsQuery := baseQuery.Session(&gorm.Session{})
	if err := countsQuery.Select("mr.record_category, COUNT(*) as count").
		Group("mr.record_category").
		Scan(&categoryCounts).Error; err != nil {
		return nil, 0, nil, err
	}
	for _, c := range categoryCounts {
		counts[c.RecordCategory] = c.Count
	}
	query := baseQuery.Session(&gorm.Session{})
	query = query.
		Joins("LEFT JOIN tbl_patient_report_attachment AS pra ON pra.record_id = mr.record_id").
		Joins("LEFT JOIN tbl_patient_diagnostic_report AS report ON report.patient_diagnostic_report_id = pra.patient_diagnostic_report_id")
	if isDeleted == 1 {
		recordCategory := string(constant.DUPLICATE)
		query = query.Where("mr.record_category = ?", recordCategory)
	}
	if category != "" {
		query = query.Where("mr.record_category = ?", category)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, nil, err
	}

	err := query.
		Offset(offset).
		Limit(limit).
		Order("COALESCE(report.report_date, mr.created_at) DESC").
		Find(&records).Error

	return records, total, counts, err
}

// GET LIST OF DIGITIZED RECORDS out of user records
func (r *tblMedicalRecordRepositoryImpl) GetDiagnosticAttachmentByRecordIDs(recordIDs []uint64) ([]models.PatientReportAttachment, error) {
	var attachments []models.PatientReportAttachment
	err := r.db.
		Where("record_id IN ?", recordIDs).
		Find(&attachments).Error
	return attachments, err
}

// GET BASIC DETAILS OF REPORT
func (r *tblMedicalRecordRepositoryImpl) GetDiagnosticReport(reportID uint64, isDeleted int) (*models.PatientDiagnosticReport, error) {
	var report models.PatientDiagnosticReport
	err := r.db.
		Where("patient_diagnostic_report_id = ? AND is_deleted = ? ", reportID, isDeleted).
		First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// get list of tests in report
func (r *tblMedicalRecordRepositoryImpl) GetDiagnosticTests(reportID uint64) ([]models.PatientDiagnosticTest, error) {
	var tests []models.PatientDiagnosticTest
	err := r.db.
		Where("patient_diagnostic_report_id = ?", reportID).
		Find(&tests).Error
	return tests, err
}

// GET test_components
func (r *tblMedicalRecordRepositoryImpl) GetTestComponents(reportID uint64, testID uint64) ([]models.PatientDiagnosticTestResultValue, error) {
	var results []models.PatientDiagnosticTestResultValue
	err := r.db.
		Where("patient_diagnostic_report_id = ? AND diagnostic_test_id = ?", reportID, testID).
		Find(&results).Error
	return results, err
}

func (r *tblMedicalRecordRepositoryImpl) GetComponentDetails(componentID uint64) (*models.DiagnosticTestComponent, error) {
	var comp models.DiagnosticTestComponent
	err := r.db.
		Where("diagnostic_test_component_id = ?", componentID).
		First(&comp).Error
	if err != nil {
		return nil, err
	}
	return &comp, nil
}

func (r *tblMedicalRecordRepositoryImpl) GetReferenceRanges(componentID uint64) ([]models.DiagnosticTestReferenceRange, error) {
	var ranges []models.DiagnosticTestReferenceRange
	err := r.db.
		Where("diagnostic_test_component_id = ?", componentID).
		Find(&ranges).Error
	return ranges, err
}

func (r *tblMedicalRecordRepositoryImpl) GetDiagnosticTestMaster(testID uint64) (*models.DiagnosticTest, error) {
	var test models.DiagnosticTest
	err := r.db.
		Where("diagnostic_test_id = ?", testID).
		First(&test).Error
	if err != nil {
		return nil, err
	}
	return &test, nil
}

func (r *tblMedicalRecordRepositoryImpl) GetMedicalRecordsByUserID(userID uint64, recordIdMap map[uint64]uint64) ([]models.TblMedicalRecord, error) {
	var records []models.TblMedicalRecord

	query := r.db.Table("tbl_medical_record").
		Select("tbl_medical_record.*").
		Joins("INNER JOIN tbl_medical_record_user_mapping ON tbl_medical_record.record_id = tbl_medical_record_user_mapping.record_id").
		Where("tbl_medical_record_user_mapping.user_id = ? and is_deleted=0", userID)

	if len(recordIdMap) > 0 {
		var recordIds []uint64
		for _, id := range recordIdMap {
			recordIds = append(recordIds, id)
		}
		query = query.Where("tbl_medical_record.record_id IN ?", recordIds)
	}
	err := query.Order("tbl_medical_record.updated_at DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

func (r *tblMedicalRecordRepositoryImpl) GetAllMedicalRecord(patientId uint64, limit, offset int) ([]models.ReportRow, int64, error) {
	recordIDs, err := r.getPaginatedRecordIDs(patientId, limit, offset)
	if err != nil || len(recordIDs) == 0 {
		return []models.ReportRow{}, 0, err
	}

	records, err := r.getMedicalRecordDetails(recordIDs)
	if err != nil {
		return nil, 0, err
	}

	total, err := r.getTotalRecordCount(patientId)
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *tblMedicalRecordRepositoryImpl) getPaginatedRecordIDs(patientId uint64, limit, offset int) ([]uint64, error) {
	var ids []uint64
	query := `
		SELECT mr.record_id
		FROM tbl_medical_record mr
		INNER JOIN tbl_medical_record_user_mapping rm ON rm.record_id = mr.record_id
		WHERE mr.is_deleted = 0 AND rm.user_id = ?
		ORDER BY mr.record_id DESC
		LIMIT ? OFFSET ?
	`
	err := r.db.Raw(query, patientId, limit, offset).Scan(&ids).Error
	return ids, err
}

func (r *tblMedicalRecordRepositoryImpl) getMedicalRecordDetails(recordIDs []uint64) ([]models.ReportRow, error) {
	var results []models.ReportRow
	if len(recordIDs) == 0 {
		return results, nil
	}

	query := `
	SELECT 
		mr.record_id, mr.record_name, mr.record_size, mr.file_type, mr.status,
		mr.upload_source, mr.source_account, mr.record_category, mr.record_url, mr.digitize_flag,
		pr.patient_diagnostic_report_id, rm.user_id AS patient_id, pr.report_name,
		format_datetime(pr.collected_date) AS collected_date,
		format_datetime(pr.report_date) AS report_date,
		pr.report_status, pr.observation, pr.comments, pr.review_by,
		format_datetime(pr.review_date) AS review_date,
		dl.diagnostic_lab_id, dl.lab_name, dl.lab_address, dl.lab_contact_number,
		rv.diagnostic_test_id, rv.diagnostic_test_component_id, rv.test_result_value_id,
		rv.result_value, rv.result_status, format_datetime(rv.result_date) AS result_date,
		rv.result_comment, dtm.test_name, dtcm.test_component_name,
		rr.normal_min, rr.normal_max, rr.units
	FROM tbl_medical_record mr
	INNER JOIN tbl_medical_record_user_mapping rm ON rm.record_id = mr.record_id
	LEFT JOIN tbl_patient_report_attachment pa ON pa.record_id = mr.record_id
	LEFT JOIN tbl_patient_diagnostic_report pr ON pr.patient_diagnostic_report_id = pa.patient_diagnostic_report_id AND pr.is_deleted = 0
	LEFT JOIN tbl_diagnostic_lab dl ON dl.diagnostic_lab_id = pr.diagnostic_lab_id
	LEFT JOIN tbl_patient_diagnostic_test_result_value rv ON rv.patient_diagnostic_report_id = pr.patient_diagnostic_report_id
	LEFT JOIN tbl_disease_profile_diagnostic_test_master dtm ON dtm.diagnostic_test_id = rv.diagnostic_test_id
	LEFT JOIN tbl_disease_profile_diagnostic_test_component_master dtcm ON dtcm.diagnostic_test_component_id = rv.diagnostic_test_component_id
	LEFT JOIN tbl_diagnostic_test_reference_range rr 
		ON rr.diagnostic_test_id = rv.diagnostic_test_id 
		AND rr.diagnostic_test_component_id = rv.diagnostic_test_component_id
		AND rr.is_deleted = 0
	WHERE mr.record_id IN ?
	ORDER BY mr.record_id DESC
	`

	err := r.db.Raw(query, recordIDs).Scan(&results).Error
	return results, err
}

func (r *tblMedicalRecordRepositoryImpl) getTotalRecordCount(patientId uint64) (int64, error) {
	var total int64
	query := `
		SELECT COUNT(DISTINCT mr.record_id)
		FROM tbl_medical_record mr
		INNER JOIN tbl_medical_record_user_mapping rm ON rm.record_id = mr.record_id
		WHERE mr.is_deleted = 0 AND rm.user_id = ?
	`
	err := r.db.Raw(query, patientId).Scan(&total).Error
	return total, err
}

func (p *tblMedicalRecordRepositoryImpl) ProcessMedicalRecordResponse(rows []models.ReportRow) []map[string]interface{} {
	if len(rows) == 0 {
		return []map[string]interface{}{}
	}

	reportMap := make(map[string]map[string]interface{})
	order := make([]string, 0)

	for _, item := range rows {
		reportID := strconv.FormatUint(item.PatientDiagnosticReportID, 10)
		if _, exists := reportMap[reportID]; !exists {
			order = append(order, reportID)
			var diagnosticLab map[string]interface{}

			if item.PatientDiagnosticReportID == 0 {
				diagnosticLab = map[string]interface{}{}
			} else {
				diagnosticLab = map[string]interface{}{
					"diagnostic_lab_id":       item.DiagnosticLabID,
					"lab_name":                item.LabName,
					"collected_date":          item.CollectedDate,
					"report_date":             item.ReportDate,
					"report_status":           item.ReportStatus,
					"report_name":             item.ReportName,
					"comments":                item.ResultComment,
					"collected_at":            item.CollectedAt,
					"patient_diagnostic_test": []map[string]interface{}{},
				}
			}

			reportMap[reportID] = map[string]interface{}{
				"record_id":                    item.RecordId,
				"record_name":                  item.RecordName,
				"status":                       item.Status,
				"record_size":                  item.RecordSize,
				"file_type":                    item.FileType,
				"upload_source":                item.UploadSource,
				"source_account":               item.SourceAccount,
				"record_category":              item.RecordCategory,
				"record_url":                   item.RecordUrl,
				"digitize_flag":                item.DigitizeFlag,
				"patient_diagnostic_report_id": reportID,
				"patient_id":                   item.PatientId,
				"uploaded_diagnostic":          diagnosticLab,
			}
		}

		if item.PatientDiagnosticReportID == 0 {
			continue
		}

		report := reportMap[reportID]
		diagnosticLab := report["uploaded_diagnostic"].(map[string]interface{})
		testResultValue := map[string]interface{}{
			"diagnostic_test_id": item.DiagnosticTestID,
			"result_value":       item.ResultValue,
			"result_status":      item.ResultStatus,
			"result_date":        item.ResultDate,
			"result_comment":     item.ResultComment,
			"qualifier":          item.Qualifier,
		}
		testReferenceRange := map[string]interface{}{
			"diagnostic_test_id":           item.DiagnosticTestID,
			"diagnostic_test_component_id": item.DiagnosticTestComponentID,
			"normal_min":                   item.NormalMin,
			"normal_max":                   item.NormalMax,
			"age":                          item.Age,
			"age_group":                    item.AgeGroup,
			"gender":                       item.Gender,
			"is_deleted":                   item.RefIsDeleted,
			"units":                        item.RefUnits,
		}
		testComponent := map[string]interface{}{
			"diagnostic_test_component_id": item.DiagnosticTestComponentID,
			"test_component_name":          item.TestComponentName,
			"units":                        item.ComponentUnit,
			"test_result_value":            []map[string]interface{}{testResultValue},
			"test_reference_range":         []map[string]interface{}{testReferenceRange},
		}

		pdtList := diagnosticLab["patient_diagnostic_test"].([]map[string]interface{})
		var existingTest map[string]interface{}
		for _, pdt := range pdtList {
			test := pdt["diagnostic_test"].(map[string]interface{})
			if test["diagnostic_test_id"] == item.DiagnosticTestID {
				existingTest = test
				break
			}
		}

		if existingTest != nil {
			existingTest["test_components"] = append(
				existingTest["test_components"].([]map[string]interface{}),
				testComponent,
			)
		} else {
			newTest := map[string]interface{}{
				"diagnostic_test_id": item.DiagnosticTestID,
				"test_name":          item.TestName,
				"test_note":          item.TestNote,
				"test_date":          item.TestDate,
				"test_components":    []map[string]interface{}{testComponent},
			}
			pdt := map[string]interface{}{
				"diagnostic_test": newTest,
			}
			diagnosticLab["patient_diagnostic_test"] = append(pdtList, pdt)
		}
	}
	finalReports := make([]map[string]interface{}, 0, len(order))
	for _, reportID := range order {
		finalReports = append(finalReports, reportMap[reportID])
	}

	return finalReports
}

func (r *tblMedicalRecordRepositoryImpl) CreateTblMedicalRecord(tx *gorm.DB, data *models.TblMedicalRecord) (*models.TblMedicalRecord, error) {
	if tx == nil {
		return nil, fmt.Errorf("transaction is nil")
	}
	if err := tx.Create(data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (r *tblMedicalRecordRepositoryImpl) CreateMultipleTblMedicalRecords(tx *gorm.DB, records []*models.TblMedicalRecord) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	if len(records) == 0 {
		return nil
	}
	if err := tx.Create(records).Error; err != nil {
		return fmt.Errorf("failed to create multiple medical records: %w", err)
	}
	return nil
}

func (r *tblMedicalRecordRepositoryImpl) UpdateTblMedicalRecord(data *models.TblMedicalRecord) (*models.TblMedicalRecord, error) {
	tx := r.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	updateFields := map[string]interface{}{}
	if data.RecordName != "" {
		updateFields["record_name"] = data.RecordName
	}
	if data.RecordSize != 0 {
		updateFields["record_size"] = data.RecordSize
	}
	if data.FileType != "" {
		updateFields["file_type"] = data.FileType
	}
	if data.UploadSource != "" {
		updateFields["upload_source"] = data.UploadSource
	}
	if data.UploadDestination != "" {
		updateFields["upload_destination"] = data.UploadDestination
	}
	if data.SourceAccount != "" {
		updateFields["source_account"] = data.SourceAccount
	}
	if data.RecordCategory != "" {
		updateFields["record_category"] = data.RecordCategory
	}
	if data.Description != "" {
		updateFields["description"] = data.Description
	}
	if data.RecordUrl != "" {
		updateFields["record_url"] = data.RecordUrl
	}
	if data.Status != "" {
		updateFields["status"] = data.Status
	}
	if data.QueueName != "" {
		updateFields["queue_name"] = data.QueueName
	}
	if data.ProcessingStartedAt != nil {
		updateFields["processing_started_at"] = data.ProcessingStartedAt
	}
	if data.CompletedAt != nil {
		updateFields["completed_at"] = data.CompletedAt
	}
	if data.NextRetryAt != nil {
		updateFields["next_retry_at"] = data.NextRetryAt
	}
	if data.IsExpired != nil {
		updateFields["is_expired"] = data.IsExpired
	}
	if data.FileData != nil {
		updateFields["file_data"] = data.FileData
	}
	if data.DigitizeFlag > 0 {
		updateFields["digitize_flag"] = data.DigitizeFlag
	}
	if data.RetryCount > 0 {
		updateFields["retry_count"] = data.RetryCount
	}
	if len(data.Metadata) != 0 {
		log.Println("Updating metadata")
		var existing models.TblMedicalRecord
		if err := tx.Where("record_id = ?", data.RecordId).First(&existing).Error; err != nil {
			log.Println("@UpdateTblMedicalRecord -> Metadata Rollback", err)
			tx.Rollback()
			return nil, err
		}
		log.Println("data.Metadata:", data.Metadata)
		var existingMap map[string]interface{}
		if err := json.Unmarshal(existing.Metadata, &existingMap); err != nil {
			log.Println("Error Unmarshal existing.metadata", err)
			existingMap = map[string]interface{}{}
		}
		log.Println("existingMap:", existingMap)
		var newMap map[string]interface{}
		if err := json.Unmarshal(data.Metadata, &newMap); err != nil {
			log.Println("Error Unmarshal data.metadata", err)
			newMap = map[string]interface{}{}
		}
		log.Println("New map created:", newMap)
		for k, v := range newMap {
			existingMap[k] = v
		}
		mergedMetadata, err := json.Marshal(existingMap)
		if err != nil {
			log.Println("Error Marshal:", err)
			tx.Rollback()
			return nil, err
		}
		updateFields["metadata"] = datatypes.JSON(mergedMetadata)
	}
	if data.IsDeleted != 0 {
		updateFields["is_deleted"] = data.IsDeleted
	}
	updateFields["is_verified"] = data.IsVerified
	updateFields["updated_at"] = time.Now()
	updateFields["error_message"] = data.ErrorMessage

	err := tx.Model(&models.TblMedicalRecord{}).
		Where("record_id = ?", data.RecordId).
		Updates(updateFields).Error

	if err != nil {
		tx.Rollback()
		return nil, err
	}
	var updatedRecord models.TblMedicalRecord
	if err := tx.Where("record_id = ?", data.RecordId).First(&updatedRecord).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &updatedRecord, nil
}

func (r *tblMedicalRecordRepositoryImpl) GetMedicalRecordByRecordId(RecordId uint64) (*models.TblMedicalRecord, error) {
	var obj models.TblMedicalRecord
	err := r.db.First(&obj, RecordId).Error
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func (r *tblMedicalRecordRepositoryImpl) DeleteTblMedicalRecord(id int, updatedBy string) error {
	return r.db.Where("record_id = ?", id).Delete(&models.TblMedicalRecord{}).Error
}

func (r *tblMedicalRecordRepositoryImpl) CreateMedicalRecordMappings(tx *gorm.DB, mappings *[]models.TblMedicalRecordUserMapping) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}
	return tx.Create(mappings).Error
}

func (r *tblMedicalRecordRepositoryImpl) UpdateMedicalRecordMappingByRecordId(tx *gorm.DB, recordId *uint64, updates map[string]interface{}) error {
	if recordId == nil || len(updates) == 0 {
		return errors.New("invalid record ID or empty update data")
	}
	log.Println("UpdateMedicalRecordMappingByRecordId ", updates)
	result := tx.Model(&models.TblMedicalRecordUserMapping{}).
		Where("record_id = ?", recordId).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no mapping found with record_id %d", recordId)
	}
	return nil
}

func (r *tblMedicalRecordRepositoryImpl) DeleteMecationRecordMappings(id int) error {
	return r.db.Where("record_id = ?", id).Delete(&models.TblMedicalRecordUserMapping{}).Error
}

func (r *tblMedicalRecordRepositoryImpl) GetMedicalRecordMappings(recordID uint64) (*models.TblMedicalRecordUserMapping, error) {
	var mapping models.TblMedicalRecordUserMapping
	err := r.db.Where("record_id=?", recordID).Find(&mapping).Error
	return &mapping, err
}

func (r *tblMedicalRecordRepositoryImpl) DeleteTblMedicalRecordWithMappings(id int, updatedBy string) error {
	tx := r.db.Begin()

	if tx.Error != nil {
		return tx.Error
	}

	result := tx.Where("record_id = ?", id).Delete(&models.TblMedicalRecord{})
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("record with id %d not found", id)
	}

	result = tx.Where("record_id = ?", id).Delete(&models.TblMedicalRecordUserMapping{})
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("no mappings found for record with id %d", id)
	}

	return tx.Commit().Error
}

func (r *tblMedicalRecordRepositoryImpl) ExistsRecordForUser(userId uint64, source, url string) (bool, error) {
	var count int64
	err := r.db.
		Table("tbl_medical_record").
		Joins("INNER JOIN tbl_medical_record_user_mapping ON tbl_medical_record.record_id = tbl_medical_record_user_mapping.record_id").
		Where("tbl_medical_record_user_mapping.user_id = ? AND tbl_medical_record.upload_source = ? AND tbl_medical_record.record_url = ?", userId, source, url).
		Count(&count).Error

	return count > 0, err
}

func (r *tblMedicalRecordRepositoryImpl) IsRecordBelongsToUser(userID uint64, recordID uint64) (bool, error) {
	var mapping models.TblMedicalRecordUserMapping
	err := r.db.Where("user_id = ? AND record_id = ?", userID, recordID).First(&mapping).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *tblMedicalRecordRepositoryImpl) MovePatientRecord(patientId, targetPatientId, recordId, reportId uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := r.UpdateMedicalRecordMappingByRecordId(tx, &recordId, map[string]interface{}{"user_id": targetPatientId}); err != nil {
			return fmt.Errorf("record mapping move failed: %w", err)
		}
		if err := r.UpdatePatientDiagnosticReports(tx, patientId, targetPatientId, reportId); err != nil {
			return fmt.Errorf("report move failed: %w", err)
		}
		if err := r.UpdatePatientDiagnosticTestResults(tx, patientId, targetPatientId, reportId); err != nil {
			return fmt.Errorf("test results move failed: %w", err)
		}
		if err := r.UpdatePatientReportAttachments(tx, patientId, targetPatientId, reportId); err != nil {
			return fmt.Errorf("attachments move failed: %w", err)
		}
		return nil
	})
}

func (r *tblMedicalRecordRepositoryImpl) UpdatePatientDiagnosticReports(tx *gorm.DB, patientId, targetPatientId, reportId uint64) error {
	return tx.Model(&models.PatientDiagnosticReport{}).
		Where("patient_id = ? AND patient_diagnostic_report_id = ?", patientId, reportId).
		Update("patient_id", targetPatientId).Error
}

func (r *tblMedicalRecordRepositoryImpl) UpdatePatientDiagnosticTestResults(tx *gorm.DB, patientId, targetPatientId, reportId uint64) error {
	return tx.Model(&models.PatientDiagnosticTestResultValue{}).
		Where("patient_id = ? AND patient_diagnostic_report_id = ?", patientId, reportId).
		Update("patient_id", targetPatientId).Error
}

func (r *tblMedicalRecordRepositoryImpl) UpdatePatientReportAttachments(tx *gorm.DB, patientId, targetPatientId, reportId uint64) error {
	return tx.Model(&models.PatientReportAttachment{}).
		Where("patient_id = ? AND patient_diagnostic_report_id = ?", patientId, reportId).
		Update("patient_id", targetPatientId).Error
}

func (r *tblMedicalRecordRepositoryImpl) GetRecordsByIDs(recordIDs []uint64, isDeleted int) ([]models.TblMedicalRecord, error) {
	var records []models.TblMedicalRecord
	err := r.db.
		Table("tbl_medical_record").
		Where("record_id IN ? AND is_deleted = ?", recordIDs, isDeleted).
		Find(&records).Error
	return records, err
}

func (r *tblMedicalRecordRepositoryImpl) GetReportRecordMapping(
	userID uint64,
	category, tag string,
	isDeleted int,
) (
	map[uint64][]uint64,
	int64,
	map[string]int64,
	error,
) {

	var attachments []struct {
		PatientDiagnosticReportId uint64 `gorm:"column:patient_diagnostic_report_id"`
		RecordId                  uint64 `gorm:"column:record_id"`
	}

	status := config.PropConfig.SystemVaribale.Status
	statuses := strings.Split(status, ",")

	query := r.db.
		Table("tbl_patient_report_attachment AS pra").
		Select("pra.patient_diagnostic_report_id, pra.record_id").
		Joins("JOIN tbl_medical_record_user_mapping AS mrum ON pra.record_id = mrum.record_id").
		Joins("JOIN tbl_medical_record AS mr ON pra.record_id = mr.record_id").
		Joins("LEFT JOIN tbl_patient_diagnostic_report AS report ON report.patient_diagnostic_report_id = pra.patient_diagnostic_report_id").
		Where("mrum.user_id = ? AND mr.is_deleted = ? AND mr.status IN (?)", userID, isDeleted, statuses)

	if tag != "" {
		subQuery := r.db.
			Table("tbl_patient_report_attachment AS pra2").
			Select("pra2.patient_diagnostic_report_id").
			Joins("JOIN tbl_medical_record AS mr2 ON pra2.record_id = mr2.record_id").
			Joins("JOIN tbl_user_tag AS ut ON ut.record_id = mr2.record_id").
			Where("ut.tag_name = ? AND ut.user_id = ?", tag, userID)

		query = query.Where("pra.patient_diagnostic_report_id IN (?)", subQuery)
	}

	if category != "" {
		query = query.Where("mr.record_category = ?", category)
	}

	query = query.Order("report.report_date DESC")

	if err := query.Find(&attachments).Error; err != nil {
		return nil, 0, nil, err
	}

	reportMap := make(map[uint64][]uint64)
	for _, a := range attachments {
		reportMap[a.PatientDiagnosticReportId] = append(reportMap[a.PatientDiagnosticReportId], a.RecordId)
	}

	type CategoryCount struct {
		RecordCategory string `gorm:"column:record_category"`
		Count          int64  `gorm:"column:count"`
	}

	categoryCountMap := make(map[string]int64)
	var totalCount int64

	var recordCounts []CategoryCount
	recordQuery := r.db.
		Table("tbl_medical_record AS mr").
		Joins("JOIN tbl_medical_record_user_mapping AS mrum ON mr.record_id = mrum.record_id").
		Where("mrum.user_id = ? AND mr.is_deleted = ? AND mr.status IN (?)", userID, isDeleted, statuses)

	if err := recordQuery.Select("mr.record_category, COUNT(*) as count").
		Group("mr.record_category").
		Scan(&recordCounts).Error; err != nil {
		return nil, 0, nil, err
	}

	for _, rc := range recordCounts {
		categoryCountMap[rc.RecordCategory] = rc.Count
		totalCount += rc.Count
	}

	var prescriptionCount int64
	r.db.
		Table("tbl_patient_prescription").
		Where("patient_id = ? AND is_deleted = ?", userID, 0).
		Count(&prescriptionCount)

	if prescriptionCount > 0 {
		categoryCountMap["medications"] = prescriptionCount
	}

	return reportMap, totalCount, categoryCountMap, nil
}

func (r *tblMedicalRecordRepositoryImpl) AddTag(tag *models.UserTag) error {
	return r.db.Create(tag).Error
}

func (r *tblMedicalRecordRepositoryImpl) GetAllReportTag(userId uint64, limit, offset int) ([]models.UserTag, int64, error) {
	var tags []models.UserTag
	var total int64

	if err := r.db.Model(&models.UserTag{}).Where("user_id = ?", userId).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Model(&models.UserTag{}).Where("user_id = ?", userId).Order("created_at DESC").Limit(limit).Offset(offset).Find(&tags).Error; err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}
