package service

import (
	"biostat/constant"
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"errors"
	"log"
	"strings"

	"gorm.io/gorm"
)

type RoleService interface {
	GetRoleById(roleId uint64) (models.RoleMaster, error)
	HasHOFMapping(patientId uint64, mappingType string) (bool, error)
	GetRoleIdByRoleName(roleName string) (models.RoleMaster, error)
	GetRoleByUserId(UserId uint64, mappingType *string) (models.RoleMaster, error)
	AddSystemUserMapping(tx *gorm.DB, patientUserId *uint64, userId uint64, userInfo *models.SystemUser_, roleId uint64, roleName string, patientRelation *models.RelationMaster) error
	CreateReverseMappingsForRelative(patientUserId, userId uint64, genderId uint64) error
	AddUserRelativeMappings(tx *gorm.DB, userId uint64, relativeId uint64, relation models.RelationMaster, roleId uint64, registeringUser *models.SystemUser_, newUser *models.SystemUser_) error
	GetMappingTypeByPatientId(patientUserId *uint64) (string, error)
}

type RoleServiceImpl struct {
	roleRepo       repository.RoleRepository
	patientService PatientService
}

func NewRoleService(repo repository.RoleRepository, patientService PatientService) RoleService {
	return &RoleServiceImpl{roleRepo: repo, patientService: patientService}
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

func (r *RoleServiceImpl) AddSystemUserMapping(tx *gorm.DB, patientUserId *uint64, userId uint64, userInfo *models.SystemUser_, roleId uint64, roleName string, patientRelation *models.RelationMaster) error {
	roleName = strings.ToLower(roleName)
	mappingType := utils.GetMappingTypeByRoleName(roleName, nil)
	isSelf := roleName == "patient"
	if mappingType == "" {
		return errors.New("invalid role name")
	}
	var patientId uint64
	var relationId uint64
	if patientUserId == nil {
		patientId = userId
	} else {
		patientId = *patientUserId
	}
	if patientRelation == nil {
		relationId = 0
	} else {
		relationId = *patientRelation.RelationId
	}
	systemUsermapping := models.SystemUserRoleMapping{
		UserId:      userId,
		RoleId:      roleId,
		MappingType: mappingType,
		IsSelf:      isSelf,
		PatientId:   patientId,
		RelationId:  relationId,
	}
	if roleName == "relative" || roleName == "caregiver" {
		usermapping := models.SystemUserRoleMapping{
			UserId:      userId,
			RoleId:      roleId,
			MappingType: mappingType,
			IsSelf:      false,
			PatientId:   patientId,
			RelationId:  relationId,
		}
		err := r.roleRepo.AddSystemUserMapping(tx, &usermapping)
		if err != nil {
			log.Println("error occures while adding AddSystemUserMapping")
		}
		log.Println("Inside MappedRelationAccordingRelationship : relationId : ", relationId)
		relationshipMapping, err := r.roleRepo.GetReverseRelationshipMapping(relationId, patientRelation.SourceGenderId, &userInfo.GenderId)
		if err != nil {
			log.Println("GetReverseRelationshipMapping error occures ", err)
		}
		log.Println("relationshipMapping : ", relationshipMapping)
		newmapping := models.SystemUserRoleMapping{
			UserId:      patientId,
			RoleId:      roleId,
			MappingType: mappingType,
			IsSelf:      false,
			PatientId:   userId,
			RelationId:  relationshipMapping.ReverseRelationshipId,
		}
		return r.roleRepo.AddSystemUserMapping(tx, &newmapping)
	} else {
		return r.roleRepo.AddSystemUserMapping(tx, &systemUsermapping)
	}
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
	err := rs.roleRepo.AddSystemUserMapping(tx, &models.SystemUserRoleMapping{
		UserId:      relativeId,
		PatientId:   userId,
		RoleId:      roleId,
		IsSelf:      false,
		MappingType: string(constant.MappingTypeR),
		RelationId:  *relation.RelationId,
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
		})
		// Add New User With Existing Relative
		err = rs.roleRepo.AddSystemUserMapping(tx, &models.SystemUserRoleMapping{
			UserId:      relativeId,
			PatientId:   relative.RelativeId,
			IsSelf:      false,
			RoleId:      roleId,
			MappingType: string(constant.MappingTypeR),
			RelationId:  inferredExisting,
		})
	}
	return nil
}

func (r *RoleServiceImpl) GetMappingTypeByPatientId(patientId *uint64) (string, error) {
	return r.roleRepo.GetMappingTypeByPatientId(patientId)
}
