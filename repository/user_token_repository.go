package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetAllTblUserTokens(limit int, offset int) ([]models.TblUserToken, int64, error)
	CreateTblUserToken(data *models.TblUserToken) (*models.TblUserToken, error)
	UpdateTblUserToken(data *models.TblUserToken, updatedBy string) (*models.TblUserToken, error)
	GetSingleTblUserToken(id uint64, provider string) (*models.TblUserToken, error)
	DeleteTblUserToken(id uint64, updatedBy string) error
	CreateSystemUser(tx *gorm.DB, systemUser models.SystemUser_) (models.SystemUser_, error)
	CreateSystemUserAddress(tx *gorm.DB, systemUserAddress models.AddressMaster) (models.AddressMaster, error)
	CreateSystemUserAddressMapping(tx *gorm.DB, userAddressMapping models.SystemUserAddressMapping) error
}

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewTblUserTokenRepository(db *gorm.DB) UserRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) GetAllTblUserTokens(limit int, offset int) ([]models.TblUserToken, int64, error) {
	var objs []models.TblUserToken
	var totalRecords int64
	err := r.db.Model(&models.TblUserToken{}).Count(&totalRecords).Error
	if err != nil {
		return nil, 0, err
	}
	err = r.db.Limit(limit).Offset(offset).Find(&objs).Error
	if err != nil {
		return nil, 0, err
	}
	return objs, totalRecords, nil
}

func (r *UserRepositoryImpl) CreateTblUserToken(data *models.TblUserToken) (*models.TblUserToken, error) {
	err := r.db.Create(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *UserRepositoryImpl) UpdateTblUserToken(data *models.TblUserToken, updatedBy string) (*models.TblUserToken, error) {
	err := r.db.Model(&models.TblUserToken{}).Where("user_token_id = ?", data.Id).Updates(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *UserRepositoryImpl) GetSingleTblUserToken(id uint64, provider string) (*models.TblUserToken, error) {
	var obj models.TblUserToken
	err := r.db.Where("user_id = ? AND provider=?", id, provider).Order("created_at DESC").First(&obj).Error
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func (r *UserRepositoryImpl) DeleteTblUserToken(id uint64, updatedBy string) error {
	return r.db.Where("user_id = ?", id).Delete(&models.TblUserToken{}).Error
}

// CreateSystemUser implements UserRepository.
func (r *UserRepositoryImpl) CreateSystemUser(tx *gorm.DB, systemUser models.SystemUser_) (models.SystemUser_, error) {
	if err := tx.Create(&systemUser).Error; err != nil {
		return models.SystemUser_{}, err
	}
	return systemUser, nil
}

func (r *UserRepositoryImpl) CreateSystemUserAddress(tx *gorm.DB, systemUserAddress models.AddressMaster) (models.AddressMaster, error) {
	if err := tx.Create(&systemUserAddress).Error; err != nil {
		return models.AddressMaster{}, err
	}
	return systemUserAddress, nil
}

func (r *UserRepositoryImpl) CreateSystemUserAddressMapping(tx *gorm.DB, userAddressMapping models.SystemUserAddressMapping) error {
	if err := tx.Create(&userAddressMapping).Error; err != nil {
		return err
	}
	return nil
}
