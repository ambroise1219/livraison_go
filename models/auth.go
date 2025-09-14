package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// OTP represents an OTP record
type OTP struct {
	ID        string    `json:"id"`
	Phone     string    `json:"phone" validate:"required,min=8,max=15"`
	Code      string    `json:"code" validate:"required,len=6"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// SendOTPRequest represents request for sending OTP
type SendOTPRequest struct {
	Phone string `json:"phone" validate:"required,min=8,max=15"`
}

// VerifyOTPRequest represents request for verifying OTP
type VerifyOTPRequest struct {
	Phone string `json:"phone" validate:"required,min=8,max=15"`
	Code  string `json:"code" validate:"required,len=6"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Phone string `json:"phone" validate:"required,min=8,max=15"`
	Code  string `json:"code" validate:"required,len=6"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token        string       `json:"token"`
	RefreshToken string       `json:"refreshToken"`
	User         *UserResponse `json:"user"`
	ExpiresAt    time.Time    `json:"expiresAt"`
}

// RefreshToken represents a refresh token record
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"createdAt"`
}

// RefreshTokenRequest represents request for refreshing token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// IsExpired checks if OTP is expired
func (o *OTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

// IsValid checks if OTP is valid for given phone and code
func (o *OTP) IsValid(phone, code string) bool {
	return o.Phone == phone && o.Code == code && !o.IsExpired()
}

// IsExpired checks if refresh token is expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid checks if refresh token is valid
func (rt *RefreshToken) IsValid() bool {
	return !rt.Revoked && !rt.IsExpired()
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID string   `json:"user_id"`
	Phone  string   `json:"phone"`
	Role   UserRole `json:"role"`
}


// PasswordResetRequest represents password reset request
type PasswordResetRequest struct {
	Phone string `json:"phone" validate:"required,min=8,max=15"`
}

// ChangePhoneRequest represents phone change request
type ChangePhoneRequest struct {
	NewPhone    string `json:"newPhone" validate:"required,min=8,max=15"`
	OTPCode     string `json:"otpCode" validate:"required,len=6"`
	CurrentPhone string `json:"currentPhone" validate:"required,min=8,max=15"`
}
