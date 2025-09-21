## Paiements, Portefeuille, Frais de plateforme

### Tables couvertes
- `Payment`, `Wallet`, `WalletTransaction`, `PlatformFee`, `PricingRule`

---

### Schémas (extraits)

```sql
-- Payment
DEFINE TABLE Payment SCHEMAFULL;
DEFINE FIELD userId ON TABLE Payment TYPE record<User>;
DEFINE FIELD amount ON TABLE Payment TYPE float;
DEFINE FIELD method ON TABLE Payment TYPE string ASSERT $value INSIDE ["CASH","WAVE","MOBILE_MONEY_ORANGE","MOBILE_MONEY_MTN","MOBILE_MONEY_MOOV"];
DEFINE FIELD paymentType ON TABLE Payment TYPE string ASSERT $value INSIDE ["DELIVERY_PAYMENT","WALLET_RECHARGE","SUBSCRIPTION","FINE","BONUS"];
DEFINE FIELD status ON TABLE Payment TYPE string DEFAULT "PENDING" ASSERT $value INSIDE ["PENDING","COMPLETED","FAILED","REFUNDED"];
DEFINE FIELD reference ON TABLE Payment TYPE string;
DEFINE INDEX unique_reference ON TABLE Payment COLUMNS reference UNIQUE;

-- Wallet & WalletTransaction
DEFINE TABLE Wallet SCHEMAFULL;
DEFINE FIELD userId ON TABLE Wallet TYPE record<User>;
DEFINE FIELD balance ON TABLE Wallet TYPE float DEFAULT 0.0;

DEFINE TABLE WalletTransaction SCHEMAFULL;
DEFINE FIELD walletId ON TABLE WalletTransaction TYPE record<Wallet>;
DEFINE FIELD amount ON TABLE WalletTransaction TYPE float;
DEFINE FIELD type ON TABLE WalletTransaction TYPE string; -- CREDIT/DEBIT

-- PricingRule
DEFINE TABLE PricingRule SCHEMAFULL;
DEFINE FIELD vehicleType ON TABLE PricingRule TYPE string ASSERT $value INSIDE ["MOTO","VOITURE","CAMIONNETTE"];
DEFINE FIELD basePrice ON TABLE PricingRule TYPE float;
DEFINE FIELD includedKm ON TABLE PricingRule TYPE float;
DEFINE FIELD perKm ON TABLE PricingRule TYPE float;
DEFINE FIELD waitingFree ON TABLE PricingRule TYPE int;
DEFINE FIELD waitingRate ON TABLE PricingRule TYPE float;

-- PlatformFee
DEFINE TABLE PlatformFee SCHEMAFULL;
DEFINE FIELD deliveryId ON TABLE PlatformFee TYPE record<Delivery>;
DEFINE FIELD driverId ON TABLE PlatformFee TYPE record<User>;
DEFINE FIELD clientId ON TABLE PlatformFee TYPE record<User>;
DEFINE FIELD baseFee ON TABLE PlatformFee TYPE float;
DEFINE FIELD commissionRate ON TABLE PlatformFee TYPE float;
DEFINE FIELD commissionAmount ON TABLE PlatformFee TYPE float;
DEFINE FIELD serviceFee ON TABLE PlatformFee TYPE float;
DEFINE FIELD totalFee ON TABLE PlatformFee TYPE float;
DEFINE FIELD driverEarnings ON TABLE PlatformFee TYPE float;
DEFINE FIELD platformEarnings ON TABLE PlatformFee TYPE float;
```

---

### Flux - Encaissement d'une livraison (exemple)

```sql
-- 1) Vérifier/charger la règle de pricing
LET $rule = SELECT * FROM PricingRule WHERE vehicleType = $vehicleType LIMIT 1;

-- 2) Calculer le prix (simplifié)
LET $distanceCost = math::max(0, $distanceKm - $rule[0].includedKm) * $rule[0].perKm;
LET $waitingCost  = math::max(0, $waitingMin - $rule[0].waitingFree) * $rule[0].waitingRate;
LET $base = $rule[0].basePrice + $distanceCost + $waitingCost;

-- 3) Appliquer commissions & frais
LET $commissionRate = 0.2; -- 20%
LET $serviceFee = 1.0;
LET $commission = $base * $commissionRate;
LET $totalFee  = $commission + $serviceFee;
LET $driverEarnings = math::max(0, $base - $totalFee);

-- 4) Enregistrer PlatformFee
CREATE PlatformFee CONTENT {
  deliveryId: type::thing("Delivery", $deliveryId),
  driverId: type::thing("User", $driverId),
  clientId: type::thing("User", $clientId),
  baseFee: $base,
  commissionRate: $commissionRate,
  commissionAmount: $commission,
  serviceFee: $serviceFee,
  totalFee: $totalFee,
  driverEarnings: $driverEarnings,
  platformEarnings: $commission + $serviceFee
};

-- 5) Marquer paiement et ajuster wallet driver
UPDATE type::thing("Delivery", $deliveryId) SET paidAt = time::now();
LET $dw = (SELECT * FROM Wallet WHERE userId = type::thing("User", $driverId))[0];
IF $dw = NONE THEN (
  LET $nw = CREATE Wallet CONTENT { userId: type::thing("User", $driverId), balance: 0.0 };
  UPDATE $nw.id SET balance += $driverEarnings;
  CREATE WalletTransaction CONTENT { walletId: $nw.id, amount: $driverEarnings, type: "CREDIT", description: "Delivery earnings" }
) ELSE (
  UPDATE $dw.id SET balance += $driverEarnings;
  CREATE WalletTransaction CONTENT { walletId: $dw.id, amount: $driverEarnings, type: "CREDIT", description: "Delivery earnings" }
) END;
```

---

### Refund/Annulation (exemple)

```sql
-- Remboursement client (si payé via wallet)
LET $cw = (SELECT * FROM Wallet WHERE userId = type::thing("User", $clientId))[0];
IF $cw != NONE THEN (
  UPDATE $cw.id SET balance += $amount;
  CREATE WalletTransaction CONTENT { walletId: $cw.id, amount: $amount, type: "CREDIT", description: "Refund" }
);

-- Débit des gains du livreur si déjà crédités
LET $dw = (SELECT * FROM Wallet WHERE userId = type::thing("User", $driverId))[0];
IF $dw != NONE THEN (
  UPDATE $dw.id SET balance -= $driverDebit;
  CREATE WalletTransaction CONTENT { walletId: $dw.id, amount: -$driverDebit, type: "DEBIT", description: "Cancellation adjustment" }
);
```

---

### Permissions & Index

```sql
DEFINE TABLE Payment SCHEMAFULL
  PERMISSIONS
    FOR select WHERE userId = $auth.id OR $auth.role = "ADMIN",
    FOR create WHERE userId = $auth.id OR $auth.role = "ADMIN",
    FOR update, delete WHERE $auth.role = "ADMIN";

DEFINE TABLE PlatformFee SCHEMAFULL
  PERMISSIONS
    FOR select WHERE $auth.role IN ["ADMIN","GESTIONNAIRE"],
    FOR create, update, delete WHERE $auth.role = "ADMIN";
```

Recommandés:
- `Payment.userId, status, createdAt`
- `Wallet.userId` UNIQUE
- `WalletTransaction.walletId, createdAt`
- `PricingRule.vehicleType` UNIQUE
- `PlatformFee.deliveryId, driverId, clientId`


