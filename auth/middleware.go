package auth

import (
	"biostat/constant"
	"biostat/models"
	"biostat/utils"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
)

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

	// fmt.Println("Response Body:", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResponse TokenExchangeResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &tokenResponse, nil
}
