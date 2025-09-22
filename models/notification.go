package models

import "time"

// NotificationType représente les différents types de notifications
type NotificationType string

const (
	NotificationTypeDeliveryUpdate NotificationType = "DELIVERY_UPDATE"
	NotificationTypePayment        NotificationType = "PAYMENT"
	NotificationTypePromotion      NotificationType = "PROMOTION"
	NotificationTypeSystem         NotificationType = "SYSTEM"
)

// Notification représente une notification utilisateur
type Notification struct {
	ID        string           `json:"id"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	IsRead    bool             `json:"isRead"`
	UserID    string           `json:"userId"`
	CreatedAt time.Time        `json:"createdAt"`
}

// CreateNotificationRequest représente une demande de création de notification
type CreateNotificationRequest struct {
	UserID  string           `json:"userId" validate:"required"`
	Type    NotificationType `json:"type" validate:"required"`
	Title   string           `json:"title" validate:"required,max=100"`
	Message string           `json:"message" validate:"required,max=500"`
}

// IncidentType représente les différents types d'incidents
type IncidentType string

const (
	IncidentTypeDamage           IncidentType = "DAMAGE"
	IncidentTypeDelay            IncidentType = "DELAY"
	IncidentTypeLostPackage      IncidentType = "LOST_PACKAGE"
	IncidentTypeCustomerComplaint IncidentType = "CUSTOMER_COMPLAINT"
	IncidentTypeDriverIssue      IncidentType = "DRIVER_ISSUE"
)

// IncidentStatus représente les statuts d'un incident
type IncidentStatus string

const (
	IncidentStatusOpen       IncidentStatus = "OPEN"
	IncidentStatusInProgress IncidentStatus = "IN_PROGRESS"
	IncidentStatusResolved   IncidentStatus = "RESOLVED"
	IncidentStatusClosed     IncidentStatus = "CLOSED"
)

// Incident représente un incident de livraison
type Incident struct {
	ID          string         `json:"id"`
	Type        IncidentType   `json:"type"`
	Description string         `json:"description"`
	Status      IncidentStatus `json:"status"`
	ReportedAt  time.Time      `json:"reportedAt"`
	ResolvedAt  *time.Time     `json:"resolvedAt,omitempty"`
	DeliveryID  string         `json:"deliveryId"`
}

// CreateIncidentRequest représente une demande de création d'incident
type CreateIncidentRequest struct {
	DeliveryID  string       `json:"deliveryId" validate:"required"`
	Type        IncidentType `json:"type" validate:"required"`
	Description string       `json:"description" validate:"required,max=1000"`
}

// UpdateIncidentRequest représente une demande de mise à jour d'incident
type UpdateIncidentRequest struct {
	Status      *IncidentStatus `json:"status,omitempty"`
	Description *string         `json:"description,omitempty"`
}

// Rating représente une évaluation de livraison
type Rating struct {
	ID         string     `json:"id"`
	Rating     int        `json:"rating"` // 1-5
	Comment    *string    `json:"comment,omitempty"`
	UserID     string     `json:"userId"`
	DeliveryID string     `json:"deliveryId"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// CreateRatingRequest représente une demande de création d'évaluation
type CreateRatingRequest struct {
	DeliveryID string  `json:"deliveryId" validate:"required"`
	Rating     int     `json:"rating" validate:"required,min=1,max=5"`
	Comment    *string `json:"comment,omitempty" validate:"omitempty,max=500"`
}

// Package redeclaré supprimé - utiliser celui de delivery.go

// CreatePackageRequest et UpdatePackageRequest déplacés vers delivery.go

// Tracking représente le suivi d'une livraison
type Tracking struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	Location   *string   `json:"location,omitempty"`
	Notes      *string   `json:"notes,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	DeliveryID string    `json:"deliveryId"`
}

// CreateTrackingRequest représente une demande de création de suivi
type CreateTrackingRequest struct {
	DeliveryID string  `json:"deliveryId" validate:"required"`
	Status     string  `json:"status" validate:"required"`
	Location   *string `json:"location,omitempty"`
	Notes      *string `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// File représente un fichier uploadé
type File struct {
	ID           string    `json:"id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"originalName"`
	MimeType     string    `json:"mimeType"`
	Size         int       `json:"size"`
	Path         string    `json:"path"`
	UserID       string    `json:"userId"`
	CreatedAt    time.Time `json:"createdAt"`
}

// CreateFileRequest représente une demande de création de fichier
type CreateFileRequest struct {
	Filename     string `json:"filename" validate:"required"`
	OriginalName string `json:"originalName" validate:"required"`
	MimeType     string `json:"mimeType" validate:"required"`
	Size         int    `json:"size" validate:"required,gt=0"`
	Path         string `json:"path" validate:"required"`
	UserID       string `json:"userId" validate:"required"`
}

// UserAddress représente une adresse favorite d'un utilisateur
type UserAddress struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`      // "Domicile", "Bureau", etc.
	Address   string  `json:"address"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	IsDefault bool    `json:"isDefault"`
	UserID    string  `json:"userId"`
}

