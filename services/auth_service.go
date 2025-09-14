package services

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"ilex-backend/config"
	"ilex-backend/db"
	"ilex-backend/models"
)

type AuthService struct {
	config *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		config: cfg,
	}
}

// SendOTP generates and sends OTP to phone number
func (s *AuthService) SendOTP(phone string) (*models.OTP, error) {
	// Delete any existing OTPs for this phone
	if err := s.deleteExistingOTPs(phone); err != nil {
		log.Printf("Warning: failed to delete existing OTPs for %s: %v", phone, err)
	}

	// Generate OTP code
	code, err := s.generateOTPCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %v", err)
	}

	// Create OTP record
	otp := &models.OTP{
		ID:        uuid.New().String(),
		Phone:     phone,
		Code:      code,
		ExpiresAt: time.Now().Add(time.Duration(s.config.OTPExpiration) * time.Minute),
		CreatedAt: time.Now(),
	}

	// Save to database
	query := `CREATE OTP SET 
		id = $id, 
		phone = $phone, 
		code = $code, 
		expiresAt = $expiresAt, 
		createdAt = $createdAt`
	
	params := map[string]interface{}{
		"id":        otp.ID,
		"phone":     otp.Phone,
		"code":      otp.Code,
		"expiresAt": otp.ExpiresAt,
		"createdAt": otp.CreatedAt,
	}

	_, err = db.Query(query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to save OTP: %v", err)
	}

	// TODO: Send SMS notification
	log.Printf("OTP %s generated for phone %s (expires at %v)", code, phone, otp.ExpiresAt)

	// Return OTP without code for security (only for development)
	if s.config.Environment == "development" {
		return otp, nil
	}

	return &models.OTP{
		ID:        otp.ID,
		Phone:     otp.Phone,
		ExpiresAt: otp.ExpiresAt,
		CreatedAt: otp.CreatedAt,
	}, nil
}

// VerifyOTP verifies OTP code and creates/returns user
func (s *AuthService) VerifyOTP(phone, code string) (*models.AuthResponse, error) {
	// Find valid OTP
	otp, err := s.findValidOTP(phone, code)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired OTP: %v", err)
	}

	// Find or create user
	user, isNewUser, err := s.findOrCreateUser(phone)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %v", err)
	}

	// Delete used OTP
	if err := s.deleteOTP(otp.ID); err != nil {
		log.Printf("Warning: failed to delete used OTP: %v", err)
	}

	// Generate tokens
	token, refreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %v", err)
	}

	// For new users, trigger welcome notifications
	if isNewUser {
		go s.sendWelcomeNotifications(user)
	}

	return &models.AuthResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
		ExpiresAt:    expiresAt,
	}, nil
}

// RefreshToken refreshes JWT token using refresh token
func (s *AuthService) RefreshToken(refreshTokenString string) (*models.AuthResponse, error) {
	// Find refresh token
	refreshToken, err := s.findRefreshToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %v", err)
	}

	if !refreshToken.IsValid() {
		// Revoke invalid token
		s.revokeRefreshToken(refreshToken.ID)
		return nil, fmt.Errorf("refresh token expired or revoked")
	}

	// Get user
	user, err := s.getUserByID(refreshToken.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	// Generate new tokens
	token, newRefreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %v", err)
	}

	// Revoke old refresh token
	s.revokeRefreshToken(refreshToken.ID)

	return &models.AuthResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
		User:         user.ToResponse(),
		ExpiresAt:    expiresAt,
	}, nil
}

// ValidateToken validates JWT token and returns claims
func (s *AuthService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Extract claims
	userID, _ := claims["user_id"].(string)
	phone, _ := claims["phone"].(string)
	roleStr, _ := claims["role"].(string)
	exp, _ := claims["exp"].(float64)
	iat, _ := claims["iat"].(float64)

	return &models.JWTClaims{
		UserID: userID,
		Phone:  phone,
		Role:   models.UserRole(roleStr),
		Exp:    int64(exp),
		Iat:    int64(iat),
	}, nil
}

