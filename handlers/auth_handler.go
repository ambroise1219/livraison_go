package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/ambroise1219/livraison_go/models"
	"github.com/ambroise1219/livraison_go/services"
)

type AuthHandler struct {
	authService *services.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

// SendOTP godoc
// @Summary Send OTP to phone number
// @Description Send OTP code to the specified phone number for authentication
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.SendOTPRequest true "Send OTP Request"
// @Success 200 {object} map[string]interface{} "OTP sent successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/otp/send [post]
func (h *AuthHandler) SendOTP(c *gin.Context) {
	var req models.SendOTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"message": err.Error(),
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": getValidationError(err),
		})
		return
	}

	otp, err := h.authService.SendOTP(req.Phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send OTP",
			"message": err.Error(),
		})
		return
	}

	response := gin.H{
		"message":   "OTP sent successfully",
		"expiresAt": otp.ExpiresAt,
	}

	// Include OTP code in development mode only
	if otp.Code != "" {
		response["code"] = otp.Code
	}

	c.JSON(http.StatusOK, response)
}

// VerifyOTP godoc
// @Summary Verify OTP and authenticate user
// @Description Verify OTP code and return authentication tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login Request"
// @Success 200 {object} models.AuthResponse "Authentication successful"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Invalid OTP"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"message": err.Error(),
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": getValidationError(err),
		})
		return
	}

	authResponse, err := h.authService.VerifyOTP(req.Phone, req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Get new JWT token using refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} models.AuthResponse "Token refreshed successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Invalid refresh token"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"message": err.Error(),
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": getValidationError(err),
		})
		return
	}

	authResponse, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Token refresh failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// Logout godoc
// @Summary Logout user
// @Description Revoke refresh token and logout user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} map[string]interface{} "Logged out successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"message": err.Error(),
		})
		return
	}

	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": getValidationError(err),
		})
		return
	}

	err := h.authService.RevokeRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Logout failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Get the profile of the authenticated user
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserResponse "User profile"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User not found in context",
		})
		return
	}

	// This would typically fetch the user from the database
	// For now, we'll return the user from the JWT claims
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "User context not found",
			"message": "Unable to retrieve user information",
		})
		return
	}

	userModel, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Invalid user context",
			"message": "User information format is invalid",
		})
		return
	}

	c.JSON(http.StatusOK, userModel.ToResponse())
}

// Helper functions

func getUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("userID"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

func getUserRoleFromContext(c *gin.Context) models.UserRole {
	if userRole, exists := c.Get("userRole"); exists {
		if role, ok := userRole.(models.UserRole); ok {
			return role
		}
	}
	return models.UserRoleClient // Default role
}

func getValidationError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		if len(validationErrors) > 0 {
			field := validationErrors[0]
			switch field.Tag() {
			case "required":
				return field.Field() + " is required"
			case "min":
				return field.Field() + " is too short"
			case "max":
				return field.Field() + " is too long"
			case "email":
				return field.Field() + " must be a valid email"
			case "len":
				return field.Field() + " must be exactly " + field.Param() + " characters"
			default:
				return field.Field() + " is invalid"
			}
		}
	}
	return err.Error()
}
