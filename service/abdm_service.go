package service

import (
	"biostat/config"
	"biostat/models"
	"biostat/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ABDMService interface {
	SendMobileOtp(mobile string) (*models.ABDMOtpResponse, error)
	VerifyOtp(txnId, otp string) (*models.ABDMOtpVerifyResponse, error)
	VerifyUser(txnId, abhaNumber, tToken string) (*models.ABDMUserVerifyResponse, error)

	SendAdhaarOtp(adharCardNo string) (*models.ABDMOtpResponse, error)
	VerifyAdharOtp(txnId, otp, mobile string) (interface{}, error)
	SetAbhaUsername(txnId, address string) (interface{}, error)
}

type ABDMServiceimpl struct {
	X_CM_ID          string
	ABDMBase         string
	ABDMDev          string
	ABDMClientID     string
	ABDMClientSecret string
}

func NewABDMService() *ABDMServiceimpl {
	return &ABDMServiceimpl{
		X_CM_ID:          config.PropConfig.ApiURL.ABDM_CMID,
		ABDMBase:         config.PropConfig.ApiURL.ADBMBase,
		ABDMDev:          config.PropConfig.ApiURL.ABDMDEV,
		ABDMClientID:     config.PropConfig.ApiURL.ABDM_CLIENT_ID,
		ABDMClientSecret: config.PropConfig.ApiURL.ABDM_CLIENT_SECRET,
	}
}

