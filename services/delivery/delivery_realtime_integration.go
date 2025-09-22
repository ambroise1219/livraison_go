package delivery

import (
	"context"
	"fmt"
	"time"

	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	"github.com/ambroise1219/livraison_go/services"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

// RealtimeDeliveryService intègre les livraisons avec les notifications temps réel
type RealtimeDeliveryService struct {
	deliveryService *DeliveryService
	realtimeService *services.RealtimeService
}

// NewRealtimeDeliveryService crée un nouveau service intégré
func NewRealtimeDeliveryService() *RealtimeDeliveryService {
	return &RealtimeDeliveryService{
		deliveryService: NewDeliveryService(),
		realtimeService: services.NewRealtimeService(),
	}
}

// UpdateDeliveryStatusWithRealtime met à jour le statut d'une livraison avec notifications temps réel
func (rds *RealtimeDeliveryService) UpdateDeliveryStatusWithRealtime(deliveryID string, newStatus models.DeliveryStatus, message string, driverInfo *DriverInfo) (*models.DeliveryResponse, error) {
	ctx := context.Background()

	// Récupérer la livraison actuelle
	currentDelivery, err := db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("livraison non trouvée: %v", err)
	}

	// Valider la transition de statut
	if !rds.isValidStatusTransition(models.DeliveryStatus(currentDelivery.Status), newStatus) {
		return nil, fmt.Errorf("transition de statut invalide de %s vers %s", currentDelivery.Status, newStatus)
	}

	// Mettre à jour le statut dans la base de données
	updatedDelivery, err := db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Update(
		prismadb.Delivery.Status.Set(prismadb.DeliveryStatus(newStatus)),
		prismadb.Delivery.UpdatedAt.Set(time.Now()),
	).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("échec de la mise à jour: %v", err)
	}

	// Créer l'entrée de tracking
	if err := rds.createTrackingEntry(deliveryID, string(newStatus), message); err != nil {
		fmt.Printf("Erreur création tracking: %v\n", err)
	}

	// Préparer la notification temps réel
	deliveryUpdate := models.DeliveryUpdate{
		DeliveryID:    deliveryID,
		Status:        string(newStatus),
		Message:       message,
		Timestamp:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Ajouter les informations du livreur si disponibles
	if driverInfo != nil {
		deliveryUpdate.DriverID = driverInfo.ID
		deliveryUpdate.DriverName = driverInfo.Name
		deliveryUpdate.DriverPhone = driverInfo.Phone
		deliveryUpdate.Latitude = driverInfo.Latitude
		deliveryUpdate.Longitude = driverInfo.Longitude
	}

	// Publier la mise à jour en temps réel
	if err := rds.realtimeService.PublishDeliveryUpdate(deliveryID, deliveryUpdate); err != nil {
		fmt.Printf("Erreur publication temps réel: %v\n", err)
	}

	// Envoyer des notifications spécifiques selon le statut
	if err := rds.sendStatusNotifications(updatedDelivery, newStatus, message); err != nil {
		fmt.Printf("Erreur envoi notifications: %v\n", err)
	}

	// Convertir en réponse
	response := rds.convertToResponse(updatedDelivery)
	return response, nil
}

