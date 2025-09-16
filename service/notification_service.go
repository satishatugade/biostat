package service

import (
	"biostat/config"
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type NotificationService interface {
	SendLoginCredentials(systemUser models.SystemUser_, password *string, patient *models.Patient, relationship string) error
	SendConnectionMail(systemUser models.SystemUser_, patient *models.Patient, relationship string) error
	SendAppointmentMail(appointment models.AppointmentResponse, userProfile models.Patient, providerInfo interface{}) error
	SendReportResultsEmail(patientInfo *models.SystemUser_, alerts []models.TestResultAlert) error
	ShareReportEmail(recipientEmail []string, userDetails *models.SystemUser_, shortURL string) error
	SendResetPasswordMail(systemUser *models.SystemUser_, token string, recipientEmail string) error

	GetUserNotifications(userId uint64) ([]models.UserNotificationMapping, error)
	ScheduleReminders(recipeintId, name string, user_id uint64, config []models.ReminderConfig) error
	UpdateReminder(userID uint64, reminder models.UpdateReminderRequest) error
	SendSOS(recipientId, familyMember, patientName, location, dateTime, deviceId string) error
	GetUserReminders(userId uint64) ([]models.UserReminder, error)

	RegisterUserInNotify(fcmToken, phone *string, email string) (uuid.UUID, error)
	UpadateUserInNotify(recipientId string, fcmToken, email, phone *string) error
	AddUsersToNotify() error
	SaveOrUpdateNotifyCreds(userId uint64, fcmToken *string, user *models.SystemUser_) error
}

type NotificationServiceImpl struct {
	notificationRepo repository.UserNotificationRepository
	userRepo         repository.UserRepository
	apiService       ApiService
}

func NewNotificationService(repo repository.UserNotificationRepository, userRepo repository.UserRepository, apiService ApiService) NotificationService {
	return &NotificationServiceImpl{notificationRepo: repo, userRepo: userRepo, apiService: apiService}
}

func (e *NotificationServiceImpl) GetUserNotifications(userId uint64) ([]models.UserNotificationMapping, error) {
	return e.notificationRepo.GetNotificationByUserId(userId)
}

func (e *NotificationServiceImpl) ScheduleReminders(recipientId, name string, user_id uint64, reminderconfig []models.ReminderConfig) error {
	startDate := time.Now()
	var failedSlots []string

	notifyServerURL := config.PropConfig.ApiURL.NotifyServerURL
	notifyAPIKey := config.PropConfig.ApiURL.NotifyAPIKey

	for _, reminder := range reminderconfig {
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
			// "recipient_mail_id": email,
			"target_type":     "recipient_id",
			"target_value":    recipientId,
			"template_code":   3,
			"channels":        []string{"email"},
			"repeat_type":     "daily",
			"repeat_interval": 1,
			"repeat_times":    reminder.DurationDays,
			"repeat_until":    repeatUntil,
			"is_recurring":    true,
			"schedule_time":   scheduleTime,
			"data": map[string]interface{}{
				"userName":  name,
				"medicines": strings.Join(medList, ", "),
			},
		}
		header := map[string]string{
			"X-API-Key": notifyAPIKey,
		}
		_, data, err := e.apiService.MakeRESTRequest(http.MethodPost, notifyServerURL+"/api/v1/notifications/schedule", scheduleBody, header)
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

func (e *NotificationServiceImpl) UpdateReminder(userID uint64, reminder models.UpdateReminderRequest) error {
	updateBody := map[string]interface{}{
		"notification_id": reminder.NotificationID,
	}
	if reminder.RepeatType != nil {
		updateBody["repeat_type"] = *reminder.RepeatType
	}
	if reminder.RepeatInterval != nil {
		updateBody["repeat_interval"] = *reminder.RepeatInterval
	}
	if reminder.RepeatTimes != nil {
		updateBody["repeat_times"] = *reminder.RepeatTimes
	}
	if reminder.RepeatUntil != nil {
		updateBody["repeat_until"] = reminder.RepeatUntil.Format(time.RFC3339)
	}
	if reminder.NextSendAt != nil {
		updateBody["next_send_at"] = reminder.NextSendAt.Format(time.RFC3339)
	}
	if reminder.IsActive != nil {
		updateBody["is_active"] = *reminder.IsActive
	}
	jsonBytes, _ := json.MarshalIndent(updateBody, "", "  ")
	log.Println(string(jsonBytes))

	var medList []string
	for _, med := range reminder.Medicines {
		medList = append(medList, fmt.Sprintf("%s (%d %s)", med.Name, med.Dose, med.Unit))
	}

	notifyServerURL := config.PropConfig.ApiURL.NotifyServerURL
	notifyAPIKey := config.PropConfig.ApiURL.NotifyAPIKey
	headers := map[string]string{
		"X-API-Key": notifyAPIKey,
	}

	_, _, err := e.apiService.MakeRESTRequest(http.MethodPost, notifyServerURL+"/api/v1/notifications/update-schedule", updateBody, headers)
	if err != nil {
		return fmt.Errorf("failed to update notification: %v", err)
	}

	// updateErr := e.notificationRepo.UpdateNotificationMapping(models.UserNotificationMapping{
	// 	UserID:           userID,
	// 	NotificationID:   reminder.NotificationID,
	// 	Message:          fmt.Sprintf("Please take medicines: %s", strings.Join(medList, ", ")),
	// 	Tags:             "medication,reminder",
	// 	NotificationType: "scheduled",
	// })
	// if updateErr != nil {
	// 	return fmt.Errorf("failed to update mapping: %v", updateErr)
	// }

	return nil
}

func (e *NotificationServiceImpl) SendSOS(recipientId, familyMember, patientName, location, dateTime, deviceId string) error {
	scheduleBody := map[string]interface{}{
		// "recipient_mail_id": recipientEmail,
		"target_type":   "recipient_id",
		"target_value":  recipientId,
		"template_code": 10,
		"channels":      []string{"email"},
		"data": map[string]interface{}{
			"familyMember": familyMember,
			"patientName":  patientName,
			"dateTime":     dateTime,
			"location":     location,
			"deviceId":     deviceId,
		},
	}
	notifyAPIKey := config.PropConfig.ApiURL.NotifyAPIKey
	notifyServerURL := config.PropConfig.ApiURL.NotifyServerURL

	header := map[string]string{
		"X-API-Key": notifyAPIKey,
	}
	_, _, err := e.apiService.MakeRESTRequest(http.MethodPost, notifyServerURL+"/api/v1/notifications/send", scheduleBody, header)
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
		"X-API-Key": config.PropConfig.ApiURL.NotifyAPIKey,
	}
	_, response, sendErr := s.apiService.MakeRESTRequest(http.MethodPost, os.Getenv("NOTIFY_SERVER_URL")+"/api/v1/notifications/info", sendBody, header)
	if sendErr != nil {
		return nil, sendErr
	}
	content, ok := response["content"].([]interface{})
	if !ok {
		return nil, errors.New("invalid or missing 'content' field")
	}
	type notifyResponse struct {
		Notification models.UserReminder `json:"notification"`
	}
	for _, res := range content {
		jsonBytes, err := json.Marshal(res)
		if err != nil {
			log.Println("@GetUserReminders=>Marshal:", res)
			continue
		}
		var wrapper notifyResponse
		if err := json.Unmarshal(jsonBytes, &wrapper); err != nil {
			log.Println("@GetUserReminders:>error unmarshaling reminder:", err)
			continue
		}
		reminder := wrapper.Notification
		if mapping, exists := metadataMap[reminder.ReminderID]; exists {
			reminder.Title = mapping.Title
			reminder.Message = mapping.Message
		}
		reminders = append(reminders, reminder)
	}
	return reminders, nil
}

func (s *NotificationServiceImpl) RegisterUserInNotify(fcmToken, phone *string, email string) (uuid.UUID, error) {
	header := map[string]string{
		"X-API-Key": config.PropConfig.ApiURL.NotifyAPIKey,
	}
	requestBody := map[string]interface{}{}
	if fcmToken != nil && strings.TrimSpace(*fcmToken) != "" {
		requestBody["fcm_token"] = *fcmToken
	}
	if strings.TrimSpace(email) != "" {
		requestBody["email"] = email
	}
	if phone != nil && strings.TrimSpace(*phone) != "" {
		requestBody["phone"] = *phone
	}
	_, response, sendErr := s.apiService.MakeRESTRequest(http.MethodPost, os.Getenv("NOTIFY_SERVER_URL")+"/api/v1/recipient/register", requestBody, header)
	if sendErr != nil {
		return uuid.Nil, sendErr
	}
	recipient_id, err := utils.ExtractRecipientID(response)
	if err != nil {
		return uuid.Nil, err
	}
	return recipient_id, nil
}

func (s *NotificationServiceImpl) UpadateUserInNotify(recipientId string, fcmToken, email, phone *string) error {
	header := map[string]string{
		"X-API-Key": config.PropConfig.ApiURL.NotifyAPIKey,
	}
	requestBody := map[string]interface{}{}
	if strings.TrimSpace(recipientId) != "" {
		requestBody["recipient_id"] = recipientId
	}
	if fcmToken != nil && strings.TrimSpace(*fcmToken) != "" {
		requestBody["fcm_token"] = *fcmToken
	}
	if email != nil && strings.TrimSpace(*email) != "" {
		requestBody["email"] = *email
	}
	if phone != nil && strings.TrimSpace(*phone) != "" {
		requestBody["phone"] = *phone
	}
	_, _, sendErr := s.apiService.MakeRESTRequest(http.MethodPost, os.Getenv("NOTIFY_SERVER_URL")+"/api/v1/recipient/update", requestBody, header)
	if sendErr != nil {
		return sendErr
	}
	return nil
}

func (s *NotificationServiceImpl) AddUsersToNotify() error {
	users, err := s.notificationRepo.GetNotifUnregisteredUsers()
	if err != nil {
		return err
	}
	for _, user := range users {
		recipientId, err := s.RegisterUserInNotify(nil, &user.MobileNo, user.Email)
		if err != nil {
			log.Println("Failed to Register User In Notify ", err)
			continue
		}
		updateInfo := map[string]interface{}{
			"notify_id": recipientId,
		}
		err = s.userRepo.UpdateUserInfo(user.AuthUserId, updateInfo)
		if err != nil {
			log.Println("Failed to Update User Info in DB ", err)
			continue
		}
	}
	return nil
}

func (e *NotificationServiceImpl) SendLoginCredentials(systemUser models.SystemUser_, password *string, patient *models.Patient, relationship string) error {
	APPURL := config.PropConfig.ApiURL.APPURL
	RESETURL := fmt.Sprintf("%s/auth/reset-password?email=%s", APPURL, systemUser.Email)
	roleName := systemUser.RoleName
	if relationship != "" {
		roleName = fmt.Sprintf("%s ( %s )", roleName, relationship)
	}
	patientFullName := ""
	if patient != nil {
		patientFullName = patient.FirstName + " " + patient.LastName
	}
	username := systemUser.Username
	if systemUser.Email != "" {
		username = systemUser.Email
	} else if systemUser.MobileNo != "" {
		username = systemUser.MobileNo
	}
	sendBody := map[string]interface{}{
		"target_type":   "recipient_id",
		"target_value":  systemUser.NotifyId,
		"template_code": 5,
		"channels":      []string{"email"},
		"data": map[string]interface{}{
			"fullName":        systemUser.FirstName + " " + systemUser.LastName,
			"patientFullName": patientFullName,
			"roleName":        roleName,
			"username":        username,
			"password":        password,
			"loginURL":        APPURL,
			"resetURL":        RESETURL,
		},
	}
	header := map[string]string{
		"X-API-Key": config.PropConfig.ApiURL.NotifyAPIKey,
	}
	_, body, sendErr := e.apiService.MakeRESTRequest(http.MethodPost, config.PropConfig.ApiURL.NotificationSendURL, sendBody, header)
	if sendErr != nil {
		return sendErr
	}

	notifId, err := utils.ExtractNotificationID(body)
	if err == nil {
		err := e.notificationRepo.CreateNotificationMapping(models.UserNotificationMapping{
			UserID:           systemUser.UserId,
			NotificationID:   notifId,
			Title:            "Welcome to BioStack",
			Message:          "Welcome to BioStack, manage your health at one place ",
			Tags:             "welcome message",
			SourceType:       "tbl_system_user_",
			SourceID:         systemUser.Username,
			NotificationType: "one-time",
		})
		if err != nil {
			log.Println("@SendLoginCredentials: failed to save mapping")
		}
	}
	return nil
}

func (e *NotificationServiceImpl) SendConnectionMail(systemUser models.SystemUser_, patient *models.Patient, relationship string) error {
	roleName := systemUser.RoleName
	if relationship != "" {
		roleName = fmt.Sprintf("%s ( %s )", roleName, relationship)
	}
	patientFullName := ""
	if patient != nil {
		patientFullName = patient.FirstName + " " + patient.LastName
	}
	sendBody := map[string]interface{}{
		// "user_id":           systemUser.UserId,
		// "recipient_mail_id": systemUser.Email,
		"target_type":   "recipient_id",
		"target_value":  systemUser.NotifyId,
		"template_code": 11,
		"channels":      []string{"email"},
		"data": map[string]interface{}{
			"fullName":        systemUser.FirstName + " " + systemUser.LastName,
			"patientFullName": patientFullName,
			"roleName":        roleName,
		},
	}
	header := map[string]string{
		"X-API-Key": config.PropConfig.ApiURL.NotifyAPIKey,
	}
	_, _, sendErr := e.apiService.MakeRESTRequest(http.MethodPost, config.PropConfig.ApiURL.NotificationSendURL, sendBody, header)
	if sendErr != nil {
		return sendErr
	}
	return nil
}

func (e *NotificationServiceImpl) SendAppointmentMail(appointment models.AppointmentResponse, userProfile models.Patient, providerInfo interface{}) error {
	apptTime, err := time.Parse("3:04 PM", appointment.AppointmentTime)
	if err != nil {
		return fmt.Errorf("failed to parse appointment time: %w", err)
	}
	appointmentDateTime := time.Date(
		appointment.AppointmentDate.Year(),
		appointment.AppointmentDate.Month(),
		appointment.AppointmentDate.Day(),
		apptTime.Hour(),
		apptTime.Minute(),
		0, 0,
		time.Local,
	)

	start := appointmentDateTime
	end := start.Add(time.Duration(appointment.DurationMinutes) * time.Minute)

	description := ""
	location := ""
	appointmentWith := ""

	if appointment.IsInperson == 0 {
		description = "Join via Zoom: " + appointment.MeetingUrl
		location = "Online"
	} else {
		switch appointment.ProviderType {
		case "doctor":
			if p, ok := providerInfo.(models.DoctorInfo); ok {
				location = "Location: " + p.ClinicAddress
				description = "Appointment with: Dr. " + p.FirstName + " " + p.LastName
			}

		case "nurse":
			if p, ok := providerInfo.(models.NurseInfo); ok {
				location = "Location: " + p.ClinicAddress
				description = "Appointment with: Nurse " + p.FirstName + " " + p.LastName
			}

		case "lab":
			if p, ok := providerInfo.(models.LabInfo); ok {
				location = "Location: " + p.LabAddress
				description = "Appointment with: " + p.LabName
			}

		default:
			description = "Appointment with: Biostack Healthcare"
			location = "Location will be shared soon"
		}
	}

	switch appointment.ProviderType {
	case "doctor":
		if p, ok := providerInfo.(models.DoctorInfo); ok {
			appointmentWith = "Dr. " + p.FirstName + " " + p.LastName
		}

	case "nurse":
		if p, ok := providerInfo.(models.NurseInfo); ok {
			appointmentWith = "Nurse " + p.FirstName + " " + p.LastName
		}

	case "lab":
		if p, ok := providerInfo.(models.LabInfo); ok {
			appointmentWith = p.LabName
		}

	default:
		appointmentWith = "Biostack Healthcare"
	}

	calendarLink := utils.GenerateGoogleCalendarLink(
		"BioStack Appointment",
		description,
		location,
		start,
		end,
	)

	sendBody := map[string]interface{}{
		"user_id":           appointment.PatientID,
		"recipient_mail_id": userProfile.Email,
		"template_code":     6,
		"channels":          []string{"email", "whatsapp"},
		"data": map[string]interface{}{
			"userName":            userProfile.FirstName + " " + userProfile.LastName,
			"appointmentDate":     utils.FormatDateTime(&appointment.AppointmentDate),
			"appointmentTime":     appointment.AppointmentTime,
			"appointmentLocation": location,
			"calendarLink":        calendarLink,
		},
	}
	header := map[string]string{
		"X-API-Key": config.PropConfig.ApiURL.NotifyAPIKey,
	}
	sendStatus, sendData, sendErr := e.apiService.MakeRESTRequest(http.MethodPost, config.PropConfig.ApiURL.NotificationSendURL, sendBody, header)

	scheduleTime := start.Add(-30 * time.Minute).Format(time.RFC3339)
	scheduleBody := map[string]interface{}{
		"user_id":           appointment.PatientID,
		"recipient_mail_id": userProfile.Email,
		"template_code":     2,
		"channels":          []string{"email"},
		"repeat_interval":   0,
		"repeat_times":      1,
		"repeat_type":       "once",
		"is_recurring":      false,
		"schedule_time":     scheduleTime,
		"data": map[string]interface{}{
			"userName":        userProfile.FirstName + " " + userProfile.LastName,
			"doctorName":      appointmentWith,
			"appointmentDate": utils.FormatDateTime(&appointment.AppointmentDate),
			"appointmentTime": appointment.AppointmentTime,
			"meetingLink":     appointment.MeetingUrl,
		},
	}

	scheduleStatus, scheduleData, scheduleErr := e.apiService.MakeRESTRequest(http.MethodPost, os.Getenv("NOTIFY_SERVER_URL")+"/api/v1/notifications/schedule", scheduleBody, header)
	var errs []string
	if sendErr != nil {
		errs = append(errs, fmt.Sprintf("send failed: %v (status: %d, data: %v)", sendErr, sendStatus, sendData))
	}
	if scheduleErr != nil {
		errs = append(errs, fmt.Sprintf("schedule failed: %v (status: %d, data: %v)", scheduleErr, scheduleStatus, scheduleData))
	}
	log.Println(errs)
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, " | "))
	}
	notifId, notifIdErr := utils.ExtractNotificationID(scheduleData)
	if notifIdErr != nil {
		log.Println(notifIdErr)
		return fmt.Errorf("Error while extracting notification Id: " + notifIdErr.Error())
	}
	mapErr := e.notificationRepo.CreateNotificationMapping(models.UserNotificationMapping{
		UserID:           appointment.PatientID,
		NotificationID:   notifId,
		Title:            "Appointment Scheduled",
		Message:          "Appointment scheduled with " + appointmentWith,
		Tags:             "appointments,reminder",
		SourceType:       "tbl_appointment_master",
		SourceID:         fmt.Sprintf("%d", appointment.AppointmentID),
		NotificationType: "one-time,scheduled",
	})
	if mapErr != nil {
		log.Println(mapErr)
		return fmt.Errorf("Error while mapping notification Id: " + mapErr.Error())
	}
	return nil
}

