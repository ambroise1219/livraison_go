package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/middlewares"
	"github.com/ambroise1219/livraison_go/models"
	"github.com/ambroise1219/livraison_go/services/auth"
	"github.com/ambroise1219/livraison_go/services/delivery"
	"github.com/ambroise1219/livraison_go/services/promo"
	"github.com/ambroise1219/livraison_go/services/validation"
	"github.com/ambroise1219/livraison_go/services/vehicle"
)

// Global validator instance
var validate *validator.Validate
var otpService *auth.OTPService
var userService *auth.UserService
var jwtService *auth.JWTService
var phoneValidator *validation.PhoneValidator

// Delivery services
var deliveryService *delivery.DeliveryService
var updateService *delivery.UpdateService
var simpleCreationService *delivery.SimpleCreationService
var expressCreationService *delivery.ExpressCreationService

// Vehicle service
var vehicleService *vehicle.VehicleService

// Promo service
var promoCodesService *promo.PromoCodesService

// InitHandlers initializes handlers with dependencies
func InitHandlers() {
	validate = validator.New()
	cfg := config.GetConfig()
	otpService = auth.NewOTPService(cfg)
	userService = auth.NewUserService()
	jwtService = auth.NewJWTService(cfg.JWTSecret, time.Duration(cfg.JWTExpiration)*time.Hour)
	phoneValidator = validation.NewPhoneValidator()
	
	// Initialize delivery services
	deliveryService = delivery.NewDeliveryService()
	updateService = delivery.NewUpdateService()
	simpleCreationService = delivery.NewSimpleCreationService()
	expressCreationService = delivery.NewExpressCreationService()
	
	// Initialize vehicle service
	vehicleService = vehicle.NewVehicleService()
	
	// Initialize promo service
	promoCodesService = promo.NewPromoCodesService(cfg)
}

// Auth handlers
func SendOTP(c *gin.Context) {
	var req models.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate request
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	// Validation spécialisée du numéro de téléphone
	if err := phoneValidator.ValidatePhone(req.Phone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Numéro de téléphone invalide", "details": err.Error()})
		return
	}

	// Normaliser le numéro de téléphone
	normalizedPhone, err := phoneValidator.NormalizePhone(req.Phone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur de normalisation du téléphone", "details": err.Error()})
		return
	}

	// Sauvegarder l'OTP avec le numéro normalisé
	_, err = otpService.SaveOTP(normalizedPhone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP", "details": err.Error()})
		return
	}

	// Envoyer l'OTP via WhatsApp avec le numéro normalisé
	if _, err := otpService.SendOTPByWhatsApp(normalizedPhone, "", ""); err != nil {
		// Don't fail if OTP sending fails, just log it
		// In production, you might want to handle this differently
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "OTP sent successfully",
		"expiresIn": "5 minutes",
	})
}

func VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate request
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	// Validation spécialisée du numéro de téléphone
	if err := phoneValidator.ValidatePhone(req.Phone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Numéro de téléphone invalide", "details": err.Error()})
		return
	}

	// Normaliser le numéro de téléphone
	normalizedPhone, err := phoneValidator.NormalizePhone(req.Phone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur de normalisation du téléphone", "details": err.Error()})
		return
	}

	// Vérifier l'OTP avec le numéro normalisé
	_, err = otpService.VerifyOTP(normalizedPhone, req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP", "details": err.Error()})
		return
	}

	// Trouver ou créer l'utilisateur avec le numéro normalisé
	user, isNewUser, err := userService.FindOrCreateUser(normalizedPhone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user", "details": err.Error()})
		return
	}

	// Générer un token JWT
	token, err := jwtService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token", "details": err.Error()})
		return
	}

	response := gin.H{
		"message":     "Authentication successful",
		"isNewUser":   isNewUser,
		"user":        user,
		"accessToken": token,
		"tokenType":   "Bearer",
		"expiresIn":   24 * 3600, // 24 hours in seconds
	}

	c.JSON(http.StatusOK, response)
}

func RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate request
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	// Rafraîchir le token JWT
	newToken, err := jwtService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to refresh token", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Token refreshed successfully",
		"accessToken": newToken,
		"tokenType":   "Bearer",
		"expiresIn":   24 * 3600, // 24 hours in seconds
	})
}

func Logout(c *gin.Context) {
	// Récupérer le token depuis l'en-tête Authorization
	auth := c.GetHeader("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token manquant ou format invalide"})
		return
	}

	token := strings.TrimPrefix(auth, "Bearer ")
	
	// Invalidation du token (ajout à une blacklist)
	err := jwtService.InvalidateToken(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la déconnexion", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Déconnexion réussie"})
}

func GetProfile(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Récupérer le profil complet depuis la base de données
	user, err := userService.GetUserByPhone(userClaims.UserID) // userID est en fait le phone
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Profil récupéré avec succès",
		"user": user,
	})
}

func UpdateProfile(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Structure pour les données de mise à jour
	type UpdateProfileRequest struct {
		FirstName     *string `json:"firstName,omitempty"`
		LastName      *string `json:"lastName,omitempty"`
		Email         *string `json:"email,omitempty"`
		Address       *string `json:"address,omitempty"`
		DateOfBirth   *string `json:"dateOfBirth,omitempty"`
		LieuResidence *string `json:"lieuResidence,omitempty"`
	}
	
	// Valider les données d'entrée
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	// Récupérer l'utilisateur actuel
	user, err := userService.GetUserByPhone(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé", "details": err.Error()})
		return
	}
	
	// Appliquer les mises à jour
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Email != nil {
		user.Email = req.Email
	}
	if req.Address != nil {
		user.Address = req.Address
	}
	if req.LieuResidence != nil {
		user.LieuResidence = req.LieuResidence
	}
	
	// Gérer dateOfBirth (conversion string vers time.Time)
	if req.DateOfBirth != nil {
		if *req.DateOfBirth != "" {
			parsedDate, err := time.Parse("2006-01-02", *req.DateOfBirth)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Format de date invalide (attendu: YYYY-MM-DD)", "details": err.Error()})
				return
			}
			user.DateOfBirth = &parsedDate
		}
	}
	
	// Sauvegarder les modifications
	err = userService.UpdateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Profil mis à jour avec succès",
		"user": user,
	})
}

// User handlers
func GetUserProfile(c *gin.Context) {
	// GetUserProfile est identique à GetProfile, redirection
	GetProfile(c)
}

func UpdateUserProfile(c *gin.Context) {
	// UpdateUserProfile est identique à UpdateProfile, redirection
	UpdateProfile(c)
}

