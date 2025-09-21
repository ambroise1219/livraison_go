package promo

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
)

// PromoReferralsService gère le système de parrainage
type PromoReferralsService struct {
	config *config.Config
}

// NewPromoReferralsService crée une nouvelle instance du service de parrainage
func NewPromoReferralsService(cfg *config.Config) *PromoReferralsService {
	return &PromoReferralsService{
		config: cfg,
	}
}

// CreateReferral crée un nouveau parrainage
func (s *PromoReferralsService) CreateReferral(referrerID string, req *models.CreateReferralRequest) (*models.ReferralResponse, error) {
	// Vérifier si un parrainage existe déjà pour ce téléphone
	exists, err := s.referralExists(referrerID, req.RefereePhone)
	if err != nil {
		return nil, fmt.Errorf("échec de la vérification du parrainage existant: %v", err)
	}

	if exists {
		return nil, fmt.Errorf("un parrainage existe déjà pour ce numéro de téléphone")
	}

	// Générer un code de parrainage unique
	code, err := s.generateReferralCode(referrerID)
	if err != nil {
		return nil, fmt.Errorf("échec de la génération du code de parrainage: %v", err)
	}

	// Récupérer les informations du parrain (pour validation)
	_, err = s.getUserByID(referrerID)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération du parrain: %v", err)
	}

	// Créer le parrainage
	referral := &models.Referral{
		ID:           uuid.New().String(),
		ReferrerID:   referrerID,
		RefereePhone: req.RefereePhone,
		Code:         code,
		Status:       models.ReferralStatusPending,
		Message:      "Parrainage créé avec succès",
		CreatedAt:    time.Now(),
	}

	// Sauvegarder en base
	if err := s.saveReferral(referral); err != nil {
		return nil, fmt.Errorf("échec de la sauvegarde: %v", err)
	}

	// Créer la réponse
	response := &models.ReferralResponse{
		ID:           referral.ID,
		ReferrerID:   referral.ReferrerID,
		RefereePhone: referral.RefereePhone,
		Code:         referral.Code,
		Message:      referral.Message,
		Status:       referral.Status,
		CreatedAt:    referral.CreatedAt,
	}

	return response, nil
}

// GetReferralByCode récupère un parrainage par son code
func (s *PromoReferralsService) GetReferralByCode(code string) (*models.Referral, error) {
	query := `SELECT id, referrer_id, referee_phone, referee_id, code, message, status, 
		created_at, completed_at, expires_at FROM referrals WHERE code = ? LIMIT 1`

	row := db.QueryRow(query, code)

	referral, err := s.scanReferralFromRow(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("code de parrainage non trouvé")
	}
	if err != nil {
		return nil, fmt.Errorf("erreur de récupération: %v", err)
	}

	return referral, nil
}

// GetReferralsByReferrer récupère les parrainages d'un parrain
func (s *PromoReferralsService) GetReferralsByReferrer(referrerID string) ([]*models.Referral, error) {
	query := `SELECT id, referrer_id, referee_phone, referee_id, code, message, status, 
		created_at, completed_at, expires_at FROM referrals WHERE referrer_id = ? ORDER BY created_at DESC`

	rows, err := db.QueryRows(query, referrerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var referrals []*models.Referral
	for rows.Next() {
		referral, err := s.scanReferralFromRow(rows)
		if err != nil {
			continue // Skip invalid records
		}
		referrals = append(referrals, referral)
	}

	return referrals, nil
}

// UpdateReferralStatus met à jour le statut d'un parrainage
func (s *PromoReferralsService) UpdateReferralStatus(referralID string, status models.ReferralStatus) (*models.Referral, error) {
	// Récupérer le parrainage existant
	referral, err := s.getReferralByID(referralID)
	if err != nil {
		return nil, fmt.Errorf("parrainage non trouvé: %v", err)
	}

	// Mettre à jour le statut
	referral.Status = status
	if status == models.ReferralStatusCompleted {
		now := time.Now()
		referral.CompletedAt = &now
	}

	// Sauvegarder les modifications
	if err := s.updateReferral(referral); err != nil {
		return nil, fmt.Errorf("échec de la mise à jour: %v", err)
	}

	return referral, nil
}

// GetReferralStats récupère les statistiques de parrainage d'un utilisateur
func (s *PromoReferralsService) GetReferralStats(userID string) (map[string]interface{}, error) {
	// Compter les parrainages par statut
	query := `SELECT status, COUNT(*) FROM referrals WHERE referrer_id = ? GROUP BY status`
	rows, err := db.QueryRows(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := map[string]interface{}{
		"total":         0,
		"pending":       0,
		"completed":     0,
		"cancelled":     0,
		"total_rewards": 0.0,
	}

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}

		stats["total"] = stats["total"].(int) + count
		stats[status] = count

		if status == string(models.ReferralStatusCompleted) {
			// Calculer les récompenses totales
			rewardQuery := `SELECT SUM(reward_amount) FROM referrals WHERE referrer_id = ? AND status = ?`
			var totalRewards sql.NullFloat64
			err := db.QueryRow(rewardQuery, userID, models.ReferralStatusCompleted).Scan(&totalRewards)
			if err == nil && totalRewards.Valid {
				stats["total_rewards"] = totalRewards.Float64
			}
		}
	}

	return stats, nil
}

// Helper methods

