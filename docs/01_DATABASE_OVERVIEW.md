# 🗄️ Base de Données Prisma/PostgreSQL - Vue d'ensemble

## 📋 Architecture Moderne Type-Safe

**⚠️ ARCHITECTURE MISE À JOUR** - Migration terminée le 21/09/2025

### 🎯 **Principe de base :**
- **PostgreSQL** = Base de données principale, robuste, scalable
- **Prisma ORM** = Client type-safe, migrations automatiques, développement rapide
- **Go Backend** = OTP via WhatsApp, logique métier complexe, calculs, validations
- **SQLite** = Environnement de développement local uniquement

---

## 🏗️ **Structure de la Base de Données**

### **Base de Données**
```
Production:  postgresql://user:pass@localhost:5432/livraison_db
Développement: file:./dev.db (SQLite local)
Prisma Schema: prisma/schema.prisma
```

### Base URLs (VPS)
```
PostgreSQL Production:     postgresql://user:pass@localhost:5432/livraison_db
SQLite Développement:       file:./dev.db  
Go API (HTTP):             http://172.187.248.156:8080
Health Check:              http://172.187.248.156:8080/health
Database Stats:            http://172.187.248.156:8080/db
```

### **Tables Principales (43 tables)**

#### 🔐 **Authentification & Sécurité**
- `users` - Utilisateurs (clients, livreurs, admins, gestionnaires, marketing)
- `otps` - Codes de vérification WhatsApp (4 chiffres)
- `refresh_tokens` - Tokens de rafraîchissement
- `emergency_contacts` - Contacts d'urgence

#### 🚚 **Livraisons & Logistique**
- `deliveries` - Livraisons principales (simple, express, groupée, déménagement)
- `packages` - Colis associés
- `locations` - Adresses géolocalisées
- `driver_locations` - Positions des livreurs (cache LMDB)
- `trackings` - Suivi GPS temps réel (cache LMDB)

#### 🚛 **Véhicules & Transport**
- `vehicles` - Véhicules des livreurs
- `service_zones` - Zones de couverture
- `depots` - Gestion des dépôts

#### 💰 **Paiements & Finances**
- `payments` - Paiements
- `wallets` - Portefeuilles électroniques
- `wallet_transactions` - Transactions de portefeuille
- `platform_fees` - Commissions plateforme
- `pricing_rules` - Règles de tarification

#### 🎯 **Livraisons Spécialisées**
- `grouped_deliveries` - Livraisons groupées
- `delivery_zones` - Zones de livraison groupée
- `moving_services` - Services de déménagement

#### 🎁 **Promotions & Parrainage**
- `promos` - Codes promotionnels
- `promo_usages` - Utilisation des promos
- `referrals` - Système de parrainage

#### 📱 **Notifications & Communication**
- `notifications` - Notifications utilisateurs
- `notification_templates` - Templates de communication
- `banners` - Gestion des bannières

#### 📊 **Analytics & Monitoring**
- `metrics` - Métriques des livreurs
- `real_time_metrics` - Monitoring temps réel (cache LMDB)
- `incidents` - Gestion des problèmes
- `ratings` - Système d'évaluation

#### ⚙️ **Configuration & Administration**
- `platform_configs` - Configuration dynamique
- `peak_hours_configs` - Heures de pointe
- `weather_configs` - Conditions météo
- `app_versions` - Versions d'application
- `dashboard_configs` - Configuration dashboards
- `user_limits` - Limites et quotas

#### 🏆 **Gamification & Motivation**
- `driver_bonuses` - Système de bonus
- `driver_penalties` - Gestion des sanctions
- `emergency_alerts` - Alertes d'urgence

#### 📁 **Fichiers & Médias**
- `files` - Gestion des fichiers
- `user_addresses` - Adresses sauvegardées

---

## 🔗 **Relations Principales**

