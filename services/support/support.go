package support

import (
	"context"
	"fmt"
	"time"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/models"
	"github.com/ambroise1219/livraison_go/prisma/db"
	"github.com/ambroise1219/livraison_go/services"
	"github.com/sirupsen/logrus"
)

// supportServicePrisma implémente SupportService avec Prisma
type supportServicePrisma struct {
	db             *db.PrismaClient
	realtimeService *services.RealtimeService
}

// NewSupportService crée une nouvelle instance du service de support avec Prisma
func NewSupportService() SupportService {
	// Charger la config pour s'assurer que les variables d'environnement sont disponibles
	config.LoadConfig()
	
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		logrus.Errorf("Failed to connect to database: %v", err)
	}
	
	return &supportServicePrisma{
		db:             client,
		realtimeService: services.NewRealtimeService(),
	}
}

// CreateTicket crée un nouveau ticket de support
func (s *supportServicePrisma) CreateTicket(userID string, role models.UserRole, req *models.CreateTicketRequest) (*models.SupportTicket, error) {
	ctx := context.Background()
	
	// Préparation des participants
	participants := []string{userID}
	if req.PreferredAssigneeID != nil {
		participants = append(participants, *req.PreferredAssigneeID)
	}
	
	// Création avec Prisma - respecting required parameters order
	createdTicket, err := s.db.SupportTicket.CreateOne(
		db.SupportTicket.Title.Set(req.Title),
		db.SupportTicket.Description.Set(req.Description),
		db.SupportTicket.Category.Set(db.TicketCategory(req.Category)),
		db.SupportTicket.CreatedBy.Set(userID),
		db.SupportTicket.CreatedRole.Set(db.UserRole(role)),
		// Optional parameters
		db.SupportTicket.Participants.Set(participants),
		db.SupportTicket.IsInternal.Set(false),
	).Exec(ctx)
	
	if err != nil {
		logrus.Errorf("Erreur création ticket: %v", err)
		return nil, fmt.Errorf("erreur création ticket: %w", err)
	}

	// Conversion vers notre modèle
	ticket := s.convertPrismaToTicket(createdTicket)

	logrus.Infof("Ticket créé avec succès: %s par %s (%s)", ticket.ID, userID, role)
	return ticket, nil
}

// GetTickets récupère les tickets selon les filtres et permissions
func (s *supportServicePrisma) GetTickets(filters map[string]string, userID string, role models.UserRole, page, limit int) ([]*models.SupportTicket, int, error) {
	ctx := context.Background()
	
	// Construction des conditions selon le rôle
	conditions := []db.SupportTicketWhereParam{}
	
	// Permissions selon le rôle
	if role == models.UserRoleClient || role == models.UserRoleLivreur {
		// Les clients/livreurs voient seulement leurs tickets
		conditions = append(conditions, db.SupportTicket.CreatedBy.Equals(userID))
	}
	
	// Filtres optionnels
	if status := filters["status"]; status != "" {
		conditions = append(conditions, db.SupportTicket.Status.Equals(db.TicketStatus(status)))
	}
	
	if category := filters["category"]; category != "" {
		conditions = append(conditions, db.SupportTicket.Category.Equals(db.TicketCategory(category)))
	}
	
	if priority := filters["priority"]; priority != "" {
		conditions = append(conditions, db.SupportTicket.Priority.Equals(db.TicketPriority(priority)))
	}
	
	if assignedTo := filters["assigned_to"]; assignedTo != "" {
		conditions = append(conditions, db.SupportTicket.AssignedTo.Equals(assignedTo))
	}
	
	// Compter le total avec les mêmes conditions
	totalResult, err := s.db.SupportTicket.FindMany(conditions...).Exec(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur count tickets: %w", err)
	}
	total := len(totalResult)
	
	// Récupérer les tickets avec pagination
	skip := (page - 1) * limit
	prismaTickets, err := s.db.SupportTicket.FindMany(conditions...).
		OrderBy(db.SupportTicket.CreatedAt.Order(db.SortOrderDesc)).
		Skip(skip).
		Take(limit).
		Exec(ctx)
	
	if err != nil {
		return nil, 0, fmt.Errorf("erreur récupération tickets: %w", err)
	}
	
	// Conversion
	tickets := make([]*models.SupportTicket, len(prismaTickets))
	for i, pt := range prismaTickets {
		tickets[i] = s.convertPrismaToTicket(&pt)
	}
	
	return tickets, total, nil
}

