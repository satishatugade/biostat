package models

import (
	"time"
)

type TblUserGtoken struct {
	Id        uint      `gorm:"column:id;primaryKey" json:"Id"`
	UserId    int64     `gorm:"column:user_id;" json:"User_id"`
	AuthToken string    `gorm:"column:auth_token;" json:"Auth_token"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"Created_at"`
}

func (TblUserGtoken) TableName() string {
	return "tbl_user_gtoken"
}
