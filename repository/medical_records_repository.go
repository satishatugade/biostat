package repository

import (
	"biostat/models"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type TblMedicalRecordRepository interface {
	GetAllMedicalRecord(patientId uint64, limit int, offset int) ([]models.ReportRow, int64, error)
	ProcessMedicalRecordResponse(data []models.ReportRow) []map[string]interface{}
	GetMedicalRecordsByUserID(userID uint64, recordIdsMap map[uint64]uint64) ([]models.TblMedicalRecord, error)
	CreateTblMedicalRecord(data *models.TblMedicalRecord) (*models.TblMedicalRecord, error)
	CreateMultipleTblMedicalRecords(data *[]models.TblMedicalRecord) error
	UpdateTblMedicalRecord(data *models.TblMedicalRecord, updatedBy string) (*models.TblMedicalRecord, error)
	GetSingleTblMedicalRecord(id int64) (*models.TblMedicalRecord, error)
	DeleteTblMedicalRecord(id int, updatedBy string) error
	IsRecordBelongsToUser(userID uint64, recordID int64) (bool, error)
	ExistsRecordForUser(userId uint64, source, url string) (bool, error)

	CreateMedicalRecordMappings(mappings *[]models.TblMedicalRecordUserMapping) error
	GetMedicalRecordMappings(recordID int64) (*models.TblMedicalRecordUserMapping, error)
	DeleteMecationRecordMappings(id int) error

	DeleteTblMedicalRecordWithMappings(id int, user_id string) error
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
				"record_url":                   item.RecordURL,
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

func (r *tblMedicalRecordRepositoryImpl) CreateTblMedicalRecord(data *models.TblMedicalRecord) (*models.TblMedicalRecord, error) {
	err := r.db.Create(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *tblMedicalRecordRepositoryImpl) CreateMultipleTblMedicalRecords(records *[]models.TblMedicalRecord) error {
	return r.db.Create(records).Error
}

func (r *tblMedicalRecordRepositoryImpl) UpdateTblMedicalRecord(data *models.TblMedicalRecord, updatedBy string) (*models.TblMedicalRecord, error) {
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
	if data.ErrorMessage != "" {
		updateFields["error_message"] = data.ErrorMessage
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
	if len(data.Metadata) != 0 {
		updateFields["metadata"] = data.Metadata
	}
	updateFields["is_verified"] = data.IsVerified
	updateFields["is_deleted"] = data.IsDeleted
	updateFields["updated_at"] = time.Now()

	err := tx.Model(&models.TblMedicalRecord{}).
		Where("record_id = ?", data.RecordId).
		Updates(updateFields).Error

	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return data, nil
}

func (r *tblMedicalRecordRepositoryImpl) GetSingleTblMedicalRecord(id int64) (*models.TblMedicalRecord, error) {
	var obj models.TblMedicalRecord
	err := r.db.First(&obj, id).Error
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func (r *tblMedicalRecordRepositoryImpl) DeleteTblMedicalRecord(id int, updatedBy string) error {
	return r.db.Where("record_id = ?", id).Delete(&models.TblMedicalRecord{}).Error
}

func (r *tblMedicalRecordRepositoryImpl) CreateMedicalRecordMappings(mappings *[]models.TblMedicalRecordUserMapping) error {
	return r.db.Create(mappings).Error
}

func (r *tblMedicalRecordRepositoryImpl) DeleteMecationRecordMappings(id int) error {
	return r.db.Where("record_id = ?", id).Delete(&models.TblMedicalRecordUserMapping{}).Error
}

func (r *tblMedicalRecordRepositoryImpl) GetMedicalRecordMappings(recordID int64) (*models.TblMedicalRecordUserMapping, error) {
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

func (r *tblMedicalRecordRepositoryImpl) IsRecordBelongsToUser(userID uint64, recordID int64) (bool, error) {
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