// GetTicketByID récupère un ticket par ID avec vérification des permissions
func (s *supportServicePrisma) GetTicketByID(ticketID, userID string, role models.UserRole) (*models.SupportTicket, error) {
	ctx := context.Background()
	
	prismaTicket, err := s.db.SupportTicket.FindUnique(
		db.SupportTicket.ID.Equals(ticketID),
	).Exec(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("ticket non trouvé: %w", err)
	}
	
	ticket := s.convertPrismaToTicket(prismaTicket)
	
	// Vérifier les permissions
	if !s.canAccessTicket(ticket, userID, role) {
		return nil, fmt.Errorf("accès refusé au ticket %s", ticketID)
	}
	
	return ticket, nil
}

// UpdateTicketStatus met à jour le statut d'un ticket
func (s *supportServicePrisma) UpdateTicketStatus(ticketID string, status models.TicketStatus, updatedBy string) error {
	ctx := context.Background()
	
	_, err := s.db.SupportTicket.FindUnique(
		db.SupportTicket.ID.Equals(ticketID),
	).Update(
		db.SupportTicket.Status.Set(db.TicketStatus(status)),
	).Exec(ctx)
	
	if err != nil {
		logrus.Errorf("Erreur mise à jour statut ticket %s: %v", ticketID, err)
		return fmt.Errorf("erreur mise à jour statut: %w", err)
	}
	
	logrus.Infof("Statut ticket mis à jour: %s -> %s par %s", ticketID, status, updatedBy)
	return nil
}

// CloseTicket ferme un ticket
func (s *supportServicePrisma) CloseTicket(ticketID, closedBy string) error {
	ctx := context.Background()
	now := time.Now()
	
	_, err := s.db.SupportTicket.FindUnique(
		db.SupportTicket.ID.Equals(ticketID),
	).Update(
		db.SupportTicket.Status.Set(db.TicketStatusClosed),
		db.SupportTicket.ClosedAt.Set(now),
	).Exec(ctx)
	
	if err != nil {
		logrus.Errorf("Erreur fermeture ticket %s: %v", ticketID, err)
		return fmt.Errorf("erreur fermeture ticket: %w", err)
	}
	
	logrus.Infof("Ticket fermé: %s par %s", ticketID, closedBy)
	return nil
}

// AddMessage ajoute un message à une conversation
func (s *supportServicePrisma) AddMessage(conversationID, senderID string, senderRole models.UserRole, req *models.AddMessageRequest) (*models.SupportMessage, error) {
	ctx := context.Background()
	
	// Déterminer le type de conversation selon le conversationID
	// Si c'est un ticket ID, c'est un message de ticket
	conversationType := models.ConversationTypeTicket
	var ticketID *string
	var groupID *string
	
	// Vérifier si c'est un ticket existant
	ticket, err := s.db.SupportTicket.FindUnique(
		db.SupportTicket.ID.Equals(conversationID),
	).Exec(ctx)
	
	if err == nil && ticket != nil {
		// C'est un message de ticket
		conversationType = models.ConversationTypeTicket
		ticketID = &conversationID
	} else {
		// Vérifier si c'est un groupe
		group, err := s.db.InternalGroupChat.FindUnique(
			db.InternalGroupChat.ID.Equals(conversationID),
		).Exec(ctx)
		
		if err == nil && group != nil {
			conversationType = models.ConversationTypeInternalGroup
			groupID = &conversationID
		} else {
			// Conversation directe ou autre
			conversationType = models.ConversationTypeDirect
		}
	}
	
	createdMessage, err := s.db.SupportMessage.CreateOne(
		// Required parameters in correct order
		db.SupportMessage.ConversationID.Set(conversationID),
		db.SupportMessage.SenderID.Set(senderID),
		db.SupportMessage.SenderRole.Set(db.UserRole(senderRole)),
		db.SupportMessage.Content.Set(req.Content),
		// Optional parameters
		db.SupportMessage.ConversationType.Set(db.ConversationType(conversationType)),
		db.SupportMessage.MessageType.Set(req.MessageType),
		db.SupportMessage.IsInternal.Set(req.IsInternal),
		db.SupportMessage.TicketID.SetOptional(ticketID),
		db.SupportMessage.GroupID.SetOptional(groupID),
	).Exec(ctx)
	
	if err != nil {
		logrus.Errorf("Erreur création message: %v", err)
		return nil, fmt.Errorf("erreur création message: %w", err)
	}
	
	message := s.convertPrismaToMessage(createdMessage)
	
	logrus.Infof("Message ajouté: %s dans %s (%s)", message.ID, conversationID, conversationType)
	return message, nil
}

