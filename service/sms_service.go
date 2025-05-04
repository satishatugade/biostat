package service

import (
	"os"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type SmsService interface {
	SendSMS(toPhoneNumber string, link string) error
}

type SmsServiceImpl struct {
}

func NewSmsService() SmsService {
	return &SmsServiceImpl{}
}

func (s *SmsServiceImpl) SendSMS(toPhoneNumber string, link string) error {
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	fromPhone := os.Getenv("TWILIO_PHONE_NUMBER")

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: accountSid,
		Password: authToken,
	})

	params := &openapi.CreateMessageParams{}
	params.SetTo(toPhoneNumber)
	params.SetFrom(fromPhone)
	params.SetBody("Hello User ! You have received a diagnostic report link from your patient. Please access the report using the secure link below:\n" + link)

	_, err := client.Api.CreateMessage(params)
	return err
}
