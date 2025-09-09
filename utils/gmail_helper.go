package utils

import (
	"encoding/base64"
	"log"
	"strings"

	"google.golang.org/api/gmail/v1"
)

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

func GetHeader(headers []*gmail.MessagePartHeader, name string) string {
	for _, h := range headers {
		if h.Name == name {
			return h.Value
		}
	}
	return ""
}

func decodeBase64Safe(data string) string {
	if data == "" {
		return ""
	}

	if decoded, err := base64.RawURLEncoding.DecodeString(data); err == nil {
		// log.Println("Retuning base64.RawURLEncoding.DecodeString")
		return string(decoded)
	} else {
		log.Printf("[decodeBase64Safe] RawURLEncoding failed: %v", err)
	}
	if decoded, err := base64.StdEncoding.DecodeString(data); err == nil {
		// log.Println("Retuning base64.StdEncoding.DecodeString")
		return string(decoded)
	} else {
		log.Printf("[decodeBase64Safe] StdEncoding failed: %v", err)
	}

	padded := data
	if m := len(padded) % 4; m != 0 {
		padded += strings.Repeat("=", 4-m)
	}
	if decoded, err := base64.URLEncoding.DecodeString(padded); err == nil {
		// log.Println("Retuning base64.URLEncoding.DecodeString")
		return string(decoded)
	} else {
		log.Printf("[decodeBase64Safe] URLEncoding with padding failed: %v", err)
	}

	log.Printf("[decodeBase64Safe] All decoding attempts failed, returning raw string")
	return data
}

func extractBodyFromParts(parts []*gmail.MessagePart) string {
	for _, part := range parts {
		if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
			// log.Println("extractBodyFromParts Detected Text/plain")
			return decodeBase64Safe(part.Body.Data)
		}
		if part.MimeType == "text/html" && part.Body != nil && part.Body.Data != "" {
			// log.Println("extractBodyFromParts Detected Text/html")
			return StripHTML(decodeBase64Safe(part.Body.Data))
		}
		if len(part.Parts) > 0 {
			if result := extractBodyFromParts(part.Parts); result != "" {
				// log.Println("extractBodyFromParts this part has its own sub-parts")
				return result
			}
		}
	}
	// log.Println("Returning empty extractBodyFromParts")
	return ""
}

func GetMessageBody(msg *gmail.Message) string {
	if msg == nil || msg.Payload == nil {
		log.Println("Message is nil returning Empty string")
		return ""
	}
	if len(msg.Payload.Parts) == 0 && msg.Payload.Body != nil && msg.Payload.Body.Data != "" {
		body := decodeBase64Safe(msg.Payload.Body.Data)
		if strings.Contains(strings.ToLower(msg.Payload.MimeType), "html") {
			// log.Println("Extracting & Returning HTML")
			return StripHTML(body)
		}
		log.Println("Returning decodeBase64Safe HTML")
		return body
	}

	// Multipart email
	return extractBodyFromParts(msg.Payload.Parts)
}
