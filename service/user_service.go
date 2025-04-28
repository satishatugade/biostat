package service

import (
	"biostat/models"
	"biostat/repository"

	"gorm.io/gorm"
)

type UserService interface {
	GetAllTblUserTokens(limit int, offset int) ([]models.TblUserToken, int64, error)
	CreateTblUserToken(data *models.TblUserToken) (*models.TblUserToken, error)
	UpdateTblUserToken(data *models.TblUserToken, updatedBy string) (*models.TblUserToken, error)
	GetSingleTblUserToken(id uint64, provider string) (*models.TblUserToken, error)
	DeleteTblUserToken(id uint64, updatedBy string) error

	CreateSystemUser(tx *gorm.DB, systemUser models.SystemUser_) (models.SystemUser_, error)
}

type UserServiceImpl struct {
	userRepo    repository.UserRepository
	patientrepo repository.PatientRepository
}

func NewTblUserTokenService(repo repository.UserRepository) UserService {
	return &UserServiceImpl{userRepo: repo}
}

func (s *UserServiceImpl) GetAllTblUserTokens(limit int, offset int) ([]models.TblUserToken, int64, error) {
	return s.userRepo.GetAllTblUserTokens(limit, offset)
}

func (s *UserServiceImpl) CreateTblUserToken(data *models.TblUserToken) (*models.TblUserToken, error) {
	return s.userRepo.CreateTblUserToken(data)
}

func (s *UserServiceImpl) UpdateTblUserToken(data *models.TblUserToken, updatedBy string) (*models.TblUserToken, error) {
	return s.userRepo.UpdateTblUserToken(data, updatedBy)
}

func (s *UserServiceImpl) GetSingleTblUserToken(id uint64, provider string) (*models.TblUserToken, error) {
	return s.userRepo.GetSingleTblUserToken(id, provider)
}

func (s *UserServiceImpl) DeleteTblUserToken(id uint64, updatedBy string) error {
	return s.userRepo.DeleteTblUserToken(id, updatedBy)
}

// CreateSystemUser implements UserService.
func (s *UserServiceImpl) CreateSystemUser(tx *gorm.DB, systemUser models.SystemUser_) (models.SystemUser_, error) {
	// return s.userRepo.CreateSystemUser(tx, systemUser)
	createdUser, err := s.userRepo.CreateSystemUser(tx, systemUser)
	if err != nil {
		return models.SystemUser_{}, err
	}
	userAddress, err := s.userRepo.CreateSystemUserAddress(tx, systemUser.UserAddress)
	if err != nil {
		return models.SystemUser_{}, err
	}
	userAddressMapping := models.SystemUserAddressMapping{
		UserId:    createdUser.UserId,
		AddressId: userAddress.AddressId,
	}
	MappingErr := s.userRepo.CreateSystemUserAddressMapping(tx, userAddressMapping)
	if MappingErr != nil {
		return models.SystemUser_{}, err
	}
	return createdUser, nil
}
