package delivery

import (
	"context"
	"fmt"
	"log"

	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

// UpdateService handles delivery updates with specialized methods per type
type UpdateService struct{}

// NewUpdateService creates a new instance of the update service
func NewUpdateService() *UpdateService {
	return &UpdateService{}
}

// UpdateSimpleDelivery updates a simple delivery
func (s *UpdateService) UpdateSimpleDelivery(deliveryID string, req *models.UpdateSimpleDeliveryRequest) (*models.Delivery, error) {
	ctx := context.Background()

	// First, get current delivery to validate transition
	currentDelivery, err := s.getDelivery(ctx, deliveryID)
	if err != nil {
		return nil, err
	}

	// Validate status transition if status is being updated
	if req.Status != nil && !currentDelivery.CanTransitionTo(*req.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", currentDelivery.Status, *req.Status)
	}

	// Prepare update parameters
	updateParams := []prismadb.DeliverySetParam{}

	if req.Status != nil {
		updateParams = append(updateParams, prismadb.Delivery.Status.Set(prismadb.DeliveryStatus(*req.Status)))
	}
	if req.FinalPrice != nil {
		updateParams = append(updateParams, prismadb.Delivery.TotalPrice.Set(*req.FinalPrice))
	}

	// Update dropoff location if provided
	if req.DropoffAddress != nil || req.DropoffLat != nil || req.DropoffLng != nil {
		err := s.updateDeliveryLocation(ctx, currentDelivery.DropoffID, req.DropoffAddress, req.DropoffLat, req.DropoffLng)
		if err != nil {
			return nil, fmt.Errorf("failed to update dropoff location: %v", err)
		}
	}

	// Update delivery
	_, err = db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Update(updateParams...).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update simple delivery: %v", err)
	}

	// Update package if provided
	if req.PackageInfo != nil {
		err = s.updatePackage(ctx, deliveryID, req.PackageInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to update package: %v", err)
		}
	}

	// Create tracking entry if status changed
	if req.Status != nil {
		err = s.createTrackingEntry(ctx, deliveryID, *req.Status, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create tracking entry: %v", err)
		}
	}

	// Return updated delivery
	return s.getDelivery(ctx, deliveryID)
}

// UpdateExpressDelivery updates an express delivery
func (s *UpdateService) UpdateExpressDelivery(deliveryID string, req *models.UpdateExpressDeliveryRequest) (*models.Delivery, error) {
	ctx := context.Background()

	// Get current delivery to validate transition
	currentDelivery, err := s.getDelivery(ctx, deliveryID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if req.Status != nil && !currentDelivery.CanTransitionTo(*req.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", currentDelivery.Status, *req.Status)
	}

	// Prepare update parameters
	updateParams := []prismadb.DeliverySetParam{}

	if req.Status != nil {
		updateParams = append(updateParams, prismadb.Delivery.Status.Set(prismadb.DeliveryStatus(*req.Status)))
	}
	if req.FinalPrice != nil {
		updateParams = append(updateParams, prismadb.Delivery.TotalPrice.Set(*req.FinalPrice))
	}

	// Update delivery
	_, err = db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Update(updateParams...).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update express delivery: %v", err)
	}

	// Update package if provided
	if req.PackageInfo != nil {
		err = s.updatePackage(ctx, deliveryID, req.PackageInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to update package: %v", err)
		}
	}

	// Update express-specific fields if provided
	if req.ExpressTime != nil || req.Priority != nil {
		err = s.updateExpressDeliveryDetails(ctx, deliveryID, req.ExpressTime, req.Priority)
		if err != nil {
			return nil, fmt.Errorf("failed to update express details: %v", err)
		}
	}

	// Create tracking entry if status changed
	if req.Status != nil {
		err = s.createTrackingEntry(ctx, deliveryID, *req.Status, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create tracking entry: %v", err)
		}
	}

	return s.getDelivery(ctx, deliveryID)
}

