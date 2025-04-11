package service

import (
	"biostat/models"
	"fmt"
	"net/smtp"
	"os"
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

	// Send Email
	return smtp.SendMail(e.SMTPHost+":"+e.SMTPPort, auth, e.SenderEmail, []string{systemUser.Email}, []byte(message))
}
