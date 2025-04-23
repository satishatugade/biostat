package models

import (
	"time"
)

type TblUserToken struct {
	Id           uint      `gorm:"column:user_token_id;primaryKey" json:"user_token_id"`
	UserId       uint64    `gorm:"column:user_id;" json:"user_id"`
	AuthToken    string    `gorm:"column:auth_token;" json:"auth_token"`
	RefreshToken string    `gorm:"column:refresh_token;" json:"refresh_token"`
	Provider     string    `gorm:"column:provider;" json:"provider"`
	ProviderId   string    `gorm:"column:provider_id;" json:"provider_id"`
	CreatedAt    time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	ExpiresAt    time.Time `gorm:"column:expires_at;" json:"expires_at"`
}

func (TblUserToken) TableName() string {
	return "tbl_user_token"
}

type ThirdPartyTokenStatus struct {
	DigiLockerPresent bool `json:"DigiLocker"`
	IsDLExpired       bool `json:"IsDLExpired"`
	GmailPresent      bool `json:"GmailPresent"`
}
