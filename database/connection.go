package database

import (
	"biostat/config"
	"biostat/models"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var sqlDB *sql.DB

func GetDBConn() *gorm.DB {
	if DB == nil {
		InitDB()
	}
	return DB
}

func InitDB() *gorm.DB {
	dbHost := config.PropConfig.Database.Host
	dbPort := config.PropConfig.Database.Port
	dbName := config.PropConfig.Database.DBName
	dbUser := config.PropConfig.Database.UserName
	dbPassword := config.PropConfig.Database.Password
	dbSSLMode := config.PropConfig.Database.SSLMode

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost, dbUser, dbPassword, dbName, dbPort, dbSSLMode)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("db.16 Error connecting", err)
	}

	sqlDB, err = database.DB()
	if err != nil {
		log.Fatal("db.20 Failed to get database object from Gorm DB", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatal("db.23 Failed to ping database", err)
	}

	log.Println("db.26 Database connection established successfully")
	database.AutoMigrate(&models.SubscriptionMasterAudit{})
	DB = database
	return DB
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if err := sqlDB.Ping(); err != nil {
		http.Error(w, "Database connection unhealthy", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Database connection healthy"))
}

func GracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")
	sqlDB.SetConnMaxIdleTime(10 * time.Second)
	sqlDB.Close()
	log.Println("Database connection closed")
}
