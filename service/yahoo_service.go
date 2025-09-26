package service

import (
	"biostat/repository"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type YahooService interface {
	GetYahooAuthURL(userID uint64) (string, error)
	GetYahooToken(ctx context.Context, code string) (*YahooOAuthToken, error)
}

type YahooServiceImpl struct {
	userService          UserService
	apiService           ApiService
	processStatusService ProcessStatusService
	gmailSyncService     GmailSyncService
	diagnosticRepo       repository.DiagnosticRepository
}

func NewYahooService(userService UserService, apiService ApiService, processStatusService ProcessStatusService, gmailSyncService GmailSyncService, diagnosticRepo repository.DiagnosticRepository) YahooService {
	return &YahooServiceImpl{userService: userService, apiService: apiService, processStatusService: processStatusService, gmailSyncService: gmailSyncService, diagnosticRepo: diagnosticRepo}
}

func (ys *YahooServiceImpl) GetYahooAuthURL(userID uint64) (string, error) {
	_, err := ys.diagnosticRepo.GetPatientLabNameAndEmail(userID)
	if err != nil {
		return "", err
	}
	clientID := os.Getenv("YAHOO_CLIENT_ID")
	redirectURI := os.Getenv("YAHOO_REDIRECT_URL")
	scope := "mail-r"
	authURL := fmt.Sprintf(
		"https://api.login.yahoo.com/oauth2/request_auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%d",
		clientID, redirectURI, scope, userID,
	)
	return authURL, nil
}

type YahooOAuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Expiry       time.Time
}

func (ys *YahooServiceImpl) GetYahooToken(ctx context.Context, code string) (*YahooOAuthToken, error) {
	clientID := os.Getenv("YAHOO_CLIENT_ID")
	clientSecret := os.Getenv("YAHOO_CLIENT_SECRET")
	redirectURI := os.Getenv("YAHOO_REDIRECT_URL")
	log.Println("@GetYahooToken:", code)
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", "https://api.login.yahoo.com/oauth2/get_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Println("Yahoo Token Response:", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed token exchange: %s", string(body))
	}

	var token YahooOAuthToken
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}
	token.Expiry = time.Now().Add(time.Second * time.Duration(token.ExpiresIn))
	return &token, nil
}
