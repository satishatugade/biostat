package models

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
