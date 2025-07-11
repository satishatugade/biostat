package repository

import (
	"biostat/constant"
	"biostat/models"
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"
)

type RoleRepository interface {
	GetRoleById(roleId uint64) (models.RoleMaster, error)
	GetRelationById(relationId uint64) (models.RelationMaster, error)
	GetRoleIdByRoleName(roleName string) (models.RoleMaster, error)
	GetSystemUserRoleMappingByUserIdMappingType(UserId uint64, mappingType *string) (models.SystemUserRoleMapping, error)
	GetSystemUserRoleMappingByPatientIdMappingType(patientId uint64, mappingType []string) ([]models.SystemUserRoleMapping, error)
	HasHOFMapping(patientId uint64, mappingType string) (bool, error)
	AddSystemUserMapping(tx *gorm.DB, systemUserMapping *models.SystemUserRoleMapping) error
	GetReverseRelationshipMapping(sourceRelationId, sourceGenderId uint64, reverseGenderId *uint64) (*models.ReverseRelationMappingResponse, error)
	CreateReverseMappingsForRelative(patientUserId, userId uint64, genderId uint64) error
	GetMappingTypeByPatientId(patientUserId *uint64) (string, error)
	GetInferredRelations(myRelationID, newRelationID, comparingRelationID uint64) (inferredNew uint64, inferredExisting uint64, err error)
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

func (r *RoleRepositoryImpl) UpdateRelationIfExists(userID, patientID uint64, relationID uint64) error {
	log.Println("Inside UpdateRelationIfExists ")
	var mapping models.SystemUserRoleMapping
	err := r.db.Where("user_id = ? AND patient_id = ? AND mapping_type = 'R' AND is_deleted = 0",
		userID, patientID).First(&mapping).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[SKIP] No mapping found for user=%d patient=%d", userID, patientID)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error fetching mapping: %w", err)
	}

	log.Printf("Before update user_id=%d → patient_id=%d updated to relation_id=%d : mapping.RelationId %d", userID, patientID, relationID, mapping.RelationId)
	if mapping.RelationId != relationID {
		log.Printf("[UPDATE] user_id=%d → patient_id=%d updated to relation_id=%d", userID, patientID, relationID)
		mapping.RelationId = relationID
		if err := r.db.Save(&mapping).Error; err != nil {
			return fmt.Errorf("failed to update relation_id: %w", err)
		}
	}
	return nil
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

func (r *RoleRepositoryImpl) GetSystemUserRoleMappingByUserIdMappingType(userId uint64, mappingType *string) (models.SystemUserRoleMapping, error) {
	var roleMapping models.SystemUserRoleMapping
	query := r.db.Model(&models.SystemUserRoleMapping{}).Where("user_id = ?", userId)

	if mappingType != nil && *mappingType != "" {
		query = query.Where("mapping_type = ?", *mappingType)
	} else {
		query = query.Where("is_self = ?", true)
	}
	if err := query.First(&roleMapping).Error; err == nil {
		return roleMapping, nil
	}
	for _, mt := range constant.FallbackMappingTypes {
		err := r.db.
			Where("user_id = ? AND mapping_type = ? AND is_self = ?", userId, string(mt), false).
			First(&roleMapping).Error
		if err == nil {
			log.Printf("Fallback role found: %s", mt)
			return roleMapping, nil
		}
	}
	return roleMapping, gorm.ErrRecordNotFound
}

func (r *RoleRepositoryImpl) GetSystemUserRoleMappingByPatientIdMappingType(patientId uint64, mappingTypes []string) ([]models.SystemUserRoleMapping, error) {
	var mappings []models.SystemUserRoleMapping

	err := r.db.
		Table("tbl_system_user_role_mapping AS m").
		Select("m.*, u.gender_id").
		Joins("JOIN tbl_system_user_ u ON u.user_id = m.user_id").
		Where("m.patient_id = ? AND m.mapping_type IN ? AND m.is_deleted = 0", patientId, mappingTypes).
		Scan(&mappings).Error

	if err != nil {
		return nil, err
	}
	return mappings, nil
}

