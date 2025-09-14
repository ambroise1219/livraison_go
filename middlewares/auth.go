package middlewares

import (
	"net/http"
	"strings"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware vérifie la validité du token JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token d'autorisation requis",
			})
			c.Abort()
			return
		}

		// Vérifier le format "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Format de token invalide. Utilisez 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]
		cfg := config.GetConfig()

		// Parser le token
		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Vérifier la méthode de signature
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token invalide",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Vérifier la validité du token
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token expiré ou invalide",
			})
			c.Abort()
			return
		}

		// Extraire les claims
		claims, ok := token.Claims.(*models.JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Claims du token invalides",
			})
			c.Abort()
			return
		}

		// Stocker les informations de l'utilisateur dans le contexte
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("user_claims", claims)

		c.Next()
	}
}

// OptionalAuthMiddleware vérifie le token s'il est présent, mais n'est pas obligatoire
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Pas de token, continuer sans authentification
			c.Next()
			return
		}

		// Si un token est présent, le valider comme dans AuthMiddleware
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			c.Next()
			return
		}

		tokenString := tokenParts[1]
		cfg := config.GetConfig()

		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err == nil && token.Valid {
			if claims, ok := token.Claims.(*models.JWTClaims); ok {
				c.Set("user_id", claims.UserID)
				c.Set("user_role", claims.Role)
				c.Set("user_claims", claims)
			}
		}

		c.Next()
	}
}

// GetCurrentUser récupère les informations de l'utilisateur depuis le contexte
func GetCurrentUser(c *gin.Context) (*models.JWTClaims, bool) {
	claims, exists := c.Get("user_claims")
	if !exists {
		return nil, false
	}
	
	userClaims, ok := claims.(*models.JWTClaims)
	return userClaims, ok
}

// GetCurrentUserID récupère l'ID de l'utilisateur depuis le contexte
func GetCurrentUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	
	id, ok := userID.(string)
	return id, ok
}

// GetCurrentUserRole récupère le rôle de l'utilisateur depuis le contexte
func GetCurrentUserRole(c *gin.Context) (models.UserRole, bool) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", false
	}
	
	userRole, ok := role.(models.UserRole)
	return userRole, ok
}
