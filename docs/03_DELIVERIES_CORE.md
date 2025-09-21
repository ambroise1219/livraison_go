## Livraisons - Coeur (Delivery, Package, Tracking, DriverLocation)

### Models couverts
- `Delivery`, `Package`, `Location`
- `Tracking` (historique GPS)
- `DriverLocation` (présence/position en attente)

---

### Modèles Prisma (extraits clés)

```prisma
// Delivery
model Delivery {
  id            String        @id @default(uuid())
  clientId      String
  livreurId     String?
  status        DeliveryStatus @default(PENDING)
  type          DeliveryType
  pickupId      String
  dropoffId     String
  vehicleType   VehicleType
  finalPrice    Float
  paymentMethod PaymentMethod
  createdAt     DateTime      @default(now())
  updatedAt     DateTime      @updatedAt
  
  // Relations
  client        User          @relation("ClientDeliveries", fields: [clientId], references: [id])
  livreur       User?         @relation("DriverDeliveries", fields: [livreurId], references: [id])
  pickup        Location      @relation("PickupDeliveries", fields: [pickupId], references: [id])
  dropoff       Location      @relation("DropoffDeliveries", fields: [dropoffId], references: [id])
  packages      Package[]
  tracking      Tracking[]
  
  @@map("deliveries")
}

enum DeliveryStatus {
  PENDING
  ACCEPTED
  PICKED_UP
  DELIVERED
  CANCELLED
  ZONE_ASSIGNED
  PICKUP_IN_PROGRESS
  PICKUP_COMPLETED
  DELIVERY_IN_PROGRESS
  ARRIVED_AT_PICKUP
  ARRIVED_AT_DROPOFF
  IN_TRANSIT
  EN_ROUTE
  SORTED
  SORTING_IN_PROGRESS
}

enum DeliveryType {
  SIMPLE
  EXPRESS
  GROUPEE
  DEMENAGEMENT
}

enum VehicleType {
  MOTO
  VOITURE
  CAMIONNETTE
}

enum PaymentMethod {
  CASH
  MOBILE_MONEY_ORANGE
  MOBILE_MONEY_MTN
  MOBILE_MONEY_MOOV
  MOBILE_MONEY_WAVE
}

// Package
model Package {
  id          String  @id @default(uuid())
  deliveryId  String
  description String?
  weightKg    Float?
  fragile     Boolean @default(false)
  
  delivery Delivery @relation(fields: [deliveryId], references: [id])
  
  @@map("packages")
}

// Location
model Location {
  id      String  @id @default(uuid())
  address String
  lat     Float?
  lng     Float?
  
  pickupDeliveries  Delivery[] @relation("PickupDeliveries")
  dropoffDeliveries Delivery[] @relation("DropoffDeliveries")
  
  @@map("locations")
}

// Tracking
model Tracking {
  id         String   @id @default(uuid())
  deliveryId String
  lat        Float?
  lng        Float?
  timestamp  DateTime @default(now())
  
  delivery Delivery @relation(fields: [deliveryId], references: [id])
  
  @@map("tracking")
}

// DriverLocation
model DriverLocation {
  id          String      @id @default(uuid())
  driverId    String      @unique
  lat         Float?
  lng         Float?
  isAvailable Boolean     @default(true)
  vehicleType VehicleType
  updatedAt   DateTime    @updatedAt
  
  driver User @relation(fields: [driverId], references: [id])
  
  @@map("driver_locations")
}
```

---

### Services Go - Delivery

