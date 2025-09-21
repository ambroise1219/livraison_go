## Handoff Frontend: Prisma/PostgreSQL vs Backend Go vs Hybride

Objectif: éviter les mauvaises intégrations. Cette page indique qui implémente quoi, avec exemples et URLs.

**⚠️ MISE À JOUR ARCHITECTURE**: Migré de SurrealDB vers **Prisma ORM + PostgreSQL**

### Base URLs (VPS)
```
PostgreSQL Database:       postgresql://user:pass@localhost:5432/livraison_db
Prisma Client:             Intégré dans Go backend
Go API (HTTP):             http://172.187.248.156:8080
Health:                    http://172.187.248.156:8080/health
Database Stats:            http://172.187.248.156:8080/db
```

---

### 1) À faire DIRECTEMENT via API Go (Frontend)

- **CRUD simples** sur tables non sensibles via API REST:
  - `Delivery` (création simple/express), `Location`, `Package`
  - `GroupedDelivery`, `DeliveryZone` (création, mise à jour statut)
  - `MovingService` (création/mise à jour des paramètres)
  - `UserAddress`, `PaymentMethod` (limité à l'utilisateur courant)
  - `Notification` (création PENDING) et lecture de `Banner`
- **Live Queries (temps réel)**:
  - `Tracking` d'une livraison, `DriverLocation` disponibles, `RealTimeMetrics` lecture
- **Lecture de configuration**:
  - `PricingRule`, `PlatformConfig`, `PeakHoursConfig`, `WeatherConfig`, `ServiceZone`
- **Promos & Parrainage (lecture & simulation)**:
  - Lire `Promo` et simuler un calcul côté app; l'usage final passe côté Go si paiement réel

Exemples rapides (API REST):
```ts
// Créer une location
const location = await fetch('/api/v1/locations', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` },
  body: JSON.stringify({ address, lat, lng })
});

// Créer une livraison
const delivery = await fetch('/api/v1/delivery/', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` },
  body: JSON.stringify({ type: 'SIMPLE', pickupId, dropoffId, vehicleType, finalPrice, paymentMethod: 'CASH' })
});

// Suivi en temps réel (WebSocket ou polling)
const tracking = await fetch(`/api/v1/delivery/client/${deliveryId}/track`);
```

Prérequis: Authentification JWT requise pour toutes les requêtes.

---

### 2) À faire UNIQUEMENT via Backend Go

- **OTP / Auth / Tokens**: envoi via WhatsApp (Wanotifier), vérif, génération JWT/refresh
- **Paiements réels**: intégrations externes, validation, webhooks
- **Calculs avancés de prix**: promotions, période de pointe, météo, minimums, règles métier
- **Assignation intelligente** des livreurs et arbitrage de capacité
- **Actions sensibles/atomiques**:
  - Annulation avec pénalité, refund, ajustement wallet
  - Application effective d'un code promo (création `PromoUsage` final)
  - Génération et vérification d'OTP de livraison (proof-of-delivery)
- **Notifications sortantes** (WhatsApp/SMS/Push/Email) effectives (changement de statut à SENT/FAILED)
- **Analytique consolidée**: écriture `Metrics`, `PlatformFee`, agrégations exigeant cohérence
- **Modération & Admin**: bannissements, sanctions (`DriverPenalty`), bonus (`DriverBonus`)
- **Cache LMDB**: gestion des positions live, données fréquentes, TTL

Raison: sécurité, intégrité transactionnelle, secrets/API keys, logique non triviale, performance optimisée.

---

### 3) Flux HYBRIDES (Frontend + Go)

- **Création de livraison avec paiement**:
  - Front: crée `Delivery` + `Location` via API + calcule prix estimé (lecture `PricingRule`, `PlatformConfig`)
  - Go: confirme prix (règles complètes), initie paiement, enregistre `Payment`, `PlatformFee`, crédite `Wallet`
- **Promo code**:
  - Front: vérifie promo via API (lecture `Promo`), affiche prix estimé
  - Go: applique promo, crée `PromoUsage`, verrouille usages/limites
- **Assignation**:
  - Front: liste `DriverLocation` dispos via API (lecture), UI sélection éventuelle
  - Go: décision finale d'assignation, maj `Delivery.livreurId`, cohérence multi-contraintes
- **Tracking**:
  - Front Driver: envoie points `Tracking` via API
  - Go: agrège/contrôle cadence, déclenche notifications/états avancés si besoin, cache LMDB
- **Grouped delivery**:
  - Front: crée `GroupedDelivery`, `DeliveryZone` via API, suit progression
  - Go: recalculs finaux, facturation globale, pénalités/bonus

---

### 4) Check-list par domaine

- Profil & adresses
  - Front: `User` (update limité), `UserAddress` CRUD, `PaymentMethod` CRUD via API
  - Go: OTP/Login via WhatsApp, Refresh, sécurité, validation
- Livraisons core
  - Front: création simple/express via API, lecture/filtrage, `Tracking` live
  - Go: transitions critiques (annulation pénalisée, delivered avec preuve), assignation
- Grouped / Moving
  - Front: création et updates zones/paramètres via API
  - Go: pricing final, paiements, metrics
- Paiements & Wallet
  - Front: affichage, lecture historique via API
  - Go: encaissement, refunds, ajustements wallet, `PlatformFee`
- Notifications
  - Front: créer en PENDING via API, lecture bannières/templates
  - Go: envoi effectif via WhatsApp et passage à SENT/FAILED

---

### 5) Écueils à éviter

- Ne jamais stocker de secrets côté Front (WhatsApp, paiements, webhooks)
- Ne pas écrire `Wallet`/`WalletTransaction` côté Front (lecture OK)
- Toujours utiliser l'authentification JWT pour toutes les requêtes API
- Utiliser les endpoints API appropriés pour les opérations CRUD
- S'appuyer sur les index SQLite existants pour les requêtes optimisées
- Respecter les contraintes de validation côté backend

---

### 6) Contacts & Debug

- Si une requête API échoue en prod: fournir l'endpoint, payload, headers, et la stack UI
- Pour tout flux hybride/paiement: passer par le backend Go et ouvrir un ticket si doute
- Vérifier les logs du serveur Go pour les erreurs de base de données
- Utiliser `/health` et `/db` pour diagnostiquer l'état du système


