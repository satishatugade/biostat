package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log  *zap.Logger
	once sync.Once
)

func SetupLogger() {
	once.Do(func() {
		err := godotenv.Load()
		if err != nil {
			log.Println("No .env file found. Using default values.")
		}

		logDir := os.Getenv("LOG_DIRECTORY")
		if logDir == "" {
			logDir = "logs"
		}

		logFile := os.Getenv("LOG_FILE")
		if logFile == "" {
			logFile = "biostat.log"
		}

		logLevelStr := os.Getenv("LOG_LEVEL")
		if logLevelStr == "" {
			logLevelStr = "info"
		}

		err = os.MkdirAll(logDir, os.ModePerm)
		if err != nil {
			log.Panicf("Failed to create log directory: %v", err)
		}

		logPath := filepath.Join(logDir, logFile)
		file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Panicf("Failed to open log file: %v", err)
		}
		log.SetFlags(log.Ldate | log.Ltime)
		log.SetOutput(file)
		// Parse the log level
		zapLevel := getZapLogLevel(logLevelStr)

		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "timestamp"
		encoderCfg.EncodeTime = customTimeEncoder
		encoderCfg.LevelKey = "level"
		encoderCfg.MessageKey = "message"
		encoderCfg.CallerKey = "caller"

		encoder := zapcore.NewConsoleEncoder(encoderCfg)

		fileWriter := zapcore.AddSync(file)
		consoleWriter := zapcore.AddSync(os.Stdout)

		core := zapcore.NewTee(
			zapcore.NewCore(encoder, fileWriter, zapLevel),
			zapcore.NewCore(encoder, consoleWriter, zapLevel),
		)

		Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	})
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("02/01/2006 15:04:05"))
}

func getZapLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
