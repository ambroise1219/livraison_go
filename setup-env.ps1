# Script PowerShell pour configurer l'environnement de développement ILEX
# Usage: .\setup-env.ps1

Write-Host "🚀 Configuration de l'environnement ILEX Backend" -ForegroundColor Green

# Vérifier si .env existe déjà
if (Test-Path ".env") {
    $overwrite = Read-Host ".env existe déjà. Voulez-vous l'écraser? (y/N)"
    if ($overwrite -ne "y" -and $overwrite -ne "Y") {
        Write-Host "❌ Configuration annulée" -ForegroundColor Yellow
        exit
    }
}

# Demander les informations de configuration
Write-Host "`n📋 Configuration SurrealDB:" -ForegroundColor Cyan
$surrealUrl = Read-Host "URL SurrealDB (défaut: ws://localhost:8000/rpc)"
if ([string]::IsNullOrEmpty($surrealUrl)) { $surrealUrl = "ws://localhost:8000/rpc" }

$surrealUser = Read-Host "Nom d'utilisateur SurrealDB (défaut: root)"
if ([string]::IsNullOrEmpty($surrealUser)) { $surrealUser = "root" }

$surrealPass = Read-Host "Mot de passe SurrealDB (défaut: root)"
if ([string]::IsNullOrEmpty($surrealPass)) { $surrealPass = "root" }

$namespace = Read-Host "Namespace (défaut: ilex)"
if ([string]::IsNullOrEmpty($namespace)) { $namespace = "ilex" }

$database = Read-Host "Database (défaut: production)"
if ([string]::IsNullOrEmpty($database)) { $database = "production" }

Write-Host "`n🔐 Configuration JWT:" -ForegroundColor Cyan
$jwtSecret = Read-Host "JWT Secret (laissez vide pour générer automatiquement)"
if ([string]::IsNullOrEmpty($jwtSecret)) { 
    $jwtSecret = [System.Web.Security.Membership]::GeneratePassword(64, 16)
    Write-Host "✅ JWT Secret généré automatiquement" -ForegroundColor Green
}

# Générer le fichier .env
$envContent = @"
# ILEX Backend Environment Configuration
# Generated on $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
ENVIRONMENT=development
DEBUG=true

# SurrealDB Configuration
SURREALDB_URL=$surrealUrl
SURREALDB_USERNAME=$surrealUser
SURREALDB_PASSWORD=$surrealPass
SURREALDB_NS=$namespace
SURREALDB_DB=$database

# JWT Configuration
JWT_SECRET=$jwtSecret
JWT_EXPIRATION=24

# OTP Configuration
OTP_EXPIRATION=5
OTP_LENGTH=6

# SMS Configuration (pour notifications OTP)
SMS_API_KEY=
SMS_API_SECRET=

# Email Configuration (pour notifications)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=

# Platform Commission Settings
DEFAULT_COMMISSION_RATE=0.15
DEFAULT_SERVICE_FEE=500.0

# Referral Settings
REFERRAL_REWARD_AMOUNT=1000.0
REFERRAL_EXPIRATION=30
"@

# Écrire le fichier
$envContent | Out-File -FilePath ".env" -Encoding UTF8

Write-Host "`n✅ Fichier .env créé avec succès!" -ForegroundColor Green
Write-Host "📝 N'oubliez pas de configurer vos clés API SMS/Email si nécessaire" -ForegroundColor Yellow
Write-Host "🚀 Vous pouvez maintenant lancer: go run main.go" -ForegroundColor Cyan

# Vérifier si Go est installé
try {
    $goVersion = & go version 2>$null
    Write-Host "`n✅ Go détecté: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "`n⚠️  Go n'est pas installé ou pas dans le PATH" -ForegroundColor Red
    Write-Host "📥 Téléchargez Go depuis: https://golang.org/dl/" -ForegroundColor Yellow
}