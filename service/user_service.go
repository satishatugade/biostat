package service

import (
	"biostat/config"
	"biostat/constant"
	"biostat/models"
	"biostat/repository"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"math/rand"

	"gorm.io/gorm"
)

type UserService interface {
	GetAllTblUserTokens(limit int, offset int) ([]models.TblUserToken, int64, error)
	CreateTblUserToken(data *models.TblUserToken) (*models.TblUserToken, error)
	UpdateTblUserToken(data *models.TblUserToken, updatedBy string) (*models.TblUserToken, error)
	GetSingleTblUserToken(userID uint64, provider string, providerID *string) (*models.TblUserToken, error)
	GetUserProviderIDs(userID uint64, provider string) ([]string, error)
	FetchAddressByPincode(postalcode string) ([]models.PincodeMaster, error)
	GetAllMappedUserAddress(patientId uint64, limit, offset int, MappingType []string) ([]models.UserAddressResponse, int64, error)
	GetUserIdBySUB(sub string) (uint64, error)
	GetSystemUserInfoByAuthUserId(authUserId string) (models.SystemUser_, error)
	GetSystemUserInfoByUserID(userId uint64) (models.SystemUser_, error)
	CreateSystemUser(tx *gorm.DB, systemUser models.SystemUser_) (models.SystemUser_, error)
	CheckUserEmailMobileExist(input *models.CheckUserMobileEmail) (bool, *models.SystemUser_, error)
	GetUserInfoByUserName(username string) (*models.UserLoginInfo, error)
	GetUserInfoByIdentifier(identifier string) (*models.UserLoginInfo, error)
	GetUserInfoByEmailId(emailId string) (*models.SystemUser_, error)
	UpdateUserInfo(authUserId string, updateInfo map[string]interface{}) error
	IsUsernameExists(username string) bool
	GenerateUniqueUsername(firstName, lastName string) string

	CreateSharedProfileLink(userId uint64, username, password string) (*models.CreateSharedLinkResponse, error)
	GetSharedProfileLink(id string) (*models.UserLoginResponse, error)

	SyncOtherDocs(patientUserId uint64, newRelative models.SystemUser_) error
}

type UserServiceImpl struct {
	userRepo   repository.UserRepository
	apiService ApiService
	recordRepo repository.TblMedicalRecordRepository
	db         *gorm.DB
}

func NewTblUserTokenService(userRepo repository.UserRepository, apiService ApiService, recordRepo repository.TblMedicalRecordRepository, db *gorm.DB) UserService {
	return &UserServiceImpl{userRepo: userRepo, apiService: apiService, recordRepo: recordRepo, db: db}
}

func (s *UserServiceImpl) GetAllTblUserTokens(limit int, offset int) ([]models.TblUserToken, int64, error) {
	return s.userRepo.GetAllTblUserTokens(limit, offset)
}

func (s *UserServiceImpl) CreateTblUserToken(data *models.TblUserToken) (*models.TblUserToken, error) {
	return s.userRepo.UpsertUserToken(data)
}

func (s *UserServiceImpl) UpdateTblUserToken(data *models.TblUserToken, updatedBy string) (*models.TblUserToken, error) {
	return s.userRepo.UpdateTblUserToken(data, updatedBy)
}

func (s *UserServiceImpl) GetSingleTblUserToken(userID uint64, provider string, providerID *string) (*models.TblUserToken, error) {
	return s.userRepo.GetUserToken(userID, provider, providerID)
}

func (s *UserServiceImpl) GetUserProviderIDs(userID uint64, provider string) ([]string, error) {
	return s.userRepo.GetUserProviderIDs(userID, provider)
}

func (s *UserServiceImpl) GetAllMappedUserAddress(patientID uint64, limit, offset int, MappingType []string) ([]models.UserAddressResponse, int64, error) {
	return s.userRepo.FetchMappedUserAddress(patientID, MappingType, limit, offset)
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

func (s *UserServiceImpl) FetchAddressByPincode(postalcode string) ([]models.PincodeMaster, error) {
	return s.userRepo.FetchAddressByPincode(postalcode)
}

func (ps *UserServiceImpl) CheckUserEmailMobileExist(input *models.CheckUserMobileEmail) (bool, *models.SystemUser_, error) {
	return ps.userRepo.CheckUserEmailMobileExist(input)
}

func (s *UserServiceImpl) GetUserInfoByUserName(username string) (*models.UserLoginInfo, error) {
	return s.userRepo.GetUserInfoByUserName(username)
}

func (s *UserServiceImpl) GetUserInfoByIdentifier(identifier string) (*models.UserLoginInfo, error) {
	return s.userRepo.GetUserInfoByIdentifier(identifier)
}

func (s *UserServiceImpl) GetUserInfoByEmailId(emailId string) (*models.SystemUser_, error) {
	return s.userRepo.GetUserInfoByEmailId(emailId)
}

func (s *UserServiceImpl) UpdateUserInfo(authUserId string, updateInfo map[string]interface{}) error {
	return s.userRepo.UpdateUserInfo(authUserId, updateInfo)
}

func (s *UserServiceImpl) GetUserIdBySUB(sub string) (uint64, error) {
	userId, err := s.userRepo.GetUserIdBySUB(sub)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func (s *UserServiceImpl) GetSystemUserInfoByAuthUserId(sub string) (models.SystemUser_, error) {
	userId, err := s.userRepo.GetUserIdBySUB(sub)
	if err != nil {
		return models.SystemUser_{}, err
	}
	return s.userRepo.GetSystemUserInfo(userId)
}

func (s *UserServiceImpl) IsUsernameExists(username string) bool {
	return s.userRepo.IsUsernameExists(username)
}

func sanitizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "")
	return name
}

