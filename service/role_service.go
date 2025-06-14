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
	AddSystemUserMapping(tx *gorm.DB, patientUserId *uint64, userId uint64, userInfo *models.SystemUser_, roleId uint64, roleName string, relationId *int) error
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

// AddSystemUserMapping implements RoleService.
func (r *RoleServiceImpl) AddSystemUserMapping(tx *gorm.DB, patientUserId *uint64, userId uint64, userInfo *models.SystemUser_, roleId uint64, roleName string, relationShipId *int) error {

	roleName = strings.ToLower(roleName)
	mappingType := utils.GetMappingType(roleName)
	isSelf := roleName == "patient"

	if mappingType == "" {
		return errors.New("invalid role name")
	}
	var patientId uint64
	var relationId int
	if patientUserId == nil {
		patientId = userId
	} else {
		patientId = *patientUserId
	}
	if relationShipId == nil {
		relationId = 0
	} else {
		relationId = *relationShipId
	}
	systemUsermapping := models.SystemUserRoleMapping{
		UserId:      userId,
		RoleId:      roleId,
		MappingType: mappingType,
		IsSelf:      isSelf,
		PatientId:   patientId,
		RelationId:  relationId,
	}
	if roleName == "relative" {
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
		newRelationId, err1 := utils.MappedRelationAccordingRelationship(userInfo, relationId)
		if err1 != nil {
			log.Println("MappedRelationAccordingRelationship error occures ")
		}
		newmapping := models.SystemUserRoleMapping{
			UserId:      patientId,
			RoleId:      roleId,
			MappingType: mappingType,
			IsSelf:      false,
			PatientId:   userId,
			RelationId:  newRelationId,
		}
		return r.roleRepo.AddSystemUserMapping(tx, &newmapping)
	} else {
		return r.roleRepo.AddSystemUserMapping(tx, &systemUsermapping)
	}
}
