# 📋 Résumé Complet des Endpoints ILEX Backend

## 🎯 Vue d'ensemble
**Total : 80+ endpoints** répartis sur **8 modules principaux**

## 📊 Répartition par module

| Module | Endpoints | Description |
|--------|-----------|-------------|
| 🏠 **System** | 4 | Health, stats, base de données |
| 🔐 **Authentication** | 4 | OTP, JWT, refresh, logout |
| 👤 **User Profile** | 4 | Profil, upload photo, test Cloudinary |
| 📦 **Delivery** | 15 | Création, gestion, suivi des livraisons |
| 🎁 **Promotions** | 5 | Codes promo, parrainage, historique |
| 👥 **Users** | 6 | Gestion utilisateurs, véhicules |
| 👑 **Admin** | 20 | Administration complète |
| ⚡ **Realtime** | 8 | SSE, WebSocket, notifications |
| 🎫 **Support** | 9 | Tickets, chat interne |
| 💬 **Internal Chat** | 4 | Groupes, messages internes |
| 📞 **Admin Support** | 1 | Contact direct administrateur |

## 🔗 URLs de base
- **Base URL** : `http://127.0.0.1:3000`
- **API Version** : `v1`
- **Prefix** : `/api/v1`

## 🔐 Authentification
- **Type** : JWT Bearer Token
- **Header** : `Authorization: Bearer <token>`
- **Flow** : OTP → Verification → Token

## 📱 Collections Postman

### 1. **ILEX_Backend_Complete.json**
- System & Health (4 endpoints)
- Authentication (4 endpoints)  
- User Profile (4 endpoints)

### 2. **Delivery_Operations.json**
- Delivery Management (5 endpoints)
- Driver Operations (4 endpoints)
- Client Operations (3 endpoints)

### 3. **Admin_Management.json**
- User Management (4 endpoints)
- Delivery Management (3 endpoints)
- Driver Management (3 endpoints)
- Promotion Management (5 endpoints)
- Vehicle Management (2 endpoints)
- Statistics (3 endpoints)

### 4. **Realtime_Support.json**
- Realtime Operations (8 endpoints)
- Support Tickets (9 endpoints)
- Internal Chat (4 endpoints)
- Admin Support (1 endpoint)

### 5. **Promotions_Users.json**
- Promotions (5 endpoints)
- User Management (6 endpoints)

## 🚀 Workflow de test

### 1. Authentification
```bash
# Envoyer OTP
POST /api/v1/auth/otp/send
{
  "phone": "+2250173226070"
}

# Vérifier OTP
POST /api/v1/auth/otp/verify
{
  "phone": "+2250173226070",
  "code": "123456"
}
```

### 2. Utilisation du token
```bash
# Tous les endpoints protégés
Authorization: Bearer <token>
```

## 📋 Endpoints détaillés

### 🏠 System (4 endpoints)
- `GET /` - Page racine
- `GET /health` - Health check
- `GET /stats` - Statistiques performance
- `GET /db` - Statistiques base de données

### 🔐 Authentication (4 endpoints)
- `POST /api/v1/auth/otp/send` - Envoyer OTP
- `POST /api/v1/auth/otp/verify` - Vérifier OTP
- `POST /api/v1/auth/refresh` - Rafraîchir token
- `POST /api/v1/auth/logout` - Déconnexion

### 👤 User Profile (4 endpoints)
- `GET /api/v1/auth/profile` - Récupérer profil
- `PUT /api/v1/auth/profile` - Mettre à jour profil
- `POST /api/v1/auth/profile/picture` - Upload photo
- `GET /api/v1/auth/test/cloudinary` - Test Cloudinary

### 📦 Delivery (15 endpoints)
- `POST /api/v1/delivery/` - Créer livraison
- `GET /api/v1/delivery/:id` - Récupérer livraison
- `PATCH /api/v1/delivery/:id/status` - Mettre à jour statut
- `POST /api/v1/delivery/:id/assign` - Assigner livraison
- `POST /api/v1/delivery/price/calculate` - Calculer prix
- `GET /api/v1/delivery/driver/available` - Livraisons disponibles
- `GET /api/v1/delivery/driver/assigned` - Livraisons assignées
- `POST /api/v1/delivery/driver/:id/accept` - Accepter livraison
- `POST /api/v1/delivery/driver/:id/location` - Mettre à jour localisation
- `GET /api/v1/delivery/client/` - Livraisons client
- `POST /api/v1/delivery/client/:id/cancel` - Annuler livraison
- `GET /api/v1/delivery/client/:id/track` - Suivre livraison