// UpdateGroupedDelivery updates a grouped delivery
func (s *UpdateService) UpdateGroupedDelivery(deliveryID string, req *models.UpdateGroupedDeliveryRequest) (*models.Delivery, error) {
	ctx := context.Background()

	// Get current delivery to validate transition
	currentDelivery, err := s.getDelivery(ctx, deliveryID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if req.Status != nil && !currentDelivery.CanTransitionTo(*req.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", currentDelivery.Status, *req.Status)
	}

	// Prepare update parameters
	updateParams := []prismadb.DeliverySetParam{}

	if req.Status != nil {
		updateParams = append(updateParams, prismadb.Delivery.Status.Set(prismadb.DeliveryStatus(*req.Status)))
	}
	if req.FinalPrice != nil {
		updateParams = append(updateParams, prismadb.Delivery.TotalPrice.Set(*req.FinalPrice))
	}

	// Update delivery
	_, err = db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Update(updateParams...).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update grouped delivery: %v", err)
	}

	// Update zones if provided
	if len(req.Zones) > 0 {
		err = s.updateGroupedZones(ctx, deliveryID, req.Zones)
		if err != nil {
			return nil, fmt.Errorf("failed to update zones: %v", err)
		}
	}

	// Update grouped delivery details
	if req.CompletedZones != nil {
		err = s.updateGroupedDeliveryDetails(ctx, deliveryID, req.CompletedZones)
		if err != nil {
			return nil, fmt.Errorf("failed to update grouped details: %v", err)
		}
	}

	// Create tracking entry if status changed
	if req.Status != nil {
		err = s.createTrackingEntry(ctx, deliveryID, *req.Status, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create tracking entry: %v", err)
		}
	}

	return s.getDelivery(ctx, deliveryID)
}

// UpdateMovingDelivery updates a moving delivery
func (s *UpdateService) UpdateMovingDelivery(deliveryID string, req *models.UpdateMovingDeliveryRequest) (*models.Delivery, error) {
	ctx := context.Background()

	// Get current delivery to validate transition
	currentDelivery, err := s.getDelivery(ctx, deliveryID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if req.Status != nil && !currentDelivery.CanTransitionTo(*req.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", currentDelivery.Status, *req.Status)
	}

	// Prepare update parameters
	updateParams := []prismadb.DeliverySetParam{}

	if req.Status != nil {
		updateParams = append(updateParams, prismadb.Delivery.Status.Set(prismadb.DeliveryStatus(*req.Status)))
	}
	if req.FinalPrice != nil {
		updateParams = append(updateParams, prismadb.Delivery.TotalPrice.Set(*req.FinalPrice))
	}

	// Update delivery
	_, err = db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Update(updateParams...).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update moving delivery: %v", err)
	}

	// Update moving items if provided
	if len(req.Items) > 0 {
		err = s.updateMovingItems(ctx, deliveryID, req.Items)
		if err != nil {
			return nil, fmt.Errorf("failed to update moving items: %v", err)
		}
	}

	// Update moving-specific details
	if req.HelpersAssigned != nil || req.EstimatedDuration != nil {
		err = s.updateMovingDeliveryDetails(ctx, deliveryID, req.HelpersAssigned, req.EstimatedDuration)
		if err != nil {
			return nil, fmt.Errorf("failed to update moving details: %v", err)
		}
	}

	// Create tracking entry if status changed
	if req.Status != nil {
		err = s.createTrackingEntry(ctx, deliveryID, *req.Status, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create tracking entry: %v", err)
		}
	}

	return s.getDelivery(ctx, deliveryID)
}

// UpdateDeliveryStatus updates only the delivery status with tracking
func (s *UpdateService) UpdateDeliveryStatus(deliveryID string, req *models.StatusUpdateRequest) (*models.Delivery, error) {
	ctx := context.Background()

	// Get current delivery to validate transition
	currentDelivery, err := s.getDelivery(ctx, deliveryID)
	if err != nil {
		return nil, err
	}

	// Validate status transition
	if !currentDelivery.CanTransitionTo(req.Status) {
		return nil, fmt.Errorf("invalid status transition from %s to %s", currentDelivery.Status, req.Status)
	}

	// Update delivery status
	_, err = db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Update(
		prismadb.Delivery.Status.Set(prismadb.DeliveryStatus(req.Status)),
	).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update delivery status: %v", err)
	}

	// Create tracking entry
	err = s.createTrackingEntry(ctx, deliveryID, req.Status, req.Location, req.Notes)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracking entry: %v", err)
	}

	return s.getDelivery(ctx, deliveryID)
}

// Helper methods

