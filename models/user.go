package models

import (
	"time"
)

// UserRole defines the role enumeration
type UserRole string

const (
	UserRoleClient      UserRole = "CLIENT"
	UserRoleLivreur     UserRole = "LIVREUR"
	UserRoleAdmin       UserRole = "ADMIN"
	UserRoleGestionnaire UserRole = "GESTIONNAIRE"
	UserRoleMarketing   UserRole = "MARKETING"
)

// DriverStatus defines the driver status enumeration
type DriverStatus string

const (
	DriverStatusOffline   DriverStatus = "OFFLINE"
	DriverStatusOnline    DriverStatus = "ONLINE"
	DriverStatusBusy      DriverStatus = "BUSY"
	DriverStatusAvailable DriverStatus = "AVAILABLE"
)

// User represents a user in the system
type User struct {
	ID                         string     `json:"id" validate:"required"`
	Phone                      string     `json:"phone" validate:"required,min=8,max=15"`
	Address                    *string    `json:"address,omitempty"`
	Role                       UserRole   `json:"role" validate:"required"`
	ReferredByID              *string    `json:"referredById,omitempty"`
	CreatedAt                 time.Time  `json:"createdAt"`
	UpdatedAt                 time.Time  `json:"updatedAt"`
	ProfilePictureID          *string    `json:"profilePictureId,omitempty"`
	LastName                  string     `json:"lastName"`
	FirstName                 string     `json:"firstName"`
	Email                     *string    `json:"email,omitempty" validate:"omitempty,email"`
	DateOfBirth               *time.Time `json:"dateOfBirth,omitempty"`
	LieuResidence             *string    `json:"lieuResidence,omitempty"`
	IsProfileCompleted        bool       `json:"is_profile_completed"`
	IsDriverComplete          bool       `json:"is_driver_complete"`
	IsDriverVehiculeComplete  bool       `json:"is_driver_vehicule_complete"`
	CNIRecto                  *string    `json:"cni_recto,omitempty"`
	CNIVerso                  *string    `json:"cni_verso,omitempty"`
	PermisRecto               *string    `json:"permis_recto,omitempty"`
	PermisVerso               *string    `json:"permis_verso,omitempty"`
	DriverStatus              DriverStatus `json:"driverStatus"`
	LastKnownLat              *float64   `json:"lastKnownLat,omitempty" validate:"omitempty,gte=-90,lte=90"`
	LastKnownLng              *float64   `json:"lastKnownLng,omitempty" validate:"omitempty,gte=-180,lte=180"`
	LastSeenAt                *time.Time `json:"lastSeenAt,omitempty"`
}

// CreateUserRequest represents request for creating a user
type CreateUserRequest struct {
	Phone         string    `json:"phone" validate:"required,min=8,max=15"`
	LastName      string    `json:"lastName" validate:"required,min=2,max=50"`
	FirstName     string    `json:"firstName" validate:"required,min=2,max=50"`
	Email         *string   `json:"email,omitempty" validate:"omitempty,email"`
	DateOfBirth   *time.Time `json:"dateOfBirth,omitempty"`
	Address       *string   `json:"address,omitempty"`
	LieuResidence *string   `json:"lieuResidence,omitempty"`
	Role          UserRole  `json:"role" validate:"required"`
	ReferredByID  *string   `json:"referredById,omitempty"`
}

// UpdateUserRequest represents request for updating a user
type UpdateUserRequest struct {
	LastName      *string    `json:"lastName,omitempty" validate:"omitempty,min=2,max=50"`
	FirstName     *string    `json:"firstName,omitempty" validate:"omitempty,min=2,max=50"`
	Email         *string    `json:"email,omitempty" validate:"omitempty,email"`
	DateOfBirth   *time.Time `json:"dateOfBirth,omitempty"`
	Address       *string    `json:"address,omitempty"`
	LieuResidence *string    `json:"lieuResidence,omitempty"`
}

