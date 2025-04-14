package repository

import (
	"biostat/constant"
	"biostat/models"
	"time"

	"gorm.io/gorm"
)

type HospitalRepository interface {
	AddHospital(hospital *models.Hospital) error
	UpdateHospital(hospital *models.Hospital, updatedBy string) error
	GetAllHospitals() ([]models.Hospital, error)
	DeleteHospitalById(hospitalId int64, updatedBy string) error
	GetHospitalById(hospitalId uint64) (models.Hospital, error)

	AddService(service *models.Service) error
	GetAllServices() ([]models.Service, error)
	GetServiceById(serviceId string) (models.Service, error)
	UpdateService(service *models.Service, updatedBy string) error
	DeleteService(serviceId uint64, deletedBy string) error

	AddServiceMapping(serviceMapping models.ServiceMapping) error

	GetHospitalAuditRecord(hospitalId uint64, hospitalAuditId uint64) ([]models.HospitalAudit, error)
	GetAllHospitalAuditRecord(page, limit int) ([]models.HospitalAudit, int64, error)

	GetServiceAuditRecord(serviceId uint64, serviceAuditId uint64) ([]models.ServiceAudit, error)
	GetAllServiceAuditRecord(page, limit int) ([]models.ServiceAudit, int64, error)
}

type HospitalRepositoryImpl struct {
	db *gorm.DB
}

func (r *HospitalRepositoryImpl) GetHospitalById(hospitalId uint64) (models.Hospital, error) {
	var hospital models.Hospital

	err := r.db.Where("hospital_id = ? ", hospitalId).First(&hospital).Error
	if err != nil {
		return models.Hospital{}, err
	}
	return hospital, nil
}

func NewHospitalRepository(db *gorm.DB) HospitalRepository {
	return &HospitalRepositoryImpl{db: db}
}

func (r *HospitalRepositoryImpl) AddHospital(h *models.Hospital) error {
	return r.db.Create(h).Error
}

func (r *HospitalRepositoryImpl) UpdateHospital(h *models.Hospital, updatedBy string) error {
	tx := r.db.Begin()

	var originalHospital models.Hospital
	if err := tx.Where("hospital_id = ?", h.HospitalId).First(&originalHospital).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.Hospital{}).Where("hospital_id = ?", h.HospitalId).Updates(h).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := r.InsertHospitalAudit(tx, h, constant.UPDATE, updatedBy); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (r *HospitalRepositoryImpl) InsertHospitalAudit(tx *gorm.DB, h *models.Hospital, operation, updatedBy string) error {
	audit := models.HospitalAudit{
		HospitalId:    h.HospitalId,
		OperationType: operation,
		UpdatedBy:     updatedBy,
		UpdatedAt:     time.Now(),
		HospitalName:  h.HospitalName,
		Address:       h.Address,
		Area:          h.Area,
		City:          h.City,
		Pincode:       h.Pincode,
		Latitude:      h.Latitude,
		Longitude:     h.Longitude,
		PhoneNumber:   h.PhoneNumber,
		Rating:        h.Rating,
		TotalReviews:  h.TotalReviews,
		IsVerified:    h.IsVerified,
		IsOpenNow:     h.IsOpenNow,
		WebsiteURL:    h.WebsiteURL,
		IsDeleted:     h.IsDeleted,
	}
	if err := tx.Create(&audit).Error; err != nil {
		return err
	}
	return nil
}

func (r *HospitalRepositoryImpl) GetAllHospitals() ([]models.Hospital, error) {
	var hospitals []models.Hospital
	err := r.db.Where("is_deleted = 0").Find(&hospitals).Error
	return hospitals, err
}

