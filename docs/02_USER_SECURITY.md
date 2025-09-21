## Utilisateurs, Sécurité et Authentification (Prisma/PostgreSQL)

### Portée
- Models Prisma: `User`, `OTP`, `RefreshToken`, `EmergencyContact`, `PaymentMethod`, `UserAddress`, `Wallet`, `WalletTransaction`
- Contraintes PostgreSQL, index et relations Prisma
- Exemples Go avec Prisma Client: CRUD, requêtes, services
- Architecture backend sécurisée avec JWT et middleware

---

### Modèles Prisma (extraits clés)

```prisma
// User
model User {
  id                   String  @id @default(uuid())
  phone               String  @unique
  role                Role    @default(CLIENT)
  firstName           String  @default("")
  lastName            String  @default("")
  email               String?
  isProfileCompleted  Boolean @default(false)
  createdAt           DateTime @default(now())
  updatedAt           DateTime @updatedAt
  
  // Relations
  refreshTokens       RefreshToken[]
  emergencyContacts   EmergencyContact[]
  paymentMethods      PaymentMethod[]
  userAddresses       UserAddress[]
  wallet              Wallet?
  
  @@map("users")
}

enum Role {
  CLIENT
  LIVREUR
  ADMIN
  GESTIONNAIRE
  MARKETING
}

// OTP
model OTP {
  id        String   @id @default(uuid())
  phone     String
  code      String
  expiresAt DateTime
  createdAt DateTime @default(now())
  
  @@map("otps")
}

// RefreshToken
model RefreshToken {
  id         String   @id @default(uuid())
  userId     String
  tokenValue String
  expiresAt  DateTime
  revoked    Boolean  @default(false)
  createdAt  DateTime @default(now())
  
  user User @relation(fields: [userId], references: [id])
  
  @@map("refresh_tokens")
}

// EmergencyContact
model EmergencyContact {
  id     String @id @default(uuid())
  userId String
  phone  String
  
  user User @relation(fields: [userId], references: [id])
  
  @@map("emergency_contacts")
}

// PaymentMethod
model PaymentMethod {
  id        String      @id @default(uuid())
  userId    String
  type      PaymentType
  isDefault Boolean     @default(false)
  
  user User @relation(fields: [userId], references: [id])
  
  @@map("payment_methods")
}

enum PaymentType {
  CASH
  WAVE
  MOBILE_MONEY_ORANGE
  MOBILE_MONEY_MTN
  MOBILE_MONEY_MOOV
}

// UserAddress
model UserAddress {
  id        String  @id @default(uuid())
  userId    String
  address   String
  lat       Float?
  lng       Float?
  isDefault Boolean @default(false)
  
  user User @relation(fields: [userId], references: [id])
  
  @@map("user_addresses")
}

// Wallet & WalletTransaction
model Wallet {
  id      String  @id @default(uuid())
  userId  String  @unique
  balance Float   @default(0.0)
  
  user         User                @relation(fields: [userId], references: [id])
  transactions WalletTransaction[]
  
  @@map("wallets")
}

model WalletTransaction {
  id          String            @id @default(uuid())
  walletId    String
  amount      Float
  type        TransactionType
  description String?
  createdAt   DateTime          @default(now())
  
  wallet Wallet @relation(fields: [walletId], references: [id])
  
  @@map("wallet_transactions")
}

enum TransactionType {
  CREDIT
  DEBIT
}
```

---

### Services Go avec Prisma Client

#### Créer un utilisateur (inscription par téléphone)
```go
func (s *UserService) CreateUser(phone, firstName, lastName string) (*models.User, error) {
    user, err := s.db.User.CreateOne(
        db.User.Phone.Set(phone),
        db.User.Role.Set(db.RoleCLIENT),
        db.User.FirstName.Set(firstName),
        db.User.LastName.Set(lastName),
    ).Exec(context.Background())
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaUserToModel(user), nil
}
```

