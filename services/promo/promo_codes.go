package promo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

// PromoCodesService gère la création et la gestion des codes promotionnels
type PromoCodesService struct {
	config *config.Config
}

// NewPromoCodesService crée une nouvelle instance du service des codes promotionnels
func NewPromoCodesService(cfg *config.Config) *PromoCodesService {
	return &PromoCodesService{
		config: cfg,
	}
}

// CreatePromo crée un nouveau code promotionnel
func (s *PromoCodesService) CreatePromo(req *models.CreatePromoRequest) (*models.Promo, error) {
	// Valider l'unicité du code promo
	exists, err := s.promoCodeExists(req.Code)
	if err != nil {
		return nil, fmt.Errorf("échec de la vérification du code promo: %v", err)
	}

	if exists {
		return nil, fmt.Errorf("le code promo existe déjà")
	}

	// Valider la plage de dates
	if req.EndDate.Before(req.StartDate) {
		return nil, fmt.Errorf("la date de fin ne peut pas être antérieure à la date de début")
	}

	if req.StartDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, fmt.Errorf("la date de début ne peut pas être dans le passé")
	}

	// Créer le promo
	promo := &models.Promo{
		ID:                uuid.New().String(),
		Code:              strings.ToUpper(strings.TrimSpace(req.Code)),
		Description:       req.Description,
		Type:              req.Type,
		Value:             req.Value,
		MinPurchaseAmount: req.MinPurchaseAmount,
		MaxUsage:          req.MaxUsage,
		UsageCount:        &[]int{0}[0],
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		IsActive:          true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Valider les valeurs
	if err := s.validatePromoValues(promo); err != nil {
		return nil, fmt.Errorf("validation échouée: %v", err)
	}

	// Sauvegarder en base
	if err := s.savePromo(promo); err != nil {
		return nil, fmt.Errorf("échec de la sauvegarde: %v", err)
	}

	return promo, nil
}

// GetPromoByCode récupère un code promotionnel par son code
func (s *PromoCodesService) GetPromoByCode(code string) (*models.Promo, error) {
	ctx := context.Background()
	
	promotion, err := db.PrismaDB.Promo.FindFirst(
		prismadb.Promo.Code.Equals(strings.ToUpper(strings.TrimSpace(code))),
		prismadb.Promo.IsActive.Equals(true),
	).Exec(ctx)

	if err != nil {
		if err == prismadb.ErrNotFound {
			return nil, fmt.Errorf("code promotionnel non trouvé")
		}
		return nil, fmt.Errorf("erreur de récupération: %v", err)
	}

	return s.convertPrismaPromoToModel(promotion), nil
}

// UpdatePromo met à jour un code promotionnel
func (s *PromoCodesService) UpdatePromo(promoID string, updates *models.UpdatePromoRequest) (*models.Promo, error) {
	// Récupérer le promo existant
	promo, err := s.getPromoByID(promoID)
	if err != nil {
		return nil, fmt.Errorf("promo non trouvé: %v", err)
	}

	// Appliquer les mises à jour
	if updates.Description != nil {
		promo.Description = updates.Description
	}
	if updates.Value != nil {
		promo.Value = *updates.Value
	}
	if updates.MinPurchaseAmount != nil {
		promo.MinPurchaseAmount = updates.MinPurchaseAmount
	}
	if updates.MaxUsage != nil {
		promo.MaxUsage = updates.MaxUsage
	}
	if updates.StartDate != nil {
		promo.StartDate = *updates.StartDate
	}
	if updates.EndDate != nil {
		promo.EndDate = *updates.EndDate
	}
	if updates.IsActive != nil {
		promo.IsActive = *updates.IsActive
	}

	promo.UpdatedAt = time.Now()

	// Valider les valeurs
	if err := s.validatePromoValues(promo); err != nil {
		return nil, fmt.Errorf("validation échouée: %v", err)
	}

	// Sauvegarder les modifications
	if err := s.updatePromo(promo); err != nil {
		return nil, fmt.Errorf("échec de la mise à jour: %v", err)
	}

	return promo, nil
}

