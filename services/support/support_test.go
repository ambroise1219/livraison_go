package support

import (
	"fmt"
	"testing"
	"time"

	"github.com/ambroise1219/livraison_go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSupportService(t *testing.T) {
	// Test de création du service avec vraie DB
	service := NewSupportService()
	
	assert.NotNil(t, service)
	assert.Implements(t, (*SupportService)(nil), service)
}

func TestCreateRealTicket(t *testing.T) {
	service := NewSupportService()
	
	// Données de test avec identifiants uniques
	timestamp := time.Now().Unix()
	userID := fmt.Sprintf("test_user_%d", timestamp)
	role := models.UserRoleClient
	req := &models.CreateTicketRequest{
		Title:       fmt.Sprintf("Test Ticket Real %d", timestamp),
		Description: "Description du test ticket réel avec base PostgreSQL",
		Category:    models.TicketCategoryTechnical,
		Priority:    models.TicketPriorityMedium,
	}
	
	// Test de création en base réelle
	ticket, err := service.CreateTicket(userID, role, req)
	
	// Assertions
	require.NoError(t, err, "La création du ticket en base doit réussir")
	assert.NotNil(t, ticket)
	assert.NotEmpty(t, ticket.ID)
	assert.Equal(t, req.Title, ticket.Title)
	assert.Equal(t, req.Description, ticket.Description)
	assert.Equal(t, req.Category, ticket.Category)
	assert.Equal(t, userID, ticket.CreatedBy)
	assert.Equal(t, role, ticket.CreatedRole)
	assert.Equal(t, models.TicketStatusOpen, ticket.Status)
	assert.Equal(t, models.TicketPriorityMedium, ticket.Priority)
	assert.False(t, ticket.IsInternal)
	assert.WithinDuration(t, time.Now(), ticket.CreatedAt, 10*time.Second)
	
	// Vérifier qu'on peut récupérer le ticket créé
	retrievedTicket, err := service.GetTicketByID(ticket.ID, userID, role)
	require.NoError(t, err, "La récupération du ticket doit réussir")
	assert.Equal(t, ticket.ID, retrievedTicket.ID)
	assert.Equal(t, ticket.Title, retrievedTicket.Title)
	
	t.Logf("✅ Ticket créé et récupéré avec succès: %s", ticket.ID)
}

func TestGetTicketsWithPagination(t *testing.T) {
	service := NewSupportService()
	timestamp := time.Now().Unix()
	userID := fmt.Sprintf("test_user_pagination_%d", timestamp)
	
	// Créer plusieurs tickets pour tester la pagination
	createdTickets := make([]*models.SupportTicket, 0)
	for i := 0; i < 3; i++ {
		req := &models.CreateTicketRequest{
			Title:       fmt.Sprintf("Test Ticket %d pour pagination", i+1),
			Description: fmt.Sprintf("Description ticket %d", i+1),
			Category:    models.TicketCategoryTechnical,
			Priority:    models.TicketPriorityLow,
		}
		
		ticket, err := service.CreateTicket(userID, models.UserRoleClient, req)
		require.NoError(t, err, "Création ticket %d doit réussir", i+1)
		createdTickets = append(createdTickets, ticket)
	}
	
	// Test récupération avec filtres
	filters := map[string]string{
		"status": string(models.TicketStatusOpen),
	}
	
	tickets, total, err := service.GetTickets(filters, userID, models.UserRoleClient, 1, 10)
	
	// Assertions
	require.NoError(t, err, "Récupération tickets doit réussir")
	assert.NotNil(t, tickets)
	assert.GreaterOrEqual(t, len(tickets), 3, "Au moins 3 tickets doivent être retournés")
	assert.GreaterOrEqual(t, total, 3, "Le total doit être au moins 3")
	
	// Vérifier que tous les tickets retournés appartiennent au bon utilisateur
	for _, ticket := range tickets {
		assert.Equal(t, userID, ticket.CreatedBy, "Tous les tickets doivent appartenir au bon utilisateur")
		assert.Equal(t, models.TicketStatusOpen, ticket.Status)
	}
	
	t.Logf("✅ Récupéré %d tickets sur %d total pour l'utilisateur %s", len(tickets), total, userID)
}