#### Lire un utilisateur par téléphone
```go
func (s *UserService) GetUserByPhone(phone string) (*models.User, error) {
    user, err := s.db.User.FindFirst(
        db.User.Phone.Equals(phone),
    ).Exec(context.Background())
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaUserToModel(user), nil
}
```

#### Mettre à jour le profil
```go
func (s *UserService) UpdateProfile(userID, firstName, lastName, email string) error {
    _, err := s.db.User.FindUnique(
        db.User.ID.Equals(userID),
    ).Update(
        db.User.FirstName.Set(firstName),
        db.User.LastName.Set(lastName),
        db.User.Email.Set(email),
        db.User.IsProfileCompleted.Set(true),
    ).Exec(context.Background())
    
    return err
}
```

#### Supprimer un utilisateur (soft delete)
```go
func (s *UserService) SoftDeleteUser(userID string) error {
    _, err := s.db.User.FindUnique(
        db.User.ID.Equals(userID),
    ).Update(
        db.User.IsDeleted.Set(true), // Si ajouté au schéma
    ).Exec(context.Background())
    
    return err
}
```

---

### OTP - Flux recommandé (backend Go)

1) Frontend demande OTP:
```http
POST /auth/send-otp { phone }
```
2) Backend Go: génère OTP et l'enregistre avec Prisma
3) Frontend vérifie OTP:
```http
POST /auth/verify-otp { phone, code }
```
4) Backend: vérifie, crée `RefreshToken`, renvoie JWT + profil

#### Service OTP en Go:
```go
func (s *OTPService) GenerateOTP(phone string) (string, error) {
    code := generateRandomCode(6) // Utils
    expiresAt := time.Now().Add(5 * time.Minute)
    
    _, err := s.db.OTP.CreateOne(
        db.OTP.Phone.Set(phone),
        db.OTP.Code.Set(code),
        db.OTP.ExpiresAt.Set(expiresAt),
    ).Exec(context.Background())
    
    return code, err
}

func (s *OTPService) VerifyOTP(phone, code string) (bool, error) {
    otp, err := s.db.OTP.FindFirst(
        db.OTP.Phone.Equals(phone),
        db.OTP.Code.Equals(code),
        db.OTP.ExpiresAt.Gt(time.Now()),
    ).Exec(context.Background())
    
    if err != nil {
        return false, err
    }
    
    return otp != nil, nil
}
```

---

### Adresses utilisateur

```go
// Ajouter une adresse
func (s *UserService) AddAddress(userID, address string, lat, lng *float64, isDefault bool) error {
    _, err := s.db.UserAddress.CreateOne(
        db.UserAddress.UserID.Set(userID),
        db.UserAddress.Address.Set(address),
        db.UserAddress.Lat.SetIfPresent(lat),
        db.UserAddress.Lng.SetIfPresent(lng),
        db.UserAddress.IsDefault.Set(isDefault),
    ).Exec(context.Background())
    
    return err
}

// Lister les adresses
func (s *UserService) GetUserAddresses(userID string) ([]*models.UserAddress, error) {
    addresses, err := s.db.UserAddress.FindMany(
        db.UserAddress.UserID.Equals(userID),
    ).OrderBy(
        db.UserAddress.CreatedAt.Order(db.DESC),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaUserAddresses(addresses), nil
}

// Mettre à jour une adresse
func (s *UserService) UpdateAddress(addressID, address string, isDefault bool) error {
    _, err := s.db.UserAddress.FindUnique(
        db.UserAddress.ID.Equals(addressID),
    ).Update(
        db.UserAddress.Address.Set(address),
        db.UserAddress.IsDefault.Set(isDefault),
    ).Exec(context.Background())
    
    return err
}
```

---

### Méthodes de paiement