func GetUserDeliveries(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Récupérer les livraisons du client
	deliveries, err := deliveryService.GetDeliveriesByClient(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des livraisons", "details": err.Error()})
		return
	}
	
	// Convertir en réponses
	var responses []gin.H
	for _, delivery := range deliveries {
		responses = append(responses, gin.H{
			"id":         delivery.ID,
			"type":       delivery.Type,
			"status":     delivery.Status,
			"createdAt":  delivery.CreatedAt,
			"updatedAt":  delivery.UpdatedAt,
			"pickupId":   delivery.PickupID,
			"dropoffId":  delivery.DropoffID,
			"finalPrice": delivery.FinalPrice,
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Livraisons récupérées avec succès",
		"deliveries": responses,
		"count": len(responses),
	})
}

func GetUserVehicles(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Récupérer les véhicules de l'utilisateur
	vehicles, err := vehicleService.GetVehiclesByOwner(userClaims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des véhicules", "details": err.Error()})
		return
	}
	
	// Convertir en réponses
	var responses []gin.H
	for _, vehicle := range vehicles {
		responses = append(responses, gin.H{
			"id":                     vehicle.ID,
			"type":                   vehicle.Type,
			"nom":                    vehicle.Nom,
			"plaqueImmatriculation":  vehicle.PlaqueImmatriculation,
			"couleur":                vehicle.Couleur,
			"marque":                 vehicle.Marque,
			"modele":                 vehicle.Modele,
			"annee":                  vehicle.Annee,
			"createdAt":              vehicle.CreatedAt,
			"isRegistrationComplete": vehicle.IsRegistrationComplete(),
			"hasRequiredDocuments":   vehicle.HasRequiredDocuments(),
		})
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Véhicules récupérés avec succès",
		"vehicles": responses,
		"count": len(responses),
	})
}

func CreateVehicle(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Valider les données d'entrée
	var req models.CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}
	
	// Créer le véhicule
	vehicle, err := vehicleService.CreateVehicle(userClaims.UserID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du véhicule", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Véhicule créé avec succès",
		"vehicle": vehicle.ToResponse(),
	})
}

func UpdateVehicle(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	vehicleID := c.Param("vehicle_id")
	if vehicleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de véhicule requis"})
		return
	}
	
	// Vérifier que le véhicule existe
	_, err := vehicleService.GetVehicleByID(vehicleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Véhicule non trouvé", "details": err.Error()})
		return
	}
	
	// Vérifier que le véhicule appartient à l'utilisateur (ou est admin)
	// Note: Le schéma actuel ne lie pas directement les véhicules aux utilisateurs
	// Cette vérification peut être adaptée selon les besoins
	if userClaims.Role != models.UserRoleAdmin {
		// Pour l'instant, on autorise tous les utilisateurs authentifiés
		// À adapter selon la logique métier
	}
	
	// Valider les données d'entrée
	var req models.UpdateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	// Mettre à jour le véhicule
	updatedVehicle, err := vehicleService.UpdateVehicle(vehicleID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour du véhicule", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Véhicule mis à jour avec succès",
		"vehicle": updatedVehicle.ToResponse(),
	})
}

// Delivery handlers
func CreateDelivery(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Valider les données d'entrée
	var req models.CreateDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	// Valider avec le validateur
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}
	
	// Router vers le bon service selon le type de livraison
	var response *models.DeliveryResponse
	var err error
	
	switch req.Type {
	case models.DeliveryTypeStandard:
		response, err = simpleCreationService.CreateSimpleDelivery(userClaims.UserID, &req)
	case models.DeliveryTypeExpress:
		response, err = expressCreationService.CreateExpressDelivery(userClaims.UserID, &req)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type de livraison non supporté"})
		return
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création de la livraison", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Livraison créée avec succès",
		"delivery": response,
	})
}

func GetDelivery(c *gin.Context) {
	// Récupérer l'ID de la livraison depuis l'URL
	deliveryID := c.Param("delivery_id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de livraison requis"})
		return
	}
	
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Récupérer la livraison
	delivery, err := deliveryService.GetDelivery(deliveryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Livraison non trouvée", "details": err.Error()})
		return
	}
	
	// Vérifier que l'utilisateur a accès à cette livraison
	if delivery.ClientID != userClaims.UserID && 
	   (delivery.LivreurID == nil || *delivery.LivreurID != userClaims.UserID) &&
	   userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès interdit à cette livraison"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"delivery": delivery.ToResponse(),
	})
}

func UpdateDeliveryStatus(c *gin.Context) {
	// Récupérer l'ID de la livraison
	deliveryID := c.Param("delivery_id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de livraison requis"})
		return
	}
	
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Structure pour la requête de mise à jour
	type UpdateStatusRequest struct {
		Status models.DeliveryStatus `json:"status" validate:"required"`
	}
	
	// Valider les données d'entrée
	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}
	
	// Vérifier que le statut est valide
	if !req.Status.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Statut de livraison invalide"})
		return
	}
	
	// Récupérer la livraison existante
	delivery, err := deliveryService.GetDelivery(deliveryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Livraison non trouvée", "details": err.Error()})
		return
	}
	
	// Vérifier les permissions (seul le livreur assigné ou admin peut mettre à jour)
	if userClaims.Role != models.UserRoleAdmin &&
	   (delivery.LivreurID == nil || *delivery.LivreurID != userClaims.UserID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès interdit pour mettre à jour cette livraison"})
		return
	}
	
	// Mettre à jour le statut
	delivery.Status = req.Status
	err = deliveryService.UpdateDelivery(delivery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Statut mis à jour avec succès",
		"delivery": delivery.ToResponse(),
	})
}

func AssignDelivery(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Seuls les admins peuvent assigner des livraisons"})
		return
	}
	
	// Structure pour la requête d'assignation
	type AssignRequest struct {
		DeliveryID string  `json:"deliveryId" validate:"required"`
		DriverID   *string `json:"driverId,omitempty"` // Si vide, auto-assign
	}
	
	// Valider les données d'entrée
	var req AssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}
	
	// Récupérer la livraison
	delivery, err := deliveryService.GetDelivery(req.DeliveryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Livraison non trouvée", "details": err.Error()})
		return
	}
	
	// Vérifier que la livraison peut être assignée
	if !delivery.CanBeAssigned() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cette livraison ne peut pas être assignée", 
			"details": fmt.Sprintf("Status actuel: %s", delivery.Status),
		})
		return
	}
	
	// Assigner le livreur
	if req.DriverID != nil {
		// Assignment manuel à un livreur spécifique
		// Vérifier que le livreur existe et est disponible
		driver, err := userService.GetUserByID(*req.DriverID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Livreur non trouvé", "details": err.Error()})
			return
		}
		
		// Vérifier que l'utilisateur est bien livreur
		if driver.Role != models.UserRoleLivreur {
			c.JSON(http.StatusBadRequest, gin.H{"error": "L'utilisateur spécifié n'est pas un livreur"})
			return
		}
		
		delivery.LivreurID = req.DriverID
		delivery.Status = models.DeliveryStatusAssigned
	} else {
		// Auto-assignment (pour plus tard)
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Auto-assignment pas encore implémenté"})
		return
	}
	
	// Mettre à jour la livraison
	err = deliveryService.UpdateDelivery(delivery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'assignation", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Livraison assignée avec succès",
		"delivery": delivery.ToResponse(),
	})
}

