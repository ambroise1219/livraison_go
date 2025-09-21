## Promos & Parrainage (Prisma/PostgreSQL)

### Models couverts
- `Promo`, `PromoUsage`, `Referral`

---

### Modèles Prisma

```prisma
// Promo
model Promo {
  id        String     @id @default(uuid())
  code      String     @unique
  type      PromoType
  value     Float
  startDate DateTime
  endDate   DateTime
  isActive  Boolean    @default(true)
  createdAt DateTime   @default(now())
  
  // Relations
  usages PromoUsage[]
  
  @@map("promos")
}

enum PromoType {
  PERCENTAGE
  FIXED_AMOUNT
  FREE_DELIVERY
}

// PromoUsage
model PromoUsage {
  id        String   @id @default(uuid())
  promoId   String
  userId    String
  amount    Float
  discount  Float
  usedAt    DateTime @default(now())
  
  promo Promo @relation(fields: [promoId], references: [id])
  user  User  @relation(fields: [userId], references: [id])
  
  @@map("promo_usages")
}

// Referral
model Referral {
  id            String        @id @default(uuid())
  referrerId    String
  refereePhone  String
  code          String        @unique
  status        ReferralStatus @default(PENDING)
  createdAt     DateTime      @default(now())
  completedAt   DateTime?
  
  referrer User @relation(fields: [referrerId], references: [id])
  
  @@map("referrals")
}

enum ReferralStatus {
  PENDING
  COMPLETED
  EXPIRED
  CANCELLED
  REWARD_CLAIMED
}
```

---

### Services Go - Appliquer un code promo

```go
type PromoService struct {
    db *db.PrismaClient
}

type ApplyPromoResult struct {
    OriginalAmount float64 `json:"original_amount"`
    Discount      float64 `json:"discount"`
    FinalAmount   float64 `json:"final_amount"`
    PromoID       string  `json:"promo_id"`
}

func (s *PromoService) ApplyPromo(code, userID string, amount float64) (*ApplyPromoResult, error) {
    // 1) Charger et valider la promo
    promo, err := s.db.Promo.FindFirst(
        db.Promo.Code.Equals(code),
        db.Promo.IsActive.Equals(true),
        db.Promo.StartDate.Lte(time.Now()),
        db.Promo.EndDate.Gte(time.Now()),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, errors.New("INVALID_PROMO")
    }
    
    // 2) Calculer la remise
    var discount float64
    switch promo.Type {
    case db.PromoTypePERCENTAGE:
        discount = amount * promo.Value / 100
    case db.PromoTypeFIXED_AMOUNT:
        discount = promo.Value
    case db.PromoTypeFREE_DELIVERY:
        discount = amount // Livraison gratuite
    default:
        discount = 0
    }
    
    finalAmount := math.Max(0, amount-discount)
    
    // 3) Enregistrer l'usage
    _, err = s.db.PromoUsage.CreateOne(
        db.PromoUsage.PromoID.Set(promo.ID),
        db.PromoUsage.UserID.Set(userID),
        db.PromoUsage.Amount.Set(amount),
        db.PromoUsage.Discount.Set(discount),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return &ApplyPromoResult{
        OriginalAmount: amount,
        Discount:      discount,
        FinalAmount:   finalAmount,
        PromoID:       promo.ID,
    }, nil
}

func (s *PromoService) ValidatePromo(code string) (*models.Promo, error) {
    promo, err := s.db.Promo.FindFirst(
        db.Promo.Code.Equals(code),
        db.Promo.IsActive.Equals(true),
        db.Promo.StartDate.Lte(time.Now()),
        db.Promo.EndDate.Gte(time.Now()),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, errors.New("PROMO_NOT_FOUND_OR_EXPIRED")
    }
    
    return ConvertPrismaPromoToModel(promo), nil
}
```

---

### Services Go - Parrainage

```go
type ReferralService struct {
    db            *db.PrismaClient
    walletService *WalletService
}

// Créer une invitation
func (s *ReferralService) CreateReferral(referrerID, refereePhone string) (*models.Referral, error) {
    // Générer un code unique
    code := generateReferralCode(refereePhone)
    
    referral, err := s.db.Referral.CreateOne(
        db.Referral.ReferrerID.Set(referrerID),
        db.Referral.RefereePhone.Set(refereePhone),
        db.Referral.Code.Set(code),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaReferralToModel(referral), nil
}

// Marquer comme complété et récompenser
func (s *ReferralService) CompleteReferral(referralID string) error {
    return s.db.Prisma.Transaction(func(tx *db.PrismaClient) error {
        // Marquer comme complété
        referral, err := tx.Referral.FindUnique(
            db.Referral.ID.Equals(referralID),
        ).Update(
            db.Referral.Status.Set(db.ReferralStatusCOMPLETED),
            db.Referral.CompletedAt.Set(time.Now()),
        ).Exec(context.Background())
        
        if err != nil {
            return err
        }
        
        // Récompenser le parrain (ex: 1000 unités)
        return s.walletService.CreditWallet(
            referral.ReferrerID, 
            1000.0, 
            "Referral reward",
        )
    })
}

// Trouver referral par code
func (s *ReferralService) GetReferralByCode(code string) (*models.Referral, error) {
    referral, err := s.db.Referral.FindFirst(
        db.Referral.Code.Equals(code),
        db.Referral.Status.Equals(db.ReferralStatusPENDING),
    ).With(
        db.Referral.Referrer.Fetch(),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaReferralToModel(referral), nil
}

// Générer un code de parrainage
func generateReferralCode(phone string) string {
    // Simple: hash du téléphone + timestamp
    hash := sha256.Sum256([]byte(phone + strconv.FormatInt(time.Now().Unix(), 10)))
    return strings.ToUpper(hex.EncodeToString(hash[:])[:8])
}
```

---

### Middleware de permissions et index

```go
// Middleware pour les promos (lecture publique, admin pour écriture)
func RequireAdminForPromos() gin.HandlerFunc {
    return RequireRole("ADMIN", "MARKETING")
}

// Middleware pour les referrals
func RequireReferralAccess() gin.HandlerFunc {
    return func(c *gin.Context) {
        referralID := c.Param("referralId")
        userID := c.GetString("user_id")
        user := c.MustGet("user").(*models.User)
        
        if user.Role == "ADMIN" {
            c.Next()
            return
        }
        
        // Vérifier si l'utilisateur est le parrain
        referral, err := referralService.GetReferral(referralID)
        if err != nil || referral.ReferrerID != userID {
            c.JSON(403, gin.H{"error": "Access denied"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

**Index PostgreSQL:**
```sql
-- Index automatiques Prisma
CREATE UNIQUE INDEX promos_code_key ON promos(code);
CREATE INDEX promos_active_dates_idx ON promos(is_active, start_date, end_date);

CREATE INDEX promo_usages_promo_user_idx ON promo_usages(promo_id, user_id);
CREATE INDEX promo_usages_used_at_idx ON promo_usages(used_at);

CREATE UNIQUE INDEX referrals_code_key ON referrals(code);
CREATE INDEX referrals_referrer_status_idx ON referrals(referrer_id, status);
```


