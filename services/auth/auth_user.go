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

// DeleteUser supprime un utilisateur
func (s *UserService) DeleteUser(userID string) error {
	ctx := context.Background()

	_, err := db.PrismaDB.User.FindUnique(
		prismadb.User.ID.Equals(userID),
	).Delete().Exec(ctx)

	return err
}
