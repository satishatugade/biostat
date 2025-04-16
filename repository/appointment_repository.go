package repository

import (
	"biostat/models"
	"errors"
	"log"

	"gorm.io/gorm"
)

type AppointmentRepository interface {
	CreateAppointment(tx *gorm.DB, appointment *models.Appointment) (*models.Appointment, error)
	GetUserAppointments(user_id uint64) ([]models.Appointment, error)
}

type AppointmentRepositoryImpl struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &AppointmentRepositoryImpl{db: db}
}

func (r *AppointmentRepositoryImpl) CreateAppointment(tx *gorm.DB, appointment *models.Appointment) (*models.Appointment, error) {
	log.Println("Creating appointment in Repo:", appointment)
	if tx == nil {
		log.Println("TX is nil")
	}
	if appointment == nil {
		return nil, errors.New("appointment is nil")
	}

	log.Println("Creating appointment in Repo:", appointment)

	if err := tx.Create(appointment).Error; err != nil {
		return nil, err
	}
	return appointment, nil
}

func (r *AppointmentRepositoryImpl) GetUserAppointments(user_id uint64) ([]models.Appointment, error) {
	var appointments []models.Appointment
	if err := r.db.Where("patient_id = ?", user_id).Find(&appointments).Error; err != nil {
		return nil, err
	}
	return appointments, nil
}
