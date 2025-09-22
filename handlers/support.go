package handlers

import (
	"net/http"
	"strconv"

	"github.com/ambroise1219/livraison_go/models"
	"github.com/ambroise1219/livraison_go/services/support"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Support service instance globale
var supportServiceInstance support.SupportService

// InitSupportHandlers initialise les handlers de support
func InitSupportHandlers() {
	supportServiceInstance = support.NewSupportService()
}

// === TICKET HANDLERS ===

// CreateSupportTicket crée un nouveau ticket de support
func CreateSupportTicket(c *gin.Context) {
	// Récupérer les claims JWT
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non autorisé"})
		return
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
		return
	}

	// Valider les données d'entrée
	var req models.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	// Créer le ticket
	ticket, err := supportServiceInstance.CreateTicket(jwtClaims.UserID, jwtClaims.Role, &req)
	if err != nil {
		logrus.Errorf("Erreur création ticket: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création ticket"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"ticket":  ticket,
		"message": "Ticket créé avec succès",
	})
}

// GetSupportTickets récupère les tickets avec filtres
func GetSupportTickets(c *gin.Context) {
	// Récupérer les claims JWT
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non autorisé"})
		return
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
		return
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Filtres
	filters := map[string]string{
		"status":      c.Query("status"),
		"category":    c.Query("category"),
		"priority":    c.Query("priority"),
		"assigned_to": c.Query("assigned_to"),
	}

	// Supprimer les filtres vides
	for key, value := range filters {
		if value == "" {
			delete(filters, key)
		}
	}

	// Récupérer les tickets
	tickets, total, err := supportServiceInstance.GetTickets(filters, jwtClaims.UserID, jwtClaims.Role, page, limit)
	if err != nil {
		logrus.Errorf("Erreur récupération tickets: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération tickets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tickets": tickets,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetSupportTicketByID récupère un ticket par ID
func GetSupportTicketByID(c *gin.Context) {
	ticketID := c.Param("ticket_id")
	if ticketID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID ticket requis"})
		return
	}

	// Récupérer les claims JWT
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non autorisé"})
		return
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
		return
	}

	// Récupérer le ticket
	ticket, err := supportServiceInstance.GetTicketByID(ticketID, jwtClaims.UserID, jwtClaims.Role)
	if err != nil {
		logrus.Errorf("Erreur récupération ticket: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket non trouvé"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"ticket":  ticket,
	})
}

// UpdateSupportTicketStatus met à jour le statut d'un ticket
func UpdateSupportTicketStatus(c *gin.Context) {
	ticketID := c.Param("ticket_id")
	if ticketID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID ticket requis"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Statut requis"})
		return
	}

	// Récupérer les claims JWT
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non autorisé"})
		return
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
		return
	}

	// Vérifier que le statut est valide
	status := models.TicketStatus(req.Status)
	if !status.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Statut invalide"})
		return
	}

	// Mettre à jour le statut
	err := supportServiceInstance.UpdateTicketStatus(ticketID, status, jwtClaims.UserID)
	if err != nil {
		logrus.Errorf("Erreur mise à jour statut: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur mise à jour statut"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Statut mis à jour avec succès",
	})
}

// === MESSAGE HANDLERS ===

// AddSupportMessage ajoute un message à un ticket
func AddSupportMessage(c *gin.Context) {
	conversationID := c.Param("ticket_id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID conversation requis"})
		return
	}

	var req models.AddMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	// Récupérer les claims JWT
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non autorisé"})
		return
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
		return
	}

	// Ajouter le message
	message, err := supportServiceInstance.AddMessage(conversationID, jwtClaims.UserID, jwtClaims.Role, &req)
	if err != nil {
		logrus.Errorf("Erreur ajout message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur ajout message"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": message,
	})
}

// GetSupportMessages récupère les messages d'une conversation
func GetSupportMessages(c *gin.Context) {
	conversationID := c.Param("ticket_id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID conversation requis"})
		return
	}

	// Récupérer les claims JWT
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non autorisé"})
		return
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
		return
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Récupérer les messages
	messages, err := supportServiceInstance.GetMessages(conversationID, jwtClaims.UserID, jwtClaims.Role, page, limit)
	if err != nil {
		logrus.Errorf("Erreur récupération messages: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"messages": messages,
	})
}

// === STATS HANDLERS ===

