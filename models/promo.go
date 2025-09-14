package models

import (
	"time"
)

// PromoType defines the promotion type enumeration
type PromoType string

const (
	PromoTypePercentage   PromoType = "PERCENTAGE"
	PromoTypeFixedAmount  PromoType = "FIXED_AMOUNT"
	PromoTypeFreeDelivery PromoType = "FREE_DELIVERY"
)

// ReferralStatus defines the referral status enumeration
type ReferralStatus string

const (
	ReferralStatusPending      ReferralStatus = "PENDING"
	ReferralStatusCompleted    ReferralStatus = "COMPLETED"
	ReferralStatusExpired      ReferralStatus = "EXPIRED"
	ReferralStatusCancelled    ReferralStatus = "CANCELLED"
	ReferralStatusRewardClaimed ReferralStatus = "REWARD_CLAIMED"
)

// Promo represents a promotional code
type Promo struct {
	ID                  string     `json:"id"`
	Code                string     `json:"code" validate:"required,min=3,max=20"`
	Description         *string    `json:"description,omitempty"`
	Type                PromoType  `json:"type" validate:"required"`
	Value               float64    `json:"value" validate:"gte=0"`
	StartDate           time.Time  `json:"startDate" validate:"required"`
	EndDate             time.Time  `json:"endDate" validate:"required"`
	IsActive            bool       `json:"isActive"`
	MaxUsage            *int       `json:"maxUsage,omitempty"`
	UsageCount          *int       `json:"usageCount,omitempty"`
	MinPurchaseAmount   *float64   `json:"minPurchaseAmount,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

// CreatePromoRequest represents request for creating a promo
type CreatePromoRequest struct {
	Code              string    `json:"code" validate:"required,min=3,max=20"`
	Description       *string   `json:"description,omitempty"`
	Type              PromoType `json:"type" validate:"required"`
	Value             float64   `json:"value" validate:"gte=0"`
	StartDate         time.Time `json:"startDate" validate:"required"`
	EndDate           time.Time `json:"endDate" validate:"required"`
	MaxUsage          *int      `json:"maxUsage,omitempty"`
	MinPurchaseAmount *float64  `json:"minPurchaseAmount,omitempty"`
}

// UpdatePromoRequest represents request for updating a promo
type UpdatePromoRequest struct {
	Description       *string   `json:"description,omitempty"`
	Value             *float64  `json:"value,omitempty" validate:"omitempty,gte=0"`
	StartDate         *time.Time `json:"startDate,omitempty"`
	EndDate           *time.Time `json:"endDate,omitempty"`
	IsActive          *bool     `json:"isActive,omitempty"`
	MaxUsage          *int      `json:"maxUsage,omitempty"`
	MinPurchaseAmount *float64  `json:"minPurchaseAmount,omitempty"`
}

// ApplyPromoRequest represents request for applying a promo
type ApplyPromoRequest struct {
	Code   string  `json:"code" validate:"required"`
	Amount float64 `json:"amount" validate:"gte=0"`
}

// PromoUsage represents usage of a promotional code
type PromoUsage struct {
	ID       string    `json:"id"`
	PromoID  string    `json:"promoId" validate:"required"`
	UserID   string    `json:"userId" validate:"required"`
	UsedAt   time.Time `json:"usedAt"`
	Amount   float64   `json:"amount" validate:"gte=0"`
	Discount float64   `json:"discount" validate:"gte=0"`
}

// Referral represents a referral
type Referral struct {
	ID              string          `json:"id"`
	ReferrerID      string          `json:"referrerId" validate:"required"`
	RefereePhone    string          `json:"refereePhone" validate:"required,min=8,max=15"`
	RefereeID       *string         `json:"refereeId,omitempty"`
	Code            string          `json:"code" validate:"required"`
	Message         string          `json:"message"`
	Status          ReferralStatus  `json:"status"`
	CreatedAt       time.Time       `json:"createdAt"`
	CompletedAt     *time.Time      `json:"completedAt,omitempty"`
	ExpiresAt       *time.Time      `json:"expiresAt,omitempty"`
	RewardClaimedAt *time.Time      `json:"rewardClaimedAt,omitempty"`
}

// CreateReferralRequest represents request for creating a referral
type CreateReferralRequest struct {
	RefereePhone string  `json:"refereePhone" validate:"required,min=8,max=15"`
	Message      *string `json:"message,omitempty"`
}

// ReferralResponse represents referral data in response
type ReferralResponse struct {
	ID              string          `json:"id"`
	ReferrerID      string          `json:"referrerId"`
	Referrer        *UserResponse   `json:"referrer,omitempty"`
	RefereePhone    string          `json:"refereePhone"`
	RefereeID       *string         `json:"refereeId,omitempty"`
	Referee         *UserResponse   `json:"referee,omitempty"`
	Code            string          `json:"code"`
	Message         string          `json:"message"`
	Status          ReferralStatus  `json:"status"`
	CreatedAt       time.Time       `json:"createdAt"`
	CompletedAt     *time.Time      `json:"completedAt,omitempty"`
	ExpiresAt       *time.Time      `json:"expiresAt,omitempty"`
	RewardClaimedAt *time.Time      `json:"rewardClaimedAt,omitempty"`
	RewardAmount    float64         `json:"rewardAmount"`
}

// PricingRule represents pricing rules for different vehicle types
type PricingRule struct {
	ID           string      `json:"id"`
	VehicleType  VehicleType `json:"vehicleType" validate:"required"`
	BasePrice    float64     `json:"basePrice" validate:"gte=0"`
	IncludedKm   float64     `json:"includedKm" validate:"gte=0"`
	PerKm        float64     `json:"perKm" validate:"gte=0"`
	WaitingFree  int         `json:"waitingFree" validate:"gte=0"` // Free waiting minutes
	WaitingRate  float64     `json:"waitingRate" validate:"gte=0"` // Price per minute after free
}

// CreatePricingRuleRequest represents request for creating a pricing rule
type CreatePricingRuleRequest struct {
	VehicleType VehicleType `json:"vehicleType" validate:"required"`
	BasePrice   float64     `json:"basePrice" validate:"gte=0"`
	IncludedKm  float64     `json:"includedKm" validate:"gte=0"`
	PerKm       float64     `json:"perKm" validate:"gte=0"`
	WaitingFree int         `json:"waitingFree" validate:"gte=0"`
	WaitingRate float64     `json:"waitingRate" validate:"gte=0"`
}

// PriceCalculation represents the result of price calculation
type PriceCalculation struct {
	BasePrice     float64 `json:"basePrice"`
	DistancePrice float64 `json:"distancePrice"`
	WaitingPrice  float64 `json:"waitingPrice"`
	SubTotal      float64 `json:"subTotal"`
	PromoDiscount float64 `json:"promoDiscount"`
	FinalPrice    float64 `json:"finalPrice"`
	PromoCode     *string `json:"promoCode,omitempty"`
}

// PromoValidationResult represents the result of promo validation
type PromoValidationResult struct {
	Valid         bool    `json:"valid"`
	Discount      float64 `json:"discount"`
	FinalPrice    float64 `json:"finalPrice"`
	Message       string  `json:"message"`
	DiscountType  string  `json:"discountType"` // "percentage", "fixed", "free_delivery"
}

// IsValidType checks if the promo type is valid
func (pt PromoType) IsValid() bool {
	return pt == PromoTypePercentage || pt == PromoTypeFixedAmount || pt == PromoTypeFreeDelivery
}

// IsValidReferralStatus checks if the referral status is valid
func (rs ReferralStatus) IsValid() bool {
	validStatuses := []ReferralStatus{
		ReferralStatusPending, ReferralStatusCompleted, ReferralStatusExpired,
		ReferralStatusCancelled, ReferralStatusRewardClaimed,
	}
	
	for _, status := range validStatuses {
		if rs == status {
			return true
		}
	}
	return false
}

// IsExpired checks if promo is expired
func (p *Promo) IsExpired() bool {
	return time.Now().After(p.EndDate)
}

// IsActive checks if promo is active and not expired
func (p *Promo) IsActive() bool {
	now := time.Now()
	return p.IsActive && now.After(p.StartDate) && now.Before(p.EndDate)
}

// HasReachedMaxUsage checks if promo has reached max usage
func (p *Promo) HasReachedMaxUsage() bool {
	if p.MaxUsage == nil {
		return false
	}
	if p.UsageCount == nil {
		return false
	}
	return *p.UsageCount >= *p.MaxUsage
}

// CanBeUsed checks if promo can be used for given amount
func (p *Promo) CanBeUsed(amount float64) bool {
	if !p.IsActive() || p.HasReachedMaxUsage() {
		return false
	}
	
	if p.MinPurchaseAmount != nil && amount < *p.MinPurchaseAmount {
		return false
	}
	
	return true
}

// CalculateDiscount calculates discount for given amount
func (p *Promo) CalculateDiscount(amount float64) float64 {
	if !p.CanBeUsed(amount) {
		return 0
	}
	
	switch p.Type {
	case PromoTypePercentage:
		return amount * (p.Value / 100)
	case PromoTypeFixedAmount:
		if p.Value > amount {
			return amount // Can't discount more than the total
		}
		return p.Value
	case PromoTypeFreeDelivery:
		return amount // Free delivery = 100% discount
	default:
		return 0
	}
}

// IsExpired checks if referral is expired
func (r *Referral) IsExpired() bool {
	if r.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*r.ExpiresAt)
}

// CanBeClaimed checks if referral reward can be claimed
func (r *Referral) CanBeClaimed() bool {
	return r.Status == ReferralStatusCompleted && r.RewardClaimedAt == nil
}

// IsCompleted checks if referral is completed
func (r *Referral) IsCompleted() bool {
	return r.Status == ReferralStatusCompleted
}

// CalculatePrice calculates delivery price based on pricing rules
func (pr *PricingRule) CalculatePrice(distanceKm, waitingMin float64) PriceCalculation {
	calc := PriceCalculation{
		BasePrice: pr.BasePrice,
	}
	
	// Calculate distance price
	if distanceKm > pr.IncludedKm {
		extraKm := distanceKm - pr.IncludedKm
		calc.DistancePrice = extraKm * pr.PerKm
	}
	
	// Calculate waiting price
	if waitingMin > float64(pr.WaitingFree) {
		extraMin := waitingMin - float64(pr.WaitingFree)
		calc.WaitingPrice = extraMin * pr.WaitingRate
	}
	
	calc.SubTotal = calc.BasePrice + calc.DistancePrice + calc.WaitingPrice
	calc.FinalPrice = calc.SubTotal - calc.PromoDiscount
	
	if calc.FinalPrice < 0 {
		calc.FinalPrice = 0
	}
	
	return calc
}

// ToResponse converts Referral to ReferralResponse
func (r *Referral) ToResponse() *ReferralResponse {
	return &ReferralResponse{
		ID:              r.ID,
		ReferrerID:      r.ReferrerID,
		RefereePhone:    r.RefereePhone,
		RefereeID:       r.RefereeID,
		Code:            r.Code,
		Message:         r.Message,
		Status:          r.Status,
		CreatedAt:       r.CreatedAt,
		CompletedAt:     r.CompletedAt,
		ExpiresAt:       r.ExpiresAt,
		RewardClaimedAt: r.RewardClaimedAt,
		RewardAmount:    1000.0, // From config
	}
}