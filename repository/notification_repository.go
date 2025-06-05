package repository

import (
	"biostat/models"
	"log"

	"gorm.io/gorm"
)

type UserNotificationRepository interface {
	CreateNotificationMapping(mapping models.UserNotificationMapping) error
	GetNotificationByUserId(userId uint64) ([]models.UserNotificationMapping, error)
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

func (r *UserNotificationRepositoryImpl) GetNotificationByUserId(userId uint64) ([]models.UserNotificationMapping, error) {
	var notifs []models.UserNotificationMapping
	err := r.db.Where("user_id = ?", userId).Find(&notifs).Error
	return notifs, err
}
