# 📚 Collections Postman ILEX Backend

## 🎯 Vue d'ensemble

Ce dossier contient des collections Postman complètes pour tester tous les endpoints du backend ILEX - Livraison Go.

## 📁 Collections disponibles

### 1. **ILEX_Backend_Complete.json** - Collection principale
- 🏠 **System & Health** : Endpoints de base (root, health, stats, db)
- 🔐 **Authentication** : OTP, JWT, refresh token, logout
- 👤 **User Profile** : Profil utilisateur, upload photo, test Cloudinary

### 2. **Delivery_Operations.json** - Opérations de livraison
- 📦 **Delivery Management** : Création, récupération, mise à jour, assignation
- 🚚 **Driver Operations** : Livraisons disponibles, acceptation, localisation
- 👥 **Client Operations** : Livraisons client, annulation, suivi

### 3. **Admin_Management.json** - Gestion administrateur
- 👥 **User Management** : Gestion des utilisateurs, rôles, suppression
- 📦 **Delivery Management** : Gestion des livraisons, statistiques
- 🚚 **Driver Management** : Gestion des livreurs, statuts
- 🎁 **Promotion Management** : Création, modification, suppression des promotions
- 🚗 **Vehicle Management** : Gestion des véhicules, vérification
- 📊 **Statistics** : Tableau de bord, revenus, utilisateurs

### 4. **Realtime_Support.json** - Temps réel et support
- ⚡ **Realtime Operations** : SSE, WebSocket, localisation, notifications
- 🎫 **Support Tickets** : Création, gestion, messages des tickets
- 💬 **Internal Chat** : Groupes internes, messages
- 📞 **Admin Support** : Contact direct administrateur

### 5. **Promotions_Users.json** - Promotions et utilisateurs
- 🎁 **Promotions** : Validation, utilisation, historique, parrainage
- 👥 **User Management** : Profils, livraisons, véhicules

## 🚀 Configuration

### Variables d'environnement
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

### Import dans Postman
1. Ouvrir Postman
2. Cliquer sur "Import"
3. Sélectionner les fichiers JSON
4. Configurer les variables d'environnement

## 🔄 Workflow de test

### 1. Authentification
```bash
# 1. Envoyer OTP
POST /api/v1/auth/otp/send
{
  "phone": "+2250173226070"
}

# 2. Vérifier OTP
POST /api/v1/auth/otp/verify
{
  "phone": "+2250173226070",
  "code": "123456"
}

# 3. Récupérer le token et le mettre dans ACCESS_TOKEN
```

### 2. Test des fonctionnalités
- **Profil utilisateur** : GET/PUT /api/v1/auth/profile
- **Upload photo** : POST /api/v1/auth/profile/picture
- **Création livraison** : POST /api/v1/delivery/
- **Gestion admin** : Utiliser les endpoints /api/v1/admin/*

## 📋 Endpoints par module

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

## 🏠 System (4 endpoints)
- `GET /` - Page racine
- `GET /health` - Health check
- `GET /stats` - Statistiques performance
- `GET /db` - Statistiques base de données

## 📊 Total : 80+ endpoints

## 🔧 Notes techniques

### Authentification
- Tous les endpoints protégés nécessitent un header `Authorization: Bearer <token>`
- Le token est obtenu via le flow OTP

### Rôles
- **CLIENT** : Peut créer des livraisons, gérer son profil
- **DRIVER** : Peut accepter des livraisons, mettre à jour sa localisation
- **ADMIN** : Accès complet à tous les endpoints
- **STAFF** : Accès aux fonctions de support

### Rate Limiting
- 100 requêtes par minute par défaut
- Headers de sécurité automatiques

### CORS
- Configuré pour accepter les requêtes depuis l'app mobile

## 🚀 Démarrage rapide

1. **Importer les collections** dans Postman
2. **Configurer les variables** d'environnement
3. **Tester l'authentification** (OTP flow)
4. **Explorer les fonctionnalités** par module

## 📱 Intégration mobile

Ces collections servent de référence pour l'intégration avec l'app Expo React Native. Tous les endpoints sont documentés avec les formats de requête/réponse attendus.