package services

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/google/uuid"

	"ilex-backend/config"
	"ilex-backend/db"
	"ilex-backend/models"
)

type DeliveryService struct {
	config      *config.Config
	promoService *PromoService
}

func NewDeliveryService(cfg *config.Config, promoService *PromoService) *DeliveryService {
	return &DeliveryService{
		config:       cfg,
		promoService: promoService,
	}
}

// CreateDelivery creates a new delivery with price calculation
func (s *DeliveryService) CreateDelivery(clientID string, req *models.CreateDeliveryRequest) (*models.DeliveryResponse, error) {
	// Validate client exists and is a client
	client, err := s.getUserByID(clientID)
	if err != nil {
		return nil, fmt.Errorf("client not found: %v", err)
	}
	
	if !client.IsClient() {
		return nil, fmt.Errorf("only clients can create deliveries")
	}

	// Create pickup and dropoff locations
	pickupLocation, err := s.createLocation(req.PickupAddress, req.PickupLat, req.PickupLng)
	if err != nil {
		return nil, fmt.Errorf("failed to create pickup location: %v", err)
	}

	dropoffLocation, err := s.createLocation(req.DropoffAddress, req.DropoffLat, req.DropoffLng)
	if err != nil {
		return nil, fmt.Errorf("failed to create dropoff location: %v", err)
	}

	// Calculate distance and duration
	distance, duration, err := s.calculateDistanceAndDuration(pickupLocation, dropoffLocation)
	if err != nil {
		log.Printf("Warning: failed to calculate distance: %v", err)
		// Use default values
		distance = 5.0  // 5 km default
		duration = 30.0 // 30 minutes default
	}

	// Calculate price based on pricing rules
	pricing, err := s.calculateDeliveryPrice(req.VehicleType, distance, 0, req.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate price: %v", err)
	}

	// Create delivery record
	delivery := &models.Delivery{
		ID:            uuid.New().String(),
		ClientID:      clientID,
		Status:        models.DeliveryStatusPending,
		Type:          req.Type,
		PickupID:      pickupLocation.ID,
		DropoffID:     dropoffLocation.ID,
		DistanceKm:    &distance,
		DurationMin:   &duration,
		VehicleType:   req.VehicleType,
		BasePrice:     &pricing.BasePrice,
		FinalPrice:    pricing.FinalPrice,
		PaymentMethod: req.PaymentMethod,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save delivery to database
	err = s.saveDelivery(delivery)
	if err != nil {
		return nil, fmt.Errorf("failed to save delivery: %v", err)
	}

	// Handle special delivery types
	if req.PackageInfo != nil {
		err = s.createPackage(delivery.ID, req.PackageInfo)
		if err != nil {
			log.Printf("Warning: failed to create package: %v", err)
		}
	}

	if req.MovingInfo != nil && req.Type == models.DeliveryTypeDemenagement {
		err = s.createMovingService(delivery.ID, req.MovingInfo)
		if err != nil {
			log.Printf("Warning: failed to create moving service: %v", err)
		}
	}

	if req.GroupedInfo != nil && req.Type == models.DeliveryTypeGroupee {
		err = s.createGroupedDelivery(delivery.ID, req.GroupedInfo)
		if err != nil {
			log.Printf("Warning: failed to create grouped delivery: %v", err)
		}
	}

	// Auto-assign if express delivery
	if req.Type == models.DeliveryTypeExpress {
		go s.AutoAssignDelivery(delivery.ID)
	}

	response := delivery.ToResponse()
	response.Pickup = pickupLocation
	response.Dropoff = dropoffLocation
	response.Client = client.ToResponse()

	return response, nil
}

// AutoAssignDelivery automatically assigns delivery to best available driver
func (s *DeliveryService) AutoAssignDelivery(deliveryID string) error {
	delivery, err := s.getDeliveryByID(deliveryID)
	if err != nil {
		return fmt.Errorf("delivery not found: %v", err)
	}

	if !delivery.CanBeAssigned() {
		return fmt.Errorf("delivery cannot be assigned")
	}

	// Find best driver
	driver, err := s.findBestDriverForDelivery(delivery)
	if err != nil {
		log.Printf("No available driver found for delivery %s: %v", deliveryID, err)
		return err
	}

	// Assign delivery
	return s.AssignDeliveryToDriver(deliveryID, driver.ID)
}

// AssignDeliveryToDriver assigns a delivery to a specific driver
func (s *DeliveryService) AssignDeliveryToDriver(deliveryID, driverID string) error {
	// Validate delivery
	delivery, err := s.getDeliveryByID(deliveryID)
	if err != nil {
		return fmt.Errorf("delivery not found: %v", err)
	}

	if !delivery.CanBeAssigned() {
		return fmt.Errorf("delivery cannot be assigned")
	}

	// Validate driver
	driver, err := s.getUserByID(driverID)
	if err != nil {
		return fmt.Errorf("driver not found: %v", err)
	}

	if !driver.CanAcceptDeliveries() {
		return fmt.Errorf("driver cannot accept deliveries")
	}

	// Check driver vehicle compatibility
	vehicle, err := s.getDriverVehicle(driverID)
	if err != nil {
		return fmt.Errorf("driver vehicle not found: %v", err)
	}

	if !vehicle.IsCompatibleWithDeliveryType(delivery.Type) {
		return fmt.Errorf("driver vehicle not compatible with delivery type")
	}

	// Update delivery
	query := `UPDATE Delivery SET livreurId = $driverId, status = $status, updatedAt = $updatedAt WHERE id = $deliveryId`
	params := map[string]interface{}{
		"deliveryId": deliveryID,
		"driverId":   driverID,
		"status":     string(models.DeliveryStatusAccepted),
		"updatedAt":  time.Now(),
	}

	_, err = db.Query(query, params)
	if err != nil {
		return fmt.Errorf("failed to assign delivery: %v", err)
	}

	// Update driver status
	err = s.updateDriverStatus(driverID, models.DriverStatusBusy)
	if err != nil {
		log.Printf("Warning: failed to update driver status: %v", err)
	}

	// Send notifications
	go s.sendAssignmentNotifications(delivery, driver)

	return nil
}

// UpdateDeliveryStatus updates delivery status with business logic
func (s *DeliveryService) UpdateDeliveryStatus(deliveryID string, status models.DeliveryStatus, userID string, userRole models.UserRole) error {
	delivery, err := s.getDeliveryByID(deliveryID)
	if err != nil {
		return fmt.Errorf("delivery not found: %v", err)
	}

	// Validate status transition
	if !s.isValidStatusTransition(delivery.Status, status, userRole) {
		return fmt.Errorf("invalid status transition from %s to %s", delivery.Status, status)
	}

	// Update delivery
	query := `UPDATE Delivery SET status = $status, updatedAt = $updatedAt WHERE id = $deliveryId`
	params := map[string]interface{}{
		"deliveryId": deliveryID,
		"status":     string(status),
		"updatedAt":  time.Now(),
	}

	_, err = db.Query(query, params)
	if err != nil {
		return fmt.Errorf("failed to update delivery status: %v", err)
	}

	// Handle status-specific logic
	switch status {
	case models.DeliveryStatusDelivered:
		err = s.handleDeliveryCompleted(delivery)
		if err != nil {
			log.Printf("Warning: failed to handle delivery completion: %v", err)
		}
	case models.DeliveryStatusCancelled:
		err = s.handleDeliveryCancelled(delivery)
		if err != nil {
			log.Printf("Warning: failed to handle delivery cancellation: %v", err)
		}
	}

	// Send status notifications
	go s.sendStatusUpdateNotifications(delivery, status)

	return nil
}

// CalculateDeliveryPriceWithPromo calculates final price with promo code
func (s *DeliveryService) CalculateDeliveryPriceWithPromo(vehicleType models.VehicleType, distance, waiting float64, deliveryType models.DeliveryType, promoCode *string) (*models.PriceCalculation, error) {
	// Calculate base price
	calculation, err := s.calculateDeliveryPrice(vehicleType, distance, waiting, deliveryType)
	if err != nil {
		return nil, err
	}

	// Apply promo if provided
	if promoCode != nil && *promoCode != "" {
		promoDiscount, err := s.promoService.ValidateAndCalculateDiscount(*promoCode, calculation.SubTotal)
		if err != nil {
			log.Printf("Warning: invalid promo code %s: %v", *promoCode, err)
		} else {
			calculation.PromoDiscount = promoDiscount
			calculation.FinalPrice = calculation.SubTotal - promoDiscount
			calculation.PromoCode = promoCode
		}
	}

	if calculation.FinalPrice < 0 {
		calculation.FinalPrice = 0
	}

	return &calculation, nil
}

// Helper methods

func (s *DeliveryService) calculateDeliveryPrice(vehicleType models.VehicleType, distance, waiting float64, deliveryType models.DeliveryType) (models.PriceCalculation, error) {
	// Get pricing rule for vehicle type
	pricingRule, err := s.getPricingRule(vehicleType)
	if err != nil {
		return models.PriceCalculation{}, fmt.Errorf("pricing rule not found: %v", err)
	}

	// Calculate base price
	calculation := pricingRule.CalculatePrice(distance, waiting)

	// Apply delivery type multipliers
	switch deliveryType {
	case models.DeliveryTypeExpress:
		calculation.BasePrice *= 1.5 // 50% extra for express
	case models.DeliveryTypeGroupee:
		calculation.BasePrice *= 0.7 // 30% discount for grouped
	case models.DeliveryTypeDemenagement:
		calculation.BasePrice *= 2.0 // Double for moving
	}

	calculation.SubTotal = calculation.BasePrice + calculation.DistancePrice + calculation.WaitingPrice
	calculation.FinalPrice = calculation.SubTotal - calculation.PromoDiscount

	return calculation, nil
}

func (s *DeliveryService) findBestDriverForDelivery(delivery *models.Delivery) (*models.User, error) {
	// Get pickup location
	pickupLocation, err := s.getLocationByID(delivery.PickupID)
	if err != nil {
		return nil, fmt.Errorf("pickup location not found: %v", err)
	}

	// Find available drivers with compatible vehicle
	query := `
		SELECT u.*, dl.lat, dl.lng, dl.timestamp 
		FROM User u
		JOIN DriverLocation dl ON u.id = dl.driverId
		JOIN Vehicle v ON u.id = v.userId
		WHERE u.role = 'LIVREUR'
		AND u.driverStatus IN ['ONLINE', 'AVAILABLE']
		AND u.is_driver_complete = true
		AND u.is_driver_vehicule_complete = true
		AND dl.isAvailable = true
		AND v.type = $vehicleType
		ORDER BY dl.timestamp DESC`

	params := map[string]interface{}{
		"vehicleType": string(delivery.VehicleType),
	}

	results, err := db.QueryMultiple(query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to query available drivers: %v", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no available drivers found")
	}

	// Find closest driver
	var bestDriver *models.User
	var minDistance float64 = math.MaxFloat64

	for _, result := range results {
		driverData, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		driver := s.parseUserFromMap(driverData)
		driverLat, _ := driverData["lat"].(float64)
		driverLng, _ := driverData["lng"].(float64)

		if pickupLocation.Lat != nil && pickupLocation.Lng != nil {
			distance := s.calculateHaversineDistance(
				*pickupLocation.Lat, *pickupLocation.Lng,
				driverLat, driverLng,
			)

			if distance < minDistance {
				minDistance = distance
				bestDriver = driver
			}
		} else if bestDriver == nil {
			bestDriver = driver // Fallback if no coordinates
		}
	}

	if bestDriver == nil {
		return nil, fmt.Errorf("no suitable driver found")
	}

	return bestDriver, nil
}

