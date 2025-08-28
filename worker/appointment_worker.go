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
	"strconv"
	"time"

	"github.com/google/uuid"
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
	redisClient          *redis.Client
	taskQueue            *asynq.Client
	apiService           service.ApiService
	patientService       service.PatientService
	diagnosticService    service.DiagnosticService
	recordRepo           repository.TblMedicalRecordRepository
	db                   *gorm.DB
	healthMonitor        *service.HealthMonitorService
	processStatusService service.ProcessStatusService
	gmailService         service.GmailSyncService
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
	processStatusService service.ProcessStatusService,
	gmailService service.GmailSyncService,

) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: config.PropConfig.ApiURL.RedisURL})
	healthMonitor := service.NewHealthMonitorService(config.RedisClient, config.PropConfig.HealthCheck.URL, time.Duration(config.PropConfig.HealthCheck.IntervalSeconds)*time.Second, time.Duration(config.PropConfig.HealthCheck.TimeoutSeconds)*time.Second)
	healthMonitor.Start()
	worker := &DigitizationWorker{
		redisClient:          config.RedisClient,
		taskQueue:            client,
		apiService:           apiService,
		patientService:       patientService,
		diagnosticService:    diagnosticService,
		recordRepo:           recordRepo,
		db:                   db,
		healthMonitor:        healthMonitor,
		processStatusService: processStatusService,
		gmailService:         gmailService,
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: config.PropConfig.ApiURL.RedisURL},
		asynq.Config{Concurrency: config.PropConfig.TaskQueue.ConcurrentTaskRun, RetryDelayFunc: func(n int, err error, t *asynq.Task) time.Duration {
			return time.Duration(config.PropConfig.Retry.MaxDelay)
		}},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("digitize:record", worker.HandleDigitizationTask)
	mux.HandleFunc("check:doctype", worker.HandleDocTypeCheckTask)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("Could not run Asynq server: %v", err)
	}
}

