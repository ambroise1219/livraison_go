package main

import (
	"log"
	"os"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Charger la configuration
	cfg := config.GetConfig()
	
	// Définir le mode Gin selon l'environnement
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.Environment == "development" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.TestMode)
	}

	// Initialiser la connexion à la base de données
	log.Println("🔗 Connexion à SurrealDB...")
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("❌ Erreur lors de l'initialisation de SurrealDB: %v", err)
	}
	log.Println("✅ Connexion à SurrealDB établie avec succès")

	// Configurer les routes
	log.Println("🚀 Configuration des routes...")
	router := routes.SetupRoutes()
	
	// Afficher les informations de démarrage
	log.Printf("🌟 Serveur ILEX Backend démarré:")
	log.Printf("   🏠 Environnement: %s", cfg.Environment)
	log.Printf("   🌐 Adresse: %s:%s", cfg.ServerHost, cfg.ServerPort)
	log.Printf("   📊 SurrealDB: %s", cfg.SurrealDBURL)
	log.Printf("   📈 Health Check: http://%s:%s/health", cfg.ServerHost, cfg.ServerPort)
	log.Printf("   📚 API Documentation: http://%s:%s/api/v1", cfg.ServerHost, cfg.ServerPort)
	
	// Démarrer le serveur
	address := cfg.ServerHost + ":" + cfg.ServerPort
	log.Printf("🚀 Serveur en écoute sur %s", address)
	
	if err := router.Run(address); err != nil {
		log.Fatalf("❌ Erreur lors du démarrage du serveur: %v", err)
	}
}

// init initialise les variables d'environnement par défaut si elles ne sont pas définies
func init() {
	// Définir des valeurs par défaut pour le développement local
	if os.Getenv("SERVER_HOST") == "" {
		os.Setenv("SERVER_HOST", "localhost")
	}
	if os.Getenv("SERVER_PORT") == "" {
		os.Setenv("SERVER_PORT", "8080")
	}
	if os.Getenv("ENVIRONMENT") == "" {
		os.Setenv("ENVIRONMENT", "development")
	}
	if os.Getenv("SURREALDB_URL") == "" {
		os.Setenv("SURREALDB_URL", "ws://localhost:8000")
	}
	if os.Getenv("SURREALDB_NAMESPACE") == "" {
		os.Setenv("SURREALDB_NAMESPACE", "ilex")
	}
	if os.Getenv("SURREALDB_DATABASE") == "" {
		os.Setenv("SURREALDB_DATABASE", "livraison")
	}
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "dev-jwt-secret-key-change-in-production")
	}
	if os.Getenv("JWT_EXPIRY_HOURS") == "" {
		os.Setenv("JWT_EXPIRY_HOURS", "24")
	}
	if os.Getenv("JWT_REFRESH_EXPIRY_DAYS") == "" {
		os.Setenv("JWT_REFRESH_EXPIRY_DAYS", "7")
	}
	if os.Getenv("OTP_EXPIRY_MINUTES") == "" {
		os.Setenv("OTP_EXPIRY_MINUTES", "5")
	}
	if os.Getenv("SMS_API_KEY") == "" {
		os.Setenv("SMS_API_KEY", "dev-sms-api-key")
	}
	if os.Getenv("SMS_SENDER") == "" {
		os.Setenv("SMS_SENDER", "ILEX")
	}
	if os.Getenv("EMAIL_HOST") == "" {
		os.Setenv("EMAIL_HOST", "smtp.gmail.com")
	}
	if os.Getenv("EMAIL_PORT") == "" {
		os.Setenv("EMAIL_PORT", "587")
	}
	if os.Getenv("EMAIL_USERNAME") == "" {
		os.Setenv("EMAIL_USERNAME", "dev@ilex.com")
	}
	if os.Getenv("EMAIL_PASSWORD") == "" {
		os.Setenv("EMAIL_PASSWORD", "dev-email-password")
	}
}
