package models

import (
	"biostat/constant"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type TblMedicalRecord struct {
	RecordId          uint64         `gorm:"column:record_id;primaryKey;autoIncrement" json:"record_id"`
	RecordName        string         `gorm:"column:record_name;not null" json:"record_name"`
	RecordSize        int64          `gorm:"column:record_size;" json:"record_size"`
	FileType          string         `gorm:"column:file_type;" json:"file_type"`
	UploadSource      string         `gorm:"column:upload_source;not null" json:"upload_source"`
	UploadDestination string         `gorm:"column:upload_destination;not null" json:"upload_destination"`
	SourceAccount     string         `gorm:"column:source_account;" json:"source_account"`
	RecordCategory    string         `gorm:"column:record_category;" json:"record_category"`
	Description       string         `gorm:"column:description;" json:"description"`
	FileData          []byte         `gorm:"column:file_data;" json:"file_data"`
	RecordUrl         string         `gorm:"column:record_url;" json:"record_url"`
	FetchedAt         time.Time      `gorm:"column:fetched_at;default:CURRENT_TIMESTAMP" json:"fetched_at"`
	UploadedBy        uint64         `gorm:"column:uploaded_by;" json:"uploaded_by"`
	IsVerified        bool           `gorm:"column:is_verified;default:false" json:"is_verified"`
	Metadata          datatypes.JSON `gorm:"column:metadata;" json:"metadata"`
	IsDeleted         int            `gorm:"column:is_deleted;default:0" json:"is_deleted"`
	DigitizeFlag      int            `gorm:"column:digitize_flag;default:0" json:"digitize_flag"`

	Status              constant.JobStatus `gorm:"type:status_enum" json:"status"`
	RetryCount          int                `gorm:"column:retry_count;default:0" json:"retry_count"`
	MaxRetry            int                `gorm:"column:max_retry" json:"max_retry"`
	QueueName           string             `gorm:"column:queue_name;type:varchar(100);not null;default:''" json:"queue_name"`
	ErrorMessage        string             `gorm:"column:error_message;type:text" json:"error_message"`
	ProcessingStartedAt *time.Time         `gorm:"column:processing_started_at" json:"processing_started_at"`
	CompletedAt         *time.Time         `gorm:"column:completed_at" json:"completed_at"`
	NextRetryAt         *time.Time         `gorm:"column:next_retry_at" json:"next_retry_at"`
	IsExpired           *bool              `gorm:"column:is_expired;default:false" json:"is_expired"`

	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TblMedicalRecord) TableName() string {
	return "tbl_medical_record"
}

type TblMedicalRecordUserMapping struct {
	MedicalRecordUserMappingId uint64 `gorm:"primaryKey;autoIncrement" json:"medical_record_user_mapping_id"`
	UserID                     uint64 `gorm:"column:user_id;not null" json:"user_id"`
	RecordID                   uint64 `gorm:"column:record_id;not null" json:"record_id"`
	IsUnknownRecord            bool   `gorm:"column:is_unknown_record" json:"is_unknown_record"`
}

func (TblMedicalRecordUserMapping) TableName() string {
	return "tbl_medical_record_user_mapping"
}

type DigiLockerFile struct {
	Data        []byte `gorm:"data" json:"data"`
	ContentType string `gorm:"content-type" json:"content-type"`
	HMAC        string `gorm:"hmac" json:"hmac"`
}

type LocalServerFile struct {
	Data        []byte `json:"data"`
	ContentType string `json:"content-type"`
}

type DiagnosticReferenceRangeRes struct {
	Age       int    `json:"age"`
	AgeGroup  string `json:"age_group"`
	Gender    string `json:"gender"`
	NormalMin string `json:"normal_min"`
	NormalMax string `json:"normal_max"`
	Units     string `json:"units"`
}

type TestResultValueRes struct {
	ResultValue   string `json:"result_value"`
	Qualifier     string `json:"qualifier"`
	ResultComment string `json:"result_comment"`
	ResultDate    string `json:"result_date"`
	ResultStatus  string `json:"result_status"`
}

type TestComponentRes struct {
	DiagnosticTestComponentID uint64                        `json:"diagnostic_test_component_id"`
	TestComponentName         string                        `json:"test_component_name"`
	Units                     string                        `json:"units"`
	TestReferenceRange        []DiagnosticReferenceRangeRes `json:"test_reference_range"`
	TestResultValue           []TestResultValueRes          `json:"test_result_value"`
}

type DiagnosticTestRes struct {
	DiagnosticTestID uint64             `json:"diagnostic_test_id"`
	TestName         string             `json:"test_name"`
	TestNote         string             `json:"test_note"`
	TestDate         time.Time          `json:"test_date"`
	TestComponents   []TestComponentRes `json:"test_components"`
}

type UploadedDiagnosticRes struct {
	CollectedDate         string              `json:"collected_date"`
	CollectedAt           string              `json:"collected_at"`
	ReportDate            string              `json:"report_date"`
	ReportName            string              `json:"report_name"`
	ReportStatus          string              `json:"report_status"`
	DiagnosticLabID       uint64              `json:"diagnostic_lab_id"`
	LabName               string              `json:"lab_name"`
	Comments              string              `json:"comments"`
	IsDeleted             int                 `json:"is_deleted"`
	PatientDiagnosticTest []DiagnosticTestRes `json:"patient_diagnostic_test"`
}

type MedicalRecordResponseRes struct {
	DigitizeFlag              int                    `json:"digitize_flag"`
	FileType                  string                 `json:"file_type"`
	PatientDiagnosticReportID string                 `json:"patient_diagnostic_report_id"`
	PatientID                 uint64                 `json:"patient_id"`
	RecordCategory            string                 `json:"record_category"`
	RecordID                  uint64                 `json:"record_id"`
	RecordName                string                 `json:"record_name"`
	RecordSize                int64                  `json:"record_size"`
	RecordURL                 string                 `json:"record_url"`
	RecordDescription         string                 `json:"record_description"`
	IsVerified                bool                   `json:"is_verified"`
	SourceAccount             string                 `json:"source_account"`
	Status                    string                 `json:"status"`
	UploadSource              string                 `json:"upload_source"`
	ErrorMessage              string                 `json:"error_message"`
	CreatedAt                 string                 `json:"created_at"`
	IsDeleted                 int                    `json:"is_deleted"`
	UploadedDiagnostic        *UploadedDiagnosticRes `json:"uploaded_diagnostic"`
}

type DigitizationPayload struct {
	RecordID    uint64    `json:"record_id"`
	UserID      uint64    `json:"user_id"`
	PatientName string    `json:"patient_name"`
	FilePath    string    `json:"file_path"`
	Category    string    `json:"category"`
	FileName    string    `json:"file_name"`
	AuthUserID  string    `json:"auth_user_id"`
	ProcessID   uuid.UUID `json:"process_id"`
}
