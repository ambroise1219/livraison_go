package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/models"
	"github.com/ambroise1219/livraison_go/services"
)

// Global validator instance
var validate *validator.Validate
var authService *services.AuthService

// InitHandlers initializes handlers with dependencies
func InitHandlers() {
	validate = validator.New()
	authService = services.NewAuthService(config.GetConfig())
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

	// Save OTP to database
	otp, err := authService.SaveOTP(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP", "details": err.Error()})
		return
	}

	// Simulate SMS sending
	if err := authService.SimulateSMSSend(req.Phone, otp.Code); err != nil {
		// Don't fail if SMS sending fails, just log it
		// In production, you might want to handle this differently
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully",
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

	// Verify OTP
	_, err := authService.VerifyOTP(req.Phone, req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP", "details": err.Error()})
		return
	}

	// Find or create user
	user, isNewUser, err := authService.FindOrCreateUser(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user", "details": err.Error()})
		return
	}

	// Generate tokens
	authResponse, err := authService.GenerateTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens", "details": err.Error()})
		return
	}

	response := gin.H{
		"message": "Authentication successful",
		"isNewUser": isNewUser,
		"token": authResponse.Token,
		"refreshToken": authResponse.RefreshToken,
		"user": authResponse.User,
		"expiresAt": authResponse.ExpiresAt,
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

	// Refresh access token
	authResponse, err := authService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"token": authResponse.Token,
		"refreshToken": authResponse.RefreshToken,
		"user": authResponse.User,
		"expiresAt": authResponse.ExpiresAt,
	})
}

func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logout - TODO: Implémenter"})
}

func GetProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetProfile - TODO: Implémenter"})
}

func UpdateProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateProfile - TODO: Implémenter"})
}

// User handlers
func GetUserProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetUserProfile - TODO: Implémenter"})
}

func UpdateUserProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateUserProfile - TODO: Implémenter"})
}

func GetUserDeliveries(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetUserDeliveries - TODO: Implémenter"})
}

func GetUserVehicles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetUserVehicles - TODO: Implémenter"})
}

func CreateVehicle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateVehicle - TODO: Implémenter"})
}

func UpdateVehicle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateVehicle - TODO: Implémenter"})
}

// Delivery handlers
func CreateDelivery(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateDelivery - TODO: Implémenter"})
}

func GetDelivery(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetDelivery - TODO: Implémenter"})
}

func UpdateDeliveryStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateDeliveryStatus - TODO: Implémenter"})
}

func AssignDelivery(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "AssignDelivery - TODO: Implémenter"})
}

func CalculateDeliveryPrice(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CalculateDeliveryPrice - TODO: Implémenter"})
}

func GetAvailableDeliveries(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAvailableDeliveries - TODO: Implémenter"})
}

func GetAssignedDeliveries(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAssignedDeliveries - TODO: Implémenter"})
}

func AcceptDelivery(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "AcceptDelivery - TODO: Implémenter"})
}

func UpdateDriverLocation(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateDriverLocation - TODO: Implémenter"})
}

func GetClientDeliveries(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetClientDeliveries - TODO: Implémenter"})
}

func CancelDelivery(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CancelDelivery - TODO: Implémenter"})
}

func TrackDelivery(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "TrackDelivery - TODO: Implémenter"})
}

// Promo handlers
func ValidatePromoCode(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ValidatePromoCode - TODO: Implémenter"})
}

func UsePromoCode(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UsePromoCode - TODO: Implémenter"})
}

func GetPromoHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetPromoHistory - TODO: Implémenter"})
}

func CreateReferral(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreateReferral - TODO: Implémenter"})
}

func GetReferralStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetReferralStats - TODO: Implémenter"})
}

// Admin handlers
func GetAllUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllUsers - TODO: Implémenter"})
}

func GetUserDetails(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetUserDetails - TODO: Implémenter"})
}

func UpdateUserRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateUserRole - TODO: Implémenter"})
}

func DeleteUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "DeleteUser - TODO: Implémenter"})
}

func GetAllDeliveries(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllDeliveries - TODO: Implémenter"})
}

func GetDeliveryStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetDeliveryStats - TODO: Implémenter"})
}

func ForceAssignDelivery(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ForceAssignDelivery - TODO: Implémenter"})
}

func GetAllDrivers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllDrivers - TODO: Implémenter"})
}

func GetDriverStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetDriverStats - TODO: Implémenter"})
}

func UpdateDriverStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdateDriverStatus - TODO: Implémenter"})
}

func GetAllPromotions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllPromotions - TODO: Implémenter"})
}

func CreatePromotion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "CreatePromotion - TODO: Implémenter"})
}

func UpdatePromotion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "UpdatePromotion - TODO: Implémenter"})
}

func DeletePromotion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "DeletePromotion - TODO: Implémenter"})
}

func GetPromotionStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetPromotionStats - TODO: Implémenter"})
}

func GetAllVehicles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetAllVehicles - TODO: Implémenter"})
}

func VerifyVehicle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "VerifyVehicle - TODO: Implémenter"})
}

func GetDashboardStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetDashboardStats - TODO: Implémenter"})
}

func GetRevenueStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetRevenueStats - TODO: Implémenter"})
}

func GetUserStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "GetUserStats - TODO: Implémenter"})
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
