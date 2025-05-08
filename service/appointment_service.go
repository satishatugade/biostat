package service

import (
	"biostat/models"
	"biostat/repository"
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
)

type AppointmentService interface {
	CreateAppointment(tx *gorm.DB, appointment *models.Appointment) (*models.Appointment, error)
	GetUserAppointments(user_id uint64) ([]models.Appointment, error)
	FindAppointmentByID(tx *gorm.DB, appointmentID uint64) (*models.Appointment, error)
	UpdateAppointmentByType(tx *gorm.DB, appointment *models.Appointment, updateType int, changedBy uint64) (*models.Appointment, error)
}
type AppointmentServiceImpl struct {
	appointmentRepo repository.AppointmentRepository
}

func NewAppointmentService(repo repository.AppointmentRepository) AppointmentService {
	return &AppointmentServiceImpl{appointmentRepo: repo}
}

func (s *AppointmentServiceImpl) CreateAppointment(tx *gorm.DB, appointment *models.Appointment) (*models.Appointment, error) {
	if tx == nil {
		log.Println("TX is nil in services")
	}
	return s.appointmentRepo.CreateAppointment(tx, appointment)
}

func (s *AppointmentServiceImpl) GetUserAppointments(user_id uint64) ([]models.Appointment, error) {
	return s.appointmentRepo.GetUserAppointments(user_id)
}

func (s *AppointmentServiceImpl) FindAppointmentByID(tx *gorm.DB, appointmentID uint64) (*models.Appointment, error) {
	return s.appointmentRepo.FindAppointmentByID(tx, appointmentID)
}

func (s *AppointmentServiceImpl) UpdateAppointmentByType(tx *gorm.DB, appointment *models.Appointment, updateType int, changedBy uint64) (*models.Appointment, error) {
	switch updateType {
	case 1:
		appointment.Status = "rescheduled"
		appointment.UpdatedAt = time.Now()
		if err := s.appointmentRepo.UpdateAppointment(tx, appointment); err != nil {
			return nil, err
		}
		if err := s.appointmentRepo.CreateAuditSnapshot(tx, appointment, "rescheduled", changedBy); err != nil {
			log.Println("Audit log error (reschedule):", err)
		}
	case 2:
		appointment.Status = "cancelled"
		appointment.UpdatedAt = time.Now()

		if err := s.appointmentRepo.UpdateAppointment(tx, appointment); err != nil {
			return nil, err
		}
		if err := s.appointmentRepo.CreateAuditSnapshot(tx, appointment, "cancelled", changedBy); err != nil {
			log.Println("Audit log error (cancel):", err)
		}
	default:
		return nil, errors.New("invalid update type")
	}
	return appointment, nil
}
