package models

import "time"

type Caregiver struct {
	CaregiverId uint      `json:"caregiver_id" gorm:"primaryKey;autoIncrement"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	ContactInfo string    `json:"contact_info"`
	Gender      string    `json:"gender"`
	DateOfBirth time.Time `json:"date_of_birth"`
	MobileNo    string    `json:"mobile_no"`
	Email       string    `json:"email"`
	Address     string    `json:"address"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Caregiver) TableName() string {
	return "tbl_caregiver"
}

type UserRelation struct {
	UserId     uint64 `json:"user_id"`
	RelationId uint64 `json:"relation_id"`
}
