# Migration Documentation SurrealDB → Prisma/PostgreSQL

## État des Fichiers

### ✅ Complétés (5/10)

1. **02_USER_SECURITY.md**
   - ✅ Modèles Prisma complets avec enums
   - ✅ Services Go avec Prisma Client
   - ✅ Authentification JWT et middleware
   - ✅ Permissions via middleware Go
   - ✅ Index PostgreSQL

2. **03_DELIVERIES_CORE.md**
   - ✅ Models Delivery, Package, Location, Tracking, DriverLocation
   - ✅ Services Go complets avec transactions
   - ✅ Middleware de permissions
   - ✅ Index PostgreSQL et optimisations

3. **06_TRANSPORT_INFRA.md**
   - ✅ Models Vehicle, Location, ServiceZone, Depot
   - ✅ Services Go avec support géospatial
   - ✅ Middleware de permissions
   - ✅ Index PostgreSQL

4. **08_PROMOS_AND_REFERRALS.md**
   - ✅ Models Promo, PromoUsage, Referral
   - ✅ Services Go avec calculs de remises
   - ✅ Logique de parrainage et récompenses
   - ✅ Middleware et index

5. **09_NOTIFICATIONS_AND_TEMPLATES.md**
   - ✅ Models Notification, NotificationTemplate, Banner
   - ✅ Services Go avec système de templating
   - ✅ Support multi-canal (SMS, Email, Push, WhatsApp)
   - ✅ Gestion des bannières actives

### 🔄 Restants à traiter (5/10)

#### 04_GROUPED_DELIVERIES.md
**Schémas à migrer :**
```prisma
model GroupedDelivery {
  id                  String @id @default(uuid())
  deliveryId          String @unique
  totalZones          Int
  completedZones      Int @default(0)
  discountPercentage  Float @default(30.0)
  originalPrice       Float
  finalPrice          Float
  
  delivery Delivery @relation(fields: [deliveryId], references: [id])
  zones    DeliveryZone[]
  
  @@map("grouped_deliveries")
}

model DeliveryZone {
  id                   String @id @default(uuid())
  groupedDeliveryId    String
  zoneNumber           Int
  recipientName        String
  recipientPhone       String
  pickupLocationId     String
  deliveryLocationId   String
  status               DeliveryStatus @default(PENDING)
  price                Float
  
  groupedDelivery      GroupedDelivery @relation(fields: [groupedDeliveryId], references: [id])
  pickupLocation       Location @relation("ZonePickup", fields: [pickupLocationId], references: [id])
  deliveryLocation     Location @relation("ZoneDelivery", fields: [deliveryLocationId], references: [id])
  
  @@unique([groupedDeliveryId, zoneNumber])
  @@map("delivery_zones")
}
```

#### 05_MOVING_SERVICES.md  
**Schémas à migrer :**
```prisma
model MovingService {
  id                  String @id @default(uuid())
  deliveryId          String @unique
  vehicleSize         MovingVehicleSize
  helpersCount        Int @default(1)
  floors              Int @default(1)
  hasElevator         Boolean @default(false)
  needsDisassembly    Boolean @default(false)
  hasFragileItems     Boolean @default(false)
  additionalServices  String[]
  specialInstructions String?
  estimatedVolume     Float?
  helpersCost         Float @default(0.0)
  vehicleCost         Float @default(0.0)
  serviceCost         Float @default(0.0)
  
  delivery Delivery @relation(fields: [deliveryId], references: [id])
  
  @@map("moving_services")
}

enum MovingVehicleSize {
  MINI
  SMALL
  MEDIUM
  LARGE
  EXTRA_LARGE
}
```

#### 07_PAYMENTS_AND_FEES.md
**Schémas à migrer :**
```prisma
model Payment {
  id          String @id @default(uuid())
  userId      String
  amount      Float
  method      PaymentMethod
  paymentType PaymentType
  status      PaymentStatus @default(PENDING)
  reference   String @unique
  createdAt   DateTime @default(now())
  
  user User @relation(fields: [userId], references: [id])
  
  @@map("payments")
}

model PlatformFee {
  id                String @id @default(uuid())
  deliveryId        String
  driverId          String
  clientId          String
  baseFee           Float
  commissionRate    Float
  commissionAmount  Float
  serviceFee        Float
  totalFee          Float
  driverEarnings    Float
  platformEarnings  Float
  createdAt         DateTime @default(now())
  
  delivery Delivery @relation(fields: [deliveryId], references: [id])
  driver   User @relation("DriverFees", fields: [driverId], references: [id])
  client   User @relation("ClientFees", fields: [clientId], references: [id])
  
  @@map("platform_fees")
}

model PricingRule {
  id           String @id @default(uuid())
  vehicleType  VehicleType @unique
  basePrice    Float
  includedKm   Float
  perKm        Float
  waitingFree  Int
  waitingRate  Float
  isActive     Boolean @default(true)
  
  @@map("pricing_rules")
}
```

