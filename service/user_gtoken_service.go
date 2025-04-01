package service

import (
	"biostat/models"
	"biostat/repository"
)

type TblUserGtokenService interface {
	GetAllTblUserGtokens(limit int, offset int) ([]models.TblUserGtoken, int64, error)
	CreateTblUserGtoken(data *models.TblUserGtoken) (*models.TblUserGtoken, error)
	UpdateTblUserGtoken(data *models.TblUserGtoken, updatedBy string) (*models.TblUserGtoken, error)
	GetSingleTblUserGtoken(id int) (*models.TblUserGtoken, error)
	DeleteTblUserGtoken(id int, updatedBy string) error
}

type tblUserGtokenServiceImpl struct {
	tblUserGtokenRepo repository.TblUserGtokenRepository
}

func NewTblUserGtokenService(repo repository.TblUserGtokenRepository) TblUserGtokenService {
	return &tblUserGtokenServiceImpl{tblUserGtokenRepo: repo}
}

func (s *tblUserGtokenServiceImpl) GetAllTblUserGtokens(limit int, offset int) ([]models.TblUserGtoken, int64, error) {
	return s.tblUserGtokenRepo.GetAllTblUserGtokens(limit, offset)
}

func (s *tblUserGtokenServiceImpl) CreateTblUserGtoken(data *models.TblUserGtoken) (*models.TblUserGtoken, error) {
	return s.tblUserGtokenRepo.CreateTblUserGtoken(data)
}

func (s *tblUserGtokenServiceImpl) UpdateTblUserGtoken(data *models.TblUserGtoken, updatedBy string) (*models.TblUserGtoken, error) {
	return s.tblUserGtokenRepo.UpdateTblUserGtoken(data, updatedBy)
}

func (s *tblUserGtokenServiceImpl) GetSingleTblUserGtoken(id int) (*models.TblUserGtoken, error) {
	return s.tblUserGtokenRepo.GetSingleTblUserGtoken(id)
}

func (s *tblUserGtokenServiceImpl) DeleteTblUserGtoken(id int, updatedBy string) error {
	return s.tblUserGtokenRepo.DeleteTblUserGtoken(id, updatedBy)
}
