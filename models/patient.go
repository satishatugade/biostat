package models

import (
	"time"
)

type Patient struct {
	PatientId          uint      `gorm:"primaryKey;autoIncrement" json:"patient_id"`
	FirstName          string    `gorm:"type:varchar(255)" json:"first_name"`
	LastName           string    `gorm:"type:varchar(255)" json:"last_name"`
	DateOfBirth        time.Time `gorm:"type:date" json:"date_of_birth"`
	Gender             string    `gorm:"type:varchar(50)" json:"gender"`
	ContactInfo        string    `gorm:"type:text" json:"contact_info"`
	Address            string    `gorm:"type:text" json:"address"`
	EmergencyContact   string    `gorm:"type:varchar(255)" json:"emergency_contact"`
	AbhaNumber         string    `gorm:"type:varchar(255)" json:"abha_number"`
	BloodGroup         string    `gorm:"type:varchar(50)" json:"blood_group"`
	Nationality        string    `gorm:"type:varchar(255)" json:"nationality"`
	CitizenshipStatus  string    `gorm:"type:varchar(255)" json:"citizenship_status"`
	PassportNumber     string    `gorm:"type:varchar(255)" json:"passport_number"`
	CountryOfResidence string    `gorm:"type:varchar(255)" json:"country_of_residence"`
	IsIndianOrigin     bool      `gorm:"type:boolean" json:"is_indian_origin"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Patient) TableName() string {
	return "tbl_patient"
}
