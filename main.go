package main

import (
	"biostat/database"
	"biostat/router"
	"biostat/utils"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatalf("Error loading .env file: %v", err)
	// }
	// env := os.Getenv("APP_ENV")
	// var envFile string

	// switch env {
	// case "dev":
	// 	envFile = ".env.dev"
	// case "uat":
	// 	envFile = ".env.uat"
	// case "prod":
	// 	envFile = ".env.prod"
	// default:
	// 	log.Println("No environment set or environment is not supported, using default .env.dev")
	// 	envFile = ".env.dev"
	// }
	// fmt.Println("ENV profile active ", envFile)
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	logDir := os.Getenv("LOG_DIRECTORY")
	logFile := os.Getenv("LOG_FILE")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Error creating logs directory: %v", err)
	}
	logPath := logDir + logFile
	file, LogFileError := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if LogFileError != nil {
		log.Fatal("Error opening log file: ", LogFileError)
	}
	defer file.Close()
	log.SetFlags(log.Ldate | log.Ltime)
	log.SetOutput(file)
	log.Println("Biostat Application Started.....")
	utils.InitKeycloak()
	database.InitDB()

	router.Routing()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit // Wait for termination signal

	log.Println("Shutting down server...")
	database.GracefulShutdown() // Close database connection
	log.Println("Server gracefully stopped.")
}
