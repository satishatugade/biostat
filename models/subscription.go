package models

import (
	"time"
)

type SubscriptionMaster struct {
	SubscriptionId uint64    `gorm:"primaryKey;column:subscription_id" json:"subscription_id"`
	PlanName       string    `gorm:"column:plan_name" json:"plan_name"`
	PlanType       string    `gorm:"column:plan_type;type:varchar(50)" json:"plan_type"`
	Price          float64   `gorm:"column:price" json:"price"`
	MaxMember      int64     `gorm:"column:max_member" json:"max_member"`
	Duration       int       `gorm:"column:duration" json:"duration"`
	VisibleToUser  bool      `gorm:"column:visible_to_user;default:true" json:"visible_to_user"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedBy      string    `gorm:"column:created_by" json:"-"`
	UpdatedBy      string    `gorm:"column:updated_by" json:"-"`

	// PatientFamilyGroup fields
	FamilyId              uint64     `gorm:"-" json:"family_id,omitempty"`
	FamilyName            string     `gorm:"-" json:"family_name,omitempty"`
	MemberId              uint64     `gorm:"-" json:"member_id,omitempty"`
	CurrentSubscriptionId *uint64    `gorm:"-" json:"current_subscription_id,omitempty"`
	SubscriptionStartDate *string    `gorm:"-" json:"subscription_start_date,omitempty"`
	SubscriptionEndDate   *string    `gorm:"-" json:"subscription_end_date,omitempty"`
	IsActive              bool       `gorm:"-" json:"is_active,omitempty"`
	IsAutoRenew           bool       `gorm:"-" json:"is_auto_renew,omitempty"`
	LastRenewedAt         *time.Time `gorm:"-" json:"last_renewed_at,omitempty"`
	LastRenewedBy         uint64     `gorm:"-" json:"last_renewed_by,omitempty"`
	LastRenewalType       string     `gorm:"-" json:"last_renewal_type,omitempty"`

	ServiceMappings []SubscriptionServiceMapping `gorm:"foreignKey:SubscriptionId;references:SubscriptionId" json:"services"`
}

func (SubscriptionMaster) TableName() string {
	return "tbl_subscription_master"
}

type SubscriptionMasterAudit struct {
	SubscriptionAuditId uint64    `gorm:"primaryKey;column:subscription_audit_id" json:"subscription_audit_id"`
	SubscriptionId      uint64    `gorm:"column:subscription_id" json:"subscription_id"`
	PlanName            string    `gorm:"column:plan_name" json:"plan_name"`
	PlanType            string    `gorm:"column:plan_type;type:varchar(50)" json:"plan_type"`
	Price               float64   `gorm:"column:price" json:"price"`
	MaxMember           int       `gorm:"column:max_member" json:"max_member"`
	Duration            int       `gorm:"column:duration" json:"duration"`
	IsActive            bool      `gorm:"column:is_active;default:true" json:"is_active"`
	VisibleToUser       bool      `gorm:"column:visible_to_user;default:true" json:"visible_to_user"`
	ActionType          string    `gorm:"column:action_type" json:"action_type"`
	ActionBy            *uint64   `gorm:"column:action_by" json:"action_by,omitempty"`
	ActionAt            time.Time `gorm:"column:action_at" json:"action_at"`
	CreatedBy           string    `gorm:"column:created_by" json:"created_by,omitempty"`
	UpdatedBy           string    `gorm:"column:updated_by" json:"updated_by,omitempty"`
}

func (SubscriptionMasterAudit) TableName() string {
	return "tbl_subscription_master_audit"
}

type PatientFamilyGroup struct {
	FamilyId              uint64     `gorm:"primaryKey;column:family_id" json:"family_id"`
	FamilyName            string     `gorm:"column:family_name" json:"family_name"`
	MemberId              uint64     `gorm:"column:member_id" json:"member_id"`
	CurrentSubscriptionId *uint64    `gorm:"column:current_subscription_id" json:"current_subscription_id,omitempty"`
	SubscriptionStartDate *time.Time `gorm:"column:subscription_start_date" json:"subscription_start_date,omitempty"`
	SubscriptionEndDate   *time.Time `gorm:"column:subscription_end_date" json:"subscription_end_date,omitempty"`
	IsActive              bool       `gorm:"column:is_active" json:"is_active"`
	IsAutoRenew           bool       `gorm:"column:is_auto_renew" json:"is_auto_renew"`
	LastRenewedAt         *time.Time `gorm:"column:last_renewed_at" json:"last_renewed_at,omitempty"`
	LastRenewedBy         uint64     `gorm:"column:last_renewed_by" json:"last_renewed_by,omitempty"`
	LastRenewalType       string     `gorm:"column:last_renewal_type" json:"last_renewal_type,omitempty"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (PatientFamilyGroup) TableName() string {
	return "tbl_patient_family_group"
}

type PatientFamilyGroupAudit struct {
	PatientFamilyGroupAuditId uint64     `gorm:"primaryKey;column:fg_audit_id" json:"fg_audit_id"`
	FamilyId                  uint64     `gorm:"column:family_id" json:"family_id"`
	FamilyName                string     `gorm:"column:family_name" json:"family_name"`
	MemberId                  uint64     `gorm:"column:member_id" json:"member_id"`
	CurrentSubscriptionId     *uint64    `gorm:"column:current_subscription_id" json:"current_subscription_id,omitempty"`
	SubscriptionStartDate     *time.Time `gorm:"column:subscription_start_date" json:"subscription_start_date,omitempty"`
	SubscriptionEndDate       *time.Time `gorm:"column:subscription_end_date" json:"subscription_end_date,omitempty"`

	IsAutoRenew     *bool      `gorm:"column:is_auto_renew" json:"is_auto_renew,omitempty"`
	LastRenewedAt   *time.Time `gorm:"column:last_renewed_at" json:"last_renewed_at,omitempty"`
	LastRenewedBy   *string    `gorm:"column:last_renewed_by" json:"last_renewed_by,omitempty"`
	LastRenewalType string     `gorm:"column:last_renewal_type" json:"last_renewal_type,omitempty"`

	ActionType string    `gorm:"column:action_type" json:"action_type"`
	ActionAt   time.Time `gorm:"column:action_at" json:"action_at"`
	ActionBy   *uint64   `gorm:"column:action_by" json:"action_by,omitempty"`
}

func (PatientFamilyGroupAudit) TableName() string {
	return "tbl_patient_family_group_audit"
}

type SubscribeFamilyRequest struct {
	FamilyId       uint64 `json:"family_id"`
	SubscriptionId uint64 `json:"subscription_id" binding:"required"`
	IsAutoRenew    bool   `json:"is_auto_renew"`
	RenewalType    string `json:"renewal_type"`
}

type SystemSetting struct {
	SettingKey   string    `gorm:"column:setting_key;primaryKey" json:"setting_key"`
	SettingValue bool      `gorm:"column:setting_value" json:"setting_value"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	UpdatedBy    string    `gorm:"column:updated_by;type:varchar(255)" json:"updated_by"`
}

func (SystemSetting) TableName() string {
	return "tbl_system_setting"
}

type SubscriptionServiceMapping struct {
	MappingId      uint64    `gorm:"primaryKey;column:mapping_id" json:"-"`
	SubscriptionId uint64    `gorm:"column:subscription_id" json:"-"`
	ServiceId      uint64    `gorm:"column:service_id" json:"-"`
	UsageLimit     *int      `gorm:"column:usage_limit" json:"usage_limit"`
	UsagePeriod    string    `gorm:"column:usage_period" json:"usage_period"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"-"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"-"`
	CreatedBy      string    `gorm:"column:created_by" json:"-"`
	UpdatedBy      string    `gorm:"column:updated_by" json:"-"`
	Service        Service   `gorm:"foreignKey:ServiceId;references:ServiceId" json:"service"`
}

func (SubscriptionServiceMapping) TableName() string {
	return "tbl_subscription_service_mapping"
}