// AssignDriverWithRealtime assigne un livreur avec notifications temps réel
func (rds *RealtimeDeliveryService) AssignDriverWithRealtime(deliveryID, driverID string, driverInfo *DriverInfo) (*models.DeliveryResponse, error) {
	ctx := context.Background()

	// Mettre à jour la livraison avec le livreur
	updatedDelivery, err := db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Update(
		prismadb.Delivery.DriverID.Set(driverID),
		prismadb.Delivery.Status.Set(prismadb.DeliveryStatusAssigned),
		prismadb.Delivery.AssignedAt.Set(time.Now()),
	).Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("échec de l'assignation: %v", err)
	}

	// Notification temps réel pour l'assignation
	deliveryUpdate := models.DeliveryUpdate{
		DeliveryID:    deliveryID,
		Status:        string(models.DeliveryStatusAssigned),
		DriverID:      driverID,
		Message:       fmt.Sprintf("Livreur assigné: %s", driverInfo.Name),
		Timestamp:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if driverInfo != nil {
		deliveryUpdate.DriverName = driverInfo.Name
		deliveryUpdate.DriverPhone = driverInfo.Phone
	}

	// Publier l'assignation
	if err := rds.realtimeService.PublishDeliveryUpdate(deliveryID, deliveryUpdate); err != nil {
		fmt.Printf("Erreur publication assignation: %v\n", err)
	}

	// Notification au client
	clientNotification := models.Notification{
		Type:    models.NotificationTypeDeliveryUpdate,
		Title:   "Livreur assigné",
		Message: fmt.Sprintf("Un livreur a été assigné à votre commande: %s", driverInfo.Name),
	}

	if err := rds.realtimeService.PublishNotification(updatedDelivery.ClientPhone, clientNotification); err != nil {
		fmt.Printf("Erreur notification client: %v\n", err)
	}

	response := rds.convertToResponse(updatedDelivery)
	return response, nil
}

// UpdateDriverLocationWithRealtime met à jour la position du livreur avec diffusion temps réel
func (rds *RealtimeDeliveryService) UpdateDriverLocationWithRealtime(deliveryID, driverID string, location models.LocationUpdate) error {
	// Enrichir les informations de position
	location.DriverID = driverID
	location.DeliveryID = deliveryID
	location.Timestamp = time.Now()
	location.UpdatedAt = time.Now()

	// Mettre en cache la position
	if err := rds.realtimeService.SetDriverLocation(driverID, location); err != nil {
		return fmt.Errorf("erreur cache position: %v", err)
	}

	// Publier la mise à jour de position
	if err := rds.realtimeService.PublishLocationUpdate(deliveryID, location); err != nil {
		return fmt.Errorf("erreur publication position: %v", err)
	}

	// Calculer et envoyer l'ETA si nécessaire
	if err := rds.calculateAndSendETA(deliveryID, location); err != nil {
		fmt.Printf("Erreur calcul ETA: %v\n", err)
	}

	return nil
}

// isValidStatusTransition valide les transitions de statut
func (rds *RealtimeDeliveryService) isValidStatusTransition(current, new models.DeliveryStatus) bool {
	validTransitions := map[models.DeliveryStatus][]models.DeliveryStatus{
		models.DeliveryStatusPending:   {models.DeliveryStatusAssigned, models.DeliveryStatusCanceled},
		models.DeliveryStatusAssigned:  {models.DeliveryStatusInProgress, models.DeliveryStatusCanceled},
		models.DeliveryStatusInProgress: {models.DeliveryStatusDelivered, models.DeliveryStatusCanceled},
		models.DeliveryStatusDelivered: {}, // État final
		models.DeliveryStatusCanceled:  {}, // État final
	}

	allowedStatuses, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, allowedStatus := range allowedStatuses {
		if allowedStatus == new {
			return true
		}
	}
	return false
}

// createTrackingEntry crée une entrée de suivi
func (rds *RealtimeDeliveryService) createTrackingEntry(deliveryID, status, notes string) error {
	ctx := context.Background()

	_, err := db.PrismaDB.Tracking.CreateOne(
		prismadb.Tracking.Status.Set(status),
		prismadb.Tracking.Delivery.Link(prismadb.Delivery.ID.Equals(deliveryID)),
	).Exec(ctx)

	return err
}

