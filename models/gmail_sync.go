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
	LLMClassifier   *ClassifierResult `json:"llm_classifier"`
	RegexClassifier *ClassifierResult `json:"regex_classifier"`
}

type ClassifierResult struct {
	DocumentType string `json:"document_type"`
	Logs         string `json:"logs"`
}