func CalculateDeliveryPrice(c *gin.Context) {
	// Structure pour la requête de calcul de prix
	type PriceCalculationRequest struct {
		Type        models.DeliveryType  `json:"type" validate:"required"`
		VehicleType models.VehicleType   `json:"vehicleType" validate:"required"`
		PickupLat   *float64             `json:"pickupLat,omitempty"`
		PickupLng   *float64             `json:"pickupLng,omitempty"`
		DropoffLat  *float64             `json:"dropoffLat,omitempty"`
		DropoffLng  *float64             `json:"dropoffLng,omitempty"`
		WeightKg    *float64             `json:"weightKg,omitempty"`
	}
	
	// Valider les données d'entrée
	var req PriceCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}
	
	// Calculer la distance si coordonnées fournies
	distance := 5.0 // Distance par défaut de 5km
	if req.PickupLat != nil && req.PickupLng != nil && req.DropoffLat != nil && req.DropoffLng != nil {
		// Calcul simple de distance (Haversine approximatif)
		distance = calculateHaversineDistance(*req.PickupLat, *req.PickupLng, *req.DropoffLat, *req.DropoffLng)
	}
	
	// Calculer le prix selon le type
	var price float64
	switch req.Type {
	case models.DeliveryTypeStandard:
		price = calculateSimplePrice(req.VehicleType, distance)
	case models.DeliveryTypeExpress:
		price = calculateExpressPrice(req.VehicleType, distance)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type de livraison non supporté"})
		return
	}
	
	// Réponse avec détails du calcul
	c.JSON(http.StatusOK, gin.H{
		"price": price,
		"details": gin.H{
			"type":        req.Type,
			"vehicleType": req.VehicleType,
			"distance":    distance,
			"currency":    "FCFA",
		},
	})
}

// Fonctions helper pour calcul de prix
func calculateSimplePrice(vehicleType models.VehicleType, distance float64) float64 {
	basePrices := map[models.VehicleType]float64{
		models.VehicleTypeMotorcycle:  500.0,
		models.VehicleTypeCar:         1000.0,
		models.VehicleTypeVan:         1500.0,
	}
	basePrice := basePrices[vehicleType]
	if basePrice == 0 {
		basePrice = 1000.0
	}
	return basePrice + (distance * 200.0)
}

func calculateExpressPrice(vehicleType models.VehicleType, distance float64) float64 {
	basePrices := map[models.VehicleType]float64{
		models.VehicleTypeMotorcycle:  1000.0,
		models.VehicleTypeCar:         2000.0,
		models.VehicleTypeVan:         3000.0,
	}
	basePrice := basePrices[vehicleType]
	if basePrice == 0 {
		basePrice = 2000.0
	}
	expressPrice := (basePrice + (distance * 400.0)) * 1.5
	return expressPrice
}

func calculateHaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	// Approximation simple pour démo
	const R = 6371 // Rayon de la Terre en km
	return R * 0.1 // Approximation basique
}

func GetAvailableDeliveries(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est livreur ou admin
	if userClaims.Role != models.UserRoleAdmin && userClaims.Role != models.UserRoleLivreur {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux livreurs et administrateurs"})
		return
	}
	
	// TODO: Implémenter GetAvailableDeliveries dans le service
	// Pour le moment, retourner une liste vide
	c.JSON(http.StatusOK, gin.H{
		"message": "Livraisons disponibles récupérées avec succès",
		"deliveries": []gin.H{},
		"count": 0,
		"note": "Service delivery.GetAvailableDeliveries à implémenter",
		"filters": gin.H{
			"status": "PENDING",
			"assignable": true,
			"location": nil,
		},
	})
}

func GetAssignedDeliveries(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est livreur ou admin
	if userClaims.Role != models.UserRoleAdmin && userClaims.Role != models.UserRoleLivreur {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux livreurs et administrateurs"})
		return
	}
	
	// TODO: Implémenter GetAssignedDeliveries dans le service pour un livreur spécifique
	// Pour le moment, retourner une liste vide
	c.JSON(http.StatusOK, gin.H{
		"message": "Livraisons assignées récupérées avec succès",
		"deliveries": []gin.H{},
		"count": 0,
		"driverId": userClaims.UserID,
		"note": "Service delivery.GetDeliveriesByDriver à implémenter",
		"filters": gin.H{
			"status": []string{"ASSIGNED", "PICKED_UP", "IN_TRANSIT"},
			"driverId": userClaims.UserID,
		},
	})
}

func AcceptDelivery(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est livreur
	if userClaims.Role != models.UserRoleLivreur {
		c.JSON(http.StatusForbidden, gin.H{"error": "Seuls les livreurs peuvent accepter des livraisons"})
		return
	}
	
	// Récupérer l'ID de la livraison
	deliveryID := c.Param("delivery_id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de livraison requis"})
		return
	}
	
	// Récupérer la livraison
	delivery, err := deliveryService.GetDelivery(deliveryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Livraison non trouvée", "details": err.Error()})
		return
	}
	
	// Vérifier que la livraison peut être acceptée
	if !delivery.CanBeAssigned() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cette livraison ne peut pas être acceptée",
			"details": fmt.Sprintf("Status actuel: %s", delivery.Status),
		})
		return
	}
	
	// Assigner le livreur et mettre à jour le statut
	delivery.LivreurID = &userClaims.UserID
	delivery.Status = models.DeliveryStatusAssigned
	
	// Mettre à jour la livraison
	err = deliveryService.UpdateDelivery(delivery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'acceptation", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Livraison acceptée avec succès",
		"delivery": delivery.ToResponse(),
		"driver": gin.H{
			"id": userClaims.UserID,
			"phone": userClaims.Phone,
		},
	})
}

