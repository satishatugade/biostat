package models

import "time"

type Nurse struct {
	NurseId           uint64        `json:"nurse_id"`
	FirstName         string        `json:"first_name"`
	LastName          string        `json:"last_name"`
	Speciality        string        `json:"speciality"`
	Gender            string        `json:"gender"`
	GenderId          uint64        `json:"gender_id"`
	MobileNo          string        `json:"mobile_no"`
	LicenseNumber     string        `json:"license_number"`
	ClinicName        string        `json:"clinic_name"`
	ClinicAddress     string        `json:"clinic_address"`
	UserAddress       AddressMaster `gorm:"-" json:"user_address"`
	Email             string        `json:"email"`
	YearsOfExperience int           `json:"years_of_experience"`
	ConsultationFee   float64       `json:"consultation_fee"`
	WorkingHours      string        `json:"working_hours"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

type Pharmacist struct {
	PharmacistId      uint64        `json:"pharmacist_id"`
	FirstName         string        `json:"first_name"`
	LastName          string        `json:"last_name"`
	Gender            string        `json:"gender"`
	GenderId          uint64        `json:"gender_id"`
	MobileNo          string        `json:"mobile_no"`
	LicenseNumber     string        `json:"license_number"`
	PharmacyName      string        `json:"pharmacy_name"`
	PharmacyAddress   string        `json:"pharmacy_address"`
	UserAddress       AddressMaster `gorm:"-" json:"user_address"`
	Email             string        `json:"email"`
	Speciality        string        `json:"speciality"`
	YearsOfExperience int           `json:"years_of_experience"`
	ConsultationFee   float64       `json:"consultation_fee"`
	WorkingHours      string        `json:"working_hours"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}
