package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/Nerzal/gocloak/v13"
)

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
	fmt.Println("KeycloakAuthURL ", KeycloakAuthURL)
	fmt.Println("KeycloakRealm ", KeycloakRealm)
	fmt.Println("KeycloakClientID ", KeycloakClientID)
	fmt.Println("KeycloakClientSecret ", KeycloakClientSecret)
	fmt.Println("KeycloakAdminUser ", KeycloakAdminUser)
	fmt.Println("KeycloakAdminPassword ", KeycloakAdminPassword)

	Client = gocloak.NewClient(KeycloakAuthURL)
}
