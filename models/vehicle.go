package models

import (
	"time"
)

// Vehicle represents a vehicle
type Vehicle struct {
	ID                     string      `json:"id"`
	Type                   VehicleType `json:"type" validate:"required"`
	UserID                 string      `json:"userId" validate:"required"`
	Nom                    *string     `json:"nom,omitempty"`
	PlaqueImmatriculation  *string     `json:"plaqueImmatriculation,omitempty"`
	Couleur                *string     `json:"couleur,omitempty"`
	Marque                 *string     `json:"marque,omitempty"`
	Modele                 *string     `json:"modele,omitempty"`
	Annee                  *int        `json:"annee,omitempty"`
	CarteGriseRecto        *string     `json:"carteGrise_recto,omitempty"`
	CarteGriseVerso        *string     `json:"carteGrise_verso,omitempty"`
	VignetteRecto          *string     `json:"vignette_recto,omitempty"`
	VignetteVerso          *string     `json:"vignette_verso,omitempty"`
	CreatedAt              time.Time   `json:"createdAt"`
}

// CreateVehicleRequest represents request for creating a vehicle
type CreateVehicleRequest struct {
	Type                  VehicleType `json:"type" validate:"required"`
	Nom                   *string     `json:"nom,omitempty"`
	PlaqueImmatriculation *string     `json:"plaqueImmatriculation,omitempty"`
	Couleur               *string     `json:"couleur,omitempty"`
	Marque                *string     `json:"marque,omitempty"`
	Modele                *string     `json:"modele,omitempty"`
	Annee                 *int        `json:"annee,omitempty"`
	CarteGriseRecto       *string     `json:"carteGrise_recto,omitempty"`
	CarteGriseVerso       *string     `json:"carteGrise_verso,omitempty"`
	VignetteRecto         *string     `json:"vignette_recto,omitempty"`
	VignetteVerso         *string     `json:"vignette_verso,omitempty"`
}

// UpdateVehicleRequest represents request for updating a vehicle
type UpdateVehicleRequest struct {
	Type                  *VehicleType `json:"type,omitempty"`
	Nom                   *string      `json:"nom,omitempty"`
	PlaqueImmatriculation *string      `json:"plaqueImmatriculation,omitempty"`
	Couleur               *string      `json:"couleur,omitempty"`
	Marque                *string      `json:"marque,omitempty"`
	Modele                *string      `json:"modele,omitempty"`
	Annee                 *int         `json:"annee,omitempty"`
	CarteGriseRecto       *string      `json:"carteGrise_recto,omitempty"`
	CarteGriseVerso       *string      `json:"carteGrise_verso,omitempty"`
	VignetteRecto         *string      `json:"vignette_recto,omitempty"`
	VignetteVerso         *string      `json:"vignette_verso,omitempty"`
}

// VehicleResponse represents vehicle data in response
type VehicleResponse struct {
	ID                    string      `json:"id"`
	Type                  VehicleType `json:"type"`
	UserID                string      `json:"userId"`
	User                  *UserResponse `json:"user,omitempty"`
	Nom                   *string     `json:"nom,omitempty"`
	PlaqueImmatriculation *string     `json:"plaqueImmatriculation,omitempty"`
	Couleur               *string     `json:"couleur,omitempty"`
	Marque                *string     `json:"marque,omitempty"`
	Modele                *string     `json:"modele,omitempty"`
	Annee                 *int        `json:"annee,omitempty"`
	CarteGriseRecto       *string     `json:"carteGrise_recto,omitempty"`
	CarteGriseVerso       *string     `json:"carteGrise_verso,omitempty"`
	VignetteRecto         *string     `json:"vignette_recto,omitempty"`
	VignetteVerso         *string     `json:"vignette_verso,omitempty"`
	CreatedAt             time.Time   `json:"createdAt"`
	IsCompletelyDocumented bool       `json:"isCompletelyDocumented"`
}