func (w *DigitizationWorker) HandleDigitizationTask(ctx context.Context, t *asynq.Task) error {
	var p models.DigitizationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	step := string(constant.DocsDigitization)
	msg := string(constant.DocsDigitizationMsg)
	errorMsg := ""
	w.processStatusService.LogStep(p.ProcessID, step, constant.Running, msg, errorMsg, &p.RecordID, nil, nil, nil, nil, nil)
	queueName := "report_digitization"
	if p.Category == string(constant.MEDICATION) {
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
		return w.failTask(ctx, queueName, p.ProcessID, p.RecordID, "Failed to read file", retryCount)
	}

	fileBuf := bytes.NewBuffer(fileBytes)
	switch p.Category {
	case string(constant.TESTREPORT):
		if err := w.handleTestReport(fileBuf, p); err != nil {
			return w.failTask(ctx, queueName, p.ProcessID, p.RecordID, err.Error(), retryCount)
		}
	case string(constant.MEDICATION):
		if err := w.handlePrescription(fileBuf, p); err != nil {
			return w.failTask(ctx, queueName, p.ProcessID, p.RecordID, err.Error(), retryCount)
		}
	}

	if err := w.logAndUpdateStatus(ctx, p.RecordID, queueName, constant.StatusSuccess, 1, nil, retryCount); err != nil {
		return err
	}

	w.processStatusService.LogStep(p.ProcessID, step, constant.Success, msg, errorMsg, &p.RecordID, nil, nil, nil, nil, nil)
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

func (w *DigitizationWorker) failTask(ctx context.Context, queueName string, processID uuid.UUID, recordID uint64, msg string, retryCount int) error {
	log.Printf("Digitization failed: recordId=%d queue Name := %s : retrying count := %d : record status := %s : error=%s", recordID, queueName, retryCount, constant.StatusFailed, msg)
	_ = w.logAndUpdateStatus(ctx, recordID, queueName, constant.StatusFailed, 0, &msg, retryCount)
	err := fmt.Errorf("digitization failed: %s queue Name := %s", msg, queueName)
	w.processStatusService.LogStepAndFail(processID, string(constant.DocsDigitization), constant.Failure, string(constant.DigitizationFailed), err.Error(), nil, &recordID, nil)
	return err
}

func (w *DigitizationWorker) handleTestReport(fileBuf *bytes.Buffer, p models.DigitizationPayload) error {
	step := string(constant.CallAIService)
	errorMsg := ""
	w.processStatusService.LogStep(p.ProcessID, step, constant.Running, string(constant.CallingAIServiceMsg), errorMsg, &p.RecordID, nil, nil, nil, nil, p.AttachmentId)
	reportData, err := w.apiService.CallGeminiService(fileBuf, p.FileName)
	if err != nil {
		aiResMsg := fmt.Sprintf("Processed record id %d %s %s %s", p.RecordID, p.Category, p.FileName, string(constant.CallingAIFailed))
		w.processStatusService.LogStepAndFail(p.ProcessID, step, constant.Failure, aiResMsg, err.Error(), nil, &p.RecordID, p.AttachmentId)
		return err
	}
	aiResMsg := fmt.Sprintf("Processed record id %d %s %s %s", p.RecordID, p.Category, p.FileName, string(constant.CallingAIServiceSuccess))
	w.processStatusService.LogStep(p.ProcessID, step, constant.Success, aiResMsg, errorMsg, &p.RecordID, nil, nil, nil, nil, p.AttachmentId)
	matchedUserID := p.UserID
	var isUnknownReport bool
	var matchName string
	var matchMessage string
	var patientNameOnReport string
	var apiResp *models.PatientDocResponse
	var apiErr error
	if reportData.ReportDetails.PatientName != "" {
		step := string(constant.MatchingReport)
		msg := string(constant.MatchingNameMsg)
		w.processStatusService.LogStep(p.ProcessID, step, constant.Running, msg, errorMsg, &p.RecordID, nil, nil, nil, nil, p.AttachmentId)
		patientNameOnReport = reportData.ReportDetails.PatientName
		apiResp, apiErr = w.gmailService.GetPatientNameOnDoc(patientNameOnReport, p.UserID)
		if apiErr != nil {
			config.Log.Error("gmailService.GetPatientNameOnDoc ERROR", zap.Error(apiErr))
			relatives, _ := w.patientService.GetRelativeList(&p.UserID)
			matchedUserID, matchName, isUnknownReport, matchMessage = service.MatchPatientNameWithRelative(relatives, reportData.ReportDetails.PatientName, p.UserID, p.PatientName)
			config.Log.Info("MatchPatientNameWithRelative ", zap.Bool("Is Unknown Report Found", isUnknownReport))
			config.Log.Info("Match Name found:", zap.String("matchName", matchName))
			if matchedUserID != p.UserID || isUnknownReport {
				w.processStatusService.LogStep(p.ProcessID, step, constant.Success, matchMessage, errorMsg, &p.RecordID, nil, nil, nil, nil, p.AttachmentId)
				tx := w.db.Begin()
				err := w.recordRepo.UpdateMedicalRecordMappingByRecordId(tx, &p.RecordID, map[string]interface{}{"user_id": matchedUserID, "is_unknown_record": isUnknownReport})
				if err != nil {
					return err
				}
				if err := tx.Commit().Error; err != nil {
					return err
				}
			}
			msg = fmt.Sprintf("Processed record id %d : %s , Patient Name on report  %s", p.RecordID, matchMessage, reportData.ReportDetails.PatientName)
			w.processStatusService.LogStep(p.ProcessID, step, constant.Success, msg, errorMsg, &p.RecordID, nil, nil, nil, nil, p.AttachmentId)
			reportData.ReportDetails.IsUnknownRecord = isUnknownReport
		} else {
			config.Log.Info("MatchPatientNameWithRelative ", zap.Bool("Is Unknown Report Found", apiResp.IsFallback))
			if apiResp.MatchedUserID != p.UserID || apiResp.IsFallback {
				matchMessage = fmt.Sprintf("Fallback :%t | Match userId :%d | Patient name on report %s | Name match with user: %s ", apiResp.IsFallback, apiResp.MatchedUserID, patientNameOnReport, apiResp.FinalPatientName)
				w.processStatusService.LogStep(p.ProcessID, step, constant.Success, matchMessage, errorMsg, &p.RecordID, nil, nil, nil, nil, p.AttachmentId)
				tx := w.db.Begin()
				err := w.recordRepo.UpdateMedicalRecordMappingByRecordId(tx, &p.RecordID, map[string]interface{}{"user_id": apiResp.MatchedUserID, "is_unknown_record": apiResp.IsFallback})
				if err != nil {
					return err
				}
				if err := tx.Commit().Error; err != nil {
					return err
				}
			}
			// msg = fmt.Sprintf("Processed record id %d : %s , Patient Name on report  %s", p.RecordID, matchMessage, reportData.ReportDetails.PatientName)
			w.processStatusService.LogStep(p.ProcessID, step, constant.Success, matchMessage, errorMsg, &p.RecordID, nil, nil, nil, nil, p.AttachmentId)
			reportData.ReportDetails.IsUnknownRecord = apiResp.IsFallback
		}
	}
	reportData.ReportDetails.IsDigital = true
	aiMetadata := map[string]interface{}{
		"ai": reportData,
	}
	if jsonBytes, err := json.Marshal(aiMetadata); err == nil {
		updateRecord := &models.TblMedicalRecord{
			RecordId: p.RecordID,
			Metadata: datatypes.JSON(jsonBytes),
		}
		if isUnknownReport || apiResp.IsFallback {
			updateRecord.RecordCategory = string(constant.OTHER)
		}
		_, err := w.recordRepo.UpdateTblMedicalRecord(updateRecord)
		if err != nil {
			log.Println("Error Worker updaing Record @UpdateTblMedicalRecord ", err)
		}
	}
	if isUnknownReport {
		if _, err := w.diagnosticService.DigitizeDiagnosticReport(reportData, matchedUserID, &p.RecordID); err != nil {
			return err
		}
	} else {
		if err := w.diagnosticService.CheckReportExistWithSampleDateTestComponent(reportData, matchedUserID, &p.RecordID, p.ProcessID, p.AttachmentId); err != nil {
			return err
		}
	}
	if reportData.ReportDetails.ReportDate != "" {
		reportDate, err := utils.ParseDate(reportData.ReportDetails.ReportDate)
		if err != nil {
			log.Println("ReportDate parsing failed:", err)
			return nil
		}
		daysSinceReport := time.Since(reportDate).Hours() / 24
		if daysSinceReport <= 7 {
			return w.diagnosticService.NotifyAbnormalResult(matchedUserID)
		}
	}
	return nil
}

func (w *DigitizationWorker) handlePrescription(fileBuf *bytes.Buffer, p models.DigitizationPayload) error {
	errorMsg := ""
	step := string(constant.CallAIService)
	w.processStatusService.LogStep(p.ProcessID, step, constant.Running, string(constant.CallingAIServiceMsg), errorMsg, &p.RecordID, nil, nil, nil, nil, p.AttachmentId)
	data, err := w.apiService.CallPrescriptionDigitizeAPI(fileBuf, p.FileName)
	if err != nil {
		w.processStatusService.LogStepAndFail(p.ProcessID, step, constant.Failure, "Prescription medication digitization failed", err.Error(), nil, &p.RecordID, p.AttachmentId)
		return err
	}

	data.PatientId = p.UserID
	data.RecordId = p.RecordID
	data.IsDigital = true
	userId := strconv.FormatUint(p.UserID, 10)
	PrescMediErr := w.patientService.AddPatientPrescription(userId, &data)
	if PrescMediErr != nil {
		w.processStatusService.LogStepAndFail(p.ProcessID, step, constant.Failure, "Failed to save prescrition in database", err.Error(), nil, &p.RecordID, p.AttachmentId)
	}
	w.processStatusService.LogStep(p.ProcessID, step, constant.Success, "Prescription saved succesfully", errorMsg, &p.RecordID, nil, nil, nil, nil, p.AttachmentId)
	return nil
}

func (w *DigitizationWorker) HandleDocTypeCheckTask(ctx context.Context, t *asynq.Task) error {
	var payload models.DocTypeCheckPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	var docTypeResp *models.DocTypeAPIResponse
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		docTypeResp, err = w.apiService.CallDocumentTypeAPI(bytes.NewReader(payload.FileBytes), payload.FileName)
		if err == nil {
			break
		}
		log.Printf("[Attempt %d] Doc type API failed for record %s: %v : %v", attempt, payload.AttachmentID, err, docTypeResp)
		time.Sleep(5 * time.Second)
	}

	// log.Printf("Doc type API response for record %s: %+v", payload.AttachmentID, docTypeResp)
	SendDocTypeResult(payload.AttachmentID, docTypeResp)
	return nil
}
func SendDocTypeResult(AttachmentID string, result *models.DocTypeAPIResponse) {
	service.DocTypeResponses.Lock()
	defer service.DocTypeResponses.Unlock()
	if ch, ok := service.DocTypeResponses.Data[AttachmentID]; ok {
		// ch <- result.Content.LLMClassifier.DocumentType
		// ch <- result.Content.RegexClassifier.DocumentType
		ch <- result
		close(ch)
		delete(service.DocTypeResponses.Data, AttachmentID)
	}
}
