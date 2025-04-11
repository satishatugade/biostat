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
	RoleId    uint64    `gorm:"primaryKey;column:role_id" json:"role_id"`
	RoleName  string    `gorm:"size:50;not null;column:role_name" json:"role_name"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
}

// TableName sets the table name explicitly
func (RoleMaster) TableName() string {
	return "tbl_role_master"
}

type SystemUserRoleMapping struct {
	SystemUserRoleMappingId uint64    `gorm:"primaryKey;column:system_user_role_mapping_id" json:"system_user_role_mapping_id"`
	UserId                  uint64    `gorm:"column:user_id;not null" json:"user_id"`
	PatientId               uint64    `gorm:"column:patient_id;not null" json:"patient_id"`
	RoleId                  uint64    `gorm:"column:role_id;not null" json:"role_id"`
	RelationId              int       `gorm:"column:relation_id;not null" json:"relation_id"`
	IsSelf                  bool      `gorm:"column:is_self;default:false" json:"is_self"`
	MappingType             string    `gorm:"column:mapping_type;type:varchar(50)" json:"mapping_type,omitempty"`
	CreatedAt               time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt               time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (SystemUserRoleMapping) TableName() string {
	return "tbl_system_user_role_mapping"
}

type UserResponse struct {
	UserId     uint64 `json:"user_id" gorm:"primaryKey"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Role       string `json:"role"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	AuthUserId string `json:"auth_user_id"`
}

type UserLoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	UserResponse UserResponse `json:"user_data"`
}

type SystemUser_ struct {
	UserId      uint64     `gorm:"primaryKey;column:user_id" json:"user_id"`
	AuthUserId  string     `gorm:"column:auth_user_id;type:varchar(100);unique" json:"auth_user_id"`
	Username    string     `gorm:"column:username;type:varchar(50);unique;not null" json:"username"`
	Password    string     `gorm:"column:password;type:varchar(255);not null" json:"password"`
	FirstName   string     `gorm:"column:first_name;type:varchar(50);not null" json:"first_name"`
	LastName    string     `gorm:"column:last_name;type:varchar(50);not null" json:"last_name"`
	Gender      string     `gorm:"column:gender;type:varchar(10)" json:"gender"`
	DateOfBirth *time.Time `gorm:"column:date_of_birth;type:date" json:"date_of_birth"`
	MobileNo    string     `gorm:"column:mobile_no;type:varchar(50);unique" json:"mobile_no"`
	Email       string     `gorm:"column:email;type:varchar(100);unique" json:"email"`
	Address     string     `gorm:"column:address;type:text" json:"address"`

	// Patient-Specific Fields
	EmergencyContact     string `gorm:"column:emergency_contact;type:varchar(50)" json:"emergency_contact,omitempty"`
	EmergencyContactName string `gorm:"-" json:"emergency_contact_person"`
	AbhaNumber           string `gorm:"column:abha_number;type:varchar(50)" json:"abha_number,omitempty"`
	BloodGroup           string `gorm:"column:blood_group;type:varchar(10)" json:"blood_group,omitempty"`
	Nationality          string `gorm:"column:nationality;type:varchar(50)" json:"nationality,omitempty"`
	CitizenshipStatus    string `gorm:"column:citizenship_status;type:varchar(50)" json:"citizenship_status,omitempty"`
	PassportNumber       string `gorm:"column:passport_number;type:varchar(50);unique" json:"passport_number,omitempty"`
	CountryOfResidence   string `gorm:"column:country_of_residence;type:varchar(50)" json:"country_of_residence,omitempty"`
	IsIndianOrigin       bool   `gorm:"column:is_indian_origin;default:false" json:"is_indian_origin,omitempty"`

	// Doctor-Specific Fields
	Specialty         string   `gorm:"column:specialty;type:varchar(100)" json:"specialty,omitempty"`
	LicenseNumber     string   `gorm:"column:license_number;type:varchar(50);unique" json:"license_number,omitempty"`
	ClinicName        string   `gorm:"column:clinic_name;type:varchar(100)" json:"clinic_name,omitempty"`
	ClinicAddress     string   `gorm:"column:clinic_address;type:text" json:"clinic_address,omitempty"`
	YearsOfExperience *int     `gorm:"column:years_of_experience" json:"years_of_experience,omitempty"`
	ConsultationFee   *float64 `gorm:"column:consultation_fee;type:decimal(10,2)" json:"consultation_fee,omitempty"`
	WorkingHours      string   `gorm:"column:working_hours;type:varchar(100)" json:"working_hours,omitempty"`

	// Authentication & User State
	UserState          string     `gorm:"column:user_state;type:varchar(50)" json:"user_state"`
	AuthType           string     `gorm:"column:auth_type;type:varchar(50)" json:"auth_type"`
	AuthStatus         string     `gorm:"column:auth_status;type:varchar(50);default:'pending'" json:"auth_status"`
	AuthDate           time.Time  `gorm:"column:auth_date;default:CURRENT_TIMESTAMP" json:"auth_date"`
	ActivationFlag     bool       `gorm:"column:activation_flag;default:false" json:"activation_flag"`
	FirstLoginFlag     bool       `gorm:"column:first_login_flag;default:false" json:"first_login_flag"`
	AccountExpired     bool       `gorm:"column:account_expired;default:false" json:"account_expired"`
	AccountLocked      bool       `gorm:"column:account_locked;default:false" json:"account_locked"`
	CredentialsExpired bool       `gorm:"column:credentials_expired;default:false" json:"credentials_expired"`
	LastLogin          *time.Time `gorm:"column:last_login" json:"last_login,omitempty"`
	IsMobileVerified   bool       `gorm:"column:is_mobile_verified;default:false" json:"is_mobile_verified"`

	// Timestamp Fields
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Transient Fields (Not Stored in DB)
	RoleId     uint64 `gorm:"-" json:"role_id"`
	RoleName   string `gorm:"-" json:"role_name"`
	RelationId uint64 `gorm:"-" json:"relation_id"`
}

func (SystemUser_) TableName() string {
	return "tbl_system_user_"
}

type SupportGroup struct {
	SupportGroupId uint64    `gorm:"column:support_group_id;primaryKey;autoIncrement" json:"support_group_id"`
	GroupName      string    `gorm:"column:group_name" json:"group_name"`
	Description    string    `gorm:"column:description" json:"description"`
	Location       string    `gorm:"column:location" json:"location"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedBy      string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy      string    `gorm:"column:updated_by" json:"updated_by"`
}

func (SupportGroup) TableName() string {
	return "tbl_support_group"
}

type SupportGroupMember struct {
	SupportGroupMemberId uint64    `gorm:"column:support_group_member_id;primaryKey;autoIncrement" json:"support_group_member_id"`
	SupportGroupId       uint64    `gorm:"column:support_group_id" json:"support_group_id"`
	UserId               uint64    `gorm:"column:user_id" json:"user_id"`
	Role                 string    `gorm:"column:role" json:"role"`
	Status               string    `gorm:"column:status" json:"status"`
	Notes                string    `gorm:"column:notes" json:"notes"`
	JoinedAt             time.Time `gorm:"column:joined_at" json:"joined_at"`
	CreatedAt            time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedBy            string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy            string    `gorm:"column:updated_by" json:"updated_by"`
}

func (SupportGroupMember) TableName() string {
	return "tbl_support_group_member"
}