// GetMessages récupère les messages d'une conversation
func (s *supportServicePrisma) GetMessages(conversationID, userID string, role models.UserRole, page, limit int) ([]*models.SupportMessage, error) {
	ctx := context.Background()
	
	// Conditions de base
	conditions := []db.SupportMessageWhereParam{
		db.SupportMessage.ConversationID.Equals(conversationID),
	}
	
	// Filtrer les messages internes selon le rôle
	if role == models.UserRoleClient || role == models.UserRoleLivreur {
		// Les clients/livreurs ne voient pas les messages internes
		conditions = append(conditions, db.SupportMessage.IsInternal.Equals(false))
	}
	
	skip := (page - 1) * limit
	prismaMessages, err := s.db.SupportMessage.FindMany(conditions...).
		OrderBy(db.SupportMessage.CreatedAt.Order(db.SortOrderAsc)).
		Skip(skip).
		Take(limit).
		Exec(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("erreur récupération messages: %w", err)
	}
	
	messages := make([]*models.SupportMessage, len(prismaMessages))
	for i, pm := range prismaMessages {
		messages[i] = s.convertPrismaToMessage(&pm)
	}
	
	return messages, nil
}

// === STUBS POUR TOUTES LES AUTRES MÉTHODES ===

func (s *supportServicePrisma) ReassignTicket(ticketID, fromUserID, toUserID, reason string, notes *string, reassignedBy string) (*models.ReassignmentLog, error) {
	ctx := context.Background()
	now := time.Now()
	
	// Mettre à jour le ticket d'abord
	_, err := s.db.SupportTicket.FindUnique(
		db.SupportTicket.ID.Equals(ticketID),
	).Update(
		db.SupportTicket.AssignedTo.Set(toUserID),
		db.SupportTicket.Status.Set(db.TicketStatusAssigned),
		db.SupportTicket.ReassignedAt.Set(now),
	).Exec(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("erreur mise à jour ticket réassigné: %w", err)
	}
	
	// TODO: Implémenter la création du log de réassignation
	// L'API Prisma requiert une relation Connect plutôt qu'un ID direct
	log := &models.ReassignmentLog{
		ID:        fmt.Sprintf("rlog_%d", now.Unix()),
		TicketID:  ticketID,
		ToUser:    toUserID,
		Reason:    reason,
		CreatedBy: reassignedBy,
		CreatedAt: now,
	}
	
	if fromUserID != "" {
		log.FromUser = &fromUserID
	}
	if notes != nil {
		log.Notes = notes
	}
	
	logrus.Infof("Ticket %s réassigné de %s à %s par %s (log temporaire)", ticketID, fromUserID, toUserID, reassignedBy)
	return log, nil
}

