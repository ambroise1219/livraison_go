package models

import (
	"time"
)

// TicketCategory représente les catégories de tickets
type TicketCategory string

const (
	TicketCategoryTechnical  TicketCategory = "TECHNICAL"
	TicketCategoryBilling    TicketCategory = "BILLING"
	TicketCategoryDelivery   TicketCategory = "DELIVERY"
	TicketCategoryAccount    TicketCategory = "ACCOUNT"
	TicketCategoryComplaint  TicketCategory = "COMPLAINT"
	TicketCategoryMarketing  TicketCategory = "MARKETING"
	TicketCategoryOther      TicketCategory = "OTHER"
)

// TicketPriority représente la priorité des tickets
type TicketPriority string

const (
	TicketPriorityLow    TicketPriority = "LOW"
	TicketPriorityMedium TicketPriority = "MEDIUM"
	TicketPriorityHigh   TicketPriority = "HIGH"
	TicketPriorityUrgent TicketPriority = "URGENT"
)

// TicketStatus représente le statut des tickets
type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "OPEN"
	TicketStatusAssigned   TicketStatus = "ASSIGNED"
	TicketStatusInProgress TicketStatus = "IN_PROGRESS"
	TicketStatusPending    TicketStatus = "PENDING"
	TicketStatusResolved   TicketStatus = "RESOLVED"
	TicketStatusClosed     TicketStatus = "CLOSED"
)

// ConversationType représente le type de conversation
type ConversationType string

const (
	ConversationTypeTicket        ConversationType = "TICKET"
	ConversationTypeInternalGroup ConversationType = "INTERNAL_GROUP"
	ConversationTypeDirect        ConversationType = "DIRECT"
)

// SupportTicket représente un ticket de support
type SupportTicket struct {
	ID                   string         `json:"id" db:"id"`
	Title                string         `json:"title" db:"title"`
	Description          string         `json:"description" db:"description"`
	Category             TicketCategory `json:"category" db:"category"`
	Priority             TicketPriority `json:"priority" db:"priority"`
	Status               TicketStatus   `json:"status" db:"status"`
	CreatedBy            string         `json:"created_by" db:"created_by"`
	CreatedRole          UserRole       `json:"created_role" db:"created_role"`
	AssignedTo           *string        `json:"assigned_to" db:"assigned_to"`
	PreviouslyAssignedTo []string       `json:"previously_assigned_to" db:"previously_assigned_to"`
	TeamChannel          *string        `json:"team_channel" db:"team_channel"`
	RelatedDeliveryID    *string        `json:"related_delivery_id" db:"related_delivery_id"`
	IsInternal           bool           `json:"is_internal" db:"is_internal"`
	Participants         []string       `json:"participants" db:"participants"`
	CreatedAt            time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at" db:"updated_at"`
	ReassignedAt         *time.Time     `json:"reassigned_at" db:"reassigned_at"`
	ClosedAt             *time.Time     `json:"closed_at" db:"closed_at"`
}

// InternalGroupChat représente un chat de groupe interne
type InternalGroupChat struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	Participants []string  `json:"participants" db:"participants"`
	CreatedBy    string    `json:"created_by" db:"created_by"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// SupportMessage représente un message dans le système de support
type SupportMessage struct {
	ID               string           `json:"id" db:"id"`
	ConversationID   string           `json:"conversation_id" db:"conversation_id"`
	ConversationType ConversationType `json:"conversation_type" db:"conversation_type"`
	SenderID         string           `json:"sender_id" db:"sender_id"`
	SenderRole       UserRole         `json:"sender_role" db:"sender_role"`
	Content          string           `json:"content" db:"content"`
	MessageType      string           `json:"message_type" db:"message_type"`
	IsInternal       bool             `json:"is_internal" db:"is_internal"`
	ReplyTo          *string          `json:"reply_to" db:"reply_to"`
	CreatedAt        time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at" db:"updated_at"`
}

