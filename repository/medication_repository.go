package repository

import (
	"biostat/constant"
	"biostat/models"

	"gorm.io/gorm"
)

type MedicationRepository interface {
	GetMedications(limit int, offset int) ([]models.Medication, int64, error)
	CreateMedication(medication *models.Medication) error
	UpdateMedication(medication *models.Medication, authUserId string) error
	DeleteMedication(medicationId uint64, authUserId string) error
	GetMedicationAuditRecord(medicationId, medicationAuditId uint64) ([]models.MedicationAudit, error)
	GetAllMedicationAuditRecord(page, limit int) ([]models.MedicationAudit, int64, error)
	AddDiseaseMedicationMapping(mapping *models.DiseaseMedicationMapping) error
}

type MedicationRepositoryImpl struct {
	db *gorm.DB
}

func NewMedicationRepository(db *gorm.DB) MedicationRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &MedicationRepositoryImpl{db: db}
}

// GetMedications implements MedicationRepository.

func (m *MedicationRepositoryImpl) GetMedications(limit int, offset int) ([]models.Medication, int64, error) {
	var medications []models.Medication
	var totalRecords int64

	// Count total records for pagination
	if err := m.db.Model(&models.Medication{}).Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	// Fetch medications with their associated types
	if err := m.db.Preload("MedicationTypes").Limit(limit).Offset(offset).Find(&medications).Error; err != nil {
		return nil, 0, err
	}

	return medications, totalRecords, nil
}

// CreateMedication implements MedicationRepository.
func (m *MedicationRepositoryImpl) CreateMedication(medication *models.Medication) error {
	return m.db.Create(medication).Error
}

func (r *MedicationRepositoryImpl) SaveMedicationAudit(tx *gorm.DB, medication *models.Medication, operationType, authUserId string) error {
	medAudit := models.MedicationAudit{
		MedicationId:   medication.MedicationId,
		MedicationName: medication.MedicationName,
		MedicationCode: medication.MedicationCode,
		Description:    medication.Description,
		OperationType:  operationType,
		IsDeleted:      medication.IsDeleted,
		CreatedBy:      medication.CreatedBy,
		UpdatedBy:      authUserId,
	}

	return tx.Create(&medAudit).Error
}

func (r *MedicationRepositoryImpl) SaveMedicationTypeAudit(tx *gorm.DB, medType *models.MedicationType, operationType string, authUserId string) error {
	typeAudit := models.MedicationTypeAudit{
		DosageId:           medType.DosageId,
		MedicationId:       medType.MedicationId,
		MedicationType:     medType.MedicationType,
		UnitValue:          medType.UnitValue,
		UnitType:           medType.UnitType,
		MedicationCost:     medType.MedicationCost,
		MedicationImageURL: medType.MedicationImageURL,
		OperationType:      operationType,
		UpdatedBy:          authUserId,
	}

	return tx.Create(&typeAudit).Error
}

func (mr *MedicationRepositoryImpl) UpdateMedication(medication *models.Medication, authUserId string) error {
	return mr.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Medication{}).
			Where("medication_id = ?", medication.MedicationId).
			Updates(medication).Error; err != nil {
			return err
		}
		if err := mr.SaveMedicationAudit(tx, medication, constant.UPDATE, authUserId); err != nil {
			return err
		}
		for _, medType := range medication.MedicationTypes {
			if medType.DosageId > 0 {
				if err := tx.Model(&models.MedicationType{}).
					Where("dosage_id = ? AND medication_id = ?", medType.DosageId, medication.MedicationId).
					Updates(medType).Error; err != nil {
					return err
				}
			} else {
				medType.MedicationId = medication.MedicationId
				if err := tx.Create(&medType).Error; err != nil {
					return err
				}
			}
			medType.MedicationId = medication.MedicationId
			if err := mr.SaveMedicationTypeAudit(tx, &medType, constant.UPDATE, authUserId); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *MedicationRepositoryImpl) DeleteMedication(medicationId uint64, authUserId string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var med *models.Medication
		if err := tx.First(&med, "medication_id = ?", medicationId).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Medication{}).
			Where("medication_id = ?", medicationId).
			Update("is_deleted", 1).Error; err != nil {
			return err
		}

		if err := s.SaveMedicationAudit(tx, med, constant.DELETE, authUserId); err != nil {
			return err
		}
		return nil
	})
}

func (repo *MedicationRepositoryImpl) GetAllMedicationAuditRecord(page, limit int) ([]models.MedicationAudit, int64, error) {
	var auditLogs []models.MedicationAudit
	var totalRecords int64

	repo.db.Model(&models.MedicationAudit{}).Count(&totalRecords)

	err := repo.db.
		Limit(limit).
		Offset((page - 1) * limit).
		Order("medication_audit_id DESC").
		Find(&auditLogs).Error

	return auditLogs, totalRecords, err
}

func (repo *MedicationRepositoryImpl) GetMedicationAuditRecord(medicationId, medicationAuditId uint64) ([]models.MedicationAudit, error) {
	var auditLogs []models.MedicationAudit
	query := repo.db

	if medicationId != 0 {
		query = query.Where("medication_id = ?", medicationId)
	}
	if medicationAuditId != 0 {
		query = query.Where("medication_audit_id = ?", medicationAuditId)
	}

	err := query.Order("medication_audit_id DESC").Find(&auditLogs).Error
	return auditLogs, err
}
func (r *MedicationRepositoryImpl) AddDiseaseMedicationMapping(mapping *models.DiseaseMedicationMapping) error {
	return r.db.Create(mapping).Error
}
