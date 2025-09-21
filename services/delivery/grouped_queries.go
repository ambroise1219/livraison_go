package delivery

import (
	"context"
	"fmt"
	"time"

	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

// GroupedQueriesService gère les requêtes pour les livraisons groupées
type GroupedQueriesService struct{}

// NewGroupedQueriesService crée une nouvelle instance du service de requêtes groupées
func NewGroupedQueriesService() *GroupedQueriesService {
	return &GroupedQueriesService{}
}

// GetGroupedDeliveries récupère les livraisons groupées d'un client
func (s *GroupedQueriesService) GetGroupedDeliveries(clientPhone string) ([]*models.Delivery, error) {
	ctx := context.Background()

	deliveries, err := db.PrismaDB.Delivery.FindMany(
		prismadb.Delivery.ClientPhone.Equals(clientPhone),
		prismadb.Delivery.Type.Equals(prismadb.DeliveryTypeGrouped),
	).OrderBy(
		prismadb.Delivery.CreatedAt.Order(prismadb.SortOrderDesc),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("erreur de récupération des livraisons groupées: %v", err)
	}

	var result []*models.Delivery
	for _, delivery := range deliveries {
		deliveryModel := s.convertPrismaDeliveryToModel(&delivery)
		result = append(result, deliveryModel)
	}

	return result, nil
}

// GetGroupedDeliveryByID récupère une livraison groupée par son ID
func (s *GroupedQueriesService) GetGroupedDeliveryByID(deliveryID string) (*models.Delivery, error) {
	ctx := context.Background()

	delivery, err := db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Exec(ctx)

	if err != nil {
		if err == prismadb.ErrNotFound {
			return nil, fmt.Errorf("livraison groupée non trouvée")
		}
		return nil, fmt.Errorf("erreur de récupération: %v", err)
	}

	// Vérifier que c'est bien une livraison groupée
	if delivery.Type != prismadb.DeliveryTypeGrouped {
		return nil, fmt.Errorf("cette livraison n'est pas de type groupé")
	}

	return s.convertPrismaDeliveryToModel(delivery), nil
}

// GetGroupedDeliveriesByStatus récupère les livraisons groupées par statut
func (s *GroupedQueriesService) GetGroupedDeliveriesByStatus(status models.DeliveryStatus) ([]*models.Delivery, error) {
	ctx := context.Background()

	deliveries, err := db.PrismaDB.Delivery.FindMany(
		prismadb.Delivery.Type.Equals(prismadb.DeliveryTypeGrouped),
		prismadb.Delivery.Status.Equals(prismadb.DeliveryStatus(status)),
	).OrderBy(
		prismadb.Delivery.CreatedAt.Order(prismadb.SortOrderDesc),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("erreur de récupération des livraisons par statut: %v", err)
	}

	var result []*models.Delivery
	for _, delivery := range deliveries {
		deliveryModel := s.convertPrismaDeliveryToModel(&delivery)
		result = append(result, deliveryModel)
	}

	return result, nil
}

// GetGroupedDeliveryStats récupère les statistiques des livraisons groupées
func (s *GroupedQueriesService) GetGroupedDeliveryStats() (map[string]interface{}, error) {
	ctx := context.Background()

	// Compter le nombre total de livraisons groupées
	deliveries, err := db.PrismaDB.Delivery.FindMany(
		prismadb.Delivery.Type.Equals(prismadb.DeliveryTypeGrouped),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("erreur de récupération des statistiques: %v", err)
	}

	total := len(deliveries)
	var revenue float64
	var totalDistance float64
	var validPrices int
	var validDistances int

	for _, delivery := range deliveries {
		if totalPrice, ok := delivery.TotalPrice(); ok {
			revenue += totalPrice
			validPrices++
		}
		if distance, ok := delivery.DistanceKm(); ok {
			totalDistance += distance
			validDistances++
		}
	}

	stats := map[string]interface{}{
		"total":   total,
		"revenue": revenue,
	}

	if validPrices > 0 {
		stats["avgPrice"] = revenue / float64(validPrices)
	} else {
		stats["avgPrice"] = 0.0
	}

	if validDistances > 0 {
		stats["avgDistance"] = totalDistance / float64(validDistances)
	} else {
		stats["avgDistance"] = 0.0
	}

	return stats, nil
}

// GetGroupedDeliveriesByDateRange récupère les livraisons groupées dans une plage de dates
func (s *GroupedQueriesService) GetGroupedDeliveriesByDateRange(startDate, endDate time.Time) ([]*models.Delivery, error) {
	ctx := context.Background()

	deliveries, err := db.PrismaDB.Delivery.FindMany(
		prismadb.Delivery.Type.Equals(prismadb.DeliveryTypeGrouped),
		prismadb.Delivery.CreatedAt.Gte(startDate),
		prismadb.Delivery.CreatedAt.Lte(endDate),
	).OrderBy(
		prismadb.Delivery.CreatedAt.Order(prismadb.SortOrderDesc),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("erreur de récupération par plage de dates: %v", err)
	}

	var result []*models.Delivery
	for _, delivery := range deliveries {
		deliveryModel := s.convertPrismaDeliveryToModel(&delivery)
		result = append(result, deliveryModel)
	}

	return result, nil
}

// UpdateGroupedDelivery met à jour une livraison groupée
func (s *GroupedQueriesService) UpdateGroupedDelivery(deliveryID string, updates *models.UpdateDeliveryRequest) (*models.Delivery, error) {
	ctx := context.Background()

	// Vérifier que la livraison existe et est de type groupé
	_, err := s.GetGroupedDeliveryByID(deliveryID)
	if err != nil {
		return nil, err
	}

	// Préparer les paramètres de mise à jour
	var updateParams []prismadb.DeliverySetParam

	if updates.Status != nil {
		updateParams = append(updateParams, prismadb.Delivery.Status.Set(prismadb.DeliveryStatus(*updates.Status)))
	}
	if updates.LivreurID != nil {
		updateParams = append(updateParams, prismadb.Delivery.DriverID.Set(*updates.LivreurID))
	}
	if updates.WaitingMin != nil {
		updateParams = append(updateParams, prismadb.Delivery.WaitingMin.Set(int(*updates.WaitingMin)))
	}
	if updates.FinalPrice != nil {
		updateParams = append(updateParams, prismadb.Delivery.TotalPrice.Set(*updates.FinalPrice))
	}

	// Toujours mettre à jour le timestamp
	updateParams = append(updateParams, prismadb.Delivery.UpdatedAt.Set(time.Now()))

	// Effectuer la mise à jour
	_, err = db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Update(updateParams...).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("échec de la mise à jour: %v", err)
	}

	// Récupérer la livraison mise à jour
	return s.GetGroupedDeliveryByID(deliveryID)
}