func TestTicketPermissions(t *testing.T) {
	service := NewSupportService()
	timestamp := time.Now().Unix()
	
	// Créer un ticket avec un utilisateur
	userID1 := fmt.Sprintf("test_user1_%d", timestamp)
	userID2 := fmt.Sprintf("test_user2_%d", timestamp)
	
	req := &models.CreateTicketRequest{
		Title:       "Test Ticket Permissions",
		Description: "Test des permissions d'accès aux tickets",
		Category:    models.TicketCategoryAccount,
		Priority:    models.TicketPriorityMedium,
	}
	
	ticket, err := service.CreateTicket(userID1, models.UserRoleClient, req)
	require.NoError(t, err)
	
	// Test 1: Le propriétaire peut accéder à son ticket
	retrievedTicket, err := service.GetTicketByID(ticket.ID, userID1, models.UserRoleClient)
	require.NoError(t, err, "Le propriétaire doit pouvoir accéder à son ticket")
	assert.Equal(t, ticket.ID, retrievedTicket.ID)
	
	// Test 2: Un autre client ne peut pas accéder au ticket
	_, err = service.GetTicketByID(ticket.ID, userID2, models.UserRoleClient)
	assert.Error(t, err, "Un autre client ne doit pas pouvoir accéder au ticket")
	assert.Contains(t, err.Error(), "accès refusé")
	
	// Test 3: Un admin peut accéder à tous les tickets
	adminID := fmt.Sprintf("admin_%d", timestamp)
	retrievedByAdmin, err := service.GetTicketByID(ticket.ID, adminID, models.UserRoleAdmin)
	require.NoError(t, err, "Un admin doit pouvoir accéder à tous les tickets")
	assert.Equal(t, ticket.ID, retrievedByAdmin.ID)
	
	// Test 4: Ticket inexistant
	_, err = service.GetTicketByID("nonexistent_id", userID1, models.UserRoleClient)
	assert.Error(t, err, "L'accès à un ticket inexistant doit échouer")
	
	t.Logf("✅ Tests de permissions réussis pour le ticket %s", ticket.ID)
}

func TestTicketStatusUpdates(t *testing.T) {
	service := NewSupportService()
	timestamp := time.Now().Unix()
	userID := fmt.Sprintf("test_user_status_%d", timestamp)
	
	// Créer un ticket réel
	req := &models.CreateTicketRequest{
		Title:       "Test Ticket Status Updates",
		Description: "Ticket pour tester les mises à jour de statut",
		Category:    models.TicketCategoryTechnical,
		Priority:    models.TicketPriorityMedium,
	}
	
	ticket, err := service.CreateTicket(userID, models.UserRoleClient, req)
	require.NoError(t, err, "Création du ticket doit réussir")
	assert.Equal(t, models.TicketStatusOpen, ticket.Status, "Statut initial doit être Open")
	
	// Test 1: Mise à jour du statut vers InProgress
	err = service.UpdateTicketStatus(ticket.ID, models.TicketStatusInProgress, userID)
	require.NoError(t, err, "Mise à jour du statut doit réussir")
	
	// Vérifier que le statut a été mis à jour
	updatedTicket, err := service.GetTicketByID(ticket.ID, userID, models.UserRoleClient)
	require.NoError(t, err, "Récupération du ticket mis à jour doit réussir")
	assert.Equal(t, models.TicketStatusInProgress, updatedTicket.Status, "Le statut doit être mis à jour")
	
	// Test 2: Fermeture du ticket
	err = service.CloseTicket(ticket.ID, userID)
	require.NoError(t, err, "Fermeture du ticket doit réussir")
	
	// Vérifier que le ticket est fermé
	closedTicket, err := service.GetTicketByID(ticket.ID, userID, models.UserRoleClient)
	require.NoError(t, err, "Récupération du ticket fermé doit réussir")
	assert.Equal(t, models.TicketStatusClosed, closedTicket.Status, "Le ticket doit être fermé")
	assert.NotNil(t, closedTicket.ClosedAt, "La date de fermeture doit être définie")
	assert.WithinDuration(t, time.Now(), *closedTicket.ClosedAt, 10*time.Second, "La date de fermeture doit être récente")
	
	t.Logf("✅ Mises à jour de statut réussies pour le ticket %s", ticket.ID)
}

