package service

import (
	"biostat/config"
	"biostat/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"time"

	"github.com/google/uuid"
)

type ApiService interface {
	CallGeminiService(file io.Reader, filename string) (models.LabReport, error)
	CallPrescriptionDigitizeAPI(file io.Reader, filename string) (models.PatientPrescription, error)
	CallSummarizeReportService(data models.PatientBasicInfo) (models.ResultSummary, error)
	AnalyzePrescriptionWithAI(data models.PatientPrescription) (string, error)
	AnalyzePharmacokineticsInfo(data models.PharmacokineticsInput) (string, error)
	SummarizeMedicalHistory(data models.PharmacokineticsInput) (string, error)
	AskAI(message string, userId uint64, patientName string, query_type string) (*models.AskAPIResponse, error)
	CheckPDFAndGetPassword(file io.Reader, fileName, emailBody string) (*models.PDFProtectionResult, error)
	CallDocumentTypeAPI(file io.Reader, filename string) (*models.DocTypeAPIResponse, error)
}

type ApiServiceImpl struct {
	GeminiAPIURL               string
	ReportSummaryAPI           string
	PrescriptionAPI            string
	PharmacokineticsAPI        string
	SummarizeMedicalHistoryAPI string
	DigitizePrescriptionAPI    string
	CheckPDFProtectionAPI      string
	PDFPasswordAPI             string
	client                     *http.Client
	sessionCache               map[uint64]string
}

func NewApiService() ApiService {
	return &ApiServiceImpl{
		GeminiAPIURL:               os.Getenv("GEMINI_API_URL"),
		ReportSummaryAPI:           os.Getenv("REPORT_SUMMARY_API"),
		PrescriptionAPI:            os.Getenv("PRESCRIPTION_API"),
		PharmacokineticsAPI:        os.Getenv("PHARMACOKINETICS_API"),
		SummarizeMedicalHistoryAPI: os.Getenv("SUMMARIZE_HISTORY_API"),
		DigitizePrescriptionAPI:    os.Getenv("DIGITIZE_PRESCRIPTION_API"),
		CheckPDFProtectionAPI:      config.PropConfig.ApiURL.CheckPDFProtectionAPI,
		PDFPasswordAPI:             config.PropConfig.ApiURL.PDFPasswordAPI,
		client:                     &http.Client{},
		sessionCache:               make(map[uint64]string),
	}
}

