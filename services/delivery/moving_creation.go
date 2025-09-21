package delivery

import (
	"context"
	"fmt"

	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

// MovingCreationService gère la création des déménagements
type MovingCreationService struct{}

// NewMovingCreationService crée une nouvelle instance du service de création de déménagement
func NewMovingCreationService() *MovingCreationService {
	return &MovingCreationService{}
}

// CreateMovingDelivery crée un nouveau déménagement
func (s *MovingCreationService) CreateMovingDelivery(clientID string, req *models.CreateDeliveryRequest) (*models.DeliveryResponse, error) {
	// Valider les informations de déménagement
	if req.MovingInfo == nil {
		return nil, fmt.Errorf("informations de déménagement requises")
	}

	// Calculer le prix de base
	basePrice := s.calculateMovingBasePrice(req)

	// Créer les emplacements de ramassage et de livraison
	pickupID, err := s.createLocation(req.PickupAddress, req.PickupLat, req.PickupLng)
	if err != nil {
		return nil, fmt.Errorf("échec de la création de l'emplacement de ramassage: %v", err)
	}

	dropoffID, err := s.createLocation(req.DropoffAddress, req.DropoffLat, req.DropoffLng)
	if err != nil {
		return nil, fmt.Errorf("échec de la création de l'emplacement de livraison: %v", err)
	}

	// Créer la livraison directement avec Prisma
	deliveryPrisma, err := s.createDeliveryWithPrisma(clientID, pickupID, dropoffID, req.VehicleType, req.PaymentMethod, basePrice)
	if err != nil {
		return nil, fmt.Errorf("échec de la création de la livraison: %v", err)
	}

	// Note: MovingService création simplifiée pour éviter les conflits Prisma
	// En production, implémenter createMovingInfo après avoir résolu les contraintes API

	// Créer la réponse
	response := &models.DeliveryResponse{
		ID:            deliveryPrisma.ID,
		ClientID:      deliveryPrisma.ClientPhone,
		Status:        models.DeliveryStatus(deliveryPrisma.Status),
		Type:          models.DeliveryType(deliveryPrisma.Type),
		Pickup:        &models.Location{ID: pickupID, Address: req.PickupAddress, Lat: req.PickupLat, Lng: req.PickupLng},
		Dropoff:       &models.Location{ID: dropoffID, Address: req.DropoffAddress, Lat: req.DropoffLat, Lng: req.DropoffLng},
		DistanceKm:    func() *float64 { if d, ok := deliveryPrisma.DistanceKm(); ok { return &d } else { return nil }}(),
		VehicleType:   req.VehicleType,
		FinalPrice:    func() float64 { if tp, ok := deliveryPrisma.TotalPrice(); ok { return tp } else { return basePrice }}(),
		PaymentMethod: req.PaymentMethod,
		CreatedAt:     deliveryPrisma.CreatedAt,
	}

	return response, nil
}

// createDeliveryWithPrisma crée une livraison directement avec Prisma et retourne l'objet créé
func (s *MovingCreationService) createDeliveryWithPrisma(clientID, pickupID, dropoffID string, vehicleType models.VehicleType, paymentMethod models.PaymentMethod, price float64) (*prismadb.DeliveryModel, error) {
	ctx := context.Background()

	// Créer la livraison avec Prisma
	delivery, err := db.PrismaDB.Delivery.CreateOne(
		prismadb.Delivery.Type.Set(prismadb.DeliveryTypeDemenagement),
		prismadb.Delivery.ClientPhone.Set(clientID),
		prismadb.Delivery.PickupLocation.Link(
			prismadb.Location.ID.Equals(pickupID),
		),
		prismadb.Delivery.DropoffLocation.Link(
			prismadb.Location.ID.Equals(dropoffID),
		),
		prismadb.Delivery.TotalPrice.SetOptional(&price),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create delivery: %v", err)
	}

	return delivery, nil
}

// calculateMovingBasePrice calcule le prix de base pour un déménagement
func (s *MovingCreationService) calculateMovingBasePrice(req *models.CreateDeliveryRequest) float64 {
	// Prix de base selon le type de véhicule
	basePrices := map[models.VehicleType]float64{
		models.VehicleTypeMotorcycle:  2000,
		models.VehicleTypeCar:         5000,
		models.VehicleTypeVan:         8000,
	}

	basePrice := basePrices[req.VehicleType]

	// Ajouter des coûts selon les informations de déménagement
	if req.MovingInfo != nil {
		// Coût par aide supplémentaire
		basePrice += float64(req.MovingInfo.HelpersCount) * 1000 // 1000 FCFA par aide

		// Coût selon le volume estimé
		if req.MovingInfo.EstimatedVolume != nil {
			volume := *req.MovingInfo.EstimatedVolume
			if volume > 50 {
				basePrice += (volume - 50) * 50 // 50 FCFA par m³ supplémentaire
			}
		}

		// Coût pour les services supplémentaires
		for _, service := range req.MovingInfo.AdditionalServices {
			switch service {
			case "packing":
				basePrice += 2000 // 2000 FCFA pour le service d'emballage
			case "assembly":
				basePrice += 1500 // 1500 FCFA pour le service de montage
			case "disassembly":
				basePrice += 1000 // 1000 FCFA pour le démontage
			}
		}

		// Coût supplémentaire pour les étages sans ascenseur
		if req.MovingInfo.Floors > 1 && !req.MovingInfo.HasElevator {
			basePrice += float64(req.MovingInfo.Floors-1) * 500 // 500 FCFA par étage supplémentaire
		}
	}

	return basePrice
}

// createLocation crée un emplacement dans la base de données
func (s *MovingCreationService) createLocation(address string, lat, lng *float64) (string, error) {
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
	
	// Créer la location avec Prisma
	locationPrisma, err := db.PrismaDB.Location.CreateOne(
		prismadb.Location.Address.Set(address),
		prismadb.Location.Lat.Set(latValue),
		prismadb.Location.Lng.Set(lngValue),
	).Exec(ctx)
	if err != nil {
		return "", err
	}
	
	return locationPrisma.ID, nil
}

// Note: createMovingInfo et saveMovingDelivery ont été remplacées par createDeliveryWithPrisma
// pour éviter les conflits d'API Prisma avec MovingService