func (s *PromoReferralsService) referralExists(referrerID, refereePhone string) (bool, error) {
	query := `SELECT COUNT(*) FROM referrals WHERE referrer_id = ? AND referee_phone = ?`
	var count int
	err := db.QueryRow(query, referrerID, refereePhone).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *PromoReferralsService) generateReferralCode(referrerID string) (string, error) {
	// Générer un code unique basé sur l'ID du parrain et un timestamp
	rand.Seed(time.Now().UnixNano())
	suffix := rand.Intn(9999)
	code := fmt.Sprintf("REF%s%04d", referrerID[:8], suffix)

	// Vérifier l'unicité
	exists, err := s.referralCodeExists(code)
	if err != nil {
		return "", err
	}

	if exists {
		// Régénérer si le code existe déjà
		return s.generateReferralCode(referrerID)
	}

	return code, nil
}

func (s *PromoReferralsService) referralCodeExists(code string) (bool, error) {
	query := `SELECT COUNT(*) FROM referrals WHERE code = ?`
	var count int
	err := db.QueryRow(query, code).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *PromoReferralsService) getUserByID(userID string) (*models.User, error) {
	query := `SELECT id, first_name, last_name, phone, email, role, created_at, updated_at 
		FROM users WHERE id = ? LIMIT 1`
	row := db.QueryRow(query, userID)

	var user models.User
	var email sql.NullString
	err := row.Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Phone, &email, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if email.Valid {
		user.Email = &email.String
	}

	return &user, nil
}

func (s *PromoReferralsService) saveReferral(referral *models.Referral) error {
	query := `INSERT INTO referrals (id, referrer_id, referee_phone, referee_id, code, message, status, 
		created_at, completed_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.ExecuteQuery(query,
		referral.ID, referral.ReferrerID, referral.RefereePhone, referral.RefereeID, referral.Code,
		referral.Message, referral.Status, referral.CreatedAt, referral.CompletedAt, referral.ExpiresAt,
	)
	return err
}

func (s *PromoReferralsService) getReferralByID(referralID string) (*models.Referral, error) {
	query := `SELECT id, referrer_id, referee_phone, referee_id, code, message, status, 
		created_at, completed_at, expires_at FROM referrals WHERE id = ? LIMIT 1`

	row := db.QueryRow(query, referralID)

	referral, err := s.scanReferralFromRow(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("parrainage non trouvé")
	}
	if err != nil {
		return nil, fmt.Errorf("erreur de récupération: %v", err)
	}

	return referral, nil
}

func (s *PromoReferralsService) updateReferral(referral *models.Referral) error {
	query := `UPDATE referrals SET status = ?, completed_at = ? WHERE id = ?`
	_, err := db.ExecuteQuery(query, referral.Status, referral.CompletedAt, referral.ID)
	return err
}

func (s *PromoReferralsService) scanReferralFromRow(scanner interface{}) (*models.Referral, error) {
	referral := &models.Referral{}

	var refereeID sql.NullString
	var completedAt, expiresAt sql.NullTime

	switch row := scanner.(type) {
	case *sql.Row:
		err := row.Scan(
			&referral.ID, &referral.ReferrerID, &referral.RefereePhone, &refereeID, &referral.Code,
			&referral.Message, &referral.Status, &referral.CreatedAt, &completedAt, &expiresAt,
		)
		if err != nil {
			return nil, err
		}
	case *sql.Rows:
		err := row.Scan(
			&referral.ID, &referral.ReferrerID, &referral.RefereePhone, &refereeID, &referral.Code,
			&referral.Message, &referral.Status, &referral.CreatedAt, &completedAt, &expiresAt,
		)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("type de scanner non supporté")
	}

	// Handle nullable fields
	if refereeID.Valid {
		referral.RefereeID = &refereeID.String
	}
	if completedAt.Valid {
		referral.CompletedAt = &completedAt.Time
	}
	if expiresAt.Valid {
		referral.ExpiresAt = &expiresAt.Time
	}

	return referral, nil
}

func (s *PromoReferralsService) generateShareURL(code string) string {
	// TODO: Implémenter la génération d'URL de partage
	return fmt.Sprintf("https://app.livraison.com/referral/%s", code)
}

// CompleteReferral complète un parrainage
func (s *PromoReferralsService) CompleteReferral(referralCode string, refereeID string) error {
	// Récupérer le parrainage par code
	referral, err := s.GetReferralByCode(referralCode)
	if err != nil {
		return fmt.Errorf("code de parrainage non trouvé: %v", err)
	}

	// Vérifier que le parrainage est en attente
	if referral.Status != models.ReferralStatusPending {
		return fmt.Errorf("le parrainage n'est pas en attente")
	}

	// Mettre à jour le statut
	referral.Status = models.ReferralStatusCompleted
	now := time.Now()
	referral.CompletedAt = &now

	// Sauvegarder les modifications
	if err := s.updateReferral(referral); err != nil {
		return fmt.Errorf("échec de la mise à jour: %v", err)
	}

	return nil
}

// ClaimReferralReward réclame la récompense de parrainage
func (s *PromoReferralsService) ClaimReferralReward(referralID, userID string) error {
	// Récupérer le parrainage
	referral, err := s.getReferralByID(referralID)
	if err != nil {
		return fmt.Errorf("parrainage non trouvé: %v", err)
	}

	// Vérifier que le parrainage est complété
	if referral.Status != models.ReferralStatusCompleted {
		return fmt.Errorf("le parrainage doit être complété avant de réclamer la récompense")
	}

	// Vérifier que l'utilisateur est le parrain
	if referral.ReferrerID != userID {
		return fmt.Errorf("vous n'êtes pas autorisé à réclamer cette récompense")
	}

	// TODO: Implémenter la logique de réclamation de récompense
	// Pour l'instant, on considère que la récompense est automatiquement accordée

	return nil
}