// ReassignmentLog représente l'historique des réassignations
type ReassignmentLog struct {
	ID        string    `json:"id" db:"id"`
	TicketID  string    `json:"ticket_id" db:"ticket_id"`
	FromUser  *string   `json:"from_user" db:"from_user"`
	ToUser    string    `json:"to_user" db:"to_user"`
	Reason    string    `json:"reason" db:"reason"`
	Notes     *string   `json:"notes" db:"notes"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Request structs

// CreateTicketRequest représente une demande de création de ticket
type CreateTicketRequest struct {
	Title               string         `json:"title" validate:"required,max=200"`
	Description         string         `json:"description" validate:"required,max=2000"`
	Category            TicketCategory `json:"category" validate:"required"`
	Priority            TicketPriority `json:"priority" validate:"required"`
	RelatedDeliveryID   *string        `json:"related_delivery_id,omitempty"`
	PreferredAssigneeID *string        `json:"preferred_assignee_id,omitempty"`
}

// UpdateTicketRequest représente une demande de mise à jour de ticket
type UpdateTicketRequest struct {
	Status     *TicketStatus   `json:"status,omitempty"`
	Priority   *TicketPriority `json:"priority,omitempty"`
	AssignedTo *string         `json:"assigned_to,omitempty"`
}

// ReassignTicketRequest représente une demande de réassignation
type ReassignTicketRequest struct {
	ToUserID string  `json:"to_user_id" validate:"required"`
	Reason   string  `json:"reason" validate:"required,max=500"`
	Notes    *string `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

// CreateGroupRequest représente une demande de création de groupe
type CreateGroupRequest struct {
	Name         string   `json:"name" validate:"required,max=100"`
	Description  string   `json:"description" validate:"required,max=500"`
	Participants []string `json:"participants" validate:"required,min=1"`
}

// AddMessageRequest représente une demande d'ajout de message
type AddMessageRequest struct {
	Content     string  `json:"content" validate:"required,max=2000"`
	MessageType string  `json:"message_type" validate:"required"`
	IsInternal  bool    `json:"is_internal"`
	ReplyTo     *string `json:"reply_to,omitempty"`
}

// DirectContactRequest représente une demande de contact direct
type DirectContactRequest struct {
	TargetUserID     string `json:"target_user_id" validate:"required"`
	InitialMessage   string `json:"initial_message" validate:"required,max=2000"`
	Subject          string `json:"subject" validate:"required,max=200"`
	CreateTicket     bool   `json:"create_ticket"`
}

// SupportStats représente les statistiques du support
type SupportStats struct {
	TotalTickets        int                        `json:"total_tickets"`
	OpenTickets         int                        `json:"open_tickets"`
	MyAssignedTickets   int                        `json:"my_assigned_tickets"`
	TicketsByStatus     map[string]int             `json:"tickets_by_status"`
	TicketsByCategory   map[string]int             `json:"tickets_by_category"`
	TicketsByPriority   map[string]int             `json:"tickets_by_priority"`
	AverageResponseTime float64                    `json:"average_response_time"` // en minutes
	ReassignmentRate    float64                    `json:"reassignment_rate"`     // pourcentage
	ActiveGroups        int                        `json:"active_groups"`
	RecentActivity      []*RecentSupportActivity   `json:"recent_activity"`
}

// RecentSupportActivity représente une activité récente
type RecentSupportActivity struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // "ticket_created", "message_sent", "reassigned", etc.
	Description string    `json:"description"`
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	Timestamp   time.Time `json:"timestamp"`
}

// Validation methods

// IsValidCategory vérifie si la catégorie est valide
func (c TicketCategory) IsValid() bool {
	return c == TicketCategoryTechnical || c == TicketCategoryBilling ||
		   c == TicketCategoryDelivery || c == TicketCategoryAccount ||
		   c == TicketCategoryComplaint || c == TicketCategoryMarketing ||
		   c == TicketCategoryOther
}

// IsValidPriority vérifie si la priorité est valide
func (p TicketPriority) IsValid() bool {
	return p == TicketPriorityLow || p == TicketPriorityMedium ||
		   p == TicketPriorityHigh || p == TicketPriorityUrgent
}

