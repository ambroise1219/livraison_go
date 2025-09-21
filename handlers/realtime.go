package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ambroise1219/livraison_go/models"
	"github.com/ambroise1219/livraison_go/services"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // En production, vérifier l'origine
	},
}

// RealtimeHandler gère les handlers temps réel
type RealtimeHandler struct {
	realtimeService *services.RealtimeService
}

// NewRealtimeHandler crée un nouveau handler temps réel
func NewRealtimeHandler() *RealtimeHandler {
	return &RealtimeHandler{
		realtimeService: services.NewRealtimeService(),
	}
}

// SSEHandler gère les connexions Server-Sent Events
func (rh *RealtimeHandler) SSEHandler(c *gin.Context) {
	// Récupérer les paramètres
	deliveryID := c.Param("deliveryId")
	userType := c.Query("type") // "client" ou "driver"
	userID := c.Query("userId")

	if deliveryID == "" || userType == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paramètres manquants"})
		return
	}

	// Configurer les headers SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// Créer un canal pour la connexion
	connectionID := fmt.Sprintf("%s_%s_%d", userType, userID, time.Now().Unix())

	// Ajouter la connexion active
	if err := rh.realtimeService.AddActiveConnection(connectionID); err != nil {
		logrus.Errorf("Erreur ajout connexion active: %v", err)
	}

	// Nettoyer à la fermeture
	defer func() {
		rh.realtimeService.RemoveActiveConnection(connectionID)
	}()

	// Envoyer un message de connexion
	connectEvent := models.SSEEvent{
		Type:      "connection",
		Data:      gin.H{"message": "Connexion établie", "connectionId": connectionID},
		Timestamp: time.Now(),
	}
	rh.sendSSEMessage(c, connectEvent)

	// S'abonner aux différents canaux selon le type d'utilisateur
	var pubsubs []*redis.PubSub

	if userType == "client" {
		// Client s'abonne aux mises à jour de livraison et chat
		pubsubs = append(pubsubs, rh.realtimeService.SubscribeToDeliveryUpdates(deliveryID))
		pubsubs = append(pubsubs, rh.realtimeService.SubscribeToChatMessages(deliveryID))
		pubsubs = append(pubsubs, rh.realtimeService.SubscribeToNotifications(userID))
	} else if userType == "driver" {
		// Driver s'abonne aux mises à jour de position et chat
		pubsubs = append(pubsubs, rh.realtimeService.SubscribeToLocationUpdates(deliveryID))
		pubsubs = append(pubsubs, rh.realtimeService.SubscribeToChatMessages(deliveryID))
		pubsubs = append(pubsubs, rh.realtimeService.SubscribeToNotifications(userID))
	}

	// Nettoyer les abonnements
	defer func() {
		for _, pubsub := range pubsubs {
			pubsub.Close()
		}
	}()

	// Créer un canal pour les messages
	messageChan := make(chan models.SSEEvent, 100)

	// Goroutine pour écouter les messages Redis
	go func() {
		for _, pubsub := range pubsubs {
			go func(ps *redis.PubSub) {
				for {
					msg, err := ps.ReceiveMessage(rh.realtimeService.Ctx)
					if err != nil {
						logrus.Errorf("Erreur réception message Redis: %v", err)
						return
					}

					// Convertir le message Redis en SSEEvent
					var event models.SSEEvent
					if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
						logrus.Errorf("Erreur désérialisation message: %v", err)
						continue
					}

					select {
					case messageChan <- event:
					default:
						logrus.Warn("Canal de message plein, message ignoré")
					}
				}
			}(pubsub)
		}
	}()

	// Envoyer les messages via SSE
	ticker := time.NewTicker(30 * time.Second) // Ping toutes les 30s
	defer ticker.Stop()

	for {
		select {
		case event := <-messageChan:
			rh.sendSSEMessage(c, event)
		case <-ticker.C:
			// Envoyer un ping pour maintenir la connexion
			pingEvent := models.SSEEvent{
				Type:      "ping",
				Data:      gin.H{"timestamp": time.Now()},
				Timestamp: time.Now(),
			}
			rh.sendSSEMessage(c, pingEvent)
		case <-c.Request.Context().Done():
			// Connexion fermée
			return
		}
	}
}

// sendSSEMessage envoie un message SSE formaté
func (rh *RealtimeHandler) sendSSEMessage(c *gin.Context, event models.SSEEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		logrus.Errorf("Erreur sérialisation event SSE: %v", err)
		return
	}

	fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	c.Writer.Flush()
}