// UpdateGroupedDeliveryStatus met à jour le statut d'une livraison groupée
func (s *GroupedQueriesService) UpdateGroupedDeliveryStatus(deliveryID string, status models.DeliveryStatus, updatedBy string) (*models.Delivery, error) {
	ctx := context.Background()

	// Mettre à jour le statut
	_, err := db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Update(
		prismadb.Delivery.Status.Set(prismadb.DeliveryStatus(status)),
		prismadb.Delivery.UpdatedAt.Set(time.Now()),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("échec de la mise à jour du statut: %v", err)
	}

	// Récupérer la livraison mise à jour
	return s.GetGroupedDeliveryByID(deliveryID)
}

// DeleteGroupedDelivery supprime une livraison groupée
func (s *GroupedQueriesService) DeleteGroupedDelivery(deliveryID string) error {
	ctx := context.Background()

	// Vérifier que la livraison existe
	_, err := s.GetGroupedDeliveryByID(deliveryID)
	if err != nil {
		return err
	}

	// Supprimer la livraison
	_, err = db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Delete().Exec(ctx)

	if err != nil {
		return fmt.Errorf("échec de la suppression: %v", err)
	}

	return nil
}

// convertPrismaDeliveryToModel convertit une livraison Prisma en modèle
func (s *GroupedQueriesService) convertPrismaDeliveryToModel(delivery *prismadb.DeliveryModel) *models.Delivery {
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