#### 10_ANALYTICS_AND_QUALITY.md
**Schémas à migrer :**
```prisma
model Metrics {
  id                   String @id @default(uuid())
  driverId             String
  date                 DateTime @db.Date
  totalDeliveries      Int @default(0)
  completedDeliveries  Int @default(0)
  cancelledDeliveries  Int @default(0)
  averageRating        Float @default(0)
  totalEarnings        Float @default(0)
  
  driver User @relation(fields: [driverId], references: [id])
  
  @@unique([driverId, date])
  @@map("metrics")
}

model Incident {
  id         String @id @default(uuid())
  deliveryId String?
  driverId   String
  clientId   String
  type       String
  severity   String
  status     String @default("OPEN")
  description String
  createdAt  DateTime @default(now())
  
  delivery Delivery? @relation(fields: [deliveryId], references: [id])
  driver   User @relation("DriverIncidents", fields: [driverId], references: [id])
  client   User @relation("ClientIncidents", fields: [clientId], references: [id])
  
  @@map("incidents")
}

model Rating {
  id           String @id @default(uuid())
  deliveryId   String
  clientId     String
  driverId     String
  clientRating Int
  driverRating Int
  comment      String?
  createdAt    DateTime @default(now())
  
  delivery Delivery @relation(fields: [deliveryId], references: [id])
  client   User @relation("ClientRatings", fields: [clientId], references: [id])
  driver   User @relation("DriverRatings", fields: [driverId], references: [id])
  
  @@unique([deliveryId])
  @@map("ratings")
}
```

#### 11_DYNAMIC_CONFIG.md
**Schémas à migrer :**
```prisma
model PlatformConfig {
  id           String @id @default(uuid())
  configType   ConfigType
  vehicleType  VehicleType
  serviceZoneId String?
  value        Float
  percentage   Float
  isActive     Boolean @default(true)
  validFrom    DateTime @default(now())
  validTo      DateTime?
  
  @@map("platform_config")
}

model PeakHoursConfig {
  id            String @id @default(uuid())
  startTime     String // "18:00"
  endTime       String // "20:00" 
  days          String[] // ["MONDAY", "TUESDAY"]
  multiplier    Float @default(1.0)
  serviceZoneId String?
  vehicleType   VehicleType
  isActive      Boolean @default(true)
  
  @@map("peak_hours_config")
}

model WeatherConfig {
  id            String @id @default(uuid())
  weatherType   String
  intensity     String
  multiplier    Float @default(1.0)
  serviceZoneId String?
  vehicleType   VehicleType
  isActive      Boolean @default(true)
  
  @@map("weather_config")
}
```

## Patterns de Migration Appliqués

### 1. Schémas
- **SurrealDB `DEFINE TABLE`** → **Prisma `model`**
- **SurrealDB `TYPE record<Table>`** → **Prisma relations avec `@relation`**
- **SurrealDB `ASSERT $value INSIDE [...]`** → **Prisma `enum`**
- **SurrealDB `DEFAULT`** → **Prisma `@default()`**
- **SurrealDB `UNIQUE INDEX`** → **Prisma `@unique` ou `@@unique`**

### 2. Requêtes
- **SurrealQL `CREATE ... CONTENT`** → **Prisma `CreateOne().Exec()`**
- **SurrealQL `SELECT * FROM ... WHERE`** → **Prisma `FindMany().Exec()`**
- **SurrealQL `UPDATE ... SET`** → **Prisma `Update().Exec()`**
- **SurrealQL `LET ... IF ... THEN`** → **Go conditionals + transactions**
- **SurrealQL `LIVE SELECT`** → **WebSocket/SSE en Go**

### 3. Permissions
- **SurrealDB RLS** → **Middleware Go avec Gin**
- **SurrealQL `$auth.id`** → **JWT claims + context**
- **SurrealQL `$auth.role`** → **User role check en Go**

### 4. Transactions
- **SurrealQL scripts complexes** → **`db.Prisma.Transaction()`**
- **SurrealQL `LET` variables** → **Variables Go locales**

## Actions à Terminer

Pour finaliser la migration :

1. **Appliquer les patterns ci-dessus** aux 5 fichiers restants
2. **Conserver la logique métier** mais l'adapter au style Go/Prisma
3. **Remplacer toutes les requêtes SurrealQL** par des services Go
4. **Mettre à jour les sections permissions** avec des middlewares
5. **Ajouter les index PostgreSQL** appropriés

## Résumé des Bénéfices

✅ **Migration réussie de 50% de la documentation**  
✅ **Cohérence avec l'architecture backend Go actuelle**  
✅ **Examples concrets de services Prisma**  
✅ **Middleware de sécurité JWT**  
✅ **Index PostgreSQL optimisés**  
✅ **Patterns réutilisables pour les 5 fichiers restants**

La documentation est maintenant alignée avec ta stack technique réelle : **Go + Gin + Prisma + PostgreSQL**.