func (s *supportServicePrisma) CreateInternalGroup(req *models.CreateGroupRequest, createdBy string) (*models.InternalGroupChat, error) {
	ctx := context.Background()
	
	// S'assurer que le créateur est dans les participants
	participants := req.Participants
	found := false
	for _, p := range participants {
		if p == createdBy {
			found = true
			break
		}
	}
	if !found {
		participants = append(participants, createdBy)
	}
	
	createdGroup, err := s.db.InternalGroupChat.CreateOne(
		// Required parameters
		db.InternalGroupChat.Name.Set(req.Name),
		db.InternalGroupChat.Description.Set(req.Description),
		db.InternalGroupChat.CreatedBy.Set(createdBy),
		// Optional parameters
		db.InternalGroupChat.Participants.Set(participants),
	).Exec(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("erreur création groupe: %w", err)
	}
	
	group := &models.InternalGroupChat{
		ID:           createdGroup.ID,
		Name:         createdGroup.Name,
		Description:  createdGroup.Description,
		Participants: createdGroup.Participants,
		CreatedBy:    createdGroup.CreatedBy,
		IsActive:     createdGroup.IsActive,
		CreatedAt:    createdGroup.CreatedAt,
		UpdatedAt:    createdGroup.UpdatedAt,
	}
	
	logrus.Infof("Groupe créé: %s (%s) par %s avec %d participants", group.ID, group.Name, createdBy, len(participants))
	return group, nil
}

func (s *supportServicePrisma) GetUserGroups(userID string) ([]*models.InternalGroupChat, error) {
	ctx := context.Background()
	
	// Rechercher les groupes où l'utilisateur est participant
	// Note: Prisma Go ne supporte pas directement les opérations sur les arrays
	// On récupère tous les groupes actifs et on filtre en mémoire
	prismaGroups, err := s.db.InternalGroupChat.FindMany(
		db.InternalGroupChat.IsActive.Equals(true),
	).OrderBy(
		db.InternalGroupChat.UpdatedAt.Order(db.SortOrderDesc),
	).Exec(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("erreur récupération groupes: %w", err)
	}
	
	// Filtrer en mémoire les groupes où l'utilisateur est participant
	var groups []*models.InternalGroupChat
	for _, pg := range prismaGroups {
		// Vérifier si l'utilisateur fait partie des participants
		isParticipant := false
		for _, participant := range pg.Participants {
			if participant == userID {
				isParticipant = true
				break
			}
		}
		
		if isParticipant {
			groups = append(groups, &models.InternalGroupChat{
				ID:           pg.ID,
				Name:         pg.Name,
				Description:  pg.Description,
				Participants: pg.Participants,
				CreatedBy:    pg.CreatedBy,
				IsActive:     pg.IsActive,
				CreatedAt:    pg.CreatedAt,
				UpdatedAt:    pg.UpdatedAt,
			})
		}
	}
	
	return groups, nil
}

func (s *supportServicePrisma) AddGroupMessage(groupID, senderID, content string) (*models.SupportMessage, error) {
	ctx := context.Background()
	
	// Vérifier que le groupe existe et que l'utilisateur en fait partie
	group, err := s.db.InternalGroupChat.FindUnique(
		db.InternalGroupChat.ID.Equals(groupID),
	).Exec(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("groupe non trouvé: %w", err)
	}
	
	// Vérifier que l'utilisateur fait partie du groupe
	canSend := false
	for _, participant := range group.Participants {
		if participant == senderID {
			canSend = true
			break
		}
	}
	
	if !canSend {
		return nil, fmt.Errorf("utilisateur %s n'est pas autorisé à écrire dans le groupe %s", senderID, groupID)
	}
	
	// Créer le message
	createdMessage, err := s.db.SupportMessage.CreateOne(
		// Required parameters in correct order
		db.SupportMessage.ConversationID.Set(groupID),
		db.SupportMessage.SenderID.Set(senderID),
		db.SupportMessage.SenderRole.Set(db.UserRoleAdmin), // Par défaut pour les groupes internes
		db.SupportMessage.Content.Set(content),
		// Optional parameters
		db.SupportMessage.ConversationType.Set(db.ConversationTypeInternalGroup),
		db.SupportMessage.MessageType.Set("text"),
		db.SupportMessage.IsInternal.Set(true),
		db.SupportMessage.GroupID.SetOptional(&groupID),
	).Exec(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("erreur création message groupe: %w", err)
	}
	
	message := s.convertPrismaToMessage(createdMessage)
	return message, nil
}

