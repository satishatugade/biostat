package repository

import (
	"biostat/models"

	"gorm.io/gorm"
)

type RoleRepository interface {
	GetRoleById(roleId uint) (models.RoleMaster, error)
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

func (r *RoleRepositoryImpl) GetRoleById(roleId uint) (models.RoleMaster, error) {
	var roles models.RoleMaster
	err := r.db.Where("role_id = ?", roleId).First(&roles).Error
	if err != nil {
		return roles, err
	}
	return roles, nil
}
