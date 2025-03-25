package models

import "time"

type Caregiver struct {
	CaregiverId uint      `json:"caregiver_id" gorm:"primaryKey;autoIncrement"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	ContactInfo string    `json:"contact_info"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Caregiver) TableName() string {
	return "tbl_caregiver"
}
