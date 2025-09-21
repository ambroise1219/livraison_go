package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

// OTPService gère la génération, l'envoi et la vérification des OTP
type OTPService struct {
	config          *config.Config
	whatsappService interface {
		SendOTP(phone, code, firstName, lastName string) error
	}
}

// NewOTPService crée une nouvelle instance du service OTP
func NewOTPService(cfg *config.Config) *OTPService {
	return &OTPService{
		config:          cfg,
		whatsappService: NewWanotifierService(cfg),
	}
}

// GenerateOTP génère un code OTP aléatoire
func (s *OTPService) GenerateOTP() string {
	otpLen := s.config.OTPLength
	if otpLen <= 0 {
		otpLen = 4
	}
	if otpLen < 4 {
		otpLen = 4
	}
	if otpLen > 10 {
		otpLen = 10
	}

	digits := "0123456789"
	buf := make([]byte, otpLen)
	for i := 0; i < otpLen; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		buf[i] = digits[n.Int64()]
	}
	return string(buf)
}

// SendOTPByWhatsApp envoie un OTP via WhatsApp
func (s *OTPService) SendOTPByWhatsApp(phone, firstName, lastName string) (*models.OTP, error) {
	// Normaliser le numéro pour WhatsApp (par défaut +225)
	normalized := s.normalizePhoneForWhatsApp(phone)

	// Générer l'OTP
	code := s.GenerateOTP()

	// Sauvegarder l'OTP en base
	otp, err := s.SaveOTPWithCode(normalized, code)
	if err != nil {
		return nil, fmt.Errorf("échec de la sauvegarde de l'OTP: %v", err)
	}

	// Envoyer via WhatsApp (simulé pour l'instant)
	if s.whatsappService != nil {
		err = s.whatsappService.SendOTP(normalized, code, firstName, lastName)
		if err != nil {
			return nil, fmt.Errorf("échec de l'envoi WhatsApp: %v", err)
		}
	}

	return otp, nil
}

// SaveOTPWithCode sauvegarde un OTP avec un code spécifique
func (s *OTPService) SaveOTPWithCode(phone, code string) (*models.OTP, error) {
	// Toujours enregistrer le numéro au format E.164 avec +225 par défaut
	normalized := s.normalizePhoneForWhatsApp(phone)

	// Rate limiting: limiter les OTP par numéro dans une fenêtre glissante
	if err := s.enforceOTPRateLimit(normalized); err != nil {
		return nil, err
	}

	// Vérifier si un OTP existe déjà pour ce numéro
	existingOTP, err := s.getExistingOTP(normalized)
	if err != nil {
		return nil, fmt.Errorf("échec de la vérification de l'OTP existant: %v", err)
	}

	otp := &models.OTP{
		ID:        uuid.New().String(),
		Phone:     normalized,
		Code:      code,
		ExpiresAt: time.Now().Add(time.Duration(s.config.OTPExpiration) * time.Minute),
		CreatedAt: time.Now(),
	}

	if existingOTP != nil {
		// Mettre à jour l'OTP existant
		otp.ID = existingOTP.ID
		err = s.updateOTP(otp)
	} else {
		// Créer un nouvel OTP
		err = s.createOTP(otp)
	}

	if err != nil {
		return nil, fmt.Errorf("échec de la sauvegarde de l'OTP: %v", err)
	}

	return otp, nil
}

// SaveOTP génère et sauvegarde un nouvel OTP
func (s *OTPService) SaveOTP(phone string) (*models.OTP, error) {
	code := s.GenerateOTP()
	return s.SaveOTPWithCode(s.normalizePhoneForWhatsApp(phone), code)
}

// VerifyOTP vérifie un code OTP
func (s *OTPService) VerifyOTP(phone, code string) (*models.OTP, error) {
	// Récupérer l'OTP depuis la base
	otp, err := s.getOTPByPhone(s.normalizePhoneForWhatsApp(phone))
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération de l'OTP: %v", err)
	}

	if otp == nil {
		return nil, fmt.Errorf("aucun OTP trouvé pour ce numéro")
	}

	// Vérifier si l'OTP est expiré
	if time.Now().After(otp.ExpiresAt) {
		return nil, fmt.Errorf("l'OTP a expiré")
	}

	// TODO: Vérifier si l'OTP est déjà utilisé (nécessite un champ IsUsed dans le modèle)

	// Vérifier le code
	if otp.Code != code {
		return nil, fmt.Errorf("code OTP incorrect")
	}

	// TODO: Marquer l'OTP comme utilisé (nécessite un champ IsUsed dans le modèle)

	return otp, nil
}

