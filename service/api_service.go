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
	"time"
)

type ApiService interface {
	CallGeminiService(file io.Reader, filename string) (models.LabReport, error)
	CallPrescriptionDigitizeAPI(file io.Reader, filename string) (models.PatientPrescription, error)
	CallSummarizeReportService(data models.PatientBasicInfo) (models.ResultSummary, error)
	AnalyzePrescriptionWithAI(data models.PatientPrescription) (string, error)
	AnalyzePharmacokineticsInfo(data models.PharmacokineticsInput) (string, error)
}

type ApiServiceImpl struct {
	GeminiAPIURL            string
	ReportSummaryAPI        string
	PrescriptionAPI         string
	PharmacokineticsAPI     string
	DigitizePrescriptionAPI string
	client                  *http.Client
}

func NewApiService() ApiService {
	return &ApiServiceImpl{
		GeminiAPIURL:            os.Getenv("GEMINI_API_URL"),
		ReportSummaryAPI:        os.Getenv("REPORT_SUMMARY_API"),
		PrescriptionAPI:         os.Getenv("PRESCRIPTION_API"),
		PharmacokineticsAPI:     os.Getenv("PHARMACOKINETICS_API"),
		DigitizePrescriptionAPI: os.Getenv("DIGITIZE_PRESCRIPTION_API"),
		client:                  &http.Client{},
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

	// Detect MIME type
	buf := make([]byte, 512)
	n, err1 := file.Read(buf)
	if err1 != nil {
		return models.PatientPrescription{}, fmt.Errorf("failed to read file for MIME detection: %w", err1)
	}
	mimeType := http.DetectContentType(buf[:n])
	log.Printf("Detected MIME type: %s", mimeType)

	// Reconstruct full reader
	fileReader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	// Create MIME header for file part
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

	// Create request
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

	// Try parsing various date formats
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
		"2006-01-02T15:04:05",       // ISO local
		"2006-01-02T15:04:05Z07:00", // RFC3339
	}

	for _, format := range dateFormats {
		parsedDate, err = time.Parse(format, prescriptionData.PrescriptionDate)
		if err == nil {
			break
		}
	}

	if err != nil {
		log.Println("Invalid date format:", err)
		return models.PatientPrescription{}, fmt.Errorf("invalid date format: %w", err)
	}

	// Now map everything
	data := models.PatientPrescription{
		PrescriptionId:            prescriptionData.PrescriptionId,
		PatientId:                 prescriptionData.PatientId,
		PrescribedBy:              prescriptionData.PrescribedBy,
		PrescriptionName:          prescriptionData.PrescriptionName,
		Description:               prescriptionData.Description,
		PrescriptionDate:          parsedDate,
		PrescriptionAttachmentUrl: prescriptionData.PrescriptionAttachmentUrl,
		PrescriptionDetails:       make([]models.PrescriptionDetail, 0),
	}

	// Map prescription details if any
	for _, detail := range prescriptionData.PrescriptionDetails {
		d := models.PrescriptionDetail{
			PrescriptionDetailId: detail.PrescriptionDetailId,
			PrescriptionId:       detail.PrescriptionId,
			MedicineName:         detail.MedicineName,
			PrescriptionType:     detail.PrescriptionType,
			DoseQuantity:         detail.DoseQuantity,
			Duration:             detail.Duration,
			UnitValue:            detail.UnitValue,
			UnitType:             detail.UnitType,
			Frequency:            detail.Frequency,
			TimesPerDay:          detail.TimesPerDay,
			IntervalHour:         detail.IntervalHour,
			Instruction:          detail.Instruction,
			CreatedAt:            detail.CreatedAt,
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
		return "", fmt.Errorf("Pharmacokinetics AI service responded with status code %d", resp.StatusCode)
	}

	var response struct {
		Summary string `json:"analysis"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.Summary, nil
}
