package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

// UserService gère la création et la recherche d'utilisateurs
type UserService struct{}

// NewUserService crée une nouvelle instance du service utilisateur
func NewUserService() *UserService {
	return &UserService{}
}

// FindOrCreateUser trouve ou crée un utilisateur par numéro de téléphone
func (s *UserService) FindOrCreateUser(phone string) (*models.User, bool, error) {
	// Chercher l'utilisateur existant
	user, err := s.findUserByPhone(phone)
	if err != nil {
		return nil, false, fmt.Errorf("échec de la recherche de l'utilisateur: %v", err)
	}

	if user != nil {
		return user, false, nil // Utilisateur existant
	}

	// Créer un nouvel utilisateur
	user, err = s.createUser(phone)
	if err != nil {
		return nil, false, fmt.Errorf("échec de la création de l'utilisateur: %v", err)
	}

	return user, true, nil // Nouvel utilisateur créé
}

// findUserByPhone recherche un utilisateur par numéro de téléphone
func (s *UserService) findUserByPhone(phone string) (*models.User, error) {
	ctx := context.Background()

	user, err := db.PrismaDB.User.FindFirst(
		prismadb.User.Phone.Equals(phone),
	).Exec(ctx)

	if err != nil {
		if err == prismadb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return s.convertPrismaUserToModel(user), nil
}

// createUser crée un nouvel utilisateur
func (s *UserService) createUser(phone string) (*models.User, error) {
	ctx := context.Background()

	user, err := db.PrismaDB.User.CreateOne(
		prismadb.User.Phone.Set(phone),
		prismadb.User.Role.Set(prismadb.UserRoleClient),
		prismadb.User.DriverStatus.Set(prismadb.DriverStatusOffline),
		prismadb.User.LastName.Set(""),
		prismadb.User.FirstName.Set(""),
		prismadb.User.IsProfileCompleted.Set(false),
		prismadb.User.IsDriverComplete.Set(false),
		prismadb.User.IsDriverVehiculeComplete.Set(false),
	).Exec(ctx)

	if err != nil {
		return nil, err
	}

	return s.convertPrismaUserToModel(user), nil
}

// convertPrismaUserToModel convertit un utilisateur Prisma en modèle
func (s *UserService) convertPrismaUserToModel(user *prismadb.UserModel) *models.User {
	modelUser := &models.User{
		ID:                       user.ID,
		Phone:                    user.Phone,
		Role:                     models.UserRole(user.Role),
		CreatedAt:                user.CreatedAt,
		UpdatedAt:                user.UpdatedAt,
		DriverStatus:             models.DriverStatus(user.DriverStatus),
		IsProfileCompleted:       user.IsProfileCompleted,
		IsDriverComplete:         user.IsDriverComplete,
		IsDriverVehiculeComplete: user.IsDriverVehiculeComplete,
		FirstName:                user.FirstName,
		LastName:                 user.LastName,
	}

	// Handle nullable fields
	if address, ok := user.Address(); ok {
		modelUser.Address = &address
	}
	if referredByID, ok := user.ReferredByID(); ok {
		modelUser.ReferredByID = &referredByID
	}
	if profilePictureID, ok := user.ProfilePictureID(); ok {
		modelUser.ProfilePictureID = &profilePictureID
	}
	if email, ok := user.Email(); ok {
		modelUser.Email = &email
	}
	if dateOfBirth, ok := user.DateOfBirth(); ok {
		modelUser.DateOfBirth = &dateOfBirth
	}
	if lieuResidence, ok := user.LieuResidence(); ok {
		modelUser.LieuResidence = &lieuResidence
	}
	if cniRecto, ok := user.CniRecto(); ok {
		modelUser.CNIRecto = &cniRecto
	}
	if cniVerso, ok := user.CniVerso(); ok {
		modelUser.CNIVerso = &cniVerso
	}
	if permisRecto, ok := user.PermisRecto(); ok {
		modelUser.PermisRecto = &permisRecto
	}
	if permisVerso, ok := user.PermisVerso(); ok {
		modelUser.PermisVerso = &permisVerso
	}
	if lastKnownLat, ok := user.LastKnownLat(); ok {
		modelUser.LastKnownLat = &lastKnownLat
	}
	if lastKnownLng, ok := user.LastKnownLng(); ok {
		modelUser.LastKnownLng = &lastKnownLng
	}
	if lastSeenAt, ok := user.LastSeenAt(); ok {
		modelUser.LastSeenAt = &lastSeenAt
	}

	return modelUser
}

// parseUserFromDB parse un utilisateur depuis les données de la base
func (s *UserService) parseUserFromDB(data map[string]interface{}) *models.User {
	user := &models.User{}

	// ID
	if id, ok := data["id"].(string); ok {
		user.ID = id
	}

	// Phone
	if phone, ok := data["phone"].(string); ok {
		user.Phone = phone
	}

	// FirstName
	if firstName, ok := data["firstName"].(string); ok {
		user.FirstName = firstName
	}

	// LastName
	if lastName, ok := data["lastName"].(string); ok {
		user.LastName = lastName
	}

	// Email
	if email, ok := data["email"].(string); ok {
		user.Email = &email
	}

	// TODO: Ajouter d'autres champs selon les besoins

	// Role
	if role, ok := data["role"].(string); ok {
		user.Role = models.UserRole(role)
	}

	// DriverStatus
	if driverStatus, ok := data["driverStatus"].(string); ok {
		user.DriverStatus = models.DriverStatus(driverStatus)
	}

	// CreatedAt
	if createdAt, ok := data["createdAt"].(time.Time); ok {
		user.CreatedAt = createdAt
	}

	// UpdatedAt
	if updatedAt, ok := data["updatedAt"].(time.Time); ok {
		user.UpdatedAt = updatedAt
	}

	return user
}

// GetUserByPhone récupère un utilisateur par son numéro de téléphone
func (s *UserService) GetUserByPhone(phone string) (*models.User, error) {
	user, err := s.findUserByPhone(phone)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("utilisateur non trouvé")
	}
	return user, nil
}

