## Analytics, Incidents, Ratings, RealTimeMetrics

### Tables couvertes
- `Metrics`, `RealTimeMetrics`, `Incident`, `Rating`, `DashboardConfig`

---

### Schémas (extraits)

```sql
-- Metrics (agrégés par jour/driver)
DEFINE TABLE Metrics SCHEMAFULL;
DEFINE FIELD driverId ON TABLE Metrics TYPE record<User>;
DEFINE FIELD date ON TABLE Metrics TYPE datetime;
DEFINE FIELD totalDeliveries ON TABLE Metrics TYPE int DEFAULT 0;
DEFINE FIELD completedDeliveries ON TABLE Metrics TYPE int DEFAULT 0;
DEFINE FIELD cancelledDeliveries ON TABLE Metrics TYPE int DEFAULT 0;
DEFINE FIELD averageRating ON TABLE Metrics TYPE float DEFAULT 0;
DEFINE FIELD totalEarnings ON TABLE Metrics TYPE float DEFAULT 0;

-- RealTimeMetrics
DEFINE TABLE RealTimeMetrics SCHEMAFULL;
DEFINE FIELD timestamp ON TABLE RealTimeMetrics TYPE datetime DEFAULT time::now();
DEFINE FIELD activeDrivers ON TABLE RealTimeMetrics TYPE int DEFAULT 0;
DEFINE FIELD pendingDeliveries ON TABLE RealTimeMetrics TYPE int DEFAULT 0;
DEFINE FIELD averageWaitTime ON TABLE RealTimeMetrics TYPE float DEFAULT 0;
DEFINE FIELD serviceZoneId ON TABLE RealTimeMetrics TYPE record<ServiceZone>;

-- Incident
DEFINE TABLE Incident SCHEMAFULL;
DEFINE FIELD deliveryId ON TABLE Incident TYPE record<Delivery>;
DEFINE FIELD driverId ON TABLE Incident TYPE record<User>;
DEFINE FIELD clientId ON TABLE Incident TYPE record<User>;
DEFINE FIELD type ON TABLE Incident TYPE string;
DEFINE FIELD severity ON TABLE Incident TYPE string;
DEFINE FIELD status ON TABLE Incident TYPE string DEFAULT "OPEN";

-- Rating
DEFINE TABLE Rating SCHEMAFULL;
DEFINE FIELD deliveryId ON TABLE Rating TYPE record<Delivery>;
DEFINE FIELD clientId ON TABLE Rating TYPE record<User>;
DEFINE FIELD driverId ON TABLE Rating TYPE record<User>;
DEFINE FIELD clientRating ON TABLE Rating TYPE int;
DEFINE FIELD driverRating ON TABLE Rating TYPE int;

-- DashboardConfig
DEFINE TABLE DashboardConfig SCHEMAFULL;
DEFINE FIELD userId ON TABLE DashboardConfig TYPE record<User>;
DEFINE FIELD role ON TABLE DashboardConfig TYPE string;
```

---

### Requêtes utiles

```sql
-- Score moyen d'un driver
SELECT math::mean(clientRating) AS avg FROM Rating WHERE driverId = type::thing("User", $driverId);

-- KPI journaliers
SELECT sum(totalDeliveries) AS deliveries, sum(totalEarnings) AS earnings
FROM Metrics WHERE date >= time::today();

-- Temps réel: s'abonner
LIVE SELECT * FROM RealTimeMetrics WHERE serviceZoneId = type::thing("ServiceZone", $zoneId);

-- Incidents ouverts par sévérité
SELECT severity, count() AS n FROM Incident WHERE status = "OPEN" GROUP BY severity;
```

---

### Permissions & Index

```sql
DEFINE TABLE Metrics SCHEMAFULL
  PERMISSIONS
    FOR select WHERE $auth.role IN ["ADMIN","GESTIONNAIRE","DRIVER"],
    FOR create, update, delete WHERE $auth.role IN ["ADMIN","GESTIONNAIRE"];

DEFINE TABLE RealTimeMetrics SCHEMAFULL
  PERMISSIONS
    FOR select WHERE true,
    FOR create, update, delete WHERE $auth.role IN ["ADMIN","GESTIONNAIRE"];

DEFINE TABLE Incident SCHEMAFULL
  PERMISSIONS
    FOR select WHERE deliveryId.clientId = $auth.id OR driverId = $auth.id OR $auth.role IN ["ADMIN","SUPPORT"],
    FOR create, update WHERE driverId = $auth.id OR deliveryId.clientId = $auth.id OR $auth.role IN ["ADMIN","SUPPORT"],
    FOR delete WHERE $auth.role = "ADMIN";

DEFINE TABLE Rating SCHEMAFULL
  PERMISSIONS
    FOR select WHERE driverId = $auth.id OR clientId = $auth.id OR $auth.role = "ADMIN",
    FOR create WHERE deliveryId.clientId = $auth.id,
    FOR update, delete WHERE $auth.role = "ADMIN";
```

Recommandés:
- `Metrics.driverId, date`
- `RealTimeMetrics.timestamp, serviceZoneId`
- `Incident.deliveryId, driverId, status, severity`
- `Rating.driverId, clientId, createdAt`