func (e *NotificationServiceImpl) SendReportResultsEmail(patientInfo *models.SystemUser_, alerts []models.TestResultAlert) error {
	if len(alerts) == 0 {
		return nil
	}
	alertData := make([]map[string]interface{}, 0, len(alerts))
	for _, a := range alerts {
		alertData = append(alertData, map[string]interface{}{
			"resultDate":        a.ResultDate,
			"testName":          a.TestName,
			"testComponentName": a.TestComponentName,
			"resultValue":       a.ResultValue,
			"normalMin":         a.NormalMin,
			"normalMax":         a.NormalMax,
			"resultComment":     a.ResultComment,
		})
	}

	sendBody := map[string]interface{}{
		"user_id":           patientInfo.UserId,
		"recipient_mail_id": patientInfo.Email,
		"target_type":       "recipient_id",
		"target_value":      patientInfo.NotifyId,
		"template_code":     7,
		"channels":          []string{"email"},
		"data": map[string]interface{}{
			"fullName": patientInfo.FirstName + " " + patientInfo.LastName,
			"alerts":   alertData,
		},
	}
	header := map[string]string{
		"X-API-Key": config.PropConfig.ApiURL.NotifyAPIKey,
	}
	_, sendData, sendErr := e.apiService.MakeRESTRequest(http.MethodPost, config.PropConfig.ApiURL.NotificationSendURL, sendBody, header)
	if sendErr != nil {
		return sendErr
	}
	notifId, err := utils.ExtractNotificationID(sendData)
	if err == nil {
		err := e.notificationRepo.CreateNotificationMapping(models.UserNotificationMapping{
			UserID:           patientInfo.UserId,
			NotificationID:   notifId,
			Title:            "Health Alert: Out-of-Range Test Results",
			Message:          "We have reviewed your recent test results and noticed some values that are outside the normal range.",
			Tags:             "alert",
			SourceType:       "TestResultAlert",
			SourceID:         "0",
			NotificationType: "one-time",
		})
		if err != nil {
			log.Println("@SendReportResultsEmail: failed to save mapping")
		}
	}
	return nil
}

