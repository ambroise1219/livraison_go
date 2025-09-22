package support

import "github.com/ambroise1219/livraison_go/models"

// SupportService interface définit les méthodes du service de support
type SupportService interface {
	// === GESTION DES TICKETS ===
	
	// CreateTicket crée un nouveau ticket de support
	CreateTicket(userID string, role models.UserRole, req *models.CreateTicketRequest) (*models.SupportTicket, error)
	
	// GetTickets récupère les tickets selon les filtres et permissions
	GetTickets(filters map[string]string, userID string, role models.UserRole, page, limit int) ([]*models.SupportTicket, int, error)
	
	// GetTicketByID récupère un ticket par ID avec vérification des permissions
	GetTicketByID(ticketID, userID string, role models.UserRole) (*models.SupportTicket, error)
	
	// UpdateTicketStatus met à jour le statut d'un ticket
	UpdateTicketStatus(ticketID string, status models.TicketStatus, updatedBy string) error
	
	// CloseTicket ferme un ticket
	CloseTicket(ticketID, closedBy string) error
	
	// ReassignTicket réassigne un ticket à un autre utilisateur
	ReassignTicket(ticketID, fromUserID, toUserID, reason string, notes *string, reassignedBy string) (*models.ReassignmentLog, error)
	
	// InitiateDirectContact initie un contact direct entre admin et utilisateur
	InitiateDirectContact(req *models.DirectContactRequest, adminID string) (*models.SupportTicket, error)
	
	// === GESTION DES MESSAGES ===
	
	// AddMessage ajoute un message à une conversation
	AddMessage(conversationID, senderID string, senderRole models.UserRole, req *models.AddMessageRequest) (*models.SupportMessage, error)
	
	// GetMessages récupère les messages d'une conversation
	GetMessages(conversationID, userID string, role models.UserRole, page, limit int) ([]*models.SupportMessage, error)
	
	// === GESTION DES GROUPES INTERNES ===
	
	// CreateInternalGroup crée un groupe de chat interne pour le staff
	CreateInternalGroup(req *models.CreateGroupRequest, createdBy string) (*models.InternalGroupChat, error)
	
	// GetUserGroups récupère les groupes d'un utilisateur
	GetUserGroups(userID string) ([]*models.InternalGroupChat, error)
	
	// AddGroupMessage ajoute un message à un groupe
	AddGroupMessage(groupID, senderID, content string) (*models.SupportMessage, error)
	
	// GetGroupMessages récupère les messages d'un groupe
	GetGroupMessages(groupID, userID string, page, limit int) ([]*models.SupportMessage, error)
	
	// AddGroupParticipant ajoute un participant à un groupe
	AddGroupParticipant(groupID, participantID, addedBy string) error
	
	// RemoveGroupParticipant retire un participant d'un groupe
	RemoveGroupParticipant(groupID, participantID, removedBy string) error
	
	// === HISTORIQUE ET STATISTIQUES ===
	
	// GetReassignmentHistory récupère l'historique des réassignations d'un ticket
	GetReassignmentHistory(ticketID string) ([]*models.ReassignmentLog, error)
	
	// GetSupportStats récupère les statistiques de support pour un utilisateur
	GetSupportStats(userID string, role models.UserRole) (*models.SupportStats, error)
}