func (s *UpdateService) getDelivery(ctx context.Context, deliveryID string) (*models.Delivery, error) {
	delivery, err := db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Exec(ctx)
	if err != nil {
		if err == prismadb.ErrNotFound {
			return nil, fmt.Errorf("delivery not found")
		}
		return nil, fmt.Errorf("failed to get delivery: %v", err)
	}

	// Convert to models.Delivery
	deliveryModel := &models.Delivery{
		ID:        delivery.ID,
		ClientID:  delivery.ClientPhone,
		Status:    models.DeliveryStatus(delivery.Status),
		Type:      models.DeliveryType(delivery.Type),
		CreatedAt: delivery.CreatedAt,
		UpdatedAt: delivery.UpdatedAt,
		PickupID:  delivery.PickupLocationID,
		DropoffID: delivery.DropoffLocationID,
	}

	// Handle nullable fields
	if driverID, ok := delivery.DriverID(); ok {
		deliveryModel.LivreurID = &driverID
	}
	if totalPrice, ok := delivery.TotalPrice(); ok {
		deliveryModel.FinalPrice = totalPrice
	}

	return deliveryModel, nil
}

func (s *UpdateService) updateDeliveryLocation(ctx context.Context, locationID string, address *string, lat, lng *float64) error {
	updateParams := []prismadb.LocationSetParam{}

	if address != nil {
		updateParams = append(updateParams, prismadb.Location.Address.Set(*address))
	}
	if lat != nil {
		updateParams = append(updateParams, prismadb.Location.Lat.Set(*lat))
	}
	if lng != nil {
		updateParams = append(updateParams, prismadb.Location.Lng.Set(*lng))
	}

	if len(updateParams) == 0 {
		return nil // Nothing to update
	}

	_, err := db.PrismaDB.Location.FindUnique(
		prismadb.Location.ID.Equals(locationID),
	).Update(updateParams...).Exec(ctx)

	return err
}

func (s *UpdateService) updatePackage(ctx context.Context, deliveryID string, packageInfo *models.PackageInfo) error {
	// For now, we'll skip package updates as the Prisma API needs to be adjusted
	// This would require finding the package first, then updating it
	log.Println("Package update skipped - API adjustment needed")
	return nil
}

func (s *UpdateService) updateExpressDeliveryDetails(ctx context.Context, deliveryID string, expressTime *int, priority *string) error {
	// Update express delivery details in related table if exists
	// This would require an ExpressDelivery table in schema
	// For now, we'll just return nil as these might be stored in delivery table or elsewhere
	return nil
}

func (s *UpdateService) updateGroupedZones(ctx context.Context, deliveryID string, zones []models.GroupedZone) error {
	// TODO: Implémenter la gestion des zones groupées
	// Le modèle GroupedZone n'existe pas encore dans le schéma Prisma
	log.Printf("Grouped zones update skipped - model not implemented: %d zones", len(zones))
	return nil
}

func (s *UpdateService) updateGroupedDeliveryDetails(ctx context.Context, deliveryID string, completedZones *int) error {
	if completedZones == nil {
		return nil
	}

	// For now, skip GroupedDelivery updates as the API needs adjustment
	log.Printf("GroupedDelivery update skipped for delivery %s - API adjustment needed", deliveryID)
	return nil
}

func (s *UpdateService) updateMovingItems(ctx context.Context, deliveryID string, items []models.MovingItem) error {
	// TODO: Implémenter la gestion des items de déménagement
	// Le modèle MovingItem n'existe pas encore dans le schéma Prisma
	log.Printf("Moving items update skipped - model not implemented: %d items", len(items))
	return nil
}

func (s *UpdateService) updateMovingDeliveryDetails(ctx context.Context, deliveryID string, helpersAssigned, estimatedDuration *int) error {
	// For now, skip MovingService updates as the API needs adjustment
	log.Printf("MovingService update skipped for delivery %s - API adjustment needed", deliveryID)
	return nil
}

func (s *UpdateService) createTrackingEntry(ctx context.Context, deliveryID string, status models.DeliveryStatus, location *models.Location, notes *string) error {
	// Create tracking according to Prisma schema (status, location optionnel, deliveryId)
	if location != nil {
		_, err := db.PrismaDB.Tracking.CreateOne(
			prismadb.Tracking.Status.Set(string(status)),
			prismadb.Tracking.Delivery.Link(prismadb.Delivery.ID.Equals(deliveryID)),
			prismadb.Tracking.Location.Set(location.Address),
		).Exec(ctx)
		return err
	} else {
		_, err := db.PrismaDB.Tracking.CreateOne(
			prismadb.Tracking.Status.Set(string(status)),
			prismadb.Tracking.Delivery.Link(prismadb.Delivery.ID.Equals(deliveryID)),
		).Exec(ctx)
		return err
	}
	// Note: Le champ 'notes' est ignoré car il n'existe pas dans le schéma Prisma
}
