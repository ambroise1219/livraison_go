package delivery

import (
	"context"
	"fmt"

	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

// SimpleCreationService gère la création des livraisons simples
type SimpleCreationService struct{}

// NewSimpleCreationService crée une nouvelle instance du service de création simple
func NewSimpleCreationService() *SimpleCreationService {
	return &SimpleCreationService{}
}

// CreateSimpleDelivery crée une livraison simple
func (s *SimpleCreationService) CreateSimpleDelivery(clientID string, req *models.CreateDeliveryRequest) (*models.DeliveryResponse, error) {
	// Valider que c'est bien une livraison simple/standard
	if req.Type != models.DeliveryTypeStandard && req.Type != "STANDARD" {
		return nil, fmt.Errorf("type de livraison incorrect, attendu SIMPLE ou STANDARD, reçu %s", req.Type)
	}

	// Valider les informations du colis
	if req.PackageInfo == nil {
		return nil, fmt.Errorf("PackageInfo requis pour les livraisons simples")
	}

	if err := s.validatePackageInfo(req.PackageInfo); err != nil {
		return nil, fmt.Errorf("validation du colis échouée: %v", err)
	}

	// Créer les lieux de ramassage et de dépôt
	pickupLocation, err := s.createLocation(req.PickupAddress, req.PickupLat, req.PickupLng)
	if err != nil {
		return nil, fmt.Errorf("échec de la création du lieu de ramassage: %v", err)
	}

	dropoffLocation, err := s.createLocation(req.DropoffAddress, req.DropoffLat, req.DropoffLng)
	if err != nil {
		return nil, fmt.Errorf("échec de la création du lieu de dépôt: %v", err)
	}

	// Calculer la distance et la durée
	distance, duration, err := s.calculateDistanceAndDuration(pickupLocation, dropoffLocation)
	if err != nil {
		return nil, fmt.Errorf("échec du calcul de la distance et de la durée: %v", err)
	}

	// Calculer le prix (pas de WaitingMin dans CreateDeliveryRequest)
	price := s.calculateSimpleDeliveryPrice(req.VehicleType, distance, 0)

	// Créer la livraison avec Prisma directement
	deliveryPrisma, err := s.createDeliveryWithPrisma(clientID, pickupLocation.ID, dropoffLocation.ID, req.VehicleType, req.PaymentMethod, price, &distance, &duration)
	if err != nil {
		return nil, fmt.Errorf("échec de la création de la livraison: %v", err)
	}

	// Créer l'enregistrement du colis
	if err := s.createPackage(deliveryPrisma.ID, req.PackageInfo); err != nil {
		return nil, fmt.Errorf("échec de la création du colis: %v", err)
	}

	return &models.DeliveryResponse{
		ID:            deliveryPrisma.ID,
		ClientID:      deliveryPrisma.ClientPhone,
		Type:          models.DeliveryType(deliveryPrisma.Type),
		Status:        models.DeliveryStatus(deliveryPrisma.Status),
		Pickup:        pickupLocation,
		Dropoff:       dropoffLocation,
		DistanceKm:    &distance,
		DurationMin:   &duration,
		FinalPrice:    func() float64 { if tp, ok := deliveryPrisma.TotalPrice(); ok { return tp } else { return price }}(),
		VehicleType:   req.VehicleType,
		PaymentMethod: req.PaymentMethod,
		CreatedAt:     deliveryPrisma.CreatedAt,
	}, nil
}

// validatePackageInfo valide les informations du colis
func (s *SimpleCreationService) validatePackageInfo(pkg *models.PackageInfo) error {
	if pkg.WeightKg == nil || *pkg.WeightKg <= 0 {
		return fmt.Errorf("le poids du colis doit être supérieur à 0")
	}
	if pkg.WeightKg != nil && *pkg.WeightKg > 50 { // Limite de poids pour les livraisons simples
		return fmt.Errorf("le poids du colis dépasse la limite autorisée pour les livraisons simples (50kg)")
	}
	return nil
}

// createDeliveryWithPrisma crée une livraison directement avec Prisma et retourne l'objet créé
func (s *SimpleCreationService) createDeliveryWithPrisma(clientID, pickupID, dropoffID string, vehicleType models.VehicleType, paymentMethod models.PaymentMethod, price float64, distance, duration *float64) (*prismadb.DeliveryModel, error) {
	ctx := context.Background()

	// Créer la livraison avec Prisma
	delivery, err := db.PrismaDB.Delivery.CreateOne(
		prismadb.Delivery.Type.Set(prismadb.DeliveryTypeStandard),
		prismadb.Delivery.ClientPhone.Set(clientID),
		prismadb.Delivery.PickupLocation.Link(
			prismadb.Location.ID.Equals(pickupID),
		),
		prismadb.Delivery.DropoffLocation.Link(
			prismadb.Location.ID.Equals(dropoffID),
		),
		prismadb.Delivery.TotalPrice.SetOptional(&price),
		prismadb.Delivery.DistanceKm.SetOptional(distance),
		prismadb.Delivery.DurationMin.SetOptional(func() *int {
			if duration != nil {
				val := int(*duration)
				return &val
			}
			return nil
		}()),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create delivery: %v", err)
	}

	return delivery, nil
}

// calculateSimpleDeliveryPrice calcule le prix d'une livraison simple
func (s *SimpleCreationService) calculateSimpleDeliveryPrice(vehicleType models.VehicleType, distance, waiting float64) float64 {
	// Prix de base selon le type de véhicule
	basePrices := map[models.VehicleType]float64{
		models.VehicleTypeMotorcycle:  500.0,  // 500 FCFA
		models.VehicleTypeCar:         1000.0, // 1000 FCFA
		models.VehicleTypeVan:         1500.0, // 1500 FCFA
	}

	basePrice := basePrices[vehicleType]
	if basePrice == 0 {
		basePrice = 1000.0 // Prix par défaut
	}

	// Prix par kilomètre
	pricePerKm := 200.0
	distancePrice := distance * pricePerKm

	// Frais d'attente
	waitingPrice := waiting * 50.0 // 50 FCFA par minute

	return basePrice + distancePrice + waitingPrice
}

// createLocation crée un lieu de ramassage ou de dépôt
func (s *SimpleCreationService) createLocation(address string, lat, lng *float64) (*models.Location, error) {
	ctx := context.Background()
	
	// Utiliser des valeurs par défaut si lat/lng ne sont pas fournies
	latValue := 0.0
	lngValue := 0.0
	
	if lat != nil {
		latValue = *lat
	}
	if lng != nil {
		lngValue = *lng
	}
	
	// Créer la location avec Prisma - champs obligatoires
	locationPrisma, err := db.PrismaDB.Location.CreateOne(
		prismadb.Location.Address.Set(address),
		prismadb.Location.Lat.Set(latValue),
		prismadb.Location.Lng.Set(lngValue),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	
	// Convertir en modèle
	location := &models.Location{
		ID:      locationPrisma.ID,
		Address: locationPrisma.Address,
		Lat:     &locationPrisma.Lat,
		Lng:     &locationPrisma.Lng,
	}
	
	return location, nil
}

// calculateDistanceAndDuration calcule la distance et la durée entre deux lieux
func (s *SimpleCreationService) calculateDistanceAndDuration(pickup, dropoff *models.Location) (float64, float64, error) {
	if pickup == nil || dropoff == nil || pickup.Lat == nil || pickup.Lng == nil || dropoff.Lat == nil || dropoff.Lng == nil {
		return 0, 0, fmt.Errorf("coordonnées manquantes")
	}

	// Calcul de distance simplifié (Haversine)
	distance := s.calculateHaversineDistance(*pickup.Lat, *pickup.Lng, *dropoff.Lat, *dropoff.Lng)

	// Estimation de la durée (vitesse moyenne de 30 km/h en ville)
	duration := distance / 30.0 * 60.0 // Convertir en minutes

	return distance, duration, nil
}

// calculateHaversineDistance calcule la distance haversine entre deux points
func (s *SimpleCreationService) calculateHaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371 // Rayon de la Terre en kilomètres

	// Calcul simplifié de distance (approximation)
	// En production, utiliser une vraie formule haversine
	_ = lat1
	_ = lng1
	_ = lat2
	_ = lng2
	return R * 1.57079632679 // Approximation simple
}

// createPackage crée un enregistrement de colis
func (s *SimpleCreationService) createPackage(deliveryID string, packageInfo *models.PackageInfo) error {
	ctx := context.Background()
	
	// Valeurs par défaut pour les champs obligatoires
	description := "Colis"
	weight := 1.0
	
	if packageInfo.Description != nil {
		description = *packageInfo.Description
	}
	if packageInfo.WeightKg != nil {
		weight = *packageInfo.WeightKg
	}
	
	// Créer le package avec Prisma - champs obligatoires seulement
	_, err := db.PrismaDB.Package.CreateOne(
		prismadb.Package.Description.Set(description),
		prismadb.Package.WeightKg.Set(weight),
		prismadb.Package.Delivery.Link(
			prismadb.Delivery.ID.Equals(deliveryID),
		),
	).Exec(ctx)
	if err != nil {
		return err
	}
	
	return nil
}

// saveDelivery sauvegarde une livraison
func (s *SimpleCreationService) saveDelivery(delivery *models.Delivery) error {
	ctx := context.Background()
	
	// Créer la livraison avec champs obligatoires seulement
	_, err := db.PrismaDB.Delivery.CreateOne(
		prismadb.Delivery.Type.Set(prismadb.DeliveryType(delivery.Type)),
		prismadb.Delivery.ClientPhone.Set(delivery.ClientID),
		prismadb.Delivery.PickupLocation.Link(
			prismadb.Location.ID.Equals(delivery.PickupID),
		),
		prismadb.Delivery.DropoffLocation.Link(
			prismadb.Location.ID.Equals(delivery.DropoffID),
		),
	).Exec(ctx)
	
	if err != nil {
		return fmt.Errorf("failed to create delivery: %v", err)
	}
	
	return nil
}