// DeletePromo supprime un code promotionnel
func (s *PromoCodesService) DeletePromo(promoID string) error {
	ctx := context.Background()
	
	// Vérifier que le promo existe
	_, err := s.getPromoByID(promoID)
	if err != nil {
		return fmt.Errorf("promo non trouvé: %v", err)
	}

	// Supprimer le promo
	_, err = db.PrismaDB.Promo.FindUnique(
		prismadb.Promo.ID.Equals(promoID),
	).Delete().Exec(ctx)
	
	if err != nil {
		return fmt.Errorf("échec de la suppression: %v", err)
	}

	return nil
}

// GetPromoStats récupère les statistiques d'un code promotionnel
func (s *PromoCodesService) GetPromoStats(promoID string) (map[string]interface{}, error) {
	// TODO: Implémenter les vraies requêtes de statistiques
	// Pour l'instant, retourner des valeurs par défaut
	return map[string]interface{}{
		"used_count":     0,
		"usage_limit":    0,
		"total_discount": 0.0,
	}, nil
}

// validatePromoValues valide les valeurs d'un code promotionnel
func (s *PromoCodesService) validatePromoValues(promo *models.Promo) error {
	if promo.Value <= 0 {
		return fmt.Errorf("la valeur du promo doit être positive")
	}

	if promo.Type == models.PromoTypePercentage && promo.Value > 100 {
		return fmt.Errorf("le pourcentage ne peut pas dépasser 100%%")
	}

	if promo.MinPurchaseAmount != nil && *promo.MinPurchaseAmount < 0 {
		return fmt.Errorf("le montant minimum d'achat ne peut pas être négatif")
	}

	if promo.MaxUsage != nil && *promo.MaxUsage < 0 {
		return fmt.Errorf("la limite d'utilisation ne peut pas être négative")
	}

	if promo.EndDate.Before(promo.StartDate) {
		return fmt.Errorf("la date de fin ne peut pas être antérieure à la date de début")
	}

	return nil
}

