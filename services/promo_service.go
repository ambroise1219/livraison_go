package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"ilex-backend/config"
	"ilex-backend/db"
	"ilex-backend/models"
)

type PromoService struct {
	config *config.Config
}

func NewPromoService(cfg *config.Config) *PromoService {
	return &PromoService{
		config: cfg,
	}
}

// CreatePromo creates a new promotional code
func (s *PromoService) CreatePromo(req *models.CreatePromoRequest) (*models.Promo, error) {
	// Validate promo code uniqueness
	exists, err := s.promoCodeExists(req.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to check promo code: %v", err)
	}
	
	if exists {
		return nil, fmt.Errorf("promo code already exists")
	}

	// Validate date range
	if req.EndDate.Before(req.StartDate) {
		return nil, fmt.Errorf("end date cannot be before start date")
	}

	if req.StartDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, fmt.Errorf("start date cannot be in the past")
	}

	// Create promo
	promo := &models.Promo{
		ID:                uuid.New().String(),
		Code:              strings.ToUpper(strings.TrimSpace(req.Code)),
		Description:       req.Description,
		Type:              req.Type,
		Value:             req.Value,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		IsActive:          true,
		MaxUsage:          req.MaxUsage,
		UsageCount:        new(int), // Initialize to 0
		MinPurchaseAmount: req.MinPurchaseAmount,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	*promo.UsageCount = 0

	// Validate promo values
	if err := s.validatePromoValues(promo); err != nil {
		return nil, fmt.Errorf("invalid promo values: %v", err)
	}

	// Save to database
	err = s.savePromo(promo)
	if err != nil {
		return nil, fmt.Errorf("failed to save promo: %v", err)
	}

	log.Printf("Created promo code: %s", promo.Code)
	return promo, nil
}

// ValidatePromoCode validates a promo code and returns validation result
func (s *PromoService) ValidatePromoCode(code string, amount float64, userID string) (*models.PromoValidationResult, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	// Get promo by code
	promo, err := s.getPromoByCode(code)
	if err != nil {
		return &models.PromoValidationResult{
			Valid:   false,
			Message: "Promo code not found",
		}, nil
	}

	// Check if promo is active and valid
	if !promo.CanBeUsed(amount) {
		message := s.getPromoInvalidMessage(promo, amount)
		return &models.PromoValidationResult{
			Valid:   false,
			Message: message,
		}, nil
	}

	// Check if user has already used this promo
	hasUsed, err := s.hasUserUsedPromo(userID, promo.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check promo usage: %v", err)
	}

	if hasUsed {
		return &models.PromoValidationResult{
			Valid:   false,
			Message: "You have already used this promo code",
		}, nil
	}

	// Calculate discount
	discount := promo.CalculateDiscount(amount)
	finalPrice := amount - discount

	if finalPrice < 0 {
		finalPrice = 0
	}

	return &models.PromoValidationResult{
		Valid:        true,
		Discount:     discount,
		FinalPrice:   finalPrice,
		Message:      fmt.Sprintf("Promo applied successfully! You saved %.0f FCFA", discount),
		DiscountType: s.getDiscountTypeString(promo.Type),
	}, nil
}

// ValidateAndCalculateDiscount validates promo and calculates discount for delivery service
func (s *PromoService) ValidateAndCalculateDiscount(code string, amount float64) (float64, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	// Get promo by code
	promo, err := s.getPromoByCode(code)
	if err != nil {
		return 0, fmt.Errorf("promo code not found")
	}

	// Check if promo can be used
	if !promo.CanBeUsed(amount) {
		return 0, fmt.Errorf("promo code cannot be used")
	}

	// Calculate discount
	discount := promo.CalculateDiscount(amount)
	return discount, nil
}

