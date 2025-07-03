package auth

import (
	"biostat/constant"
	"biostat/models"
	"biostat/repository"
	"biostat/service"
	"biostat/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
)

type AuthService interface {
	SendResetPasswordLink(email string) error
	SendOTP(email string) error
	// VerifyOTPAndLogin(email string, otp string) (map[string]interface{}, error)
	CreateUserInKeycloak(user models.SystemUser_) (string, bool, error)
}

type AuthServiceImpl struct {
	userRepo     repository.UserRepository
	userService  service.UserService
	emailService service.EmailService
}

func NewAuthService(repo repository.UserRepository, userService service.UserService, emailService service.EmailService) AuthService {
	return &AuthServiceImpl{
		userRepo:     repo,
		userService:  userService,
		emailService: emailService,
	}
}

type TokenData struct {
	AuthUserID string
	ExpiresAt  time.Time
}

var (
	tokenStore = make(map[string]TokenData)
	tokenMutex = sync.RWMutex{}
)

func StoreToken(token, AuthuserID string, duration time.Duration) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()
	tokenStore[token] = TokenData{
		AuthUserID: AuthuserID,
		ExpiresAt:  time.Now().Add(duration),
	}
}

func GetTokenData(token string) (TokenData, bool) {
	tokenMutex.RLock()
	defer tokenMutex.RUnlock()
	data, found := tokenStore[token]
	if !found || time.Now().After(data.ExpiresAt) {
		return TokenData{}, false
	}
	return data, true
}

func DeleteToken(token string) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()
	delete(tokenStore, token)
}

func AuthToken(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Authorization header missing", nil, nil)
			c.Abort()
			return
		}
		tokenStr := authHeader[len("Bearer "):]
		client := utils.Client
		ctx := context.Background()
		introspection, err := client.RetrospectToken(ctx, tokenStr, utils.KeycloakClientID, utils.KeycloakClientSecret, utils.KeycloakRealm)
		if err != nil {
			log.Println("Error introspecting token:", err)
			models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Invalid token", nil, err)
			c.Abort()
			return
		}
		if !*introspection.Active {
			log.Println("introspection.Active ", introspection)
			models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Token is expired!", nil, err)
			c.Abort()
			return
		}

		_, claims, err := client.DecodeAccessToken(ctx, tokenStr, utils.KeycloakRealm)
		if err != nil {
			log.Println("Error decoding access token:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		if claims == nil {
			models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Invalid token claims", nil, err)
			c.Abort()
			return
		}

		roles, ok := (*claims)["realm_access"].(map[string]interface{})["roles"].([]interface{})
		if !ok {
			models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Invalid token claims", nil, err)
			c.Abort()
			return
		}

		sub, ok := (*claims)["sub"]
		if !ok {
			models.ErrorResponse(c, constant.Failure, http.StatusUnauthorized, "Invalid token claims", nil, err)
			c.Abort()
			return
		}

		var userRoles []string

		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			for _, role := range roles {
				userRoles = append(userRoles, role.(string))
				if role == requiredRole {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			models.ErrorResponse(c, constant.Failure, http.StatusForbidden, "Access denied", nil, err)
			c.Abort()
			return
		}

		// Store the claims in the context
		c.Set("claims", claims)
		c.Set("sub", sub)
		c.Set("userRoles", userRoles)
		c.Next()
	}
}

func Authenticate(path string, protectedRoutes map[string][]string, handler gin.HandlerFunc) gin.HandlerFunc {
	for protectedPrefix, roles := range protectedRoutes {
		if strings.HasPrefix(path, protectedPrefix) {
			return gin.HandlerFunc(func(c *gin.Context) {
				AuthToken(roles...)(c)
				if c.IsAborted() {
					return
				}
				handler(c)
			})
		}
	}
	return handler
}

type TokenExchangeResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	TokenType        string `json:"token_type"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	Scope            string `json:"scope"`
}

var Client *gocloak.GoCloak

func ExchangeToken(subjectToken string) (*TokenExchangeResponse, error) {
	formData := url.Values{}
	// formData.Set("grant_type", "urn:ietf:params:oauth:grant-type:token-exchange")
	formData.Set("grant_type", "client_credentials")
	formData.Set("subject_token", subjectToken)
	formData.Set("requested_token_type", "urn:ietf:params:oauth:token-type:access_token")
	formData.Set("client_id", os.Getenv("KEYCLOAK_CLIENT_ID"))
	formData.Set("client_secret", os.Getenv("KEYCLOAK_CLIENT_SECRET"))

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		os.Getenv("KEYCLOAK_AUTH_URL"),
		os.Getenv("KEYCLOAK_REALM"),
	)

	// Make the POST request
	resp, err := http.PostForm(tokenURL, formData)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResponse TokenExchangeResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &tokenResponse, nil
}

func ResetPasswordInKeycloak(username, newPassword string) error {
	ctx := context.Background()
	token, err := utils.Client.LoginAdmin(ctx, utils.KeycloakAdminUser, utils.KeycloakAdminPassword, "master")
	if err != nil {
		log.Println("ResetPasswordInKeycloak Failed to login to admin ", err)
		return fmt.Errorf("admin login failed: %w", err)
	}
	users, err := utils.Client.GetUsers(ctx, token.AccessToken, utils.KeycloakRealm, gocloak.GetUsersParams{
		Username: gocloak.StringP(username),
	})
	if err != nil {
		return fmt.Errorf("failed to fetch user: %w", err)
	}
	if len(users) == 0 || users[0].ID == nil {
		return fmt.Errorf("user '%s' not found in Keycloak", username)
	}
	userID := *users[0].ID
	err = utils.Client.SetPassword(ctx, token.AccessToken, userID, utils.KeycloakRealm, newPassword, false)
	if err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}
	return nil
}

func (as *AuthServiceImpl) SendResetPasswordLink(email string) error {
	userInfo, err := as.userRepo.GetUserInfoByEmailId(email)
	if err != nil {
		return errors.New("user not found")
	}
	token := uuid.New().String()
	// tokenData := &models.TblUserToken{
	// 	UserId:    userInfo.UserId,
	// 	AuthToken: token,
	// 	Provider:  "Email",
	// 	ExpiresAt: time.Now().Add(15 * time.Minute),
	// }
	// _, err1 := as.userService.CreateTblUserToken(tokenData)
	// if err1 != nil {
	// 	return errors.New("reset password token not saved")
	// }
	StoreToken(token, userInfo.AuthUserId, 15*time.Minute)

	MailErr := as.emailService.SendResetPasswordMail(userInfo, token, email)
	if MailErr != nil {
		return errors.New("reset password email not sent")
	}
	return nil
}

// func GenerateOTP() string {
// 	rand.Seed(time.Now().UnixNano())
// 	return fmt.Sprintf("%04d", rand.Intn(10000))
// }

// var otpStore = make(map[string]string)

// func (a *AuthServiceImpl) SendOTP(email string) error {
// 	otp := GenerateOTP()
// 	otpStore[email] = otp

// 	subject := "Your OTP Code"
// 	body := fmt.Sprintf("<h2>Your OTP code is: <strong>%s</strong></h2><p>This OTP is valid for 5 minutes.</p>", otp)

// 	if err := a.emailService.SendEmail(email, subject, body); err != nil {
// 		log.Printf("Failed to send OTP email: %v", err)
// 		return fmt.Errorf("email sending failed")
// 	}

// 	fmt.Println("OTP sent to:", email)
// 	return nil
// }

func (a *AuthServiceImpl) SendOTP(username string) error {
	// Prepare form data for Keycloak token endpoint
	form := url.Values{}
	form.Set("client_id", os.Getenv("KEYCLOAK_CLIENT_ID"))
	form.Set("client_secret", os.Getenv("KEYCLOAK_CLIENT_SECRET")) // Omit if public client
	form.Set("grant_type", "password")
	form.Set("username", username)
	form.Set("password", "password") // Any value; login will fail but triggers OTP

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		os.Getenv("KEYCLOAK_AUTH_URL"),
		os.Getenv("KEYCLOAK_REALM"),
	)

	// Make the dummy login attempt to trigger OTP email
	resp, err := http.PostForm(tokenURL, form)
	if err != nil {
		return fmt.Errorf("failed to contact Keycloak: %w", err)
	}
	defer resp.Body.Close()
	// io.Copy(io.Discard, resp.Body) // Discard the response body
	io.ReadAll(resp.Body)
	fmt.Println("data ", resp.Body)
	// No need to check resp.StatusCode; we expect failure
	return nil // Always return nil; purpose is to trigger OTP email
}

type VerifyOTPRequest struct {
	Username string `json:"username"`
	OTP      string `json:"otp"`
}

func VerifyOTPAndLogin(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.OTP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and OTP are required"})
		return
	}

	form := url.Values{}
	form.Set("client_id", os.Getenv("KEYCLOAK_CLIENT_ID"))
	form.Set("client_secret", os.Getenv("KEYCLOAK_CLIENT_SECRET")) // Omit if public client
	form.Set("grant_type", "password")
	form.Set("username", req.Username)
	form.Set("totp", req.OTP) // Use 'totp' for OTP-based login

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		os.Getenv("KEYCLOAK_AUTH_URL"),
		os.Getenv("KEYCLOAK_REALM"),
	)

	resp, err := http.PostForm(tokenURL, form)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact Keycloak"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusUnauthorized, gin.H{"error": string(body)})
		return
	}

	var tokenResp map[string]interface{}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token response"})
		return
	}
	c.JSON(http.StatusOK, tokenResp)
}

func getAdminToken() (string, error) {
	client := gocloak.NewClient(os.Getenv("KEYCLOAK_AUTH_URL"))
	ctx := context.Background()
	token, err := client.LoginClient(ctx, os.Getenv("KEYCLOAK_CLIENT_ID"), os.Getenv("KEYCLOAK_CLIENT_SECRET"), os.Getenv("KEYCLOAK_REALM"))
	if err != nil {
		return "", fmt.Errorf("admin token error: %w", err)
	}
	return token.AccessToken, nil
}

func findUserByEmail(email, adminToken string) (string, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/users?email=%s",
		os.Getenv("KEYCLOAK_AUTH_URL"),
		os.Getenv("KEYCLOAK_REALM"),
		email)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("find user failed: %s", string(body))
	}

	var users []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&users)

	if len(users) == 0 {
		return "", fmt.Errorf("no user found with email %s", email)
	}

	return users[0]["id"].(string), nil
}

func impersonateUser(userID, adminToken string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/users/%s/impersonation",
		os.Getenv("KEYCLOAK_AUTH_URL"),
		os.Getenv("KEYCLOAK_REALM"),
		userID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("impersonation failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("impersonation error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode impersonation response")
	}

	return result, nil
}

func (a *AuthServiceImpl) CreateUserInKeycloak(user models.SystemUser_) (string, bool, error) {
	client := utils.Client
	ctx := context.Background()
	token, err := client.LoginClient(ctx, utils.KeycloakClientID, utils.KeycloakClientSecret, utils.KeycloakRealm)
	if err != nil {
		log.Printf("[ERROR] Failed to login to Keycloak: %v", err)
		return "", false, err
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
		RealmRoles: &[]string{(user.RoleName)},
	}

	role, roleErr := client.GetRealmRole(ctx, token.AccessToken, utils.KeycloakRealm, user.RoleName)
	if roleErr != nil {
		log.Printf("[ERROR] Role not found in Keycloak: %v", err)
		return "User role not found at keycloak server", false, roleErr
	}

	userID, err := client.CreateUser(ctx, token.AccessToken, utils.KeycloakRealm, newuser)
	if err != nil {
		var loginInfo *models.UserLoginInfo
		loginInfo, err = a.userService.GetUserInfoByIdentifier(user.Email)
		if err != nil {
			loginInfo, err = a.userService.GetUserInfoByIdentifier(user.MobileNo)
			if err != nil {
				log.Printf("[ERROR] User not found in DB with email: %s or mobile: %s", user.Email, user.MobileNo)
				return "", false, fmt.Errorf("user not found")
			}
		}
		existingUsers, fetchErr := client.GetUsers(ctx, token.AccessToken, utils.KeycloakRealm, gocloak.GetUsersParams{
			Username: gocloak.StringP(loginInfo.Username),
			Email:    gocloak.StringP(user.Email),
			Max:      gocloak.IntP(1),
		})
		if fetchErr != nil {
			log.Printf("[ERROR] Error while fetching existing users: %v", fetchErr)
			return "", false, fmt.Errorf("user creation failed and checking existing users also failed: %v", fetchErr)
		}
		if len(existingUsers) > 0 && existingUsers[0].ID != nil {
			userID = *existingUsers[0].ID
			log.Printf("[INFO] User already exists with ID: %s. Proceeding to assign role in keycloak...", userID)

			roleErr := client.AddRealmRoleToUser(ctx, token.AccessToken, utils.KeycloakRealm, userID, []gocloak.Role{*role})
			if roleErr != nil {
				log.Printf("[ERROR] Failed to assign role '%s' to existing user ID %s: %v", role.Name, userID, roleErr)
				return "", true, fmt.Errorf("unable to assign role to existing user: %v", roleErr)
			}
			log.Printf("[INFO] Role '%s' assigned to existing user ID: %s", role.Name, userID)
			return userID, true, nil
		}
		log.Printf("existingUsers id ", existingUsers)
		log.Printf("[ERROR] User creation failed and no existing user found (username: %s, email: %s)", user.Username, user.Email)
		return "", false, fmt.Errorf("user creation failed: %v", err)
	}

	adderr := client.AddRealmRoleToUser(ctx, token.AccessToken, utils.KeycloakRealm, userID, []gocloak.Role{*role})
	if adderr != nil {
		return "Unable to add role to user", false, adderr
	}
	return userID, false, nil
}
