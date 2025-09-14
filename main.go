package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"ilex-backend/config"
	"ilex-backend/db"
	"ilex-backend/routes"
)

// @title ILEX Backend API
// @version 1.0
// @description Backend API for ILEX delivery platform with SurrealDB
// @termsOfService http://swagger.io/terms/

// @contact.name ILEX Support
// @contact.email support@ilex.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Starting ILEX Backend Server in %s mode", cfg.Environment)

	// Initialize database connection
	if err := db.InitDB(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.CloseDB(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Setup routes
	router := routes.SetupRoutes(cfg)

	// Setup graceful shutdown
	setupGracefulShutdown()

	// Start server
	log.Printf("Server starting on %s:%s", cfg.ServerHost, cfg.ServerPort)
	log.Printf("Health check available at http://%s:%s/health", cfg.ServerHost, cfg.ServerPort)
	log.Printf("API documentation available at http://%s:%s/swagger/index.html", cfg.ServerHost, cfg.ServerPort)

	if err := routes.StartServer(router, cfg); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupGracefulShutdown sets up graceful shutdown handling
func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		
		// Close database connection
		if err := db.CloseDB(); err != nil {
			log.Printf("Error closing database during shutdown: %v", err)
		}

		log.Println("Server shutdown complete")
		os.Exit(0)
	}()
}