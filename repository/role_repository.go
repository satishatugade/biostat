package repository

import (
	"biostat/models"
	"fmt"
	"log"

	"gorm.io/gorm"
)

type RoleRepository interface {
	GetRoleById(roleId uint64) (models.RoleMaster, error)
	GetRoleIdByRoleName(roleName string) (models.RoleMaster, error)
	GetRoleByUserId(UserId uint64, mappingType *string) (models.SystemUserRoleMapping, error)
	AddSystemUserMapping(tx *gorm.DB, systemUserMapping *models.SystemUserRoleMapping) error
	GetReverseRelationshipMapping(sourceRelationId, sourceGenderId uint64, reverseGenderId *uint64) (*models.ReverseRelationMappingResponse, error)
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
		log.Println("SystemUserRoleMapping not found:")
		mappingTypes := []string{"A", "D", "R", "C"}
		for _, mt := range mappingTypes {
			err = r.db.Where("user_id = ? AND mapping_type = ? AND is_self = ? ", UserId, mt, false).First(&rolesMapping).Error
			if err == nil {
				log.Println("SystemUserRoleMapping found with fallback mapping_type =", mt)
				return rolesMapping, nil
			}
		}
		return rolesMapping, err
	}

	log.Println("SystemUserRoleMapping:", rolesMapping)
	return rolesMapping, nil
}

// AddSystemUserMapping implements RoleRepository.
func (r *RoleRepositoryImpl) AddSystemUserMapping(tx *gorm.DB, systemUserMapping *models.SystemUserRoleMapping) error {
	return tx.Create(systemUserMapping).Error
}

func (r *RoleRepositoryImpl) GetReverseRelationshipMapping(sourceRelationId, sourceGenderId uint64, reverseGenderId *uint64) (*models.ReverseRelationMappingResponse, error) {
	var result models.ReverseRelationMappingResponse
	fmt.Println("GetReverseRelationshipMapping reverseGenderId ", *reverseGenderId)
	err := r.db.Debug().Raw(`
		SELECT
			rm.relation_id AS relation_id,
			rrm.source_gender_id AS source_gender_id,
			rrm.reverse_relation_gender_id AS reverse_relation_gender_id,
			rmg.relationship AS reverse_relationship_name,
			rrm.reverse_relationship_id AS reverse_relationship_id
		FROM
			tbl_reverse_relationship_mapping rrm
		JOIN
			tbl_relation_master rm 
			ON rm.relation_id = rrm.source_relation_id
		JOIN
			tbl_relation_master rmg 
			ON rmg.relation_id = rrm.reverse_relationship_id
			AND rmg.source_gender_id = rrm.reverse_relation_gender_id
		WHERE
			rrm.source_relation_id = ?
			AND rrm.source_gender_id = ?
			AND rrm.reverse_relation_gender_id = ?`, sourceRelationId, sourceGenderId, *reverseGenderId).Scan(&result).Error

	if err != nil {
		return nil, err
	}
	return &result, nil
}
