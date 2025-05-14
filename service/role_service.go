package service

import (
	"biostat/models"
	"biostat/repository"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type RoleService interface {
	GetRoleById(roleId uint64) (models.RoleMaster, error)
	GetRoleIdByRoleName(roleName string) (models.RoleMaster, error)
	GetRoleByUserId(UserId uint64, mappingType *string) (models.RoleMaster, error)
	AddSystemUserMapping(tx *gorm.DB, patientUserId *uint64, userId, roleId uint64, roleName string, relationId *int) error
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
func (r *RoleServiceImpl) AddSystemUserMapping(tx *gorm.DB, patientUserId *uint64, userId uint64, roleId uint64, roleName string, relationShipId *int) error {

	roleName = strings.ToLower(roleName)
	mappingType := map[string]string{"patient": "S", "doctor": "D", "nurse": "N", "relative": "R", "caregiver": "C", "admin": "A", "pharmacist": "P"}[roleName]
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
	return r.roleRepo.AddSystemUserMapping(tx, &systemUsermapping)
}