func (s *UserServiceImpl) GenerateUniqueUsername(firstName, lastName string) string {
	rand.Seed(time.Now().UnixNano())
	base := fmt.Sprintf("%s.%s", sanitizeName(firstName), sanitizeName(lastName))

	for i := 0; i < 5; i++ {
		suffix := rand.Intn(10000)
		username := fmt.Sprintf("%s.%04d", base, suffix)
		if s.IsUsernameExists(username) {
			return username
		}
	}
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s.%d", base, timestamp)
}

func (s *UserServiceImpl) GetSystemUserInfoByUserID(userId uint64) (models.SystemUser_, error) {
	return s.userRepo.GetSystemUserInfo(userId)
}

func (s *UserServiceImpl) CreateSharedProfileLink(userId uint64, username, password string) (*models.CreateSharedLinkResponse, error) {
	client := config.Client
	ctx := context.Background()
	token, err := client.Login(ctx, config.KeycloakPublicClientID, config.KeycloakPublicClientSecret, config.KeycloakRealm, username, password)
	if err != nil {
		return nil, err
	}

	link := &models.SharedProfileLink{
		UserID:    userId,
		Token:     token.AccessToken,
		ExpiresAt: time.Now().Add(time.Duration(token.ExpiresIn) * time.Second),
	}
	err = s.userRepo.CreateShareProfileLink(link)
	if err != nil {
		return nil, err
	}
	log.Println(link)
	APPURL := config.PropConfig.ApiURL.APPURL
	response := &models.CreateSharedLinkResponse{
		ShareURL:  fmt.Sprintf("%s/view/profile/%s", APPURL, link.SharedPorfileLinkId),
		ExpiresAt: link.ExpiresAt.Format(time.RFC3339),
	}

	return response, nil
}

func (s *UserServiceImpl) GetSharedProfileLink(id string) (*models.UserLoginResponse, error) {

	linkResp, err := s.userRepo.GetSharedProfileLinkById(id)
	if err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetSystemUserInfo(linkResp.UserID)
	if err != nil {
		return nil, err
	}
	userLoginResponse := &models.UserLoginResponse{
		AccessToken: linkResp.Token,
		UserResponse: models.UserResponse{
			UserId:     user.UserId,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			Email:      user.Email,
			Username:   user.Username,
			Role:       string(constant.Patient),
			AuthUserId: user.AuthUserId,
			AccessType: constant.ViewOnly,
		},
	}
	return userLoginResponse, nil

}

func (s *UserServiceImpl) SyncOtherDocs(patientUserId uint64, newRelative models.SystemUser_) error {
	// log.Println("Starting to Sync Docs in Other bucket")
	records, err := s.userRepo.GetOtherBucketRecordsByPatientID(newRelative.UserId)
	if err != nil {
		log.Println("Error fetching Records from other bucket:", err)
		return err
	}
	var errorList []error
	newMmeberName := fmt.Sprintf("%s %s", newRelative.FirstName, newRelative.LastName)
	for _, record := range records {
		// log.Println("Processing record:", i, ":", record.RecordID)
		shouldMove, err := s.apiService.CallCheckOwnerOtherTypeAPI(record.RecordName, record.UserID, record.DocumentOwner, newMmeberName)
		if err != nil {
			// log.Println("Error Processing record:", record.RecordID, "Error:", err)
			errorList = append(errorList, err)
			continue
		}
		if shouldMove == "YES" {
			// log.Println("Move ", record.RecordID, " To ", newRelative.UserId)
			tx := s.db.Begin()
			err := s.recordRepo.UpdateMedicalRecordMappingByRecordId(tx, &record.RecordID, map[string]interface{}{"user_id": newRelative.UserId, "is_unknown_record": false})
			if err != nil {
				// log.Println("Error moving Record ", err)
				errorList = append(errorList, err)
				continue
			}
			updatedRecord := &models.TblMedicalRecord{
				RecordId:       record.RecordID,
				RecordCategory: record.DocumentBucket,
			}

			if err := tx.Commit().Error; err != nil {
				errorList = append(errorList, err)
				continue
			}
			_, err = s.recordRepo.UpdateTblMedicalRecord(updatedRecord)
			if err != nil {
				// log.Println("Error Updating Record ", err)
				errorList = append(errorList, err)
				continue
			}
		}
	}
	finalErr := errors.Join(errorList...)
	return finalErr
}
