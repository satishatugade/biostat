package service

import (
	"biostat/config"
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

type EmailService interface {
	SendLoginCredentials(systemUser models.SystemUser_, password *string, patient *models.Patient, relationship string) error
	SendConnectionMail(systemUser models.SystemUser_, patient *models.Patient, relationship string) error
	SendAppointmentMail(appointment models.AppointmentResponse, userProfile models.Patient, providerInfo interface{}) error
	SendReportResultsEmail(patientInfo *models.SystemUser_, alerts []models.TestResultAlert) error
	ShareReportEmail(recipientEmail []string, userDetails *models.SystemUser_, shortURL string) error
	SendResetPasswordMail(systemUser *models.SystemUser_, token string, recipientEmail string) error
}

type EmailServiceImpl struct {
	notificationRepo repository.UserNotificationRepository
	userRepo         repository.UserRepository
	apiService       ApiService
}

func NewEmailService(notificationRepo repository.UserNotificationRepository, userRepo repository.UserRepository, apiService ApiService) *EmailServiceImpl {
	return &EmailServiceImpl{notificationRepo: notificationRepo, userRepo: userRepo, apiService: apiService}
}

func (e *EmailServiceImpl) SendLoginCredentials(systemUser models.SystemUser_, password *string, patient *models.Patient, relationship string) error {
	APPURL := os.Getenv("APP_URL")
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

func (e *EmailServiceImpl) SendConnectionMail(systemUser models.SystemUser_, patient *models.Patient, relationship string) error {
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

func (e *EmailServiceImpl) SendAppointmentMail(appointment models.AppointmentResponse, userProfile models.Patient, providerInfo interface{}) error {
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

func (e *EmailServiceImpl) SendReportResultsEmail(patientInfo *models.SystemUser_, alerts []models.TestResultAlert) error {
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

func (e *EmailServiceImpl) ShareReportEmail(recipientEmail []string, userDetails *models.SystemUser_, shortURL string) error {
	var errs []string
	header := map[string]string{
		"X-API-Key": config.PropConfig.ApiURL.NotifyAPIKey,
	}
	notifyServerURL := config.PropConfig.ApiURL.NotifyServerURL
	for _, email := range recipientEmail {
		userInfo, err := e.userRepo.GetUserInfoByEmailId(strings.ToLower(email))
		if err != nil {
			log.Println("User Info not found with this email, Notify Id is required to send Email : ", email)
			config.Log.Warn("User Info not found with this email, Notify Id is required to send Email : ", zap.String("Email", email))
			continue
		}
		sendBody := map[string]interface{}{
			// "user_id":           userDetails.UserId,
			// "recipient_mail_id": email,
			"target_type":   "recipient_id",
			"target_value":  userInfo.NotifyId,
			"template_code": 8,
			"channels":      []string{"email"},
			"data": map[string]interface{}{
				"fullName":   userDetails.FirstName + " " + userDetails.LastName,
				"reportLink": shortURL,
			},
		}
		_, _, sendErr := e.apiService.MakeRESTRequest(http.MethodPost, notifyServerURL+"/api/v1/notifications/send", sendBody, header)
		if sendErr != nil {
			log.Println("Error sending share report email ", sendErr)
			errs = append(errs, fmt.Sprintf("send failed for %v: %v", email, sendErr))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, " | "))
	}
	return nil
}

func (e *EmailServiceImpl) SendResetPasswordMail(systemUser *models.SystemUser_, token string, recipientEmail string) error {
	APPURL := os.Getenv("APP_URL")
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
