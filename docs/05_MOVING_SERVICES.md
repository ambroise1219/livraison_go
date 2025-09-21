## Déménagement (MovingService)

### Objectif
Modéliser et piloter un déménagement: taille véhicule, nombre d'aides, services supplémentaires, calculs de coûts et suivi.

---

### Schéma (extraits)

```sql
DEFINE TABLE MovingService SCHEMAFULL;
DEFINE FIELD deliveryId        ON TABLE MovingService TYPE record<Delivery>;
DEFINE FIELD vehicleSize       ON TABLE MovingService TYPE string ASSERT $value INSIDE ["MINI","SMALL","MEDIUM","LARGE","EXTRA_LARGE"];
DEFINE FIELD helpersCount      ON TABLE MovingService TYPE int DEFAULT 1;
DEFINE FIELD floors            ON TABLE MovingService TYPE int DEFAULT 1;
DEFINE FIELD hasElevator       ON TABLE MovingService TYPE bool DEFAULT false;
DEFINE FIELD needsDisassembly  ON TABLE MovingService TYPE bool DEFAULT false;
DEFINE FIELD hasFragileItems   ON TABLE MovingService TYPE bool DEFAULT false;
DEFINE FIELD additionalServices ON TABLE MovingService TYPE array<string> DEFAULT [];
DEFINE FIELD specialInstructions ON TABLE MovingService TYPE option<string>;
DEFINE FIELD estimatedVolume   ON TABLE MovingService TYPE option<float>;
DEFINE FIELD helpersCost       ON TABLE MovingService TYPE float DEFAULT 0.0;
DEFINE FIELD vehicleCost       ON TABLE MovingService TYPE float DEFAULT 0.0;
DEFINE FIELD serviceCost       ON TABLE MovingService TYPE float DEFAULT 0.0;
```

---

### Flux SurrealQL - Création

```sql
-- 1) Créer la Delivery mère (type DEMENAGEMENT)
LET $client = type::thing("User", $clientId);
LET $pickup  = CREATE Location CONTENT { address: $pickupAddress, lat: $plat, lng: $plng };
LET $dropoff = CREATE Location CONTENT { address: $dropAddress, lat: $dlat, lng: $dlng };
LET $del = CREATE Delivery CONTENT {
  clientId: $client,
  type: "DEMENAGEMENT",
  pickupId: $pickup.id,
  dropoffId: $dropoff.id,
  vehicleType: $vehicleType,
  finalPrice: 0.0,
  paymentMethod: $payment
};

-- 2) Enregistrer les détails MovingService
LET $ms = CREATE MovingService CONTENT {
  deliveryId: $del.id,
  vehicleSize: $vehicleSize,
  helpersCount: $helpersCount,
  floors: $floors,
  hasElevator: $hasElevator,
  needsDisassembly: $needsDisassembly,
  hasFragileItems: $hasFragileItems,
  additionalServices: $additionalServices,
  specialInstructions: $specialInstructions,
  estimatedVolume: $estimatedVolume
};

-- 3) Calculs de coûts (exemple indicatif)
LET $vehicleBase = switch($vehicleSize) {
  case "MINI"        => 10,
  case "SMALL"       => 15,
  case "MEDIUM"      => 20,
  case "LARGE"       => 30,
  case "EXTRA_LARGE" => 45,
  default => 20
};
LET $helpersBase = $helpersCount * 8;         -- coût par aide
LET $floorCost   = math::max(0,$floors-1) * 5; -- coût étages
LET $fragile     = $hasFragileItems ? 5 : 0;
LET $services    = array::len($additionalServices) * 5; -- coût simple par service

LET $vehicleCost = $vehicleBase;
LET $helpersCost = $helpersBase + $floorCost + $fragile;
LET $serviceCost = $services;
LET $total       = $vehicleCost + $helpersCost + $serviceCost;

UPDATE $ms.id SET vehicleCost = $vehicleCost, helpersCost = $helpersCost, serviceCost = $serviceCost;
UPDATE $del.id SET finalPrice = $total;
```

---

### Mise à jour & lecture

```sql
-- Mise à jour des paramètres (recalculer ensuite)
UPDATE type::thing("MovingService", $msId) MERGE {
  helpersCount: $helpersCount,
  floors: $floors,
  hasElevator: $hasElevator,
  needsDisassembly: $needsDisassembly,
  hasFragileItems: $hasFragileItems,
  additionalServices: $additionalServices,
  estimatedVolume: $estimatedVolume
};

-- Récupérer la fiche complète
SELECT * FROM MovingService WHERE deliveryId = type::thing("Delivery", $deliveryId);
```

---

### Permissions suggérées

```sql
DEFINE TABLE MovingService SCHEMAFULL
  PERMISSIONS
    FOR select WHERE deliveryId.clientId = $auth.id OR deliveryId.livreurId = $auth.id OR $auth.role = "ADMIN",
    FOR create, update WHERE deliveryId.clientId = $auth.id OR $auth.role = "ADMIN",
    FOR delete WHERE $auth.role = "ADMIN";
```

---

### Index recommandés

- `MovingService.deliveryId`
- `MovingService.vehicleSize`