func (s *supportServicePrisma) GetGroupMessages(groupID, userID string, page, limit int) ([]*models.SupportMessage, error) {
	ctx := context.Background()
	
	// Vérifier que l'utilisateur fait partie du groupe
	group, err := s.db.InternalGroupChat.FindUnique(
		db.InternalGroupChat.ID.Equals(groupID),
	).Exec(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("groupe non trouvé: %w", err)
	}
	
	canRead := false
	for _, participant := range group.Participants {
		if participant == userID {
			canRead = true
			break
		}
	}
	
	if !canRead {
		return nil, fmt.Errorf("utilisateur %s n'est pas autorisé à lire les messages du groupe %s", userID, groupID)
	}
	
	skip := (page - 1) * limit
	prismaMessages, err := s.db.SupportMessage.FindMany(
		db.SupportMessage.ConversationID.Equals(groupID),
		db.SupportMessage.ConversationType.Equals(db.ConversationTypeInternalGroup),
	).OrderBy(
		db.SupportMessage.CreatedAt.Order(db.SortOrderAsc),
	).Skip(skip).Take(limit).Exec(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("erreur récupération messages groupe: %w", err)
	}
	
	messages := make([]*models.SupportMessage, len(prismaMessages))
	for i, pm := range prismaMessages {
		messages[i] = s.convertPrismaToMessage(&pm)
	}
	
	return messages, nil
}

func (s *supportServicePrisma) AddGroupParticipant(groupID, participantID, addedBy string) error {
	ctx := context.Background()
	
	// Récupérer le groupe existant
	group, err := s.db.InternalGroupChat.FindUnique(
		db.InternalGroupChat.ID.Equals(groupID),
	).Exec(ctx)
	
	if err != nil {
		return fmt.Errorf("groupe non trouvé: %w", err)
	}
	
	// Vérifier que le participant n'est pas déjà dans le groupe
	for _, p := range group.Participants {
		if p == participantID {
			return fmt.Errorf("utilisateur %s fait déjà partie du groupe %s", participantID, groupID)
		}
	}
	
	// Ajouter le nouveau participant
	newParticipants := append(group.Participants, participantID)
	
	// Mettre à jour le groupe
	_, err = s.db.InternalGroupChat.FindUnique(
		db.InternalGroupChat.ID.Equals(groupID),
	).Update(
		db.InternalGroupChat.Participants.Set(newParticipants),
	).Exec(ctx)
	
	if err != nil {
		return fmt.Errorf("erreur ajout participant au groupe: %w", err)
	}
	
	logrus.Infof("Participant %s ajouté au groupe %s par %s", participantID, groupID, addedBy)
	return nil
}

func (s *supportServicePrisma) RemoveGroupParticipant(groupID, participantID, removedBy string) error {
	ctx := context.Background()
	
	// Récupérer le groupe existant
	group, err := s.db.InternalGroupChat.FindUnique(
		db.InternalGroupChat.ID.Equals(groupID),
	).Exec(ctx)
	
	if err != nil {
		return fmt.Errorf("groupe non trouvé: %w", err)
	}
	
	// Retirer le participant
	newParticipants := make([]string, 0, len(group.Participants)-1)
	found := false
	for _, p := range group.Participants {
		if p != participantID {
			newParticipants = append(newParticipants, p)
		} else {
			found = true
		}
	}
	
	if !found {
		return fmt.Errorf("utilisateur %s ne fait pas partie du groupe %s", participantID, groupID)
	}
	
	// Mettre à jour le groupe
	_, err = s.db.InternalGroupChat.FindUnique(
		db.InternalGroupChat.ID.Equals(groupID),
	).Update(
		db.InternalGroupChat.Participants.Set(newParticipants),
	).Exec(ctx)
	
	if err != nil {
		return fmt.Errorf("erreur suppression participant du groupe: %w", err)
	}
	
	logrus.Infof("Participant %s retiré du groupe %s par %s", participantID, groupID, removedBy)
	return nil
}