func (r *HospitalRepositoryImpl) DeleteHospitalById(id int64, updatedBy string) error {
	tx := r.db.Begin()

	var hospital models.Hospital
	if err := tx.Where("hospital_id = ?", id).First(&hospital).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&models.Hospital{}).Where("hospital_id = ?", id).Updates(map[string]interface{}{
		"is_deleted": 1,
		"updated_by": updatedBy,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := r.InsertHospitalAudit(tx, &hospital, constant.DELETE, updatedBy); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
func (s *HospitalRepositoryImpl) UpdateService(service *models.Service, updatedBy string) error {
	tx := s.db.Begin() // Start the transaction with the repository's db connection

	// Update the service record
	if err := tx.Model(&models.Service{}).Where("service_id = ?", service.ServiceId).Updates(service).Error; err != nil {
		tx.Rollback() // Rollback the transaction if update fails
		return err
	}

	if err := SaveServiceAudit(tx, service, constant.UPDATE, updatedBy); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func SaveServiceAudit(tx *gorm.DB, service *models.Service, operation, updatedBy string) error {
	audit := models.ServiceAudit{
		ServiceId:     service.ServiceId,
		ServiceName:   service.ServiceName,
		CreatedBy:     service.CreatedBy,
		CreatedAt:     service.CreatedAt,
		OperationType: operation,
		UpdatedBy:     updatedBy,
		UpdatedAt:     time.Now(),
	}
	return tx.Create(&audit).Error
}

func (r *HospitalRepositoryImpl) AddService(service *models.Service) error {
	return r.db.Create(service).Error
}

func (r *HospitalRepositoryImpl) GetAllServices() ([]models.Service, error) {
	var services []models.Service
	err := r.db.Find(&services).Error
	return services, err
}

func (r *HospitalRepositoryImpl) GetServiceById(serviceId string) (models.Service, error) {
	var service models.Service
	err := r.db.Where("service_id = ?", serviceId).First(&service).Error
	return service, err
}

func (r *HospitalRepositoryImpl) DeleteService(serviceId uint64, deletedBy string) error {
	tx := r.db.Begin()
	var service models.Service
	if err := tx.Where("service_id = ?", serviceId).First(&service).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&models.Service{}).Where("service_id = ?", serviceId).Updates(map[string]interface{}{
		"is_deleted": 1,
		"updated_by": deletedBy,
		"updated_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := SaveServiceAudit(tx, &service, constant.DELETE, deletedBy); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (repo *HospitalRepositoryImpl) GetHospitalAuditRecord(hospitalId, hospitalAuditId uint64) ([]models.HospitalAudit, error) {
	var auditLogs []models.HospitalAudit
	query := repo.db

	if hospitalId != 0 {
		query = query.Where("hospital_id = ?", hospitalId)
	}
	if hospitalAuditId != 0 {
		query = query.Where("hospital_audit_id = ?", hospitalAuditId)
	}

	err := query.Order("hospital_audit_id DESC").Find(&auditLogs).Error
	if err != nil {
		return nil, err
	}
	return auditLogs, nil
}

func (repo *HospitalRepositoryImpl) GetAllHospitalAuditRecord(page, limit int) ([]models.HospitalAudit, int64, error) {
	var totalRecords int64
	var auditLogs []models.HospitalAudit

	err := repo.db.Model(&models.HospitalAudit{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}

	err = repo.db.Offset((page - 1) * limit).Limit(limit).Find(&auditLogs).Error
	if err != nil {
		return nil, 0, err
	}

	return auditLogs, totalRecords, nil
}

func (repo *HospitalRepositoryImpl) GetServiceAuditRecord(serviceId, serviceAuditId uint64) ([]models.ServiceAudit, error) {
	var auditLogs []models.ServiceAudit
	query := repo.db

	if serviceId != 0 {
		query = query.Where("service_id = ?", serviceId)
	}
	if serviceAuditId != 0 {
		query = query.Where("service_audit_id = ?", serviceAuditId)
	}

	err := query.Order("service_audit_id DESC").Find(&auditLogs).Error
	if err != nil {
		return nil, err
	}
	return auditLogs, nil
}

func (repo *HospitalRepositoryImpl) GetAllServiceAuditRecord(page, limit int) ([]models.ServiceAudit, int64, error) {
	var totalRecords int64
	var auditLogs []models.ServiceAudit

	err := repo.db.Model(&models.ServiceAudit{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}

	err = repo.db.Offset((page - 1) * limit).Limit(limit).Find(&auditLogs).Error
	if err != nil {
		return nil, 0, err
	}

	return auditLogs, totalRecords, nil
}

func (repo *HospitalRepositoryImpl) AddServiceMapping(serviceMapping models.ServiceMapping) error {
	err := repo.db.Create(&serviceMapping).Error
	if err != nil {
		return err
	}
	return nil
}