func UpdateDriverLocation(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est livreur
	if userClaims.Role != models.UserRoleLivreur {
		c.JSON(http.StatusForbidden, gin.H{"error": "Seuls les livreurs peuvent mettre à jour leur position"})
		return
	}
	
	// Structure pour la mise à jour de localisation
	type UpdateLocationRequest struct {
		Lat         float64 `json:"lat" validate:"required,gte=-90,lte=90"`
		Lng         float64 `json:"lng" validate:"required,gte=-180,lte=180"`
		IsAvailable *bool   `json:"isAvailable,omitempty"`
	}
	
	// Valider les données d'entrée
	var req UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}
	
	// TODO: Implémenter la mise à jour de la localisation du conducteur
	// Pour le moment, simuler la sauvegarde
	isAvailable := true
	if req.IsAvailable != nil {
		isAvailable = *req.IsAvailable
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Position mise à jour avec succès",
		"location": gin.H{
			"lat": req.Lat,
			"lng": req.Lng,
			"timestamp": time.Now(),
		},
		"isAvailable": isAvailable,
		"driver": gin.H{
			"id": userClaims.UserID,
			"phone": userClaims.Phone,
		},
	})
}

// Delivery Update Handlers

// UpdateSimpleDelivery updates a simple delivery
func UpdateSimpleDelivery(c *gin.Context) {
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}

	deliveryID := c.Param("delivery_id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de livraison requis"})
		return
	}

	var req models.UpdateSimpleDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	// Check permissions (owner, driver assigned, or admin)
	if !canUpdateDelivery(userClaims, deliveryID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès non autorisé pour cette livraison"})
		return
	}

	updatedDelivery, err := updateService.UpdateSimpleDelivery(deliveryID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update delivery", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Livraison simple mise à jour avec succès",
		"delivery": updatedDelivery.ToResponse(),
	})
}

// UpdateExpressDelivery updates an express delivery
func UpdateExpressDelivery(c *gin.Context) {
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}

	deliveryID := c.Param("delivery_id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de livraison requis"})
		return
	}

	var req models.UpdateExpressDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	if !canUpdateDelivery(userClaims, deliveryID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès non autorisé pour cette livraison"})
		return
	}

	updatedDelivery, err := updateService.UpdateExpressDelivery(deliveryID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update delivery", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Livraison express mise à jour avec succès",
		"delivery": updatedDelivery.ToResponse(),
	})
}

// UpdateGroupedDelivery updates a grouped delivery
func UpdateGroupedDelivery(c *gin.Context) {
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}

	deliveryID := c.Param("delivery_id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de livraison requis"})
		return
	}

	var req models.UpdateGroupedDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	if !canUpdateDelivery(userClaims, deliveryID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès non autorisé pour cette livraison"})
		return
	}

	updatedDelivery, err := updateService.UpdateGroupedDelivery(deliveryID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update delivery", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Livraison groupée mise à jour avec succès",
		"delivery": updatedDelivery.ToResponse(),
	})
}

// UpdateMovingDelivery updates a moving delivery
func UpdateMovingDelivery(c *gin.Context) {
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}

	deliveryID := c.Param("delivery_id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de livraison requis"})
		return
	}

	var req models.UpdateMovingDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	if !canUpdateDelivery(userClaims, deliveryID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès non autorisé pour cette livraison"})
		return
	}

	updatedDelivery, err := updateService.UpdateMovingDelivery(deliveryID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update delivery", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Livraison déménagement mise à jour avec succès",
		"delivery": updatedDelivery.ToResponse(),
	})
}

// UpdateDeliveryStatus updates only the delivery status with tracking
func UpdateDeliveryStatus(c *gin.Context) {
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}

	deliveryID := c.Param("delivery_id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de livraison requis"})
		return
	}

	var req models.StatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}

	// Only drivers and admins can update status
	if userClaims.Role != models.UserRoleLivreur && userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Seuls les livreurs et administrateurs peuvent mettre à jour le statut"})
		return
	}

	updatedDelivery, err := updateService.UpdateDeliveryStatus(deliveryID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update delivery status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Statut de livraison mis à jour avec succès",
		"delivery": updatedDelivery.ToResponse(),
		"tracking": gin.H{
			"status": req.Status,
			"timestamp": time.Now(),
			"location": req.Location,
			"notes": req.Notes,
		},
	})
}

// Helper functions

// canUpdateDelivery checks if user can update the delivery
func canUpdateDelivery(userClaims *middlewares.UserClaims, deliveryID string) bool {
	// Admin can always update
	if userClaims.Role == models.UserRoleAdmin {
		return true
	}

	// Get delivery to check ownership and assignment
	delivery, err := deliveryService.GetDelivery(deliveryID)
	if err != nil {
		return false
	}

	// Client can update their own delivery (before assignment)
	if userClaims.Role == models.UserRoleClient && delivery.ClientID == userClaims.Phone {
		return delivery.Status == models.DeliveryStatusPending
	}

	// Driver can update assigned delivery
	if userClaims.Role == models.UserRoleLivreur && delivery.LivreurID != nil && *delivery.LivreurID == userClaims.UserID {
		return true
	}

	return false
	
	// Construire la réponse avec la localisation mise à jour
	location := gin.H{
		"driverId":   userClaims.UserID,
		"lat":        req.Lat,
		"lng":        req.Lng,
		"isAvailable": isAvailable,
		"timestamp":  time.Now(),
		"accuracy":   nil, // TODO: Gérer la précision GPS
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Position mise à jour avec succès",
		"location": location,
		"note": "Service location tracking à implémenter avec Redis/cache temps réel",
	})
}

func GetClientDeliveries(c *gin.Context) {
	// GetClientDeliveries est identique à GetUserDeliveries
	GetUserDeliveries(c)
}

func CancelDelivery(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Récupérer l'ID de la livraison
	deliveryID := c.Param("delivery_id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de livraison requis"})
		return
	}
	
	// Récupérer la livraison
	delivery, err := deliveryService.GetDelivery(deliveryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Livraison non trouvée", "details": err.Error()})
		return
	}
	
	// Vérifier que l'utilisateur a le droit d'annuler cette livraison
	if delivery.ClientID != userClaims.UserID && userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès interdit pour annuler cette livraison"})
		return
	}
	
	// Vérifier que la livraison peut être annulée
	if !delivery.CanBeCancelled() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cette livraison ne peut plus être annulée",
			"details": fmt.Sprintf("Status actuel: %s", delivery.Status),
		})
		return
	}
	
	// Mettre à jour le statut à annulé
	delivery.Status = models.DeliveryStatusCancelled
	err = deliveryService.UpdateDelivery(delivery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'annulation", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Livraison annulée avec succès",
		"delivery": delivery.ToResponse(),
	})
}