```go
// Ajouter une méthode
func (s *UserService) AddPaymentMethod(userID string, paymentType db.PaymentType, isDefault bool) error {
    _, err := s.db.PaymentMethod.CreateOne(
        db.PaymentMethod.UserID.Set(userID),
        db.PaymentMethod.Type.Set(paymentType),
        db.PaymentMethod.IsDefault.Set(isDefault),
    ).Exec(context.Background())
    
    return err
}

// Lister les méthodes
func (s *UserService) GetPaymentMethods(userID string) ([]*models.PaymentMethod, error) {
    methods, err := s.db.PaymentMethod.FindMany(
        db.PaymentMethod.UserID.Equals(userID),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaPaymentMethods(methods), nil
}

// Basculer la méthode par défaut
func (s *UserService) SetDefaultPaymentMethod(userID, methodID string) error {
    // Désactiver toutes les méthodes existantes
    _, err := s.db.PaymentMethod.FindMany(
        db.PaymentMethod.UserID.Equals(userID),
    ).Update(
        db.PaymentMethod.IsDefault.Set(false),
    ).Exec(context.Background())
    
    if err != nil {
        return err
    }
    
    // Activer la nouvelle méthode par défaut
    _, err = s.db.PaymentMethod.FindUnique(
        db.PaymentMethod.ID.Equals(methodID),
    ).Update(
        db.PaymentMethod.IsDefault.Set(true),
    ).Exec(context.Background())
    
    return err
}
```

---

### Wallet & Transactions

```go
// Obtenir ou créer le wallet
func (s *WalletService) GetOrCreateWallet(userID string) (*models.Wallet, error) {
    wallet, err := s.db.Wallet.FindUnique(
        db.Wallet.UserID.Equals(userID),
    ).Exec(context.Background())
    
    if err != nil {
        // Créer si n'existe pas
        wallet, err = s.db.Wallet.CreateOne(
            db.Wallet.UserID.Set(userID),
            db.Wallet.Balance.Set(0.0),
        ).Exec(context.Background())
        
        if err != nil {
            return nil, err
        }
    }
    
    return ConvertPrismaWalletToModel(wallet), nil
}

// Créditer le wallet
func (s *WalletService) CreditWallet(userID string, amount float64, description string) error {
    return s.db.Prisma.Transaction(func(tx *db.PrismaClient) error {
        // Mettre à jour le solde
        _, err := tx.Wallet.FindUnique(
            db.Wallet.UserID.Equals(userID),
        ).Update(
            db.Wallet.Balance.Increment(amount),
        ).Exec(context.Background())
        
        if err != nil {
            return err
        }
        
        // Créer la transaction
        wallet, err := tx.Wallet.FindUnique(
            db.Wallet.UserID.Equals(userID),
        ).Exec(context.Background())
        
        if err != nil {
            return err
        }
        
        _, err = tx.WalletTransaction.CreateOne(
            db.WalletTransaction.WalletID.Set(wallet.ID),
            db.WalletTransaction.Amount.Set(amount),
            db.WalletTransaction.Type.Set(db.TransactionTypeCREDIT),
            db.WalletTransaction.Description.SetIfPresent(&description),
        ).Exec(context.Background())
        
        return err
    })
}

// Débiter le wallet (avec vérification)
func (s *WalletService) DebitWallet(userID string, amount float64, description string) error {
    return s.db.Prisma.Transaction(func(tx *db.PrismaClient) error {
        wallet, err := tx.Wallet.FindUnique(
            db.Wallet.UserID.Equals(userID),
        ).Exec(context.Background())
        
        if err != nil {
            return err
        }
        
        if wallet.Balance < amount {
            return errors.New("INSUFFICIENT_FUNDS")
        }
        
        // Débiter le solde
        _, err = tx.Wallet.FindUnique(
            db.Wallet.UserID.Equals(userID),
        ).Update(
            db.Wallet.Balance.Decrement(amount),
        ).Exec(context.Background())
        
        if err != nil {
            return err
        }
        
        // Créer la transaction
        _, err = tx.WalletTransaction.CreateOne(
            db.WalletTransaction.WalletID.Set(wallet.ID),
            db.WalletTransaction.Amount.Set(-amount),
            db.WalletTransaction.Type.Set(db.TransactionTypeDEBIT),
            db.WalletTransaction.Description.SetIfPresent(&description),
        ).Exec(context.Background())
        
        return err
    })
}
```