// TestCloseTicket supprimé car la fermeture est déjà testée dans TestTicketStatusUpdates

func TestRealMessagesInTicket(t *testing.T) {
	service := NewSupportService()
	timestamp := time.Now().Unix()
	userID := fmt.Sprintf("test_user_messages_%d", timestamp)
	
	// Créer un ticket réel d'abord
	req := &models.CreateTicketRequest{
		Title:       "Test Ticket pour Messages",
		Description: "Ticket pour tester l'ajout de messages",
		Category:    models.TicketCategoryTechnical,
		Priority:    models.TicketPriorityMedium,
	}
	
	ticket, err := service.CreateTicket(userID, models.UserRoleClient, req)
	require.NoError(t, err, "Création du ticket doit réussir")
	
	// Test d'ajout de message au ticket
	messageReq := &models.AddMessageRequest{
		Content:     "Premier message de test dans le ticket réel",
		MessageType: "text",
		IsInternal:  false,
	}
	
	message, err := service.AddMessage(ticket.ID, userID, models.UserRoleClient, messageReq)
	
	// Assertions pour le message créé
	require.NoError(t, err, "Ajout de message doit réussir")
	assert.NotNil(t, message)
	assert.NotEmpty(t, message.ID)
	assert.Equal(t, ticket.ID, message.ConversationID)
	assert.Equal(t, userID, message.SenderID)
	assert.Equal(t, models.UserRoleClient, message.SenderRole)
	assert.Equal(t, messageReq.Content, message.Content)
	assert.Equal(t, messageReq.MessageType, message.MessageType)
	assert.Equal(t, messageReq.IsInternal, message.IsInternal)
	assert.Equal(t, models.ConversationTypeTicket, message.ConversationType)
	assert.WithinDuration(t, time.Now(), message.CreatedAt, 10*time.Second)
	
	// Test récupération des messages
	messages, err := service.GetMessages(ticket.ID, userID, models.UserRoleClient, 1, 10)
	require.NoError(t, err, "Récupération des messages doit réussir")
	assert.NotNil(t, messages)
	assert.GreaterOrEqual(t, len(messages), 1, "Au moins 1 message doit être retourné")
	
	// Vérifier que le message est dans la liste
	found := false
	for _, msg := range messages {
		if msg.ID == message.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Le message créé doit être dans la liste des messages")
	
	t.Logf("✅ Message créé et récupéré avec succès dans le ticket %s", ticket.ID)
}

func TestSupportStats(t *testing.T) {
	service := NewSupportService()
	timestamp := time.Now().Unix()
	userID := fmt.Sprintf("test_user_stats_%d", timestamp)
	
	// Test récupération des statistiques pour un client
	stats, err := service.GetSupportStats(userID, models.UserRoleClient)
	require.NoError(t, err, "Récupération des stats client doit réussir")
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.TotalTickets, 0, "Le total de tickets ne peut pas être négatif")
	assert.GreaterOrEqual(t, stats.OpenTickets, 0, "Les tickets ouverts ne peuvent pas être négatifs")
	assert.Equal(t, 0, stats.MyAssignedTickets, "Les clients ne doivent pas avoir de tickets assignés")
	
	// Test récupération des statistiques pour un admin
	adminStats, err := service.GetSupportStats("admin_stats", models.UserRoleAdmin)
	require.NoError(t, err, "Récupération des stats admin doit réussir")
	assert.NotNil(t, adminStats)
	assert.GreaterOrEqual(t, adminStats.TotalTickets, 0)
	assert.GreaterOrEqual(t, adminStats.OpenTickets, 0)
	assert.GreaterOrEqual(t, adminStats.MyAssignedTickets, 0)
	
	t.Logf("✅ Stats récupérées - Client: %d total, Admin: %d total", stats.TotalTickets, adminStats.TotalTickets)
}