// getExistingOTP récupère un OTP existant pour un numéro
func (s *OTPService) getExistingOTP(phone string) (*models.OTP, error) {
	ctx := context.Background()
	
	otp, err := db.PrismaDB.Otp.FindFirst(
		prismadb.Otp.Phone.Equals(phone),
	).OrderBy(
		prismadb.Otp.CreatedAt.Order(prismadb.SortOrderDesc),
	).Exec(ctx)
	
	if err != nil {
		if err == prismadb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return &models.OTP{
		ID:        otp.ID,
		Phone:     otp.Phone,
		Code:      otp.Code,
		ExpiresAt: otp.ExpiresAt,
		CreatedAt: otp.CreatedAt,
	}, nil
}

// getOTPByPhone récupère un OTP par numéro de téléphone
func (s *OTPService) getOTPByPhone(phone string) (*models.OTP, error) {
	ctx := context.Background()
	
	otp, err := db.PrismaDB.Otp.FindFirst(
		prismadb.Otp.Phone.Equals(phone),
	).OrderBy(
		prismadb.Otp.CreatedAt.Order(prismadb.SortOrderDesc),
	).Exec(ctx)
	
	if err != nil {
		if err == prismadb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	
	return &models.OTP{
		ID:        otp.ID,
		Phone:     otp.Phone,
		Code:      otp.Code,
		ExpiresAt: otp.ExpiresAt,
		CreatedAt: otp.CreatedAt,
	}, nil
}

// normalizePhoneForWhatsApp retourne un numéro au format E.164 avec +225 par défaut.
// Règles:
// - Supprime tout caractère non numérique (sauf + en tête)
// - Si commence par '+', retourne tel quel
// - Si commence par '225' (sans +), préfixe avec '+'
// - Si commence par '0', retire le 0 et préfixe '+225'
// - Sinon, préfixe '+225'
func (s *OTPService) normalizePhoneForWhatsApp(input string) string {
	if input == "" {
		return input
	}
	// Conserver un éventuel '+' en tête, sinon retirer non-digits
	hasPlus := len(input) > 0 && input[0] == '+'
	digits := make([]rune, 0, len(input))
	for i, r := range input {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		} else if r == '+' && i == 0 {
			hasPlus = true
		}
	}
	num := string(digits)
	if hasPlus {
		return "+" + num
	}
	// 225XXXXXXXX / 0XXXXXXXXX / XXXXXXXXXX
	if len(num) >= 3 && num[:3] == "225" {
		return "+" + num
	}
	if len(num) > 0 && num[0] == '0' {
		// Garder le 0 de tête pour les numéros ivoiriens
		return "+225" + num
	}
	return "+225" + num
}

// createOTP crée un nouvel OTP en base
func (s *OTPService) createOTP(otp *models.OTP) error {
	ctx := context.Background()
	
	_, err := db.PrismaDB.Otp.CreateOne(
		prismadb.Otp.Phone.Set(otp.Phone),
		prismadb.Otp.Code.Set(otp.Code),
		prismadb.Otp.ExpiresAt.Set(otp.ExpiresAt),
	).Exec(ctx)
	
	return err
}

// updateOTP met à jour un OTP existant
func (s *OTPService) updateOTP(otp *models.OTP) error {
	ctx := context.Background()
	
	_, err := db.PrismaDB.Otp.FindUnique(
		prismadb.Otp.ID.Equals(otp.ID),
	).Update(
		prismadb.Otp.Code.Set(otp.Code),
		prismadb.Otp.ExpiresAt.Set(otp.ExpiresAt),
	).Exec(ctx)
	
	return err
}

// enforceOTPRateLimit vérifie le nombre d'OTP envoyés pour ce numéro dans la fenêtre
func (s *OTPService) enforceOTPRateLimit(phone string) error {
	windowMinutes := s.config.OTPWindowMinutes
	maxPerWindow := s.config.OTPMaxPerWindow
	if windowMinutes <= 0 || maxPerWindow <= 0 {
		return nil
	}

	ctx := context.Background()
	windowStart := time.Now().Add(-time.Duration(windowMinutes) * time.Minute)

	otps, err := db.PrismaDB.Otp.FindMany(
		prismadb.Otp.Phone.Equals(phone),
		prismadb.Otp.CreatedAt.Gte(windowStart),
	).Exec(ctx)
	
	if err != nil {
		return err
	}
	count := len(otps)

	if count >= maxPerWindow {
		return fmt.Errorf("too_many_requests: wait %d minutes", windowMinutes)
	}

	return nil
}
