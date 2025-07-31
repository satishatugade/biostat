package config

import (
	"os"
	"strconv"
)

func getEnv(key string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return ""
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key)
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	valPtr := getEnv(key)
	if valPtr != "" {
		switch valPtr {
		case "true", "TRUE", "1":
			return true
		case "false", "FALSE", "0":
			return false
		}
	}
	return defaultVal
}

type PropertyConfig struct {
	HealthCheck struct {
		Enabled         bool
		URL             string
		IntervalSeconds int
		TimeoutSeconds  int
		MaxRetries      int
		RetryDelay      int
		Expiration      int
	}
	Retry struct {
		Enabled      bool
		MaxAttempts  int
		Strategy     string
		InitialDelay int
		MaxDelay     int
	}
	TaskQueue struct {
		ConcurrentTaskRun int
		RetryCount        int
		Delay             int
		Expiration        int
		Retention         int
	}
	SystemVaribale struct {
		Score int
	}
	Database struct {
		Host     string
		Port     string
		DBName   string
		UserName string
		Password string
		SSLMode  string
	}
	ApiURL struct {
		RedisURL                    string
		NotifyServerURL             string
		ShareReportBaseURL          string
		ShortBaseURL                string
		AskChatBotURL               string
		ReportSummaryURL            string
		SummaryHistoryURL           string
		PrescriptionURL             string
		PrescriptionDigitizationURL string
		PharmacokineticsURL         string
		DigilockerRedirectURL       string
		GoogleRedirectURL           string
		ReportDigitizationURL       string
	}
}

var PropConfig *PropertyConfig = LoadConfigFromEnv()

func LoadConfigFromEnv() *PropertyConfig {
	cfg := &PropertyConfig{}
	cfg.SystemVaribale.Score = getEnvAsInt("SYSTEM_DUPLICATE_REPORT_SCORE", 60)

	cfg.Database.Host = getEnv("DB_HOST")
	cfg.Database.Port = getEnv("DB_PORT")
	cfg.Database.DBName = getEnv("DB_NAME")
	cfg.Database.UserName = getEnv("DB_USER")
	cfg.Database.Password = getEnv("DB_PASSWORD")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE")

	// HealthCheck Config
	cfg.HealthCheck.Enabled = getEnvAsBool("HEALTHCHECK_ENABLED", true)
	cfg.HealthCheck.URL = getEnv("SERVICE_HEALTH_CHECK_URL")
	cfg.HealthCheck.IntervalSeconds = getEnvAsInt("HEALTHCHECK_INTERVAL_SECONDS", 60)
	cfg.HealthCheck.TimeoutSeconds = getEnvAsInt("HEALTHCHECK_TIMEOUT_SECONDS", 30)
	cfg.HealthCheck.MaxRetries = getEnvAsInt("HEALTHCHECK_MAX_RETRIES", 3)
	cfg.HealthCheck.RetryDelay = getEnvAsInt("HEALTHCHECK_RETRY_DELAY_SECONDS", 120)
	cfg.HealthCheck.Expiration = getEnvAsInt("HEALTHCHECK_EXPIRATION_SECONDS", 300)

	// Retry Config
	cfg.Retry.Enabled = getEnvAsBool("RETRY_ENABLED", true)
	cfg.Retry.MaxAttempts = getEnvAsInt("RETRY_MAX_ATTEMPTS", 3)
	cfg.Retry.Strategy = getEnv("RETRY_BACKOFF_STRATEGY")
	cfg.Retry.InitialDelay = getEnvAsInt("RETRY_INITIAL_DELAY_SECONDS", 5)
	cfg.Retry.MaxDelay = getEnvAsInt("RETRY_MAX_DELAY_SECONDS", 300)

	// TaskQueue Config
	cfg.TaskQueue.ConcurrentTaskRun = getEnvAsInt("TASK_CONCURRENT_RUN_COUNT", 50)
	cfg.TaskQueue.RetryCount = getEnvAsInt("TASK_QUEUE_RETRY_COUNT", 2)
	cfg.TaskQueue.Delay = getEnvAsInt("TASK_PROCESS_DELAY_SECONDS", 5)
	cfg.TaskQueue.Expiration = getEnvAsInt("TASK_QUEUE_EXPIRATION", 0)
	cfg.TaskQueue.Retention = getEnvAsInt("TASK_QUEUE_RETENTION", 86400)

	// API Endpoints URLS
	cfg.ApiURL.RedisURL = getEnv("REDIS_ADDR")
	cfg.ApiURL.ReportDigitizationURL = getEnv("GEMINI_API_URL")
	cfg.ApiURL.NotifyServerURL = getEnv("NOTIFY_SERVER_URL")
	cfg.ApiURL.ShareReportBaseURL = getEnv("SHARE_REPORT_BASE_URL")
	cfg.ApiURL.ShortBaseURL = getEnv("SHORT_URL_BASE")
	cfg.ApiURL.AskChatBotURL = getEnv("ASK_API")
	cfg.ApiURL.ReportSummaryURL = getEnv("REPORT_SUMMARY_API")
	cfg.ApiURL.SummaryHistoryURL = getEnv("SUMMARIZE_HISTORY_API")
	cfg.ApiURL.PrescriptionURL = getEnv("PRESCRIPTION_API")
	cfg.ApiURL.PrescriptionDigitizationURL = getEnv("DIGITIZE_PRESCRIPTION_API")
	cfg.ApiURL.PharmacokineticsURL = getEnv("PHARMACOKINETICS_API")
	cfg.ApiURL.DigilockerRedirectURL = getEnv("DIGILOCKER_REDIRECT_URI")
	cfg.ApiURL.GoogleRedirectURL = getEnv("GOOGLE_REDIRECT_URI")

	return cfg
}
