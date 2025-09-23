package routes

import (
	"time"

	"github.com/ambroise1219/livraison_go/db"
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

	// Page racine simple pour vérifier le service
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Backend en cours d'exécution",
		})
	})

	// Health check endpoint optimisé (pas d'authentification requise)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "ilex-backend",
			"version":   "1.0.0",
			"database":  "PostgreSQL connected",
		})
	})

	// Performance stats endpoint
	router.GET("/stats", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"performance":     "TODO: Implémenter",
			"connection_pool": "TODO: Implémenter",
		})
	})

	// Database management interface (pas d'authentification requise)
	router.GET("/db", func(c *gin.Context) {
		stats, err := db.GetDatabaseStats()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"message": "Database management interface",
			"stats":   stats,
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

	// Routes temps réel (SSE + WebSocket)
	setupRealtimeRoutes(protected)

	// Routes de support
	setupSupportRoutes(protected)
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

		// Upload photo de profil (multipart/form-data, champ: file)
		auth.POST("/profile/picture", handlers.UploadProfilePicture)
		auth.GET("/test/cloudinary", handlers.TestCloudinaryUploader)
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

// setupRealtimeRoutes configure les routes temps réel (SSE + WebSocket)
func setupRealtimeRoutes(rg *gin.RouterGroup) {
	// Initialiser le handler temps réel
	realtimeHandler := handlers.NewRealtimeHandler()

	// Routes SSE (Server-Sent Events) - connexions temps réel
	sse := rg.Group("/sse")
	sse.Use(middlewares.AuthMiddleware())
	{
		// SSE pour suivi livraison (client + driver)
		sse.GET("/delivery/:deliveryId", realtimeHandler.SSEHandler)
	}

	// Routes WebSocket - chat et interactivité
	ws := rg.Group("/ws")
	ws.Use(middlewares.AuthMiddleware())
	{
		// WebSocket pour chat temps réel
		ws.GET("/chat/:deliveryId", realtimeHandler.WebSocketHandler)
	}

	// Routes API temps réel - mises à jour
	realtime := rg.Group("/realtime")
	realtime.Use(middlewares.AuthMiddleware())
	{
		// Mise à jour position livreur
		realtime.POST("/location/:driverId/:deliveryId", middlewares.RequireDriver(), realtimeHandler.UpdateLocationHandler)
		// Récupérer position livreur
		realtime.GET("/location/:driverId", realtimeHandler.GetDriverLocationHandler)
		// Mise à jour statut livraison
		realtime.POST("/delivery/:deliveryId/status", middlewares.RequireDriverOrAdmin(), realtimeHandler.UpdateDeliveryStatusHandler)
		// Envoyer notification
		realtime.POST("/notification/:userId", middlewares.RequireAdmin(), realtimeHandler.SendNotificationHandler)
		// Calculer ETA
		realtime.GET("/eta/:deliveryId", realtimeHandler.CalculateETAHandler)
		// Statistiques temps réel
		realtime.GET("/stats", middlewares.RequireAdmin(), realtimeHandler.GetRealtimeStatsHandler)
	}
}

// setupSupportRoutes configure les routes du système de support
func setupSupportRoutes(rg *gin.RouterGroup) {
	// Initialiser les handlers de support
	handlers.InitSupportHandlers()

	// Routes de support (tickets)
	support := rg.Group("/support")
	{
		// Routes tickets - accessibles à tous les utilisateurs authentifiés
		tickets := support.Group("/tickets")
		{
			// Créer un nouveau ticket (client/livreur)
			tickets.POST("/", handlers.CreateSupportTicket)

			// Lister les tickets (selon permissions)
			tickets.GET("/", handlers.GetSupportTickets)

			// Détails d'un ticket
			tickets.GET("/:ticket_id", handlers.GetSupportTicketByID)

			// Messages d'un ticket
			tickets.GET("/:ticket_id/messages", handlers.GetSupportMessages)
			tickets.POST("/:ticket_id/messages", handlers.AddSupportMessage)

			// Historique des réassignations (staff seulement)
			tickets.GET("/:ticket_id/history", middlewares.RequireStaff(), handlers.GetReassignmentHistory)

			// Routes staff seulement
			staffRoutes := tickets.Group("/")
			staffRoutes.Use(middlewares.RequireStaff())
			{
				// Mettre à jour le statut d'un ticket
				staffRoutes.PUT("/:ticket_id/status", handlers.UpdateSupportTicketStatus)

				// Réassigner un ticket
				staffRoutes.POST("/:ticket_id/reassign", handlers.ReassignSupportTicket)
			}
		}

		// Statistiques de support
		support.GET("/stats", handlers.GetSupportStats)
	}

	// Routes de chat interne (staff seulement)
	internal := rg.Group("/internal")
	internal.Use(middlewares.RequireStaff())
	{
		// Groupes de chat interne
		groups := internal.Group("/groups")
		{
			// Créer un groupe
			groups.POST("/", handlers.CreateInternalGroup)

			// Lister mes groupes
			groups.GET("/", handlers.GetInternalGroups)

			// Messages d'un groupe
			groups.GET("/:group_id/messages", handlers.GetGroupMessages)
			groups.POST("/:group_id/messages", handlers.AddGroupMessage)
		}
	}

	// Routes admin pour contact direct
	adminSupport := rg.Group("/admin/support")
	adminSupport.Use(middlewares.RequireStaff())
	{
		// Initier un contact direct avec un utilisateur
		adminSupport.POST("/contact", handlers.InitiateDirectContact)
	}
}
