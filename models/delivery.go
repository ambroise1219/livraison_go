package models

import (
	"time"
)

// DeliveryStatus defines the delivery status enumeration
type DeliveryStatus string

const (
	DeliveryStatusPending              DeliveryStatus = "PENDING"
	DeliveryStatusAccepted             DeliveryStatus = "ACCEPTED"
	DeliveryStatusPickedUp             DeliveryStatus = "PICKED_UP"
	DeliveryStatusDelivered            DeliveryStatus = "DELIVERED"
	DeliveryStatusCancelled            DeliveryStatus = "CANCELLED"
	DeliveryStatusZoneAssigned         DeliveryStatus = "ZONE_ASSIGNED"
	DeliveryStatusPickupInProgress     DeliveryStatus = "PICKUP_IN_PROGRESS"
	DeliveryStatusPickupCompleted      DeliveryStatus = "PICKUP_COMPLETED"
	DeliveryStatusDeliveryInProgress   DeliveryStatus = "DELIVERY_IN_PROGRESS"
	DeliveryStatusAssignedToHelper     DeliveryStatus = "ASSIGNED_TO_HELPER"
	DeliveryStatusHelpersConfirmed     DeliveryStatus = "HELPERS_CONFIRMED"
	DeliveryStatusArrivedAtPickup      DeliveryStatus = "ARRIVED_AT_PICKUP"
	DeliveryStatusLoadingInProgress    DeliveryStatus = "LOADING_IN_PROGRESS"
	DeliveryStatusLoadingCompleted     DeliveryStatus = "LOADING_COMPLETED"
	DeliveryStatusInTransit            DeliveryStatus = "IN_TRANSIT"
	DeliveryStatusArrivedAtDestination DeliveryStatus = "ARRIVED_AT_DESTINATION"
	DeliveryStatusUnloadingInProgress  DeliveryStatus = "UNLOADING_IN_PROGRESS"
	DeliveryStatusUnloadingCompleted   DeliveryStatus = "UNLOADING_COMPLETED"
	DeliveryStatusArrivedAtDropoff     DeliveryStatus = "ARRIVED_AT_DROPOFF"
	DeliveryStatusEnRoute              DeliveryStatus = "EN_ROUTE"
	DeliveryStatusDispatchInProgress   DeliveryStatus = "DISPATCH_IN_PROGRESS"
	DeliveryStatusSorted               DeliveryStatus = "SORTED"
	DeliveryStatusSortingInProgress    DeliveryStatus = "SORTING_IN_PROGRESS"
)

// DeliveryType defines the delivery type enumeration
type DeliveryType string

const (
	DeliveryTypeSimple      DeliveryType = "SIMPLE"
	DeliveryTypeExpress     DeliveryType = "EXPRESS"
	DeliveryTypeGroupee     DeliveryType = "GROUPEE"
	DeliveryTypeDemenagement DeliveryType = "DEMENAGEMENT"
)

// VehicleType defines the vehicle type enumeration
type VehicleType string

const (
	VehicleTypeMoto        VehicleType = "MOTO"
	VehicleTypeVoiture     VehicleType = "VOITURE"
	VehicleTypeCamionnette VehicleType = "CAMIONNETTE"
)

// PaymentMethod defines the payment method enumeration
type PaymentMethod string

const (
	PaymentMethodCash              PaymentMethod = "CASH"
	PaymentMethodMobileMoneyOrange PaymentMethod = "MOBILE_MONEY_ORANGE"
	PaymentMethodMobileMoneyMTN    PaymentMethod = "MOBILE_MONEY_MTN"
	PaymentMethodMobileMoneyMoov   PaymentMethod = "MOBILE_MONEY_MOOV"
	PaymentMethodMobileMoneyWave   PaymentMethod = "MOBILE_MONEY_WAVE"
)

// Location represents a location
type Location struct {
	ID      string   `json:"id"`
	Address string   `json:"address" validate:"required"`
	Lat     *float64 `json:"lat,omitempty" validate:"omitempty,gte=-90,lte=90"`
	Lng     *float64 `json:"lng,omitempty" validate:"omitempty,gte=-180,lte=180"`
}

