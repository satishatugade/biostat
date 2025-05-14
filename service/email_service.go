package service

import (
	"biostat/models"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

// EmailService handles sending emails
type EmailService struct {
	SMTPHost    string
	SMTPPort    string
	SenderEmail string
	SenderPass  string
}

// NewEmailService initializes the email service
func NewEmailService() *EmailService {
	return &EmailService{
		SMTPHost:    os.Getenv("SMTP_HOST"), // e.g., "smtp.gmail.com"
		SMTPPort:    os.Getenv("SMTP_PORT"), // e.g., "587"
		SenderEmail: os.Getenv("SMTP_EMAIL"),
		SenderPass:  os.Getenv("SMTP_PASSWORD"),
	}
}

func (e *EmailService) SendLoginCredentials(systemUser models.SystemUser_, password string, patient *models.Patient) error {
	auth := smtp.PlainAuth("", e.SenderEmail, e.SenderPass, e.SMTPHost)
	APPURL := os.Getenv("APP_URL")                                                  // Application Login URL
	RESETURL := fmt.Sprintf("%s/reset-password?email=%s", APPURL, systemUser.Email) // Password Reset URL

	var additionalInfo string
	if patient != nil {
		additionalInfo = fmt.Sprintf(
			"<p style='text-align: left;'>You’ve been successfully added as a <strong>%s</strong> by patient <strong>%s %s</strong> in the Biostat Healthcare System.</p>",
			systemUser.RoleName, patient.FirstName, patient.LastName,
		)
	}

	// HTML Email Body with Login and Reset Password Links
	message := fmt.Sprintf("Subject: Welcome to our Biostat Healthcare System\r\n"+
		"From: Biostat Healthcare <%s>\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n"+
		"<html><body>"+
		"<div style='text-align: center;'>"+
		"<h3>Welcome to our Biostat Healthcare System</h3>"+
		"<p style='text-align: left;'>Hello %s %s,</p>"+
		"%s"+
		"<p><strong>Username:</strong> %s</p>"+
		"<p><strong>Password:</strong> %s</p>"+
		"<p>Please change your password after logging in.</p>"+
		"<p><a href='%s' style='background-color:blue;color:white;padding:10px 20px;text-decoration:none;border-radius:5px;'>Login Here</a></p>"+
		"<p>If you need to reset your password, click below:</p>"+
		"<p><a href='%s' style='background-color:red;color:white;padding:10px 20px;text-decoration:none;border-radius:5px;'>Reset Password</a></p>"+
		"<div><br>Best Regards,<br>Biostat Healthcare Team</div>"+
		"</div></body></html>",
		e.SenderEmail,
		systemUser.FirstName, systemUser.LastName,
		additionalInfo,
		systemUser.Username,
		password,
		APPURL,
		RESETURL,
	)

	return smtp.SendMail(e.SMTPHost+":"+e.SMTPPort, auth, e.SenderEmail, []string{systemUser.Email}, []byte(message))
}

func (e *EmailService) SendAppointmentMail(appointment models.AppointmentResponse, userProfile interface{}) error {
	auth := smtp.PlainAuth("", e.SenderEmail, e.SenderPass, e.SMTPHost)
	log.Println("userProfile ", userProfile)

	profileMap, ok := userProfile.(models.Patient)
	if !ok {
		return fmt.Errorf("invalid userProfile format")
	}
	email := profileMap.Email
	fullName := profileMap.FirstName + " " + profileMap.LastName

	var providerName string
	var providerEmail string
	switch info := appointment.ProviderInfo.(type) {
	case map[string]interface{}:
		firstName, _ := info["first_name"].(string)
		lastName, _ := info["last_name"].(string)
		email, _ := info["email"].(string)
		providerEmail = email
		providerName = fmt.Sprintf("%s %s", firstName, lastName)
	default:
		providerName = "Biostat healthcare"
		providerEmail = "satish123@yopmail.com"
	}

	subject := fmt.Sprintf("Appointment Scheduled with %s", strings.Title(appointment.ProviderType))
	to := []string{email, providerEmail}
	mode := map[int]string{0: "Online", 1: "In-Person"}[appointment.IsInperson]

	message := fmt.Sprintf("Subject: %s\r\n"+
		"From: Biostat Healthcare <%s>\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n"+
		"<html><body style='font-family: Arial, sans-serif; color: #333;'>"+
		"<div style='max-width: 600px; margin: auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 10px;'>"+
		"<div style='text-align: center; margin-bottom: 20px;'>"+
		"<h2 style='color: #2d6cdf;'>Appointment Confirmation</h2>"+
		"</div>"+
		"<p>Hi <strong>%s</strong>,</p>"+
		"<p>Your appointment has been successfully scheduled. Please find the details below:</p>"+
		"<table style='width: 100%%; border-collapse: collapse;'>"+
		"<tr><td style='padding: 8px;'><strong>Appointment Id:</strong></td><td style='padding: 8px;'>%d</td></tr>"+
		"<tr><td style='padding: 8px;'><strong>Date:</strong></td><td style='padding: 8px;'>%s</td></tr>"+
		"<tr><td style='padding: 8px;'><strong>Time:</strong></td><td style='padding: 8px;'>%s</td></tr>"+
		"<tr><td style='padding: 8px;'><strong>Mode:</strong></td><td style='padding: 8px;'>%s</td></tr>"+
		"<tr><td style='padding: 8px;'><strong>Scheduled with:</strong></td><td style='padding: 8px;'>%s (%s)</td></tr>"+
		"</table>",
		subject,
		e.SenderEmail,
		fullName,
		appointment.AppointmentID,
		appointment.AppointmentDate.Format("02 Jan 2006"),
		appointment.AppointmentTime,
		mode,
		providerName, strings.Title(appointment.ProviderType),
	)

	if appointment.IsInperson == 0 && appointment.MeetingUrl != "" {
		message += fmt.Sprintf(
			"<div style='margin-top: 20px; text-align: center;'>"+
				"<a href='%s' style='display: inline-block; background-color: #28a745; color: #fff; padding: 12px 25px; text-decoration: none; border-radius: 5px; font-weight: bold;'>Join Meeting</a>"+
				"</div>",
			appointment.MeetingUrl,
		)
	}

	message += "<p style='margin-top: 30px;'>Thank you for choosing Biostat Healthcare.<br>We look forward to assisting you.</p>" +
		"<p style='margin-top: 20px;'>Best regards,<br><strong>Biostat Healthcare Team</strong></p>" +
		"</div></body></html>"

	return smtp.SendMail(e.SMTPHost+":"+e.SMTPPort, auth, e.SenderEmail, to, []byte(message))
}

func (e *EmailService) SendReportResultsEmail(patientInfo *models.SystemUser_, alerts []models.TestResultAlert) error {
	if len(alerts) == 0 {
		return nil
	}
	log.Println("alerts ", alerts)
	auth := smtp.PlainAuth("", e.SenderEmail, e.SenderPass, e.SMTPHost)

	subject := "Important: Abnormal Test Results Notification"
	to := []string{patientInfo.Email}
	fullName := patientInfo.FirstName + " " + patientInfo.LastName

	var body strings.Builder
	body.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	body.WriteString(fmt.Sprintf("From: Biostat Healthcare <%s>\r\n", e.SenderEmail))
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
	body.WriteString("<html><body style='font-family: Arial, sans-serif; color: #333;'>")
	body.WriteString("<div style='max-width: 600px; margin: auto; padding: 20px; border: 1px solid #e0e0e0; border-radius: 10px;'>")
	body.WriteString("<div style='text-align: center; margin-bottom: 20px;'>")
	body.WriteString("<h2 style='color: #dc3545;'>Health Alert: Out-of-Range Test Results</h2>")
	body.WriteString("</div>")

	body.WriteString(fmt.Sprintf("<p>Hi <strong>%s</strong>,</p>", fullName))
	body.WriteString("<p>We have reviewed your recent test results and noticed some values that are outside the normal range. Please review the details below:</p>")

	body.WriteString("<table style='width: 100%; border-collapse: collapse; margin-top: 15px;'>")
	body.WriteString("<tr style='background-color: #f2f2f2;'>")
	body.WriteString("<th style='padding: 8px; border: 1px solid #ddd;'>Date</th>")
	body.WriteString("<th style='padding: 8px; border: 1px solid #ddd;'>Test</th>")
	body.WriteString("<th style='padding: 8px; border: 1px solid #ddd;'>Component</th>")
	body.WriteString("<th style='padding: 8px; border: 1px solid #ddd;'>Result</th>")
	body.WriteString("<th style='padding: 8px; border: 1px solid #ddd;'>Normal Range</th>")
	body.WriteString("<th style='padding: 8px; border: 1px solid #ddd;'>Status</th>")
	body.WriteString("</tr>")

	for _, a := range alerts {
		body.WriteString("<tr>")
		body.WriteString(fmt.Sprintf("<td style='padding: 8px; border: 1px solid #ddd;'>%s</td>", a.ResultDate.Format("02 Jan 2006")))
		body.WriteString(fmt.Sprintf("<td style='padding: 8px; border: 1px solid #ddd;'>%s</td>", a.TestName))
		body.WriteString(fmt.Sprintf("<td style='padding: 8px; border: 1px solid #ddd;'>%s</td>", a.TestComponentName))
		body.WriteString(fmt.Sprintf("<td style='padding: 8px; border: 1px solid #ddd;'>%.2f</td>", a.ResultValue))
		body.WriteString(fmt.Sprintf("<td style='padding: 8px; border: 1px solid #ddd;'>%.2f - %.2f</td>", a.NormalMin, a.NormalMax))
		body.WriteString(fmt.Sprintf("<td style='padding: 8px; border: 1px solid #ddd;'>%s</td>", a.ResultComment))
		body.WriteString("</tr>")
	}
	body.WriteString("</table>")

	body.WriteString("<p style='margin-top: 20px;'>We strongly recommend consulting your doctor regarding these results.</p>")
	body.WriteString("<p style='margin-top: 20px;'>Best regards,<br><strong>Biostat Healthcare Team</strong></p>")
	body.WriteString("</div></body></html>")
	log.Println("Report abnormal values body prepared")
	return smtp.SendMail(e.SMTPHost+":"+e.SMTPPort, auth, e.SenderEmail, to, []byte(body.String()))
}
