package service

import (
	"biostat/models"
	"biostat/utils"
	"fmt"
	"os"
	"strings"
	"time"
)

// EmailService handles sending emails
type EmailService struct {
	SMTPHost    string
	SMTPPort    string
	SenderEmail string
	SenderPass  string
}

func NewEmailService() *EmailService {
	return &EmailService{
		SMTPHost:    os.Getenv("SMTP_HOST"),
		SMTPPort:    os.Getenv("SMTP_PORT"),
		SenderEmail: os.Getenv("SMTP_EMAIL"),
		SenderPass:  os.Getenv("SMTP_PASSWORD"),
	}
}

func (e *EmailService) SendLoginCredentials(systemUser models.SystemUser_, password string, patient *models.Patient) error {
	APPURL := os.Getenv("APP_URL")
	RESETURL := fmt.Sprintf("%s/auth/reset-password?email=%s", APPURL, systemUser.Email)

	sendBody := map[string]interface{}{
		"user_id":           systemUser.UserId,
		"recipient_mail_id": systemUser.Email,
		"template_code":     5,
		"channels":          []string{"email", "whatsapp"},
		"data": map[string]interface{}{
			"fullName":        systemUser.FirstName + " " + systemUser.LastName,
			"patientFullName": patient.FirstName + " " + patient.LastName,
			"roleName":        systemUser.RoleName,
			"username":        systemUser.Username,
			"password":        password,
			"loginURL":        APPURL,
			"resetURL":        RESETURL,
		},
	}
	_, _, sendErr := utils.MakeRESTRequest("POST", os.Getenv("NOTIFY_SERVER_URL")+"/api/v1/notifications/send", sendBody, nil)
	return sendErr
}

func (e *EmailService) SendAppointmentMail(appointment models.AppointmentResponse, userProfile models.Patient, providerInfo interface{}) error {
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
		"BioStat Appointment",
		description,
		location,
		start,
		end,
	)

	sendBody := map[string]interface{}{
		"user_id":           appointment.PatientID,
		"recipient_mail_id": []string{userProfile.Email},
		"template_code":     6,
		"channels":          []string{"email", "whatsapp"},
		"data": map[string]interface{}{
			"userName":            userProfile.FirstName + " " + userProfile.LastName,
			"appointmentDate":     appointment.AppointmentDate.Format("02-01-2006"),
			"appointmentTime":     appointment.AppointmentTime,
			"appointmentLocation": location,
			"calendarLink":        calendarLink,
		},
	}
	sendStatus, sendData, sendErr := utils.MakeRESTRequest("POST", os.Getenv("NOTIFY_SERVER_URL")+"/api/v1/notifications/send", sendBody, nil)

	scheduleTime := start.Add(-30 * time.Minute).Format(time.RFC3339)
	scheduleBody := map[string]interface{}{
		"user_id":           appointment.PatientID,
		"recipient_mail_id": []string{userProfile.Email},
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
			"appointmentDate": appointment.AppointmentDate.Format("02-01-2006"),
			"appointmentTime": appointment.AppointmentTime,
			"meetingLink":     appointment.MeetingUrl,
		},
	}

	scheduleStatus, scheduleData, scheduleErr := utils.MakeRESTRequest("POST", os.Getenv("NOTIFY_SERVER_URL")+"/api/v1/notifications/schedule", scheduleBody, nil)

	var errs []string
	if sendErr != nil {
		errs = append(errs, fmt.Sprintf("send failed: %v (status: %d, data: %v)", sendErr, sendStatus, sendData))
	}
	if scheduleErr != nil {
		errs = append(errs, fmt.Sprintf("schedule failed: %v (status: %d, data: %v)", scheduleErr, scheduleStatus, scheduleData))
	}

	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, " | "))
	}
	return nil
}

func (e *EmailService) SendReportResultsEmail(patientInfo *models.SystemUser_, alerts []models.TestResultAlert) error {
	if len(alerts) == 0 {
		return nil
	}
	alertData := make([]map[string]interface{}, 0, len(alerts))
	for _, a := range alerts {
		alertData = append(alertData, map[string]interface{}{
			"resultDate":        a.ResultDate.Format("02 Jan 2006"),
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
		"template_code":     7,
		"channels":          []string{"email", "whatsapp"},
		"data": map[string]interface{}{
			"fullName": patientInfo.FirstName + " " + patientInfo.LastName,
			"alerts":   alertData,
		},
	}
	_, _, sendErr := utils.MakeRESTRequest("POST", os.Getenv("NOTIFY_SERVER_URL")+"/api/v1/notifications/send", sendBody, nil)

	return sendErr
}

func (e *EmailService) ShareReportEmail(recipientEmail []string, userDetails *models.SystemUser_, shortURL string) error {
	var errs []string
	for _, email := range recipientEmail {
		sendBody := map[string]interface{}{
			"user_id":           userDetails.UserId,
			"recipient_mail_id": email,
			"template_code":     8,
			"channels":          []string{"email", "whatsapp"},
			"data": map[string]interface{}{
				"fullName":   userDetails.FirstName + " " + userDetails.LastName,
				"reportLink": shortURL,
			},
		}
		_, _, sendErr := utils.MakeRESTRequest("POST", os.Getenv("NOTIFY_SERVER_URL")+"/api/v1/notifications/send", sendBody, nil)
		if sendErr != nil {
			errs = append(errs, fmt.Sprintf("send failed for %v: %v", email, sendErr))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, " | "))
	}
	return nil
}

func (e *EmailService) SendResetPasswordMail(systemUser *models.SystemUser_, token string, recipientEmail string) error {
	APPURL := os.Getenv("APP_URL")
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", APPURL, token)

	sendBody := map[string]interface{}{
		"user_id":           systemUser.UserId,
		"recipient_mail_id": systemUser.Email,
		"template_code":     5,
		"channels":          []string{"email", "whatsapp"},
		"data": map[string]interface{}{
			"fullName": systemUser.FirstName + " " + systemUser.LastName,
			"resetURL": resetURL,
		},
	}
	_, _, sendErr := utils.MakeRESTRequest("POST", os.Getenv("NOTIFY_SERVER_URL")+"/api/v1/notifications/send", sendBody, nil)
	return sendErr
}
