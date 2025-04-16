package service

import (
	"biostat/models"
	"biostat/repository"
	"log"

	"gorm.io/gorm"
)

type AppointmentService interface {
	CreateAppointment(tx *gorm.DB, appointment *models.Appointment) (*models.Appointment, error)
	GetUserAppointments(user_id uint64)([]models.Appointment, error)
}
type AppointmentServiceImpl struct {
	appointmentRepo repository.AppointmentRepository
}

func NewAppointmentService(repo repository.AppointmentRepository) AppointmentService {
	return &AppointmentServiceImpl{appointmentRepo: repo}
}

func (s *AppointmentServiceImpl) CreateAppointment(tx *gorm.DB, appointment *models.Appointment) (*models.Appointment, error) {
	log.Println("inside appointment in Service:", appointment)
	if tx == nil {
		log.Println("TX is nil in services")
	}
	return s.appointmentRepo.CreateAppointment(tx, appointment)
}

func (s *AppointmentServiceImpl) GetUserAppointments(user_id uint64)([]models.Appointment, error){
	return s.appointmentRepo.GetUserAppointments(user_id)
}