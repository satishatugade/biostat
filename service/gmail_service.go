package service

import (
	"biostat/models"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
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

func CreateGmailServiceFromToken(ctx context.Context, accessToken string) (*gmail.Service, error) {
    token := &oauth2.Token{AccessToken: accessToken}
    config := &oauth2.Config{}
    client := config.Client(ctx, token)
    return gmail.New(client)
}

func FetchEmailsWithAttachments(service *gmail.Service, userId uint64) ([]models.TblMedicalRecord, error) {
	profile, err := service.Users.GetProfile("me").Do()
	if err != nil {
		return nil, err
	}
	userEmail := profile.EmailAddress
	query := `("health report" OR "lab result" OR "blood test") has:attachment`
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

		attachments := ExtractAttachments(service, message, userEmail, userId)
		records = append(records, attachments...)
	}
	log.Println("Gmail Records found:", len(records), "userEmail: ", userEmail)
	return records, nil
}

func ExtractAttachments(service *gmail.Service, message *gmail.Message, userEmail string, userId uint64) []models.TblMedicalRecord {
	var records []models.TblMedicalRecord
	for _, part := range message.Payload.Parts {
		if part.Filename != "" {
			attachmentData, err := DownloadAttachment(service, message.Id, part.Body.AttachmentId)
			if err != nil {
				log.Println("Failed to download attachment %s: %v", part.Filename, err)
				continue
			}
			decodedName, err := url.QueryUnescape(part.Filename)
			if err != nil {
				decodedName = part.Filename
			}
			decodedName = strings.ReplaceAll(decodedName, " ", "_")
			re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
			safeName := re.ReplaceAllString(decodedName, "_")
			originalName := strings.TrimSuffix(safeName, filepath.Ext(safeName))
			extension := filepath.Ext(part.Filename)
			uniqueSuffix := time.Now().Format("20060102150405") + "-" + uuid.New().String()[:8]
			safeFileName := fmt.Sprintf("%s_%s%s", originalName, uniqueSuffix, extension)
			destinationPath := filepath.Join("uploads", safeFileName)

			if err := os.WriteFile(destinationPath, attachmentData, 0644); err != nil {
				log.Printf("Failed to save attachment locally %s: %v", part.Filename, err)
				continue
			}

			newRecord := &models.TblMedicalRecord{
				RecordName:        safeFileName,
				RecordSize:        int64(len(attachmentData)),
				FileType:          part.MimeType,
				RecordUrl:         fmt.Sprintf("%s/uploads/%s", os.Getenv("SHORT_URL_BASE"), safeFileName),
				UploadDestination: "LocalServer",
				Description:       getHeader(message.Payload.Headers, "Subject"),
				UploadSource:      "Gmail",
				RecordCategory:    "report",
				SourceAccount:     userEmail,
				UploadedBy:        userId,
				FetchedAt:         time.Now(),
			}
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
