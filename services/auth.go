package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
)

type AuthService struct {
	config *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		config: cfg,
	}
}

// GenerateOTP generates a random OTP code
func (s *AuthService) GenerateOTP() string {
	max := int64(999999) // 6 digits max
	min := int64(100000) // 6 digits min
	
	n, _ := rand.Int(rand.Reader, big.NewInt(max-min+1))
	return fmt.Sprintf("%06d", n.Int64()+min)
}

// SaveOTP saves OTP to database with expiration
func (s *AuthService) SaveOTP(phone string) (*models.OTP, error) {
	code := s.GenerateOTP()
	now := time.Now()
	expiresAt := now.Add(time.Duration(s.config.OTPExpiration) * time.Minute)
	
	// Delete any existing OTP for this phone
	query := "DELETE OTP WHERE phone = $phone"
	_, err := db.DB.Query(query, map[string]interface{}{
		"phone": phone,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing OTP: %w", err)
	}
	
	// Create new OTP using SurrealDB datetime functions
	// SurrealDB utilise le format 5m pour 5 minutes, mais il faut utiliser la fonction duration
	query = fmt.Sprintf("CREATE OTP SET phone = $phone, code = $code, expiresAt = time::now() + %dm, createdAt = time::now()", s.config.OTPExpiration)
	result, err := db.DB.Query(query, map[string]interface{}{
		"phone": phone,
		"code": code,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save OTP: %w", err)
	}
	
	// Check if result contains an error
	if resultArray, ok := result.([]interface{}); ok && len(resultArray) > 0 {
		if firstResult, ok := resultArray[0].(map[string]interface{}); ok {
			if status, ok := firstResult["status"].(string); ok && status == "ERR" {
				return nil, fmt.Errorf("SurrealDB error: %v", firstResult["result"])
			}
		}
	}
	
	otp := &models.OTP{
		Phone:     phone,
		Code:      code,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}
	
	return otp, nil
}

// VerifyOTP verifies the OTP code against database
func (s *AuthService) VerifyOTP(phone, code string) (*models.OTP, error) {
	query := "SELECT * FROM OTP WHERE phone = $phone AND code = $code LIMIT 1"
	result, err := db.DB.Query(query, map[string]interface{}{
		"phone": phone,
		"code":  code,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query OTP: %w", err)
	}
	
	resultArray, ok := result.([]interface{})
	if !ok || len(resultArray) == 0 {
		return nil, fmt.Errorf("invalid OTP code")
	}
	
	// SurrealDB retourne un tableau d'objets
	firstResult, ok := resultArray[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}
	
	// Vérifie si le résultat est vide
	otpArray, ok := firstResult["result"].([]interface{})
	if !ok || len(otpArray) == 0 {
		return nil, fmt.Errorf("invalid OTP code")
	}
	
	// Parse the first result
	var otp models.OTP
	data, ok := otpArray[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected OTP data format")
	}
	
	// Parse l'ID qui peut être un format spécial dans SurrealDB
	switch id := data["id"].(type) {
	case string:
		otp.ID = id
	case map[string]interface{}:
		// Si l'ID est un objet complexe, on prend sa représentation en string
		otp.ID = fmt.Sprintf("%v", id)
	default:
		otp.ID = fmt.Sprintf("%v", data["id"])
	}
	
	// Parse les autres champs
	if phone, ok := data["phone"].(string); ok {
		otp.Phone = phone
	}
	if code, ok := data["code"].(string); ok {
		otp.Code = code
	}
	
	// Parse timestamps
	if expiresAt, ok := data["expiresAt"].(string); ok {
		otp.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAt)
	}
	if createdAt, ok := data["createdAt"].(string); ok {
		otp.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	}
	
	// Check if OTP is expired
	if otp.IsExpired() {
		return nil, fmt.Errorf("OTP has expired")
	}
	
	// Delete the used OTP
	query = "DELETE $id"
	_, err = db.DB.Query(query, map[string]interface{}{
		"id": otp.ID,
	})
	if err != nil {
		// Log error but don't fail the verification
		fmt.Printf("Warning: failed to delete used OTP: %v\n", err)
	}
	
	return &otp, nil
}

// FindOrCreateUser finds user by phone or creates new one
func (s *AuthService) FindOrCreateUser(phone string) (*models.User, bool, error) {
	// Try to find existing user
	query := "SELECT * FROM User WHERE phone = $phone LIMIT 1"
	result, err := db.DB.Query(query, map[string]interface{}{
		"phone": phone,
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to query user: %w", err)
	}
	
	// Parse les résultats de SurrealDB
	resultArray, ok := result.([]interface{})
	if ok && len(resultArray) > 0 {
		firstResult, ok := resultArray[0].(map[string]interface{})
		if ok {
			userArray, ok := firstResult["result"].([]interface{})
			if ok && len(userArray) > 0 {
				// User exists, parse and return
				data, ok := userArray[0].(map[string]interface{})
				if ok {
					user := s.parseUserFromDB(data)
					return user, false, nil
				}
			}
		}
	}
	
	// User doesn't exist, create new one
	now := time.Now()
	
	// Create user in database using SurrealDB datetime functions
	query = `CREATE User SET 
		phone = $phone, 
		role = $role, 
		createdAt = time::now(), 
		updatedAt = time::now(), 
		firstName = $firstName, 
		lastName = $lastName, 
		is_profile_completed = $profileCompleted, 
		is_driver_complete = $driverComplete, 
		is_driver_vehicule_complete = $vehiculeComplete, 
		driverStatus = $driverStatus`
	result, err = db.DB.Query(query, map[string]interface{}{
		"phone": phone,
		"role": string(models.UserRoleClient),
		"firstName": "",
		"lastName": "",
		"profileCompleted": false,
		"driverComplete": false,
		"vehiculeComplete": false,
		"driverStatus": string(models.DriverStatusOffline),
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to create user: %w", err)
	}
	
	// Check if result contains an error
	if resultArray, ok := result.([]interface{}); ok && len(resultArray) > 0 {
		if firstResult, ok := resultArray[0].(map[string]interface{}); ok {
			if status, ok := firstResult["status"].(string); ok && status == "ERR" {
				return nil, false, fmt.Errorf("SurrealDB error: %v", firstResult["result"])
			}
		}
	}
	
	// Create the User struct to return
	user := &models.User{
		Phone:                      phone,
		Role:                       models.UserRoleClient,
		CreatedAt:                 now,
		UpdatedAt:                 now,
		FirstName:                 "",
		LastName:                  "",
		IsProfileCompleted:        false,
		IsDriverComplete:          false,
		IsDriverVehiculeComplete:  false,
		DriverStatus:              models.DriverStatusOffline,
	}
	
	return user, true, nil
}

// parseUserFromDB converts database result to User model
func (s *AuthService) parseUserFromDB(data map[string]interface{}) *models.User {
	user := &models.User{
		Phone: "",
		Role:  models.UserRoleClient,
	}
	
	// Parse l'ID qui peut être un format spécial dans SurrealDB
	switch id := data["id"].(type) {
	case string:
		user.ID = id
	case map[string]interface{}:
		// Si l'ID est un objet complexe, on prend sa représentation en string
		user.ID = fmt.Sprintf("%v", id)
	default:
		user.ID = fmt.Sprintf("%v", data["id"])
	}
	
	// Parse les champs obligatoires avec vérification de type
	if phone, ok := data["phone"].(string); ok {
		user.Phone = phone
	}
	if role, ok := data["role"].(string); ok {
		user.Role = models.UserRole(role)
	}
	
	// Parse optional fields
	if address, ok := data["address"].(string); ok && address != "" {
		user.Address = &address
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
	if lieuRes, ok := data["lieuResidence"].(string); ok && lieuRes != "" {
		user.LieuResidence = &lieuRes
	}
	if profileCompleted, ok := data["is_profile_completed"].(bool); ok {
		user.IsProfileCompleted = profileCompleted
	}
	if driverComplete, ok := data["is_driver_complete"].(bool); ok {
		user.IsDriverComplete = driverComplete
	}
	if vehiculeComplete, ok := data["is_driver_vehicule_complete"].(bool); ok {
		user.IsDriverVehiculeComplete = vehiculeComplete
	}
	if driverStatus, ok := data["driverStatus"].(string); ok {
		user.DriverStatus = models.DriverStatus(driverStatus)
	}
	
	// Parse timestamps
	if createdAtStr, ok := data["createdAt"].(string); ok {
		user.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	}
	if updatedAtStr, ok := data["updatedAt"].(string); ok {
		user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
	}
	
	return user
}

// GenerateTokens generates JWT and refresh tokens
func (s *AuthService) GenerateTokens(user *models.User) (*models.AuthResponse, error) {
	// Generate JWT token
	now := time.Now()
	expiresAt := now.Add(time.Duration(s.config.JWTExpiration) * time.Hour)
	
	claims := &models.JWTClaims{
		UserID: user.ID,
		Phone:  user.Phone,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "ilex-backend",
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign JWT token: %w", err)
	}
	
	// Generate refresh token
	refreshTokenValue := uuid.New().String()
	
	// Save refresh token to database using SurrealDB datetime functions
	query := "CREATE RefreshToken SET userId = $userId, token = $token, expiresAt = time::now() + 7d, revoked = false, createdAt = time::now()"
	result, err := db.DB.Query(query, map[string]interface{}{
		"userId": user.ID,
		"token": refreshTokenValue,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}
	
	// Check if result contains an error
	if resultArray, ok := result.([]interface{}); ok && len(resultArray) > 0 {
		if firstResult, ok := resultArray[0].(map[string]interface{}); ok {
			if status, ok := firstResult["status"].(string); ok && status == "ERR" {
				return nil, fmt.Errorf("SurrealDB error: %v", firstResult["result"])
			}
		}
	}
	
	response := &models.AuthResponse{
		Token:        tokenString,
		RefreshToken: refreshTokenValue,
		User:         user.ToResponse(),
		ExpiresAt:    expiresAt,
	}
	
	return response, nil
}

// RefreshAccessToken generates new access token from refresh token
func (s *AuthService) RefreshAccessToken(refreshTokenStr string) (*models.AuthResponse, error) {
	// Find refresh token
	query := "SELECT * FROM RefreshToken WHERE token = $token AND revoked = false LIMIT 1"
	result, err := db.DB.Query(query, map[string]interface{}{
		"token": refreshTokenStr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query refresh token: %w", err)
	}
	
	// Parse les résultats de SurrealDB
	resultArray, ok := result.([]interface{})
	if !ok || len(resultArray) == 0 {
		return nil, fmt.Errorf("invalid refresh token")
	}
	
	// SurrealDB retourne un tableau d'objets
	firstResult, ok := resultArray[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}
	
	// Vérifie si le résultat est vide
	tokenArray, ok := firstResult["result"].([]interface{})
	if !ok || len(tokenArray) == 0 {
		return nil, fmt.Errorf("invalid refresh token")
	}
	
	// Parse refresh token
	data, ok := tokenArray[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected token data format")
	}
	
	refreshToken := &models.RefreshToken{}
	
	// Parse l'ID qui peut être un format spécial dans SurrealDB
	switch id := data["id"].(type) {
	case string:
		refreshToken.ID = id
	case map[string]interface{}:
		// Si l'ID est un objet complexe, on prend sa représentation en string
		refreshToken.ID = fmt.Sprintf("%v", id)
	default:
		refreshToken.ID = fmt.Sprintf("%v", data["id"])
	}
	
	// Parse les autres champs
	if userId, ok := data["userId"].(string); ok {
		refreshToken.UserID = userId
	}
	if token, ok := data["token"].(string); ok {
		refreshToken.Token = token
	}
	if revoked, ok := data["revoked"].(bool); ok {
		refreshToken.Revoked = revoked
	}
	
	if expiresAtStr, ok := data["expiresAt"].(string); ok {
		refreshToken.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAtStr)
	}
	
	// Check if refresh token is expired
	if refreshToken.IsExpired() {
		return nil, fmt.Errorf("refresh token has expired")
	}
	
	// Get user
	query = "SELECT * FROM User WHERE id = $userId LIMIT 1"
	result, err = db.DB.Query(query, map[string]interface{}{
		"userId": refreshToken.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	
	// Parse les résultats de SurrealDB
	resultArray, ok = result.([]interface{})
	if !ok || len(resultArray) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	
	// SurrealDB retourne un tableau d'objets
	firstResult, ok = resultArray[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}
	
	// Vérifie si le résultat est vide
	userArray, ok := firstResult["result"].([]interface{})
	if !ok || len(userArray) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	
	// Parse user data
	userData, ok := userArray[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected user data format")
	}
	
	user := s.parseUserFromDB(userData)
	
	// Generate new tokens
	return s.GenerateTokens(user)
}

// ValidateToken validates JWT token and returns claims
func (s *AuthService) ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Make sure the token method conforms to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// SimulateSMSSend simulates sending SMS (for development)
func (s *AuthService) SimulateSMSSend(phone, code string) error {
	// In production, this would integrate with a real SMS service
	// For now, we just log the OTP code
	fmt.Printf("📱 SMS to %s: Your ILEX verification code is: %s (expires in %d minutes)\n", 
		phone, code, s.config.OTPExpiration)
	return nil
}
