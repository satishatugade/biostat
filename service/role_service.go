package service

import (
	"biostat/models"
	"biostat/repository"
)

type RoleService interface {
	GetRoleById(roleId uint) (models.RoleMaster, error)
}

type RoleServiceImpl struct {
	roleRepo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) RoleService {
	return &RoleServiceImpl{roleRepo: repo}
}

// GetRoleById implements RoleService.
func (r *RoleServiceImpl) GetRoleById(roleId uint) (models.RoleMaster, error) {
	return r.roleRepo.GetRoleById(roleId)

}
