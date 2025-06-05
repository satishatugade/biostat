package service

import (
	"biostat/models"
	"biostat/repository"
)

type NotificationService interface {
	GetUserNotifications(userId uint64) ([]models.UserNotificationMapping, error)
}

type NotificationServiceImpl struct {
	notificationRepo repository.UserNotificationRepository
}

func NewNotificationService(repo repository.UserNotificationRepository) NotificationService {
	return &NotificationServiceImpl{notificationRepo: repo}
}

func (e *NotificationServiceImpl) GetUserNotifications(userId uint64) ([]models.UserNotificationMapping, error) {
	return e.notificationRepo.GetNotificationByUserId(userId)
}