func TrackDelivery(c *gin.Context) {
	// Récupérer l'ID de la livraison
	deliveryID := c.Param("delivery_id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de livraison requis"})
		return
	}
	
	// Récupérer l'utilisateur authentifié (optionnel pour le tracking)
	userClaims, _ := middlewares.GetCurrentUser(c)
	
	// Récupérer la livraison
	delivery, err := deliveryService.GetDelivery(deliveryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Livraison non trouvée", "details": err.Error()})
		return
	}
	
	// Si l'utilisateur est authentifié, vérifier l'accès
	if userClaims != nil {
		if delivery.ClientID != userClaims.UserID && 
		   (delivery.LivreurID == nil || *delivery.LivreurID != userClaims.UserID) &&
		   userClaims.Role != models.UserRoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Accès interdit à cette livraison"})
			return
		}
	}
	
	// Informations de tracking (simulation)
	trackingInfo := gin.H{
		"deliveryId":    delivery.ID,
		"status":        delivery.Status,
		"type":          delivery.Type,
		"createdAt":     delivery.CreatedAt,
		"updatedAt":     delivery.UpdatedAt,
		"pickupId":      delivery.PickupID,
		"dropoffId":     delivery.DropoffID,
		"estimatedTime": nil, // TODO: Calculer le temps estimé
		"currentLocation": nil, // TODO: Localisation en temps réel du livreur
	}
	
	if delivery.LivreurID != nil {
		trackingInfo["driverId"] = *delivery.LivreurID
		// TODO: Récupérer les infos du livreur (nom, téléphone)
	}
	
	if delivery.DistanceKm != nil {
		trackingInfo["distance"] = *delivery.DistanceKm
	}
	
	if delivery.DurationMin != nil {
		trackingInfo["duration"] = *delivery.DurationMin
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Informations de suivi récupérées avec succès",
		"tracking": trackingInfo,
	})
}

// Promo handlers
func ValidatePromoCode(c *gin.Context) {
	// Structure pour la validation de code promo
	type ValidatePromoRequest struct {
		Code   string  `json:"code" validate:"required"`
		Amount float64 `json:"amount" validate:"required,gt=0"`
	}
	
	// Valider les données d'entrée
	var req ValidatePromoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}
	
	// Récupérer l'utilisateur authentifié (optionnel pour validation future)
	_, _ = middlewares.GetCurrentUser(c)
	
	// TODO: Implémenter le service de validation des promos
	// cfg := config.GetConfig()
	// promoService := promo.NewPromoValidationService(cfg)
	// result, err := promoService.ValidatePromoCode(req.Code, req.Amount, userID)
	
	// Simulation de validation pour le moment
	isValid := req.Code != "INVALID" && req.Amount >= 1000
	discount := 0.0
	finalPrice := req.Amount
	
	if isValid {
		if req.Code == "WELCOME10" {
			discount = req.Amount * 0.1 // 10% de réduction
		} else if req.Code == "SAVE500" {
			discount = 500.0 // 500 FCFA de réduction
		}
		finalPrice = req.Amount - discount
	}
	
	c.JSON(http.StatusOK, gin.H{
		"valid":      isValid,
		"code":       req.Code,
		"discount":   discount,
		"finalPrice": finalPrice,
		"message":    func() string {
			if isValid {
				return "Code promotionnel valide"
			}
			return "Code promotionnel invalide ou montant insuffisant"
		}(),
		"note": "Service promo à implémenter - validation simulée",
	})
}

func UsePromoCode(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Structure pour l'utilisation de code promo
	type UsePromoRequest struct {
		Code       string `json:"code" validate:"required"`
		Amount     float64 `json:"amount" validate:"required,gt=0"`
		DeliveryID string `json:"deliveryId" validate:"required"`
	}
	
	// Valider les données d'entrée
	var req UsePromoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
		return
	}
	
	// Vérifier que la livraison existe et appartient à l'utilisateur
	delivery, err := deliveryService.GetDelivery(req.DeliveryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Livraison non trouvée", "details": err.Error()})
		return
	}
	
	if delivery.ClientID != userClaims.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès interdit à cette livraison"})
		return
	}
	
	// TODO: Implémenter le service d'application des promos
	// cfg := config.GetConfig()
	// promoService := promo.NewPromoValidationService(cfg)
	// usage, err := promoService.ApplyPromo(req.Code, req.Amount, userClaims.UserID)
	
	// Simulation d'application pour le moment
	isValid := req.Code != "INVALID" && req.Amount >= 1000
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code promotionnel invalide"})
		return
	}
	
	discount := 0.0
	if req.Code == "WELCOME10" {
		discount = req.Amount * 0.1
	} else if req.Code == "SAVE500" {
		discount = 500.0
	}
	
	finalPrice := req.Amount - discount
	
	// Mettre à jour le prix de la livraison
	delivery.FinalPrice = finalPrice
	err = deliveryService.UpdateDelivery(delivery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'application du code", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":    "Code promotionnel appliqué avec succès",
		"code":       req.Code,
		"discount":   discount,
		"finalPrice": finalPrice,
		"delivery":   delivery.ToResponse(),
		"usage": gin.H{
			"id":       fmt.Sprintf("usage_%d", time.Now().Unix()),
			"userId":   userClaims.UserID,
			"code":     req.Code,
			"discount": discount,
			"usedAt":   time.Now(),
		},
		"note": "Service promo à implémenter - application simulée",
	})
}

func GetPromoHistory(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// TODO: Implémenter la récupération de l'historique des promos
	// Pour le moment, retourner une liste vide
	c.JSON(http.StatusOK, gin.H{
		"message": "Historique des codes promotionnels récupéré avec succès",
		"userId": userClaims.UserID,
		"promoUsages": []gin.H{},
		"totalSavings": 0.0,
		"count": 0,
		"note": "Service promo.GetUserPromoUsages à implémenter",
	})
}

func CreateReferral(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Générer un code de parrainage unique
	referralCode := fmt.Sprintf("%s_%d", strings.ToUpper(userClaims.UserID[:8]), time.Now().Unix())
	
	// TODO: Implémenter la création de code de parrainage dans le service
	// Pour le moment, simuler la création
	referral := gin.H{
		"id": fmt.Sprintf("ref_%d", time.Now().Unix()),
		"code": referralCode,
		"referrerId": userClaims.UserID,
		"isUsed": false,
		"usedAt": nil,
		"createdAt": time.Now(),
		"reward": gin.H{
			"type": "discount",
			"value": 1000.0,
			"description": "1000 FCFA de réduction pour le parrain et le filleul",
		},
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Code de parrainage créé avec succès",
		"referral": referral,
		"note": "Service referral.CreateReferral à implémenter",
	})
}

