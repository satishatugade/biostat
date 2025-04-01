package models

import (
	"time"

	"gorm.io/datatypes"
)

type TblMedicalRecord struct {
	RecordId       uint           `gorm:"column:record_id;primaryKey;autoIncrement" json:"record_id"`
	RecordName     string         `gorm:"column:record_name;not null" json:"record_name"`
	RecordSize     int64          `gorm:"column:record_size;" json:"record_size"`
	RecordExt      string         `gorm:"column:record_ext;" json:"record_ext"`
	UploadSource   string         `gorm:"column:upload_source;not null" json:"upload_source"`
	RecordType     string         `gorm:"column:record_type;" json:"record_type"`
	FilePath       string         `gorm:"column:file_path;" json:"file_path"`
	Description    string         `gorm:"column:description;" json:"description"`
	SourceMetadata datatypes.JSON `gorm:"column:source_metadata;" json:"source_metadata"`
	FileData       []byte         `gorm:"column:file_data;" json:"file_data"`
	CreatedAt      time.Time      `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	IsActive       bool           `gorm:"column:is_active;default:true" json:"is_active"`
}

func (TblMedicalRecord) TableName() string {
	return "tbl_medical_record"
}

type TblMedicalRecordUserMapping struct {
	ID       uint  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID   int64 `gorm:"column:user_id;not null" json:"user_id"`
	RecordID int64 `gorm:"column:record_id;not null" json:"record_id"`
}

func (TblMedicalRecordUserMapping) TableName() string {
	return "tbl_medical_record_user_mapping"
}