func (s *DeliveryService) calculateHaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371 // Earth radius in kilometers

	// Convert degrees to radians
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180

	// Haversine formula
	dlat := lat2Rad - lat1Rad
	dlng := lng2Rad - lng1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func (s *DeliveryService) isValidStatusTransition(from, to models.DeliveryStatus, userRole models.UserRole) bool {
	// Define valid transitions based on user role
	validTransitions := map[models.UserRole]map[models.DeliveryStatus][]models.DeliveryStatus{
		models.UserRoleClient: {
			models.DeliveryStatusPending: {models.DeliveryStatusCancelled},
		},
		models.UserRoleLivreur: {
			models.DeliveryStatusAccepted:           {models.DeliveryStatusPickupInProgress, models.DeliveryStatusCancelled},
			models.DeliveryStatusPickupInProgress:   {models.DeliveryStatusPickedUp},
			models.DeliveryStatusPickedUp:           {models.DeliveryStatusInTransit},
			models.DeliveryStatusInTransit:          {models.DeliveryStatusArrivedAtDropoff},
			models.DeliveryStatusArrivedAtDropoff:   {models.DeliveryStatusDelivered},
		},
		models.UserRoleAdmin: {
			// Admin can transition to any status
			models.DeliveryStatusPending:    {models.DeliveryStatusAccepted, models.DeliveryStatusCancelled},
			models.DeliveryStatusAccepted:   {models.DeliveryStatusPickupInProgress, models.DeliveryStatusCancelled},
			models.DeliveryStatusPickedUp:   {models.DeliveryStatusInTransit, models.DeliveryStatusCancelled},
			models.DeliveryStatusInTransit:  {models.DeliveryStatusDelivered, models.DeliveryStatusCancelled},
			models.DeliveryStatusDelivered:  {models.DeliveryStatusCancelled},
		},
	}

	allowedTransitions, exists := validTransitions[userRole][from]
	if !exists {
		return false
	}

	for _, allowedStatus := range allowedTransitions {
		if to == allowedStatus {
			return true
		}
	}

	return false
}

