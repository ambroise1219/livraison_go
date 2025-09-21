package models

import (
	"time"
)

// RealtimeConnection représente une connexion temps réel
type RealtimeConnection struct {
	ID             string    `json:"id" db:"id"`
	UserID         string    `json:"user_id" db:"user_id"`
	ConnectionType string    `json:"connection_type" db:"connection_type"` // 'client', 'driver', 'admin'
	DeliveryID     *string   `json:"delivery_id,omitempty" db:"delivery_id"`
	IPAddress      string    `json:"ip_address" db:"ip_address"`
	UserAgent      string    `json:"user_agent" db:"user_agent"`
	ConnectedAt    time.Time `json:"connected_at" db:"connected_at"`
	LastPing       time.Time `json:"last_ping" db:"last_ping"`
	IsActive       bool      `json:"is_active" db:"is_active"`
}

// DeliveryTracking représente le tracking d'une livraison
type DeliveryTracking struct {
	ID         string    `json:"id" db:"id"`
	DeliveryID string    `json:"delivery_id" db:"delivery_id"`
	DriverID   string    `json:"driver_id" db:"driver_id"`
	Latitude   float64   `json:"latitude" db:"latitude"`
	Longitude  float64   `json:"longitude" db:"longitude"`
	Accuracy   float64   `json:"accuracy" db:"accuracy"`
	Speed      float64   `json:"speed" db:"speed"`
	Heading    float64   `json:"heading" db:"heading"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// ChatMessage représente un message de chat
type ChatMessage struct {
	ID          string                 `json:"id" db:"id"`
	ChatID      string                 `json:"chat_id"`
	DeliveryID  string                 `json:"delivery_id" db:"delivery_id"`
	SenderID    string                 `json:"sender_id" db:"sender_id"`
	ReceiverID  string                 `json:"receiver_id" db:"receiver_id"`
	MessageType string                 `json:"message_type" db:"message_type"` // 'text', 'image', 'location'
	Content     string                 `json:"content" db:"content"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	IsRead      bool                   `json:"is_read" db:"is_read"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}


// SSEEvent représente un événement Server-Sent Events
type SSEEvent struct {
	ID        string      `json:"id"`
	Event     string      `json:"event"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Retry     int         `json:"retry,omitempty"`
}

// DeliveryUpdate représente une mise à jour de livraison
type DeliveryUpdate struct {
	DeliveryID    string    `json:"delivery_id"`
	Status        string    `json:"status"`
	DriverID      string    `json:"driver_id,omitempty"`
	DriverName    string    `json:"driver_name,omitempty"`
	DriverPhone   string    `json:"driver_phone,omitempty"`
	Latitude      float64   `json:"latitude,omitempty"`
	Longitude     float64   `json:"longitude,omitempty"`
	EstimatedTime int       `json:"estimated_time,omitempty"` // en minutes
	Distance      float64   `json:"distance,omitempty"`       // en km
	Message       string    `json:"message,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// LocationUpdate représente une mise à jour de position
type LocationUpdate struct {
	DriverID   string    `json:"driver_id"`
	DeliveryID string    `json:"delivery_id"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Accuracy   float64   `json:"accuracy"`
	Speed      float64   `json:"speed"`
	Heading    float64   `json:"heading"`
	Timestamp  time.Time `json:"timestamp"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ChatUpdate représente une mise à jour de chat
type ChatUpdate struct {
	DeliveryID string      `json:"delivery_id"`
	Message    ChatMessage `json:"message"`
	Timestamp  time.Time   `json:"timestamp"`
}

// NotificationUpdate représente une mise à jour de notification
type NotificationUpdate struct {
	UserID       string       `json:"user_id"`
	Notification Notification `json:"notification"`
	Timestamp    time.Time    `json:"timestamp"`
}

// RealtimeStats représente les statistiques temps réel
type RealtimeStats struct {
	ActiveConnections int `json:"active_connections"`
	ActiveDeliveries  int `json:"active_deliveries"`
	MessagesPerSecond int `json:"messages_per_second"`
	AverageLatency    int `json:"average_latency"` // en ms
	MemoryUsage       int `json:"memory_usage"`    // en MB
	CPUUsage          int `json:"cpu_usage"`       // en %
}