// RevokeRefreshToken revokes a refresh token
func (s *AuthService) RevokeRefreshToken(tokenString string) error {
	refreshToken, err := s.findRefreshToken(tokenString)
	if err != nil {
		return fmt.Errorf("refresh token not found: %v", err)
	}

	return s.revokeRefreshToken(refreshToken.ID)
}

// Helper methods

func (s *AuthService) generateOTPCode() (string, error) {
	const digits = "0123456789"
	code := make([]byte, s.config.OTPLength)

	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}

	return string(code), nil
}

func (s *AuthService) deleteExistingOTPs(phone string) error {
	query := "DELETE FROM OTP WHERE phone = $phone"
	params := map[string]interface{}{"phone": phone}
	
	_, err := db.Query(query, params)
	return err
}

func (s *AuthService) findValidOTP(phone, code string) (*models.OTP, error) {
	query := `SELECT * FROM OTP WHERE phone = $phone AND code = $code AND expiresAt > $now LIMIT 1`
	params := map[string]interface{}{
		"phone": phone,
		"code":  code,
		"now":   time.Now(),
	}

	result, err := db.QuerySingle(query, params)
	if err != nil {
		return nil, err
	}

	otpData, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid OTP data format")
	}

	return s.parseOTPFromMap(otpData), nil
}

func (s *AuthService) findOrCreateUser(phone string) (*models.User, bool, error) {
	// Try to find existing user
	query := `SELECT * FROM User WHERE phone = $phone LIMIT 1`
	params := map[string]interface{}{"phone": phone}

	result, err := db.QuerySingle(query, params)
	if err == nil {
		// User exists
		userData, ok := result.(map[string]interface{})
		if ok {
			user := s.parseUserFromMap(userData)
			return user, false, nil
		}
	}

	// Create new user
	user := &models.User{
		ID:           uuid.New().String(),
		Phone:        phone,
		Role:         models.UserRoleClient,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		DriverStatus: models.DriverStatusOffline,
		FirstName:    "",
		LastName:     "",
	}

	createQuery := `CREATE User SET 
		id = $id,
		phone = $phone,
		role = $role,
		createdAt = $createdAt,
		updatedAt = $updatedAt,
		driverStatus = $driverStatus,
		firstName = $firstName,
		lastName = $lastName,
		is_profile_completed = false,
		is_driver_complete = false,
		is_driver_vehicule_complete = false`
	
	createParams := map[string]interface{}{
		"id":           user.ID,
		"phone":        user.Phone,
		"role":         string(user.Role),
		"createdAt":    user.CreatedAt,
		"updatedAt":    user.UpdatedAt,
		"driverStatus": string(user.DriverStatus),
		"firstName":    user.FirstName,
		"lastName":     user.LastName,
	}

	_, err = db.Query(createQuery, createParams)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create user: %v", err)
	}

	return user, true, nil
}

func (s *AuthService) generateTokens(user *models.User) (string, string, time.Time, error) {
	// Generate JWT token
	expiresAt := time.Now().Add(time.Duration(s.config.JWTExpiration) * time.Hour)
	
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"phone":   user.Phone,
		"role":    string(user.Role),
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to sign JWT token: %v", err)
	}

	// Generate refresh token
	refreshToken := &models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     s.generateRandomString(64),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
		CreatedAt: time.Now(),
		Revoked:   false,
	}

	// Save refresh token
	query := `CREATE RefreshToken SET 
		id = $id,
		userId = $userId,
		token = $token,
		expiresAt = $expiresAt,
		revoked = $revoked,
		createdAt = $createdAt`
	
	params := map[string]interface{}{
		"id":        refreshToken.ID,
		"userId":    refreshToken.UserID,
		"token":     refreshToken.Token,
		"expiresAt": refreshToken.ExpiresAt,
		"revoked":   refreshToken.Revoked,
		"createdAt": refreshToken.CreatedAt,
	}

	_, err = db.Query(query, params)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to save refresh token: %v", err)
	}

	return tokenString, refreshToken.Token, expiresAt, nil
}

