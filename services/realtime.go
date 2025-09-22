package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/models"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// RealtimeService gère les services temps réel
type RealtimeService struct {
	redisClient *redis.Client
	Ctx         context.Context
}

// NewRealtimeService crée une nouvelle instance du service temps réel
func NewRealtimeService() *RealtimeService {
	return &RealtimeService{
		redisClient: config.GetRedisClient(),
		Ctx:         context.Background(),
	}
}

// PublishDeliveryUpdate publie une mise à jour de livraison
func (rs *RealtimeService) PublishDeliveryUpdate(deliveryID string, update models.DeliveryUpdate) error {
	channel := fmt.Sprintf("delivery:%s", deliveryID)
	update.Timestamp = time.Now()

	data, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("erreur sérialisation update: %v", err)
	}

	return rs.redisClient.Publish(rs.Ctx, channel, string(data)).Err()
}

// PublishLocationUpdate publie une mise à jour de position
func (rs *RealtimeService) PublishLocationUpdate(deliveryID string, location models.LocationUpdate) error {
	channel := fmt.Sprintf("location:%s", deliveryID)
	location.Timestamp = time.Now()

	data, err := json.Marshal(location)
	if err != nil {
		return fmt.Errorf("erreur sérialisation location: %v", err)
	}

	return rs.redisClient.Publish(rs.Ctx, channel, string(data)).Err()
}

// PublishChatMessage publie un message de chat
func (rs *RealtimeService) PublishChatMessage(deliveryID string, message models.ChatMessage) error {
	channel := fmt.Sprintf("chat:%s", deliveryID)
	message.CreatedAt = time.Now()

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("erreur sérialisation message: %v", err)
	}

	return rs.redisClient.Publish(rs.Ctx, channel, string(data)).Err()
}

// PublishNotification publie une notification
func (rs *RealtimeService) PublishNotification(userID string, notification models.Notification) error {
	channel := fmt.Sprintf("notifications:%s", userID)
	notification.CreatedAt = time.Now()

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("erreur sérialisation notification: %v", err)
	}

	return rs.redisClient.Publish(rs.Ctx, channel, string(data)).Err()
}

// SubscribeToDeliveryUpdates s'abonne aux mises à jour d'une livraison
func (rs *RealtimeService) SubscribeToDeliveryUpdates(deliveryID string) *redis.PubSub {
	channel := fmt.Sprintf("delivery:%s", deliveryID)
	return rs.redisClient.Subscribe(rs.Ctx, channel)
}

// SubscribeToLocationUpdates s'abonne aux mises à jour de position
func (rs *RealtimeService) SubscribeToLocationUpdates(deliveryID string) *redis.PubSub {
	channel := fmt.Sprintf("location:%s", deliveryID)
	return rs.redisClient.Subscribe(rs.Ctx, channel)
}

// SubscribeToChatMessages s'abonne aux messages de chat
func (rs *RealtimeService) SubscribeToChatMessages(deliveryID string) *redis.PubSub {
	channel := fmt.Sprintf("chat:%s", deliveryID)
	return rs.redisClient.Subscribe(rs.Ctx, channel)
}

// SubscribeToNotifications s'abonne aux notifications d'un utilisateur
func (rs *RealtimeService) SubscribeToNotifications(userID string) *redis.PubSub {
	channel := fmt.Sprintf("notifications:%s", userID)
	return rs.redisClient.Subscribe(rs.Ctx, channel)
}

// CalculateDistance calcule la distance entre deux points GPS
func (rs *RealtimeService) CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Rayon de la Terre en km

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// CalculateEstimatedTime calcule le temps estimé d'arrivée
func (rs *RealtimeService) CalculateEstimatedTime(distance float64, speed float64) int {
	if speed <= 0 {
		speed = 30 // Vitesse moyenne en km/h
	}

	timeInHours := distance / speed
	timeInMinutes := int(timeInHours * 60)

	// Minimum 1 minute, maximum 120 minutes
	if timeInMinutes < 1 {
		timeInMinutes = 1
	} else if timeInMinutes > 120 {
		timeInMinutes = 120
	}

	return timeInMinutes
}