func GetReferralStats(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// TODO: Implémenter les statistiques de parrainage
	// Pour le moment, retourner des stats vides
	stats := gin.H{
		"userId": userClaims.UserID,
		"totalReferrals": 0,
		"successfulReferrals": 0,
		"pendingReferrals": 0,
		"totalEarnings": 0.0,
		"currentReferralCode": nil,
		"referralHistory": []gin.H{},
		"rewards": gin.H{
			"earned": 0.0,
			"pending": 0.0,
			"currency": "FCFA",
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Statistiques de parrainage récupérées avec succès",
		"stats": stats,
		"note": "Service referral.GetUserReferralStats à implémenter",
	})
}

// Admin handlers
func GetAllUsers(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Paramètres de pagination
	page := 1
	limit := 10
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	// Filtre par rôle
	roleFilter := c.Query("role")
	
	// Récupérer tous les utilisateurs avec pagination
	users, total, err := userService.GetAllUsers(page, limit, roleFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des utilisateurs"})
		return
	}
	
	// Calculer le nombre de pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Utilisateurs récupérés avec succès",
		"users": users,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
		},
		"filters": gin.H{
			"role":     roleFilter,
			"search":   c.Query("search"),
			"dateFrom": c.Query("dateFrom"),
			"dateTo":   c.Query("dateTo"),
		},
	})
}

func GetUserDetails(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Récupérer l'ID utilisateur depuis l'URL
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID utilisateur requis"})
		return
	}
	
	// Récupérer les détails de l'utilisateur
	user, err := userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}
	
	// Récupérer les statistiques additionnelles de l'utilisateur
	stats, err := userService.GetUserStats(userID)
	if err != nil {
		// Log l'erreur mais continue avec les données de base
		stats = map[string]interface{}{
			"deliveriesCount": 0,
			"vehiclesCount": 0,
			"averageRating": 0.0,
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Détails utilisateur récupérés avec succès",
		"user": user,
		"stats": stats,
	})
}

func UpdateUserRole(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Récupérer l'ID utilisateur depuis l'URL
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID utilisateur requis"})
		return
	}
	
	// Empêcher l'admin de modifier son propre rôle
	if userID == userClaims.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Impossible de modifier votre propre rôle"})
		return
	}
	
	// Structure pour recevoir le nouveau rôle
	type UpdateRoleRequest struct {
		Role string `json:"role" binding:"required"`
	}
	
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}
	
	// Valider le rôle
	if req.Role != string(models.UserRoleClient) && req.Role != string(models.UserRoleLivreur) && req.Role != string(models.UserRoleAdmin) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rôle invalide"})
		return
	}
	
	// Mettre à jour le rôle de l'utilisateur
	err := userService.UpdateUserRole(userID, models.UserRole(req.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour du rôle"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Rôle utilisateur mis à jour avec succès",
		"userID": userID,
		"newRole": req.Role,
	})
}

func DeleteUser(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Récupérer l'ID utilisateur depuis l'URL
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID utilisateur requis"})
		return
	}
	
	// Empêcher l'admin de supprimer son propre compte
	if userID == userClaims.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Impossible de supprimer votre propre compte"})
		return
	}
	
	// Supprimer l'utilisateur
	err := userService.DeleteUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression de l'utilisateur"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Utilisateur supprimé avec succès",
		"userID": userID,
	})
}

func GetAllDeliveries(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Paramètres de pagination
	page := 1
	limit := 10
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	// Filtres
	filters := map[string]string{
		"status": c.Query("status"),
		"type":   c.Query("type"),
		"date":   c.Query("date"),
	}
	
	// Récupérer toutes les livraisons avec pagination
	deliveries, total, err := deliveryService.GetAllDeliveries(page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des livraisons"})
		return
	}
	
	// Calculer le nombre de pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Livraisons récupérées avec succès",
		"deliveries": deliveries,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
		},
		"filters": filters,
	})
}

func GetDeliveryStats(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// TODO: Implémenter GetDeliveryStats dans le service
	// Simulation de statistiques pour le moment
	stats := gin.H{
		"total": gin.H{
			"count":   0,
			"revenue": 0.0,
		},
		"byStatus": gin.H{
			"pending":   0,
			"assigned":  0,
			"pickedUp":  0,
			"delivered": 0,
			"cancelled": 0,
		},
		"byType": gin.H{
			"standard": 0,
			"express":  0,
			"grouped":  0,
			"moving":   0,
		},
		"byVehicle": gin.H{
			"motorcycle": 0,
			"car":        0,
			"van":        0,
		},
		"recent": gin.H{
			"today":     0,
			"week":      0,
			"month":     0,
		},
		"performance": gin.H{
			"averageDeliveryTime": 0,
			"successRate":         0.0,
			"customerSatisfaction": 0.0,
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Statistiques des livraisons récupérées avec succès",
		"stats": stats,
		"note": "Service delivery stats à implémenter - données simulées",
	})
}

func ForceAssignDelivery(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Récupérer l'ID livraison depuis l'URL
	deliveryID := c.Param("id")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID livraison requis"})
		return
	}
	
	// Structure pour recevoir les données d'assignation
	type ForceAssignRequest struct {
		DriverID string `json:"driverId" binding:"required"`
		Reason   string `json:"reason,omitempty"`
	}
	
	var req ForceAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}
	
	// Vérifier que la livraison existe
	delivery, err := deliveryService.GetDelivery(deliveryID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Livraison non trouvée"})
		return
	}
	
	// Vérifier que le livreur existe et est actif
	driver, err := userService.GetUserByID(req.DriverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Livreur non trouvé"})
		return
	}
	
	if driver.Role != models.UserRoleLivreur {
		c.JSON(http.StatusBadRequest, gin.H{"error": "L'utilisateur n'est pas un livreur"})
		return
	}
	
	// Forcer l'assignation de la livraison
	err = deliveryService.ForceAssignDelivery(deliveryID, req.DriverID, userClaims.UserID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'assignation forcée"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Livraison assignée avec succès",
		"deliveryID": deliveryID,
		"driverID": req.DriverID,
		"assignedBy": userClaims.UserID,
		"reason": req.Reason,
	})
}

func GetAllDrivers(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Paramètres de pagination
	page := 1
	limit := 10
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	// Filtre par statut de livreur
	statusFilter := c.Query("status")
	
	// Récupérer tous les livreurs avec pagination
	drivers, total, err := userService.GetAllDrivers(page, limit, statusFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des livreurs"})
		return
	}
	
	// Calculer le nombre de pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Livreurs récupérés avec succès",
		"drivers": drivers,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
		},
		"filters": gin.H{
			"status": statusFilter,
		},
	})
}

