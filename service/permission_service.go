package service

import (
	"biostat/models"
	"biostat/repository"
	"log"
)

type PermissionService interface {
	ManagePermission(input models.ManagePermissionsRequest) error
	GetAllPermissions() ([]models.PermissionMaster, error)
	AssignMultiplePermissions(userID, relativeID uint64, permissions map[string]bool) error
}

type PermissionServiceImpl struct {
	permissionRepo repository.PermissionRepository
}

func NewPermissionService(repo repository.PermissionRepository) PermissionService {
	return &PermissionServiceImpl{permissionRepo: repo}
}

func (s *PermissionServiceImpl) GetAllPermissions() ([]models.PermissionMaster, error) {
	return s.permissionRepo.GetAllPermissions()
}

func (s *PermissionServiceImpl) ManagePermission(input models.ManagePermissionsRequest) error {
	for _, rel := range input.RelativePermissions {
		for _, perm := range rel.Permissions {
			// Lookup permission_id
			permMaster, err := s.permissionRepo.GetPermissionByCode(perm.Code)
			if err != nil {
				log.Printf("Invalid permission code: %s", perm.Code)
				continue
			}

			err = s.permissionRepo.UpsertUserRelativePermission(
				perm.GrantedTo,
				rel.RelativeID,
				permMaster.PermissionID,
				perm.Value,
			)
			if err != nil {
				log.Println("Failed to upsert permission:", err)
			}
		}
	}
	return nil
}

func (s *PermissionServiceImpl) AssignPermission(userID, relativeID uint64, code string, granted bool) error {
	perm, err := s.permissionRepo.GetPermissionByCode(code)
	if err != nil {
		return err
	}
	return s.permissionRepo.GrantPermission(userID, relativeID, perm.PermissionID, granted)
}

func (s *PermissionServiceImpl) AssignMultiplePermissions(userID, relativeID uint64, permissions map[string]bool) error {
	for code, value := range permissions {
		perm, err := s.permissionRepo.GetPermissionByCode(code)
		if err != nil {
			log.Printf("Permission code '%s' not found", code)
			continue
		}

		exists, currentValue, _ := s.permissionRepo.CheckPermissionValue(userID, relativeID, perm.PermissionID)
		if !exists {
			err := s.permissionRepo.GrantPermission(userID, relativeID, perm.PermissionID, value)
			if err != nil {
				log.Printf("Failed to create mapping for permission %s: %v", code, err)
			}
		} else if currentValue != value {
			err := s.permissionRepo.UpdatePermissionValue(userID, relativeID, perm.PermissionID, value)
			if err != nil {
				log.Printf("Failed to update permission %s: %v", code, err)
			}
		}
	}
	return nil
}