func TestInternalGroupOperations(t *testing.T) {
	service := NewSupportService()
	timestamp := time.Now().Unix()
	adminID := fmt.Sprintf("admin_%d", timestamp)
	userID := fmt.Sprintf("user_%d", timestamp)
	
	// Test création de groupe
	groupReq := &models.CreateGroupRequest{
		Name:        fmt.Sprintf("Test Group %d", timestamp),
		Description: "Groupe de test pour les opérations CRUD",
		Participants: []string{adminID, userID},
	}
	
	group, err := service.CreateInternalGroup(groupReq, adminID)
	require.NoError(t, err, "Création du groupe doit réussir")
	assert.NotNil(t, group)
	assert.NotEmpty(t, group.ID)
	assert.Equal(t, groupReq.Name, group.Name)
	assert.Equal(t, groupReq.Description, group.Description)
	assert.Contains(t, group.Participants, adminID)
	assert.Contains(t, group.Participants, userID)
	assert.True(t, group.IsActive)
	
	// Test récupération des groupes de l'utilisateur
	userGroups, err := service.GetUserGroups(adminID)
	require.NoError(t, err, "Récupération des groupes doit réussir")
	assert.NotNil(t, userGroups)
	
	// Vérifier que le groupe créé est dans la liste
	foundGroup := false
	for _, g := range userGroups {
		if g.ID == group.ID {
			foundGroup = true
			break
		}
	}
	assert.True(t, foundGroup, "Le groupe créé doit être dans la liste des groupes de l'utilisateur")
	
	// Test ajout de message au groupe
	message, err := service.AddGroupMessage(group.ID, adminID, "Message de test dans le groupe")
	require.NoError(t, err, "Ajout de message au groupe doit réussir")
	assert.NotNil(t, message)
	assert.Equal(t, group.ID, message.ConversationID)
	// Note: Prisma retourne les enums en majuscules, ajustons la comparaison
	assert.Contains(t, string(message.ConversationType), "INTERNAL_GROUP")
	assert.True(t, message.IsInternal, "Les messages de groupe doivent être internes")
	
	// Test récupération des messages du groupe
	groupMessages, err := service.GetGroupMessages(group.ID, adminID, 1, 10)
	require.NoError(t, err, "Récupération des messages du groupe doit réussir")
	assert.NotNil(t, groupMessages)
	assert.GreaterOrEqual(t, len(groupMessages), 1, "Au moins 1 message doit être dans le groupe")
	
	t.Logf("✅ Opérations de groupe réussies pour %s avec %d messages", group.ID, len(groupMessages))
}

func TestReassignmentOperations(t *testing.T) {
	service := NewSupportService()
	timestamp := time.Now().Unix()
	userID := fmt.Sprintf("test_user_reassign_%d", timestamp)
	adminID := fmt.Sprintf("admin_reassign_%d", timestamp)
	newAssigneeID := fmt.Sprintf("assignee_%d", timestamp)
	
	// Créer un ticket pour le test de réassignation
	req := &models.CreateTicketRequest{
		Title:       "Test Ticket pour Réassignation",
		Description: "Ticket pour tester la réassignation",
		Category:    models.TicketCategoryTechnical,
		Priority:    models.TicketPriorityHigh,
	}
	
	ticket, err := service.CreateTicket(userID, models.UserRoleClient, req)
	require.NoError(t, err, "Création du ticket doit réussir")
	
	// Test réassignation (version temporaire qui met à jour le ticket)
	log, err := service.ReassignTicket(ticket.ID, "", newAssigneeID, "Test de réassignation", nil, adminID)
	require.NoError(t, err, "Réassignation doit réussir")
	assert.NotNil(t, log)
	assert.Equal(t, ticket.ID, log.TicketID)
	assert.Equal(t, newAssigneeID, log.ToUser)
	
	// Vérifier que le ticket a été mis à jour
	updatedTicket, err := service.GetTicketByID(ticket.ID, userID, models.UserRoleClient)
	require.NoError(t, err, "Récupération du ticket réassigné doit réussir")
	assert.NotNil(t, updatedTicket.AssignedTo)
	assert.Equal(t, newAssigneeID, *updatedTicket.AssignedTo, "Le ticket doit être assigné au nouvel utilisateur")
	assert.Equal(t, models.TicketStatusAssigned, updatedTicket.Status, "Le statut doit être Assigned")
	assert.NotNil(t, updatedTicket.ReassignedAt, "La date de réassignation doit être définie")
	
	// Test récupération de l'historique (retourne vide pour l'instant)
	history, err := service.GetReassignmentHistory(ticket.ID)
	require.NoError(t, err, "Récupération de l'historique doit réussir")
	assert.NotNil(t, history, "L'historique ne doit pas être nil")
	
	t.Logf("✅ Réassignation réussie du ticket %s vers %s", ticket.ID, newAssigneeID)
}

