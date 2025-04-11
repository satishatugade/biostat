package service

import (
	"biostat/models"
	"biostat/repository"
)

type SupportGroupService interface {
	AddSupportGroup(supportGroup *models.SupportGroup) error
	GetSupportGroupById(supportGrpId uint64) (*models.SupportGroup, error)
	GetAllSupportGroups(limit, offset int) ([]models.SupportGroup, int64, error)
	UpdateSupportGroup(updatedData *models.SupportGroup, authUserId string) error
	DeleteSupportGroup(supportGrpId uint64, authUserId string) error
	GetSupportGroupAuditRecord(supportGroupId, supportGroupAuditId uint64) ([]models.SupportGroupAudit, error)
	GetAllSupportGroupAuditRecord(limit, offset int) ([]models.SupportGroupAudit, int64, error)
}

type SupportGroupServiceImpl struct {
	supportGroupRepo repository.SupportGroupRepository
}

func NewSupportGroupService(repo repository.SupportGroupRepository) SupportGroupService {
	return &SupportGroupServiceImpl{supportGroupRepo: repo}
}

func (s *SupportGroupServiceImpl) AddSupportGroup(supportGroup *models.SupportGroup) error {
	return s.supportGroupRepo.AddSupportGroup(supportGroup)
}

func (s *SupportGroupServiceImpl) GetAllSupportGroups(limit, offset int) ([]models.SupportGroup, int64, error) {
	return s.supportGroupRepo.GetAllSupportGroups(limit, offset)
}

func (s *SupportGroupServiceImpl) UpdateSupportGroup(updatedData *models.SupportGroup, authUserId string) error {
	return s.supportGroupRepo.UpdateSupportGroup(updatedData, authUserId)
}

func (s *SupportGroupServiceImpl) DeleteSupportGroup(supportGrpId uint64, authUserId string) error {
	return s.supportGroupRepo.DeleteSupportGroup(supportGrpId, authUserId)
}

func (s *SupportGroupServiceImpl) GetSupportGroupById(supportGrpId uint64) (*models.SupportGroup, error) {
	return s.supportGroupRepo.GetSupportGroupById(supportGrpId)
}

// support_group_service_impl.go
func (s *SupportGroupServiceImpl) GetSupportGroupAuditRecord(supportGroupId, supportGroupAuditId uint64) ([]models.SupportGroupAudit, error) {
	return s.supportGroupRepo.GetSupportGroupAuditRecord(supportGroupId, supportGroupAuditId)
}

func (s *SupportGroupServiceImpl) GetAllSupportGroupAuditRecord(limit, offset int) ([]models.SupportGroupAudit, int64, error) {
	return s.supportGroupRepo.GetAllSupportGroupAuditRecord(limit, offset)
}