func (s *ABDMServiceimpl) SendMobileOtp(mobile string) (*models.ABDMOtpResponse, error) {
	token, err := s.GetAbdmSession()
	if err != nil {
		return nil, err
	}
	cert, err := s.GetAbdmPublicCertificate(token.AccessToken)
	if err != nil {
		return nil, err
	}

	encryptedMobile, err := utils.EncryptWithPublicKey(cert.PublicKey, mobile)
	if err != nil {
		return nil, err
	}
	resp, err := s.RequestAbdmMobileOtp(token.AccessToken, encryptedMobile)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *ABDMServiceimpl) VerifyOtp(txnId, otp string) (*models.ABDMOtpVerifyResponse, error) {
	token, err := s.GetAbdmSession()
	if err != nil {
		return nil, err
	}

	// Step 2: Get public key
	cert, err := s.GetAbdmPublicCertificate(token.AccessToken)
	if err != nil {
		return nil, err
	}

	// Step 3: Encrypt OTP
	encryptedOtp, err := utils.EncryptWithPublicKey(cert.PublicKey, otp)
	if err != nil {
		return nil, err
	}

	// Step 4: Call Verify OTP API
	resp, err := s.VerifyAbdmMobileOtp(token.AccessToken, txnId, encryptedOtp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *ABDMServiceimpl) VerifyUser(txnId, abhaNumber, tToken string) (*models.ABDMUserVerifyResponse, error) {
	token, err := s.GetAbdmSession()
	if err != nil {
		return nil, err
	}

	resp, err := s.VerifyAbdmUser(token.AccessToken, tToken, abhaNumber, txnId)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (s *ABDMServiceimpl) SendAdhaarOtp(adharCardNo string) (*models.ABDMOtpResponse, error) {
	token, err := s.GetAbdmSession()
	if err != nil {
		return nil, err
	}
	cert, err := s.GetAbdmPublicCertificate(token.AccessToken)
	if err != nil {
		return nil, err
	}

	encryptedAdhar, err := utils.EncryptWithPublicKey(cert.PublicKey, adharCardNo)
	if err != nil {
		return nil, err
	}
	resp, err := s.RequestAadhaarOtp(token.AccessToken, encryptedAdhar)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *ABDMServiceimpl) VerifyAdharOtp(txnId, otp, mobile string) (interface{}, error) {
	token, err := s.GetAbdmSession()
	if err != nil {
		return nil, err
	}

	cert, err := s.GetAbdmPublicCertificate(token.AccessToken)
	if err != nil {
		return nil, err
	}

	encryptedOtp, err := utils.EncryptWithPublicKey(cert.PublicKey, otp)
	if err != nil {
		return nil, err
	}

	createAbhaResp, err := s.CreateAbhaByAadhaar(token.AccessToken, txnId, encryptedOtp, mobile)
	if err != nil {
		return nil, err
	}

	address, err := s.GetAbhaAddressSuggestions(token.AccessToken, txnId)
	if err != nil {
		return nil, err
	}

	resp := map[string]interface{}{
		"abha":    createAbhaResp,
		"address": address,
	}
	return resp, nil
}

func (s *ABDMServiceimpl) SetAbhaUsername(txnId, address string) (interface{}, error) {
	token, err := s.GetAbdmSession()
	if err != nil {
		return nil, err
	}

	abhaCard, err := s.CreateAbhaAddress(token.AccessToken, txnId, address, 1)
	if err != nil {
		return nil, err
	}

	return abhaCard, nil
}

// ABDM API service Function

func (s *ABDMServiceimpl) GetAbdmSession() (models.ABDMTokenResponse, error) {
	body := models.ABDMSessionRequest{
		ClientID:     s.ABDMClientID,
		ClientSecret: s.ABDMClientSecret,
		GrantType:    "client_credentials",
	}

	requestURL := fmt.Sprintf("%s/hiecm/gateway/v3/sessions", s.ABDMDev)

	res, err := s.doRequest("POST", requestURL, body, nil)
	if err != nil {
		return models.ABDMTokenResponse{}, fmt.Errorf("request failed: %w", err)
	}
	if res.StatusCode != 202 {
		return models.ABDMTokenResponse{}, fmt.Errorf("unexpected status: %d", res.StatusCode)
	}

	var success models.ABDMTokenResponse
	if err := json.Unmarshal(res.Body, &success); err == nil && success.AccessToken != "" {
		return success, nil
	}
	var fail models.ABDMSessionErrorResponse
	_ = json.Unmarshal(res.Body, &fail)
	return models.ABDMTokenResponse{}, fmt.Errorf("%s: %s", fail.Error.Code, fail.Error.Message)
}

func (s *ABDMServiceimpl) GetAbdmPublicCertificate(accessToken string) (models.ABDMPublicKeyResponse, error) {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}
	requestURL := fmt.Sprintf("%s/v3/profile/public/certificate", s.ABDMBase)
	res, err := s.doRequest("GET", requestURL, nil, headers)
	if err != nil {
		return models.ABDMPublicKeyResponse{}, err
	}

	if res.StatusCode != 200 {
		var fail map[string]string
		_ = json.Unmarshal(res.Body, &fail)
		return models.ABDMPublicKeyResponse{}, fmt.Errorf("%s: %s", fail["code"], fail["message"])
	}

	var success models.ABDMPublicKeyResponse
	_ = json.Unmarshal(res.Body, &success)
	return success, nil
}

func (s *ABDMServiceimpl) RequestAbdmMobileOtp(accessToken, encryptedMobile string) (models.ABDMOtpResponse, error) {
	body := models.ABDMOtpRequest{
		Scope:     []string{"abha-login", "mobile-verify"},
		LoginHint: "mobile",
		LoginId:   encryptedMobile,
		OtpSystem: "abdm",
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}

	requestURL := fmt.Sprintf("%s/v3/profile/login/request/otp", s.ABDMBase)
	res, err := s.doRequest("POST", requestURL, body, headers)
	if err != nil {

		return models.ABDMOtpResponse{}, err
	}

	if res.StatusCode != 200 {
		var fail map[string]string
		_ = json.Unmarshal(res.Body, &fail)
		log.Println(fail)
		return models.ABDMOtpResponse{}, fmt.Errorf("%v", fail)
	}

	var success models.ABDMOtpResponse
	_ = json.Unmarshal(res.Body, &success)
	return success, nil
}

func (s *ABDMServiceimpl) VerifyAbdmMobileOtp(accessToken, txnId, encryptedOtp string) (models.ABDMOtpVerifyResponse, error) {
	body := models.ABDMVerifyOtpRequest{
		Scope: []string{"abha-login", "mobile-verify"},
	}
	body.AuthData.AuthMethods = []string{"otp"}
	body.AuthData.Otp.TxnID = txnId
	body.AuthData.Otp.OtpValue = encryptedOtp

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}
	requestURL := fmt.Sprintf("%s/v3/profile/login/verify", s.ABDMBase)

	res, err := s.doRequest("POST", requestURL, body, headers)
	if err != nil {
		return models.ABDMOtpVerifyResponse{}, err
	}

	if res.StatusCode != 200 {
		var fail map[string]string
		_ = json.Unmarshal(res.Body, &fail)
		return models.ABDMOtpVerifyResponse{}, fmt.Errorf("%v", fail)
	}

	var success models.ABDMOtpVerifyResponse
	_ = json.Unmarshal(res.Body, &success)
	if strings.ToLower(success.AuthResult) != "success" {
		return success, fmt.Errorf("OTP verification failed: %s", success.Message)
	}

	return success, nil
}

func (s *ABDMServiceimpl) VerifyAbdmUser(accessToken, tToken, abhaNumber, txnId string) (models.ABDMUserVerifyResponse, error) {
	body := models.ABDMUserVerifyRequest{
		ABHANumber: abhaNumber,
		TxnID:      txnId,
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
		"T-token":       fmt.Sprintf("Bearer %s", tToken),
	}
	requestURL := fmt.Sprintf("%s/v3/profile/login/verify/user", s.ABDMBase)

	res, err := s.doRequest("POST", requestURL, body, headers)
	if err != nil {
		return models.ABDMUserVerifyResponse{}, err
	}

	if res.StatusCode != 200 {
		var fail map[string]string
		_ = json.Unmarshal(res.Body, &fail)
		return models.ABDMUserVerifyResponse{}, fmt.Errorf("%v", fail)
	}

	var success models.ABDMUserVerifyResponse
	_ = json.Unmarshal(res.Body, &success)
	return success, nil
}