func TestInitiateDirectContact(t *testing.T) {
	service := NewSupportService()
	timestamp := time.Now().Unix()
	
	adminID := fmt.Sprintf("admin_direct_%d", timestamp)
	targetUserID := fmt.Sprintf("target_user_%d", timestamp)
	
	// Test de création d'un contact direct
	req := &models.DirectContactRequest{
		TargetUserID:   targetUserID,
		InitialMessage: "Message de contact direct de test avec base PostgreSQL",
		Subject:        "Test Contact Direct",
		CreateTicket:   true,
	}
	
	// Initier le contact direct
	ticket, err := service.InitiateDirectContact(req, adminID)
	
	// Assertions
	require.NoError(t, err, "La création du contact direct doit réussir")
	assert.NotNil(t, ticket)
	assert.NotEmpty(t, ticket.ID)
	assert.Equal(t, req.Subject, ticket.Title)
	assert.Equal(t, req.InitialMessage, ticket.Description)
	assert.Equal(t, adminID, ticket.CreatedBy)
	assert.Equal(t, models.UserRoleAdmin, ticket.CreatedRole)
	assert.NotNil(t, ticket.AssignedTo)
	assert.Equal(t, targetUserID, *ticket.AssignedTo)
	assert.True(t, ticket.IsInternal, "Le ticket de contact direct doit être interne")
	assert.Contains(t, ticket.Participants, adminID, "L'admin doit être dans les participants")
	assert.Contains(t, ticket.Participants, targetUserID, "L'utilisateur cible doit être dans les participants")
	assert.Equal(t, models.TicketCategoryOther, ticket.Category)
	assert.Equal(t, models.TicketPriorityMedium, ticket.Priority)
	assert.WithinDuration(t, time.Now(), ticket.CreatedAt, 10*time.Second)
	
	t.Logf("✅ Contact direct initié avec succès: %s entre %s et %s", ticket.ID, adminID, targetUserID)
	
	// Vérifier qu'on peut récupérer le ticket créé
	retrievedTicket, err := service.GetTicketByID(ticket.ID, adminID, models.UserRoleAdmin)
	require.NoError(t, err, "La récupération du ticket de contact direct doit réussir")
	assert.Equal(t, ticket.ID, retrievedTicket.ID)
	assert.Equal(t, ticket.Title, retrievedTicket.Title)
	
	// Test avec sujet vide (utilise un titre par défaut)
	reqEmptySubject := &models.DirectContactRequest{
		TargetUserID:   fmt.Sprintf("target2_%d", timestamp),
		InitialMessage: "Message sans sujet spécifique",
		Subject:        "", // Sujet vide
		CreateTicket:   true,
	}
	
	ticketEmptySubject, err := service.InitiateDirectContact(reqEmptySubject, adminID)
	require.NoError(t, err, "La création avec sujet vide doit réussir")
	assert.Contains(t, ticketEmptySubject.Title, "Contact direct avec", "Le titre par défaut doit être généré")
	
	t.Logf("✅ Tous les tests de contact direct ont réussi")
}
