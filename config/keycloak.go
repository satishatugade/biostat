package config

import (
	"log"
	"os"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/Nerzal/gocloak/v13"
)

var (
	AsynqClient *asynq.Client
	RedisClient *redis.Client
)

var (
	GoogleClientID     = os.Getenv("GOOGLE_CLIENT_ID")
	GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	RedirectURI        = os.Getenv("GOOGLE_REDIRECT_URI")
)

func InitRedisAndAsynq() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	AsynqClient = asynq.NewClient(asynq.RedisClientOpt{
		Addr: os.Getenv("REDIS_ADDR"),
	})
}

var (
	KeycloakAuthURL       string
	KeycloakRealm         string
	KeycloakClientID      string
	KeycloakClientSecret  string
	KeycloakAdminUser     string
	KeycloakAdminPassword string
)

var Client *gocloak.GoCloak

func InitKeycloak() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	KeycloakAuthURL = os.Getenv("KEYCLOAK_AUTH_URL")
	KeycloakRealm = os.Getenv("KEYCLOAK_REALM")
	KeycloakClientID = os.Getenv("KEYCLOAK_CLIENT_ID")
	KeycloakClientSecret = os.Getenv("KEYCLOAK_CLIENT_SECRET")
	KeycloakAdminUser = os.Getenv("KEYCLOAK_ADMIN_USER")
	KeycloakAdminPassword = os.Getenv("KEYCLOAK_ADMIN_PASSWORD")

	Client = gocloak.NewClient(KeycloakAuthURL)
}
