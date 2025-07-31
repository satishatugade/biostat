package main

import (
	"biostat/config"
	"biostat/database"
	"biostat/router"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()
	env := os.Getenv("APP_ENV")
	var envFile string
	switch env {
	case "dev":
		envFile = ".env.dev"
	case "uat":
		envFile = ".env.uat"
	case "prod":
		envFile = ".env.prod"
	default:
		envFile = ".env.dev"
	}
	err := godotenv.Overload(envFile)
	if err != nil {
		log.Fatalf("Error loading %s file: %v", envFile, err)
	}
	config.PropConfig = config.LoadConfigFromEnv()
	config.SetupLogger()
	defer config.Log.Sync()
	config.Log.Info("Biostack Application Started.....")
	config.Log.Info("ENV profile active", zap.String("env", env))
	config.InitKeycloak()
	database.InitDB()
	config.InitRedisAndAsynq()
	router.Routing(env)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	database.GracefulShutdown()
	log.Println("Server gracefully stopped.")
}
