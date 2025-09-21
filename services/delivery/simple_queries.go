package delivery

import (
	"context"
	"fmt"
	"time"

	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

// SimpleQueriesService gère les requêtes pour les livraisons simples
type SimpleQueriesService struct{}

// NewSimpleQueriesService crée une nouvelle instance du service de requêtes simples
func NewSimpleQueriesService() *SimpleQueriesService {
	return &SimpleQueriesService{}
}

// GetSimpleDeliveries récupère les livraisons simples d'un client
func (s *SimpleQueriesService) GetSimpleDeliveries(clientPhone string) ([]*models.Delivery, error) {
	ctx := context.Background()

	deliveries, err := db.PrismaDB.Delivery.FindMany(
		prismadb.Delivery.ClientPhone.Equals(clientPhone),
		prismadb.Delivery.Type.Equals(prismadb.DeliveryTypeStandard),
	).OrderBy(
		prismadb.Delivery.CreatedAt.Order(prismadb.SortOrderDesc),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("erreur de récupération des livraisons simples: %v", err)
	}

	var result []*models.Delivery
	for _, delivery := range deliveries {
		deliveryModel := s.convertPrismaDeliveryToModel(&delivery)
		result = append(result, deliveryModel)
	}

	return result, nil
}

// GetSimpleDeliveryByID récupère une livraison simple par son ID
func (s *SimpleQueriesService) GetSimpleDeliveryByID(deliveryID string) (*models.Delivery, error) {
	ctx := context.Background()

	delivery, err := db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Exec(ctx)

	if err != nil {
		if err == prismadb.ErrNotFound {
			return nil, fmt.Errorf("livraison simple non trouvée")
		}
		return nil, fmt.Errorf("erreur de récupération: %v", err)
	}

	// Vérifier que c'est bien une livraison simple
	if delivery.Type != prismadb.DeliveryTypeStandard {
		return nil, fmt.Errorf("cette livraison n'est pas de type standard")
	}

	return s.convertPrismaDeliveryToModel(delivery), nil
}

// GetSimpleDeliveriesByStatus récupère les livraisons simples par statut
func (s *SimpleQueriesService) GetSimpleDeliveriesByStatus(status models.DeliveryStatus) ([]*models.Delivery, error) {
	ctx := context.Background()

	deliveries, err := db.PrismaDB.Delivery.FindMany(
		prismadb.Delivery.Type.Equals(prismadb.DeliveryTypeStandard),
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

// GetSimpleDeliveryStats récupère les statistiques des livraisons simples
func (s *SimpleQueriesService) GetSimpleDeliveryStats() (map[string]interface{}, error) {
	ctx := context.Background()

	// Compter le nombre total de livraisons simples
	deliveries, err := db.PrismaDB.Delivery.FindMany(
		prismadb.Delivery.Type.Equals(prismadb.DeliveryTypeStandard),
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

// GetSimpleDeliveriesByDateRange récupère les livraisons simples dans une plage de dates
func (s *SimpleQueriesService) GetSimpleDeliveriesByDateRange(startDate, endDate time.Time) ([]*models.Delivery, error) {
	ctx := context.Background()

	deliveries, err := db.PrismaDB.Delivery.FindMany(
		prismadb.Delivery.Type.Equals(prismadb.DeliveryTypeStandard),
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

// UpdateSimpleDelivery met à jour une livraison simple
func (s *SimpleQueriesService) UpdateSimpleDelivery(deliveryID string, updates *models.UpdateDeliveryRequest) (*models.Delivery, error) {
	ctx := context.Background()

	// Vérifier que la livraison existe et est de type simple
	_, err := s.GetSimpleDeliveryByID(deliveryID)
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
	return s.GetSimpleDeliveryByID(deliveryID)
}

// UpdateSimpleDeliveryStatus met à jour le statut d'une livraison simple
func (s *SimpleQueriesService) UpdateSimpleDeliveryStatus(deliveryID string, status models.DeliveryStatus, updatedBy string) (*models.Delivery, error) {
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
	return s.GetSimpleDeliveryByID(deliveryID)
}

// DeleteSimpleDelivery supprime une livraison simple
func (s *SimpleQueriesService) DeleteSimpleDelivery(deliveryID string) error {
	ctx := context.Background()

	// Vérifier que la livraison existe
	_, err := s.GetSimpleDeliveryByID(deliveryID)
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
func (s *SimpleQueriesService) convertPrismaDeliveryToModel(delivery *prismadb.DeliveryModel) *models.Delivery {
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