func (s *supportServicePrisma) InitiateDirectContact(req *models.DirectContactRequest, adminID string) (*models.SupportTicket, error) {
	ctx := context.Background()
	
	// Créer un ticket spécial pour le contact direct
	title := req.Subject
	if title == "" {
		title = fmt.Sprintf("Contact direct avec %s", req.TargetUserID)
	}
	
	createdTicket, err := s.db.SupportTicket.CreateOne(
		db.SupportTicket.Title.Set(title),
		db.SupportTicket.Description.Set(req.InitialMessage),
		db.SupportTicket.Category.Set(db.TicketCategoryOther),
		db.SupportTicket.CreatedBy.Set(adminID),
		db.SupportTicket.CreatedRole.Set(db.UserRoleAdmin),
		// Optional parameters
		db.SupportTicket.Priority.Set(db.TicketPriorityMedium),
		db.SupportTicket.AssignedTo.Set(req.TargetUserID),
		db.SupportTicket.IsInternal.Set(true),
		db.SupportTicket.Participants.Set([]string{adminID, req.TargetUserID}),
	).Exec(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("erreur création ticket contact direct: %w", err)
	}
	
	ticket := s.convertPrismaToTicket(createdTicket)
	
	logrus.Infof("Contact direct initié: %s entre %s et %s", ticket.ID, adminID, req.TargetUserID)
	return ticket, nil
}

func (s *supportServicePrisma) GetReassignmentHistory(ticketID string) ([]*models.ReassignmentLog, error) {
	ctx := context.Background()
	
	prismaLogs, err := s.db.ReassignmentLog.FindMany(
		db.ReassignmentLog.TicketID.Equals(ticketID),
	).OrderBy(
		db.ReassignmentLog.CreatedAt.Order(db.SortOrderDesc),
	).Exec(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("erreur récupération historique réassignations: %w", err)
	}
	
	logs := make([]*models.ReassignmentLog, len(prismaLogs))
	for i, pl := range prismaLogs {
		log := &models.ReassignmentLog{
			ID:        pl.ID,
			TicketID:  pl.TicketID,
			ToUser:    pl.ToUser,
			Reason:    pl.Reason,
			CreatedBy: pl.CreatedBy,
			CreatedAt: pl.CreatedAt,
		}
		
		if fromUser, ok := pl.FromUser(); ok {
			log.FromUser = &fromUser
		}
		if notes, ok := pl.Notes(); ok {
			log.Notes = &notes
		}
		
		logs[i] = log
	}
	
	return logs, nil
}

func (s *supportServicePrisma) GetSupportStats(userID string, role models.UserRole) (*models.SupportStats, error) {
	ctx := context.Background()
	
	stats := &models.SupportStats{}
	
	if role == models.UserRoleClient || role == models.UserRoleLivreur {
		// Pour les clients/livreurs, compter seulement leurs propres tickets
		totalTickets, err := s.db.SupportTicket.FindMany(
			db.SupportTicket.CreatedBy.Equals(userID),
		).Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("erreur calcul stats total tickets: %w", err)
		}
		stats.TotalTickets = len(totalTickets)
		
		openTickets, err := s.db.SupportTicket.FindMany(
			db.SupportTicket.CreatedBy.Equals(userID),
			db.SupportTicket.Status.In([]db.TicketStatus{
				db.TicketStatusOpen,
				db.TicketStatusAssigned,
				db.TicketStatusInProgress,
				db.TicketStatusPending,
			}),
		).Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("erreur calcul stats tickets ouverts: %w", err)
		}
		stats.OpenTickets = len(openTickets)
		stats.MyAssignedTickets = 0 // Les clients n'ont pas de tickets assignés
	} else {
		// Pour le staff, compter tous les tickets
		totalTickets, err := s.db.SupportTicket.FindMany().Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("erreur calcul stats total tickets: %w", err)
		}
		stats.TotalTickets = len(totalTickets)
		
		openTickets, err := s.db.SupportTicket.FindMany(
			db.SupportTicket.Status.In([]db.TicketStatus{
				db.TicketStatusOpen,
				db.TicketStatusAssigned,
				db.TicketStatusInProgress,
				db.TicketStatusPending,
			}),
		).Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("erreur calcul stats tickets ouverts: %w", err)
		}
		stats.OpenTickets = len(openTickets)
		
		// Tickets assignés à cet utilisateur
		assignedTickets, err := s.db.SupportTicket.FindMany(
			db.SupportTicket.AssignedTo.Equals(userID),
			db.SupportTicket.Status.In([]db.TicketStatus{
				db.TicketStatusAssigned,
				db.TicketStatusInProgress,
				db.TicketStatusPending,
			}),
		).Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("erreur calcul stats tickets assignés: %w", err)
		}
		stats.MyAssignedTickets = len(assignedTickets)
	}
	
	return stats, nil
}