// GetSupportStats récupère les statistiques de support
func GetSupportStats(c *gin.Context) {
	// Récupérer les claims JWT
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non autorisé"})
		return
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
		return
	}

	// Récupérer les statistiques
	stats, err := supportServiceInstance.GetSupportStats(jwtClaims.UserID, jwtClaims.Role)
	if err != nil {
		logrus.Errorf("Erreur récupération stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération statistiques"})
		return
	}

c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
	})
}

// === HANDLERS MANQUANTS ===

// GetReassignmentHistory récupère l'historique de réassignation d'un ticket
func GetReassignmentHistory(c *gin.Context) {
	ticketID := c.Param("ticket_id")
	if ticketID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID ticket requis"})
		return
	}

	// Récupérer l'historique
	history, err := supportServiceInstance.GetReassignmentHistory(ticketID)
	if err != nil {
		logrus.Errorf("Erreur récupération historique: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération historique"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"history": history,
	})
}

// ReassignSupportTicket réassigne un ticket à un autre membre du staff
func ReassignSupportTicket(c *gin.Context) {
	ticketID := c.Param("ticket_id")
	if ticketID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID ticket requis"})
		return
	}

	var req struct {
		FromUserID *string `json:"from_user_id"`
		ToUserID   string  `json:"to_user_id" binding:"required"`
		Reason     string  `json:"reason" binding:"required"`
		Notes      *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	// Récupérer les claims JWT
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non autorisé"})
		return
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
		return
	}

	// Si fromUserID n'est pas spécifié, utiliser l'utilisateur courant
	fromUserID := ""
	if req.FromUserID != nil {
		fromUserID = *req.FromUserID
	}

	// Réassigner le ticket
	reassignmentLog, err := supportServiceInstance.ReassignTicket(
		ticketID, fromUserID, req.ToUserID, req.Reason, req.Notes, jwtClaims.UserID,
	)
	if err != nil {
		logrus.Errorf("Erreur réassignation ticket: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur réassignation ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"reassignment_log": reassignmentLog,
		"message":          "Ticket réassigné avec succès",
	})
}

// CreateInternalGroup crée un nouveau groupe de chat interne
func CreateInternalGroup(c *gin.Context) {
	var req models.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	// Récupérer les claims JWT
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non autorisé"})
		return
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
		return
	}

	// Créer le groupe
	group, err := supportServiceInstance.CreateInternalGroup(&req, jwtClaims.UserID)
	if err != nil {
		logrus.Errorf("Erreur création groupe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création groupe"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"group":   group,
		"message": "Groupe créé avec succès",
	})
}

// GetInternalGroups récupère les groupes d'un utilisateur
func GetInternalGroups(c *gin.Context) {
	// Récupérer les claims JWT
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non autorisé"})
		return
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
		return
	}

	// Récupérer les groupes
	groups, err := supportServiceInstance.GetUserGroups(jwtClaims.UserID)
	if err != nil {
		logrus.Errorf("Erreur récupération groupes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération groupes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"groups":  groups,
	})
}

// GetGroupMessages récupère les messages d'un groupe
func GetGroupMessages(c *gin.Context) {
	groupID := c.Param("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID groupe requis"})
		return
	}

	// Récupérer les claims JWT
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non autorisé"})
		return
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
		return
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Récupérer les messages
	messages, err := supportServiceInstance.GetGroupMessages(groupID, jwtClaims.UserID, page, limit)
	if err != nil {
		logrus.Errorf("Erreur récupération messages groupe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur récupération messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"messages": messages,
	})
}

// AddGroupMessage ajoute un message à un groupe
func AddGroupMessage(c *gin.Context) {
	groupID := c.Param("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID groupe requis"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contenu requis"})
		return
	}

	// Récupérer les claims JWT
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non autorisé"})
		return
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
		return
	}

	// Ajouter le message
	message, err := supportServiceInstance.AddGroupMessage(groupID, jwtClaims.UserID, req.Content)
	if err != nil {
		logrus.Errorf("Erreur ajout message groupe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur ajout message"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": message,
	})
}

// InitiateDirectContact initie un contact direct admin -> utilisateur
func InitiateDirectContact(c *gin.Context) {
	var req models.DirectContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	// Récupérer les claims JWT
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Non autorisé"})
		return
	}

	jwtClaims, ok := claims.(*models.JWTClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalide"})
		return
	}

	// Initier le contact direct
	ticket, err := supportServiceInstance.InitiateDirectContact(&req, jwtClaims.UserID)
	if err != nil {
		logrus.Errorf("Erreur contact direct: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur initiation contact"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"ticket":  ticket,
		"message": "Contact direct initié avec succès",
	})
}
