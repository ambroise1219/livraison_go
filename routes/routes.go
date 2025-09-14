package routes

import (
	"time"

	"github.com/ambroise1219/livraison_go/handlers"
	"github.com/ambroise1219/livraison_go/middlewares"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configure toutes les routes de l'application
func SetupRoutes() *gin.Engine {
	// Créer le routeur Gin
	router := gin.New()

	// Middlewares globaux
	router.Use(middlewares.RecoveryMiddleware())
	router.Use(middlewares.LoggerMiddleware())
	router.Use(middlewares.CORSMiddleware())
	router.Use(middlewares.SecurityHeadersMiddleware())
	
	// Rate limiting global (100 requêtes par minute)
	router.Use(middlewares.RateLimitMiddleware(100, time.Minute))

	// Health check endpoint (pas d'authentification requise)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "ilex-backend",
			"version":   "1.0.0",
		})
	})

	// API v1
	v1 := router.Group("/api/v1")
	
	// Routes publiques (pas d'authentification requise)
	setupPublicRoutes(v1)
	
	// Routes protégées (authentification requise)
	setupProtectedRoutes(v1)

	return router
}

// setupPublicRoutes configure les routes publiques
func setupPublicRoutes(rg *gin.RouterGroup) {
	// Groupe des routes d'authentification
	auth := rg.Group("/auth")
	{
		// Envoi OTP
		auth.POST("/otp/send", handlers.SendOTP)
		
		// Vérification OTP et connexion
		auth.POST("/otp/verify", handlers.VerifyOTP)
		
		// Rafraîchissement du token
		auth.POST("/refresh", handlers.RefreshToken)
	}

	// Routes de livraison publiques (pour calculer prix sans authentification)
	delivery := rg.Group("/delivery")
	{
		// Calcul de prix (optionnel: authentifié pour appliquer promos)
		delivery.POST("/price/calculate", middlewares.OptionalAuthMiddleware(), handlers.CalculateDeliveryPrice)
	}

	// Routes de promotion publiques
	promo := rg.Group("/promo")
	{
		// Valider un code promo (optionnel: authentifié pour l'associer à un utilisateur)
		promo.POST("/validate", middlewares.OptionalAuthMiddleware(), handlers.ValidatePromoCode)
	}
}

// setupProtectedRoutes configure les routes protégées
func setupProtectedRoutes(rg *gin.RouterGroup) {
	// Appliquer l'authentification à toutes les routes protégées
	protected := rg.Group("/")
	protected.Use(middlewares.AuthMiddleware())

	// Routes d'authentification protégées
	setupAuthRoutes(protected)
	
	// Routes utilisateur
	setupUserRoutes(protected)
	
	// Routes de livraison
	setupDeliveryRoutes(protected)
	
	// Routes de promotion
	setupPromoRoutes(protected)
	
	// Routes administrateur
	setupAdminRoutes(protected)
}

// setupAuthRoutes configure les routes d'authentification protégées
func setupAuthRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		// Déconnexion
		auth.POST("/logout", handlers.Logout)
		
		// Profil utilisateur
		auth.GET("/profile", handlers.GetProfile)
		auth.PUT("/profile", handlers.UpdateProfile)
	}
}

// setupUserRoutes configure les routes utilisateur
func setupUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	{
		// Profil utilisateur (accessible par l'utilisateur lui-même ou admin)
		users.GET("/:user_id", middlewares.RequireResourceOwner("user_id"), handlers.GetUserProfile)
		users.PUT("/:user_id", middlewares.RequireResourceOwner("user_id"), handlers.UpdateUserProfile)
		
		// Historique des livraisons de l'utilisateur
		users.GET("/:user_id/deliveries", middlewares.RequireResourceOwner("user_id"), handlers.GetUserDeliveries)
		
		// Véhicules de l'utilisateur (pour les livreurs)
		users.GET("/:user_id/vehicles", middlewares.RequireResourceOwner("user_id"), middlewares.RequireDriverOrAdmin(), handlers.GetUserVehicles)
		users.POST("/:user_id/vehicles", middlewares.RequireResourceOwner("user_id"), middlewares.RequireDriver(), handlers.CreateVehicle)
		users.PUT("/:user_id/vehicles/:vehicle_id", middlewares.RequireResourceOwner("user_id"), middlewares.RequireDriver(), handlers.UpdateVehicle)
	}
}