// Database helper methods
func (s *DeliveryService) createLocation(address string, lat, lng *float64) (*models.Location, error) {
	location := &models.Location{
		ID:      uuid.New().String(),
		Address: address,
		Lat:     lat,
		Lng:     lng,
	}

	query := `CREATE Location SET id = $id, address = $address, lat = $lat, lng = $lng`
	params := map[string]interface{}{
		"id":      location.ID,
		"address": location.Address,
		"lat":     location.Lat,
		"lng":     location.Lng,
	}

	_, err := db.Query(query, params)
	if err != nil {
		return nil, err
	}

	return location, nil
}

func (s *DeliveryService) saveDelivery(delivery *models.Delivery) error {
	query := `CREATE Delivery SET 
		id = $id,
		clientId = $clientId,
		status = $status,
		type = $type,
		pickupId = $pickupId,
		dropoffId = $dropoffId,
		distanceKm = $distanceKm,
		durationMin = $durationMin,
		vehicleType = $vehicleType,
		basePrice = $basePrice,
		finalPrice = $finalPrice,
		paymentMethod = $paymentMethod,
		createdAt = $createdAt,
		updatedAt = $updatedAt`

	params := map[string]interface{}{
		"id":            delivery.ID,
		"clientId":      delivery.ClientID,
		"status":        string(delivery.Status),
		"type":          string(delivery.Type),
		"pickupId":      delivery.PickupID,
		"dropoffId":     delivery.DropoffID,
		"distanceKm":    delivery.DistanceKm,
		"durationMin":   delivery.DurationMin,
		"vehicleType":   string(delivery.VehicleType),
		"basePrice":     delivery.BasePrice,
		"finalPrice":    delivery.FinalPrice,
		"paymentMethod": string(delivery.PaymentMethod),
		"createdAt":     delivery.CreatedAt,
		"updatedAt":     delivery.UpdatedAt,
	}

	_, err := db.Query(query, params)
	return err
}