// Delivery represents a delivery
type Delivery struct {
	ID            string         `json:"id"`
	ClientID      string         `json:"clientId" validate:"required"`
	LivreurID     *string        `json:"livreurId,omitempty"`
	Status        DeliveryStatus `json:"status"`
	Type          DeliveryType   `json:"type" validate:"required"`
	PickupID      string         `json:"pickupId" validate:"required"`
	DropoffID     string         `json:"dropoffId" validate:"required"`
	DistanceKm    *float64       `json:"distanceKm,omitempty"`
	DurationMin   *float64       `json:"durationMin,omitempty"`
	VehicleType   VehicleType    `json:"vehicleType" validate:"required"`
	BasePrice     *float64       `json:"basePrice,omitempty"`
	WaitingMin    *float64       `json:"waitingMin,omitempty"`
	FinalPrice    float64        `json:"finalPrice" validate:"gte=0"`
	PaymentMethod PaymentMethod  `json:"paymentMethod" validate:"required"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	PaidAt        *time.Time     `json:"paidAt,omitempty"`
}

// CreateDeliveryRequest represents request for creating a delivery
type CreateDeliveryRequest struct {
	Type          DeliveryType  `json:"type" validate:"required"`
	PickupAddress string        `json:"pickupAddress" validate:"required"`
	PickupLat     *float64      `json:"pickupLat,omitempty" validate:"omitempty,gte=-90,lte=90"`
	PickupLng     *float64      `json:"pickupLng,omitempty" validate:"omitempty,gte=-180,lte=180"`
	DropoffAddress string       `json:"dropoffAddress" validate:"required"`
	DropoffLat    *float64      `json:"dropoffLat,omitempty" validate:"omitempty,gte=-90,lte=90"`
	DropoffLng    *float64      `json:"dropoffLng,omitempty" validate:"omitempty,gte=-180,lte=180"`
	VehicleType   VehicleType   `json:"vehicleType" validate:"required"`
	PaymentMethod PaymentMethod `json:"paymentMethod" validate:"required"`
	PackageInfo   *PackageInfo  `json:"packageInfo,omitempty"`
	MovingInfo    *MovingInfo   `json:"movingInfo,omitempty"`
	GroupedInfo   *GroupedInfo  `json:"groupedInfo,omitempty"`
}

// UpdateDeliveryRequest represents request for updating a delivery
type UpdateDeliveryRequest struct {
	Status      *DeliveryStatus `json:"status,omitempty"`
	LivreurID   *string         `json:"livreurId,omitempty"`
	WaitingMin  *float64        `json:"waitingMin,omitempty"`
	FinalPrice  *float64        `json:"finalPrice,omitempty" validate:"omitempty,gte=0"`
}

// AssignDeliveryRequest represents request for assigning delivery to driver
type AssignDeliveryRequest struct {
	DeliveryID string  `json:"deliveryId" validate:"required"`
	DriverID   *string `json:"driverId,omitempty"` // If empty, auto-assign
}

// Package represents a package
type Package struct {
	ID          string   `json:"id"`
	DeliveryID  string   `json:"deliveryId" validate:"required"`
	Description *string  `json:"description,omitempty"`
	WeightKg    *float64 `json:"weightKg,omitempty" validate:"omitempty,gte=0"`
	Size        *string  `json:"size,omitempty"`
	Fragile     bool     `json:"fragile"`
}

// PackageInfo for creating deliveries
type PackageInfo struct {
	Description *string  `json:"description,omitempty"`
	WeightKg    *float64 `json:"weightKg,omitempty" validate:"omitempty,gte=0"`
	Size        *string  `json:"size,omitempty"`
	Fragile     bool     `json:"fragile"`
}

// MovingInfo for moving deliveries
type MovingInfo struct {
	VehicleSize       string   `json:"vehicleSize" validate:"required"`
	HelpersCount      int      `json:"helpersCount" validate:"gte=1"`
	Floors            int      `json:"floors" validate:"gte=1"`
	HasElevator       bool     `json:"hasElevator"`
	NeedsDisassembly  bool     `json:"needsDisassembly"`
	HasFragileItems   bool     `json:"hasFragileItems"`
	AdditionalServices []string `json:"additionalServices,omitempty"`
	SpecialInstructions *string `json:"specialInstructions,omitempty"`
	EstimatedVolume   *float64 `json:"estimatedVolume,omitempty"`
}

// GroupedInfo for grouped deliveries
type GroupedInfo struct {
	Zones []GroupedZone `json:"zones" validate:"min=2,dive"`
}

// GroupedZone represents a zone in grouped delivery
type GroupedZone struct {
	ZoneNumber       int     `json:"zoneNumber" validate:"gte=1"`
	RecipientName    string  `json:"recipientName" validate:"required"`
	RecipientPhone   string  `json:"recipientPhone" validate:"required,min=8,max=15"`
	PickupAddress    string  `json:"pickupAddress" validate:"required"`
	PickupLat        *float64 `json:"pickupLat,omitempty"`
	PickupLng        *float64 `json:"pickupLng,omitempty"`
	DeliveryAddress  string  `json:"deliveryAddress" validate:"required"`
	DeliveryLat      *float64 `json:"deliveryLat,omitempty"`
	DeliveryLng      *float64 `json:"deliveryLng,omitempty"`
}

// DeliveryResponse represents delivery data in response
type DeliveryResponse struct {
	ID            string         `json:"id"`
	ClientID      string         `json:"clientId"`
	Client        *UserResponse  `json:"client,omitempty"`
	LivreurID     *string        `json:"livreurId,omitempty"`
	Livreur       *UserResponse  `json:"livreur,omitempty"`
	Status        DeliveryStatus `json:"status"`
	Type          DeliveryType   `json:"type"`
	Pickup        *Location      `json:"pickup"`
	Dropoff       *Location      `json:"dropoff"`
	DistanceKm    *float64       `json:"distanceKm,omitempty"`
	DurationMin   *float64       `json:"durationMin,omitempty"`
	VehicleType   VehicleType    `json:"vehicleType"`
	BasePrice     *float64       `json:"basePrice,omitempty"`
	WaitingMin    *float64       `json:"waitingMin,omitempty"`
	FinalPrice    float64        `json:"finalPrice"`
	PaymentMethod PaymentMethod  `json:"paymentMethod"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	PaidAt        *time.Time     `json:"paidAt,omitempty"`
	Package       *Package       `json:"package,omitempty"`
	Moving        *MovingService `json:"moving,omitempty"`
	Grouped       *GroupedDelivery `json:"grouped,omitempty"`
}