### 🎁 Promotions (5 endpoints)
- `POST /api/v1/promo/validate` - Valider code promo
- `POST /api/v1/promo/use` - Utiliser code promo
- `GET /api/v1/promo/history` - Historique promotions
- `POST /api/v1/promo/referral/create` - Créer parrainage
- `GET /api/v1/promo/referral/stats` - Statistiques parrainage

### 👥 Users (6 endpoints)
- `GET /api/v1/users/:id` - Profil utilisateur
- `PUT /api/v1/users/:id` - Mettre à jour profil
- `GET /api/v1/users/:id/deliveries` - Livraisons utilisateur
- `GET /api/v1/users/:id/vehicles` - Véhicules utilisateur
- `POST /api/v1/users/:id/vehicles` - Créer véhicule
- `PUT /api/v1/users/:id/vehicles/:id` - Mettre à jour véhicule

### 👑 Admin (20 endpoints)
- **Users** : GET, GET/:id, PUT/:id/role, DELETE/:id
- **Deliveries** : GET, GET/stats, POST/:id/assign/:driver_id
- **Drivers** : GET, GET/:id/stats, PUT/:id/status
- **Promotions** : GET, POST, PUT/:id, DELETE/:id, GET/:id/stats
- **Vehicles** : GET, PUT/:id/verify
- **Stats** : GET/dashboard, GET/revenue, GET/users

### ⚡ Realtime (8 endpoints)
- `GET /api/v1/sse/delivery/:id` - SSE suivi livraison
- `GET /api/v1/ws/chat/:id` - WebSocket chat
- `POST /api/v1/realtime/location/:driver_id/:delivery_id` - Mettre à jour localisation
- `GET /api/v1/realtime/location/:driver_id` - Récupérer localisation
- `POST /api/v1/realtime/delivery/:id/status` - Mettre à jour statut
- `POST /api/v1/realtime/notification/:user_id` - Envoyer notification
- `GET /api/v1/realtime/eta/:id` - Calculer ETA
- `GET /api/v1/realtime/stats` - Statistiques temps réel

### 🎫 Support (9 endpoints)
- `POST /api/v1/support/tickets/` - Créer ticket
- `GET /api/v1/support/tickets/` - Lister tickets
- `GET /api/v1/support/tickets/:id` - Détails ticket
- `PUT /api/v1/support/tickets/:id/status` - Mettre à jour statut
- `POST /api/v1/support/tickets/:id/messages` - Ajouter message
- `GET /api/v1/support/tickets/:id/messages` - Récupérer messages
- `GET /api/v1/support/tickets/:id/history` - Historique réassignations
- `POST /api/v1/support/tickets/:id/reassign` - Réassigner ticket
- `GET /api/v1/support/stats` - Statistiques support

### 💬 Internal Chat (4 endpoints)
- `POST /api/v1/internal/groups/` - Créer groupe
- `GET /api/v1/internal/groups/` - Lister groupes
- `GET /api/v1/internal/groups/:id/messages` - Messages groupe
- `POST /api/v1/internal/groups/:id/messages` - Ajouter message

### 📞 Admin Support (1 endpoint)
- `POST /api/v1/admin/support/contact` - Contact direct

## 🔧 Configuration

### Variables d'environnement Postman
```json
{
  "BASE_URL": "http://127.0.0.1:3000",
  "CLIENT_PHONE": "+2250173226070",
  "OTP_CODE": "123456",
  "ACCESS_TOKEN": "",
  "USER_ID": "",
  "DELIVERY_ID": "",
  "DRIVER_ID": "",
  "VEHICLE_ID": "",
  "TICKET_ID": "",
  "GROUP_ID": "",
  "PROMO_ID": ""
}
```

### Rôles et permissions
- **CLIENT** : Création livraisons, gestion profil
- **DRIVER** : Acceptation livraisons, mise à jour localisation
- **ADMIN** : Accès complet à tous les endpoints
- **STAFF** : Accès aux fonctions de support

## 🚀 Démarrage rapide

1. **Importer les collections** Postman
2. **Configurer les variables** d'environnement
3. **Tester l'authentification** (OTP flow)
4. **Explorer les fonctionnalités** par module

## 📱 Intégration mobile

Ces endpoints sont prêts pour l'intégration avec l'app Expo React Native. Tous les formats de requête/réponse sont documentés dans les collections Postman.

## 🧪 Tests

- **Script de test** : `./test_all_endpoints.sh`
- **Collections Postman** : 5 collections complètes
- **Documentation** : README détaillé dans `postman_collections/`

## ✅ Status

- **Backend** : ✅ Opérationnel sur port 3000
- **Redis** : ✅ Intégré et fonctionnel
- **Cloudinary** : ✅ Upload de photos configuré
- **Collections Postman** : ✅ Complètes et mises à jour
- **Documentation** : ✅ Complète de A à Z