```go
// Créer une livraison simple
func (s *DeliveryService) CreateDelivery(req *models.CreateDeliveryRequest) (*models.Delivery, error) {
    return s.db.Prisma.Transaction(func(tx *db.PrismaClient) error {
        // Créer les locations
        pickup, err := tx.Location.CreateOne(
            db.Location.Address.Set(req.PickupAddress),
            db.Location.Lat.SetIfPresent(req.PickupLat),
            db.Location.Lng.SetIfPresent(req.PickupLng),
        ).Exec(context.Background())
        if err != nil {
            return nil, err
        }
        
        dropoff, err := tx.Location.CreateOne(
            db.Location.Address.Set(req.DropoffAddress),
            db.Location.Lat.SetIfPresent(req.DropoffLat),
            db.Location.Lng.SetIfPresent(req.DropoffLng),
        ).Exec(context.Background())
        if err != nil {
            return nil, err
        }
        
        // Créer la livraison
        delivery, err := tx.Delivery.CreateOne(
            db.Delivery.ClientID.Set(req.ClientID),
            db.Delivery.Type.Set(req.Type),
            db.Delivery.PickupID.Set(pickup.ID),
            db.Delivery.DropoffID.Set(dropoff.ID),
            db.Delivery.VehicleType.Set(req.VehicleType),
            db.Delivery.FinalPrice.Set(req.FinalPrice),
            db.Delivery.PaymentMethod.Set(req.PaymentMethod),
        ).Exec(context.Background())
        
        if err != nil {
            return nil, err
        }
        
        return ConvertPrismaDeliveryToModel(delivery), nil
    })
}

// Lire les livraisons d'un client
func (s *DeliveryService) GetClientDeliveries(clientID string, limit int) ([]*models.Delivery, error) {
    deliveries, err := s.db.Delivery.FindMany(
        db.Delivery.ClientID.Equals(clientID),
    ).OrderBy(
        db.Delivery.CreatedAt.Order(db.DESC),
    ).Take(limit).With(
        db.Delivery.Pickup.Fetch(),
        db.Delivery.Dropoff.Fetch(),
        db.Delivery.Livreur.Fetch(),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaDeliveries(deliveries), nil
}

// Mettre à jour le statut
func (s *DeliveryService) UpdateStatus(deliveryID string, status db.DeliveryStatus) error {
    _, err := s.db.Delivery.FindUnique(
        db.Delivery.ID.Equals(deliveryID),
    ).Update(
        db.Delivery.Status.Set(status),
    ).Exec(context.Background())
    
    return err
}

// Affecter un livreur
func (s *DeliveryService) AssignDriver(deliveryID, driverID string) error {
    _, err := s.db.Delivery.FindUnique(
        db.Delivery.ID.Equals(deliveryID),
    ).Update(
        db.Delivery.LivreurID.Set(&driverID),
        db.Delivery.Status.Set(db.DeliveryStatusACCEPTED),
    ).Exec(context.Background())
    
    return err
}

// Annuler une livraison
func (s *DeliveryService) CancelDelivery(deliveryID string) error {
    _, err := s.db.Delivery.FindUnique(
        db.Delivery.ID.Equals(deliveryID),
    ).Update(
        db.Delivery.Status.Set(db.DeliveryStatusCANCELLED),
    ).Exec(context.Background())
    
    return err
}
```

---

### Services Go - Packages

```go
// Ajouter un colis à une livraison
func (s *PackageService) AddPackage(deliveryID, description string, weightKg *float64, fragile bool) error {
    _, err := s.db.Package.CreateOne(
        db.Package.DeliveryID.Set(deliveryID),
        db.Package.Description.SetIfPresent(&description),
        db.Package.WeightKg.SetIfPresent(weightKg),
        db.Package.Fragile.Set(fragile),
    ).Exec(context.Background())
    
    return err
}

// Lister les colis d'une livraison
func (s *PackageService) GetDeliveryPackages(deliveryID string) ([]*models.Package, error) {
    packages, err := s.db.Package.FindMany(
        db.Package.DeliveryID.Equals(deliveryID),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaPackages(packages), nil
}
```

---

### Services Go - Tracking (historique GPS)

```go
// Enregistrer un point GPS
func (s *TrackingService) AddTrackingPoint(deliveryID string, lat, lng *float64) error {
    _, err := s.db.Tracking.CreateOne(
        db.Tracking.DeliveryID.Set(deliveryID),
        db.Tracking.Lat.SetIfPresent(lat),
        db.Tracking.Lng.SetIfPresent(lng),
    ).Exec(context.Background())
    
    return err
}

// Récupérer l'historique GPS
func (s *TrackingService) GetDeliveryTracking(deliveryID string) ([]*models.Tracking, error) {
    tracking, err := s.db.Tracking.FindMany(
        db.Tracking.DeliveryID.Equals(deliveryID),
    ).OrderBy(
        db.Tracking.Timestamp.Order(db.ASC),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaTracking(tracking), nil
}

// Pour le temps réel: WebSocket/SSE depuis les handlers
// Les clients s'abonnent via WebSocket pour recevoir les updates
```

---

### Services Go - DriverLocation (présence/zone d'attente)

