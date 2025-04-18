package service

import (
	"biostat/models"
	"biostat/repository"

	"gorm.io/gorm"
)

type UserService interface {
	GetAllTblUserGtokens(limit int, offset int) ([]models.TblUserGtoken, int64, error)
	CreateTblUserGtoken(data *models.TblUserGtoken) (*models.TblUserGtoken, error)
	UpdateTblUserGtoken(data *models.TblUserGtoken, updatedBy string) (*models.TblUserGtoken, error)
	GetSingleTblUserGtoken(id uint64) (*models.TblUserGtoken, error)
	DeleteTblUserGtoken(id uint64, updatedBy string) error

	CreateSystemUser(tx *gorm.DB, systemUser models.SystemUser_) (models.SystemUser_, error)
}

type UserServiceImpl struct {
	tblUserGtokenRepo repository.UserRepository
}

func NewTblUserGtokenService(repo repository.UserRepository) UserService {
	return &UserServiceImpl{tblUserGtokenRepo: repo}
}

func (s *UserServiceImpl) GetAllTblUserGtokens(limit int, offset int) ([]models.TblUserGtoken, int64, error) {
	return s.tblUserGtokenRepo.GetAllTblUserGtokens(limit, offset)
}

func (s *UserServiceImpl) CreateTblUserGtoken(data *models.TblUserGtoken) (*models.TblUserGtoken, error) {
	return s.tblUserGtokenRepo.CreateTblUserGtoken(data)
}

func (s *UserServiceImpl) UpdateTblUserGtoken(data *models.TblUserGtoken, updatedBy string) (*models.TblUserGtoken, error) {
	return s.tblUserGtokenRepo.UpdateTblUserGtoken(data, updatedBy)
}

func (s *UserServiceImpl) GetSingleTblUserGtoken(id uint64) (*models.TblUserGtoken, error) {
	return s.tblUserGtokenRepo.GetSingleTblUserGtoken(id)
}

func (s *UserServiceImpl) DeleteTblUserGtoken(id uint64, updatedBy string) error {
	return s.tblUserGtokenRepo.DeleteTblUserGtoken(id, updatedBy)
}

// CreateSystemUser implements UserService.
func (s *UserServiceImpl) CreateSystemUser(tx *gorm.DB, systemUser models.SystemUser_) (models.SystemUser_, error) {
	return s.tblUserGtokenRepo.CreateSystemUser(tx, systemUser)

}
