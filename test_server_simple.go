package main

import (
	"log"
	"time"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/database"
	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("🚀 Démarrage serveur de test simplifié")
	
	// Charger la configuration
	cfg := config.GetConfig()
	gin.SetMode(gin.DebugMode)
	
	// Initialiser la base de données
	if err := database.InitPrisma(); err != nil {
		log.Fatalf("❌ Erreur DB: %v", err)
	}
	defer database.ClosePrisma()
	
	if err := db.InitializePrisma(); err != nil {
		log.Fatalf("❌ Erreur db client: %v", err)
	}
	defer db.ClosePrisma()
	
	// Initialiser les handlers
	handlers.InitHandlers()
	
	// Créer le routeur simplifié
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "ilex-backend-test",
		})
	})
	
	// API v1 simplifié
	v1 := router.Group("/api/v1")
	
	// Routes publiques
	auth := v1.Group("/auth")
	{
		auth.POST("/otp/send", handlers.SendOTP)
		auth.POST("/otp/verify", handlers.VerifyOTP)
	}
	
	delivery := v1.Group("/delivery")
	{
		delivery.POST("/price/calculate", handlers.CalculateDeliveryPrice)
	}
	
	log.Printf("🌟 Serveur test démarré sur %s:%s", cfg.ServerHost, cfg.ServerPort)
	log.Printf("📈 Health Check: http://%s:%s/health", cfg.ServerHost, cfg.ServerPort)
	
	address := cfg.ServerHost + ":" + cfg.ServerPort
	if err := router.Run(address); err != nil {
		log.Fatalf("❌ Erreur serveur: %v", err)
	}
}