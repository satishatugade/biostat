package service

import (
	"biostat/config"
	"biostat/models"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var googleOauthConfig = &oauth2.Config{
	ClientID:     config.GoogleClientID,
	ClientSecret: config.GoogleClientSecret,
	RedirectURL:  config.RedirectURI,
	Scopes:       []string{"https://mail.google.com/"},
	Endpoint:     google.Endpoint,
}

func GetGmailAuthURL(userId string) (authUrl string) {
	authURL := googleOauthConfig.AuthCodeURL(userId, oauth2.AccessTypeOffline)
	return authURL
}

func CreateGmailServiceClient(accessToken string) (*gmail.Service, error) {
	creds := oauth2.Token{AccessToken: accessToken}
	client := googleOauthConfig.Client(context.Background(), &creds)
	return gmail.New(client)
}

func FetchEmailsWithAttachments(service *gmail.Service, userId int64) ([]models.TblMedicalRecord, error) {
	profile, err := service.Users.GetProfile("me").Do()
	if err != nil {
		return nil, err
	}
	userEmail := profile.EmailAddress
	query := fmt.Sprintf("subject:report has:attachment")
	results, err := service.Users.Messages.List("me").Q(query).Do()
	if err != nil {
		return nil, err
	}

	var records []models.TblMedicalRecord

	for _, msg := range results.Messages {
		message, err := service.Users.Messages.Get("me", msg.Id).Do()
		if err != nil {
			continue
		}

		attachments := ExtractAttachments(service, message, userEmail)
		records = append(records, attachments...)
	}
	log.Println("Gmail Records found:", len(records), "userEmail: ", userEmail)
	return records, nil
}

func ExtractAttachments(service *gmail.Service, message *gmail.Message, userEmail string) []models.TblMedicalRecord {
	var records []models.TblMedicalRecord

	for _, part := range message.Payload.Parts {
		if part.Filename != "" {
			attachmentData, err := DownloadAttachment(service, message.Id, part.Body.AttachmentId)
			if err != nil {
				log.Println("Failed to download attachment %s: %v", part.Filename, err)
				continue
			}

			record := models.TblMedicalRecord{
				RecordName:   part.Filename,
				RecordSize:   int64(len(attachmentData)),
				RecordExt:    part.MimeType,
				FileData:     attachmentData,
				Description:  getHeader(message.Payload.Headers, "Subject"),
				UploadSource: userEmail,
				RecordType:   "report",
				CreatedAt:    time.Now(),
				IsActive:     true,
			}
			records = append(records, record)
		}
	}
	return records
}

func DownloadAttachment(service *gmail.Service, messageID, attachmentID string) ([]byte, error) {
	attachment, err := service.Users.Messages.Attachments.Get("me", messageID, attachmentID).Do()
	if err != nil {
		return nil, err
	}

	decodedData, err := base64.URLEncoding.DecodeString(attachment.Data)
	if err != nil {
		return nil, err
	}

	return decodedData, nil
}

func getHeader(headers []*gmail.MessagePartHeader, name string) string {
	for _, h := range headers {
		if h.Name == name {
			return h.Value
		}
	}
	return ""
}