// CreateUserAddressRequest représente une demande de création d'adresse
type CreateUserAddressRequest struct {
	Name      string  `json:"name" validate:"required,max=50"`
	Address   string  `json:"address" validate:"required,max=200"`
	Lat       float64 `json:"lat" validate:"required,gte=-90,lte=90"`
	Lng       float64 `json:"lng" validate:"required,gte=-180,lte=180"`
	IsDefault *bool   `json:"isDefault,omitempty"`
}

// UpdateUserAddressRequest représente une demande de mise à jour d'adresse
type UpdateUserAddressRequest struct {
	Name      *string  `json:"name,omitempty" validate:"omitempty,max=50"`
	Address   *string  `json:"address,omitempty" validate:"omitempty,max=200"`
	Lat       *float64 `json:"lat,omitempty" validate:"omitempty,gte=-90,lte=90"`
	Lng       *float64 `json:"lng,omitempty" validate:"omitempty,gte=-180,lte=180"`
	IsDefault *bool    `json:"isDefault,omitempty"`
}

// Response methods pour les modèles

// ToResponse convertit une Notification en réponse API
func (n *Notification) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":        n.ID,
		"type":      n.Type,
		"title":     n.Title,
		"message":   n.Message,
		"isRead":    n.IsRead,
		"userId":    n.UserID,
		"createdAt": n.CreatedAt,
	}
}

// ToResponse convertit un Incident en réponse API
func (i *Incident) ToResponse() map[string]interface{} {
	response := map[string]interface{}{
		"id":          i.ID,
		"type":        i.Type,
		"description": i.Description,
		"status":      i.Status,
		"reportedAt":  i.ReportedAt,
		"deliveryId":  i.DeliveryID,
	}
	
	if i.ResolvedAt != nil {
		response["resolvedAt"] = *i.ResolvedAt
	}
	
	return response
}

// ToResponse convertit un Rating en réponse API
func (r *Rating) ToResponse() map[string]interface{} {
	response := map[string]interface{}{
		"id":         r.ID,
		"rating":     r.Rating,
		"userId":     r.UserID,
		"deliveryId": r.DeliveryID,
		"createdAt":  r.CreatedAt,
	}
	
	if r.Comment != nil {
		response["comment"] = *r.Comment
	}
	
	return response
}

// ToResponse pour Package déplacé vers delivery.go

// ToResponse convertit un Tracking en réponse API
func (t *Tracking) ToResponse() map[string]interface{} {
	response := map[string]interface{}{
		"id":         t.ID,
		"status":     t.Status,
		"timestamp":  t.Timestamp,
		"deliveryId": t.DeliveryID,
	}
	
	if t.Location != nil {
		response["location"] = *t.Location
	}
	if t.Notes != nil {
		response["notes"] = *t.Notes
	}
	
	return response
}

// ToResponse convertit un File en réponse API
func (f *File) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":           f.ID,
		"filename":     f.Filename,
		"originalName": f.OriginalName,
		"mimeType":     f.MimeType,
		"size":         f.Size,
		"path":         f.Path,
		"userId":       f.UserID,
		"createdAt":    f.CreatedAt,
	}
}

// ToResponse convertit une UserAddress en réponse API
func (ua *UserAddress) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":        ua.ID,
		"name":      ua.Name,
		"address":   ua.Address,
		"lat":       ua.Lat,
		"lng":       ua.Lng,
		"isDefault": ua.IsDefault,
		"userId":    ua.UserID,
	}
}