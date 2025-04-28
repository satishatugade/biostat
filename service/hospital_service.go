package service

import (
	"biostat/models"
	"biostat/repository"
)

type HospitalService interface {
	AddHospital(hospitaL *models.Hospital) error
	UpdateHospital(hospital *models.Hospital, authUserId string) error
	GetAllHospitals(isDeleted *int) ([]models.Hospital, error)
	DeleteHospitalById(hospitalId int64, updatedBy string) error
	GetHospitalById(hospitalId uint64) (models.Hospital, error)

	//service
	AddService(service *models.Service) error
	GetAllServices() ([]models.Service, error)
	GetServiceById(serviceId string) (models.Service, error)
	UpdateService(service *models.Service, updatedBy string) error
	DeleteService(serviceId uint64, deletedBy string) error

	AddServiceMapping(serviceMapping models.ServiceMapping) error

	GetHospitalAuditRecord(hospitalId uint64, hospitalAuditId uint64) ([]models.HospitalAudit, error)
	GetAllHospitalAuditRecord(limit, offset int) ([]models.HospitalAudit, int64, error)

	GetServiceAuditRecord(serviceId uint64, serviceAuditId uint64) ([]models.ServiceAudit, error)
	GetAllServiceAuditRecord(limit, offset int) ([]models.ServiceAudit, int64, error)
}

type HospitalServiceImpl struct {
	hospitalRepo repository.HospitalRepository
}

func NewHospitalService(repo repository.HospitalRepository) HospitalService {
	return &HospitalServiceImpl{hospitalRepo: repo}
}

func (s *HospitalServiceImpl) AddHospital(h *models.Hospital) error {
	return s.hospitalRepo.AddHospital(h)
}

func (s *HospitalServiceImpl) UpdateHospital(hospital *models.Hospital, authUserId string) error {
	return s.hospitalRepo.UpdateHospital(hospital, authUserId)
}

func (s *HospitalServiceImpl) GetAllHospitals(isDeleted *int) ([]models.Hospital, error) {
	return s.hospitalRepo.GetAllHospitals(isDeleted)
}

func (s *HospitalServiceImpl) DeleteHospitalById(hospitalId int64, updatedBy string) error {
	return s.hospitalRepo.DeleteHospitalById(hospitalId, updatedBy)
}

func (s *HospitalServiceImpl) GetHospitalById(hospitalId uint64) (models.Hospital, error) {
	return s.hospitalRepo.GetHospitalById(hospitalId)
}

// AddService implements HospitalService.
func (s *HospitalServiceImpl) AddService(service *models.Service) error {
	return s.hospitalRepo.AddService(service)
}

// DeleteService implements HospitalService.
func (s *HospitalServiceImpl) DeleteService(serviceId uint64, deletedBy string) error {
	return s.hospitalRepo.DeleteService(serviceId, deletedBy)

}

// GetAllServices implements HospitalService.
func (s *HospitalServiceImpl) GetAllServices() ([]models.Service, error) {
	return s.hospitalRepo.GetAllServices()

}

// GetServiceById implements HospitalService.
func (s *HospitalServiceImpl) GetServiceById(serviceId string) (models.Service, error) {
	return s.hospitalRepo.GetServiceById(serviceId)

}

// UpdateService implements HospitalService.
func (s *HospitalServiceImpl) UpdateService(service *models.Service, updatedBy string) error {
	return s.hospitalRepo.UpdateService(service, updatedBy)
}

func (s *HospitalServiceImpl) GetAllHospitalAuditRecord(limit, offset int) ([]models.HospitalAudit, int64, error) {
	return s.hospitalRepo.GetAllHospitalAuditRecord(limit, offset)
}

func (s *HospitalServiceImpl) GetHospitalAuditRecord(hospitalId, hospitalAuditId uint64) ([]models.HospitalAudit, error) {
	return s.hospitalRepo.GetHospitalAuditRecord(hospitalId, hospitalAuditId)
}

func (s *HospitalServiceImpl) GetAllServiceAuditRecord(limit, offset int) ([]models.ServiceAudit, int64, error) {
	return s.hospitalRepo.GetAllServiceAuditRecord(limit, offset)
}

func (s *HospitalServiceImpl) GetServiceAuditRecord(serviceId, serviceAuditId uint64) ([]models.ServiceAudit, error) {
	return s.hospitalRepo.GetServiceAuditRecord(serviceId, serviceAuditId)
}

func (s *HospitalServiceImpl) AddServiceMapping(serviceMapping models.ServiceMapping) error {
	return s.hospitalRepo.AddServiceMapping(serviceMapping)
}
