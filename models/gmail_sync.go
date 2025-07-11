package models

type GmailSyncRequest struct {
	AccessToken string `json:"access_token"`
	UserID      uint64 `json:"user_id"`
}

type DocumentTypeResponse struct {
	DocumentType string `json:"document_type"`
	Error        string `json:"error"`
}