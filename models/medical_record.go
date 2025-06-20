package models

import (
	"biostat/constant"
	"time"

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
	QueueName           string             `gorm:"column:queue_name;type:varchar(100);not null;default:''" json:"queue_name"`
	RetryCount          int                `gorm:"column:retry_count;default:0" json:"retry_count"`
	MaxRetry            int                `gorm:"column:max_retry" json:"max_retry"`
	ErrorMessage        string             `gorm:"column:error_message;type:text" json:"error_message"`
	ProcessingStartedAt *time.Time         `gorm:"column:processing_started_at" json:"processing_started_at"`
	CompletedAt         *time.Time         `gorm:"column:completed_at" json:"completed_at"`
	NextRetryAt         *time.Time         `gorm:"column:next_retry_at" json:"next_retry_at"`
	IsExpired           bool               `gorm:"column:is_expired;default:false" json:"is_expired"`

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
