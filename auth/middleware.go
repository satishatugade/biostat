package auth

import (
	"biostat/utils"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthToken(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}
		tokenStr := authHeader[len("Bearer "):]

		client := utils.Client
		ctx := context.Background()
		fmt.Println("tokenStr ", tokenStr)
		_, claims, err := client.DecodeAccessToken(ctx, tokenStr, utils.KeycloakRealm)
		if err != nil {
			fmt.Println("Error decoding access token:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		introspection, err := client.RetrospectToken(ctx, tokenStr, utils.KeycloakClientID, utils.KeycloakClientSecret, utils.KeycloakRealm)
		if err != nil {
			fmt.Println("Error introspecting token:", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		fmt.Println("introspection.Active ", introspection.Active)
		if !*introspection.Active {
			fmt.Println("introspection.Active ", introspection)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is expired!"})
			c.Abort()
			return
		}

		// _, claims, err := client.DecodeAccessToken(ctx, tokenStr, utils.KeycloakRealm)
		// if err != nil {
		// 	fmt.Println("Error decoding access token:", err)
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		// 	c.Abort()
		// 	return
		// }
		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		roles, ok := (*claims)["realm_access"].(map[string]interface{})["roles"].([]interface{})
		if !ok {
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

		var userRoles []string

		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			for _, role := range roles {
				fmt.Println("roles ", roles)
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
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to perform this action."})
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

func ApplyMiddleware(path string, protectedRoutes map[string][]string, handler gin.HandlerFunc) gin.HandlerFunc {
	fmt.Println("Full path:", path)

	// Check if path starts with any protected route prefix
	for protectedPrefix, roles := range protectedRoutes {
		if strings.HasPrefix(path, protectedPrefix) {
			fmt.Println("Protected route matched:", protectedPrefix)
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