func (s *ApiServiceImpl) CallGeminiService(file io.Reader, filename string) (models.LabReport, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	log.Println("File Name : ", filename)

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return models.LabReport{}, fmt.Errorf("failed to read file for MIME detection: %w", err)
	}
	mimeType := http.DetectContentType(buf[:n])
	log.Printf("Detected MIME type: %s", mimeType)

	fileReader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filename))
	h.Set("Content-Type", mimeType)

	part, err := writer.CreatePart(h)
	if err != nil {
		return models.LabReport{}, fmt.Errorf("failed to create form part: %w", err)
	}

	if _, err := io.Copy(part, fileReader); err != nil {
		return models.LabReport{}, fmt.Errorf("failed to copy file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return models.LabReport{}, fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", s.GeminiAPIURL, body)
	if err != nil {
		return models.LabReport{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		return models.LabReport{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return models.LabReport{}, fmt.Errorf("service returned error: %s, body: %s", resp.Status, string(respBody))
	}

	var reportData models.LabReport
	if err := json.NewDecoder(resp.Body).Decode(&reportData); err != nil {
		return models.LabReport{}, fmt.Errorf("failed to parse JSON response: %w", err)
	}
	prettyJSON, err := json.MarshalIndent(reportData, "", "  ")
	if err != nil {
		log.Println("Failed to format JSON:", err)
	} else {
		log.Println("Digitize report response (pretty):\n", string(prettyJSON))
	}
	return reportData, nil
}

func (a *ApiServiceImpl) CallSummarizeReportService(data models.PatientBasicInfo) (models.ResultSummary, error) {

	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON (pretty): %v", err)
		return models.ResultSummary{}, err
	}
	log.Println("Sending JSON Payload to Report Summary API:", string(prettyJSON))

	jsonData, err := json.Marshal(data)
	if err != nil {
		return models.ResultSummary{}, err
	}
	var result models.ResultSummary
	resp, err := http.Post(a.ReportSummaryAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return models.ResultSummary{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.ResultSummary{}, fmt.Errorf("API returned status: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.ResultSummary{}, fmt.Errorf("error decoding response: %w", err)
	}

	return result, nil
}

func (api *ApiServiceImpl) AnalyzePrescriptionWithAI(prescription models.PatientPrescription) (string, error) {

	jsonData, err := json.Marshal(prescription)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, api.PrescriptionAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI service responded with status code %d", resp.StatusCode)
	}

	var response struct {
		Summary string `json:"pharmacodynamics_explanation"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.Summary, nil
}

func (s *ApiServiceImpl) CallPrescriptionDigitizeAPI(file io.Reader, filename string) (models.PatientPrescription, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	log.Println("Prescription File Name:", filename)
	buf := make([]byte, 512)
	n, err1 := file.Read(buf)
	if err1 != nil {
		return models.PatientPrescription{}, fmt.Errorf("failed to read file for MIME detection: %w", err1)
	}
	mimeType := http.DetectContentType(buf[:n])
	log.Printf("Detected MIME type: %s", mimeType)

	fileReader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filename))
	h.Set("Content-Type", mimeType)

	part, err1 := writer.CreatePart(h)
	if err1 != nil {
		return models.PatientPrescription{}, fmt.Errorf("failed to create form part: %w", err1)
	}

	if _, err := io.Copy(part, fileReader); err != nil {
		return models.PatientPrescription{}, fmt.Errorf("failed to copy file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return models.PatientPrescription{}, fmt.Errorf("failed to close writer: %w", err)
	}

	req, err1 := http.NewRequest("POST", s.DigitizePrescriptionAPI, body)
	if err1 != nil {
		return models.PatientPrescription{}, fmt.Errorf("failed to create request: %w", err1)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err1 := s.client.Do(req)
	if err1 != nil {
		return models.PatientPrescription{}, fmt.Errorf("failed to send request: %w", err1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return models.PatientPrescription{}, fmt.Errorf("service returned error: %s, body: %s", resp.Status, string(respBody))
	}

	var prescriptionData models.PatientPrescriptionData
	if err := json.NewDecoder(resp.Body).Decode(&prescriptionData); err != nil {
		return models.PatientPrescription{}, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	var parsedDate time.Time
	var err error

	dateFormats := []string{
		"02-Jan-2006",
		"02-Jan-06",
		"02/01/2006",
		"02-01-2006",
		"2006-01-02",
		"2006/01/02",
		"2006.01.02",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
	}

	for _, format := range dateFormats {
		if prescriptionData.PrescriptionDate != "" {
			parsedDate, err = time.Parse(format, prescriptionData.PrescriptionDate)
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		log.Println("Invalid date format:", err)
		return models.PatientPrescription{}, fmt.Errorf("invalid date format: %w", err)
	}
	data := models.PatientPrescription{
		PrescriptionId:      prescriptionData.PrescriptionId,
		PatientId:           prescriptionData.PatientId,
		PrescribedBy:        prescriptionData.PrescribedBy,
		PrescriptionName:    &prescriptionData.PrescriptionName,
		Description:         prescriptionData.Description,
		PrescriptionDate:    &parsedDate,
		PrescriptionDetails: make([]models.PrescriptionDetail, 0),
	}

	for _, detail := range prescriptionData.PrescriptionDetails {
		d := models.PrescriptionDetail{
			PrescriptionDetailId: detail.PrescriptionDetailId,
			PrescriptionId:       detail.PrescriptionId,
			MedicineName:         detail.MedicineName,
			PrescriptionType:     detail.PrescriptionType,
			Duration:             detail.Duration,
			DosageInfo: []models.PrescriptionDoseSchedule{
				{
					DoseQuantity: detail.DoseQuantity,
					UnitType:     detail.UnitType,
					Instruction:  detail.Instruction,
				},
			},
		}
		data.PrescriptionDetails = append(data.PrescriptionDetails, d)
	}

	return data, nil
}

func (api *ApiServiceImpl) AnalyzePharmacokineticsInfo(input models.PharmacokineticsInput) (string, error) {

	prettyJSON, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		log.Println("Failed to generate pretty JSON:", err)
	} else {
		log.Println("PharmacokineticsInput Payload:\n", string(prettyJSON))
	}

	jsonData, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, api.PharmacokineticsAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("pharmacokinetics AI service responded with status code %d", resp.StatusCode)
	}

	var response struct {
		Summary string `json:"analysis"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.Summary, nil
}

func (api *ApiServiceImpl) SummarizeMedicalHistory(input models.PharmacokineticsInput) (string, error) {

	prettyJSON, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		log.Println("Failed to generate pretty JSON:", err)
	} else {
		log.Println("Summarize Medical History Payload:\n", string(prettyJSON))
	}

	jsonData, err := json.Marshal(input)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, api.SummarizeMedicalHistoryAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("summarize Medical History responded with status code %d", resp.StatusCode)
	}

	var response struct {
		Summary string `json:"summary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	prettyJSON, err1 := json.MarshalIndent(response, "", "  ")
	if err1 != nil {
		log.Println("Failed to generate pretty JSON:", err1)
	} else {
		log.Println("Summarize Medical History reponse Payload:\n", string(prettyJSON))
	}

	return response.Summary, nil
}

func (s *ApiServiceImpl) AskAI(message string, userId uint64, patientName string, query_type string) (*models.AskAPIResponse, error) {
	apiURL := os.Getenv("ASK_API")
	if apiURL == "" {
		apiURL = "http://bio.alrn.in/api/ask"
	}
	// log.Println("AskAI message body : ", message)
	data := message
	var reportData interface{}
	sessionID, exists := s.sessionCache[userId]
	if !exists {
		sessionID = uuid.New().String()    // generate new only first time
		s.sessionCache[userId] = sessionID // store for future requests
	}
	if query_type == "report" || query_type == "prescription" {
		data = message
		message = ""
		if err := json.Unmarshal([]byte(data), &reportData); err != nil {
			reportData = data
		}
	}
	reqBody := map[string]interface{}{
		"name":       patientName,
		"user_id":    fmt.Sprintf("%d", userId),
		"session_id": sessionID,
		"query_type": query_type,
	}
	if query_type == "chat" {
		reqBody["message"] = message
	}

	// If report/prescription, include report_data
	if query_type == "report" {
		reqBody["report_data"] = reportData
	}
	if query_type == "prescription" {
		reqBody["prescription_data"] = reportData
	}

	jsonData, _ := json.Marshal(reqBody)
	prettyJSON, _ := json.MarshalIndent(reqBody, "", "  ")
	log.Println("AskAI reqBody body:\n", string(prettyJSON))
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call AI API: %w", err)
	}
	defer resp.Body.Close()

	var apiResp models.AskAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}
	log.Println("AskAPIResponse : ", apiResp)
	return &apiResp, nil
}

func (s *ApiServiceImpl) CheckPDFAndGetPassword(file io.Reader, fileName, emailBody string) (*models.PDFProtectionResult, error) {
	// Step 1: Check if PDF is protected
	isProtected, err := s.CallCheckPDFProtectionAPI(file, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to check PDF protection: %w", err)
	}

	if !isProtected {
		return &models.PDFProtectionResult{IsProtected: false}, nil
	}

	// Step 2: Get password from email body
	password, err := s.CallPDFPasswordAPI(emailBody)
	if err != nil {
		return nil, fmt.Errorf("failed to get PDF password: %w", err)
	}

	return &models.PDFProtectionResult{
		IsProtected: true,
		Password:    password,
	}, nil
}

func (s *ApiServiceImpl) CallCheckPDFProtectionAPI(file io.Reader, fileName string) (bool, error) {
	// Read first 512 bytes to detect MIME type
	buf := make([]byte, 512)
	n, err := io.ReadFull(file, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return false, fmt.Errorf("failed to read file for MIME detection: %w", err)
	}
	mimeType := http.DetectContentType(buf[:n])

	// Reset reader to include the bytes we already read + remaining content
	fileReader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	// Build multipart body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Explicit MIME headers for the file part
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileName))
	h.Set("Content-Type", mimeType)

	part, err := writer.CreatePart(h)
	if err != nil {
		return false, fmt.Errorf("failed to create multipart part: %w", err)
	}

	if _, err := io.Copy(part, fileReader); err != nil {
		return false, fmt.Errorf("failed to copy file to multipart: %w", err)
	}

	writer.Close()

	// Create request
	req, err := http.NewRequest(http.MethodPost, s.CheckPDFProtectionAPI, body)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse JSON response
	var result struct {
		Protected bool `json:"password_protected"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode failed: %w", err)
	}

	return result.Protected, nil
}

func (s *ApiServiceImpl) CallPDFPasswordAPI(emailBody string) (string, error) {
	reqBody := map[string]string{"body": emailBody}
	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(http.MethodPost, s.PDFPasswordAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Password, nil
}

func (s *ApiServiceImpl) CallDocumentTypeAPI(file io.Reader, filename string) (*models.DocTypeAPIResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// log.Println("File Name : ", filename)

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read file for MIME detection: %w", err)
	}
	mimeType := http.DetectContentType(buf[:n])
	// log.Printf("Detected MIME type: %s", mimeType)

	fileReader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filename))
	h.Set("Content-Type", mimeType)

	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("failed to create form part: %w", err)
	}

	if _, err := io.Copy(part, fileReader); err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", os.Getenv("DOCUMENT_TYPE_API"), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("service returned error: %s, body: %s", resp.Status, string(respBody))
	}

	var apiResp models.DocTypeAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}
	prettyJSON, err := json.MarshalIndent(apiResp, "", "  ")
	if err != nil {
		log.Println("Failed to format JSON:", err)
	} else {
		log.Println("Digitize report response (pretty):\n", string(prettyJSON))
	}
	if apiResp.Error != nil && *apiResp.Error != "" {
		return &apiResp, fmt.Errorf("API returned error: %s", *apiResp.Error)
	}
	if apiResp.Content == nil {
		return &apiResp, fmt.Errorf("API response missing content")
	}
	return &apiResp, nil
}
