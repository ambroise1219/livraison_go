package delivery

import (
	"context"
	"fmt"

	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

// ExpressQueriesService gère les requêtes pour les livraisons express
type ExpressQueriesService struct{}

// NewExpressQueriesService crée une nouvelle instance du service de requêtes express
func NewExpressQueriesService() *ExpressQueriesService {
	return &ExpressQueriesService{}
}

// GetExpressDeliveries récupère les livraisons express d'un client
func (s *ExpressQueriesService) GetExpressDeliveries(clientID string) ([]*models.Delivery, error) {
	ctx := context.Background()

	deliveries, err := db.PrismaDB.Delivery.FindMany(
		prismadb.Delivery.ClientPhone.Equals(clientID),
		prismadb.Delivery.Type.Equals(prismadb.DeliveryTypeExpress),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("erreur de récupération des livraisons express: %v", err)
	}

	var result []*models.Delivery
	for _, delivery := range deliveries {
		deliveryModel := s.convertPrismaDeliveryToModel(&delivery)
		result = append(result, deliveryModel)
	}

	return result, nil
}

// convertPrismaDeliveryToModel convertit une livraison Prisma en modèle
func (s *ExpressQueriesService) convertPrismaDeliveryToModel(delivery *prismadb.DeliveryModel) *models.Delivery {
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

	return deliveryModel
}

// GetExpressDeliveryByID récupère une livraison express par son ID
func (s *ExpressQueriesService) GetExpressDeliveryByID(deliveryID string) (*models.Delivery, error) {
	ctx := context.Background()

	// Chercher d'abord la livraison par ID
	delivery, err := db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Exec(ctx)

	if err != nil {
		if err == prismadb.ErrNotFound {
			return nil, fmt.Errorf("livraison non trouvée")
		}
		return nil, fmt.Errorf("erreur de récupération: %v", err)
	}

	// Vérifier que c'est bien une livraison express
	if delivery.Type != prismadb.DeliveryTypeExpress {
		return nil, fmt.Errorf("cette livraison n'est pas de type express")
	}

	return s.convertPrismaDeliveryToModel(delivery), nil
}