// UpdateDriverLocationRequest represents request for updating driver location
type UpdateDriverLocationRequest struct {
	Lat    *float64      `json:"lat" validate:"omitempty,gte=-90,lte=90"`
	Lng    *float64      `json:"lng" validate:"omitempty,gte=-180,lte=180"`
	Status *DriverStatus `json:"status,omitempty"`
}

// UserResponse represents user data in response
type UserResponse struct {
	ID                       string        `json:"id"`
	Phone                    string        `json:"phone"`
	Address                  *string       `json:"address,omitempty"`
	Role                     UserRole      `json:"role"`
	ReferredByID            *string       `json:"referredById,omitempty"`
	CreatedAt               time.Time     `json:"createdAt"`
	UpdatedAt               time.Time     `json:"updatedAt"`
	ProfilePictureID        *string       `json:"profilePictureId,omitempty"`
	LastName                string        `json:"lastName"`
	FirstName               string        `json:"firstName"`
	Email                   *string       `json:"email,omitempty"`
	DateOfBirth             *time.Time    `json:"dateOfBirth,omitempty"`
	LieuResidence           *string       `json:"lieuResidence,omitempty"`
	IsProfileCompleted      bool          `json:"is_profile_completed"`
	IsDriverComplete        bool          `json:"is_driver_complete"`
	IsDriverVehiculeComplete bool         `json:"is_driver_vehicule_complete"`
	DriverStatus            *DriverStatus `json:"driverStatus,omitempty"`
	LastKnownLat            *float64      `json:"lastKnownLat,omitempty"`
	LastKnownLng            *float64      `json:"lastKnownLng,omitempty"`
	LastSeenAt              *time.Time    `json:"lastSeenAt,omitempty"`
}

// IsValidRole checks if the role is valid
func (r UserRole) IsValid() bool {
	return r == UserRoleClient || r == UserRoleLivreur || r == UserRoleAdmin || 
		   r == UserRoleGestionnaire || r == UserRoleMarketing
}

// IsValidDriverStatus checks if the driver status is valid
func (s DriverStatus) IsValid() bool {
	return s == DriverStatusOffline || s == DriverStatusOnline || 
		   s == DriverStatusBusy || s == DriverStatusAvailable
}

// CanAcceptDeliveries checks if user can accept deliveries
func (u *User) CanAcceptDeliveries() bool {
	return u.Role == UserRoleLivreur && 
		   u.IsDriverComplete && 
		   u.IsDriverVehiculeComplete &&
		   (u.DriverStatus == DriverStatusOnline || u.DriverStatus == DriverStatusAvailable)
}

// IsDriver checks if user is a driver
func (u *User) IsDriver() bool {
	return u.Role == UserRoleLivreur
}

// IsClient checks if user is a client
func (u *User) IsClient() bool {
	return u.Role == UserRoleClient
}

// IsAdmin checks if user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin || u.Role == UserRoleGestionnaire
}

// GetFullName returns the full name of the user
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	resp := &UserResponse{
		ID:                       u.ID,
		Phone:                    u.Phone,
		Address:                  u.Address,
		Role:                     u.Role,
		ReferredByID:            u.ReferredByID,
		CreatedAt:               u.CreatedAt,
		UpdatedAt:               u.UpdatedAt,
		ProfilePictureID:        u.ProfilePictureID,
		LastName:                u.LastName,
		FirstName:               u.FirstName,
		Email:                   u.Email,
		DateOfBirth:             u.DateOfBirth,
		LieuResidence:           u.LieuResidence,
		IsProfileCompleted:      u.IsProfileCompleted,
		IsDriverComplete:        u.IsDriverComplete,
		IsDriverVehiculeComplete: u.IsDriverVehiculeComplete,
	}

	// Only include driver-specific fields for drivers
	if u.IsDriver() {
		resp.DriverStatus = &u.DriverStatus
		resp.LastKnownLat = u.LastKnownLat
		resp.LastKnownLng = u.LastKnownLng
		resp.LastSeenAt = u.LastSeenAt
	}

	return resp
}
