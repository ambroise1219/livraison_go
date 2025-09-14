package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"ilex-backend/models"
	"ilex-backend/services"
)

type DeliveryHandler struct {
	deliveryService *services.DeliveryService
	validator       *validator.Validate
}

func NewDeliveryHandler(deliveryService *services.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{
		deliveryService: deliveryService,
		validator:       validator.New(),
	}
}

// CreateDelivery godoc
// @Summary Create a new delivery
// @Description Create a new delivery order with automatic price calculation
// @Tags Deliveries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateDeliveryRequest true "Create Delivery Request"
// @Success 201 {object} models.DeliveryResponse "Delivery created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Only clients can create deliveries"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /deliveries [post]
func (h *DeliveryHandler) CreateDelivery(c *gin.Context) {
	userID := getUserIDFromContext(c)
	userRole := getUserRoleFromContext(c)

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User authentication required",
		})
		return
	}

	if userRole != models.UserRoleClient {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "Only clients can create deliveries",
		})
		return
	}

	var req models.CreateDeliveryRequest
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

	// Validate delivery type specific requirements
	if err := h.validateDeliveryTypeRequirements(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"message": err.Error(),
		})
		return
	}

	delivery, err := h.deliveryService.CreateDelivery(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create delivery",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, delivery)
}

// AssignDelivery godoc
// @Summary Assign delivery to driver
// @Description Assign a delivery to a specific driver or auto-assign to best available driver
// @Tags Deliveries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.AssignDeliveryRequest true "Assign Delivery Request"
// @Success 200 {object} map[string]interface{} "Delivery assigned successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Only admins can assign deliveries"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /deliveries/assign [post]
func (h *DeliveryHandler) AssignDelivery(c *gin.Context) {
	userRole := getUserRoleFromContext(c)

	if !isAdminRole(userRole) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "Only administrators can assign deliveries",
		})
		return
	}

	var req models.AssignDeliveryRequest
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

	var err error
	if req.DriverID != nil && *req.DriverID != "" {
		// Manual assignment
		err = h.deliveryService.AssignDeliveryToDriver(req.DeliveryID, *req.DriverID)
	} else {
		// Auto-assignment
		err = h.deliveryService.AutoAssignDelivery(req.DeliveryID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to assign delivery",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Delivery assigned successfully",
	})
}

// UpdateDeliveryStatus godoc
// @Summary Update delivery status
// @Description Update the status of a delivery with proper validation based on user role
// @Tags Deliveries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param deliveryId path string true "Delivery ID"
// @Param status query string true "New Status" Enums(ACCEPTED,PICKED_UP,DELIVERED,CANCELLED)
// @Success 200 {object} map[string]interface{} "Status updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - Invalid status transition"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /deliveries/{deliveryId}/status [put]
func (h *DeliveryHandler) UpdateDeliveryStatus(c *gin.Context) {
	userID := getUserIDFromContext(c)
	userRole := getUserRoleFromContext(c)

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User authentication required",
		})
		return
	}

	deliveryID := c.Param("deliveryId")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad request",
			"message": "Delivery ID is required",
		})
		return
	}

	statusStr := c.Query("status")
	if statusStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad request",
			"message": "Status is required",
		})
		return
	}

	status := models.DeliveryStatus(statusStr)
	if !status.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad request",
			"message": "Invalid delivery status",
		})
		return
	}

	err := h.deliveryService.UpdateDeliveryStatus(deliveryID, status, userID, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update delivery status",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Delivery status updated successfully",
		"status":  status,
	})
}

// GetDelivery godoc
// @Summary Get delivery by ID
// @Description Get detailed information about a specific delivery
// @Tags Deliveries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param deliveryId path string true "Delivery ID"
// @Success 200 {object} models.DeliveryResponse "Delivery details"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Delivery not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /deliveries/{deliveryId} [get]
func (h *DeliveryHandler) GetDelivery(c *gin.Context) {
	userID := getUserIDFromContext(c)
	userRole := getUserRoleFromContext(c)

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User authentication required",
		})
		return
	}

	deliveryID := c.Param("deliveryId")
	if deliveryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad request",
			"message": "Delivery ID is required",
		})
		return
	}

	// TODO: Implement GetDeliveryByID in delivery service
	// This is a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"message":    "Delivery details retrieved successfully",
		"deliveryId": deliveryID,
		"userRole":   userRole,
	})
}

