package repository

import (
	"biostat/models"
	"log"

	"gorm.io/gorm"
)

type RoleRepository interface {
	GetRoleById(roleId uint64) (models.RoleMaster, error)
	GetRoleIdByRoleName(roleName string) (models.RoleMaster, error)
	GetRoleByUserId(UserId uint64, mappingType *string) (models.SystemUserRoleMapping, error)
	AddSystemUserMapping(*models.SystemUserRoleMapping) error
}

type RoleRepositoryImpl struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &RoleRepositoryImpl{db: db}
}

func (r *RoleRepositoryImpl) GetRoleById(roleId uint64) (models.RoleMaster, error) {
	var roles models.RoleMaster
	err := r.db.Where("role_id = ?", roleId).First(&roles).Error
	if err != nil {
		return roles, err
	}
	return roles, nil
}

func (r *RoleRepositoryImpl) GetRoleIdByRoleName(roleName string) (models.RoleMaster, error) {
	var roles models.RoleMaster
	err := r.db.Where("role_name = ?", roleName).First(&roles).Error
	if err != nil {
		return roles, err
	}
	return roles, nil
}

func (r *RoleRepositoryImpl) GetRoleByUserId(UserId uint64, mappingType *string) (models.SystemUserRoleMapping, error) {
	var rolesMapping models.SystemUserRoleMapping
	query := r.db.Where("user_id = ?", UserId)

	if mappingType != nil && *mappingType != "" {
		query = query.Where("mapping_type = ?", *mappingType)
	} else {
		// If mappingType is nil or empty, check is_self condition and do login as patient
		query = query.Where("is_self = ?", true)
	}

	err := query.First(&rolesMapping).Error
	if err != nil {
		return rolesMapping, err
	}

	log.Println("SystemUserRoleMapping:", rolesMapping)
	return rolesMapping, nil
}

// AddSystemUserMapping implements RoleRepository.
func (r *RoleRepositoryImpl) AddSystemUserMapping(systemUserMapping *models.SystemUserRoleMapping) error {
	if err := r.db.Create(systemUserMapping).Error; err != nil {
		return err
	}
	return nil
}
