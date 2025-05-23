package service

import (
	"biostat/models"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func GetGmailAuthURL(userId string) (authUrl string) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URI")
	var googleOauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"https://mail.google.com/"},
		Endpoint:     google.Endpoint,
	}
	authURL := googleOauthConfig.AuthCodeURL(userId, oauth2.AccessTypeOffline)
	return authURL
}

func CreateGmailServiceClient(accessToken string, googleOauthConfig *oauth2.Config) (*gmail.Service, error) {
	creds := oauth2.Token{AccessToken: accessToken}
	client := googleOauthConfig.Client(context.Background(), &creds)
	return gmail.New(client)
}

func FetchEmailsWithAttachments(service *gmail.Service, userId uint64, accessToken string) ([]models.TblMedicalRecord, error) {
	profile, err := service.Users.GetProfile("me").Do()
	if err != nil {
		return nil, err
	}
	userEmail := profile.EmailAddress
	query := fmt.Sprintf(`subject:"health record" has:attachment`)
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

		attachments := ExtractAttachments(service, message, userEmail, userId, accessToken)
		records = append(records, attachments...)
	}
	log.Println("Gmail Records found:", len(records), "userEmail: ", userEmail)
	return records, nil
}

func ExtractAttachments(service *gmail.Service, message *gmail.Message, userEmail string, userId uint64, accessToken string) []models.TblMedicalRecord {
	var records []models.TblMedicalRecord

	for _, part := range message.Payload.Parts {
		if part.Filename != "" {
			attachmentData, err := DownloadAttachment(service, message.Id, part.Body.AttachmentId)
			if err != nil {
				log.Println("Failed to download attachment %s: %v", part.Filename, err)
				continue
			}
			newRecord, err := SaveRecordToDigiLocker(accessToken, attachmentData, part.Filename, part.MimeType)
			if err != nil {
				log.Println("Failed to Save attachment to digiLocker %s: %v %v", part.Filename, userId, err)
				continue
			}
			newRecord.Description = getHeader(message.Payload.Headers, "Subject")
			newRecord.UploadSource = "Gmail"
			newRecord.SourceAccount = userEmail
			newRecord.UploadedBy = userId
			newRecord.RecordCategory = "report"
			newRecord.UpdatedAt = time.Now()
			// record := models.TblMedicalRecord{
			// 	RecordName:     part.Filename,
			// 	RecordSize:     int64(len(attachmentData)),
			// 	FileType:       part.MimeType,
			// 	FileData:       attachmentData,
			// 	Description:    getHeader(message.Payload.Headers, "Subject"),
			// 	UploadSource:   "gmail",
			// 	SourceAccount:  userEmail,
			// 	UploadedBy:     userId,
			// 	RecordCategory: "report",
			// 	CreatedAt:      time.Now(),
			// 	UpdatedAt:      time.Now(),
			// 	RecordUrl:      fmt.Sprintf("https://mail.google.com/mail/u/0/?ui=2&ik=%s&attid=%s", message.Id, part.Body.AttachmentId),
			// }
			records = append(records, *newRecord)
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
