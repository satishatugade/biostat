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
	"net/http"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
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
	// log.Println("Scheduler: checking for completed appointments ...")
	err := s.MarkCompletedAppointments()
	if err != nil {
		log.Println("Error @ MarkCompletedAppointments", err)
	}
}

type DigitizationWorker struct {
	redisClient       *redis.Client
	taskQueue         *asynq.Client
	apiService        service.ApiService
	patientService    service.PatientService
	diagnosticService service.DiagnosticService
	recordRepo        repository.TblMedicalRecordRepository
	db                *gorm.DB
}

func NewDigitizationWorker(db *gorm.DB) *DigitizationWorker {
	if db == nil {
		panic("database instance is null")
	}
	return &DigitizationWorker{db: db}
}

func InitAsynqWorker(
	apiService service.ApiService,
	patientService service.PatientService,
	diagnosticService service.DiagnosticService,
	recordRepo repository.TblMedicalRecordRepository,
	db *gorm.DB,
) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: os.Getenv("REDIS_ADDR")})
	worker := &DigitizationWorker{
		redisClient:       config.RedisClient,
		taskQueue:         client,
		apiService:        apiService,
		patientService:    patientService,
		diagnosticService: diagnosticService,
		recordRepo:        recordRepo,
		db:                db,
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: os.Getenv("REDIS_ADDR")},
		asynq.Config{Concurrency: utils.GetConcurrentTaskCount(), RetryDelayFunc: func(n int, err error, t *asynq.Task) time.Duration {
			return 5 * time.Minute
		}},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("digitize:record", worker.HandleDigitizationTask)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("Could not run Asynq server: %v", err)
	}
}

func (w *DigitizationWorker) HandleDigitizationTask(ctx context.Context, t *asynq.Task) error {
	var p models.DigitizationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	queueName := "report_digitization"
	if p.Category == "Prescriptions" {
		queueName = "prescription_digitization"
	}
	retryCount, _ := asynq.GetRetryCount(ctx)
	status := constant.StatusProcessing
	if retryCount > 0 {
		status = constant.StatusRetrying
	}
	if !CheckServiceHealth(w.redisClient, os.Getenv("SERVICE_HEALTH_CHECK")) {
		log.Println("AI service is down, retrying later.")
		_ = w.logAndUpdateStatus(ctx, p.RecordID, queueName, constant.StatusQueued, 0, &constant.ServiceError, p.AuthUserID, retryCount)
		newTask := asynq.NewTask("digitize:record", t.Payload())
		_, err := w.taskQueue.Enqueue(newTask, asynq.ProcessIn(2*time.Minute))
		if err != nil {
			log.Printf("Failed to reschedule task for record %d: %v", p.RecordID, err)
			return err
		}

		return nil
	}
	log.Printf("Digitization started: recordId=%d queue Name := %s : retrying count := %d : record status := %s", p.RecordID, queueName, retryCount, status)

	_ = w.logAndUpdateStatus(ctx, p.RecordID, queueName, status, 0, nil, p.AuthUserID, retryCount)

	fileBytes, err := os.ReadFile(p.FilePath)
	if err != nil {
		return w.failTask(ctx, queueName, p.RecordID, "Failed to read file", p.AuthUserID, retryCount)
	}

	fileBuf := bytes.NewBuffer(fileBytes)

	switch p.Category {
	case "Test Reports", "test_report":
		if err := w.handleTestReport(fileBuf, p); err != nil {
			return w.failTask(ctx, queueName, p.RecordID, err.Error(), p.AuthUserID, retryCount)
		}
	case "Prescriptions":
		if err := w.handlePrescription(fileBuf, p); err != nil {
			return w.failTask(ctx, queueName, p.RecordID, err.Error(), p.AuthUserID, retryCount)
		}
	}

	if err := w.logAndUpdateStatus(ctx, p.RecordID, queueName, constant.StatusSuccess, 1, nil, p.AuthUserID, retryCount); err != nil {
		return err
	}

	log.Printf("Digitization success: recordId=%d queue Name := %s : retrying count := %d : record status := %s", p.RecordID, queueName, retryCount, status)
	_ = os.Remove(p.FilePath)
	return nil
}

