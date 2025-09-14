package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware middleware de logging personnalisé
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Traiter la requête
		c.Next()

		// Calculer la latence
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		// Construire l'URL complète
		if raw != "" {
			path = path + "?" + raw
		}

		// Log avec couleurs selon le status code
		var statusColor, methodColor, resetColor string
		if gin.IsDebugging() {
			statusColor = getStatusColor(statusCode)
			methodColor = getMethodColor(method)
			resetColor = "\033[0m"
		}

		log.Printf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %s\n%s",
			start.Format("2006/01/02 - 15:04:05"),
			statusColor, statusCode, resetColor,
			latency,
			clientIP,
			methodColor, method, resetColor,
			path,
			c.Errors.String(),
		)
	}
}

// CORSMiddleware middleware pour gérer CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Liste des origines autorisées (à adapter selon vos besoins)
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:3001", 
			"http://127.0.0.1:3000",
			"https://ilex-app.com", // Votre domaine de production
		}

		// Vérifier si l'origine est autorisée
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 heures

		// Gérer les requêtes OPTIONS (preflight)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware ajoute des headers de sécurité
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Protection contre le clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// Protection XSS
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Empêcher le sniffing MIME
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Politique de référent stricte
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Content Security Policy de base
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self';")
		
		c.Next()
	}
}

// RateLimitMiddleware middleware simple de limitation de débit
func RateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	// Map pour stocker les compteurs par IP
	requestCounts := make(map[string][]time.Time)
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Nettoyer les anciennes requêtes
		if timestamps, exists := requestCounts[clientIP]; exists {
			var validTimestamps []time.Time
			for _, timestamp := range timestamps {
				if now.Sub(timestamp) <= window {
					validTimestamps = append(validTimestamps, timestamp)
				}
			}
			requestCounts[clientIP] = validTimestamps
		}

		// Vérifier si la limite est atteinte
		if len(requestCounts[clientIP]) >= maxRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Trop de requêtes",
				"message": fmt.Sprintf("Limite de %d requêtes par %v dépassée", maxRequests, window),
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		// Ajouter la requête actuelle
		requestCounts[clientIP] = append(requestCounts[clientIP], now)
		
		c.Next()
	}
}

// RecoveryMiddleware middleware de récupération personnalisé
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("Panic recovered: %v", recovered)
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erreur interne du serveur",
			"message": "Une erreur inattendue s'est produite",
		})
	})
}

// Fonctions utilitaires pour les couleurs de log
func getStatusColor(code int) string {
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return "\033[97;42m" // Blanc sur fond vert
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return "\033[90;47m" // Gris sur fond blanc
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return "\033[90;43m" // Gris sur fond jaune
	default:
		return "\033[97;41m" // Blanc sur fond rouge
	}
}

func getMethodColor(method string) string {
	switch method {
	case http.MethodGet:
		return "\033[97;44m" // Blanc sur fond bleu
	case http.MethodPost:
		return "\033[97;42m" // Blanc sur fond vert
	case http.MethodPut:
		return "\033[97;43m" // Blanc sur fond jaune
	case http.MethodDelete:
		return "\033[97;41m" // Blanc sur fond rouge
	case http.MethodPatch:
		return "\033[97;46m" // Blanc sur fond cyan
	case http.MethodHead:
		return "\033[97;45m" // Blanc sur fond magenta
	case http.MethodOptions:
		return "\033[90;47m" // Gris sur fond blanc
	default:
		return "\033[0m" // Reset
	}
}
