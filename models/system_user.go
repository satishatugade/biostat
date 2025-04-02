package models

import (
	"time"
)

type SystemUser struct {
	SystemUserId       uint      `gorm:"primaryKey;autoIncrement;column:system_user_id"`
	UserId             uint      `gorm:"not null;column:user_id"`
	AuthUserId         string    `gorm:"not null;column:auth_user_id"`
	RoleId             uint      `gorm:"not null;column:role_id"`
	Username           string    `gorm:"size:255;column:username"`
	Password           string    `gorm:"size:255;column:password"`
	UserState          string    `gorm:"size:50;column:user_state"`
	AuthType           string    `gorm:"size:50;column:auth_type"`
	AuthStatus         string    `gorm:"size:50;column:auth_status"`
	AuthDate           time.Time `gorm:"default:CURRENT_TIMESTAMP;column:auth_date"`
	ActivationFlag     bool      `gorm:"default:false;column:activation_flag"`
	FirstLoginFlag     bool      `gorm:"default:true;column:first_login_flag"`
	AccountExpired     bool      `gorm:"default:false;column:account_expired"`
	AccountLocked      bool      `gorm:"default:false;column:account_locked"`
	CredentialsExpired bool      `gorm:"default:false;column:credentials_expired"`
	LastLogin          time.Time `gorm:"column:last_login"`
	IsMobileVerified   bool      `gorm:"default:false;column:is_mobile_verified"`
	CreatedAt          time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime;column:updated_at"`

	// Transient fields (not stored in DB)
	Email        string `gorm:"-" json:"email,omitempty"`
	FirstName    string `gorm:"-" json:"first_name,omitempty"`
	LastName     string `gorm:"-" json:"last_name,omitempty"`
	PatientId    uint   `gorm:"-" json:"patient_id,omitempty"`
	Relationship string `gorm:"-" json:"relationship,omitempty"`
	Role         string `gorm:"-" json:"role,omitempty"`

	DateOfBirth        string `gorm:"-" json:"date_of_birth,omitempty"`
	Gender             string `gorm:"-" json:"gender,omitempty"`
	MobileNo           string `gorm:"-" json:"mobile_no,omitempty"`
	Address            string `gorm:"-" json:"address,omitempty"`
	EmergencyContact   string `gorm:"-" json:"emergency_contact,omitempty"`
	AbhaNumber         string `gorm:"-" json:"abha_number,omitempty"`
	BloodGroup         string `gorm:"-" json:"blood_group,omitempty"`
	Nationality        string `gorm:"-" json:"nationality,omitempty"`
	CitizenshipStatus  string `gorm:"-" json:"citizenship_status,omitempty"`
	PassportNumber     string `gorm:"-" json:"passport_number,omitempty"`
	CountryOfResidence string `gorm:"-" json:"country_of_residence,omitempty"`
	IsIndianOrigin     bool   `gorm:"-" json:"is_indian_origin,omitempty"`
}

// TableName sets the table name explicitly
func (SystemUser) TableName() string {
	return "tbl_system_user"
}

type RoleMaster struct {
	RoleId    uint      `gorm:"primaryKey;column:role_id" json:"role_id"`
	RoleName  string    `gorm:"size:50;not null;column:role_name" json:"role_name"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
}

// TableName sets the table name explicitly
func (RoleMaster) TableName() string {
	return "tbl_role_master"
}

type UserResponse struct {
	UserId     int    `json:"user_id" gorm:"primaryKey"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Role       string `json:"role"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	AuthUserId string `json:"auth_user_id"`
}
