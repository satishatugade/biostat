package service

import (
	"biostat/repository"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type YahooService interface {
	GetYahooAuthURL(userID uint64) (string, error)
	GetYahooToken(ctx context.Context, code string) (*YahooOAuthToken, error)
	// SyncYahooWeb(ctx context.Context, userId uint64, token *oauth2.Token) error
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

func generateCodeVerifier() string {
	rand.Seed(time.Now().UnixNano())
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	b := make([]byte, 64)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func generateCodeChallenge(verifier string) string {
	h := sha256.New()
	h.Write([]byte(verifier))
	sha := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(sha)
}

func (ys *YahooServiceImpl) GetYahooAuthURL(userID uint64) (string, error) {
	_, err := ys.diagnosticRepo.GetPatientLabNameAndEmail(userID)
	if err != nil {
		return "", err
	}
	clientID := os.Getenv("YAHOO_CLIENT_ID")
	redirectURI := os.Getenv("YAHOO_REDIRECT_URL")
	scope := "openid"
	codeVerifier := generateCodeVerifier()
	codeChallenge := generateCodeChallenge(codeVerifier)
	authURL := fmt.Sprintf(
		"https://api.login.yahoo.com/oauth2/request_auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%d&code_challenge=%s&code_challenge_method=S256",
		clientID, redirectURI, scope, userID, codeChallenge,
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
	redirectURI := os.Getenv("YAHOO_REDIRECT_URI")
	data := fmt.Sprintf("client_id=%s&client_secret=%s&redirect_uri=%s&code=%s&grant_type=authorization_code",
		clientID, clientSecret, redirectURI, code,
	)
	req, _ := http.NewRequest("POST", "https://api.login.yahoo.com/oauth2/get_token", bytes.NewBufferString(data))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed token exchange: %s", string(body))
	}

	var token YahooOAuthToken
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}
	token.Expiry = time.Now().Add(time.Second * time.Duration(token.ExpiresIn))
	return &token, nil
}