// setupDeliveryRoutes configure les routes de livraison
func setupDeliveryRoutes(rg *gin.RouterGroup) {
	delivery := rg.Group("/delivery")
	{
		// Création de livraison (clients seulement)
		delivery.POST("/", middlewares.RequireClientOrAdmin(), handlers.CreateDelivery)
		
		// Récupération des détails d'une livraison
		delivery.GET("/:delivery_id", handlers.GetDelivery) // Validation de propriété dans le handler
		
		// Mise à jour du statut (livreurs et admins)
		delivery.PATCH("/:delivery_id/status", middlewares.RequireDriverOrAdmin(), handlers.UpdateDeliveryStatus)
		
		// Assignation de livreur (admins seulement ou auto-assignation pour livreurs disponibles)
		delivery.POST("/:delivery_id/assign", handlers.AssignDelivery) // Logique de rôle dans le handler
		
		// Routes spécifiques aux livreurs
		driverRoutes := delivery.Group("/driver")
		driverRoutes.Use(middlewares.RequireDriver())
		{
			// Livraisons disponibles pour assignation
			driverRoutes.GET("/available", handlers.GetAvailableDeliveries)
			
			// Livraisons assignées au livreur
			driverRoutes.GET("/assigned", handlers.GetAssignedDeliveries)
			
			// Accepter une livraison
			driverRoutes.POST("/:delivery_id/accept", handlers.AcceptDelivery)
			
			// Mettre à jour la position
			driverRoutes.POST("/:delivery_id/location", handlers.UpdateDriverLocation)
		}
		
		// Routes spécifiques aux clients
		clientRoutes := delivery.Group("/client")
		clientRoutes.Use(middlewares.RequireClientOrAdmin())
		{
			// Livraisons du client
			clientRoutes.GET("/", handlers.GetClientDeliveries)
			
			// Annuler une livraison
			clientRoutes.POST("/:delivery_id/cancel", handlers.CancelDelivery)
			
			// Suivre une livraison
			clientRoutes.GET("/:delivery_id/track", handlers.TrackDelivery)
		}
	}
}

// setupPromoRoutes configure les routes de promotion
func setupPromoRoutes(rg *gin.RouterGroup) {
	promo := rg.Group("/promo")
	{
		// Utiliser un code promo
		promo.POST("/use", handlers.UsePromoCode)
		
		// Historique des promotions utilisées
		promo.GET("/history", handlers.GetPromoHistory)
		
		// Routes de parrainage
		referral := promo.Group("/referral")
		{
			// Créer un lien de parrainage
			referral.POST("/create", handlers.CreateReferral)
			
			// Obtenir les statistiques de parrainage
			referral.GET("/stats", handlers.GetReferralStats)
		}
	}
}

// setupAdminRoutes configure les routes administrateur
func setupAdminRoutes(rg *gin.RouterGroup) {
	admin := rg.Group("/admin")
	admin.Use(middlewares.RequireAdmin())
	{
		// Gestion des utilisateurs
		users := admin.Group("/users")
		{
			users.GET("/", handlers.GetAllUsers)
			users.GET("/:user_id", handlers.GetUserDetails)
			users.PUT("/:user_id/role", handlers.UpdateUserRole)
			users.DELETE("/:user_id", handlers.DeleteUser)
		}
		
		// Gestion des livraisons
		deliveries := admin.Group("/deliveries")
		{
			deliveries.GET("/", handlers.GetAllDeliveries)
			deliveries.GET("/stats", handlers.GetDeliveryStats)
			deliveries.POST("/:delivery_id/assign/:driver_id", handlers.ForceAssignDelivery)
		}
		
		// Gestion des livreurs
		drivers := admin.Group("/drivers")
		{
			drivers.GET("/", handlers.GetAllDrivers)
			drivers.GET("/:driver_id/stats", handlers.GetDriverStats)
			drivers.PUT("/:driver_id/status", handlers.UpdateDriverStatus)
		}
		
		// Gestion des promotions
		promotions := admin.Group("/promotions")
		{
			promotions.GET("/", handlers.GetAllPromotions)
			promotions.POST("/", handlers.CreatePromotion)
			promotions.PUT("/:promo_id", handlers.UpdatePromotion)
			promotions.DELETE("/:promo_id", handlers.DeletePromotion)
			promotions.GET("/:promo_id/stats", handlers.GetPromotionStats)
		}
		
		// Gestion des véhicules
		vehicles := admin.Group("/vehicles")
		{
			vehicles.GET("/", handlers.GetAllVehicles)
			vehicles.PUT("/:vehicle_id/verify", handlers.VerifyVehicle)
		}
		
		// Statistiques générales
		stats := admin.Group("/stats")
		{
			stats.GET("/dashboard", handlers.GetDashboardStats)
			stats.GET("/revenue", handlers.GetRevenueStats)
			stats.GET("/users", handlers.GetUserStats)
		}
	}
}

// Routes WebSocket pour le temps réel (à implémenter plus tard)
func setupWebSocketRoutes(rg *gin.RouterGroup) {
	ws := rg.Group("/ws")
	ws.Use(middlewares.AuthMiddleware())
	{
		// Suivi en temps réel des livraisons
		ws.GET("/delivery/:delivery_id", handlers.DeliveryWebSocket)
		
		// Notifications en temps réel pour les livreurs
		ws.GET("/driver/notifications", middlewares.RequireDriver(), handlers.DriverNotificationsWebSocket)
		
		// Notifications en temps réel pour les clients
		ws.GET("/client/notifications", middlewares.RequireClient(), handlers.ClientNotificationsWebSocket)
	}
}
