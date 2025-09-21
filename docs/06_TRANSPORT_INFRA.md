## Transport & Infrastructure (Prisma/PostgreSQL)

### Models couverts
- `Vehicle`, `Location`, `ServiceZone`, `Depot`

---

### Modèles Prisma

```prisma
// Vehicle
model Vehicle {
  id                    String      @id @default(uuid())
  userId                String      @unique
  type                  VehicleType
  marque                String?
  modele                String?
  plaqueImmatriculation String?
  isActive              Boolean     @default(true)
  createdAt             DateTime    @default(now())
  
  user User @relation(fields: [userId], references: [id])
  
  @@map("vehicles")
}

// Location (déjà défini dans 03_DELIVERIES_CORE.md)
model Location {
  id      String  @id @default(uuid())
  address String
  lat     Float?
  lng     Float?
  
  // Relations
  pickupDeliveries  Delivery[] @relation("PickupDeliveries")
  dropoffDeliveries Delivery[] @relation("DropoffDeliveries")
  depots           Depot[]
  
  @@map("locations")
}

// ServiceZone
model ServiceZone {
  id       String                 @id @default(uuid())
  name     String?
  city     String
  polygon  Json?                  // Coordonnées GeoJSON du polygone
  isActive Boolean                @default(true)
  
  @@map("service_zones")
}

// Depot
model Depot {
  id         String   @id @default(uuid())
  name       String?
  locationId String
  isActive   Boolean  @default(true)
  capacity   Int?
  
  location Location @relation(fields: [locationId], references: [id])
  
  @@map("depots")
}
```

---

### Services Go

```go
// VehicleService
type VehicleService struct {
    db *db.PrismaClient
}

// Associer un véhicule à un livreur
func (s *VehicleService) CreateVehicle(userID string, vehicleType db.VehicleType, marque, modele, plaque *string) (*models.Vehicle, error) {
    vehicle, err := s.db.Vehicle.CreateOne(
        db.Vehicle.UserID.Set(userID),
        db.Vehicle.Type.Set(vehicleType),
        db.Vehicle.Marque.SetIfPresent(marque),
        db.Vehicle.Modele.SetIfPresent(modele),
        db.Vehicle.PlaqueImmatriculation.SetIfPresent(plaque),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaVehicleToModel(vehicle), nil
}

// Obtenir le véhicule d'un livreur
func (s *VehicleService) GetDriverVehicle(driverID string) (*models.Vehicle, error) {
    vehicle, err := s.db.Vehicle.FindUnique(
        db.Vehicle.UserID.Equals(driverID),
    ).With(
        db.Vehicle.User.Fetch(),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaVehicleToModel(vehicle), nil
}

// LocationService
type LocationService struct {
    db *db.PrismaClient
}

// Créer une localisation
func (s *LocationService) CreateLocation(address string, lat, lng *float64) (*models.Location, error) {
    location, err := s.db.Location.CreateOne(
        db.Location.Address.Set(address),
        db.Location.Lat.SetIfPresent(lat),
        db.Location.Lng.SetIfPresent(lng),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaLocationToModel(location), nil
}

// ServiceZoneService
type ServiceZoneService struct {
    db *db.PrismaClient
}

// Trouver zones actives par ville
func (s *ServiceZoneService) GetActiveZonesByCity(city string) ([]*models.ServiceZone, error) {
    zones, err := s.db.ServiceZone.FindMany(
        db.ServiceZone.City.Equals(city),
        db.ServiceZone.IsActive.Equals(true),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaServiceZones(zones), nil
}

// Vérifier si point dans zone (exemple basique)
func (s *ServiceZoneService) IsPointInZone(zoneID string, lat, lng float64) (bool, error) {
    zone, err := s.db.ServiceZone.FindUnique(
        db.ServiceZone.ID.Equals(zoneID),
    ).Exec(context.Background())
    
    if err != nil {
        return false, err
    }
    
    // Logic pour vérifier si le point est dans le polygone
    // Utiliser une lib de géométrie ou PostGIS
    return pointInPolygon(lat, lng, zone.Polygon), nil
}

// DepotService
type DepotService struct {
    db *db.PrismaClient
}

// Créer un dépôt
func (s *DepotService) CreateDepot(name, address string, lat, lng *float64, capacity *int) (*models.Depot, error) {
    return s.db.Prisma.Transaction(func(tx *db.PrismaClient) (*models.Depot, error) {
        // Créer d'abord la location
        location, err := tx.Location.CreateOne(
            db.Location.Address.Set(address),
            db.Location.Lat.SetIfPresent(lat),
            db.Location.Lng.SetIfPresent(lng),
        ).Exec(context.Background())
        
        if err != nil {
            return nil, err
        }
        
        // Créer le dépôt
        depot, err := tx.Depot.CreateOne(
            db.Depot.Name.SetIfPresent(&name),
            db.Depot.LocationID.Set(location.ID),
            db.Depot.Capacity.SetIfPresent(capacity),
        ).With(
            db.Depot.Location.Fetch(),
        ).Exec(context.Background())
        
        if err != nil {
            return nil, err
        }
        
        return ConvertPrismaDepotToModel(depot), nil
    })
}

// Lister dépôts actifs
func (s *DepotService) GetActiveDepots() ([]*models.Depot, error) {
    depots, err := s.db.Depot.FindMany(
        db.Depot.IsActive.Equals(true),
    ).With(
        db.Depot.Location.Fetch(),
    ).Exec(context.Background())
    
    if err != nil {
        return nil, err
    }
    
    return ConvertPrismaDepots(depots), nil
}

// Fonction utilitaire pour point dans polygone (exemple basique)
func pointInPolygon(lat, lng float64, polygon interface{}) bool {
    // Implémentation basique ou utilisation de PostGIS
    // Pour une implémentation complète, utiliser une lib de géométrie
    return true // placeholder
}
```

---

### Middleware de permissions

```go
// Middleware pour accès véhicule
func RequireVehicleAccess() gin.HandlerFunc {
    return func(c *gin.Context) {
        vehicleUserID := c.Param("userId") // ou via le véhicule ID
        userID := c.GetString("user_id")
        user := c.MustGet("user").(*models.User)
        
        if user.Role == "ADMIN" || userID == vehicleUserID {
            c.Next()
            return
        }
        
        c.JSON(403, gin.H{"error": "Access denied"})
        c.Abort()
    }
}

// Middleware pour zones et dépôts (admin seulement pour modification)
func RequireAdminForInfrastructure() gin.HandlerFunc {
    return RequireRole("ADMIN", "GESTIONNAIRE")
}
```

**Index PostgreSQL:**
```sql
-- Véhicules
CREATE UNIQUE INDEX vehicles_user_id_key ON vehicles(user_id);
CREATE INDEX vehicles_type_active_idx ON vehicles(type, is_active);

-- Zones de service
CREATE INDEX service_zones_city_active_idx ON service_zones(city, is_active);
CREATE INDEX service_zones_active_idx ON service_zones(is_active);

-- Dépôts
CREATE INDEX depots_location_id_idx ON depots(location_id);
CREATE INDEX depots_active_idx ON depots(is_active);

-- Locations avec extension PostGIS pour géolocalisation
-- CREATE INDEX locations_point_idx ON locations USING GIST (ST_Point(lng, lat));
```