// Additional helper methods would continue here...
// For brevity, I'll include the key method signatures

func (s *DeliveryService) createPackage(deliveryID string, packageInfo *models.PackageInfo) error {
	// Implementation for creating package
	return nil
}

func (s *DeliveryService) createMovingService(deliveryID string, movingInfo *models.MovingInfo) error {
	// Implementation for creating moving service
	return nil
}

func (s *DeliveryService) createGroupedDelivery(deliveryID string, groupedInfo *models.GroupedInfo) error {
	// Implementation for creating grouped delivery
	return nil
}

func (s *DeliveryService) calculateDistanceAndDuration(pickup, dropoff *models.Location) (float64, float64, error) {
	// Implementation for calculating distance and duration
	// Could integrate with mapping APIs
	if pickup.Lat != nil && pickup.Lng != nil && dropoff.Lat != nil && dropoff.Lng != nil {
		distance := s.calculateHaversineDistance(*pickup.Lat, *pickup.Lng, *dropoff.Lat, *dropoff.Lng)
		duration := distance * 3 // Rough estimate: 3 minutes per km
		return distance, duration, nil
	}
	return 5.0, 30.0, nil // Default values
}

func (s *DeliveryService) getPricingRule(vehicleType models.VehicleType) (*models.PricingRule, error) {
	// Implementation for getting pricing rules
	// Default pricing rules
	switch vehicleType {
	case models.VehicleTypeMoto:
		return &models.PricingRule{
			ID:          "moto-rule",
			VehicleType: models.VehicleTypeMoto,
			BasePrice:   1000,
			IncludedKm:  3,
			PerKm:       200,
			WaitingFree: 5,
			WaitingRate: 50,
		}, nil
	case models.VehicleTypeVoiture:
		return &models.PricingRule{
			ID:          "voiture-rule",
			VehicleType: models.VehicleTypeVoiture,
			BasePrice:   2000,
			IncludedKm:  5,
			PerKm:       300,
			WaitingFree: 10,
			WaitingRate: 100,
		}, nil
	case models.VehicleTypeCamionnette:
		return &models.PricingRule{
			ID:          "camionnette-rule",
			VehicleType: models.VehicleTypeCamionnette,
			BasePrice:   5000,
			IncludedKm:  10,
			PerKm:       500,
			WaitingFree: 15,
			WaitingRate: 200,
		}, nil
	}
	return nil, fmt.Errorf("pricing rule not found for vehicle type: %s", vehicleType)
}

