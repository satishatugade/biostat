package models

import "time"

type PatientDietPlan struct {
	PatientDietPlanId  uint       `gorm:"primaryKey;autoIncrement" json:"patient_diet_plan_id"`
	PatientId          uint       `json:"patient_id"`
	DietPlanTemplateId uint       `json:"diet_plan_template_id"`
	StartDate          time.Time  `json:"start_date"`
	EndDate            time.Time  `json:"end_date"`
	Customizations     string     `json:"customizations"`
	PaymentStatus      string     `json:"payment_status"`
	PaymentDate        *time.Time `json:"payment_date,omitempty"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`

	// Relations
	DietCreator      DietCreator      `json:"diet_creator" gorm:"foreignKey:DietCreatorId"`
	DietPlanTemplate DietPlanTemplate `gorm:"foreignKey:DietPlanTemplateId" json:"diet_plan_template"`
}

type DietCreator struct {
	DietCreatorId  uint      `json:"diet_creator_id" gorm:"primaryKey"`
	CreatorName    string    `json:"creator_name"`
	Email          string    `json:"email"`
	Specialization string    `json:"specialization"`
	CreatedAt      time.Time `json:"created_at"`
}

func (DietCreator) TableName() string {
	return "tbl_diet_creator"
}

func (PatientDietPlan) TableName() string {
	return "tbl_patient_diet_plan"
}
