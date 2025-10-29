package repository

import (
	"biostat/models"
	"database/sql"
	"log"
	"time"

	"gorm.io/gorm"
)

type UserNotificationRepository interface {
	CreateNotificationMapping(mapping models.UserNotificationMapping) error
	UpdateNotificationMapping(mapping models.UserNotificationMapping) error
	GetNotificationByUserId(userId uint64) ([]models.UserNotificationMapping, error)
	GetRemindersByUserId(userId uint64) ([]models.UserNotificationMapping, error)
	GetNotifUnregisteredUsers() ([]models.SystemUser_, error)
	GetMailUnregisteredUsers() ([]models.SystemUser_, error)
	GetUserNotifyId(userId uint64) (string, error)
}

type UserNotificationRepositoryImpl struct {
	db *gorm.DB
}

func NewUserNotificationRepository(db *gorm.DB) UserNotificationRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &UserNotificationRepositoryImpl{db: db}
}

func (r *UserNotificationRepositoryImpl) CreateNotificationMapping(mapping models.UserNotificationMapping) error {
	log.Println(mapping)
	return r.db.Create(&mapping).Error
}

func (r *UserNotificationRepositoryImpl) UpdateNotificationMapping(mapping models.UserNotificationMapping) error {
	return r.db.Model(&models.UserNotificationMapping{}).
		Where("notification_id = ? AND user_id = ?", mapping.NotificationID, mapping.UserID).
		Updates(map[string]interface{}{
			"title":             mapping.Title,
			"message":           mapping.Message,
			"tags":              mapping.Tags,
			"notification_type": mapping.NotificationType,
			"updated_at":        time.Now(),
		}).Error
}

func (r *UserNotificationRepositoryImpl) GetNotificationByUserId(userId uint64) ([]models.UserNotificationMapping, error) {
	var notifs []models.UserNotificationMapping
	err := r.db.Where("user_id = ?", userId).Order("created_at DESC").Find(&notifs).Error
	return notifs, err
}

func (r *UserNotificationRepositoryImpl) GetRemindersByUserId(userId uint64) ([]models.UserNotificationMapping, error) {
	var reminders []models.UserNotificationMapping
	err := r.db.Where("user_id = ? and tags='medication,reminder'", userId).Find(&reminders).Error
	return reminders, err
}

func (r *UserNotificationRepositoryImpl) GetNotifUnregisteredUsers() ([]models.SystemUser_, error) {
	var users []models.SystemUser_
	err := r.db.Where("notify_id IS NULL").Find(&users).Error
	return users, err
}

func (r *UserNotificationRepositoryImpl) GetMailUnregisteredUsers() ([]models.SystemUser_, error) {
	var users []models.SystemUser_
	err := r.db.Where("(biomail_id IS NULL OR biomail_id = '') AND mobile_no != ''").Find(&users).Error
	return users, err
}

func (r *UserNotificationRepositoryImpl) GetUserNotifyId(userId uint64) (string, error) {
	var notifyId sql.NullString
	err := r.db.Model(&models.SystemUser_{}).
		Select("notify_id").Where("user_id = ?", userId).
		Scan(&notifyId).Error
	if err != nil {
		return "", err
	}
	if notifyId.Valid {
		return notifyId.String, nil
	}
	return "", nil
}
