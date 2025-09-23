package models

// ErrorResponse représente une réponse d'erreur standard
type ErrorResponse struct {
	Error   string `json:"error" example:"Message d'erreur"`
	Details string `json:"details,omitempty" example:"Détails supplémentaires"`
}

// OTPResponse représente la réponse d'envoi d'OTP
type OTPResponse struct {
	Message   string `json:"message" example:"OTP sent successfully"`
	ExpiresIn string `json:"expiresIn" example:"5 minutes"`
}

// AuthResponse représente la réponse d'authentification
type AuthResponse struct {
	Message      string `json:"message" example:"Authentication successful"`
	AccessToken  string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn    int64  `json:"expiresIn" example:"3600"`
	User         User   `json:"user"`
}

// UserProfileResponse représente la réponse du profil utilisateur
type UserProfileResponse struct {
	ID                string  `json:"id" example:"user_123"`
	Phone             string  `json:"phone" example:"+2250173226070"`
	FirstName         string  `json:"firstName" example:"John"`
	LastName          string  `json:"lastName" example:"Doe"`
	Email             *string `json:"email,omitempty" example:"john.doe@example.com"`
	ProfilePictureUrl *string `json:"profilePictureUrl,omitempty" example:"https://res.cloudinary.com/livraison/image/upload/v1234567890/photo_profil_livraison/abc123.jpg"`
	IsDriver          bool    `json:"isDriver" example:"false"`
	IsAdmin           bool    `json:"isAdmin" example:"false"`
	CreatedAt         string  `json:"createdAt" example:"2024-01-15T10:30:00Z"`
	UpdatedAt         string  `json:"updatedAt" example:"2024-01-15T10:30:00Z"`
}

// ProfilePictureResponse représente la réponse d'upload de photo de profil
type ProfilePictureResponse struct {
	Message           string  `json:"message" example:"Photo de profil uploadée avec succès"`
	PublicID          string  `json:"publicId" example:"photo_profil_livraison/abc123"`
	URL               string  `json:"url" example:"https://res.cloudinary.com/livraison/image/upload/v1234567890/photo_profil_livraison/abc123.jpg"`
	ProfilePictureUrl string  `json:"profilePictureUrl" example:"https://res.cloudinary.com/livraison/image/upload/v1234567890/photo_profil_livraison/abc123.jpg"`
	User              User    `json:"user"`
}

// DeliveryResponse représente une réponse de livraison
type DeliveryResponse struct {
	ID          string  `json:"id" example:"delivery_123"`
	ClientID    string  `json:"clientId" example:"user_123"`
	DriverID    *string `json:"driverId,omitempty" example:"driver_456"`
	Status      string  `json:"status" example:"PENDING"`
	Type        string  `json:"type" example:"EXPRESS"`
	Description string  `json:"description" example:"Livraison urgente"`
	CreatedAt   string  `json:"createdAt" example:"2024-01-15T10:30:00Z"`
	UpdatedAt   string  `json:"updatedAt" example:"2024-01-15T10:30:00Z"`
}

// SuccessResponse représente une réponse de succès générique
type SuccessResponse struct {
	Message string      `json:"message" example:"Opération réussie"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse représente une réponse paginée
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page" example:"1"`
	Limit      int         `json:"limit" example:"10"`
	Total      int64       `json:"total" example:"100"`
	TotalPages int         `json:"totalPages" example:"10"`
}