// GetUserByID récupère un utilisateur par son ID
func (s *UserService) GetUserByID(userID string) (*models.User, error) {
	ctx := context.Background()

	user, err := db.PrismaDB.User.FindUnique(
		prismadb.User.ID.Equals(userID),
	).Exec(ctx)

	if err != nil {
		if err == prismadb.ErrNotFound {
			return nil, fmt.Errorf("utilisateur non trouvé")
		}
		return nil, err
	}

	return s.convertPrismaUserToModel(user), nil
}

// UpdateUser met à jour un utilisateur
func (s *UserService) UpdateUser(user *models.User) error {
	ctx := context.Background()

	updateData := []prismadb.UserSetParam{
		prismadb.User.FirstName.Set(user.FirstName),
		prismadb.User.LastName.Set(user.LastName),
	}

	if user.Email != nil {
		updateData = append(updateData, prismadb.User.Email.Set(*user.Email))
	}

	_, err := db.PrismaDB.User.FindUnique(
		prismadb.User.ID.Equals(user.ID),
	).Update(updateData...).Exec(ctx)

	return err
}

// GetAllDrivers récupère tous les livreurs avec pagination et filtre de statut
func (s *UserService) GetAllDrivers(page int, limit int, statusFilter string) ([]*models.User, int, error) {
	ctx := context.Background()

	// Calculer offset
	offset := (page - 1) * limit

	// Construire les conditions de filtre pour les livreurs seulement
	conditions := []prismadb.UserWhereParam{
		prismadb.User.Role.Equals(prismadb.UserRoleLivreur),
	}

	// Ajouter le filtre de statut si spécifié
	if statusFilter != "" {
		switch statusFilter {
		case "ONLINE":
			conditions = append(conditions, prismadb.User.DriverStatus.Equals(prismadb.DriverStatusOnline))
		case "OFFLINE":
			conditions = append(conditions, prismadb.User.DriverStatus.Equals(prismadb.DriverStatusOffline))
		case "BUSY":
			conditions = append(conditions, prismadb.User.DriverStatus.Equals(prismadb.DriverStatusBusy))
		case "AVAILABLE":
			conditions = append(conditions, prismadb.User.DriverStatus.Equals(prismadb.DriverStatusAvailable))
		}
	}

	// Récupérer les livreurs avec pagination
	drivers, err := db.PrismaDB.User.FindMany(
		conditions...,
	).Skip(offset).Take(limit).OrderBy(
		prismadb.User.CreatedAt.Order(prismadb.SortOrderDesc),
	).Exec(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Compter le total avec les mêmes conditions
	allDrivers, err := db.PrismaDB.User.FindMany(
		conditions...,
	).Exec(ctx)
	if err != nil {
		return nil, 0, err
	}
	total := len(allDrivers)

	// Convertir les utilisateurs
	driverModels := make([]*models.User, len(drivers))
	for i, driver := range drivers {
		driverModels[i] = s.convertPrismaUserToModel(&driver)
	}

	return driverModels, total, nil
}

// GetDriverStats récupère les statistiques d'un livreur
func (s *UserService) GetDriverStats(driverID string) (map[string]interface{}, error) {
	ctx := context.Background()

	// Vérifier si l'utilisateur existe et est un livreur
	driver, err := db.PrismaDB.User.FindUnique(
		prismadb.User.ID.Equals(driverID),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}

	if driver.Role != prismadb.UserRoleLivreur {
		return nil, fmt.Errorf("l'utilisateur n'est pas un livreur")
	}

	// Compter les livraisons du livreur
	deliveries, err := db.PrismaDB.Delivery.FindMany(
		prismadb.Delivery.DriverID.Equals(driverID),
	).Exec(ctx)
	if err != nil {
		deliveries = []prismadb.DeliveryModel{} // En cas d'erreur, retourner 0
	}

	// Compter les livraisons par statut
	var completedCount, cancelledCount, activeCount int
	for _, delivery := range deliveries {
		switch delivery.Status {
		case prismadb.DeliveryStatusDelivered:
			completedCount++
		case prismadb.DeliveryStatusCancelled, prismadb.DeliveryStatusFailed:
			cancelledCount++
		case prismadb.DeliveryStatusAssigned, prismadb.DeliveryStatusPickedUp, prismadb.DeliveryStatusInTransit:
			activeCount++
		}
	}

	// Compter les évaluations moyennes
	ratings, err := db.PrismaDB.Rating.FindMany(
		prismadb.Rating.UserID.Equals(driverID),
	).Exec(ctx)
	if err != nil {
		ratings = []prismadb.RatingModel{} // En cas d'erreur, retourner 0
	}

	averageRating := 0.0
	if len(ratings) > 0 {
		var totalRating int
		for _, rating := range ratings {
			totalRating += rating.Rating
		}
		averageRating = float64(totalRating) / float64(len(ratings))
	}

	stats := map[string]interface{}{
		"totalDeliveries":    len(deliveries),
		"completedDeliveries": completedCount,
		"cancelledDeliveries": cancelledCount,
		"activeDeliveries":   activeCount,
		"averageRating":      averageRating,
		"ratingsCount":       len(ratings),
		"driverStatus":       string(driver.DriverStatus),
		"isActive":           driver.DriverStatus == prismadb.DriverStatusOnline || driver.DriverStatus == prismadb.DriverStatusAvailable,
	}

	return stats, nil
}

// UpdateDriverStatus met à jour le statut d'un livreur
func (s *UserService) UpdateDriverStatus(driverID string, status models.DriverStatus) error {
	ctx := context.Background()

	// Convertir le statut vers le type Prisma
	var prismaStatus prismadb.DriverStatus
	switch status {
	case models.DriverStatusOffline:
		prismaStatus = prismadb.DriverStatusOffline
	case models.DriverStatusOnline:
		prismaStatus = prismadb.DriverStatusOnline
	case models.DriverStatusBusy:
		prismaStatus = prismadb.DriverStatusBusy
	case models.DriverStatusAvailable:
		prismaStatus = prismadb.DriverStatusAvailable
	default:
		return fmt.Errorf("statut de livreur invalide: %s", status)
	}

	// Vérifier que l'utilisateur est un livreur
	user, err := db.PrismaDB.User.FindUnique(
		prismadb.User.ID.Equals(driverID),
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("utilisateur non trouvé: %v", err)
	}

	if user.Role != prismadb.UserRoleLivreur {
		return fmt.Errorf("l'utilisateur n'est pas un livreur")
	}

	// Mettre à jour le statut
	_, err = db.PrismaDB.User.FindUnique(
		prismadb.User.ID.Equals(driverID),
	).Update(
		prismadb.User.DriverStatus.Set(prismaStatus),
	).Exec(ctx)

	return err
}

// DeleteUser supprime un utilisateur
func (s *UserService) DeleteUser(userID string) error {
	ctx := context.Background()

	_, err := db.PrismaDB.User.FindUnique(
		prismadb.User.ID.Equals(userID),
	).Delete().Exec(ctx)

	return err
}

// GetAllUsers récupère tous les utilisateurs avec pagination et filtre
func (s *UserService) GetAllUsers(page int, limit int, roleFilter string) ([]*models.User, int, error) {
	ctx := context.Background()

	// Calculer offset
	offset := (page - 1) * limit

	// Construire les conditions de filtre
	var conditions []prismadb.UserWhereParam

	if roleFilter != "" {
		switch roleFilter {
		case "CLIENT":
			conditions = append(conditions, prismadb.User.Role.Equals(prismadb.UserRoleClient))
		case "LIVREUR":
			conditions = append(conditions, prismadb.User.Role.Equals(prismadb.UserRoleLivreur))
		case "ADMIN":
			conditions = append(conditions, prismadb.User.Role.Equals(prismadb.UserRoleAdmin))
		case "GESTIONNAIRE":
			conditions = append(conditions, prismadb.User.Role.Equals(prismadb.UserRoleGestionnaire))
		}
	}

	// Récupérer les utilisateurs avec pagination
	users, err := db.PrismaDB.User.FindMany(
		conditions...,
	).Skip(offset).Take(limit).OrderBy(
		prismadb.User.CreatedAt.Order(prismadb.SortOrderDesc),
	).Exec(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Compter le total avec les mêmes conditions
	allUsers, err := db.PrismaDB.User.FindMany(
		conditions...,
	).Exec(ctx)
	if err != nil {
		return nil, 0, err
	}
	total := len(allUsers)

	// Convertir les utilisateurs
	userModels := make([]*models.User, len(users))
	for i, user := range users {
		userModels[i] = s.convertPrismaUserToModel(&user)
	}

	return userModels, total, nil
}

// GetUserStats récupère les statistiques d'un utilisateur
func (s *UserService) GetUserStats(userID string) (map[string]interface{}, error) {
	ctx := context.Background()

	// Vérifier si l'utilisateur existe
	_, err := db.PrismaDB.User.FindUnique(
		prismadb.User.ID.Equals(userID),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}

	// Compter les livraisons de l'utilisateur
	deliveries, err := db.PrismaDB.Delivery.FindMany(
		prismadb.Delivery.UserID.Equals(prismadb.String(userID)),
	).Exec(ctx)
	if err != nil {
		deliveries = []prismadb.DeliveryModel{} // En cas d'erreur, retourner 0
	}

	// Compter les évaluations moyennes (si applicable)
	ratings, err := db.PrismaDB.Rating.FindMany(
		prismadb.Rating.UserID.Equals(userID),
	).Exec(ctx)
	if err != nil {
		ratings = []prismadb.RatingModel{} // En cas d'erreur, retourner 0
	}

	averageRating := 0.0
	if len(ratings) > 0 {
		var totalRating int
		for _, rating := range ratings {
			totalRating += rating.Rating
		}
		averageRating = float64(totalRating) / float64(len(ratings))
	}

	stats := map[string]interface{}{
		"deliveriesCount": len(deliveries),
		"vehiclesCount":   0, // TODO: Implémenter si nécessaire
		"averageRating":   averageRating,
		"ratingsCount":    len(ratings),
	}

	return stats, nil
}

// UpdateUserRole met à jour le rôle d'un utilisateur
func (s *UserService) UpdateUserRole(userID string, role models.UserRole) error {
	ctx := context.Background()

	// Convertir le rôle vers le type Prisma
	var prismaRole prismadb.UserRole
	switch role {
	case models.UserRoleClient:
		prismaRole = prismadb.UserRoleClient
	case models.UserRoleLivreur:
		prismaRole = prismadb.UserRoleLivreur
	case models.UserRoleAdmin:
		prismaRole = prismadb.UserRoleAdmin
	case models.UserRoleGestionnaire:
		prismaRole = prismadb.UserRoleGestionnaire
	default:
		return fmt.Errorf("rôle invalide: %s", role)
	}

	// Mettre à jour le rôle
	_, err := db.PrismaDB.User.FindUnique(
		prismadb.User.ID.Equals(userID),
	).Update(
		prismadb.User.Role.Set(prismaRole),
	).Exec(ctx)

	return err
}