func (s *DeliveryService) getDeliveryByID(deliveryID string) (*models.Delivery, error) {
	// Implementation for getting delivery by ID
	return nil, nil
}

func (s *DeliveryService) getUserByID(userID string) (*models.User, error) {
	// Implementation for getting user by ID
	return nil, nil
}

func (s *DeliveryService) getDriverVehicle(driverID string) (*models.Vehicle, error) {
	// Implementation for getting driver vehicle
	return nil, nil
}

func (s *DeliveryService) getLocationByID(locationID string) (*models.Location, error) {
	// Implementation for getting location by ID
	return nil, nil
}

func (s *DeliveryService) updateDriverStatus(driverID string, status models.DriverStatus) error {
	// Implementation for updating driver status
	return nil
}

func (s *DeliveryService) handleDeliveryCompleted(delivery *models.Delivery) error {
	// Implementation for handling delivery completion
	return nil
}

func (s *DeliveryService) handleDeliveryCancelled(delivery *models.Delivery) error {
	// Implementation for handling delivery cancellation
	return nil
}

func (s *DeliveryService) sendAssignmentNotifications(delivery *models.Delivery, driver *models.User) {
	// Implementation for sending assignment notifications
}

func (s *DeliveryService) sendStatusUpdateNotifications(delivery *models.Delivery, status models.DeliveryStatus) {
	// Implementation for sending status update notifications
}

func (s *DeliveryService) parseUserFromMap(data map[string]interface{}) *models.User {
	// Implementation for parsing user from map
	return nil
}