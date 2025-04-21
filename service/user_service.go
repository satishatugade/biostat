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
	GetSingleTblUserToken(id uint64,provider string) (*models.TblUserToken, error)
	DeleteTblUserToken(id uint64, updatedBy string) error

	CreateSystemUser(tx *gorm.DB, systemUser models.SystemUser_) (models.SystemUser_, error)
}

type UserServiceImpl struct {
	tblUserGtokenRepo repository.UserRepository
}

func NewTblUserTokenService(repo repository.UserRepository) UserService {
	return &UserServiceImpl{tblUserGtokenRepo: repo}
}

func (s *UserServiceImpl) GetAllTblUserTokens(limit int, offset int) ([]models.TblUserToken, int64, error) {
	return s.tblUserGtokenRepo.GetAllTblUserTokens(limit, offset)
}

func (s *UserServiceImpl) CreateTblUserToken(data *models.TblUserToken) (*models.TblUserToken, error) {
	return s.tblUserGtokenRepo.CreateTblUserToken(data)
}

func (s *UserServiceImpl) UpdateTblUserToken(data *models.TblUserToken, updatedBy string) (*models.TblUserToken, error) {
	return s.tblUserGtokenRepo.UpdateTblUserToken(data, updatedBy)
}

func (s *UserServiceImpl) GetSingleTblUserToken(id uint64,provider string) (*models.TblUserToken, error) {
	return s.tblUserGtokenRepo.GetSingleTblUserToken(id,provider)
}

func (s *UserServiceImpl) DeleteTblUserToken(id uint64, updatedBy string) error {
	return s.tblUserGtokenRepo.DeleteTblUserToken(id, updatedBy)
}

// CreateSystemUser implements UserService.
func (s *UserServiceImpl) CreateSystemUser(tx *gorm.DB, systemUser models.SystemUser_) (models.SystemUser_, error) {
	return s.tblUserGtokenRepo.CreateSystemUser(tx, systemUser)

}