// GetRealtimeStats récupère les statistiques temps réel
func (rs *RealtimeService) GetRealtimeStats() (*models.RealtimeStats, error) {
	// Compter les connexions actives
	activeConnections, err := rs.redisClient.SCard(rs.Ctx, "active_connections").Result()
	if err != nil {
		logrus.Warnf("Erreur récupération connexions actives: %v", err)
		activeConnections = 0
	}

	// Compter les livraisons actives
	activeDeliveries, err := rs.redisClient.SCard(rs.Ctx, "active_deliveries").Result()
	if err != nil {
		logrus.Warnf("Erreur récupération livraisons actives: %v", err)
		activeDeliveries = 0
	}

	// Récupérer les métriques de performance
	info, err := rs.redisClient.Info(rs.Ctx, "memory").Result()
	if err != nil {
		logrus.Warnf("Erreur récupération info Redis: %v", err)
	}

	// Parser les métriques de mémoire (simplifié)
	memoryUsage := 0
	if info != "" {
		// Extraction simplifiée de la mémoire utilisée
		// Dans un vrai projet, on utiliserait un parser plus robuste
		memoryUsage = 100 // Placeholder
	}

	return &models.RealtimeStats{
		ActiveConnections: int(activeConnections),
		ActiveDeliveries:  int(activeDeliveries),
		MessagesPerSecond: 0,  // À implémenter avec des métriques
		AverageLatency:    50, // Placeholder
		MemoryUsage:       memoryUsage,
		CPUUsage:          0, // À implémenter avec des métriques système
	}, nil
}

// AddActiveConnection ajoute une connexion active
func (rs *RealtimeService) AddActiveConnection(connectionID string) error {
	return rs.redisClient.SAdd(rs.Ctx, "active_connections", connectionID).Err()
}

// RemoveActiveConnection supprime une connexion active
func (rs *RealtimeService) RemoveActiveConnection(connectionID string) error {
	return rs.redisClient.SRem(rs.Ctx, "active_connections", connectionID).Err()
}

// AddActiveDelivery ajoute une livraison active
func (rs *RealtimeService) AddActiveDelivery(deliveryID string) error {
	return rs.redisClient.SAdd(rs.Ctx, "active_deliveries", deliveryID).Err()
}

// RemoveActiveDelivery supprime une livraison active
func (rs *RealtimeService) RemoveActiveDelivery(deliveryID string) error {
	return rs.redisClient.SRem(rs.Ctx, "active_deliveries", deliveryID).Err()
}

// SetDriverLocation met en cache la position d'un livreur
func (rs *RealtimeService) SetDriverLocation(driverID string, location models.LocationUpdate) error {
	key := fmt.Sprintf("driver_location:%s", driverID)

	data, err := json.Marshal(location)
	if err != nil {
		return fmt.Errorf("erreur sérialisation location: %v", err)
	}

	return rs.redisClient.Set(rs.Ctx, key, string(data), 5*time.Minute).Err()
}

// GetDriverLocation récupère la position d'un livreur
func (rs *RealtimeService) GetDriverLocation(driverID string) (*models.LocationUpdate, error) {
	key := fmt.Sprintf("driver_location:%s", driverID)

	data, err := rs.redisClient.Get(rs.Ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var location models.LocationUpdate
	if err := json.Unmarshal([]byte(data), &location); err != nil {
		return nil, fmt.Errorf("erreur désérialisation location: %v", err)
	}

	return &location, nil
}

// BroadcastToAllClients diffuse un message à tous les clients connectés
func (rs *RealtimeService) BroadcastToAllClients(event models.SSEEvent) error {
	channel := "broadcast:all"

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("erreur sérialisation event: %v", err)
	}

	return rs.redisClient.Publish(rs.Ctx, channel, string(data)).Err()
}

// BroadcastToDrivers diffuse un message à tous les livreurs
func (rs *RealtimeService) BroadcastToDrivers(event models.SSEEvent) error {
	channel := "broadcast:drivers"

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("erreur sérialisation event: %v", err)
	}

	return rs.redisClient.Publish(rs.Ctx, channel, string(data)).Err()
}

// BroadcastToClients diffuse un message à tous les clients
func (rs *RealtimeService) BroadcastToClients(event models.SSEEvent) error {
	channel := "broadcast:clients"

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("erreur sérialisation event: %v", err)
	}

	return rs.redisClient.Publish(rs.Ctx, channel, string(data)).Err()
}