// DriverLocation represents driver location for tracking
type DriverLocation struct {
	ID          string      `json:"id"`
	DriverID    string      `json:"driverId" validate:"required"`
	Lat         *float64    `json:"lat,omitempty" validate:"omitempty,gte=-90,lte=90"`
	Lng         *float64    `json:"lng,omitempty" validate:"omitempty,gte=-180,lte=180"`
	Timestamp   time.Time   `json:"timestamp"`
	IsAvailable bool        `json:"isAvailable"`
	VehicleType VehicleType `json:"vehicleType" validate:"required"`
}


// HasRequiredDocuments checks if vehicle has all required documents
func (v *Vehicle) HasRequiredDocuments() bool {
	return v.CarteGriseRecto != nil && v.CarteGriseVerso != nil &&
		   v.VignetteRecto != nil && v.VignetteVerso != nil &&
		   v.PlaqueImmatriculation != nil && *v.PlaqueImmatriculation != ""
}

// IsRegistrationComplete checks if basic registration info is complete
func (v *Vehicle) IsRegistrationComplete() bool {
	return v.Marque != nil && *v.Marque != "" &&
		   v.Modele != nil && *v.Modele != "" &&
		   v.Couleur != nil && *v.Couleur != "" &&
		   v.PlaqueImmatriculation != nil && *v.PlaqueImmatriculation != ""
}

// GetDisplayName returns a display name for the vehicle
func (v *Vehicle) GetDisplayName() string {
	if v.Nom != nil && *v.Nom != "" {
		return *v.Nom
	}
	
	if v.Marque != nil && v.Modele != nil {
		return *v.Marque + " " + *v.Modele
	}
	
	if v.PlaqueImmatriculation != nil && *v.PlaqueImmatriculation != "" {
		return *v.PlaqueImmatriculation
	}
	
	return string(v.Type)
}

// CanBeUsedForDelivery checks if vehicle can be used for delivery
func (v *Vehicle) CanBeUsedForDelivery() bool {
	return v.IsRegistrationComplete() && v.HasRequiredDocuments()
}

// IsCompatibleWithDeliveryType checks if vehicle type is compatible with delivery type
func (v *Vehicle) IsCompatibleWithDeliveryType(deliveryType DeliveryType) bool {
	switch deliveryType {
	case DeliveryTypeSimple, DeliveryTypeExpress:
		return true // All vehicle types can handle simple and express
	case DeliveryTypeGroupee:
		return v.Type == VehicleTypeVoiture || v.Type == VehicleTypeCamionnette
	case DeliveryTypeDemenagement:
		return v.Type == VehicleTypeCamionnette
	default:
		return false
	}
}

// GetCapacityWeight returns vehicle capacity in kg
func (v *Vehicle) GetCapacityWeight() float64 {
	switch v.Type {
	case VehicleTypeMoto:
		return 50.0 // 50 kg max
	case VehicleTypeVoiture:
		return 200.0 // 200 kg max
	case VehicleTypeCamionnette:
		return 1000.0 // 1000 kg max
	default:
		return 50.0
	}
}

// GetCapacityVolume returns vehicle capacity in cubic meters
func (v *Vehicle) GetCapacityVolume() float64 {
	switch v.Type {
	case VehicleTypeMoto:
		return 0.1 // 0.1 m³
	case VehicleTypeVoiture:
		return 2.0 // 2 m³
	case VehicleTypeCamionnette:
		return 10.0 // 10 m³
	default:
		return 0.1
	}
}

// ToResponse converts Vehicle to VehicleResponse
func (v *Vehicle) ToResponse() *VehicleResponse {
	return &VehicleResponse{
		ID:                     v.ID,
		Type:                   v.Type,
		UserID:                 v.UserID,
		Nom:                    v.Nom,
		PlaqueImmatriculation:  v.PlaqueImmatriculation,
		Couleur:                v.Couleur,
		Marque:                 v.Marque,
		Modele:                 v.Modele,
		Annee:                  v.Annee,
		CarteGriseRecto:        v.CarteGriseRecto,
		CarteGriseVerso:        v.CarteGriseVerso,
		VignetteRecto:          v.VignetteRecto,
		VignetteVerso:          v.VignetteVerso,
		CreatedAt:              v.CreatedAt,
		IsCompletelyDocumented: v.HasRequiredDocuments(),
	}
}
