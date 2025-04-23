package models

import (
	"time"
)

type TblUserToken struct {
	Id         uint      `gorm:"column:user_token_id;primaryKey" json:"user_token_id"`
	UserId     uint64    `gorm:"column:user_id;" json:"user_id"`
	AuthToken  string    `gorm:"column:auth_token;" json:"auth_token"`
	Provider   string    `gorm:"column:provider;" json:"provider"`
	ProviderId string    `gorm:"column:provider_id;" json:"provider_id"`
	CreatedAt  time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (TblUserToken) TableName() string {
	return "tbl_user_token"
}

type ThirdPartyTokenStatus struct {
	DigiLockerPresent bool `json:"DigiLocker"`
	IsDLExpired       bool `json:"IsDLExpired"`
	GmailPresent      bool `json:"GmailPresent"`
}
