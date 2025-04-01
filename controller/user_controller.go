package controller

import (
	"biostat/database"
	"biostat/models"
	"biostat/utils"
	"context"

	"fmt"
	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
}

func NewUserController() *UserController {
	return &UserController{}
}

func (uc *UserController) RegisterUser(c *gin.Context) {
	// var user models.User
	var user models.Patient
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	//Add User in Keycloak
	keyCloakID, err := createUserInKeycloak(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user.AuthUserId = keyCloakID
	user.Password = string(hashedPassword)
	//store user in DB
	if err := database.DB.Create(&user).Error; err != nil {
		// if err := connection.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func createUserInKeycloak(user models.Patient) (string, error) {
	client := utils.Client
	fmt.Println("client ", client)
	ctx := context.Background()
	token, err := client.LoginClient(ctx, utils.KeycloakClientID, utils.KeycloakClientSecret, utils.KeycloakRealm)
	if err != nil {
		return "", err
	}
	newuser := gocloak.User{
		Username:      gocloak.StringP(user.Username),
		Email:         gocloak.StringP(user.Email),
		FirstName:     gocloak.StringP(user.FirstName),
		LastName:      gocloak.StringP(user.LastName),
		Enabled:       gocloak.BoolP(true),
		EmailVerified: gocloak.BoolP(true),
		Credentials: &[]gocloak.CredentialRepresentation{
			{
				Type:      gocloak.StringP("password"),
				Value:     gocloak.StringP(user.Password),
				Temporary: gocloak.BoolP(false),
			},
		},
		RealmRoles: &[]string{(user.Role)},
	}

	role, roleErr := client.GetRealmRole(ctx, token.AccessToken, utils.KeycloakRealm, user.Role)
	if roleErr != nil {
		return "", roleErr
	}

	userID, err := client.CreateUser(ctx, token.AccessToken, utils.KeycloakRealm, newuser)
	if err != nil {
		return "", err
	}

	adderr := client.AddRealmRoleToUser(ctx, token.AccessToken, utils.KeycloakRealm, userID, []gocloak.Role{*role})
	if adderr != nil {
		return "", adderr
	}
	return userID, nil
}

func (uc *UserController) LoginUser(c *gin.Context) {
	var user models.Patient
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	client := utils.Client
	ctx := context.Background()

	token, err := client.Login(ctx, utils.KeycloakClientID, utils.KeycloakClientSecret, utils.KeycloakRealm, input.Username, input.Password)
	if err != nil {
		fmt.Println("Error logging in to Keycloak:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err})
		return
	}

	_, claims, err := client.DecodeAccessToken(ctx, token.AccessToken, utils.KeycloakRealm)
	if err != nil {
		fmt.Println("Error decoding access token:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		c.Abort()
		return
	}

	sub, ok := (*claims)["sub"]
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		c.Abort()
		return
	}

	err1 := database.DB.Where("auth_user_id = ?", sub).First(&user).Error
	if err1 != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": err1.Error()})
		return
	}

	type UserResponse struct {
		ID         int    `json:"id" gorm:"primaryKey"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Role       string `json:"role"`
		Username   string `json:"username"`
		Email      string `json:"email"`
		AuthUserId string `json:"auth_user_id"`
	}

	useResponse := UserResponse{ID: int(user.PatientId), FirstName: user.FirstName, LastName: user.LastName, Email: user.Email, Username: user.Username, Role: user.Role, AuthUserId: user.AuthUserId}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"userData":      useResponse})
}

func (uc *UserController) LogoutUser(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := utils.Client
	ctx := context.Background()

	err := client.Logout(ctx, utils.KeycloakClientID, utils.KeycloakClientSecret, utils.KeycloakRealm, input.RefreshToken)
	if err != nil {
		fmt.Println("Error logging out from Keycloak:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User logged out successfully"})
}

func GetUserInfoById(c *gin.Context) {
	var user models.Patient
	userID := c.Param("id")

	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func GetAllUsersInfo(c *gin.Context) {
	var users []models.Patient
	if err := database.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}