### **users (Utilisateur central)**
```
users -> vehicles (1:1)
users -> wallets (1:1)
users -> user_addresses (1:N)
users -> deliveries (1:N) [as client]
users -> deliveries (1:N) [as livreur]
users -> notifications (1:N)
users -> payments (1:N)
users -> referrals (1:N)
```

### **deliveries (Livraison centrale)**
```
deliveries -> locations (2:1) [pickup + dropoff]
deliveries -> packages (1:N)
deliveries -> trackings (1:N)
deliveries -> payments (1:1)
deliveries -> grouped_deliveries (1:1) [if grouped]
deliveries -> moving_services (1:1) [if moving]
```

### **grouped_deliveries (Livraisons groupées)**
```
grouped_deliveries -> delivery_zones (1:N)
delivery_zones -> locations (2:1) [pickup + delivery]
```

---

## ⚡ **Avantages de Prisma + PostgreSQL**

### **Performance**
- ⚡ **PostgreSQL optimisé** (transactions ACID, concurrence)
- ⚡ **Prisma Query Engine** (requêtes optimisées automatiquement)
- ⚡ **Connection pooling** (gestion connexions efficace) 
- ⚡ **Index automatiques** (sur foreign keys et contraintes)

### **Développement**
- 🔧 **Type safety** (modèles Go générés automatiquement)
- 🔧 **Migrations automatiques** (schema evolution sécurisée)
- 🔧 **IDE auto-completion** (intellisense sur toutes les requêtes)
- 🔧 **Relations type-safe** (pas d'erreurs de jointures)

### **Scalabilité**
- 📈 **PostgreSQL production-ready** (millions de records)
- 📈 **Réplication master-slave** (haute disponibilité)
- 📈 **SQLite développement** (tests rapides locaux)

---

## 🚀 **Architecture Moderne Type-Safe**

### **🟢 Prisma ORM (Type-safe database access)**
```go
// CRUD ultra-sécurisés avec auto-completion
delivery, err := db.PrismaDB.Delivery.CreateOne(
    prismadb.Delivery.ClientPhone.Set(clientID),
    prismadb.Delivery.Type.Set(prismadb.DeliveryTypeStandard),
    prismadb.Delivery.PickupLocation.Link(
        prismadb.Location.ID.Equals(pickupID),
    ),
).Exec(ctx)

// Requêtes type-safe avec relations
deliveries, err := db.PrismaDB.Delivery.FindMany(
    prismadb.Delivery.ClientPhone.Equals(clientID),
).With(
    prismadb.Delivery.PickupLocation.Fetch(),
    prismadb.Delivery.DropoffLocation.Fetch(),
).Exec(ctx)
```

### **🟡 PostgreSQL (Production robuste)**
```sql
-- Transactions ACID garanties
BEGIN;
INSERT INTO deliveries (...) VALUES (...);
UPDATE users SET total_deliveries = total_deliveries + 1;
COMMIT;

-- Index automatiques sur foreign keys
-- Performance optimisée pour millions de records
```

### **🔴 Go Backend (Logique complexe)**
```go
// Calculs complexes avec validation Prisma
func CalculateComplexPrice(delivery *models.Delivery) float64 {
    // Algorithme de pricing avancé avec type safety
}

// Assignation intelligente
func AutoAssignDelivery(deliveryID string) error {
    // IA d'assignation avec validations Prisma
}
```

---

## 📊 **Statistiques de la Base**

- **43 tables** principales
- **200+ champs** au total
- **60+ index** pour l'optimisation
- **30+ relations** complexes
- **15+ types d'énumérations**
- **Cache LMDB** avec TTL automatique

---

## 🎯 **Prochaines Étapes**

1. **Documentation détaillée** de chaque table SQLite
2. **Exemples API REST** pour chaque opération
3. **Guide d'intégration** frontend
4. **Patterns d'optimisation** avancés
5. **Tests de performance** SQLite + LMDB

---

**Cette architecture hybride vous donne le meilleur des deux mondes : la rapidité de SQLite pour les opérations simples, la puissance de LMDB pour le cache, et la robustesse de Go pour la logique complexe !** 🚀
