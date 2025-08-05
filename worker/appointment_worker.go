package worker

import (
	"biostat/config"
	"biostat/constant"
	"biostat/models"
	"biostat/repository"
	"biostat/service"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/datatypes"
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
	healthMonitor     *service.HealthMonitorService
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
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: config.PropConfig.ApiURL.RedisURL})
	healthMonitor := service.NewHealthMonitorService(config.RedisClient, config.PropConfig.HealthCheck.URL, time.Duration(config.PropConfig.HealthCheck.IntervalSeconds)*time.Second, time.Duration(config.PropConfig.HealthCheck.TimeoutSeconds)*time.Second)
	healthMonitor.Start()
	worker := &DigitizationWorker{
		redisClient:       config.RedisClient,
		taskQueue:         client,
		apiService:        apiService,
		patientService:    patientService,
		diagnosticService: diagnosticService,
		recordRepo:        recordRepo,
		db:                db,
		healthMonitor:     healthMonitor,
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: config.PropConfig.ApiURL.RedisURL},
		asynq.Config{Concurrency: config.PropConfig.TaskQueue.ConcurrentTaskRun, RetryDelayFunc: func(n int, err error, t *asynq.Task) time.Duration {
			return time.Duration(config.PropConfig.Retry.MaxDelay)
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
	if !w.healthMonitor.IsServiceUp() {
		log.Println("AI service is down, retrying later.")
		_ = w.logAndUpdateStatus(ctx, p.RecordID, queueName, constant.StatusQueued, 0, &constant.ServiceError, retryCount)
		newTask := asynq.NewTask("digitize:record", t.Payload())
		_, err := w.taskQueue.Enqueue(newTask, asynq.ProcessIn(time.Duration(config.PropConfig.TaskQueue.Delay)))
		if err != nil {
			log.Printf("Failed to reschedule task for record %d: %v", p.RecordID, err)
			return err
		}
		return nil
	}

	log.Printf("Digitization started: recordId=%d queue Name := %s : retrying count := %d : record status := %s", p.RecordID, queueName, retryCount, status)

	_ = w.logAndUpdateStatus(ctx, p.RecordID, queueName, status, 0, nil, retryCount)

	fileBytes, err := os.ReadFile(p.FilePath)
	if err != nil {
		return w.failTask(ctx, queueName, p.RecordID, "Failed to read file", retryCount)
	}

	fileBuf := bytes.NewBuffer(fileBytes)

	switch p.Category {
	case "Test Reports", "test_report":
		if err := w.handleTestReport(fileBuf, p); err != nil {
			return w.failTask(ctx, queueName, p.RecordID, err.Error(), retryCount)
		}
	case "Prescriptions":
		if err := w.handlePrescription(fileBuf, p); err != nil {
			return w.failTask(ctx, queueName, p.RecordID, err.Error(), retryCount)
		}
	}

	if err := w.logAndUpdateStatus(ctx, p.RecordID, queueName, constant.StatusSuccess, 1, nil, retryCount); err != nil {
		return err
	}

	log.Printf("Digitization success: recordId=%d queue Name := %s : retrying count := %d : record status := %s", p.RecordID, queueName, retryCount, constant.StatusSuccess)
	_ = os.Remove(p.FilePath)
	return nil
}

func (w *DigitizationWorker) logAndUpdateStatus(ctx context.Context, recordID uint64, queue string,
	status constant.JobStatus, flag int, errMsg *string, retryCount int) error {
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
		if errMsg != nil {
			update.ErrorMessage = *errMsg
		} else {
			update.ErrorMessage = ""
		}
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

func (w *DigitizationWorker) failTask(ctx context.Context, queueName string, recordID uint64, msg string, retryCount int) error {
	log.Printf("Digitization failed: recordId=%d queue Name := %s : retrying count := %d : record status := %s : error=%s", recordID, queueName, retryCount, constant.StatusFailed, msg)
	_ = w.logAndUpdateStatus(ctx, recordID, queueName, constant.StatusFailed, 0, &msg, retryCount)
	return fmt.Errorf("digitization failed: %s queue Name := %s", msg, queueName)
}

func (w *DigitizationWorker) handleTestReport(fileBuf *bytes.Buffer, p models.DigitizationPayload) error {
	reportData, err := w.apiService.CallGeminiService(fileBuf, p.FileName)
	if err != nil {
		return err
	}

	relatives, _ := w.patientService.GetRelativeList(&p.UserID)
	matchedUserID := p.UserID
	var isUnknownReport bool
	if reportData.ReportDetails.PatientName != "" {
		matchedUserID, isUnknownReport = service.MatchPatientNameWithRelative(relatives, reportData.ReportDetails.PatientName, p.UserID, p.PatientName)
		config.Log.Info("MatchPatientNameWithRelative ", zap.Bool("Is Unknown Report Found", isUnknownReport))
		if matchedUserID != p.UserID || isUnknownReport {
			tx := w.db.Begin()
			err := w.recordRepo.UpdateMedicalRecordMappingByRecordId(tx, &p.RecordID, map[string]interface{}{"user_id": matchedUserID, "is_unknown_record": isUnknownReport})
			if err != nil {
				return err
			}
			if err := tx.Commit().Error; err != nil {
				return err
			}
		}
	}
	reportData.ReportDetails.IsDigital = true
	reportData.ReportDetails.IsUnknownRecord = isUnknownReport
	if jsonBytes, err := json.Marshal(reportData); err == nil {
		updateRecord := &models.TblMedicalRecord{
			RecordId: p.RecordID,
			Metadata: datatypes.JSON(jsonBytes),
		}
		if isUnknownReport {
			updateRecord.RecordCategory = string(constant.OTHER)
		}
		_, _ = w.recordRepo.UpdateTblMedicalRecord(updateRecord)
	}
	if isUnknownReport {
		if _, err := w.diagnosticService.DigitizeDiagnosticReport(reportData, matchedUserID, &p.RecordID); err != nil {
			return err
		}
	} else {
		if err := w.diagnosticService.CheckReportExistWithSampleDateTestComponent(reportData, matchedUserID, &p.RecordID); err != nil {
			return err
		}
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
