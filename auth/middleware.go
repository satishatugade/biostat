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
}

type AuthServiceImpl struct {
	userRepo         repository.UserRepository
	userTokenService service.UserService
	emailService     *service.EmailService
}

func NewAuthService(repo repository.UserRepository, userTokenService service.UserService, emailService *service.EmailService) AuthService {
	return &AuthServiceImpl{
		userRepo:         repo,
		userTokenService: userTokenService,
		emailService:     emailService,
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
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
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
	// _, err1 := as.userTokenService.CreateTblUserToken(tokenData)
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
