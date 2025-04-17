package auth

import (
	"biostat/constant"
	"biostat/models"
	"biostat/utils"
	"context"
	"log"
	"net/http"
	"strings"

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
				log.Println(" keycloak roles ", roles)
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