func (s *ABDMServiceimpl) RequestAadhaarOtp(accessToken, aadhaarEncrypted string) (*models.ABDMOtpResponse, error) {
	payload := map[string]interface{}{
		"txnId":     "",
		"scope":     []string{"abha-enrol"},
		"loginHint": "aadhaar",
		"loginId":   aadhaarEncrypted,
		"otpSystem": "aadhaar",
	}

	requestURL := fmt.Sprintf("%s/v3/enrollment/request/otp", s.ABDMBase)
	log.Println(requestURL)

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}
	res, err := s.doRequest("POST", requestURL, payload, headers)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		var fail map[string]string
		_ = json.Unmarshal(res.Body, &fail)
		return nil, fmt.Errorf("%v", fail)
	}

	var otpResp models.ABDMOtpResponse
	_ = json.Unmarshal(res.Body, &otpResp)
	return &otpResp, nil
}

func (s *ABDMServiceimpl) CreateAbhaByAadhaar(accessToken, txnId, otpEncrypted, mobile string) (*models.AbdmCreateAbhaByAadhaarResponse, error) {
	requestURL := fmt.Sprintf("%s/v3/enrollment/enrol/byAadhaar", s.ABDMBase)

	payload := map[string]interface{}{
		"authData": map[string]interface{}{
			"authMethods": []string{"otp"},
			"otp": map[string]interface{}{
				"timeStamp": time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
				"txnId":     txnId,
				"otpValue":  otpEncrypted,
				"mobile":    mobile,
			},
		},
		"consent": map[string]interface{}{
			"code":    "abha-enrollment",
			"version": "1.4",
		},
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}

	res, err := s.doRequest("POST", requestURL, payload, headers)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		var fail map[string]string
		_ = json.Unmarshal(res.Body, &fail)
		return nil, fmt.Errorf("%v", fail)
	}

	var enrolResp models.AbdmCreateAbhaByAadhaarResponse
	_ = json.Unmarshal(res.Body, &enrolResp)
	return &enrolResp, nil
}

func (s *ABDMServiceimpl) GetAbhaAddressSuggestions(accessToken, txnId string) (*models.AbdmAbhaAddressSuggestionResponse, error) {
	requestURL := fmt.Sprintf("%s/v3/enrollment/enrol/suggestion", s.ABDMBase)

	headers := map[string]string{
		"Authorization":  fmt.Sprintf("Bearer %s", accessToken),
		"Transaction_Id": txnId,
	}

	res, err := s.doRequest("GET", requestURL, nil, headers)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		var fail map[string]string
		_ = json.Unmarshal(res.Body, &fail)
		return nil, fmt.Errorf("%v", fail)
	}

	var suggestionResp models.AbdmAbhaAddressSuggestionResponse
	_ = json.Unmarshal(res.Body, &suggestionResp)
	return &suggestionResp, nil
}

func (s *ABDMServiceimpl) CreateAbhaAddress(accessToken, txnId, abhaAddress string, preferred int) (*models.AbdmCreateAbhaAddressResponse, error) {
	requestURL := fmt.Sprintf("%s/v3/enrollment/enrol/abha-address", s.ABDMBase)
	payload := map[string]interface{}{
		"txnId":       txnId,
		"abhaAddress": abhaAddress,
		"preferred":   preferred,
	}
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}

	res, err := s.doRequest("POST", requestURL, payload, headers)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		var fail map[string]string
		_ = json.Unmarshal(res.Body, &fail)
		return nil, fmt.Errorf("%v", fail)
	}

	var addrResp models.AbdmCreateAbhaAddressResponse
	_ = json.Unmarshal(res.Body, &addrResp)
	return &addrResp, nil
}

type HTTPResponse struct {
	StatusCode int
	Body       []byte
}

func (s *ABDMServiceimpl) doRequest(method, url string, body any, headers map[string]string) (HTTPResponse, error) {
	log.Println("Request URL:", url)

	var reqBody io.Reader
	if body != nil {
		b, _ := json.MarshalIndent(body, "", " ")
		log.Println("Body:", string(b))
		reqBody = bytes.NewReader(b)
	}

	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("REQUEST-ID", uuid.New().String())
	req.Header.Add("TIMESTAMP", time.Now().UTC().Format("2006-01-02T15:04:05.000Z"))
	req.Header.Add("X-CM-ID", s.X_CM_ID)

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Got error making request: ", resp)
		return HTTPResponse{}, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err != nil {
		log.Println("Response body (raw):")
		log.Println(string(bodyBytes))
	} else {
		log.Println("Response body (formatted):")
		log.Println(prettyJSON.String())
	}
	return HTTPResponse{StatusCode: resp.StatusCode, Body: bodyBytes}, nil
}
