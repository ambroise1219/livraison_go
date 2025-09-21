package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ambroise1219/livraison_go/models"
)

// JWTClaims représente les claims du token JWT
type JWTClaims struct {
	UserID             string           `json:"user_id"`
	Phone              string           `json:"phone"`
	Role               models.UserRole  `json:"role"`
	IsProfileCompleted bool            `json:"is_profile_completed"`
	jwt.RegisteredClaims
}

// JWTService gère les opérations JWT
type JWTService struct {
	secretKey     []byte
	tokenDuration time.Duration
}

// NewJWTService crée une nouvelle instance du service JWT
func NewJWTService(secretKey string, tokenDuration time.Duration) *JWTService {
	return &JWTService{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

// GenerateToken génère un token JWT pour un utilisateur
func (j *JWTService) GenerateToken(user *models.User) (string, error) {
	now := time.Now()
	
	claims := JWTClaims{
		UserID:             user.ID,
		Phone:              user.Phone,
		Role:               user.Role,
		IsProfileCompleted: user.IsProfileCompleted,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.tokenDuration)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "livraison-go",
			Audience:  []string{"livraison-app"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken valide un token JWT et retourne les claims
func (j *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Vérifier la méthode de signature
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("méthode de signature invalide")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token invalide")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("claims invalides")
	}

	return claims, nil
}

// RefreshToken génère un nouveau token si l'ancien est encore valide mais proche de l'expiration
func (j *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Vérifier si le token expire dans moins d'1 heure
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", errors.New("le token n'a pas besoin d'être rafraîchi")
	}

	// Créer un nouveau token avec les mêmes claims mais nouvelle expiration
	user := &models.User{
		ID:                 claims.UserID,
		Phone:              claims.Phone,
		Role:               claims.Role,
		IsProfileCompleted: claims.IsProfileCompleted,
	}

	return j.GenerateToken(user)
}

// ExtractUserFromToken extrait les informations utilisateur du token
func (j *JWTService) ExtractUserFromToken(tokenString string) (*models.User, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:                 claims.UserID,
		Phone:              claims.Phone,
		Role:               claims.Role,
		IsProfileCompleted: claims.IsProfileCompleted,
	}, nil
}