package models

import "time"

type Caregiver struct {
	PatientId    uint64             `json:"user_id"` //patient_id
	FirstName    string             `json:"first_name"`
	MiddleName   string             `json:"middle_name"`
	LastName     string             `json:"last_name"`
	ContactInfo  string             `json:"contact_info"`
	Gender       string             `json:"gender"`
	GenderId     uint64             `json:"gender_id"`
	DateOfBirth  time.Time          `json:"date_of_birth"`
	MobileNo     string             `json:"mobile_no"`
	Email        string             `json:"email"`
	MappingType  string             `json:"mapping_type"`
	RelationId   uint64             `json:"relation_id"`
	Relationship string             `json:"relationship"`
	Address      string             `json:"address"`
	UserAddress  AddressMaster      `gorm:"-" json:"user_address"`
	CreatedAt    time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time          `json:"updated_at" gorm:"autoUpdateTime"`
	Permissions  []PermissionResult `json:"permissions" gorm:"-"`
	HealthScore  int                `json:"health_score" gorm:"-"`
}

type UserRelation struct {
	UserId      uint64 `json:"user_id"`
	PatientId   uint64 `json:"patient_id"`
	RelationId  uint64 `json:"relation_id"`
	MappingType string `json:"mapping_type"`
}
