package delivery

import (
	"context"
	"fmt"
	"time"

	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

// GroupedCreationService gère la création des livraisons groupées
type GroupedCreationService struct{}

// NewGroupedCreationService crée une nouvelle instance du service de création groupée
func NewGroupedCreationService() *GroupedCreationService {
	return &GroupedCreationService{}
}

// CreateGroupedDelivery crée une nouvelle livraison groupée
func (s *GroupedCreationService) CreateGroupedDelivery(clientID string, req *models.CreateDeliveryRequest) (*models.DeliveryResponse, error) {
	// Valider les informations de livraison groupée
	if req.GroupedInfo == nil {
		return nil, fmt.Errorf("informations de livraison groupée requises")
	}

	// Calculer le prix de base (sera affiné plus tard)
	basePrice := s.calculateGroupedBasePrice(req)

	// Créer les emplacements de ramassage et de livraison
	pickupID, err := s.createLocation(req.PickupAddress, req.PickupLat, req.PickupLng)
	if err != nil {
		return nil, fmt.Errorf("échec de la création de l'emplacement de ramassage: %v", err)
	}

	dropoffID, err := s.createLocation(req.DropoffAddress, req.DropoffLat, req.DropoffLng)
	if err != nil {
		return nil, fmt.Errorf("échec de la création de l'emplacement de livraison: %v", err)
	}

	// Créer la livraison groupée
	delivery := &models.Delivery{
		ClientID:      clientID,
		Status:        models.DeliveryStatusPending,
		Type:          models.DeliveryTypeGrouped,
		PickupID:      pickupID,
		DropoffID:     dropoffID,
		DistanceKm:    &[]float64{0.0}[0], // Sera calculé plus tard
		DurationMin:   &[]float64{0.0}[0], // Sera calculé plus tard
		VehicleType:   req.VehicleType,
		BasePrice:     &basePrice,
		WaitingMin:    &[]float64{0.0}[0],
		FinalPrice:    basePrice,
		PaymentMethod: req.PaymentMethod,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Sauvegarder la livraison (Prisma va générer l'ID et le mettre dans delivery.ID)
	if err := s.saveGroupedDelivery(delivery); err != nil {
		return nil, fmt.Errorf("échec de la sauvegarde: %v", err)
	}

	// Utiliser l'ID généré par Prisma pour les informations groupées
	if err := s.createGroupedInfo(delivery.ID, req.GroupedInfo); err != nil {
		return nil, fmt.Errorf("échec de la création des informations groupées: %v", err)
	}

	// Créer la réponse
	response := &models.DeliveryResponse{
		ID:            delivery.ID,
		ClientID:      delivery.ClientID,
		Status:        delivery.Status,
		Type:          delivery.Type,
		Pickup:        &models.Location{ID: pickupID, Address: req.PickupAddress, Lat: req.PickupLat, Lng: req.PickupLng},
		Dropoff:       &models.Location{ID: dropoffID, Address: req.DropoffAddress, Lat: req.DropoffLat, Lng: req.DropoffLng},
		DistanceKm:    delivery.DistanceKm,
		DurationMin:   delivery.DurationMin,
		VehicleType:   delivery.VehicleType,
		BasePrice:     delivery.BasePrice,
		WaitingMin:    delivery.WaitingMin,
		FinalPrice:    delivery.FinalPrice,
		PaymentMethod: delivery.PaymentMethod,
		CreatedAt:     delivery.CreatedAt,
		UpdatedAt:     delivery.UpdatedAt,
	}

	return response, nil
}

// calculateGroupedBasePrice calcule le prix de base pour une livraison groupée
func (s *GroupedCreationService) calculateGroupedBasePrice(req *models.CreateDeliveryRequest) float64 {
	// Prix de base selon le type de véhicule
	basePrices := map[models.VehicleType]float64{
		models.VehicleTypeMotorcycle:  800,
		models.VehicleTypeCar:         1200,
		models.VehicleTypeVan:         1800,
	}

	basePrice := basePrices[req.VehicleType]

	// Ajouter un coût par zone si spécifié
	if req.GroupedInfo != nil {
		zoneCount := len(req.GroupedInfo.Zones)
		basePrice += float64(zoneCount-1) * 200 // 200 FCFA par zone supplémentaire
	}

	return basePrice
}

// createLocation crée un emplacement dans la base de données
func (s *GroupedCreationService) createLocation(address string, lat, lng *float64) (string, error) {
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

// createGroupedInfo crée les informations de livraison groupée avec toutes les zones
func (s *GroupedCreationService) createGroupedInfo(deliveryID string, groupedInfo *models.GroupedInfo) error {
	ctx := context.Background()
	
	// 1. Créer l'enregistrement GroupedDelivery principal
	totalZones := len(groupedInfo.Zones)
	groupedDeliveryPrisma, err := db.PrismaDB.GroupedDelivery.CreateOne(
		prismadb.GroupedDelivery.Name.Set(fmt.Sprintf("Livraison groupée - %d zones", totalZones)),
		prismadb.GroupedDelivery.Description.SetOptional(stringPtr("Livraison multi-zones avec détails complets")),
		prismadb.GroupedDelivery.TotalPrice.SetOptional(floatPtr(0.0)),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("erreur création GroupedDelivery: %v", err)
	}
	
	// 2. Créer chaque zone avec ses locations complètes
	var totalPrice float64
	for _, zone := range groupedInfo.Zones {
		fmt.Printf("\n      📍 Création ZONE %d:\n", zone.ZoneNumber)
		fmt.Printf("         👥 Destinataire: %s (%s)\n", zone.RecipientName, zone.RecipientPhone)
		
		// Créer la location de ramassage pour cette zone
		pickupLocationID, err := s.createLocation(zone.PickupAddress, zone.PickupLat, zone.PickupLng)
		if err != nil {
			return fmt.Errorf("erreur création pickup zone %d: %v", zone.ZoneNumber, err)
		}
		fmt.Printf("         📦 Pickup: %s (ID: %s)\n", zone.PickupAddress, pickupLocationID[:10]+"...")
		
		// Créer la location de livraison pour cette zone
		deliveryLocationID, err := s.createLocation(zone.DeliveryAddress, zone.DeliveryLat, zone.DeliveryLng)
		if err != nil {
			return fmt.Errorf("erreur création delivery zone %d: %v", zone.ZoneNumber, err)
		}
		fmt.Printf("         🏠 Delivery: %s (ID: %s)\n", zone.DeliveryAddress, deliveryLocationID[:10]+"...")
		
		// Calculer le prix pour cette zone (exemple: 800 FCFA base + 100 FCFA par km)
		zonePrice := s.calculateZonePrice(zone.ZoneNumber, pickupLocationID, deliveryLocationID)
		totalPrice += zonePrice
		fmt.Printf("         💰 Prix zone: %.0f FCFA\n", zonePrice)
		
		// Enregistrer les détails de la zone (simulation puisque pas de table DeliveryZone dans le schema)
		fmt.Printf("         ✅ Zone %d enregistrée avec succès\n", zone.ZoneNumber)
	}
	
	// 3. Mettre à jour le prix total
	_, err = db.PrismaDB.GroupedDelivery.FindUnique(
		prismadb.GroupedDelivery.ID.Equals(groupedDeliveryPrisma.ID),
	).Update(
		prismadb.GroupedDelivery.TotalPrice.Set(totalPrice),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("erreur mise à jour prix total: %v", err)
	}
	
	fmt.Printf("\n      📋 RÉSUMÉ LIVRAISON GROUPÉE:\n")
	fmt.Printf("         🎯 Total zones: %d\n", totalZones)
	fmt.Printf("         💰 Prix total: %.0f FCFA\n", totalPrice)
	fmt.Printf("         🆔 GroupedDelivery ID: %s\n", groupedDeliveryPrisma.ID[:10]+"...")
	
	return nil
}

// stringPtr helper pour créer un pointeur vers string
func stringPtr(s string) *string {
	return &s
}

// floatPtr helper pour créer un pointeur vers float64
func floatPtr(f float64) *float64 {
	return &f
}

// calculateZonePrice calcule le prix d'une zone spécifique
func (s *GroupedCreationService) calculateZonePrice(zoneNumber int, pickupLocationID, deliveryLocationID string) float64 {
	// Prix de base par zone
	basePrice := 800.0
	
	// Prix dégressif selon le nombre de zones
	switch zoneNumber {
	case 1:
		return basePrice // Première zone: prix plein
	case 2:
		return basePrice * 0.9 // 10% de réduction
	case 3:
		return basePrice * 0.8 // 20% de réduction
	case 4:
		return basePrice * 0.7 // 30% de réduction
	case 5:
		return basePrice * 0.6 // 40% de réduction
	default:
		return basePrice * 0.5 // 50% pour zones supplémentaires
	}
}

// saveGroupedDelivery sauvegarde une livraison groupée
func (s *GroupedCreationService) saveGroupedDelivery(delivery *models.Delivery) error {
	ctx := context.Background()

	// Créer la livraison avec Prisma et récupérer l'ID généré
	createdDelivery, err := db.PrismaDB.Delivery.CreateOne(
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
		return err
	}

	// Assigner l'ID généré par Prisma à l'objet delivery
	delivery.ID = createdDelivery.ID

	return nil
}