// IsValidStatus vérifie si le statut est valide
func (s TicketStatus) IsValid() bool {
	return s == TicketStatusOpen || s == TicketStatusAssigned ||
		   s == TicketStatusInProgress || s == TicketStatusPending ||
		   s == TicketStatusResolved || s == TicketStatusClosed
}

// Business logic methods

// CanBeReassigned vérifie si un ticket peut être réassigné
func (t *SupportTicket) CanBeReassigned() bool {
	return t.Status != TicketStatusClosed && t.Status != TicketStatusResolved
}

// CanAddMessage vérifie si on peut ajouter un message
func (t *SupportTicket) CanAddMessage() bool {
	return t.Status != TicketStatusClosed
}

// GetDefaultAssigneeByCategory retourne l'assigné par défaut selon la catégorie
func (c TicketCategory) GetDefaultAssigneeRole() UserRole {
	switch c {
	case TicketCategoryTechnical:
		return UserRoleAdmin
	case TicketCategoryBilling:
		return UserRoleGestionnaire
	case TicketCategoryMarketing:
		return UserRoleMarketing
	default:
		return UserRoleAdmin
	}
}

// ToResponse methods

// ToResponse convertit un SupportTicket en réponse API
func (t *SupportTicket) ToResponse() map[string]interface{} {
	response := map[string]interface{}{
		"id":             t.ID,
		"title":          t.Title,
		"description":    t.Description,
		"category":       t.Category,
		"priority":       t.Priority,
		"status":         t.Status,
		"created_by":     t.CreatedBy,
		"created_role":   t.CreatedRole,
		"assigned_to":    t.AssignedTo,
		"is_internal":    t.IsInternal,
		"participants":   t.Participants,
		"created_at":     t.CreatedAt,
		"updated_at":     t.UpdatedAt,
	}

	if t.RelatedDeliveryID != nil {
		response["related_delivery_id"] = *t.RelatedDeliveryID
	}
	if t.ReassignedAt != nil {
		response["reassigned_at"] = *t.ReassignedAt
	}
	if t.ClosedAt != nil {
		response["closed_at"] = *t.ClosedAt
	}
	if len(t.PreviouslyAssignedTo) > 0 {
		response["previously_assigned_to"] = t.PreviouslyAssignedTo
	}

	return response
}

// ToResponse convertit un SupportMessage en réponse API
func (m *SupportMessage) ToResponse() map[string]interface{} {
	response := map[string]interface{}{
		"id":                m.ID,
		"conversation_id":   m.ConversationID,
		"conversation_type": m.ConversationType,
		"sender_id":         m.SenderID,
		"sender_role":       m.SenderRole,
		"content":           m.Content,
		"message_type":      m.MessageType,
		"is_internal":       m.IsInternal,
		"created_at":        m.CreatedAt,
		"updated_at":        m.UpdatedAt,
	}

	if m.ReplyTo != nil {
		response["reply_to"] = *m.ReplyTo
	}

	return response
}

// ToResponse convertit un InternalGroupChat en réponse API
func (g *InternalGroupChat) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":           g.ID,
		"name":         g.Name,
		"description":  g.Description,
		"participants": g.Participants,
		"created_by":   g.CreatedBy,
		"is_active":    g.IsActive,
		"created_at":   g.CreatedAt,
		"updated_at":   g.UpdatedAt,
	}
}

// ToResponse convertit un ReassignmentLog en réponse API
func (r *ReassignmentLog) ToResponse() map[string]interface{} {
	response := map[string]interface{}{
		"id":         r.ID,
		"ticket_id":  r.TicketID,
		"to_user":    r.ToUser,
		"reason":     r.Reason,
		"created_by": r.CreatedBy,
		"created_at": r.CreatedAt,
	}

	if r.FromUser != nil {
		response["from_user"] = *r.FromUser
	}
	if r.Notes != nil {
		response["notes"] = *r.Notes
	}

	return response
}