---

### Authentification JWT et Middleware

```go
// JWT Service
type JWTService struct {
    secretKey string
    db        *db.PrismaClient
}

func (s *JWTService) GenerateTokens(userID string) (string, string, error) {
    // Access Token (15 min)
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Minute * 15).Unix(),
        "type":    "access",
    })
    
    // Refresh Token (7 jours)
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": userID,
        "exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
        "type":    "refresh",
    })
    
    accessTokenString, err := accessToken.SignedString([]byte(s.secretKey))
    if err != nil {
        return "", "", err
    }
    
    refreshTokenString, err := refreshToken.SignedString([]byte(s.secretKey))
    if err != nil {
        return "", "", err
    }
    
    // Sauvegarder le refresh token en DB
    _, err = s.db.RefreshToken.CreateOne(
        db.RefreshToken.UserID.Set(userID),
        db.RefreshToken.TokenValue.Set(refreshTokenString),
        db.RefreshToken.ExpiresAt.Set(time.Now().Add(time.Hour * 24 * 7)),
    ).Exec(context.Background())
    
    return accessTokenString, refreshTokenString, err
}

// Middleware d'authentification
func (s *JWTService) AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(401, gin.H{"error": "No token provided"})
            c.Abort()
            return
        }
        
        tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
        
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(s.secretKey), nil
        })
        
        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        claims := token.Claims.(jwt.MapClaims)
        c.Set("user_id", claims["user_id"])
        c.Next()
    }
}
```

---

### Sécurité et Permissions (Middleware Go)

Contrairement à SurrealDB, avec Prisma/PostgreSQL, les permissions sont gérées via des middlewares Go:

```go
// Middleware de vérification des rôles
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if userID == "" {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        user, err := userService.GetUserByID(userID)
        if err != nil {
            c.JSON(401, gin.H{"error": "User not found"})
            c.Abort()
            return
        }
        
        allowed := false
        for _, role := range allowedRoles {
            if user.Role == role {
                allowed = true
                break
            }
        }
        
        if !allowed {
            c.JSON(403, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }
        
        c.Set("user", user)
        c.Next()
    }
}

// Middleware de vérification de propriété
func RequireOwnershipOrAdmin(resourceUserIDParam string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        user := c.MustGet("user").(*models.User)
        resourceUserID := c.Param(resourceUserIDParam)
        
        if user.Role != "ADMIN" && userID != resourceUserID {
            c.JSON(403, gin.H{"error": "Access denied"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

---

### Index PostgreSQL et performances

Index automatiques avec Prisma:
- `users.phone` UNIQUE pour login rapide
- `users.role` pour filtres par rôle
- `user_addresses.user_id` pour chargement profil
- `payment_methods.user_id` pour méthodes de paiement
- `wallets.user_id` UNIQUE pour accès direct
- `wallet_transactions.wallet_id, created_at` pour historiques

---

### Bonnes pratiques backend Go/Prisma

- **Sécurité**: Toujours valider les permissions dans les handlers
- **Transactions**: Utiliser `db.Prisma.Transaction()` pour opérations atomiques
- **Validation**: Valider les entrées avec des structs et tags de validation
- **Erreurs**: Retourner des erreurs explicites sans exposer les détails internes
- **Logs**: Logger toutes les opérations sensibles (OTP, auth, wallet)
- **Conversion**: Utiliser des fonctions de conversion entre modèles Prisma et domain
- **Cache**: Considérer Redis pour les sessions et données fréquemment accédées