// ApplyPromo applies promo code and records usage
func (s *PromoService) ApplyPromo(code string, amount float64, userID string) (*models.PromoUsage, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	// Validate promo
	validation, err := s.ValidatePromoCode(code, amount, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate promo: %v", err)
	}

	if !validation.Valid {
		return nil, fmt.Errorf("invalid promo: %s", validation.Message)
	}

	// Get promo
	promo, err := s.getPromoByCode(code)
	if err != nil {
		return nil, fmt.Errorf("promo not found: %v", err)
	}

	// Record usage
	usage := &models.PromoUsage{
		ID:       uuid.New().String(),
		PromoID:  promo.ID,
		UserID:   userID,
		UsedAt:   time.Now(),
		Amount:   amount,
		Discount: validation.Discount,
	}

	err = s.savePromoUsage(usage)
	if err != nil {
		return nil, fmt.Errorf("failed to record promo usage: %v", err)
	}

	// Increment usage count
	err = s.incrementPromoUsage(promo.ID)
	if err != nil {
		log.Printf("Warning: failed to increment promo usage count: %v", err)
	}

	log.Printf("Applied promo %s for user %s: %.0f FCFA discount", code, userID, validation.Discount)
	return usage, nil
}

// CreateReferral creates a new referral
func (s *PromoService) CreateReferral(referrerID string, req *models.CreateReferralRequest) (*models.ReferralResponse, error) {
	// Validate referrer
	referrer, err := s.getUserByID(referrerID)
	if err != nil {
		return nil, fmt.Errorf("referrer not found: %v", err)
	}

	// Check if phone is already referred by this user
	exists, err := s.referralExists(referrerID, req.RefereePhone)
	if err != nil {
		return nil, fmt.Errorf("failed to check referral: %v", err)
	}

	if exists {
		return nil, fmt.Errorf("you have already referred this phone number")
	}

	// Check if phone is already a user
	existingUser, _ := s.getUserByPhone(req.RefereePhone)
	if existingUser != nil {
		return nil, fmt.Errorf("this phone number is already registered")
	}

	// Generate unique referral code
	code, err := s.generateReferralCode(referrerID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate referral code: %v", err)
	}

	// Create referral
	referral := &models.Referral{
		ID:           uuid.New().String(),
		ReferrerID:   referrerID,
		RefereePhone: req.RefereePhone,
		Code:         code,
		Message:      s.buildReferralMessage(referrer, req.Message),
		Status:       models.ReferralStatusPending,
		CreatedAt:    time.Now(),
		ExpiresAt:    s.calculateReferralExpiry(),
	}

	// Save referral
	err = s.saveReferral(referral)
	if err != nil {
		return nil, fmt.Errorf("failed to save referral: %v", err)
	}

	// Send referral notification
	go s.sendReferralNotification(referral, referrer)

	response := referral.ToResponse()
	response.Referrer = referrer.ToResponse()
	response.RewardAmount = s.config.ReferralRewardAmount

	log.Printf("Created referral from %s to %s", referrerID, req.RefereePhone)
	return response, nil
}

// CompleteReferral completes a referral when referee signs up
func (s *PromoService) CompleteReferral(referralCode string, refereeID string) error {
	// Find referral by code
	referral, err := s.getReferralByCode(referralCode)
	if err != nil {
		return fmt.Errorf("referral not found: %v", err)
	}

	if referral.Status != models.ReferralStatusPending {
		return fmt.Errorf("referral is not pending")
	}

	if referral.IsExpired() {
		// Update status to expired
		s.updateReferralStatus(referral.ID, models.ReferralStatusExpired)
		return fmt.Errorf("referral has expired")
	}

	// Update referral
	completedAt := time.Now()
	query := `UPDATE Referral SET 
		refereeId = $refereeId, 
		status = $status, 
		completedAt = $completedAt 
		WHERE id = $referralId`

	params := map[string]interface{}{
		"referralId":  referral.ID,
		"refereeId":   refereeID,
		"status":      string(models.ReferralStatusCompleted),
		"completedAt": completedAt,
	}

	_, err = db.Query(query, params)
	if err != nil {
		return fmt.Errorf("failed to complete referral: %v", err)
	}

	// Send completion notifications
	go s.sendReferralCompletionNotifications(referral.ReferrerID, refereeID)

	log.Printf("Completed referral %s", referralCode)
	return nil
}

