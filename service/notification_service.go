package service

import (
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type NotificationService interface {
	GetUserNotifications(userId uint64) ([]models.UserNotificationMapping, error)
	ScheduleReminders(email, name string, user_id uint64, config []models.ReminderConfig) error
	SendSOS(recipientEmail, familyMember, patientName, location, dateTime, deviceId string) error
	GetUserReminders(userId uint64) ([]models.UserReminder, error)
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

func (e *NotificationServiceImpl) ScheduleReminders(email, name string, user_id uint64, config []models.ReminderConfig) error {
	startDate := time.Now()
	var failedSlots []string

	for _, reminder := range config {
		reminderTimeStr := fmt.Sprintf("%s %s", startDate.Format("2006-01-02"), reminder.Time)
		reminderTime, err := time.ParseInLocation("2006-01-02 15:04", reminderTimeStr, time.Local)
		if err != nil {
			log.Printf("@ScheduleReminders Error parsing time for %s slot: %v", reminder.TimeSlot, err)
			failedSlots = append(failedSlots, fmt.Sprintf("%s: invalid time format", reminder.TimeSlot))
			continue
		}
		scheduleTime := reminderTime.Format(time.RFC3339)
		repeatUntil := reminderTime.AddDate(0, 0, reminder.DurationDays).Format(time.RFC3339)
		var medList []string
		for _, med := range reminder.Medicines {
			medList = append(medList, fmt.Sprintf("%s (%d %s)", med.Name, med.Dose, med.Unit))
		}

		scheduleBody := map[string]interface{}{
			"recipient_mail_id": email,
			"template_code":     3,
			"channels":          []string{"email"},
			"repeat_type":       "daily",
			"repeat_interval":   1,
			"repeat_times":      reminder.DurationDays,
			"repeat_until":      repeatUntil,
			"is_recurring":      true,
			"schedule_time":     scheduleTime,
			"data": map[string]interface{}{
				"userName":  name,
				"medicines": strings.Join(medList, ", "),
			},
		}
		header := map[string]string{
			"X-API-Key": os.Getenv("NOTIFY_API_KEY"),
		}
		_, data, err := utils.MakeRESTRequest("POST", os.Getenv("NOTIFY_SERVER_URL")+"/api/v1/notifications/schedule", scheduleBody, header)
		if err != nil {
			log.Printf("@ScheduleReminders ->MakeRESTRequest [%s] -> Failed to schedule notification: %v", reminder.TimeSlot, err)
			failedSlots = append(failedSlots, fmt.Sprintf("%s: scheduling failed", reminder.TimeSlot))
			continue
		}
		notifId, notifIdErr := utils.ExtractNotificationID(data)
		if notifIdErr != nil {
			log.Printf("@ScheduleReminders -> ExtractNotificationID [%s] -> Error extracting notification ID: %v", reminder.TimeSlot, notifIdErr)
			failedSlots = append(failedSlots, fmt.Sprintf("%s: could not extract ID", reminder.TimeSlot))
			continue
		}
		err = e.notificationRepo.CreateNotificationMapping(models.UserNotificationMapping{
			UserID:           user_id,
			NotificationID:   notifId,
			SourceType:       "Medicine_Reminder",
			SourceID:         "0",
			Title:            fmt.Sprintf("%s Medicine Reminder", reminder.TimeSlot),
			Message:          fmt.Sprintf("Please take following medicines without fail %s", strings.Join(medList, ", ")),
			Tags:             "medication,reminder",
			NotificationType: "scheduled",
		})
		if err != nil {
			log.Printf("@ScheduleReminders -> CreateNotificationMapping [%s] -> Failed to save mapping: %v", reminder.TimeSlot, err)
			failedSlots = append(failedSlots, fmt.Sprintf("%s: mapping failed", reminder.TimeSlot))
		}
	}
	if len(failedSlots) > 0 {
		return fmt.Errorf("failed to schedule reminders for: %s", strings.Join(failedSlots, "; "))
	}
	return nil
}

func (e *NotificationServiceImpl) SendSOS(recipientEmail, familyMember, patientName, location, dateTime, deviceId string) error {
	scheduleBody := map[string]interface{}{
		"recipient_mail_id": recipientEmail,
		"template_code":     10,
		"channels":          []string{"email"},
		"data": map[string]interface{}{
			"familyMember": familyMember,
			"patientName":  patientName,
			"dateTime":     dateTime,
			"location":     location,
			"deviceId":     deviceId,
		},
	}
	header := map[string]string{
		"X-API-Key": os.Getenv("NOTIFY_API_KEY"),
	}
	_, _, err := utils.MakeRESTRequest("POST", os.Getenv("NOTIFY_SERVER_URL")+"/api/v1/notifications/send", scheduleBody, header)
	return err
}

func (s *NotificationServiceImpl) GetUserReminders(userId uint64) ([]models.UserReminder, error) {
	var reminders []models.UserReminder
	reminderMapping, err := s.notificationRepo.GetRemindersByUserId(userId)
	if err != nil {
		return nil, err
	}
	metadataMap := make(map[uuid.UUID]models.UserNotificationMapping)
	var notificationIds []uuid.UUID

	for _, m := range reminderMapping {
		notificationIds = append(notificationIds, m.NotificationID)
		metadataMap[m.NotificationID] = m
	}
	sendBody := map[string]interface{}{
		"notification_id": notificationIds,
	}
	header := map[string]string{
		"X-API-Key": os.Getenv("NOTIFY_API_KEY"),
	}
	_, response, sendErr := utils.MakeRESTRequest("POST", os.Getenv("NOTIFY_SERVER_URL")+"/api/v1/notifications/info", sendBody, header)
	if sendErr != nil {
		return nil, sendErr
	}
	content, ok := response["content"].([]interface{})
	if !ok {
		return nil, errors.New("invalid or missing 'content' field")
	}
	for _, res := range content {
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			log.Println("@GetUserReminders=>Marshal:", res)
			continue
		}
		var reminder models.UserReminder
		if err := json.Unmarshal(jsonBytes, &reminder); err != nil {
			log.Println("@GetUserReminders:>error unmarshaling reminder:", err)
			continue
		}
		if mapping, exists := metadataMap[reminder.ReminderID]; exists {
			reminder.Title = mapping.Title
			reminder.Message = mapping.Message
		}
		reminders = append(reminders, reminder)
	}
	return reminders, nil
}
