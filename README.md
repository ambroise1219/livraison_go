# ILEX Backend Go

Backend de l'application ILEX de livraison développé en Go avec SurrealDB.

## 🚀 Fonctionnalités

- 🔐 Authentification JWT avec OTP par SMS
- 👥 Gestion des rôles (CLIENT, LIVREUR, ADMIN)
- 📦 Gestion des livraisons avec suivi en temps réel
- 🚗 Gestion des véhicules pour livreurs
- 💰 Système de promotions et parrainage
- 📊 Tableaux de bord administrateurs
- 🔒 API REST sécurisée avec middlewares
- 🎯 Architecture propre et modulaire

## 📋 Prérequis

- **Go 1.19+** - [Télécharger Go](https://golang.org/dl/)
- **SurrealDB** - [Installation SurrealDB](https://surrealdb.com/docs/installation)
- **Git** - Pour cloner le projet

## 🛠️ Installation

### 1. Cloner le projet
```bash
git clone https://github.com/ambroise1219/livraison_go.git
cd livraison_go
```

### 2. Installer Go (si nécessaire)

#### Windows (avec Chocolatey)
```powershell
# En tant qu'administrateur
choco install golang -y
```

#### Windows (manuel)
1. Télécharger depuis https://golang.org/dl/
2. Exécuter l'installateur
3. Redémarrer le terminal

#### Linux/macOS
```bash
# Via le gestionnaire de paquets ou depuis golang.org
# Ubuntu/Debian
sudo apt install golang-go

# macOS (avec Homebrew)
brew install go
```

### 3. Vérifier l'installation de Go
```bash
go version
# Devrait afficher: go version go1.x.x
```

### 4. Installer les dépendances
```bash
go mod download
go mod tidy
```

### 5. Installer et démarrer SurrealDB

#### Avec Docker (recommandé)
```bash
docker run --name surrealdb -d -p 8000:8000 surrealdb/surrealdb:latest start --log trace --user root --pass root memory
```

#### Installation manuelle
```bash
# Voir https://surrealdb.com/docs/installation
```

## ⚙️ Configuration

### Variables d'environnement

Créer un fichier `.env` à la racine du projet :

```env
# Serveur
SERVER_HOST=localhost
SERVER_PORT=8080
ENVIRONMENT=development

# SurrealDB
SURREALDB_URL=ws://localhost:8000
SURREALDB_NAMESPACE=ilex
SURREALDB_DATABASE=livraison
SURREALDB_USERNAME=root
SURREALDB_PASSWORD=root

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY_HOURS=24
JWT_REFRESH_EXPIRY_DAYS=7

# OTP
OTP_EXPIRY_MINUTES=5

# SMS (remplacer par vos vraies clés)
SMS_API_KEY=your-sms-api-key
SMS_API_URL=https://api.sms-provider.com/send
SMS_SENDER=ILEX

# Email (remplacer par vos vrais paramètres)
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
```

### Configuration par défaut

Si aucun fichier `.env` n'est fourni, les valeurs par défaut de développement sont utilisées (voir `main.go`).

## 🚀 Démarrage

### 1. Démarrer SurrealDB
```bash
# Si utilisant Docker
docker start surrealdb

# Si installation manuelle
surreal start --log trace --user root --pass root memory
```

### 2. Lancer le backend
```bash
# Développement
go run main.go

# Ou compiler puis exécuter
go build -o ilex-backend
./ilex-backend
```

Le serveur démarrera sur `http://localhost:8080`

## 📚 API Endpoints

### 🔐 Authentification
```
POST /api/v1/auth/otp/send      - Envoyer OTP
POST /api/v1/auth/otp/verify    - Vérifier OTP et se connecter
POST /api/v1/auth/refresh       - Rafraîchir le token
POST /api/v1/auth/logout        - Se déconnecter
GET  /api/v1/auth/profile       - Profil utilisateur
```

### 📦 Livraisons
```
POST /api/v1/delivery/                    - Créer livraison (CLIENT)
GET  /api/v1/delivery/:id                 - Détails livraison
POST /api/v1/delivery/price/calculate     - Calculer prix (public)
PATCH /api/v1/delivery/:id/status         - Mettre à jour statut (LIVREUR/ADMIN)
```

### 🚚 Livreurs
```
GET  /api/v1/delivery/driver/available    - Livraisons disponibles
GET  /api/v1/delivery/driver/assigned     - Livraisons assignées
POST /api/v1/delivery/driver/:id/accept   - Accepter livraison
POST /api/v1/delivery/driver/:id/location - Mettre à jour position
```

### 👥 Utilisateurs
```
GET  /api/v1/users/:id                    - Profil utilisateur
PUT  /api/v1/users/:id                    - Mettre à jour profil
GET  /api/v1/users/:id/deliveries         - Historique livraisons
```

### 🎁 Promotions
```
POST /api/v1/promo/validate               - Valider code promo
POST /api/v1/promo/use                    - Utiliser code promo
GET  /api/v1/promo/history                - Historique promos
```

### 👑 Administration (ADMIN uniquement)
```
GET  /api/v1/admin/users                  - Liste utilisateurs
GET  /api/v1/admin/deliveries             - Liste livraisons
GET  /api/v1/admin/drivers                - Liste livreurs
GET  /api/v1/admin/stats/dashboard        - Statistiques dashboard
```

## 🧪 Tests

```bash
# Lancer les tests
go test ./...

# Tests avec couverture
go test -cover ./...

# Tests détaillés
go test -v ./...
```

## 🏗️ Architecture

```
├── cmd/                 # Points d'entrée
├── config/             # Configuration
├── db/                 # Connexion base de données
├── handlers/           # Contrôleurs HTTP
├── middlewares/        # Middlewares (auth, CORS, etc.)
├── models/             # Modèles de données
├── routes/             # Définition des routes
├── services/           # Logique métier
├── tests/              # Tests
└── main.go            # Point d'entrée principal
```

## 🔧 Développement

### Ajouter un nouveau endpoint

1. **Définir le modèle** dans `models/`
2. **Créer le service** dans `services/`
3. **Ajouter le handler** dans `handlers/`
4. **Configurer la route** dans `routes/routes.go`
5. **Écrire les tests** dans `tests/`

### Middlewares disponibles

- `AuthMiddleware()` - Authentification JWT requise
- `RequireRole(roles...)` - Vérification de rôles
- `RequireAdmin()` - Admin uniquement
- `RequireDriver()` - Livreur uniquement
- `RequireClient()` - Client uniquement
- `CORSMiddleware()` - Gestion CORS
- `RateLimitMiddleware()` - Limitation de débit

## 📊 Monitoring

### Health Check
```bash
curl http://localhost:8080/health
```

### Logs
Les logs sont affichés dans la console avec codes couleur selon le niveau.

## 🔒 Sécurité

- Tokens JWT avec expiration
- Middleware de validation des rôles
- Rate limiting par IP
- Headers de sécurité (XSS, CSRF, etc.)
- Validation des données d'entrée
- Hashage sécurisé des mots de passe

## 🚀 Déploiement

### Production

1. Définir `ENVIRONMENT=production`
2. Configurer les vraies variables d'environnement
3. Utiliser HTTPS
4. Configurer un reverse proxy (nginx)
5. Mettre en place la supervision

### Docker (à venir)
```bash
docker build -t ilex-backend .
docker run -p 8080:8080 ilex-backend
```

## 🤝 Contribution

1. Fork le projet
2. Créer une branche feature (`git checkout -b feature/AmazingFeature`)
3. Commit les changes (`git commit -m 'Add some AmazingFeature'`)
4. Créer une Pull Request

## 📝 License

Ce projet est sous licence MIT. Voir le fichier [LICENSE](LICENSE) pour plus de détails.

## 📞 Support

- 📧 Email: support@ilex.com
- 📱 GitHub Issues: [Issues](https://github.com/ambroise1219/livraison_go/issues)
- 💬 Discord: [Lien Discord]

---

**Développé avec ❤️ par l'équipe ILEX**