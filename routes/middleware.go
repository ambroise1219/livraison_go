package routes

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ambroise1219/livraison_go/models"
	// "github.com/ambroise1219/livraison_go/services/auth"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Token is required",
			})
			c.Abort()
			return
		}

		// TODO: Implémenter la validation JWT
		// Pour l'instant, on accepte tous les tokens

		// Check if token is expired (handled automatically by JWT library in token.Valid)
		// Additional check if needed:
		// if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		//     c.JSON(http.StatusUnauthorized, gin.H{
		//         "error":   "Unauthorized",
		//         "message": "Token has expired",
		//     })
		//     c.Abort()
		//     return
		// }

		// Set user context (temporaire)
		c.Set("userID", "temp_user")
		c.Set("userPhone", "temp_phone")
		c.Set("userRole", "CLIENT")

		// Optional: Set user object (would require database call)
		// user := &models.User{
		//     ID:    claims.UserID,
		//     Phone: claims.Phone,
		//     Role:  claims.Role,
		// }
		// c.Set("user", user)

		c.Next()
	})
}

// RequireRole middleware checks if user has required role
func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userRoleInterface, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "User role not found in context",
			})
			c.Abort()
			return
		}

		userRole, ok := userRoleInterface.(models.UserRole)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Invalid user role format",
			})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRequiredRole := false
		for _, role := range roles {
			if userRole == role {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequireClient middleware - shorthand for client role
func RequireClient() gin.HandlerFunc {
	return RequireRole(models.UserRoleClient)
}

// RequireDriver middleware - shorthand for driver role
func RequireDriver() gin.HandlerFunc {
	return RequireRole(models.UserRoleLivreur)
}

// RequireAdmin middleware - shorthand for admin roles
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.UserRoleAdmin, models.UserRoleGestionnaire)
}

// RequireAnyUser middleware - allows any authenticated user
func RequireAnyUser() gin.HandlerFunc {
	return RequireRole(
		models.UserRoleClient,
		models.UserRoleLivreur,
		models.UserRoleAdmin,
		models.UserRoleGestionnaire,
		models.UserRoleMarketing,
	)
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// LoggerMiddleware logs requests
func LoggerMiddleware() gin.HandlerFunc {
	return gin.Logger()
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.Recovery()
}

// RateLimitMiddleware (placeholder for rate limiting)
func RateLimitMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// TODO: Implement rate limiting logic
		// This could use Redis or in-memory storage
		c.Next()
	})
}

// RequestIDMiddleware adds request ID to context
func RequestIDMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate a simple request ID
			requestID = generateRequestID()
		}

		c.Set("requestID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	})
}

// Helper function to generate request ID
func generateRequestID() string {
	// Simple implementation - in production, use UUID or similar
	return "req-" + strings.Replace(strings.Replace(gin.Mode(), "debug", "dev", 1), "release", "prod", 1) + "-" + "123456"
}

// ValidationMiddleware can be used to add custom validation
func ValidationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Add custom validation logic here if needed
		c.Next()
	})
}
