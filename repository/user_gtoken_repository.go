package repository

import (
	"biostat/models"
	"log"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetAllTblUserGtokens(limit int, offset int) ([]models.TblUserGtoken, int64, error)
	CreateTblUserGtoken(data *models.TblUserGtoken) (*models.TblUserGtoken, error)
	UpdateTblUserGtoken(data *models.TblUserGtoken, updatedBy string) (*models.TblUserGtoken, error)
	GetSingleTblUserGtoken(id uint64) (*models.TblUserGtoken, error)
	DeleteTblUserGtoken(id uint64, updatedBy string) error
	CreateSystemUser(tx *gorm.DB, systemUser models.SystemUser_) (models.SystemUser_, error)
}

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewTblUserGtokenRepository(db *gorm.DB) UserRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) GetAllTblUserGtokens(limit int, offset int) ([]models.TblUserGtoken, int64, error) {
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

func (r *UserRepositoryImpl) CreateTblUserGtoken(data *models.TblUserGtoken) (*models.TblUserGtoken, error) {
	err := r.db.Create(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *UserRepositoryImpl) UpdateTblUserGtoken(data *models.TblUserGtoken, updatedBy string) (*models.TblUserGtoken, error) {
	err := r.db.Model(&models.TblUserGtoken{}).Where("id = ?", data.Id).Updates(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *UserRepositoryImpl) GetSingleTblUserGtoken(id uint64) (*models.TblUserGtoken, error) {
	var obj models.TblUserGtoken
	err := r.db.Where("user_id = ?", id).First(&obj).Error
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func (r *UserRepositoryImpl) DeleteTblUserGtoken(id uint64, updatedBy string) error {
	return r.db.Where("user_id = ?", id).Delete(&models.TblUserGtoken{}).Error
}

// CreateSystemUser implements UserRepository.
func (r *UserRepositoryImpl) CreateSystemUser(tx *gorm.DB, systemUser models.SystemUser_) (models.SystemUser_, error) {
	log.Println("User pass:", systemUser.Password)
	if err := tx.Create(&systemUser).Error; err != nil {
		return models.SystemUser_{}, err
	}
	return systemUser, nil
}
