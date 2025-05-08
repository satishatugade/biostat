package repository

import (
	"biostat/models"
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
)

type AppointmentRepository interface {
	CreateAppointment(tx *gorm.DB, appointment *models.Appointment) (*models.Appointment, error)
	GetUserAppointments(user_id uint64) ([]models.Appointment, error)
	FindAppointmentByID(tx *gorm.DB, appointmentID uint64) (*models.Appointment, error)
	UpdateAppointment(tx *gorm.DB, appointment *models.Appointment) error
	CreateAuditSnapshot(tx *gorm.DB, appointment *models.Appointment, action string, changedBy uint64) error
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

func (r *AppointmentRepositoryImpl) FindAppointmentByID(tx *gorm.DB, appointmentID uint64) (*models.Appointment, error) {
	var appointment models.Appointment
	if err := tx.First(&appointment, "appointment_id = ?", appointmentID).Error; err != nil {
		return nil, err
	}
	return &appointment, nil
}

func (r *AppointmentRepositoryImpl) UpdateAppointment(tx *gorm.DB, appointment *models.Appointment) error {
	if tx == nil {
		return errors.New("nil DB transaction")
	}
	return tx.Model(&models.Appointment{}).
		Where("appointment_id = ?", appointment.AppointmentID).
		Updates(appointment).Error
}

func (r *AppointmentRepositoryImpl) CreateAuditSnapshot(tx *gorm.DB, appointment *models.Appointment, action string, changedBy uint64) error {
	audit := &models.AppointmentAudit{
		Action:          action,
		ChangedBy:       changedBy,
		ChangeTimestamp: time.Now(),

		AppointmentID:   appointment.AppointmentID,
		PatientID:       appointment.PatientID,
		ProviderID:      appointment.ProviderID,
		ProviderType:    appointment.ProviderType,
		ScheduledBy:     appointment.ScheduledBy,
		AppointmentType: appointment.AppointmentType,
		AppointmentDate: appointment.AppointmentDate,
		AppointmentTime: appointment.AppointmentTime,
		DurationMinutes: appointment.DurationMinutes,
		IsInperson:      appointment.IsInperson,
		MeetingUrl:      appointment.MeetingUrl,
		Status:          appointment.Status,
		PaymentStatus:   appointment.PaymentStatus,
		PaymentID:       appointment.PaymentID,
		Notes:           appointment.Notes,
		CreatedAt:       appointment.CreatedAt,
		UpdatedAt:       appointment.UpdatedAt,
		IsDeleted:       appointment.IsDeleted,
	}
	return tx.Create(audit).Error
}
