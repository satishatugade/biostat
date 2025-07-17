package models

import "time"

type Hospital struct {
	HospitalId   uint64  `gorm:"column:hospital_id;primaryKey;autoIncrement" json:"hospital_id"`
	HospitalName string  `gorm:"column:hospital_name" json:"hospital_name"`
	Address      string  `gorm:"column:address" json:"address"`
	State        string  `gorm:"column:state" json:"state"`
	City         string  `gorm:"column:city" json:"city"`
	PostalCode   string  `gorm:"column:postal_code" json:"postal_code"`
	Latitude     *string `gorm:"column:latitude" json:"latitude"`
	Longitude    *string `gorm:"column:longitude" json:"longitude"`
	PhoneNumber  string  `gorm:"column:phone_number" json:"phone_number"`
	Rating       *string `gorm:"column:rating" json:"rating"`
	TotalReviews *string `gorm:"column:total_reviews" json:"total_reviews"`
	IsVerified   *string `gorm:"column:is_verified" json:"is_verified"`
	IsOpenNow    *string `gorm:"column:is_open_now" json:"is_open_now"`
	WebsiteURL   string  `gorm:"column:website_url" json:"website_url"`
	// DiscountOffer  string    `gorm:"column:discount_offer" json:"discount_offer"`
	// PaidStatus     bool      `gorm:"column:paid_status" json:"paid_status"`
	// OperationHours string    `gorm:"column:operation_hours" json:"operation_hour"`
	IsDeleted int       `gorm:"column:is_deleted;default:0" json:"is_deleted"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedBy string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy string    `gorm:"column:updated_by" json:"updated_by"`
}

func (Hospital) TableName() string {
	return "tbl_hospital"
}

type HospitalAudit struct {
	HospitalAuditId uint64    `gorm:"column:hospital_audit_id;primaryKey;autoIncrement" json:"hospital_audit_id"`
	HospitalId      uint64    `gorm:"column:hospital_id" json:"hospital_id"`
	HospitalName    string    `gorm:"column:hospital_name" json:"hospital_name"`
	Address         string    `gorm:"column:address" json:"address"`
	State           string    `gorm:"column:state" json:"state"`
	City            string    `gorm:"column:city" json:"city"`
	PostalCode      string    `gorm:"column:postal_code" json:"postal_code"`
	Latitude        *string   `gorm:"column:latitude" json:"latitude"`
	Longitude       *string   `gorm:"column:longitude" json:"longitude"`
	PhoneNumber     string    `gorm:"column:phone_number" json:"phone_number"`
	Rating          *string   `gorm:"column:rating" json:"rating"`
	TotalReviews    *string   `gorm:"column:total_reviews" json:"total_reviews"`
	IsVerified      *string   `gorm:"column:is_verified" json:"is_verified"`
	IsOpenNow       *string   `gorm:"column:is_open_now" json:"is_open_now"`
	WebsiteURL      string    `gorm:"column:website_url" json:"website_url"`
	IsDeleted       int       `gorm:"column:is_deleted" json:"is_deleted"`
	OperationType   string    `gorm:"column:operation_type" json:"operation_type"`
	CreatedBy       string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy       string    `gorm:"column:updated_by" json:"updated_by"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (HospitalAudit) TableName() string {
	return "tbl_hospital_audit"
}

type Service struct {
	ServiceId   uint64    `gorm:"column:service_id;primaryKey;autoIncrement" json:"service_id"`
	ServiceType string    `gorm:"column:service_type" json:"service_type"`
	ServiceName string    `gorm:"column:service_name" json:"service_name"`
	IsDeleted   int       `gorm:"column:is_deleted;default:0" json:"is_deleted"`
	BulkUpload  int       `gorm:"column:bulk_upload;default:0" json:"bulk_upload"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"-"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"-"`
	CreatedBy   string    `gorm:"column:created_by" json:"-"`
	UpdatedBy   string    `gorm:"column:updated_by" json:"-"`
}

func (Service) TableName() string {
	return "tbl_service"
}

type ServiceMapping struct {
	ServiceProviderId uint64    `gorm:"column:service_provider_id;primaryKey" json:"service_provider_id"`
	ServiceId         int       `gorm:"column:service_id;primaryKey" json:"service_id"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedBy         string    `gorm:"column:created_by" json:"created_by"`
	UpdatedBy         string    `gorm:"column:updated_by" json:"updated_by"`
}

func (ServiceMapping) TableName() string {
	return "tbl_service_mapping"
}

type ServiceAudit struct {
	ServiceAuditId uint64    `gorm:"column:service_audit_id;primaryKey;autoIncrement" json:"service_audit_id"` // Unique ID for the audit record
	ServiceId      uint64    `gorm:"column:service_id" json:"service_id"`                                      // Service ID being audited
	ServiceName    string    `gorm:"column:service_name" json:"service_name"`
	IsDeleted      int       `gorm:"column:is_deleted" json:"is_deleted"`
	OperationType  string    `gorm:"column:operation_type" json:"operation_type"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	CreatedBy      string    `gorm:"column:created_by" json:"created_by"` // User who performed the operation	// Operation type (e.g., "CREATE", "UPDATE", "DELETE")
	UpdatedBy      string    `gorm:"column:updated_by" json:"updated_by"` // User who performed the operation
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"` // Timestamp of the operation
}

func (ServiceAudit) TableName() string {
	return "tbl_service_audit"
}