func (w *DigitizationWorker) logAndUpdateStatus(ctx context.Context, recordID uint64, queue string,
	status constant.JobStatus, flag int, errMsg *string, authUserID string, retryCount int) error {
	now := time.Now()
	update := &models.TblMedicalRecord{
		RecordId:     recordID,
		DigitizeFlag: flag,
		Status:       status,
		QueueName:    queue,
		RetryCount:   retryCount,
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

	if _, err := w.recordRepo.UpdateTblMedicalRecord(update); err != nil {
		return err
	}

	w.redisClient.Set(ctx, fmt.Sprintf("record_status:%d", recordID), status, 0)
	w.redisClient.Set(ctx, fmt.Sprintf("record_queue:%d", recordID), queue, 0)

	return nil
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func (w *DigitizationWorker) failTask(ctx context.Context, queueName string, recordID uint64, msg, authUserID string, retryCount int) error {
	log.Printf("Digitization failed: recordId=%d queue Name := %s : retrying count := %d : record status := %s : error=%s", recordID, queueName, retryCount, constant.StatusFailed, msg)
	_ = w.logAndUpdateStatus(ctx, recordID, queueName, constant.StatusFailed, 0, &msg, authUserID, retryCount)
	return fmt.Errorf("digitization failed: %s queue Name := %s", msg, queueName)
}

func (w *DigitizationWorker) handleTestReport(fileBuf *bytes.Buffer, p models.DigitizationPayload) error {
	reportData, err := w.apiService.CallGeminiService(fileBuf, p.FileName)
	if err != nil {
		return err
	}

	relatives, _ := w.patientService.GetRelativeList(&p.UserID)
	matchedUserID := p.UserID
	if reportData.ReportDetails.PatientName != "" {
		matchedUserID = service.MatchPatientNameWithRelative(relatives, reportData.ReportDetails.PatientName, p.UserID, p.PatientName)
		if matchedUserID != p.UserID {
			tx := w.db.Begin()
			err := w.recordRepo.UpdateMedicalRecordMappingByRecordId(tx, &p.RecordID, map[string]interface{}{"user_id": matchedUserID})
			if err != nil {
				return err
			}
			if err := tx.Commit().Error; err != nil {
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

func (w *DigitizationWorker) handlePrescription(fileBuf *bytes.Buffer, p models.DigitizationPayload) error {
	data, err := w.apiService.CallPrescriptionDigitizeAPI(fileBuf, p.FileName)
	if err != nil {
		return err
	}

	data.PatientId = p.UserID
	data.RecordId = p.RecordID
	data.IsDigital = true

	return w.patientService.AddPatientPrescription(p.AuthUserID, &data)
}

func CheckServiceHealth(redisClient *redis.Client, serviceURL string) bool {
	log.Println("CheckServiceHealth for AI service ....")
	ctx := context.Background()

	// Check Redis cache first
	key := fmt.Sprintf("ai-service_status:%s", serviceURL)
	cachedStatus, err := redisClient.Get(ctx, key).Result()
	if err == nil {
		if cachedStatus == "up" {
			log.Println("[CheckServiceHealth] Service is marked UP in cache.")
			return true
		}
		if cachedStatus == "down" {
			log.Println("[CheckServiceHealth] Service is marked DOWN in cache.")
			return false
		}
	}

	// If no cache, hit the actual health endpoint
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(serviceURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		// Service DOWN cache result for 2 minutes
		log.Println("CheckServiceHealth Marking service as DOWN in cache for 2 minutes.")
		_ = redisClient.Set(ctx, key, "down", 2*time.Minute).Err()
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("CheckServiceHealth Marking service as DOWN in cache for 2 minutes RESPONSE CODE : %d ", resp.StatusCode)
		_ = redisClient.Set(ctx, key, "down", 2*time.Minute).Err()
		return false
	}

	log.Println("CheckServiceHealth Service is UP. Caching status for 5 minutes.")
	_ = redisClient.Set(ctx, key, "up", 5*time.Minute).Err()
	return true
}
