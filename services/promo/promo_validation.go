package promo

import (
	"context"
	"fmt"
	"time"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

// PromoValidationService gère la validation et l'application des codes promotionnels
type PromoValidationService struct {
	config *config.Config
}

// NewPromoValidationService crée une nouvelle instance du service de validation
func NewPromoValidationService(cfg *config.Config) *PromoValidationService {
	return &PromoValidationService{
		config: cfg,
	}
}

// ValidatePromoCode valide un code promotionnel
func (s *PromoValidationService) ValidatePromoCode(code string, amount float64, userID string) (*models.PromoValidationResult, error) {
	// Récupérer le promo
	promo, err := s.getPromoByCode(code)
	if err != nil {
		return nil, fmt.Errorf("échec de la récupération du promo: %v", err)
	}

	if promo == nil {
		return &models.PromoValidationResult{
			Valid:   false,
			Message: "Code promotionnel non trouvé",
		}, nil
	}

	// Vérifier si le promo est actif
	if !promo.IsActive {
		return &models.PromoValidationResult{
			Valid:   false,
			Message: "Ce code promotionnel n'est plus actif",
		}, nil
	}

	// Vérifier les dates
	now := time.Now()
	if now.Before(promo.StartDate) {
		return &models.PromoValidationResult{
			Valid:   false,
			Message: "Ce code promotionnel n'est pas encore valide",
		}, nil
	}

	if now.After(promo.EndDate) {
		return &models.PromoValidationResult{
			Valid:   false,
			Message: "Ce code promotionnel a expiré",
		}, nil
	}

	// Vérifier le montant minimum d'achat
	if promo.MinPurchaseAmount != nil && amount < *promo.MinPurchaseAmount {
		return &models.PromoValidationResult{
			Valid:   false,
			Message: fmt.Sprintf("Montant minimum d'achat requis: %.2f FCFA", *promo.MinPurchaseAmount),
		}, nil
	}

	// Vérifier la limite d'utilisation
	if promo.MaxUsage != nil && promo.UsageCount != nil && *promo.UsageCount >= *promo.MaxUsage {
		return &models.PromoValidationResult{
			Valid:   false,
			Message: "Ce code promotionnel a atteint sa limite d'utilisation",
		}, nil
	}

	// Vérifier si l'utilisateur a déjà utilisé ce promo
	hasUsed, err := s.hasUserUsedPromo(userID, promo.ID)
	if err != nil {
		return nil, fmt.Errorf("échec de la vérification de l'utilisation: %v", err)
	}

	if hasUsed {
		return &models.PromoValidationResult{
			Valid:   false,
			Message: "Vous avez déjà utilisé ce code promotionnel",
		}, nil
	}

	// Calculer la remise
	discount, err := s.calculateDiscount(promo, amount)
	if err != nil {
		return nil, fmt.Errorf("échec du calcul de la remise: %v", err)
	}

	return &models.PromoValidationResult{
		Valid:      true,
		Message:    "Code promotionnel valide",
		Discount:   discount,
		FinalPrice: amount - discount,
	}, nil
}

// ValidateAndCalculateDiscount valide et calcule la remise
func (s *PromoValidationService) ValidateAndCalculateDiscount(code string, amount float64) (float64, error) {
	result, err := s.ValidatePromoCode(code, amount, "")
	if err != nil {
		return 0, err
	}

	if !result.Valid {
		return 0, fmt.Errorf("code promotionnel invalide: %s", result.Message)
	}

	return result.Discount, nil
}

// ApplyPromo applique un code promotionnel
func (s *PromoValidationService) ApplyPromo(code string, amount float64, userID string) (*models.PromoUsage, error) {
	// Valider le code
	result, err := s.ValidatePromoCode(code, amount, userID)
	if err != nil {
		return nil, fmt.Errorf("échec de la validation: %v", err)
	}

	if !result.Valid {
		return nil, fmt.Errorf("code promotionnel invalide: %s", result.Message)
	}

	// Créer l'enregistrement d'utilisation
	usage := &models.PromoUsage{
		ID:       fmt.Sprintf("usage_%d", time.Now().Unix()),
		UserID:   userID,
		PromoID:  "temp-promo-id", // TODO: récupérer le vrai ID du promo
		Amount:   amount,
		Discount: result.Discount,
		UsedAt:   time.Now(),
	}

	// Sauvegarder l'utilisation
	if err := s.savePromoUsage(usage); err != nil {
		return nil, fmt.Errorf("échec de la sauvegarde de l'utilisation: %v", err)
	}

	// TODO: Incrémenter le compteur d'utilisation
	// if err := s.incrementPromoUsage(promoID); err != nil {
	//     return nil, fmt.Errorf("échec de l'incrémentation du compteur: %v", err)
	// }

	return usage, nil
}

// calculateDiscount calcule la remise selon le type de promo
func (s *PromoValidationService) calculateDiscount(promo *models.Promo, amount float64) (float64, error) {
	var discount float64

	switch promo.Type {
	case models.PromoTypeFixedAmount:
		discount = promo.Value
	case models.PromoTypePercentage:
		discount = amount * (promo.Value / 100)
	case models.PromoTypeFreeDelivery:
		// Pour la livraison gratuite, on ne calcule pas de remise sur le montant
		discount = 0
	default:
		return 0, fmt.Errorf("type de promo non supporté: %s", promo.Type)
	}

	// S'assurer que la remise ne dépasse pas le montant de la commande
	if discount > amount {
		discount = amount
	}

	return discount, nil
}

// getPromoByCode récupère un promo par son code
func (s *PromoValidationService) getPromoByCode(code string) (*models.Promo, error) {
	ctx := context.Background()
	
	promotion, err := db.PrismaDB.Promo.FindFirst(
		prismadb.Promo.Code.Equals(code),
		prismadb.Promo.IsActive.Equals(true),
	).Exec(ctx)

	if err != nil {
		if err == prismadb.ErrNotFound {
			return nil, fmt.Errorf("code promotionnel non trouvé")
		}
		return nil, err
	}

	return s.convertPrismaPromoToModel(promotion), nil
}

// hasUserUsedPromo vérifie si un utilisateur a déjà utilisé un promo
func (s *PromoValidationService) hasUserUsedPromo(userID, promoID string) (bool, error) {
	ctx := context.Background()
	
	usages, err := db.PrismaDB.PromoUsage.FindMany(
		prismadb.PromoUsage.UserID.Equals(userID),
		prismadb.PromoUsage.PromoID.Equals(promoID),
	).Exec(ctx)
	
	if err != nil {
		return false, err
	}

	return len(usages) > 0, nil
}

// savePromoUsage sauvegarde une utilisation de promo
func (s *PromoValidationService) savePromoUsage(usage *models.PromoUsage) error {
	ctx := context.Background()
	
	_, err := db.PrismaDB.PromoUsage.CreateOne(
		prismadb.PromoUsage.Promo.Link(prismadb.Promo.ID.Equals(usage.PromoID)),
		prismadb.PromoUsage.User.Link(prismadb.User.ID.Equals(usage.UserID)),
	).Exec(ctx)
	
	return err
}

// incrementPromoUsage incrémente le compteur d'utilisation d'un promo
func (s *PromoValidationService) incrementPromoUsage(promoID string) error {
	ctx := context.Background()
	
	// Récupérer le promo actuel pour obtenir le compteur
	promo, err := db.PrismaDB.Promo.FindUnique(
		prismadb.Promo.ID.Equals(promoID),
	).Exec(ctx)
	if err != nil {
		return err
	}
	
	// Calculer le nouveau compteur
	currentCount := promo.UsageCount
	newCount := currentCount + 1
	
	// Mettre à jour
	_, err = db.PrismaDB.Promo.FindUnique(
		prismadb.Promo.ID.Equals(promoID),
	).Update(
		prismadb.Promo.UsageCount.Set(newCount),
		prismadb.Promo.UpdatedAt.Set(time.Now()),
	).Exec(ctx)
	
	return err
}

// convertPrismaPromoToModel convertit un modèle Prisma en modèle de domaine
func (s *PromoValidationService) convertPrismaPromoToModel(prismaPromo *prismadb.PromoModel) *models.Promo {
	promo := &models.Promo{
		ID:        prismaPromo.ID,
		Code:      prismaPromo.Code,
		Type:      models.PromoType(prismaPromo.Type),
		Value:     prismaPromo.Value,
		StartDate: prismaPromo.ValidFrom,
		EndDate:   prismaPromo.ValidUntil,
		IsActive:  prismaPromo.IsActive,
		CreatedAt: prismaPromo.CreatedAt,
		UpdatedAt: prismaPromo.UpdatedAt,
	}
	
	// Gérer les champs optionnels
	if description, ok := prismaPromo.Description(); ok {
		promo.Description = &description
	}
	if minAmount, ok := prismaPromo.MinPurchaseAmount(); ok {
		promo.MinPurchaseAmount = &minAmount
	}
	if maxUsage, ok := prismaPromo.MaxUsage(); ok {
		promo.MaxUsage = &maxUsage
	}
	promo.UsageCount = &prismaPromo.UsageCount
	
	return promo
}