func GetDriverStats(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Récupérer l'ID livreur depuis l'URL
	driverID := c.Param("id")
	if driverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID livreur requis"})
		return
	}
	
	// Vérifier que l'utilisateur est bien un livreur
	driver, err := userService.GetUserByID(driverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Livreur non trouvé"})
		return
	}
	
	if driver.Role != models.UserRoleLivreur {
		c.JSON(http.StatusBadRequest, gin.H{"error": "L'utilisateur n'est pas un livreur"})
		return
	}
	
	// Récupérer les statistiques détaillées du livreur
	stats, err := userService.GetDriverStats(driverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des statistiques"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Statistiques du livreur récupérées avec succès",
		"driver": driver,
		"stats": stats,
	})
}

func UpdateDriverStatus(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Récupérer l'ID livreur depuis l'URL
	driverID := c.Param("id")
	if driverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID livreur requis"})
		return
	}
	
	// Structure pour recevoir le nouveau statut
	type UpdateDriverStatusRequest struct {
		Status string `json:"status" binding:"required"`
	}
	
	var req UpdateDriverStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}
	
	// Valider le statut
	newStatus := models.DriverStatus(req.Status)
	if !newStatus.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Statut livreur invalide"})
		return
	}
	
	// Vérifier que l'utilisateur existe et est un livreur
	driver, err := userService.GetUserByID(driverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Livreur non trouvé"})
		return
	}
	
	if driver.Role != models.UserRoleLivreur {
		c.JSON(http.StatusBadRequest, gin.H{"error": "L'utilisateur n'est pas un livreur"})
		return
	}
	
	// Mettre à jour le statut du livreur
	err = userService.UpdateDriverStatus(driverID, newStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour du statut"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Statut livreur mis à jour avec succès",
		"driverID": driverID,
		"newStatus": req.Status,
	})
}

func GetAllPromotions(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Paramètres de pagination
	page := 1
	limit := 10
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	// Filtres
	filters := map[string]string{
		"status": c.Query("status"), // active, inactive
		"type":   c.Query("type"),   // PERCENTAGE, FIXED_AMOUNT, FREE_DELIVERY
	}
	
	// Récupérer toutes les promotions avec pagination
	promotions, total, err := promoCodesService.GetAllPromotions(page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des promotions"})
		return
	}
	
	// Calculer le nombre de pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Promotions récupérées avec succès",
		"promotions": promotions,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
		},
		"filters": filters,
	})
}

func CreatePromotion(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Valider les données d'entrée
	var req models.CreatePromoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}
	
	// Validation avec le validateur
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation échouée", "details": err.Error()})
		return
	}
	
	// Créer la promotion
	promotion, err := promoCodesService.CreatePromo(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création de la promotion", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Promotion créée avec succès",
		"promotion": promotion,
	})
}

func UpdatePromotion(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Récupérer l'ID promotion depuis l'URL
	promoID := c.Param("id")
	if promoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID promotion requis"})
		return
	}
	
	// Valider les données d'entrée
	var req models.UpdatePromoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}
	
	// Validation avec le validateur
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation échouée", "details": err.Error()})
		return
	}
	
	// Mettre à jour la promotion
	promotion, err := promoCodesService.UpdatePromo(promoID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour de la promotion", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Promotion mise à jour avec succès",
		"promotion": promotion,
	})
}

func DeletePromotion(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Récupérer l'ID promotion depuis l'URL
	promoID := c.Param("id")
	if promoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID promotion requis"})
		return
	}
	
	// Supprimer la promotion
	err := promoCodesService.DeletePromo(promoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression de la promotion", "details": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Promotion supprimée avec succès",
		"promoID": promoID,
	})
}

func GetPromotionStats(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Récupérer l'ID promotion depuis l'URL (optionnel)
	promoID := c.Param("id")
	
	var stats map[string]interface{}
	var err error
	
	if promoID != "" {
		// Statistiques pour une promotion spécifique
		stats, err = promoCodesService.GetPromoStats(promoID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Promotion non trouvée", "details": err.Error()})
			return
		}
	} else {
		// Statistiques globales des promotions
		stats, err = promoCodesService.GetAllPromotionStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des statistiques"})
			return
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Statistiques des promotions récupérées avec succès",
		"stats": stats,
	})
}

func GetAllVehicles(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Paramètres de pagination
	page := 1
	limit := 10
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	// Filtres
	filters := map[string]string{
		"type":     c.Query("type"),
		"isActive": c.Query("isActive"),
	}
	
	// Récupérer tous les véhicules avec pagination
	vehicles, total, err := vehicleService.GetAllVehicles(page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des véhicules"})
		return
	}
	
	// Calculer le nombre de pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Véhicules récupérés avec succès",
		"vehicles": vehicles,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": totalPages,
		},
		"filters": filters,
	})
}

func VerifyVehicle(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Récupérer l'ID véhicule depuis l'URL
	vehicleID := c.Param("id")
	if vehicleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID véhicule requis"})
		return
	}
	
	// Structure pour recevoir les données de vérification
	type VerifyVehicleRequest struct {
		Verified bool   `json:"verified" binding:"required"`
		Notes    string `json:"notes,omitempty"`
	}
	
	var req VerifyVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}
	
	// Vérifier que le véhicule existe
	vehicle, err := vehicleService.GetVehicleByID(vehicleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Véhicule non trouvé"})
		return
	}
	
	// Mettre à jour le statut de vérification
	err = vehicleService.VerifyVehicle(vehicleID, req.Verified, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la vérification du véhicule"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Véhicule vérifié avec succès",
		"vehicleID": vehicleID,
		"verified": req.Verified,
		"vehicle": vehicle,
	})
}

func GetDashboardStats(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Récupérer les statistiques du tableau de bord
	dashboardStats, err := getDashboardStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des statistiques"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Statistiques du tableau de bord récupérées avec succès",
		"stats": dashboardStats,
	})
}

func GetRevenueStats(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Paramètres de période (optionnels)
	period := c.Query("period") // day, week, month, year
	if period == "" {
		period = "month" // Par défaut
	}
	
	// Récupérer les statistiques de revenus
	revenueStats, err := getRevenueStatistics(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des statistiques de revenus"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Statistiques de revenus récupérées avec succès",
		"stats": revenueStats,
		"period": period,
	})
}

func GetUserStats(c *gin.Context) {
	// Récupérer l'utilisateur authentifié
	userClaims, exists := middlewares.GetCurrentUser(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Utilisateur non authentifié"})
		return
	}
	
	// Vérifier que l'utilisateur est admin
	if userClaims.Role != models.UserRoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès réservé aux administrateurs"})
		return
	}
	
	// Paramètres de période (optionnels)
	period := c.Query("period") // day, week, month, year
	if period == "" {
		period = "month" // Par défaut
	}
	
	// Récupérer les statistiques d'utilisateurs
	userStats, err := getUserStatistics(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des statistiques utilisateur"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Statistiques utilisateur récupérées avec succès",
		"stats": userStats,
		"period": period,
	})
}