// CleanupInactiveConnections nettoie les connexions inactives
func (rs *RealtimeService) CleanupInactiveConnections() error {
	// Récupérer toutes les connexions actives
	connections, err := rs.redisClient.SMembers(rs.Ctx, "active_connections").Result()
	if err != nil {
		return err
	}

	now := time.Now()
	cutoff := now.Add(-5 * time.Minute) // Connexions inactives depuis 5 minutes

	for _, connID := range connections {
		// Vérifier si la connexion est inactive
		lastPingKey := fmt.Sprintf("connection_ping:%s", connID)
		lastPingStr, err := rs.redisClient.Get(rs.Ctx, lastPingKey).Result()
		if err != nil {
			// Connexion probablement inactive, la supprimer
			rs.RemoveActiveConnection(connID)
			continue
		}

		lastPing, err := time.Parse(time.RFC3339, lastPingStr)
		if err != nil || lastPing.Before(cutoff) {
			// Connexion inactive, la supprimer
			rs.RemoveActiveConnection(connID)
			rs.redisClient.Del(rs.Ctx, lastPingKey)
		}
	}

	return nil
}

// StartCleanupRoutine démarre la routine de nettoyage
func (rs *RealtimeService) StartCleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			if err := rs.CleanupInactiveConnections(); err != nil {
				logrus.Errorf("Erreur nettoyage connexions: %v", err)
			}
		}
	}()
}

// === SUPPORT METHODS ===

// PublishSupportTicketUpdate publie une mise à jour de ticket
func (rs *RealtimeService) PublishSupportTicketUpdate(ticketID string, update interface{}) error {
	channel := fmt.Sprintf("support:ticket:%s", ticketID)

	data, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("erreur sérialisation ticket update: %v", err)
	}

	return rs.redisClient.Publish(rs.Ctx, channel, string(data)).Err()
}

// PublishSupportMessage publie un message de support
func (rs *RealtimeService) PublishSupportMessage(conversationID string, message interface{}) error {
	channel := fmt.Sprintf("support:ticket:%s", conversationID)

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("erreur sérialisation support message: %v", err)
	}

	return rs.redisClient.Publish(rs.Ctx, channel, string(data)).Err()
}

// PublishInternalGroupMessage publie un message de groupe interne
func (rs *RealtimeService) PublishInternalGroupMessage(groupID string, message interface{}) error {
	channel := fmt.Sprintf("internal:group:%s", groupID)

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("erreur sérialisation group message: %v", err)
	}

	return rs.redisClient.Publish(rs.Ctx, channel, string(data)).Err()
}

// PublishMessage publie un message sur un canal spécifique
func (rs *RealtimeService) PublishMessage(channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("erreur sérialisation message: %v", err)
	}

	return rs.redisClient.Publish(rs.Ctx, channel, string(data)).Err()
}

// BroadcastToStaff diffuse un message à tout le staff
func (rs *RealtimeService) BroadcastToStaff(event models.SSEEvent) error {
	channel := "broadcast:staff"

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("erreur sérialisation staff event: %v", err)
	}

	return rs.redisClient.Publish(rs.Ctx, channel, string(data)).Err()
}

// SubscribeToSupportTicket s'abonne aux mises à jour d'un ticket
func (rs *RealtimeService) SubscribeToSupportTicket(ticketID string) *redis.PubSub {
	channel := fmt.Sprintf("support:ticket:%s", ticketID)
	return rs.redisClient.Subscribe(rs.Ctx, channel)
}

// SubscribeToInternalGroup s'abonne aux messages d'un groupe interne
func (rs *RealtimeService) SubscribeToInternalGroup(groupID string) *redis.PubSub {
	channel := fmt.Sprintf("internal:group:%s", groupID)
	return rs.redisClient.Subscribe(rs.Ctx, channel)
}

// SubscribeToStaffBroadcast s'abonne au canal broadcast staff
func (rs *RealtimeService) SubscribeToStaffBroadcast() *redis.PubSub {
	return rs.redisClient.Subscribe(rs.Ctx, "broadcast:staff")
}

// NotifyDirectContact notifie un contact direct admin->user
func (rs *RealtimeService) NotifyDirectContact(adminID, userID string, message interface{}) error {
	channel := fmt.Sprintf("direct:admin:%s:user:%s", adminID, userID)

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("erreur sérialisation direct contact: %v", err)
	}

	return rs.redisClient.Publish(rs.Ctx, channel, string(data)).Err()
}
