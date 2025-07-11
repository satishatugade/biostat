package repository

import (
	"biostat/models"
	"fmt"
	"strings"

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
	FetchAddressByPincode(postalcode string) ([]models.PincodeMaster, error)
	FetchMappedUserAddress(patientId uint64, mappingType string, limit, offset int) ([]models.UserAddressResponse, int64, error)
	CheckUserEmailMobileExist(input *models.CheckUserMobileEmail) (bool, error)
	GetUserInfoByUserName(username string) (*models.UserLoginInfo, error)
	GetUserInfoByIdentifier(identifier string) (*models.UserLoginInfo, error)
	UpdateUserInfo(authUserId string, updateInfo map[string]interface{}) error
	GetUserInfoByEmailId(emailId string) (*models.SystemUser_, error)
	GetUserIdBySUB(sub string) (uint64, error)
	GetSystemUserInfo(userId uint64) (models.SystemUser_, error)
	IsUsernameExists(username string) bool
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

func (ds *UserRepositoryImpl) FetchAddressByPincode(postalcode string) ([]models.PincodeMaster, error) {
	var addresses []models.PincodeMaster
	if err := ds.db.Where("pincode = ?", postalcode).Find(&addresses).Error; err != nil {
		return nil, err
	}
	return addresses, nil
}

func (r *UserRepositoryImpl) FetchMappedUserAddress(patientID uint64, mappingType string, limit, offset int) ([]models.UserAddressResponse, int64, error) {
	var data []models.UserAddressResponse
	var total int64

	query := r.db.Table("tbl_system_user_role_mapping AS urm").
		Select(`su.user_id,urm.mapping_type, su.first_name, su.last_name,
				am.address_id, am.address_line1, am.address_line2,
				am.city, am.state, am.country, am.postal_code`).
		Joins("JOIN tbl_system_user_ AS su ON urm.user_id = su.user_id").
		Joins("JOIN tbl_system_user_address_mapping AS suad ON suad.user_id = su.user_id").
		Joins("JOIN tbl_address_master AS am ON am.address_id = suad.address_id").
		Where(`urm.patient_id = ? AND urm.mapping_type = ?
		AND COALESCE(am.address_line1, '') <> '' AND COALESCE(am.city, '') <> ''
		AND COALESCE(am.state, '') <> '' AND COALESCE(am.postal_code, '') <> ''`, patientID, mappingType)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Scan(&data).Error
	return data, total, err
}

func (ur *UserRepositoryImpl) CheckUserEmailMobileExist(input *models.CheckUserMobileEmail) (bool, error) {
	var count int64

	if input.Mobile != "" {
		err := ur.db.Model(&models.SystemUser_{}).
			Where("mobile_no = ?", input.Mobile).
			Count(&count).Error
		if err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}
	if input.Email != "" {
		err := ur.db.Model(&models.SystemUser_{}).
			Where("email = ?", strings.ToLower(input.Email)).
			Count(&count).Error
		if err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}
	return false, nil
}

func (ur *UserRepositoryImpl) GetUserInfoByUserName(username string) (*models.UserLoginInfo, error) {
	var info models.UserLoginInfo

	err := ur.db.Debug().
		Model(&models.SystemUser_{}).
		Select("auth_user_id", "password", "login_count").
		Where("username = ?", username).
		Scan(&info).Error

	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (ur *UserRepositoryImpl) GetUserInfoByIdentifier(identifier string) (*models.UserLoginInfo, error) {
	var info models.UserLoginInfo

	err := ur.db.Model(&models.SystemUser_{}).
		Select("auth_user_id", "username", "password", "login_count").
		Where("username = ? OR email = ? OR mobile_no = ?", identifier, identifier, identifier).
		Limit(1).
		Scan(&info).Error

	if err != nil {
		return nil, err
	}

	if info.Username == "" {
		return nil, fmt.Errorf("user not found with identifier: %s", identifier)
	}

	return &info, nil
}

func (ur *UserRepositoryImpl) GetUserInfoByEmailId(emailId string) (*models.SystemUser_, error) {
	var user models.SystemUser_

	err := ur.db.Where("email = ?", emailId).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepositoryImpl) UpdateUserInfo(userID string, updates map[string]interface{}) error {
	tx := ur.db.Begin()

	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Model(&models.SystemUser_{}).
		Where("auth_user_id = ?", userID).
		Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (u *UserRepositoryImpl) GetUserIdBySUB(SUB string) (uint64, error) {
	var user models.SystemUser_
	err := u.db.Select("user_id").Where("auth_user_id=?", SUB).First(&user).Error
	if err != nil {
		return 0, err
	}
	return user.UserId, nil
}

func (u *UserRepositoryImpl) GetSystemUserInfo(userId uint64) (models.SystemUser_, error) {
	var user models.SystemUser_
	err := u.db.Where("user_id = ?", userId).First(&user).Error
	if err != nil {
		return models.SystemUser_{}, err
	}
	return user, nil
}

func (u *UserRepositoryImpl) IsUsernameExists(username string) bool {
	var count int64
	u.db.Table("tbl_system_user_").Where("username=?", username).Count(&count)
	return count > 0
}