func (e *NotificationServiceImpl) ShareReportEmail(recipientEmail []string, userDetails *models.SystemUser_, shortURL string) error {
	var errs []string
	header := map[string]string{
		"X-API-Key": config.PropConfig.ApiURL.NotifyAPIKey,
	}
	for _, email := range recipientEmail {
		sendBody := map[string]interface{}{
			// "user_id":           userDetails.UserId,
			// "recipient_mail_id": email,
			"target_type":   "recipient_id",
			"target_value":  userDetails.NotifyId,
			"template_code": 8,
			"channels":      []string{"email"},
			"data": map[string]interface{}{
				"fullName":   userDetails.FirstName + " " + userDetails.LastName,
				"reportLink": shortURL,
			},
		}
		_, _, sendErr := e.apiService.MakeRESTRequest(http.MethodPost, config.PropConfig.ApiURL.NotificationSendURL, sendBody, header)
		if sendErr != nil {
			errs = append(errs, fmt.Sprintf("send failed for %v: %v", email, sendErr))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, " | "))
	}
	return nil
}

func (e *NotificationServiceImpl) SendResetPasswordMail(systemUser *models.SystemUser_, token string, recipientEmail string) error {
	APPURL := config.PropConfig.ApiURL.APPURL
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", APPURL, token)
	header := map[string]string{
		"X-API-Key": config.PropConfig.ApiURL.NotifyAPIKey,
	}
	sendBody := map[string]interface{}{
		// "user_id":           systemUser.UserId,
		// "recipient_mail_id": systemUser.Email,
		"target_type":   "recipient_id",
		"target_value":  systemUser.NotifyId,
		"template_code": 9,
		"channels":      []string{"email"},
		"data": map[string]interface{}{
			"fullName": systemUser.FirstName + " " + systemUser.LastName,
			"resetURL": resetURL,
		},
	}
	_, _, sendErr := e.apiService.MakeRESTRequest(http.MethodPost, config.PropConfig.ApiURL.NotificationSendURL, sendBody, header)
	return sendErr
}

func (ns *NotificationServiceImpl) SaveOrUpdateNotifyCreds(userId uint64, fcmToken *string, user *models.SystemUser_) error {
	notifyId, err := ns.notificationRepo.GetUserNotifyId(userId)
	if err != nil {
		return fmt.Errorf("error fetching notify_id: %w", err)
	}
	if notifyId == "" {
		recipientId, err := ns.RegisterUserInNotify(fcmToken, &user.MobileNo, user.Email)
		if err != nil {
			return fmt.Errorf("error registering user in notify system: %w", err)
		}
		updateInfo := map[string]interface{}{
			"notify_id": recipientId,
		}
		err = ns.userRepo.UpdateUserInfo(user.AuthUserId, updateInfo)
		if err != nil {
			return fmt.Errorf("error updating notify_id in DB: %w", err)
		}
		return nil
	}
	return ns.UpadateUserInNotify(user.NotifyId, fcmToken, &user.Email, &user.MobileNo)
}