// Fonctions utilitaires pour les statistiques Admin

// getDashboardStatistics calcule les statistiques générales du tableau de bord
func getDashboardStatistics() (map[string]interface{}, error) {
	// Statistiques utilisateurs
	allUsers, total, _ := userService.GetAllUsers(1, 1000, "")
	totalUsers := total
	clientCount := 0
	driverCount := 0
	adminCount := 0
	
	for _, user := range allUsers {
		switch user.Role {
		case models.UserRoleClient:
			clientCount++
		case models.UserRoleLivreur:
			driverCount++
		case models.UserRoleAdmin:
			adminCount++
		}
	}
	
	// Statistiques livreurs
	driversOnline := 0
	driversAvailable := 0
	allDrivers, _, _ := userService.GetAllDrivers(1, 1000, "")
	for _, driver := range allDrivers {
		if driver.DriverStatus == models.DriverStatusOnline {
			driversOnline++
		}
		if driver.DriverStatus == models.DriverStatusAvailable {
			driversAvailable++
		}
	}
	
	// Statistiques véhicules
	allVehicles, totalVehicles, _ := vehicleService.GetAllVehicles(1, 1000, map[string]string{})
	activeVehicles := 0
	for _, vehicle := range allVehicles {
		// Vérifier si le véhicule est actif (dépend de votre modèle)
		activeVehicles++
	}
	
	// Statistiques livraisons
	allDeliveries, totalDeliveries, _ := deliveryService.GetAllDeliveries(1, 1000, map[string]string{})
	deliveredCount := 0
	pendingCount := 0
	totalRevenue := 0.0
	
	for _, delivery := range allDeliveries {
		if delivery.Status == models.DeliveryStatusDelivered {
			deliveredCount++
			if delivery.FinalPrice > 0 {
				totalRevenue += delivery.FinalPrice
			}
		} else if delivery.Status == models.DeliveryStatusPending {
			pendingCount++
		}
	}
	
	// Statistiques promotions
	allPromotions, totalPromotions, _ := promoCodesService.GetAllPromotions(1, 1000, map[string]string{})
	activePromotions := 0
	for _, promo := range allPromotions {
		if promo.IsActive {
			activePromotions++
		}
	}
	
	return map[string]interface{}{
		"users": map[string]interface{}{
			"total":   totalUsers,
			"clients": clientCount,
			"drivers": driverCount,
			"admins":  adminCount,
		},
		"drivers": map[string]interface{}{
			"total":     driverCount,
			"online":    driversOnline,
			"available": driversAvailable,
			"offline":   driverCount - driversOnline - driversAvailable,
		},
		"vehicles": map[string]interface{}{
			"total":  totalVehicles,
			"active": activeVehicles,
		},
		"deliveries": map[string]interface{}{
			"total":     totalDeliveries,
			"delivered": deliveredCount,
			"pending":   pendingCount,
			"inProgress": totalDeliveries - deliveredCount - pendingCount,
		},
		"promotions": map[string]interface{}{
			"total":  totalPromotions,
			"active": activePromotions,
		},
		"revenue": map[string]interface{}{
			"total":    totalRevenue,
			"currency": "FCFA",
		},
	}, nil
}

// getRevenueStatistics calcule les statistiques de revenus par période
func getRevenueStatistics(period string) (map[string]interface{}, error) {
	// Récupérer les livraisons livrées
	deliveredDeliveries, _, _ := deliveryService.GetAllDeliveries(1, 10000, map[string]string{
		"status": string(models.DeliveryStatusDelivered),
	})
	
	totalRevenue := 0.0
	deliveryCount := 0
	averageOrder := 0.0
	
	for _, delivery := range deliveredDeliveries {
		if delivery.FinalPrice > 0 {
			totalRevenue += delivery.FinalPrice
			deliveryCount++
		}
	}
	
	if deliveryCount > 0 {
		averageOrder = totalRevenue / float64(deliveryCount)
	}
	
	// TODO: Implémenter le filtrage par période avec des requêtes Prisma appropriées
	// Pour l'instant, retourner les statistiques globales
	
	return map[string]interface{}{
		"period": period,
		"revenue": map[string]interface{}{
			"total":        totalRevenue,
			"deliveryCount": deliveryCount,
			"averageOrder": averageOrder,
			"currency":     "FCFA",
		},
		"growth": map[string]interface{}{
			"percentage": 0.0, // TODO: Calculer la croissance par rapport à la période précédente
			"trend":      "stable",
		},
		"topPerformers": []map[string]interface{}{},
	}, nil
}

// getUserStatistics calcule les statistiques d'utilisateurs par période
func getUserStatistics(period string) (map[string]interface{}, error) {
	// Récupérer tous les utilisateurs
	allUsers, totalUsers, _ := userService.GetAllUsers(1, 10000, "")
	
	// Statistiques par rôle
	clientCount := 0
	driverCount := 0
	adminCount := 0
	activeDrivers := 0
	newUsersThisMonth := 0 // TODO: Implémenter le filtrage par date
	
	for _, user := range allUsers {
		switch user.Role {
		case models.UserRoleClient:
			clientCount++
		case models.UserRoleLivreur:
			driverCount++
			if user.DriverStatus == models.DriverStatusOnline || user.DriverStatus == models.DriverStatusAvailable {
				activeDrivers++
			}
		case models.UserRoleAdmin:
			adminCount++
		}
		
		// TODO: Vérifier si l'utilisateur a été créé ce mois-ci
		// if user.CreatedAt.After(time.Now().AddDate(0, -1, 0)) {
		//     newUsersThisMonth++
		// }
	}
	
	return map[string]interface{}{
		"period": period,
		"total": map[string]interface{}{
			"users":   totalUsers,
			"clients": clientCount,
			"drivers": driverCount,
			"admins":  adminCount,
		},
		"activity": map[string]interface{}{
			"activeDrivers": activeDrivers,
			"newUsers":     newUsersThisMonth,
		},
		"growth": map[string]interface{}{
			"newThisMonth": newUsersThisMonth,
			"percentage":   0.0, // TODO: Calculer la croissance
			"trend":        "stable",
		},
	}, nil
}

// WebSocket handlers (stubs pour plus tard)
func DeliveryWebSocket(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "DeliveryWebSocket - WebSocket not implemented yet"})
}

func DriverNotificationsWebSocket(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "DriverNotificationsWebSocket - WebSocket not implemented yet"})
}

func ClientNotificationsWebSocket(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "ClientNotificationsWebSocket - WebSocket not implemented yet"})
}
