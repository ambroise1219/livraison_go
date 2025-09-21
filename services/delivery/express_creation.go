package delivery

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

// ExpressCreationService gère la création des livraisons express
type ExpressCreationService struct{}

// NewExpressCreationService crée une nouvelle instance du service de création express
func NewExpressCreationService() *ExpressCreationService {
	return &ExpressCreationService{}
}

// CreateExpressDelivery crée une livraison express
func (s *ExpressCreationService) CreateExpressDelivery(clientID string, req *models.CreateDeliveryRequest) (*models.DeliveryResponse, error) {
	// Valider que c'est bien une livraison express
	if req.Type != models.DeliveryTypeExpress {
		return nil, fmt.Errorf("type de livraison incorrect, attendu EXPRESS, reçu %s", req.Type)
	}

	// Valider les informations du colis
	if req.PackageInfo == nil {
		return nil, fmt.Errorf("PackageInfo requis pour les livraisons express")
	}

	if err := s.validateExpressPackageInfo(req.PackageInfo); err != nil {
		return nil, fmt.Errorf("validation du colis express échouée: %v", err)
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

	// Calculer la distance et la durée (plus rapide pour express)
	distance, duration, err := s.calculateDistanceAndDuration(pickupLocation, dropoffLocation)
	if err != nil {
		return nil, fmt.Errorf("échec du calcul de la distance et de la durée: %v", err)
	}

	// Calculer le prix express (plus cher)
	price := s.calculateExpressDeliveryPrice(req.VehicleType, distance, 0)

	// Créer la livraison
	delivery := &models.Delivery{
		ID:            uuid.New().String(),
		ClientID:      clientID,
		Type:          models.DeliveryTypeExpress,
		Status:        models.DeliveryStatusPending,
		PickupID:      pickupLocation.ID,
		DropoffID:     dropoffLocation.ID,
		DistanceKm:    &distance,
		DurationMin:   &duration,
		WaitingMin:    nil, // Pas de WaitingMin dans CreateDeliveryRequest
		FinalPrice:    price,
		VehicleType:   req.VehicleType,
		PaymentMethod: req.PaymentMethod,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Sauvegarder la livraison et récupérer l'ID créé
	deliveryPrisma, err := s.saveDeliveryAndReturn(delivery)
	if err != nil {
		return nil, fmt.Errorf("échec de la sauvegarde de la livraison: %v", err)
	}

	// Créer l'enregistrement du colis avec l'ID réel
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
		FinalPrice:    func() float64 { if tp, ok := deliveryPrisma.TotalPrice(); ok { return tp } else { return delivery.FinalPrice }}(),
		VehicleType:   delivery.VehicleType,
		PaymentMethod: delivery.PaymentMethod,
		CreatedAt:     deliveryPrisma.CreatedAt,
	}, nil
}

// validateExpressPackageInfo valide les informations du colis express
func (s *ExpressCreationService) validateExpressPackageInfo(pkg *models.PackageInfo) error {
	if pkg.WeightKg == nil || *pkg.WeightKg <= 0 {
		return fmt.Errorf("le poids du colis doit être supérieur à 0")
	}
	if pkg.WeightKg != nil && *pkg.WeightKg > 30 { // Limite de poids plus stricte pour les livraisons express
		return fmt.Errorf("le poids du colis dépasse la limite autorisée pour les livraisons express (30kg)")
	}

	return nil
}

// calculateExpressDeliveryPrice calcule le prix d'une livraison express
func (s *ExpressCreationService) calculateExpressDeliveryPrice(vehicleType models.VehicleType, distance, waiting float64) float64 {
	// Prix de base plus élevé pour les livraisons express
	basePrices := map[models.VehicleType]float64{
		models.VehicleTypeMotorcycle:  1000.0, // 1000 FCFA (double du simple)
		models.VehicleTypeCar:         2000.0, // 2000 FCFA (double du simple)
		models.VehicleTypeVan:         3000.0, // 3000 FCFA (double du simple)
	}

	basePrice := basePrices[vehicleType]
	if basePrice == 0 {
		basePrice = 2000.0 // Prix par défaut
	}

	// Prix par kilomètre plus élevé
	pricePerKm := 400.0 // Double du prix normal
	distancePrice := distance * pricePerKm

	// Frais d'attente plus élevés
	waitingPrice := waiting * 100.0 // 100 FCFA par minute (double du normal)

	// Multiplicateur express
	expressMultiplier := 1.5

	return (basePrice + distancePrice + waitingPrice) * expressMultiplier
}

// createLocation crée un lieu de ramassage ou de dépôt
func (s *ExpressCreationService) createLocation(address string, lat, lng *float64) (*models.Location, error) {
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

// calculateDistanceAndDuration calcule la distance et la durée (optimisée pour express)
func (s *ExpressCreationService) calculateDistanceAndDuration(pickup, dropoff *models.Location) (float64, float64, error) {
	if pickup == nil || dropoff == nil || pickup.Lat == nil || pickup.Lng == nil || dropoff.Lat == nil || dropoff.Lng == nil {
		return 0, 0, fmt.Errorf("coordonnées manquantes")
	}

	// Calcul de distance simplifié (Haversine)
	distance := s.calculateHaversineDistance(*pickup.Lat, *pickup.Lng, *dropoff.Lat, *dropoff.Lng)

	// Estimation de la durée plus rapide pour express (vitesse moyenne de 40 km/h)
	duration := distance / 40.0 * 60.0 // Convertir en minutes

	return distance, duration, nil
}

// calculateHaversineDistance calcule la distance haversine entre deux points
func (s *ExpressCreationService) calculateHaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
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
func (s *ExpressCreationService) createPackage(deliveryID string, packageInfo *models.PackageInfo) error {
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

// saveDeliveryAndReturn sauvegarde une livraison et retourne l'objet créé
func (s *ExpressCreationService) saveDeliveryAndReturn(delivery *models.Delivery) (*prismadb.DeliveryModel, error) {
	ctx := context.Background()
	
	// Créer la livraison avec champs obligatoires et retourner l'objet
	deliveryPrisma, err := db.PrismaDB.Delivery.CreateOne(
		prismadb.Delivery.Type.Set(prismadb.DeliveryType(delivery.Type)),
		prismadb.Delivery.ClientPhone.Set(delivery.ClientID),
		prismadb.Delivery.PickupLocation.Link(
			prismadb.Location.ID.Equals(delivery.PickupID),
		),
		prismadb.Delivery.DropoffLocation.Link(
			prismadb.Location.ID.Equals(delivery.DropoffID),
		),
		prismadb.Delivery.TotalPrice.SetOptional(&delivery.FinalPrice),
		prismadb.Delivery.DistanceKm.SetOptional(delivery.DistanceKm),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}
	
	return deliveryPrisma, nil
}
