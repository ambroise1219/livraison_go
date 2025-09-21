package delivery

import (
	"context"
	"fmt"

	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

// DeliveryService gère les opérations sur les livraisons
type DeliveryService struct{}

// NewDeliveryService crée une nouvelle instance du service de livraison
func NewDeliveryService() *DeliveryService {
	return &DeliveryService{}
}

// CreateDelivery crée une nouvelle livraison
func (s *DeliveryService) CreateDelivery(delivery *models.Delivery) error {
	ctx := context.Background()

	// Création directe avec champs obligatoires seulement
	createdDelivery, err := db.PrismaDB.Delivery.CreateOne(
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

	// Mise à jour de l'ID dans l'objet delivery
	delivery.ID = createdDelivery.ID

	return nil
}

// GetDelivery récupère une livraison par son ID
func (s *DeliveryService) GetDelivery(deliveryID string) (*models.Delivery, error) {
	ctx := context.Background()

	delivery, err := db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).With(
		prismadb.Delivery.PickupLocation.Fetch(),
		prismadb.Delivery.DropoffLocation.Fetch(),
	).Exec(ctx)

	if err != nil {
		if err == prismadb.ErrNotFound {
			return nil, fmt.Errorf("livraison non trouvée")
		}
		return nil, fmt.Errorf("erreur de récupération: %v", err)
	}

	// Convertir la livraison Prisma en modèle
	deliveryModel := &models.Delivery{
		ID:        delivery.ID,
		ClientID:  delivery.ClientPhone,
		Status:    models.DeliveryStatus(delivery.Status),
		Type:      models.DeliveryType(delivery.Type),
		CreatedAt:   delivery.CreatedAt,
		UpdatedAt:   delivery.UpdatedAt,
		PickupID:    delivery.PickupLocationID,
		DropoffID:   delivery.DropoffLocationID,
	}

	// Handle nullable fields
	if driverID, ok := delivery.DriverID(); ok {
		deliveryModel.LivreurID = &driverID
	}
	if distanceKm, ok := delivery.DistanceKm(); ok {
		deliveryModel.DistanceKm = &distanceKm
	}
	if durationMin, ok := delivery.DurationMin(); ok {
		durationFloat := float64(durationMin)
		deliveryModel.DurationMin = &durationFloat
	}
	if basePrice, ok := delivery.BasePrice(); ok {
		deliveryModel.BasePrice = &basePrice
	}
	if waitingMin, ok := delivery.WaitingMin(); ok {
		waitingFloat := float64(waitingMin)
		deliveryModel.WaitingMin = &waitingFloat
	}
	if paidAt, ok := delivery.PaidAt(); ok {
		deliveryModel.PaidAt = &paidAt
	}
	if paymentMethod, ok := delivery.PaymentMethod(); ok {
		deliveryModel.PaymentMethod = models.PaymentMethod(paymentMethod)
	}
	if totalPrice, ok := delivery.TotalPrice(); ok {
		deliveryModel.FinalPrice = totalPrice
	}

	return deliveryModel, nil
}

// UpdateDelivery met à jour une livraison
func (s *DeliveryService) UpdateDelivery(delivery *models.Delivery) error {
	ctx := context.Background()

	// Préparation des paramètres de mise à jour
	updateParams := []prismadb.DeliverySetParam{
		prismadb.Delivery.Status.Set(prismadb.DeliveryStatus(delivery.Status)),
	}

	// Ajout conditionnel des champs optionnels
	if delivery.LivreurID != nil {
		updateParams = append(updateParams, prismadb.Delivery.DriverID.Set(*delivery.LivreurID))
	}
	if delivery.DistanceKm != nil {
		updateParams = append(updateParams, prismadb.Delivery.DistanceKm.Set(*delivery.DistanceKm))
	}
	if delivery.DurationMin != nil {
		updateParams = append(updateParams, prismadb.Delivery.DurationMin.Set(int(*delivery.DurationMin)))
	}
	if delivery.BasePrice != nil {
		updateParams = append(updateParams, prismadb.Delivery.BasePrice.Set(*delivery.BasePrice))
	}
	if delivery.WaitingMin != nil {
		updateParams = append(updateParams, prismadb.Delivery.WaitingMin.Set(int(*delivery.WaitingMin)))
	}
	if delivery.PaidAt != nil {
		updateParams = append(updateParams, prismadb.Delivery.PaidAt.Set(*delivery.PaidAt))
	}
	if delivery.PaymentMethod != "" {
		updateParams = append(updateParams, prismadb.Delivery.PaymentMethod.Set(prismadb.PaymentMethod(delivery.PaymentMethod)))
	}
	if delivery.FinalPrice != 0 {
		updateParams = append(updateParams, prismadb.Delivery.TotalPrice.Set(delivery.FinalPrice))
	}

	_, err := db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(delivery.ID),
	).Update(updateParams...).Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update delivery: %v", err)
	}

	return nil
}

