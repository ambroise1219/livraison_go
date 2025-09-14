package middlewares

import (
	"net/http"

	"github.com/ambroise1219/livraison_go/models"
	"github.com/gin-gonic/gin"
)

// RequireRole middleware qui vérifie que l'utilisateur a l'un des rôles requis
func RequireRole(allowedRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Récupérer le rôle de l'utilisateur depuis le contexte
		userRole, exists := GetCurrentUserRole(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentification requise",
			})
			c.Abort()
			return
		}

		// Vérifier si le rôle de l'utilisateur est autorisé
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		// Si on arrive ici, l'utilisateur n'a pas les droits nécessaires
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Accès interdit",
			"message": "Vous n'avez pas les permissions nécessaires pour accéder à cette ressource",
			"required_roles": allowedRoles,
			"your_role": userRole,
		})
		c.Abort()
	}
}

// RequireAdmin middleware spécifique pour les admins
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.UserRoleAdmin)
}

// RequireDriver middleware spécifique pour les livreurs
func RequireDriver() gin.HandlerFunc {
	return RequireRole(models.UserRoleLivreur)
}

// RequireClient middleware spécifique pour les clients
func RequireClient() gin.HandlerFunc {
	return RequireRole(models.UserRoleClient)
}

// RequireDriverOrAdmin middleware pour livreurs ou admins
func RequireDriverOrAdmin() gin.HandlerFunc {
	return RequireRole(models.UserRoleLivreur, models.UserRoleAdmin)
}

// RequireClientOrAdmin middleware pour clients ou admins
func RequireClientOrAdmin() gin.HandlerFunc {
	return RequireRole(models.UserRoleClient, models.UserRoleAdmin)
}

// RequireAuthenticatedUser middleware qui vérifie seulement qu'un utilisateur est connecté
func RequireAuthenticatedUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentification requise",
				"message": "Vous devez être connecté pour accéder à cette ressource",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireResourceOwner middleware qui vérifie que l'utilisateur accède à ses propres ressources
func RequireResourceOwner(resourceUserIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Récupérer l'ID de l'utilisateur connecté
		currentUserID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentification requise",
			})
			c.Abort()
			return
		}

		// Récupérer l'ID de la ressource depuis les paramètres de la route
		resourceUserID := c.Param(resourceUserIDParam)
		if resourceUserID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "ID utilisateur requis",
			})
			c.Abort()
			return
		}

		// Vérifier si l'utilisateur est admin (accès à toutes les ressources)
		userRole, _ := GetCurrentUserRole(c)
		if userRole == models.UserRoleAdmin {
			c.Next()
			return
		}

		// Vérifier si l'utilisateur accède à ses propres ressources
		if currentUserID != resourceUserID {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Accès interdit",
				"message": "Vous ne pouvez accéder qu'à vos propres ressources",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireDriverStatus middleware qui vérifie le statut du livreur
func RequireDriverStatus(allowedStatuses ...models.DriverStatus) gin.HandlerFunc {
	return func(c *gin.Context) {
		// D'abord vérifier que c'est un livreur
		userRole, exists := GetCurrentUserRole(c)
		if !exists || userRole != models.UserRoleLivreur {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Accès réservé aux livreurs",
			})
			c.Abort()
			return
		}

		// TODO: Récupérer le statut du livreur depuis la base de données
		// Pour l'instant, on assume que le livreur est actif
		// Dans une vraie implémentation, il faudrait faire une requête à SurrealDB
		
		// Exemple de ce qu'on ferait :
		// userID, _ := GetCurrentUserID(c)
		// driver, err := services.GetDriverByUserID(userID)
		// if err != nil || driver == nil {
		//     c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de récupérer le statut du livreur"})
		//     c.Abort()
		//     return
		// }
		
		// Pour l'instant, on laisse passer
		c.Next()
	}
}
