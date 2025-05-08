package service

import (
	"biostat/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type ApiService interface {
	CallGeminiService(image io.Reader) (models.LabReport, error)
	CallSummarizeReportService(data models.PatientBasicInfo) (models.ResultSummary, error)
}

type ApiServiceImpl struct {
	GeminiAPIURL     string
	ReportSummaryAPI string
	client           *http.Client
}

func NewApiService() ApiService {
	return &ApiServiceImpl{
		GeminiAPIURL:     os.Getenv("GEMINI_API_URL"),
		ReportSummaryAPI: os.Getenv("REPORT_SUMMARY_API"),
		client:           &http.Client{},
	}
}

func (s *ApiServiceImpl) CallGeminiService(image io.Reader) (models.LabReport, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a form file field with dummy filename
	part, err := writer.CreateFormFile("image", "report_image.jpg")
	if err != nil {
		return models.LabReport{}, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, image); err != nil {
		return models.LabReport{}, fmt.Errorf("failed to copy image data: %w", err)
	}
	writer.Close()

	// Create and send request
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.LabReport{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return models.LabReport{}, fmt.Errorf("Python service returned error: %s, body: %s", resp.Status, string(respBody))
	}

	var reportData models.LabReport
	if err := json.Unmarshal(respBody, &reportData); err != nil {
		return models.LabReport{}, fmt.Errorf("failed to parse JSON response: %w, body: %s", err, string(respBody))
	}

	return reportData, nil
}

func (a *ApiServiceImpl) CallSummarizeReportService(data models.PatientBasicInfo) (models.ResultSummary, error) {
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
