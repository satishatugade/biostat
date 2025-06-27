package worker

import (
	"biostat/config"
	"biostat/constant"
	"biostat/models"
	"biostat/repository"
	"biostat/service"
	"biostat/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func StartAppointmentScheduler(service service.AppointmentService) {
	log.Println("Appointment Schedular running")

	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		for range ticker.C {
			processAppointments(service)
		}
	}()
}

func processAppointments(s service.AppointmentService) {
	log.Println("Scheduler: checking for completed appointments ...")
	err := s.MarkCompletedAppointments()
	if err != nil {
		log.Println("Error @ MarkCompletedAppointments", err)
	}
}

type DigitizationWorker struct {
	redisClient       *redis.Client
	apiService        service.ApiService
	patientService    service.PatientService
	diagnosticService service.DiagnosticService
	recordRepo        repository.TblMedicalRecordRepository
}

func InitAsynqWorker(
	apiService service.ApiService,
	patientService service.PatientService,
	diagnosticService service.DiagnosticService,
	recordRepo repository.TblMedicalRecordRepository,
) {
	worker := &DigitizationWorker{
		redisClient:       config.RedisClient,
		apiService:        apiService,
		patientService:    patientService,
		diagnosticService: diagnosticService,
		recordRepo:        recordRepo,
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: os.Getenv("REDIS_ADDR")},
		asynq.Config{Concurrency: utils.GetConcurrentTaskCount()},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("digitize:record", worker.HandleDigitizationTask)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("Could not run Asynq server: %v", err)
	}
}
func (w *DigitizationWorker) HandleDigitizationTask(ctx context.Context, t *asynq.Task) error {
	var p service.DigitizationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	queueName := "report_digitization"
	if p.Category == "Prescriptions" {
		queueName = "prescription_digitization"
	}
	log.Printf("Digitization started: recordId=%d queue Name := %s ", p.RecordID, queueName)

	_ = w.logAndUpdateStatus(ctx, p.RecordID, queueName, constant.StatusProcessing, 0, nil, p.AuthUserID)

	fileBytes, err := os.ReadFile(p.FilePath)
	if err != nil {
		return w.failTask(ctx, queueName, p.RecordID, "Failed to read file", p.AuthUserID)
	}

	fileBuf := bytes.NewBuffer(fileBytes)

	switch p.Category {
	case "Test Reports":
		if err := w.handleTestReport(fileBuf, p); err != nil {
			return w.failTask(ctx, queueName, p.RecordID, err.Error(), p.AuthUserID)
		}
	case "Prescriptions":
		if err := w.handlePrescription(fileBuf, p); err != nil {
			return w.failTask(ctx, queueName, p.RecordID, err.Error(), p.AuthUserID)
		}
	}

	if err := w.logAndUpdateStatus(ctx, p.RecordID, queueName, constant.StatusSuccess, 1, nil, p.AuthUserID); err != nil {
		return err
	}

	log.Printf("Digitization success: recordId=%d queue Name := %s ", p.RecordID, queueName)
	_ = os.Remove(p.FilePath)
	return nil
}

func (w *DigitizationWorker) logAndUpdateStatus(ctx context.Context, recordID uint64, queue string,
	status constant.JobStatus, flag int, errMsg *string, authUserID string) error {
	now := time.Now()

	update := &models.TblMedicalRecord{
		RecordId:     recordID,
		DigitizeFlag: flag,
		Status:       status,
		QueueName:    queue,
	}

	switch status {
	case constant.StatusProcessing:
		update.ProcessingStartedAt = &now
	case constant.StatusSuccess, constant.StatusFailed:
		update.CompletedAt = &now
	}

	if status == constant.StatusFailed {
		update.NextRetryAt = ptrTime(now.Add(5 * time.Minute))
	}

	if errMsg != nil {
		update.ErrorMessage = *errMsg
	}

	if _, err := w.recordRepo.UpdateTblMedicalRecord(update, authUserID); err != nil {
		return err
	}

	w.redisClient.Set(ctx, fmt.Sprintf("record_status:%d", recordID), status, 0)
	w.redisClient.Set(ctx, fmt.Sprintf("record_queue:%d", recordID), queue, 0)

	return nil
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func (w *DigitizationWorker) failTask(ctx context.Context, queueName string, recordID uint64, msg, authUserID string) error {
	log.Printf("Digitization failed: recordId=%d queue Name := %s error=%s", recordID, queueName, msg)
	_ = w.logAndUpdateStatus(ctx, recordID, queueName, constant.StatusFailed, 0, &msg, authUserID)
	return fmt.Errorf("digitization failed: %s queue Name := %s", msg, queueName)
}

func (w *DigitizationWorker) handleTestReport(fileBuf *bytes.Buffer, p service.DigitizationPayload) error {
	reportData, err := w.apiService.CallGeminiService(fileBuf, p.FileName)
	if err != nil {
		return err
	}

	relatives, _ := w.patientService.GetRelativeList(&p.UserID)
	matchedUserID := p.UserID
	if reportData.ReportDetails.PatientName != "" {
		matchedUserID = service.MatchPatientNameWithRelative(relatives, reportData.ReportDetails.PatientName, p.UserID)
		if matchedUserID != p.UserID {
			err := w.recordRepo.UpdateMedicalRecordMappingByRecordId(&p.RecordID, map[string]interface{}{
				"user_id": matchedUserID,
			})
			if err != nil {
				return err
			}
		}
	}
	reportData.ReportDetails.IsDigital = true

	if _, err := w.diagnosticService.DigitizeDiagnosticReport(reportData, matchedUserID, &p.RecordID); err != nil {
		return err
	}

	return w.diagnosticService.NotifyAbnormalResult(matchedUserID)
}

func (w *DigitizationWorker) handlePrescription(fileBuf *bytes.Buffer, p service.DigitizationPayload) error {
	data, err := w.apiService.CallPrescriptionDigitizeAPI(fileBuf, p.FileName)
	if err != nil {
		return err
	}

	data.PatientId = p.UserID
	data.RecordId = p.RecordID
	data.IsDigital = true

	return w.patientService.AddPatientPrescription(p.AuthUserID, &data)
}