// sendStatusNotifications envoie les notifications spécifiques au statut
func (rds *RealtimeDeliveryService) sendStatusNotifications(delivery *prismadb.DeliveryModel, status models.DeliveryStatus, message string) error {
	var notification models.Notification

	switch status {
	case models.DeliveryStatusAssigned:
		notification = models.Notification{
			Type:    models.NotificationTypeDeliveryUpdate,
			Title:   "Livreur assigné",
			Message: "Un livreur a été assigné à votre commande",
		}
	case models.DeliveryStatusInProgress:
		notification = models.Notification{
			Type:    models.NotificationTypeDeliveryUpdate,
			Title:   "Livraison en cours",
			Message: "Votre commande est en cours de livraison",
		}
	case models.DeliveryStatusDelivered:
		notification = models.Notification{
			Type:    models.NotificationTypeDeliveryUpdate,
			Title:   "Livraison terminée",
			Message: "Votre commande a été livrée avec succès",
		}
		// Supprimer de la liste des livraisons actives
		rds.realtimeService.RemoveActiveDelivery(delivery.ID)
	case models.DeliveryStatusCanceled:
		notification = models.Notification{
			Type:    models.NotificationTypeDeliveryUpdate,
			Title:   "Livraison annulée",
			Message: "Votre commande a été annulée",
		}
		// Supprimer de la liste des livraisons actives
		rds.realtimeService.RemoveActiveDelivery(delivery.ID)
	default:
		return nil // Pas de notification pour ce statut
	}

	return rds.realtimeService.PublishNotification(delivery.ClientPhone, notification)
}

// calculateAndSendETA calcule et envoie l'ETA mis à jour
func (rds *RealtimeDeliveryService) calculateAndSendETA(deliveryID string, driverLocation models.LocationUpdate) error {
	ctx := context.Background()

	// Récupérer les informations de la livraison
	delivery, err := db.PrismaDB.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(deliveryID),
	).Exec(ctx)
	if err != nil {
		return err
	}

	// Récupérer la position de destination
	dropoffLocation := delivery.DropoffLocation()
	if dropoffLocation == nil {
		return fmt.Errorf("destination non trouvée")
	}

	dropoffLat := dropoffLocation.Lat
	dropoffLng := dropoffLocation.Lng

	// Calculer la distance
	distance := rds.realtimeService.CalculateDistance(
		driverLocation.Latitude, driverLocation.Longitude,
		dropoffLat, dropoffLng,
	)

	// Calculer l'ETA
	eta := rds.realtimeService.CalculateEstimatedTime(distance, 30) // 30 km/h

	// Publier l'ETA mis à jour
	etaUpdate := models.DeliveryUpdate{
		DeliveryID:      deliveryID,
		Status:          string(delivery.Status),
		Latitude:        driverLocation.Latitude,
		Longitude:       driverLocation.Longitude,
		EstimatedTime:   eta,
		Distance:        distance,
		Message:         fmt.Sprintf("ETA mis à jour: %d minutes", eta),
		Timestamp:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return rds.realtimeService.PublishDeliveryUpdate(deliveryID, etaUpdate)
}

// convertToResponse convertit un modèle Prisma en réponse API
func (rds *RealtimeDeliveryService) convertToResponse(delivery *prismadb.DeliveryModel) *models.DeliveryResponse {
	response := &models.DeliveryResponse{
		ID:            delivery.ID,
		ClientID:      delivery.ClientPhone,
		Status:        models.DeliveryStatus(delivery.Status),
		Type:          models.DeliveryType(delivery.Type),
		CreatedAt:     delivery.CreatedAt,
		UpdatedAt:     delivery.UpdatedAt,
	}

	// Ajouter les champs optionnels
	if driverID, ok := delivery.DriverID(); ok {
		response.LivreurID = &driverID
	}
	if totalPrice, ok := delivery.TotalPrice(); ok {
		response.FinalPrice = totalPrice
	}

	// Ajouter les locations si disponibles
	if pickup := delivery.PickupLocation(); pickup != nil {
		response.Pickup = &models.Location{
			ID:      pickup.ID,
			Address: pickup.Address,
		}
		response.Pickup.Lat = &pickup.Lat
		response.Pickup.Lng = &pickup.Lng
	}

	if dropoff := delivery.DropoffLocation(); dropoff != nil {
		response.Dropoff = &models.Location{
			ID:      dropoff.ID,
			Address: dropoff.Address,
		}
		response.Dropoff.Lat = &dropoff.Lat
		response.Dropoff.Lng = &dropoff.Lng
	}

	return response
}

// DriverInfo représente les informations d'un livreur
type DriverInfo struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Phone     string  `json:"phone"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}