// MovingService represents moving service details
type MovingService struct {
	ID                  string   `json:"id"`
	DeliveryID          string   `json:"deliveryId"`
	VehicleSize         string   `json:"vehicleSize"`
	HelpersCount        int      `json:"helpersCount"`
	Floors              int      `json:"floors"`
	HasElevator         bool     `json:"hasElevator"`
	NeedsDisassembly    bool     `json:"needsDisassembly"`
	HasFragileItems     bool     `json:"hasFragileItems"`
	AdditionalServices  []string `json:"additionalServices"`
	SpecialInstructions *string  `json:"specialInstructions,omitempty"`
	EstimatedVolume     *float64 `json:"estimatedVolume,omitempty"`
	HelpersCost         float64  `json:"helpersCost"`
	VehicleCost         float64  `json:"vehicleCost"`
	ServiceCost         float64  `json:"serviceCost"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// GroupedDelivery represents grouped delivery details
type GroupedDelivery struct {
	ID                 string    `json:"id"`
	DeliveryID         string    `json:"deliveryId"`
	TotalZones         int       `json:"totalZones"`
	CompletedZones     int       `json:"completedZones"`
	DiscountPercentage float64   `json:"discountPercentage"`
	OriginalPrice      float64   `json:"originalPrice"`
	FinalPrice         float64   `json:"finalPrice"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// IsValidStatus checks if the delivery status is valid
func (s DeliveryStatus) IsValid() bool {
	validStatuses := []DeliveryStatus{
		DeliveryStatusPending, DeliveryStatusAccepted, DeliveryStatusPickedUp,
		DeliveryStatusDelivered, DeliveryStatusCancelled, DeliveryStatusZoneAssigned,
		DeliveryStatusPickupInProgress, DeliveryStatusPickupCompleted,
		DeliveryStatusDeliveryInProgress, DeliveryStatusAssignedToHelper,
		DeliveryStatusHelpersConfirmed, DeliveryStatusArrivedAtPickup,
		DeliveryStatusLoadingInProgress, DeliveryStatusLoadingCompleted,
		DeliveryStatusInTransit, DeliveryStatusArrivedAtDestination,
		DeliveryStatusUnloadingInProgress, DeliveryStatusUnloadingCompleted,
		DeliveryStatusArrivedAtDropoff, DeliveryStatusEnRoute,
		DeliveryStatusDispatchInProgress, DeliveryStatusSorted,
		DeliveryStatusSortingInProgress,
	}
	
	for _, status := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// IsValidType checks if the delivery type is valid
func (t DeliveryType) IsValid() bool {
	return t == DeliveryTypeSimple || t == DeliveryTypeExpress || 
		   t == DeliveryTypeGroupee || t == DeliveryTypeDemenagement
}

// IsValidVehicleType checks if the vehicle type is valid
func (vt VehicleType) IsValid() bool {
	return vt == VehicleTypeMoto || vt == VehicleTypeVoiture || vt == VehicleTypeCamionnette
}

// IsValidPaymentMethod checks if the payment method is valid
func (pm PaymentMethod) IsValid() bool {
	validMethods := []PaymentMethod{
		PaymentMethodCash, PaymentMethodMobileMoneyOrange,
		PaymentMethodMobileMoneyMTN, PaymentMethodMobileMoneyMoov,
		PaymentMethodMobileMoneyWave,
	}
	
	for _, method := range validMethods {
		if pm == method {
			return true
		}
	}
	return false
}

// CanBeAssigned checks if delivery can be assigned to a driver
func (d *Delivery) CanBeAssigned() bool {
	return d.Status == DeliveryStatusPending && d.LivreurID == nil
}

// CanBeCancelled checks if delivery can be cancelled
func (d *Delivery) CanBeCancelled() bool {
	return d.Status != DeliveryStatusDelivered && 
		   d.Status != DeliveryStatusCancelled &&
		   d.PaidAt == nil
}

// IsCompleted checks if delivery is completed
func (d *Delivery) IsCompleted() bool {
	return d.Status == DeliveryStatusDelivered
}

// IsPaid checks if delivery is paid
func (d *Delivery) IsPaid() bool {
	return d.PaidAt != nil
}

// RequiresSpecialVehicle checks if delivery requires special vehicle
func (d *Delivery) RequiresSpecialVehicle() bool {
	return d.Type == DeliveryTypeDemenagement
}

// IsGroupedDelivery checks if this is a grouped delivery
func (d *Delivery) IsGroupedDelivery() bool {
	return d.Type == DeliveryTypeGroupee
}

// IsMovingDelivery checks if this is a moving delivery
func (d *Delivery) IsMovingDelivery() bool {
	return d.Type == DeliveryTypeDemenagement
}

// GetExpectedDuration returns expected duration based on type
func (d *Delivery) GetExpectedDuration() int {
	switch d.Type {
	case DeliveryTypeExpress:
		return 30 // 30 minutes
	case DeliveryTypeSimple:
		return 60 // 1 hour
	case DeliveryTypeGroupee:
		return 120 // 2 hours
	case DeliveryTypeDemenagement:
		return 480 // 8 hours
	default:
		return 60
	}
}

// ToResponse converts Delivery to DeliveryResponse
func (d *Delivery) ToResponse() *DeliveryResponse {
	return &DeliveryResponse{
		ID:            d.ID,
		ClientID:      d.ClientID,
		LivreurID:     d.LivreurID,
		Status:        d.Status,
		Type:          d.Type,
		DistanceKm:    d.DistanceKm,
		DurationMin:   d.DurationMin,
		VehicleType:   d.VehicleType,
		BasePrice:     d.BasePrice,
		WaitingMin:    d.WaitingMin,
		FinalPrice:    d.FinalPrice,
		PaymentMethod: d.PaymentMethod,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
		PaidAt:        d.PaidAt,
	}
}