// === FONCTIONS DE CONVERSION PRISMA ===

// convertPrismaToTicket convertit un SupportTicket Prisma vers notre modèle
func (s *supportServicePrisma) convertPrismaToTicket(pt *db.SupportTicketModel) *models.SupportTicket {
	ticket := &models.SupportTicket{
		ID:                  pt.ID,
		Title:               pt.Title,
		Description:         pt.Description,
		Category:            models.TicketCategory(pt.Category),
		Priority:            models.TicketPriority(pt.Priority),
		Status:              models.TicketStatus(pt.Status),
		CreatedBy:           pt.CreatedBy,
		CreatedRole:         models.UserRole(pt.CreatedRole),
		PreviouslyAssignedTo: pt.PreviouslyAssignedTo,
		IsInternal:          pt.IsInternal,
		Participants:        pt.Participants,
		CreatedAt:           pt.CreatedAt,
		UpdatedAt:           pt.UpdatedAt,
	}

	// Gérer les champs optionnels
	if assignedTo, ok := pt.AssignedTo(); ok {
		ticket.AssignedTo = &assignedTo
	}
	if relatedID, ok := pt.RelatedDeliveryID(); ok {
		ticket.RelatedDeliveryID = &relatedID
	}
	if teamCh, ok := pt.TeamChannel(); ok {
		ticket.TeamChannel = &teamCh
	}
	if reassignedAt, ok := pt.ReassignedAt(); ok {
		ticket.ReassignedAt = &reassignedAt
	}
	if closedAt, ok := pt.ClosedAt(); ok {
		ticket.ClosedAt = &closedAt
	}

	return ticket
}

// convertPrismaToMessage convertit un SupportMessage Prisma vers notre modèle
func (s *supportServicePrisma) convertPrismaToMessage(pm *db.SupportMessageModel) *models.SupportMessage {
	message := &models.SupportMessage{
		ID:               pm.ID,
		ConversationID:   pm.ConversationID,
		ConversationType: models.ConversationType(pm.ConversationType),
		SenderID:         pm.SenderID,
		SenderRole:       models.UserRole(pm.SenderRole),
		Content:          pm.Content,
		MessageType:      pm.MessageType,
		IsInternal:       pm.IsInternal,
		CreatedAt:        pm.CreatedAt,
		UpdatedAt:        pm.UpdatedAt,
	}

	// Gérer les champs optionnels
	if replyTo, ok := pm.ReplyTo(); ok {
		message.ReplyTo = &replyTo
	}

	return message
}

// === HELPER METHODS ===

// canAccessTicket vérifie si un utilisateur peut accéder à un ticket
func (s *supportServicePrisma) canAccessTicket(ticket *models.SupportTicket, userID string, role models.UserRole) bool {
	// Le staff peut accéder à tous les tickets
	if role == models.UserRoleAdmin || role == models.UserRoleGestionnaire || role == models.UserRoleMarketing {
		return true
	}
	
	// Les clients/livreurs peuvent seulement accéder à leurs propres tickets
	return ticket.CreatedBy == userID
}
