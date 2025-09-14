package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ilex-backend/config"
	"ilex-backend/handlers"
	"ilex-backend/services"
)

// SetupRoutes configures all application routes
func SetupRoutes(cfg *config.Config) *gin.Engine {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin engine
	r := gin.New()

	// Apply global middleware
	r.Use(LoggerMiddleware())
	r.Use(RecoveryMiddleware())
	r.Use(CORSMiddleware())
	r.Use(RequestIDMiddleware())
	r.Use(RateLimitMiddleware())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "ilex-backend",
			"version": "1.0.0",
		})
	})

	// Initialize services
	authService := services.NewAuthService(cfg)
	promoService := services.NewPromoService(cfg)
	deliveryService := services.NewDeliveryService(cfg, promoService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	deliveryHandler := handlers.NewDeliveryHandler(deliveryService)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/otp/send", authHandler.SendOTP)
			auth.POST("/login", authHandler.VerifyOTP)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)

			// Protected auth routes
			authProtected := auth.Group("")
			authProtected.Use(AuthMiddleware(authService))
			{
				authProtected.GET("/profile", authHandler.GetProfile)
			}
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(AuthMiddleware(authService))
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/profile", RequireAnyUser(), authHandler.GetProfile)
				// TODO: Add more user endpoints
				// users.PUT("/profile", RequireAnyUser(), userHandler.UpdateProfile)
				// users.POST("/upload-document", RequireDriver(), userHandler.UploadDocument)
			}

			// Delivery routes
			deliveries := protected.Group("/deliveries")
			{
				deliveries.POST("", RequireClient(), deliveryHandler.CreateDelivery)
				deliveries.GET("", RequireAnyUser(), deliveryHandler.GetDeliveries)
				deliveries.GET("/:deliveryId", RequireAnyUser(), deliveryHandler.GetDelivery)
				deliveries.PUT("/:deliveryId/status", RequireAnyUser(), deliveryHandler.UpdateDeliveryStatus)
				deliveries.GET("/calculate-price", RequireClient(), deliveryHandler.CalculatePrice)
				
				// Admin only routes
				deliveries.POST("/assign", RequireAdmin(), deliveryHandler.AssignDelivery)
			}

			// Driver routes
			drivers := protected.Group("/drivers")
			drivers.Use(RequireDriver())
			{
				// TODO: Add driver-specific endpoints
				// drivers.PUT("/location", driverHandler.UpdateLocation)
				// drivers.GET("/deliveries/available", driverHandler.GetAvailableDeliveries)
				// drivers.POST("/deliveries/:id/accept", driverHandler.AcceptDelivery)
			}

			// Vehicle routes
			vehicles := protected.Group("/vehicles")
			vehicles.Use(RequireDriver())
			{
				// TODO: Add vehicle management endpoints
				// vehicles.POST("", vehicleHandler.CreateVehicle)
				// vehicles.GET("", vehicleHandler.GetVehicles)
				// vehicles.PUT("/:id", vehicleHandler.UpdateVehicle)
			}

			// Promo routes
			promos := protected.Group("/promos")
			{
				// Client routes
				promos.POST("/validate", RequireClient(), func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Promo validation endpoint - TODO"})
				})
				promos.POST("/apply", RequireClient(), func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Apply promo endpoint - TODO"})
				})

				// Admin routes
				adminPromos := promos.Group("")
				adminPromos.Use(RequireAdmin())
				{
					adminPromos.POST("", func(c *gin.Context) {
						c.JSON(http.StatusOK, gin.H{"message": "Create promo endpoint - TODO"})
					})
					adminPromos.GET("", func(c *gin.Context) {
						c.JSON(http.StatusOK, gin.H{"message": "List promos endpoint - TODO"})
					})
					adminPromos.PUT("/:id", func(c *gin.Context) {
						c.JSON(http.StatusOK, gin.H{"message": "Update promo endpoint - TODO"})
					})
				}
			}

			// Referral routes
			referrals := protected.Group("/referrals")
			{
				referrals.POST("", RequireClient(), func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Create referral endpoint - TODO"})
				})
				referrals.GET("", RequireClient(), func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Get referrals endpoint - TODO"})
				})
				referrals.POST("/:id/claim", RequireClient(), func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Claim referral reward endpoint - TODO"})
				})
			}

			// Payment routes
			payments := protected.Group("/payments")
			{
				payments.GET("/methods", RequireAnyUser(), func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Payment methods endpoint - TODO"})
				})
				payments.POST("/process", RequireClient(), func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Process payment endpoint - TODO"})
				})
			}

			// Wallet routes
			wallet := protected.Group("/wallet")
			{
				wallet.GET("", RequireAnyUser(), func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Get wallet endpoint - TODO"})
				})
				wallet.POST("/recharge", RequireClient(), func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Recharge wallet endpoint - TODO"})
				})
				wallet.GET("/transactions", RequireAnyUser(), func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Wallet transactions endpoint - TODO"})
				})
			}

			// Notification routes
			notifications := protected.Group("/notifications")
			notifications.Use(RequireAnyUser())
			{
				notifications.GET("", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Get notifications endpoint - TODO"})
				})
				notifications.PUT("/:id/read", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Mark notification as read endpoint - TODO"})
				})
			}

			// Admin routes
			admin := protected.Group("/admin")
			admin.Use(RequireAdmin())
			{
				// Dashboard
				admin.GET("/dashboard", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Admin dashboard endpoint - TODO"})
				})

				// User management
				admin.GET("/users", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Admin users list endpoint - TODO"})
				})
				admin.PUT("/users/:id/status", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Update user status endpoint - TODO"})
				})

				// Delivery management
				admin.GET("/deliveries/stats", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Delivery statistics endpoint - TODO"})
				})

				// Financial reports
				admin.GET("/reports/financial", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Financial reports endpoint - TODO"})
				})

				// Platform configuration
				admin.GET("/config", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Platform config endpoint - TODO"})
				})
				admin.PUT("/config", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Update platform config endpoint - TODO"})
				})
			}
		}
	}

	// WebSocket routes for real-time features
	r.GET("/ws/delivery/:deliveryId", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "WebSocket delivery tracking endpoint - TODO"})
	})

	r.GET("/ws/driver/location", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "WebSocket driver location endpoint - TODO"})
	})

	// File upload routes
	files := r.Group("/files")
	{
		files.POST("/upload", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "File upload endpoint - TODO"})
		})
		files.GET("/:fileId", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "File serving endpoint - TODO"})
		})
	}

	// 404 handler
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "The requested resource was not found",
			"path":    c.Request.URL.Path,
		})
	})

	return r
}

// StartServer starts the HTTP server
func StartServer(r *gin.Engine, cfg *config.Config) error {
	address := cfg.ServerHost + ":" + cfg.ServerPort
	return r.Run(address)
}