// ClaimReferralReward allows referrer to claim reward
func (s *PromoService) ClaimReferralReward(referralID, userID string) error {
	// Get referral
	referral, err := s.getReferralByID(referralID)
	if err != nil {
		return fmt.Errorf("referral not found: %v", err)
	}

	// Validate ownership
	if referral.ReferrerID != userID {
		return fmt.Errorf("unauthorized: not your referral")
	}

	// Check if can be claimed
	if !referral.CanBeClaimed() {
		return fmt.Errorf("referral reward cannot be claimed")
	}

	// Update referral
	claimedAt := time.Now()
	query := `UPDATE Referral SET 
		status = $status, 
		rewardClaimedAt = $claimedAt 
		WHERE id = $referralId`

	params := map[string]interface{}{
		"referralId": referralID,
		"status":     string(models.ReferralStatusRewardClaimed),
		"claimedAt":  claimedAt,
	}

	_, err = db.Query(query, params)
	if err != nil {
		return fmt.Errorf("failed to claim reward: %v", err)
	}

	// Add reward to user wallet
	err = s.addWalletCredit(userID, s.config.ReferralRewardAmount, "Referral reward")
	if err != nil {
		log.Printf("Warning: failed to add wallet credit: %v", err)
	}

	log.Printf("User %s claimed referral reward: %.0f FCFA", userID, s.config.ReferralRewardAmount)
	return nil
}

// Helper methods

func (s *PromoService) validatePromoValues(promo *models.Promo) error {
	if promo.Value <= 0 {
		return fmt.Errorf("promo value must be positive")
	}

	switch promo.Type {
	case models.PromoTypePercentage:
		if promo.Value > 100 {
			return fmt.Errorf("percentage discount cannot exceed 100%%")
		}
	case models.PromoTypeFixedAmount:
		if promo.Value > 50000 { // Max 50k FCFA discount
			return fmt.Errorf("fixed discount cannot exceed 50,000 FCFA")
		}
	case models.PromoTypeFreeDelivery:
		// No specific validation needed
	default:
		return fmt.Errorf("invalid promo type: %s", promo.Type)
	}

	if promo.MinPurchaseAmount != nil && *promo.MinPurchaseAmount < 0 {
		return fmt.Errorf("minimum purchase amount cannot be negative")
	}

	if promo.MaxUsage != nil && *promo.MaxUsage <= 0 {
		return fmt.Errorf("max usage must be positive")
	}

	return nil
}

func (s *PromoService) getPromoInvalidMessage(promo *models.Promo, amount float64) string {
	if promo.IsExpired() {
		return "Promo code has expired"
	}

	if !promo.IsActive {
		return "Promo code is not active"
	}

	if time.Now().Before(promo.StartDate) {
		return "Promo code is not yet active"
	}

	if promo.HasReachedMaxUsage() {
		return "Promo code usage limit reached"
	}

	if promo.MinPurchaseAmount != nil && amount < *promo.MinPurchaseAmount {
		return fmt.Sprintf("Minimum purchase amount is %.0f FCFA", *promo.MinPurchaseAmount)
	}

	return "Promo code cannot be used"
}

func (s *PromoService) getDiscountTypeString(promoType models.PromoType) string {
	switch promoType {
	case models.PromoTypePercentage:
		return "percentage"
	case models.PromoTypeFixedAmount:
		return "fixed"
	case models.PromoTypeFreeDelivery:
		return "free_delivery"
	default:
		return "unknown"
	}
}

func (s *PromoService) generateReferralCode(referrerID string) (string, error) {
	// Simple code generation: REF + first 6 chars of user ID + timestamp suffix
	timestamp := fmt.Sprintf("%d", time.Now().Unix()%10000) // Last 4 digits
	userSuffix := referrerID[:6] // First 6 chars of user ID
	code := fmt.Sprintf("REF%s%s", strings.ToUpper(userSuffix), timestamp)
	
	// Ensure uniqueness
	exists, err := s.referralCodeExists(code)
	if err != nil {
		return "", err
	}
	
	if exists {
		// Add random suffix if collision
		code = fmt.Sprintf("%s%d", code, time.Now().Nanosecond()%100)
	}
	
	return code, nil
}

