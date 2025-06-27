package models

import (
	"time"

	"github.com/google/uuid"
)

type Admin struct {
	UserId    uint64 `json:"user_id"`
	RoleName  string `json:"role_name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	MobileNo  string `json:"mobile_no"`
	Email     string `json:"email"`
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
	RelationId              uint64    `gorm:"column:relation_id;not null" json:"relation_id"`
	IsSelf                  bool      `gorm:"column:is_self;default:false" json:"is_self"`
	MappingType             string    `gorm:"column:mapping_type;type:varchar(50)" json:"mapping_type,omitempty"`
	IsDeleted               int       `gorm:"column:is_deleted;default:0" json:"is_deleted"`
	CreatedAt               time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt               time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (SystemUserRoleMapping) TableName() string {
	return "tbl_system_user_role_mapping"
}

type ReverseRelationMappingResponse struct {
	RelationId              uint64 `json:"relation_id"`
	SourceGenderId          uint64 `json:"source_gender_id"`
	ReverseRelationGenderId uint64 `json:"reverse_relation_gender_id"`
	ReverseRelationshipId   uint64 `json:"reverse_relationship_id"`
	ReverseRelationshipName string `json:"reverse_relationship_name"`
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
	ExpiresIN    int          `json:"expires_in"`
	UserResponse UserResponse `json:"user_data"`
}

type SystemUser_ struct {
	UserId        uint64        `gorm:"primaryKey;column:user_id" json:"user_id"`
	AuthUserId    string        `gorm:"column:auth_user_id;type:varchar(100);unique" json:"auth_user_id"`
	Username      string        `gorm:"column:username;type:varchar(50);unique;not null" json:"username"`
	Password      string        `gorm:"column:password;type:varchar(255);not null" json:"password"`
	FirstName     string        `gorm:"column:first_name;type:varchar(50);not null" json:"first_name"`
	MiddleName    string        `gorm:"column:middle_name;type:varchar(50)" json:"middle_name"`
	LastName      string        `gorm:"column:last_name;type:varchar(50);not null" json:"last_name"`
	Gender        string        `gorm:"-" json:"gender"`
	GenderId      uint64        `gorm:"column:gender_id" json:"gender_id"`
	DateOfBirth   *time.Time    `gorm:"column:date_of_birth;type:date" json:"date_of_birth"`
	MobileNo      string        `gorm:"column:mobile_no;type:varchar(50);unique" json:"mobile_no"`
	Email         string        `gorm:"column:email;type:varchar(100);unique" json:"email"`
	MaritalStatus string        `gorm:"column:marital_status;type:varchar(100)" json:"marital_status"`
	Address       string        `gorm:"column:address;type:text" json:"address"`
	UserAddress   AddressMaster `gorm:"-" json:"user_address"`

	// Patient-Specific Fields
	FormatDateTimePattern string `gorm:"column:format_datetime_pattern;type:text" json:"format_datetime_pattern"`
	EmergencyContact      string `gorm:"column:emergency_contact;type:varchar(50)" json:"emergency_contact,omitempty"`
	EmergencyContactName  string `gorm:"-" json:"emergency_contact_person"`
	AbhaNumber            string `gorm:"column:abha_number;type:varchar(50)" json:"abha_number,omitempty"`
	BloodGroup            string `gorm:"column:blood_group;type:varchar(10)" json:"blood_group,omitempty"`
	Nationality           string `gorm:"column:nationality;type:varchar(50)" json:"nationality,omitempty"`
	CitizenshipStatus     string `gorm:"column:citizenship_status;type:varchar(50)" json:"citizenship_status,omitempty"`
	PassportNumber        string `gorm:"column:passport_number;type:varchar(50);unique" json:"passport_number,omitempty"`
	CountryOfResidence    string `gorm:"column:country_of_residence;type:varchar(50)" json:"country_of_residence,omitempty"`
	IsIndianOrigin        bool   `gorm:"column:is_indian_origin;default:false" json:"is_indian_origin,omitempty"`

	// Doctor-Specific Fields
	Speciality        string   `gorm:"column:speciality;type:varchar(100)" json:"speciality,omitempty"`
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
	LoginCount         int        `gorm:"column:login_count" json:"login_count"`
	LastLoginIP        string     `gorm:"column:last_login_ip" json:"last_login_ip"`
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

	AddressMapping SystemUserAddressMapping `gorm:"foreignKey:UserId;references:UserId" json:"address_mappings"`
}

func (SystemUser_) TableName() string {
	return "tbl_system_user_"
}

type UserLoginInfo struct {
	AuthUserID string
	Username   string
	Password   string
	LoginCount int
}

func (AddressMaster) TableName() string {
	return "tbl_address_master"
}

type AddressMaster struct {
	AddressId    uint64     `gorm:"column:address_id;primaryKey;autoIncrement" json:"address_id"`
	AddressLine1 string     `gorm:"column:address_line1" json:"address_line1"`
	AddressLine2 string     `gorm:"column:address_line2" json:"address_line2"`
	Landmark     string     `gorm:"column:landmark" json:"landmark"`
	City         string     `gorm:"column:city" json:"city"`
	State        string     `gorm:"column:state" json:"state"`
	Country      string     `gorm:"column:country" json:"country"`
	PostalCode   string     `gorm:"column:postal_code" json:"postal_code"`
	Latitude     float64    `gorm:"column:latitude" json:"latitude"`
	Longitude    float64    `gorm:"column:longitude" json:"longitude"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

type SystemUserAddressMapping struct {
	UserId    uint64        `gorm:"column:user_id" json:"user_id"`
	AddressId uint64        `gorm:"column:address_id" json:"address_id"`
	CreatedAt time.Time     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt *time.Time    `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	Address   AddressMaster `gorm:"foreignKey:AddressId;references:AddressId" json:"address"`
}

func (SystemUserAddressMapping) TableName() string {
	return "tbl_system_user_address_mapping"
}

type SupportGroup struct {
	SupportGroupId uint64    `gorm:"column:support_group_id;primaryKey;autoIncrement" json:"support_group_id"`
	GroupName      string    `gorm:"column:group_name" json:"group_name"`
	Description    string    `gorm:"column:description" json:"description"`
	Location       string    `gorm:"column:location" json:"location"`
	IsDeleted      int       `gorm:"column:is_deleted" json:"is_deleted"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedBy      string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy      string    `gorm:"column:updated_by" json:"updated_by"`
}

func (SupportGroup) TableName() string {
	return "tbl_support_group"
}

type SupportGroupAudit struct {
	SupportGrpAuditId uint64    `gorm:"column:support_group_audit_id;primaryKey;autoIncrement" json:"support_group_audit_id"`
	SupportGroupId    uint64    `gorm:"column:support_group_id" json:"support_group_id"`
	GroupName         string    `gorm:"column:group_name" json:"group_name"`
	Description       string    `gorm:"column:description" json:"description"`
	Location          string    `gorm:"column:location" json:"location"`
	IsDeleted         int       `gorm:"column:is_deleted" json:"is_deleted"`
	OperationType     string    `gorm:"column:operation_type" json:"operation_type"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedBy         string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy         string    `gorm:"column:updated_by" json:"updated_by"`
}

func (SupportGroupAudit) TableName() string {
	return "tbl_support_group_audit"
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

type CheckUserMobileEmail struct {
	Mobile string `json:"mobile" binding:"omitempty"`
	Email  string `json:"email" binding:"omitempty,email"`
}

type PincodeMaster struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Pincode  string `json:"pincode"`
	City     string `json:"city"`
	District string `json:"district"`
	State    string `json:"state"`
	Country  string `json:"country"`
}

func (PincodeMaster) TableName() string {
	return "tbl_pincode_master"
}

type SendOTPRequest struct {
	Email string `json:"email" binding:"required"`
}

type NotificationUserMapping struct {
	ID             uuid.UUID `db:"id" json:"id"`
	UserID         int64     `db:"user_id" json:"user_id"`
	NotificationID uuid.UUID `db:"notification_id" json:"notification_id"`
	SourceType     string    `db:"source_type" json:"source_type"`
	SourceID       string    `db:"source_id" json:"source_id"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type PermissionMaster struct {
	PermissionID int64     `gorm:"primaryKey;column:permission_id" json:"permission_id"`
	Code         string    `gorm:"uniqueIndex;column:code" json:"code"`
	Name         string    `gorm:"column:name" json:"name"`
	Description  string    `gorm:"column:description" json:"description"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (PermissionMaster) TableName() string {
	return "tbl_permissions_master"
}

type UserRelativePermissionMapping struct {
	MappingID    uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"mapping_id"`
	UserID       uint64    `gorm:"column:user_id" json:"user_id"`
	RelativeID   uint64    `gorm:"column:relative_id" json:"relative_id"`
	PermissionID int64     `gorm:"column:permission_id" json:"permission_id"`
	Granted      bool      `gorm:"column:granted" json:"granted"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (UserRelativePermissionMapping) TableName() string {
	return "tbl_user_relative_permission_mappings"
}

type PermissionResult struct {
	UserID     uint64 `json:"user_id"`
	RelativeID uint64 `json:"relative_id"`
	Code       string `json:"code"`
	Granted    bool   `json:"granted"`
}

type AddRelationRequest struct {
	UserID      uint64 `json:"user_id" binding:"required"`
	CurrentRole string `json:"current_role" binding:"required"`
	RoleName    string `json:"role_name" binding:"required"`
}

type UserShare struct {
	UserID       uint64 `json:"user_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	MappingType  string `json:"mapping_type"`
	RoleName     string `json:"role_name"`
	Relationship string `json:"relationship"`
	Email        string `json:"email"`
	MobileNo     string `json:"mobile_no"`
}

type UserAddressResponse struct {
	UserId       uint64 `json:"user_id"`
	MappingType  string `json:"mapping_type"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	AddressId    uint64 `json:"address_id"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	PostalCode   string `json:"postal_code"`
}
