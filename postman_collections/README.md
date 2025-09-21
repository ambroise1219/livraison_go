# Collections Postman - ILEX Livraison API

Ce dossier contient des collections Postman complètes pour tester l'API ILEX Livraison.

## 📁 Collections Disponibles

### 1. `auth_otp_jwt.json` - 🔐 Authentification
- Envoi d'OTP via WhatsApp
- Vérification OTP et génération JWT
- Gestion des profils utilisateur

### 2. `delivery_realtime_and_core.json` - 🚚 Livraisons
- CRUD des livraisons
- Temps réel via Server-Sent Events (SSE)
- WebSocket pour chat client-livreur

### 3. `driver_operations.json` - 🏍️ Opérations Livreurs
- Tableau de bord livreur
- Acceptation de livraisons
- Tracking GPS en temps réel

### 4. `admin_management.json` - 👨‍💼 Administration
- Gestion utilisateurs (CRUD, rôles)
- Analytics et rapports
- Monitoring système

### 5. `promos_pricing.json` - 💰 Prix & Promotions
- Calcul de prix dynamique
- Validation de codes promotionnels
- Application de remises

## 🚀 Utilisation

1. **Importer** les collections dans Postman
2. **Configurer** les variables d'environnement :
   - `BASE_URL` : `http://localhost:8080`
   - `CLIENT_PHONE` : `+2250701234567`
   - etc.

3. **Commencer** par la collection Auth pour obtenir des tokens
4. **Tester** les workflows dans l'ordre des collections

## 📊 Variables

Chaque collection utilise des variables automatiques qui s'extraient des réponses :
- `ACCESS_TOKEN` : Token JWT
- `DELIVERY_ID` : ID de livraison
- `USER_ID` : ID utilisateur

---

> **💡 Conseil** : Commencez par `auth_otp_jwt.json` pour créer vos premiers comptes utilisateur.

> **⚠️ Important** : Ces collections sont conçues pour un environnement de développement local.