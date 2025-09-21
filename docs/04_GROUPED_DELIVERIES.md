## Livraisons groupées (GroupedDelivery, DeliveryZone)

### Objectif
Optimiser le coût pour un client avec plusieurs arrêts (zones), en créant une livraison mère (`Delivery` + `GroupedDelivery`) et des sous-zones (`DeliveryZone`).

---

### Schémas (extraits)

```sql
-- GroupedDelivery
DEFINE TABLE GroupedDelivery SCHEMAFULL;
DEFINE FIELD deliveryId        ON TABLE GroupedDelivery TYPE record<Delivery>;
DEFINE FIELD totalZones        ON TABLE GroupedDelivery TYPE int;
DEFINE FIELD completedZones    ON TABLE GroupedDelivery TYPE int DEFAULT 0;
DEFINE FIELD discountPercentage ON TABLE GroupedDelivery TYPE float DEFAULT 30.0;
DEFINE FIELD originalPrice     ON TABLE GroupedDelivery TYPE float;
DEFINE FIELD finalPrice        ON TABLE GroupedDelivery TYPE float;

-- DeliveryZone
DEFINE TABLE DeliveryZone SCHEMAFULL;
DEFINE FIELD groupedDeliveryId  ON TABLE DeliveryZone TYPE record<GroupedDelivery>;
DEFINE FIELD zoneNumber         ON TABLE DeliveryZone TYPE int; -- 1..N
DEFINE FIELD recipientName      ON TABLE DeliveryZone TYPE string;
DEFINE FIELD recipientPhone     ON TABLE DeliveryZone TYPE string;
DEFINE FIELD pickupLocationId   ON TABLE DeliveryZone TYPE record<Location>;
DEFINE FIELD deliveryLocationId ON TABLE DeliveryZone TYPE record<Location>;
DEFINE FIELD status ON TABLE DeliveryZone TYPE string ASSERT $value INSIDE [
  "PENDING","ACCEPTED","PICKED_UP","DELIVERED","CANCELLED",
  "ZONE_ASSIGNED","PICKUP_IN_PROGRESS","PICKUP_COMPLETED","DELIVERY_IN_PROGRESS",
  "ARRIVED_AT_PICKUP","IN_TRANSIT","ARRIVED_AT_DESTINATION","UNLOADING_COMPLETED","ARRIVED_AT_DROPOFF"
];
DEFINE FIELD price ON TABLE DeliveryZone TYPE float;
```

---

### Flux SurrealQL - Création complète (frontend direct)

```sql
-- 1) Créer la Delivery mère
LET $client = type::thing("User", $clientId);
LET $pickup  = CREATE Location CONTENT { address: $pickupAddress, lat: $plat, lng: $plng };
LET $dropoff = CREATE Location CONTENT { address: $dropAddress, lat: $dlat, lng: $dlng };
LET $del = CREATE Delivery CONTENT {
  clientId: $client,
  type: "GROUPEE",
  pickupId: $pickup.id,
  dropoffId: $dropoff.id,
  vehicleType: $vehicleType,
  finalPrice: 0.0,
  paymentMethod: $payment
};

-- 2) Créer GroupedDelivery (N zones prévues)
LET $gd = CREATE GroupedDelivery CONTENT {
  deliveryId: $del.id,
  totalZones: array::len($zones),
  originalPrice: 0.0,
  finalPrice: 0.0
};

-- 3) Créer chaque zone avec ses locations
FOREACH $z IN $zones THEN (
  LET $pl = CREATE Location CONTENT { address: $z.pickupAddress, lat: $z.plat, lng: $z.plng };
  LET $dl = CREATE Location CONTENT { address: $z.dropAddress, lat: $z.dlat, lng: $z.dlng };
  CREATE DeliveryZone CONTENT {
    groupedDeliveryId: $gd.id,
    zoneNumber: $z.zoneNumber,
    recipientName: $z.recipientName,
    recipientPhone: $z.recipientPhone,
    pickupLocationId: $pl.id,
    deliveryLocationId: $dl.id,
    price: $z.price,
    status: "PENDING"
  };
);

-- 4) Calculer les prix: somme zones et remise
LET $sum = SELECT math::sum(price) AS total FROM DeliveryZone WHERE groupedDeliveryId = $gd.id;
LET $orig = $sum[0].total ?? 0;
LET $discount = $orig * ($discountPct ?? 30.0) / 100;
LET $final = math::max(0, $orig - $discount);
UPDATE $gd.id SET originalPrice = $orig, finalPrice = $final, discountPercentage = $discountPct ?? 30.0;
UPDATE $del.id SET finalPrice = $final;
```

---

### Avancement d'une zone

```sql
-- Marquer une zone comme récupérée
UPDATE type::thing("DeliveryZone", $zoneId) SET status = "PICKED_UP", pickedUpAt = time::now();

-- Marquer une zone comme livrée (et incrémenter le compteur)
LET $z = SELECT * FROM type::thing("DeliveryZone", $zoneId);
LET $gd = $z[0].groupedDeliveryId;
UPDATE $z[0].id SET status = "DELIVERED", deliveredAt = time::now();
UPDATE $gd SET completedZones += 1;

-- Si toutes les zones terminées, terminer la delivery mère
LET $g = SELECT totalZones, completedZones, deliveryId FROM $gd;
IF $g[0].completedZones >= $g[0].totalZones THEN (
  UPDATE $g[0].deliveryId SET status = "DELIVERED"
) END;
```

---

### Lecture & vues utiles

```sql
-- Détails d'une livraison groupée
SELECT * FROM GroupedDelivery WHERE deliveryId = type::thing("Delivery", $deliveryId);

-- Zones d'une livraison groupée (dans l'ordre)
SELECT * FROM DeliveryZone WHERE groupedDeliveryId = $gd ORDER BY zoneNumber ASC;

-- Statistiques simples
SELECT count() AS total, math::sum(price) AS sum FROM DeliveryZone WHERE groupedDeliveryId = $gd;
```

---

### Permissions suggérées

```sql
DEFINE TABLE GroupedDelivery SCHEMAFULL
  PERMISSIONS
    FOR select WHERE deliveryId.clientId = $auth.id OR deliveryId.livreurId = $auth.id OR $auth.role = "ADMIN",
    FOR create, update WHERE deliveryId.clientId = $auth.id OR $auth.role = "ADMIN",
    FOR delete WHERE $auth.role = "ADMIN";

DEFINE TABLE DeliveryZone SCHEMAFULL
  PERMISSIONS
    FOR select WHERE groupedDeliveryId.deliveryId.clientId = $auth.id OR groupedDeliveryId.deliveryId.livreurId = $auth.id OR $auth.role = "ADMIN",
    FOR create, update WHERE groupedDeliveryId.deliveryId.clientId = $auth.id OR $auth.role = "ADMIN",
    FOR delete WHERE $auth.role = "ADMIN";
```

---

### Index recommandés

- `GroupedDelivery.deliveryId` (UNIQUE)
- `DeliveryZone.groupedDeliveryId`
- `DeliveryZone.groupedDeliveryId, zoneNumber` (UNIQUE pour ordre)
- `DeliveryZone.status`


