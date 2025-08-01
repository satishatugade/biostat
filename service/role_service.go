package service

import (
	"biostat/constant"
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"errors"
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

type RoleService interface {
	GetRoleById(roleId uint64) (models.RoleMaster, error)
	HasHOFMapping(patientId uint64, mappingType string) (bool, error)
	GetRoleIdByRoleName(roleName string) (models.RoleMaster, error)
	GetRoleByUserId(UserId uint64, mappingType *string) (models.RoleMaster, error)
	AddSystemUserMapping(tx *gorm.DB, patientUserId *uint64, newUserInfo models.SystemUser_, userInfo *models.SystemUser_, roleId uint64, roleName string, patientRelation *models.RelationMaster, isExistingUser *bool, RelativeIds []uint64) error
	CreateReverseMappingsForRelative(patientUserId, userId uint64, genderId uint64) error
	AddUserRelativeMappings(tx *gorm.DB, userId uint64, relativeId uint64, relation models.RelationMaster, roleId uint64, registeringUser *models.SystemUser_, newUser *models.SystemUser_) error
	GetMappingTypeByPatientId(patientUserId *uint64) (string, error)
}

type RoleServiceImpl struct {
	roleRepo         repository.RoleRepository
	patientService   PatientService
	userRepo         repository.UserRepository
	subscriptionRepo repository.SubscriptionRepository
}

func NewRoleService(repo repository.RoleRepository, patientService PatientService, userRepo repository.UserRepository, subscriptionRepo repository.SubscriptionRepository) RoleService {
	return &RoleServiceImpl{roleRepo: repo, patientService: patientService, userRepo: userRepo, subscriptionRepo: subscriptionRepo}
}

// GetRoleById implements RoleService.
func (r *RoleServiceImpl) GetRoleById(roleId uint64) (models.RoleMaster, error) {
	return r.roleRepo.GetRoleById(roleId)
}

func (r *RoleServiceImpl) GetRoleIdByRoleName(roleName string) (models.RoleMaster, error) {
	return r.roleRepo.GetRoleIdByRoleName(roleName)

}

func (r *RoleServiceImpl) GetRoleByUserId(UserId uint64, mappingType *string) (models.RoleMaster, error) {
	role, err := r.roleRepo.GetSystemUserRoleMappingByUserIdMappingType(UserId, mappingType)
	if err != nil {
		return models.RoleMaster{}, err
	}
	return r.roleRepo.GetRoleById(role.RoleId)
}

func (r *RoleServiceImpl) HasHOFMapping(patinetId uint64, mappingType string) (bool, error) {
	return r.roleRepo.HasHOFMapping(patinetId, mappingType)
}

func (r *RoleServiceImpl) AddSystemUserMapping(tx *gorm.DB, patientUserId *uint64, newUserInfo models.SystemUser_, userInfo *models.SystemUser_, roleId uint64, roleName string, patientRelation *models.RelationMaster, isExistingUser *bool, RelativeIds []uint64) error {
	roleName = strings.ToLower(roleName)
	mappingType := utils.GetMappingTypeByRoleName(roleName, nil)
	isSelf := roleName == "patient"
	if mappingType == "" {
		return errors.New("invalid role name")
	}
	var patientId uint64
	var relationId uint64
	if patientUserId == nil {
		patientId = newUserInfo.UserId
	} else {
		patientId = *patientUserId
	}
	if patientRelation == nil {
		relationId = 0
	} else {
		relationId = *patientRelation.RelationId
	}
	if roleName == string(constant.Caregiver) {
		if isExistingUser != nil && *isExistingUser {
			log.Println("patientId: ", patientId)
			exist, existingMapping := r.CheckDeletedUserMappingWithPatient(patientId, newUserInfo.Email, mappingType)
			log.Println("CheckDeletedUserMappingWithPatient ", exist, existingMapping)
			if exist {
				log.Printf("[AddSystemUserMapping] Found deleted mapping (id=%d), restoring it...", existingMapping.UserId)
				existingMapping.IsDeleted = 0
				existingMapping.RelationId = newUserInfo.RelationId
				return r.roleRepo.UpdateSystemUserMapping(existingMapping)
			}
		}
	} else if roleName == string(constant.Doctor) {
		systemUsermapping := models.SystemUserRoleMapping{
			UserId:      newUserInfo.UserId,
			RoleId:      roleId,
			MappingType: mappingType,
			IsSelf:      isSelf,
			PatientId:   patientId,
			RelationId:  relationId,
		}
		// realtives, err := r.patientService.GetRelativeList(&patientId)
		// if err != nil {
		// 	return err
		// }
		// log.Println("List of ALL Relatives for User to add family doctor mapping :", patientId)
		if len(RelativeIds) > 0 {
			for _, relativeId := range RelativeIds {
				err := r.roleRepo.AddSystemUserMapping(tx, &models.SystemUserRoleMapping{
					UserId:      newUserInfo.UserId,
					PatientId:   relativeId,
					IsSelf:      false,
					RoleId:      2,
					MappingType: string(constant.MappingTypeD),
					RelationId:  21,
					FamilyId:    nil,
				})
				if err != nil {
					log.Printf("failed to add system user mapping: %v", err)
				}
			}
		}
		return r.roleRepo.AddSystemUserMapping(tx, &systemUsermapping)

	}
	systemUsermapping := models.SystemUserRoleMapping{
		UserId:      newUserInfo.UserId,
		RoleId:      roleId,
		MappingType: mappingType,
		IsSelf:      isSelf,
		PatientId:   patientId,
		RelationId:  relationId,
	}
	return r.roleRepo.AddSystemUserMapping(tx, &systemUsermapping)
}

func (r *RoleServiceImpl) CheckDeletedUserMappingWithPatient(patientId uint64, email, mappingType string) (bool, *models.SystemUserRoleMapping) {
	log.Println("CheckDeletedUserMappingWithPatient ", email)
	exist, existingUser, _ := r.userRepo.CheckUserEmailMobileExist(&models.CheckUserMobileEmail{Email: email})
	if exist {
		return r.roleRepo.CheckDeletedUserMappingWithPatient(existingUser.UserId, patientId, mappingType)
	}
	return false, nil
}

// func (r *RoleServiceImpl) AddSystemUserMapping(tx *gorm.DB, patientUserId *uint64, userId uint64, userInfo *models.SystemUser_, roleId uint64, roleName string, patientRelation *models.RelationMaster) error {
// 	roleName = strings.ToLower(roleName)
// 	mappingType := utils.GetMappingTypeByRoleName(roleName, nil)
// 	if mappingType == "" {
// 		return errors.New("invalid role name")
// 	}
// 	relationType, _ := r.roleRepo.GetMappingTypeByPatientId(patientUserId)

// 	isSelf := roleName == "patient"
// 	var patientId uint64
// 	if patientUserId == nil {
// 		patientId = userId
// 	} else {
// 		patientId = *patientUserId
// 	}

// 	var relationId uint64
// 	if patientRelation == nil {
// 		relationId = 0
// 	} else {
// 		relationId = *patientRelation.RelationId
// 	}

// 	// mapping creation function
// 	CreateMapping := func(userId, patientId, roleId, relationId uint64, isSelf bool, mappingType string) error {
// 		mapping := models.SystemUserRoleMapping{
// 			UserId:      userId,
// 			RoleId:      roleId,
// 			MappingType: mappingType,
// 			IsSelf:      isSelf,
// 			PatientId:   patientId,
// 			RelationId:  relationId,
// 		}
// 		return r.roleRepo.AddSystemUserMapping(tx, &mapping)
// 	}

// 	// Handle relative, caregiver, doctor, nurse case (bi-directional)
// 	if roleName == string(constant.Relative) || roleName == string(constant.Caregiver) {
// 		// 1. Add mapping for user => patient
// 		if err := CreateMapping(userId, patientId, roleId, relationId, false, mappingType); err != nil {
// 			log.Println("Error adding user-to-patient mapping:", err)
// 			// return err
// 		}

// 		// 2. Fetch reverse relation
// 		reverseRelation, err := r.roleRepo.GetReverseRelationshipMapping(relationId, patientRelation.SourceGenderId, &userInfo.GenderId)
// 		if err != nil {
// 			log.Println("Error fetching reverse relationship:", err)
// 			return err
// 		}

// 		// Add self mapping with patient => patient
// 		if err := CreateMapping(userId, userId, roleId, 0, false, string(constant.MappingTypeS)); err != nil {
// 			log.Println("Error adding patient-to-patient self mapping:", err)
// 			// return err
// 		}
// 		// 3. Add reverse mapping for patient => user
// 		return CreateMapping(patientId, userId, roleId, reverseRelation.ReverseRelationshipId, false, relationType)
// 	}
// 	// For other roles like patient, admin etc.
// 	return CreateMapping(userId, patientId, roleId, relationId, isSelf, mappingType)

// }

func (r *RoleServiceImpl) CreateReverseMappingsForRelative(patientUserId uint64, userId uint64, genderId uint64) error {
	return r.roleRepo.CreateReverseMappingsForRelative(patientUserId, userId, genderId)
}

func (rs *RoleServiceImpl) AddUserRelativeMappings(tx *gorm.DB, userId uint64, relativeId uint64, relation models.RelationMaster, roleId uint64, registeringUser *models.SystemUser_, newUser *models.SystemUser_) error {
	log.Println("@AddUserRelativeMappings")
	log.Println("User id:", relativeId, " to be added as ", relation.RelationShip, "for userId:", userId)
	// ADD MAPPING WITH NEW USER FOR User Adding Relative
	var family *models.PatientFamilyGroup
	enabled, enabledStatusErr := rs.subscriptionRepo.GetSubscriptionShowStatus()
	if enabledStatusErr != nil {
		log.Println("Not able to fetch subscription_enabled status : ")
	}
	if enabled {
		var familyGrpErr error
		family, familyGrpErr = rs.subscriptionRepo.GetFamilyGroupByMemberID(userId)
		if familyGrpErr != nil {
			return fmt.Errorf("cannot fetch family for member: %w", familyGrpErr)
		}
	}
	var familyIdPtr *uint64
	if family != nil {
		familyIdPtr = &family.FamilyId
	}
	err := rs.roleRepo.AddSystemUserMapping(tx, &models.SystemUserRoleMapping{
		UserId:      relativeId,
		PatientId:   userId,
		RoleId:      roleId,
		IsSelf:      false,
		MappingType: string(constant.MappingTypeR),
		RelationId:  *relation.RelationId,
		FamilyId:    familyIdPtr,
	})
	if err != nil {
		log.Println("@AddUserRelativeMappings->AddSystemUserMapping1:", err)
		return err
	}
	reverseRel := utils.GetReverseRelation(int(*relation.RelationId), int(registeringUser.GenderId))
	// TODO:ADD MAPPING OF RELATIVE IN USERS
	err = rs.roleRepo.AddSystemUserMapping(tx, &models.SystemUserRoleMapping{
		UserId:      userId,
		PatientId:   relativeId,
		IsSelf:      false,
		RoleId:      roleId,
		MappingType: string(constant.MappingTypeR),
		RelationId:  uint64(*reverseRel),
		FamilyId:    familyIdPtr,
	})

	realtives, err := rs.patientService.GetRelativeList(&userId)
	if err != nil {
		return err
	}
	// UPDATE MAPPING IN ALL OTHER RELATIVES
	log.Println("List of ALL Relatives for User:", userId)
	for _, relative := range realtives {
		log.Println("Relative:", relative.RelativeId, "my_relation_id:", *reverseRel, " new_relation_id:", *relation.RelationId, " comparing_relation_id: ", relative.RelationId)
		inferredNew, inferredExisting, err := rs.roleRepo.GetInferredRelations(uint64(*reverseRel), *relation.RelationId, relative.RelationId)
		if err != nil {
			log.Println("@AddUserRelativeMappings->@GetInferredRelations:", err)
			continue
		}

		// Add Existing User TO Newly added User
		err = rs.roleRepo.AddSystemUserMapping(tx, &models.SystemUserRoleMapping{
			UserId:      relative.RelativeId,
			PatientId:   relativeId,
			IsSelf:      false,
			RoleId:      roleId,
			MappingType: string(constant.MappingTypeR),
			RelationId:  inferredNew,
			FamilyId:    familyIdPtr,
		})
		// Add New User With Existing Relative
		err = rs.roleRepo.AddSystemUserMapping(tx, &models.SystemUserRoleMapping{
			UserId:      relativeId,
			PatientId:   relative.RelativeId,
			IsSelf:      false,
			RoleId:      roleId,
			MappingType: string(constant.MappingTypeR),
			RelationId:  inferredExisting,
			FamilyId:    familyIdPtr,
		})
	}
	return nil
}

func (r *RoleServiceImpl) GetMappingTypeByPatientId(patientId *uint64) (string, error) {
	return r.roleRepo.GetMappingTypeByPatientId(patientId)
}