// WebSocketHandler gère les connexions WebSocket pour le chat
func (rh *RealtimeHandler) WebSocketHandler(c *gin.Context) {
	deliveryID := c.Param("deliveryId")
	userType := c.Query("type")
	userID := c.Query("userId")

	if deliveryID == "" || userType == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paramètres manquants"})
		return
	}

	// Upgrader la connexion HTTP en WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("Erreur upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// Ajouter la connexion active
	connectionID := fmt.Sprintf("ws_%s_%s_%d", userType, userID, time.Now().Unix())
	rh.realtimeService.AddActiveConnection(connectionID)
	defer rh.realtimeService.RemoveActiveConnection(connectionID)

	// S'abonner aux messages de chat
	chatPubsub := rh.realtimeService.SubscribeToChatMessages(deliveryID)
	defer chatPubsub.Close()

	// Goroutine pour écouter les messages Redis
	go func() {
		for {
			msg, err := chatPubsub.ReceiveMessage(rh.realtimeService.Ctx)
			if err != nil {
				logrus.Errorf("Erreur réception message chat: %v", err)
				return
			}

			// Envoyer le message via WebSocket
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				logrus.Errorf("Erreur envoi message WebSocket: %v", err)
				return
			}
		}
	}()

	// Écouter les messages entrants du WebSocket
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			logrus.Errorf("Erreur lecture message WebSocket: %v", err)
			break
		}

		// Parser le message de chat
		var chatMessage models.ChatMessage
		if err := json.Unmarshal(message, &chatMessage); err != nil {
			logrus.Errorf("Erreur parsing message chat: %v", err)
			continue
		}

		// Valider et enrichir le message
		chatMessage.ChatID = deliveryID
		chatMessage.SenderID = userID
		chatMessage.CreatedAt = time.Now()

		// Publier le message via Redis
		if err := rh.realtimeService.PublishChatMessage(deliveryID, chatMessage); err != nil {
			logrus.Errorf("Erreur publication message chat: %v", err)
		}
	}
}

// UpdateLocationHandler met à jour la position d'un livreur
func (rh *RealtimeHandler) UpdateLocationHandler(c *gin.Context) {
	driverID := c.Param("driverId")
	deliveryID := c.Param("deliveryId")

	var locationUpdate models.LocationUpdate
	if err := c.ShouldBindJSON(&locationUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Valider les coordonnées
	if locationUpdate.Latitude < -90 || locationUpdate.Latitude > 90 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Latitude invalide"})
		return
	}
	if locationUpdate.Longitude < -180 || locationUpdate.Longitude > 180 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Longitude invalide"})
		return
	}

	// Enrichir les données
	locationUpdate.DriverID = driverID
	locationUpdate.DeliveryID = deliveryID
	locationUpdate.UpdatedAt = time.Now()

	// Mettre en cache la position
	if err := rh.realtimeService.SetDriverLocation(driverID, locationUpdate); err != nil {
		logrus.Errorf("Erreur mise en cache position: %v", err)
	}

	// Publier la mise à jour
	if err := rh.realtimeService.PublishLocationUpdate(deliveryID, locationUpdate); err != nil {
		logrus.Errorf("Erreur publication position: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur publication position"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Position mise à jour", "location": locationUpdate})
}

// GetDriverLocationHandler récupère la position d'un livreur
func (rh *RealtimeHandler) GetDriverLocationHandler(c *gin.Context) {
	driverID := c.Param("driverId")

	location, err := rh.realtimeService.GetDriverLocation(driverID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Position non trouvée"})
		return
	}

	c.JSON(http.StatusOK, location)
}

// UpdateDeliveryStatusHandler met à jour le statut d'une livraison
func (rh *RealtimeHandler) UpdateDeliveryStatusHandler(c *gin.Context) {
	deliveryID := c.Param("deliveryId")

	var statusUpdate models.DeliveryUpdate
	if err := c.ShouldBindJSON(&statusUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enrichir les données
	statusUpdate.DeliveryID = deliveryID
	statusUpdate.UpdatedAt = time.Now()

	// Publier la mise à jour
	if err := rh.realtimeService.PublishDeliveryUpdate(deliveryID, statusUpdate); err != nil {
		logrus.Errorf("Erreur publication statut: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur publication statut"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Statut mis à jour", "update": statusUpdate})
}

// SendNotificationHandler envoie une notification
func (rh *RealtimeHandler) SendNotificationHandler(c *gin.Context) {
	userID := c.Param("userId")

	var notification models.Notification
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enrichir les données
	notification.UserID = userID
	notification.CreatedAt = time.Now()

	// Publier la notification
	if err := rh.realtimeService.PublishNotification(userID, notification); err != nil {
		logrus.Errorf("Erreur publication notification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur publication notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification envoyée", "notification": notification})
}

// GetRealtimeStatsHandler récupère les statistiques temps réel
func (rh *RealtimeHandler) GetRealtimeStatsHandler(c *gin.Context) {
	stats, err := rh.realtimeService.GetRealtimeStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CalculateETAHandler calcule le temps estimé d'arrivée
func (rh *RealtimeHandler) CalculateETAHandler(c *gin.Context) {
	deliveryID := c.Param("deliveryId")
	clientLatStr := c.Query("client_lat")
	clientLngStr := c.Query("client_lng")

	if clientLatStr == "" || clientLngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Coordonnées client requises"})
		return
	}

	clientLat, err := strconv.ParseFloat(clientLatStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Latitude client invalide"})
		return
	}

	clientLng, err := strconv.ParseFloat(clientLngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Longitude client invalide"})
		return
	}

	// Récupérer la position du livreur (simulée pour l'exemple)
	// En réalité, on récupérerait la position depuis la base de données
	driverLat := 48.8566 // Paris
	driverLng := 2.3522

	// Calculer la distance
	distance := rh.realtimeService.CalculateDistance(driverLat, driverLng, clientLat, clientLng)

	// Calculer l'ETA
	eta := rh.realtimeService.CalculateEstimatedTime(distance, 30) // 30 km/h

	c.JSON(http.StatusOK, gin.H{
		"delivery_id": deliveryID,
		"distance_km": fmt.Sprintf("%.2f", distance),
		"eta_minutes": eta,
		"driver_position": gin.H{
			"lat": driverLat,
			"lng": driverLng,
		},
		"client_position": gin.H{
			"lat": clientLat,
			"lng": clientLng,
		},
	})
}
