package models

import "time"

type Nurse struct {
	NurseId           uint64    `json:"nurse_id"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Specialty         string    `json:"specialty"`
	Gender            string    `json:"gender"`
	MobileNo          string    `json:"mobile_no"`
	LicenseNumber     string    `json:"license_number"`
	ClinicName        string    `json:"clinic_name"`
	ClinicAddress     string    `json:"clinic_address"`
	Email             string    `json:"email"`
	YearsOfExperience int       `json:"years_of_experience"`
	ConsultationFee   float64   `json:"consultation_fee"`
	WorkingHours      string    `json:"working_hours"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
