package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Stubs pour tous les handlers référencés dans les routes
// À implémenter progressivement selon les besoins

// Auth handlers
func SendOTP(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "SendOTP - TODO: Implémenter"})
}

func VerifyOTP(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "VerifyOTP - TODO: Implémenter"})
}

func RefreshToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "RefreshToken - TODO: Implémenter"})
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
