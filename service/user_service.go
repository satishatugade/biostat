package service

import (
	"biostat/models"
	"biostat/repository"
)

type UserService interface {
	GetAllTblUserGtokens(limit int, offset int) ([]models.TblUserGtoken, int64, error)
	CreateTblUserGtoken(data *models.TblUserGtoken) (*models.TblUserGtoken, error)
	UpdateTblUserGtoken(data *models.TblUserGtoken, updatedBy string) (*models.TblUserGtoken, error)
	GetSingleTblUserGtoken(id int) (*models.TblUserGtoken, error)
	DeleteTblUserGtoken(id int, updatedBy string) error

	CreateSystemUser(systemUser models.SystemUser_) (models.SystemUser_, error)
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

func (s *UserServiceImpl) GetSingleTblUserGtoken(id int) (*models.TblUserGtoken, error) {
	return s.tblUserGtokenRepo.GetSingleTblUserGtoken(id)
}

func (s *UserServiceImpl) DeleteTblUserGtoken(id int, updatedBy string) error {
	return s.tblUserGtokenRepo.DeleteTblUserGtoken(id, updatedBy)
}

// CreateSystemUser implements UserService.
func (s *UserServiceImpl) CreateSystemUser(systemUser models.SystemUser_) (models.SystemUser_, error) {
	return s.tblUserGtokenRepo.CreateSystemUser(systemUser)

}
