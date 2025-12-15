package models

import "time"

type OTPMaster struct {
	ID            uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID        *uint64    `gorm:"column:user_id"`
	IdentityType  string     `gorm:"column:identity_type;size:20;not null"`
	IdentityValue string     `gorm:"column:identity_value;size:100;not null"`
	Purpose       string     `gorm:"column:purpose;size:50;not null"`
	ReferenceType *string    `gorm:"column:reference_type;size:50"`
	ReferenceID   *uint64    `gorm:"column:reference_id"`
	FamilyID      *uint64    `gorm:"column:family_id"`
	OTPHash       string     `gorm:"column:otp_hash;size:255;not null"`
	Channel       string     `gorm:"column:channel;size:20;not null"`
	ExpiresAt     time.Time  `gorm:"column:expires_at;not null"`
	IsUsed        bool       `gorm:"column:is_used;default:false"`
	AttemptCount  int        `gorm:"column:attempt_count;default:0"`
	MaxAttempts   int        `gorm:"column:max_attempts;default:5"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime"`
	UsedAt        *time.Time `gorm:"column:used_at"`
}

func (OTPMaster) TableName() string {
	return "tbl_otp_master"
}

type GenerateOTPRequest struct {
	UserID        *uint64
	IdentityType  string
	IdentityValue string
	Purpose       string

	ReferenceType *string
	ReferenceID   *uint64
	FamilyID      *uint64

	Channel string
}

type ValidateOTPRequest struct {
	IdentityType  string
	IdentityValue string
	Purpose       string
	OTP           string
}
