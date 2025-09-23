// @title ILEX Backend API
// @version 1.0
// @description API complète pour le système de livraison ILEX
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host 127.0.0.1:3000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"log"
	"os"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/database"
	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/handlers"
	"github.com/ambroise1219/livraison_go/routes"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/ambroise1219/livraison_go/docs" // Import des docs générés
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

	// Initialiser PostgreSQL via Prisma
	log.Println("🔗 Connexion à PostgreSQL...")
	if err := database.InitPrisma(); err != nil {
		log.Fatalf("❌ Erreur lors de l'initialisation de PostgreSQL: %v", err)
	}
	defer database.ClosePrisma()
	log.Println("✅ Connexion à PostgreSQL établie avec succès")

	// Initialiser le client Prisma dans le package db
	if err := db.InitializePrisma(); err != nil {
		log.Fatalf("❌ Erreur lors de l'initialisation du client Prisma: %v", err)
	}
	defer db.ClosePrisma()

	// Initialiser les handlers
	log.Println("🔧 Initialisation des handlers...")
	handlers.InitHandlers()
	log.Println("✅ Handlers initialisés avec succès")

	// Vérifier l'uploader Cloudinary
	log.Println("🔍 Vérification de l'uploader Cloudinary...")
	log.Println("🔍 Test d'initialisation Cloudinary...")

	// Configurer les routes
	log.Println("🚀 Configuration des routes...")
	router := routes.SetupRoutes()
	
	// Ajouter la route Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Afficher les informations de démarrage
	log.Printf("🌟 Serveur ILEX Backend démarré:")
	log.Printf("   🏠 Environnement: %s", cfg.Environment)
	log.Printf("   🌐 Adresse: %s:%s", cfg.ServerHost, cfg.ServerPort)
	log.Printf("   🐘 PostgreSQL: livraison_db")
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
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgresql://neondb_owner:npg_9pkHjaIsTc6Z@ep-purple-king-agho52sv-pooler.c-2.eu-central-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require")
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