func (s *AuthService) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}
	
	return string(result)
}

func (s *AuthService) findRefreshToken(token string) (*models.RefreshToken, error) {
	query := `SELECT * FROM RefreshToken WHERE token = $token LIMIT 1`
	params := map[string]interface{}{"token": token}

	result, err := db.QuerySingle(query, params)
	if err != nil {
		return nil, err
	}

	tokenData, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid refresh token data format")
	}

	return s.parseRefreshTokenFromMap(tokenData), nil
}

func (s *AuthService) revokeRefreshToken(tokenID string) error {
	query := `UPDATE RefreshToken SET revoked = true WHERE id = $id`
	params := map[string]interface{}{"id": tokenID}
	
	_, err := db.Query(query, params)
	return err
}

func (s *AuthService) deleteOTP(otpID string) error {
	query := `DELETE FROM OTP WHERE id = $id`
	params := map[string]interface{}{"id": otpID}
	
	_, err := db.Query(query, params)
	return err
}

func (s *AuthService) getUserByID(userID string) (*models.User, error) {
	query := `SELECT * FROM User WHERE id = $id LIMIT 1`
	params := map[string]interface{}{"id": userID}

	result, err := db.QuerySingle(query, params)
	if err != nil {
		return nil, err
	}

	userData, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid user data format")
	}

	return s.parseUserFromMap(userData), nil
}

func (s *AuthService) sendWelcomeNotifications(user *models.User) {
	// TODO: Implement welcome notifications
	log.Printf("Sending welcome notifications to user %s", user.ID)
}

// Parser helper methods
func (s *AuthService) parseOTPFromMap(data map[string]interface{}) *models.OTP {
	otp := &models.OTP{}
	
	if id, ok := data["id"].(string); ok {
		otp.ID = id
	}
	if phone, ok := data["phone"].(string); ok {
		otp.Phone = phone
	}
	if code, ok := data["code"].(string); ok {
		otp.Code = code
	}
	if expiresAt, ok := data["expiresAt"].(time.Time); ok {
		otp.ExpiresAt = expiresAt
	}
	if createdAt, ok := data["createdAt"].(time.Time); ok {
		otp.CreatedAt = createdAt
	}
	
	return otp
}

func (s *AuthService) parseUserFromMap(data map[string]interface{}) *models.User {
	user := &models.User{}
	
	if id, ok := data["id"].(string); ok {
		user.ID = id
	}
	if phone, ok := data["phone"].(string); ok {
		user.Phone = phone
	}
	if role, ok := data["role"].(string); ok {
		user.Role = models.UserRole(role)
	}
	if firstName, ok := data["firstName"].(string); ok {
		user.FirstName = firstName
	}
	if lastName, ok := data["lastName"].(string); ok {
		user.LastName = lastName
	}
	if email, ok := data["email"].(string); ok && email != "" {
		user.Email = &email
	}
	if createdAt, ok := data["createdAt"].(time.Time); ok {
		user.CreatedAt = createdAt
	}
	if updatedAt, ok := data["updatedAt"].(time.Time); ok {
		user.UpdatedAt = updatedAt
	}
	if driverStatus, ok := data["driverStatus"].(string); ok {
		user.DriverStatus = models.DriverStatus(driverStatus)
	}
	
	return user
}

func (s *AuthService) parseRefreshTokenFromMap(data map[string]interface{}) *models.RefreshToken {
	rt := &models.RefreshToken{}
	
	if id, ok := data["id"].(string); ok {
		rt.ID = id
	}
	if userID, ok := data["userId"].(string); ok {
		rt.UserID = userID
	}
	if token, ok := data["token"].(string); ok {
		rt.Token = token
	}
	if expiresAt, ok := data["expiresAt"].(time.Time); ok {
		rt.ExpiresAt = expiresAt
	}
	if revoked, ok := data["revoked"].(bool); ok {
		rt.Revoked = revoked
	}
	if createdAt, ok := data["createdAt"].(time.Time); ok {
		rt.CreatedAt = createdAt
	}
	
	return rt
}