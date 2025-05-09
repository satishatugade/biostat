package service

import (
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
)

type ApiService interface {
	CallGeminiService(file io.Reader, filename string) (models.LabReport, error)
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