// DeleteDelivery supprime une livraison
func (s *DeliveryService) DeleteDelivery(deliveryID string) error {
	ctx := context.Background()

	_, err := db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Delete().Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete delivery: %v", err)
	}

	return nil
}

// GetDeliveriesByClient récupère les livraisons d'un client par téléphone
func (s *DeliveryService) GetDeliveriesByClient(clientPhone string) ([]models.Delivery, error) {
	ctx := context.Background()

	deliveries, err := db.PrismaDB.Delivery.FindMany(
		prismadb.Delivery.ClientPhone.Equals(clientPhone),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get client deliveries: %v", err)
	}

	var result []models.Delivery
	for _, delivery := range deliveries {
		// Convertir la livraison Prisma en modèle
		// Conversion avec gestion des champs optionnels
		deliveryModel := models.Delivery{
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
		if distanceKm, ok := delivery.DistanceKm(); ok {
			deliveryModel.DistanceKm = &distanceKm
		}
		if durationMin, ok := delivery.DurationMin(); ok {
			durationFloat := float64(durationMin)
			deliveryModel.DurationMin = &durationFloat
		}
		if basePrice, ok := delivery.BasePrice(); ok {
			deliveryModel.BasePrice = &basePrice
		}
		if waitingMin, ok := delivery.WaitingMin(); ok {
			waitingFloat := float64(waitingMin)
			deliveryModel.WaitingMin = &waitingFloat
		}
		if paidAt, ok := delivery.PaidAt(); ok {
			deliveryModel.PaidAt = &paidAt
		}
		if paymentMethod, ok := delivery.PaymentMethod(); ok {
			deliveryModel.PaymentMethod = models.PaymentMethod(paymentMethod)
		}
		if totalPrice, ok := delivery.TotalPrice(); ok {
			deliveryModel.FinalPrice = totalPrice
		}

		result = append(result, deliveryModel)
	}

	return result, nil
}

// GetDeliveriesByDriver récupère les livraisons d'un livreur
func (s *DeliveryService) GetDeliveriesByDriver(driverID string) ([]models.Delivery, error) {
	ctx := context.Background()

	deliveries, err := db.PrismaDB.Delivery.FindMany(
		prismadb.Delivery.DriverID.Equals(driverID),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get driver deliveries: %v", err)
	}

	var result []models.Delivery
	for _, delivery := range deliveries {
		// Convertir la livraison Prisma en modèle
		deliveryModel := models.Delivery{
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
		if distanceKm, ok := delivery.DistanceKm(); ok {
			deliveryModel.DistanceKm = &distanceKm
		}
		if durationMin, ok := delivery.DurationMin(); ok {
			durationFloat := float64(durationMin)
			deliveryModel.DurationMin = &durationFloat
		}
		if basePrice, ok := delivery.BasePrice(); ok {
			deliveryModel.BasePrice = &basePrice
		}
		if waitingMin, ok := delivery.WaitingMin(); ok {
			waitingFloat := float64(waitingMin)
			deliveryModel.WaitingMin = &waitingFloat
		}
		if paidAt, ok := delivery.PaidAt(); ok {
			deliveryModel.PaidAt = &paidAt
		}
		if paymentMethod, ok := delivery.PaymentMethod(); ok {
			deliveryModel.PaymentMethod = models.PaymentMethod(paymentMethod)
		}
		if totalPrice, ok := delivery.TotalPrice(); ok {
			deliveryModel.FinalPrice = totalPrice
		}

		result = append(result, deliveryModel)
	}

	return result, nil
}