// promoCodeExists vérifie si un code promotionnel existe déjà
func (s *PromoCodesService) promoCodeExists(code string) (bool, error) {
	ctx := context.Background()
	
	promotion, err := db.PrismaDB.Promo.FindFirst(
		prismadb.Promo.Code.Equals(strings.ToUpper(strings.TrimSpace(code))),
	).Exec(ctx)
	
	if err != nil {
		if err == prismadb.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	
	return promotion != nil, nil
}

// savePromo sauvegarde un code promotionnel en base
func (s *PromoCodesService) savePromo(promo *models.Promo) error {
	ctx := context.Background()
	
	// TODO: Simplification temporaire - implémenter les champs optionnels plus tard
	
	_, err := db.PrismaDB.Promo.CreateOne(
		prismadb.Promo.Code.Set(promo.Code),
		prismadb.Promo.Type.Set(prismadb.PromoType(promo.Type)),
		prismadb.Promo.Value.Set(promo.Value),
		prismadb.Promo.ValidFrom.Set(promo.StartDate),
		prismadb.Promo.ValidUntil.Set(promo.EndDate),
	).Exec(ctx)
	return err
}

// getPromoByID récupère un code promotionnel par son ID
func (s *PromoCodesService) getPromoByID(promoID string) (*models.Promo, error) {
	ctx := context.Background()
	
	promotion, err := db.PrismaDB.Promo.FindUnique(
		prismadb.Promo.ID.Equals(promoID),
	).Exec(ctx)

	if err != nil {
		if err == prismadb.ErrNotFound {
			return nil, fmt.Errorf("code promotionnel non trouvé")
		}
		return nil, fmt.Errorf("erreur de récupération: %v", err)
	}

	return s.convertPrismaPromoToModel(promotion), nil
}

// updatePromo met à jour un code promotionnel en base
func (s *PromoCodesService) updatePromo(promo *models.Promo) error {
	ctx := context.Background()
	
	_, err := db.PrismaDB.Promo.FindUnique(
		prismadb.Promo.ID.Equals(promo.ID),
	).Update(
		prismadb.Promo.Value.Set(promo.Value),
		prismadb.Promo.ValidFrom.Set(promo.StartDate),
		prismadb.Promo.ValidUntil.Set(promo.EndDate),
		prismadb.Promo.IsActive.Set(promo.IsActive),
		prismadb.Promo.UpdatedAt.Set(promo.UpdatedAt),
	).Exec(ctx)
	
	return err
}

// convertPrismaPromoToModel convertit un modèle Prisma en modèle de domaine
func (s *PromoCodesService) convertPrismaPromoToModel(prismaPromo *prismadb.PromoModel) *models.Promo {
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

// GetAllPromotions récupère toutes les promotions avec pagination et filtres (admin seulement)
func (s *PromoCodesService) GetAllPromotions(page int, limit int, filters map[string]string) ([]*models.Promo, int, error) {
	ctx := context.Background()

	// Calculer offset
	offset := (page - 1) * limit

	// Construire les conditions de filtre
	var conditions []prismadb.PromoWhereParam

	// Appliquer les filtres
	if status, exists := filters["status"]; exists && status != "" {
		if status == "active" {
			conditions = append(conditions, prismadb.Promo.IsActive.Equals(true))
		} else if status == "inactive" {
			conditions = append(conditions, prismadb.Promo.IsActive.Equals(false))
		}
	}

	if promoType, exists := filters["type"]; exists && promoType != "" {
		switch promoType {
		case "PERCENTAGE":
			conditions = append(conditions, prismadb.Promo.Type.Equals(prismadb.PromoTypePercentage))
		case "FIXED_AMOUNT":
			conditions = append(conditions, prismadb.Promo.Type.Equals(prismadb.PromoTypeFixedAmount))
		case "FREE_DELIVERY":
			conditions = append(conditions, prismadb.Promo.Type.Equals(prismadb.PromoTypeFreeDelivery))
		}
	}

	// Récupérer les promotions avec pagination
	promos, err := db.PrismaDB.Promo.FindMany(
		conditions...,
	).Skip(offset).Take(limit).OrderBy(
		prismadb.Promo.CreatedAt.Order(prismadb.SortOrderDesc),
	).Exec(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Compter le total avec les mêmes filtres
	allPromos, err := db.PrismaDB.Promo.FindMany(
		conditions...,
	).Exec(ctx)
	if err != nil {
		return nil, 0, err
	}
	total := len(allPromos)

	// Convertir les promotions
	promoModels := make([]*models.Promo, len(promos))
	for i, promo := range promos {
		promoModels[i] = s.convertPrismaPromoToModel(&promo)
	}

	return promoModels, total, nil
}

// GetAllPromotionStats récupère les statistiques de toutes les promotions (admin seulement)
func (s *PromoCodesService) GetAllPromotionStats() (map[string]interface{}, error) {
	ctx := context.Background()

	// Compter le total des promotions
	allPromos, err := db.PrismaDB.Promo.FindMany().Exec(ctx)
	if err != nil {
		return nil, err
	}

	// Compter les promotions actives
	activePromos, err := db.PrismaDB.Promo.FindMany(
		prismadb.Promo.IsActive.Equals(true),
	).Exec(ctx)
	if err != nil {
		return nil, err
	}

	// Compter les promotions par type
	percentagePromos, _ := db.PrismaDB.Promo.FindMany(
		prismadb.Promo.Type.Equals(prismadb.PromoTypePercentage),
	).Exec(ctx)
	fixedAmountPromos, _ := db.PrismaDB.Promo.FindMany(
		prismadb.Promo.Type.Equals(prismadb.PromoTypeFixedAmount),
	).Exec(ctx)
	freeDeliveryPromos, _ := db.PrismaDB.Promo.FindMany(
		prismadb.Promo.Type.Equals(prismadb.PromoTypeFreeDelivery),
	).Exec(ctx)

	// Calculer les statistiques d'utilisation
	var totalUsage int
	var totalSavings float64
	for _, promo := range allPromos {
		totalUsage += promo.UsageCount
		// TODO: Calculer les économies totales en fonction du type et de la valeur
	}

	stats := map[string]interface{}{
		"total": map[string]interface{}{
			"count":   len(allPromos),
			"active":  len(activePromos),
			"inactive": len(allPromos) - len(activePromos),
		},
		"byType": map[string]interface{}{
			"percentage":   len(percentagePromos),
			"fixedAmount":  len(fixedAmountPromos),
			"freeDelivery": len(freeDeliveryPromos),
		},
		"usage": map[string]interface{}{
			"totalUsage":   totalUsage,
			"totalSavings": totalSavings,
			"averageUsage": func() float64 {
				if len(allPromos) > 0 {
					return float64(totalUsage) / float64(len(allPromos))
				}
				return 0
			}(),
		},
	}

	return stats, nil
}
