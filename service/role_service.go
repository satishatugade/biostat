package service

import (
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
	GetRoleIdByRoleName(roleName string) (models.RoleMaster, error)
	GetRoleByUserId(UserId uint64, mappingType *string) (models.RoleMaster, error)
	AddSystemUserMapping(tx *gorm.DB, patientUserId *uint64, userId uint64, userInfo *models.SystemUser_, roleId uint64, roleName string, patientRelation *models.PatientRelation) error
}

type RoleServiceImpl struct {
	roleRepo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) RoleService {
	return &RoleServiceImpl{roleRepo: repo}
}

// GetRoleById implements RoleService.
func (r *RoleServiceImpl) GetRoleById(roleId uint64) (models.RoleMaster, error) {
	return r.roleRepo.GetRoleById(roleId)

}

func (r *RoleServiceImpl) GetRoleIdByRoleName(roleName string) (models.RoleMaster, error) {
	return r.roleRepo.GetRoleIdByRoleName(roleName)

}

func (r *RoleServiceImpl) GetRoleByUserId(UserId uint64, mappingType *string) (models.RoleMaster, error) {
	role, err := r.roleRepo.GetRoleByUserId(UserId, mappingType)
	if err != nil {
		return models.RoleMaster{}, err
	}
	return r.roleRepo.GetRoleById(role.RoleId)
}

func (r *RoleServiceImpl) AddSystemUserMapping(tx *gorm.DB, patientUserId *uint64, userId uint64, userInfo *models.SystemUser_, roleId uint64, roleName string, patientRelation *models.PatientRelation) error {
	roleName = strings.ToLower(roleName)
	mappingType := utils.GetMappingType(roleName, nil)
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