```go
// Upsert de présence driver
func (s *DriverLocationService) UpdateDriverLocation(driverID string, lat, lng *float64, isAvailable bool, vehicleType db.VehicleType) error {
    // Tentative de mise à jour
    _, err := s.db.DriverLocation.FindUnique(
        db.DriverLocation.DriverID.Equals(driverID),
    ).Update(
        db.DriverLocation.Lat.SetIfPresent(lat),
        db.DriverLocation.Lng.SetIfPresent(lng),
        db.DriverLocation.IsAvailable.Set(isAvailable),
        db.DriverLocation.VehicleType.Set(vehicleType),
    ).Exec(context.Background())
    
    if err != nil {
        // Si pas trouvé, créer
        _, err = s.db.DriverLocation.CreateOne(
            db.DriverLocation.DriverID.Set(driverID),
            db.DriverLocation.Lat.SetIfPresent(lat),
            db.DriverLocation.Lng.SetIfPresent(lng),
            db.DriverLocation.IsAvailable.Set(isAvailable),
            db.DriverLocation.VehicleType.Set(vehicleType),
        ).Exec(context.Background())
    }
    
    return err
}

// Chercher des drivers disponibles par véhicule
// Note: Pour le calcul de distance, utiliser une fonction Go ou une extension PostgreSQL
func (s *DriverLocationService) FindAvailableDrivers(lat, lng float64, vehicleType db.VehicleType, limit int) ([]*models.DriverLocation, error) {
    drivers, err := s.db.DriverLocation.FindMany(
        db.DriverLocation.IsAvailable.Equals(true),
        db.DriverLocation.VehicleType.Equals(vehicleType),
    ).Take(limit).With(
        db.DriverLocation.Driver.Fetch(),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    // Tri par distance en Go (ou utiliser une query PostgreSQL avec ST_Distance)
    result := ConvertPrismaDriverLocations(drivers)
    sort.Slice(result, func(i, j int) bool {
        distI := calculateDistance(lat, lng, *result[i].Lat, *result[i].Lng)
        distJ := calculateDistance(lat, lng, *result[j].Lat, *result[j].Lng)
        return distI < distJ
    })
    
    return result, nil
}

func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
    return math.Sqrt(math.Pow(lat2-lat1, 2) + math.Pow(lng2-lng1, 2))
}
```

---

### Middleware de permissions (Go)

```go
// Middleware pour accès livraisons
func RequireDeliveryAccess() gin.HandlerFunc {
    return func(c *gin.Context) {
        deliveryID := c.Param("deliveryId")
        userID := c.GetString("user_id")
        user := c.MustGet("user").(*models.User)
        
        // Admin accès total
        if user.Role == "ADMIN" || user.Role == "GESTIONNAIRE" {
            c.Next()
            return
        }
        
        // Vérifier si user est client ou livreur de cette delivery
        delivery, err := deliveryService.GetDelivery(deliveryID)
        if err != nil {
            c.JSON(404, gin.H{"error": "Delivery not found"})
            c.Abort()
            return
        }
        
        if delivery.ClientID != userID && (delivery.LivreurID == nil || *delivery.LivreurID != userID) {
            c.JSON(403, gin.H{"error": "Access denied"})
            c.Abort()
            return
        }
        
        c.Set("delivery", delivery)
        c.Next()
    }
}

// Middleware pour accès driver location
func RequireDriverLocationAccess() gin.HandlerFunc {
    return func(c *gin.Context) {
        driverID := c.Param("driverId")
        userID := c.GetString("user_id")
        user := c.MustGet("user").(*models.User)
        
        if user.Role == "ADMIN" || user.Role == "GESTIONNAIRE" || userID == driverID {
            c.Next()
            return
        }
        
        c.JSON(403, gin.H{"error": "Access denied"})
        c.Abort()
    }
}
```

---

### Index PostgreSQL et performances

Index automatiques avec Prisma et suggestions d'optimisation:

```sql
-- Index automatiques Prisma
CREATE INDEX deliveries_client_id_idx ON deliveries(client_id);
CREATE INDEX deliveries_livreur_id_idx ON deliveries(livreur_id);
CREATE INDEX deliveries_status_idx ON deliveries(status);
CREATE INDEX deliveries_created_at_idx ON deliveries(created_at);

-- Index composites recommandés (via @@index dans schema.prisma)
CREATE INDEX deliveries_client_status_idx ON deliveries(client_id, status);
CREATE INDEX deliveries_type_vehicle_idx ON deliveries(type, vehicle_type);

-- Pour packages
CREATE INDEX packages_delivery_id_idx ON packages(delivery_id);

-- Pour tracking (historique GPS)
CREATE INDEX tracking_delivery_timestamp_idx ON tracking(delivery_id, timestamp);

-- Pour driver locations
CREATE INDEX driver_locations_available_vehicle_idx ON driver_locations(is_available, vehicle_type);
```

**Optimisations supplémentaires:**
- Utiliser des index partiels pour les drivers disponibles seulement
- Considérer PostGIS pour les calculs de distance géospatiale
- Partitioning des tables tracking pour les gros volumes