func (s *PromoService) buildReferralMessage(referrer *models.User, customMessage *string) string {
	defaultMessage := fmt.Sprintf("Join ILEX using my referral code and get %.0f FCFA bonus! - %s", 
		s.config.ReferralRewardAmount, referrer.GetFullName())
	
	if customMessage != nil && strings.TrimSpace(*customMessage) != "" {
		return fmt.Sprintf("%s\n\n%s", *customMessage, defaultMessage)
	}
	
	return defaultMessage
}

func (s *PromoService) calculateReferralExpiry() *time.Time {
	expiry := time.Now().Add(time.Duration(s.config.ReferralExpiration) * 24 * time.Hour)
	return &expiry
}

// Database operations

func (s *PromoService) promoCodeExists(code string) (bool, error) {
	query := `SELECT id FROM Promo WHERE code = $code LIMIT 1`
	params := map[string]interface{}{"code": strings.ToUpper(code)}
	
	_, err := db.QuerySingle(query, params)
	if err != nil {
		if err.Error() == "no result found" {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}

func (s *PromoService) referralExists(referrerID, phone string) (bool, error) {
	query := `SELECT id FROM Referral WHERE referrerId = $referrerId AND refereePhone = $phone LIMIT 1`
	params := map[string]interface{}{
		"referrerId": referrerID,
		"phone":      phone,
	}
	
	_, err := db.QuerySingle(query, params)
	if err != nil {
		if err.Error() == "no result found" {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}

func (s *PromoService) referralCodeExists(code string) (bool, error) {
	query := `SELECT id FROM Referral WHERE code = $code LIMIT 1`
	params := map[string]interface{}{"code": code}
	
	_, err := db.QuerySingle(query, params)
	if err != nil {
		if err.Error() == "no result found" {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}

// Placeholder methods - would need full implementation

func (s *PromoService) savePromo(promo *models.Promo) error {
	// Implementation for saving promo to database
	return nil
}

func (s *PromoService) getPromoByCode(code string) (*models.Promo, error) {
	// Implementation for getting promo by code
	return nil, fmt.Errorf("not implemented")
}

func (s *PromoService) hasUserUsedPromo(userID, promoID string) (bool, error) {
	// Implementation for checking if user has used promo
	return false, nil
}

func (s *PromoService) savePromoUsage(usage *models.PromoUsage) error {
	// Implementation for saving promo usage
	return nil
}

func (s *PromoService) incrementPromoUsage(promoID string) error {
	// Implementation for incrementing promo usage count
	return nil
}

func (s *PromoService) getUserByID(userID string) (*models.User, error) {
	// Implementation for getting user by ID
	return nil, fmt.Errorf("not implemented")
}

func (s *PromoService) getUserByPhone(phone string) (*models.User, error) {
	// Implementation for getting user by phone
	return nil, fmt.Errorf("not implemented")
}

func (s *PromoService) saveReferral(referral *models.Referral) error {
	// Implementation for saving referral
	return nil
}

func (s *PromoService) getReferralByCode(code string) (*models.Referral, error) {
	// Implementation for getting referral by code
	return nil, fmt.Errorf("not implemented")
}

func (s *PromoService) getReferralByID(referralID string) (*models.Referral, error) {
	// Implementation for getting referral by ID
	return nil, fmt.Errorf("not implemented")
}

func (s *PromoService) updateReferralStatus(referralID string, status models.ReferralStatus) error {
	// Implementation for updating referral status
	return nil
}

func (s *PromoService) addWalletCredit(userID string, amount float64, description string) error {
	// Implementation for adding wallet credit
	return nil
}

func (s *PromoService) sendReferralNotification(referral *models.Referral, referrer *models.User) {
	// Implementation for sending referral notification
}

func (s *PromoService) sendReferralCompletionNotifications(referrerID, refereeID string) {
	// Implementation for sending referral completion notifications
}