// GetDeliveries godoc
// @Summary Get deliveries list
// @Description Get list of deliveries based on user role and filters
// @Tags Deliveries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param status query string false "Filter by delivery status"
// @Param type query string false "Filter by delivery type"
// @Success 200 {object} map[string]interface{} "Deliveries list"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /deliveries [get]
func (h *DeliveryHandler) GetDeliveries(c *gin.Context) {
	userID := getUserIDFromContext(c)
	userRole := getUserRoleFromContext(c)

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User authentication required",
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Parse filter parameters
	statusFilter := c.Query("status")
	typeFilter := c.Query("type")

	// Validate filters
	if statusFilter != "" {
		status := models.DeliveryStatus(statusFilter)
		if !status.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad request",
				"message": "Invalid status filter",
			})
			return
		}
	}

	if typeFilter != "" {
		deliveryType := models.DeliveryType(typeFilter)
		if !deliveryType.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad request",
				"message": "Invalid type filter",
			})
			return
		}
	}

	// TODO: Implement GetDeliveries in delivery service with proper filtering
	// This is a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"message": "Deliveries retrieved successfully",
		"data": gin.H{
			"deliveries": []interface{}{},
			"pagination": gin.H{
				"page":       page,
				"limit":      limit,
				"totalItems": 0,
				"totalPages": 0,
			},
		},
		"filters": gin.H{
			"status": statusFilter,
			"type":   typeFilter,
		},
	})
}

// CalculatePrice godoc
// @Summary Calculate delivery price
// @Description Calculate price for a delivery with optional promo code
// @Tags Deliveries
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param vehicleType query string true "Vehicle Type" Enums(MOTO,VOITURE,CAMIONNETTE)
// @Param distance query float64 true "Distance in kilometers"
// @Param waiting query float64 false "Waiting time in minutes (default: 0)"
// @Param type query string true "Delivery Type" Enums(SIMPLE,EXPRESS,GROUPEE,DEMENAGEMENT)
// @Param promoCode query string false "Promo code for discount"
// @Success 200 {object} models.PriceCalculation "Price calculation result"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /deliveries/calculate-price [get]
func (h *DeliveryHandler) CalculatePrice(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "User authentication required",
		})
		return
	}

	// Parse query parameters
	vehicleTypeStr := c.Query("vehicleType")
	distanceStr := c.Query("distance")
	waitingStr := c.DefaultQuery("waiting", "0")
	deliveryTypeStr := c.Query("type")
	promoCode := c.Query("promoCode")

	if vehicleTypeStr == "" || distanceStr == "" || deliveryTypeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad request",
			"message": "vehicleType, distance, and type are required",
		})
		return
	}

	// Parse and validate parameters
	vehicleType := models.VehicleType(vehicleTypeStr)
	if !vehicleType.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad request",
			"message": "Invalid vehicle type",
		})
		return
	}

	deliveryType := models.DeliveryType(deliveryTypeStr)
	if !deliveryType.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad request",
			"message": "Invalid delivery type",
		})
		return
	}

	distance, err := strconv.ParseFloat(distanceStr, 64)
	if err != nil || distance <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad request",
			"message": "Invalid distance value",
		})
		return
	}

	waiting, err := strconv.ParseFloat(waitingStr, 64)
	if err != nil || waiting < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad request",
			"message": "Invalid waiting time value",
		})
		return
	}

	var promoCodePtr *string
	if promoCode != "" {
		promoCodePtr = &promoCode
	}

	// Calculate price
	calculation, err := h.deliveryService.CalculateDeliveryPriceWithPromo(
		vehicleType, distance, waiting, deliveryType, promoCodePtr,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to calculate price",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, calculation)
}

// Helper functions

func (h *DeliveryHandler) validateDeliveryTypeRequirements(req *models.CreateDeliveryRequest) error {
	switch req.Type {
	case models.DeliveryTypeDemenagement:
		if req.MovingInfo == nil {
			return fmt.Errorf("moving information is required for moving deliveries")
		}
		// Only allow camionnette for moving
		if req.VehicleType != models.VehicleTypeCamionnette {
			return fmt.Errorf("moving deliveries require a camionnette")
		}
	case models.DeliveryTypeGroupee:
		if req.GroupedInfo == nil {
			return fmt.Errorf("grouped information is required for grouped deliveries")
		}
		if len(req.GroupedInfo.Zones) < 2 {
			return fmt.Errorf("grouped deliveries require at least 2 zones")
		}
	}
	return nil
}

func isAdminRole(role models.UserRole) bool {
	return role == models.UserRoleAdmin || role == models.UserRoleGestionnaire
}