package models

import "net/http"

type GmailSyncRequest struct {
	AccessToken string `json:"access_token"`
	UserID      uint64 `json:"user_id"`
}

type DocTypeAPIResponse struct {
	Content *Content `json:"content"`
	Error   *string  `json:"error"`
}

type Content struct {
	PatientName     string            `json:"patient_name"`
	LLMClassifier   *ClassifierResult `json:"llm_classifier"`
	RegexClassifier *ClassifierResult `json:"regex_classifier"`
}

type ClassifierResult struct {
	DocumentType string `json:"document_type"`
	Logs         string `json:"logs"`
}

type Relative struct {
	UserID uint64 `json:"user_id"`
	Name   string `json:"name"`
}

type PatientDocRequest struct {
	UserID            uint64     `json:"user_id"`
	PatientName       string     `json:"patient_name"`
	Relatives         []Relative `json:"relatives"`
	ReportPatientName string     `json:"report_patient_name"`
}

type PatientDocResponse struct {
	UserID           uint64 `json:"user_id"`
	FinalPatientName string `json:"final_patient_name"`
	MatchedWith      string `json:"matched_with"`    // "patient" or "relative"
	MatchedUserID    uint64 `json:"matched_user_id"` // id of patient or relative
	IsFallback       bool   `json:"is_fallback"`
}

type GmailReSyncRequest struct {
	ProviderID string `json:"provider_id"`
}

type OutlookService struct {
	Client *http.Client
}

type OutlookMessage struct {
	ID             string    `json:"id"`
	Subject        string    `json:"subject"`
	From           FromField `json:"from"`
	Received       string    `json:"receivedDateTime"`
	HasAttachments bool      `json:"hasAttachments"`
	BodyPreview    string    `json:"bodyPreview"`
}

type FromField struct {
	EmailAddress struct {
		Address string `json:"address"`
		Name    string `json:"name"`
	} `json:"emailAddress"`
}

type OutlookMessagesResponse struct {
	Value    []OutlookMessage `json:"value"`
	NextLink string           `json:"@odata.nextLink"`
}

type OutlookAttachment struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Size         int    `json:"size"`
	ContentType  string `json:"contentType"`
	ContentBytes string `json:"contentBytes"`
}
