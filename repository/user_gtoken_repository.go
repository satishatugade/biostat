package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type TblUserGtokenRepository interface {
	GetAllTblUserGtokens(limit int, offset int) ([]models.TblUserGtoken, int64, error)
	CreateTblUserGtoken(data *models.TblUserGtoken) (*models.TblUserGtoken, error)
	UpdateTblUserGtoken(data *models.TblUserGtoken, updatedBy string) (*models.TblUserGtoken, error)
	GetSingleTblUserGtoken(id int) (*models.TblUserGtoken, error)
	DeleteTblUserGtoken(id int, updatedBy string) error
}

type tblUserGtokenRepositoryImpl struct {
	db *gorm.DB
}

func NewTblUserGtokenRepository(db *gorm.DB) TblUserGtokenRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &tblUserGtokenRepositoryImpl{db: db}
}

func (r *tblUserGtokenRepositoryImpl) GetAllTblUserGtokens(limit int, offset int) ([]models.TblUserGtoken, int64, error) {
	var objs []models.TblUserGtoken
	var totalRecords int64
	err := r.db.Model(&models.TblUserGtoken{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}
	err = r.db.Limit(limit).Offset(offset).Find(&objs).Error
	if err != nil {
		return nil, 0, err
	}
	return objs, totalRecords, nil
}

func (r *tblUserGtokenRepositoryImpl) CreateTblUserGtoken(data *models.TblUserGtoken) (*models.TblUserGtoken, error) {
	err := r.db.Create(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *tblUserGtokenRepositoryImpl) UpdateTblUserGtoken(data *models.TblUserGtoken, updatedBy string) (*models.TblUserGtoken, error) {
	err := r.db.Model(&models.TblUserGtoken{}).Where("id = ?", data.Id).Updates(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *tblUserGtokenRepositoryImpl) GetSingleTblUserGtoken(id int) (*models.TblUserGtoken, error) {
	var obj models.TblUserGtoken
	err := r.db.Where("user_id = ?", id).First(&obj).Error
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func (r *tblUserGtokenRepositoryImpl) DeleteTblUserGtoken(id int, updatedBy string) error {
	return r.db.Where("user_id = ?", id).Delete(&models.TblUserGtoken{}).Error
}