func (r *RoleRepositoryImpl) AddSystemUserMapping(tx *gorm.DB, systemUserMapping *models.SystemUserRoleMapping) error {
	return tx.Create(systemUserMapping).Error
}

func (r *RoleRepositoryImpl) GetReverseRelationshipMapping(sourceRelationId, sourceGenderId uint64, reverseGenderId *uint64) (*models.ReverseRelationMappingResponse, error) {
	var result models.ReverseRelationMappingResponse
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

func (ps *RoleRepositoryImpl) HasHOFMapping(patientId uint64, mappingType string) (bool, error) {
	var count int64
	err := ps.db.Model(&models.SystemUserRoleMapping{}).
		Where("user_id = ? AND mapping_type = ? AND is_deleted = 0", patientId, mappingType).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *RoleRepositoryImpl) CreateReverseMappingsForRelative(userId, patientId uint64, genderId uint64) error {
	log.Printf("[CreateReverseMappingsForRelative] Starting for userId=%d, patientId=%d", userId, patientId)

	mappingTypes := []string{
		string(constant.MappingTypeR),
		string(constant.MappingTypeHOF),
	}

	allRelativeMappings, err := r.GetSystemUserRoleMappingByPatientIdMappingType(userId, mappingTypes)
	if err != nil {
		return fmt.Errorf("failed to fetch relative mappings: %w", err)
	}

	for _, mapping := range allRelativeMappings {
		log.Printf("Creating reverse mapping => userId=%d, patientId=%d, roleId=%d, relationId=%d, mappingType=%s",
			mapping.UserId, patientId, mapping.RoleId, mapping.RelationId, mapping.MappingType)
		if mapping.UserId != patientId {
			reverseRelation, err := r.GetReverseRelationshipMapping(mapping.RelationId, mapping.GenderId, &genderId)
			if err != nil {
				log.Println("Error fetching reverse relationship:", err)
				return err
			}

			reverseMapping := models.SystemUserRoleMapping{
				UserId:      mapping.UserId,
				RoleId:      mapping.RoleId,
				MappingType: mapping.MappingType,
				IsSelf:      false,
				PatientId:   patientId,
				RelationId:  reverseRelation.RelationId,
			}

			tx := r.db.Begin()

			if err := r.AddSystemUserMapping(tx, &reverseMapping); err != nil {
				log.Printf("Failed insert: userId=%d → patientId=%d, reason: %v", mapping.UserId, patientId, err)
				tx.Rollback()
				continue
			}

			if err := tx.Commit().Error; err != nil {
				log.Printf("Commit failed for userId=%d → patientId=%d, reason: %v", mapping.UserId, patientId, err)
				continue
			}

			log.Printf("Reverse mapping inserted for userId=%d → patientId=%d", mapping.UserId, patientId)
		}
	}

	log.Println("[CreateReverseMappingsForRelative] Completed")
	return nil
}

func (p *RoleRepositoryImpl) GetRelationById(relationId uint64) (models.RelationMaster, error) {
	var relation models.RelationMaster
	err := p.db.First(&relation, relationId).Error
	return relation, err
}
func (r *RoleRepositoryImpl) GetMappingTypeByPatientId(patientId *uint64) (string, error) {
	var mapping models.SystemUserRoleMapping

	err := r.db.
		Where("user_id = ? AND patient_id = ? AND is_deleted = 0", patientId, patientId).
		First(&mapping).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil // No mapping found
		}
		return "", fmt.Errorf("failed to fetch mapping type: %w", err)
	}

	return mapping.MappingType, nil
}

func (r *RoleRepositoryImpl) GetInferredRelations(myRelationID, newRelationID, comparingRelationID uint64) (uint64, uint64, error) {
	type result struct {
		InferredRelationWithNew      uint64
		InferredRelationWithExisting uint64
	}
	var res result

	err := r.db.Table("tbl_relation_inference").
		Select("inferred_relation_with_new, inferred_relation_with_existing").
		Where("my_relation_id = ? AND new_relation_id = ? AND comparing_relation_id = ?", myRelationID, newRelationID, comparingRelationID).
		First(&res).Error

	if err != nil {
		return 0, 0, err
	}

	return res.InferredRelationWithNew, res.InferredRelationWithExisting, nil
}
