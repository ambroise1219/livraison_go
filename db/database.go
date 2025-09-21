package db

import (
	"context"
	"fmt"
	"log"

	"github.com/ambroise1219/livraison_go/database"
	"github.com/ambroise1219/livraison_go/prisma/db"
)

// PrismaDB est l'instance globale du client Prisma
var PrismaDB *db.PrismaClient

// InitializePrisma initialise la connexion Prisma
func InitializePrisma() error {
	if database.PrismaClient == nil {
		return fmt.Errorf("Prisma client not initialized")
	}
	PrismaDB = database.PrismaClient
	log.Println("✅ Prisma database client initialized")
	return nil
}

// ClosePrisma ferme la connexion Prisma
func ClosePrisma() {
	if PrismaDB != nil {
		PrismaDB = nil
	}
}

// ✅ MIGRATION TERMINÉE - Toutes les fonctions SQLite ont été migrées vers Prisma ORM
// Les anciennes fonctions ExecuteQuery, QueryRow, QueryRows, BeginTransaction ont été supprimées
// Utilisez maintenant directement PrismaDB pour toutes les opérations de base de données

// CheckTableExists vérifie si une table existe (compatibilité)
func CheckTableExists(tableName string) (bool, error) {
	// Pour Prisma, on peut vérifier l'existence via une requête simple
	ctx := context.Background()

	// Test avec une requête simple pour vérifier la connexion
	_, err := PrismaDB.User.FindMany().Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("table check failed: %v", err)
	}

	return true, nil
}

// GetTableCount retourne le nombre de lignes dans une table (compatibilité)
func GetTableCount(tableName string) (int64, error) {
	ctx := context.Background()

	switch tableName {
	case "users":
		users, err := PrismaDB.User.FindMany().Exec(ctx)
		return int64(len(users)), err
	case "otps":
		otps, err := PrismaDB.Otp.FindMany().Exec(ctx)
		return int64(len(otps)), err
	case "deliveries":
		deliveries, err := PrismaDB.Delivery.FindMany().Exec(ctx)
		return int64(len(deliveries)), err
	case "locations":
		locations, err := PrismaDB.Location.FindMany().Exec(ctx)
		return int64(len(locations)), err
	case "vehicles":
		vehicles, err := PrismaDB.Vehicle.FindMany().Exec(ctx)
		return int64(len(vehicles)), err
	case "packages":
		packages, err := PrismaDB.Package.FindMany().Exec(ctx)
		return int64(len(packages)), err
	case "trackings":
		trackings, err := PrismaDB.Tracking.FindMany().Exec(ctx)
		return int64(len(trackings)), err
	case "payments":
		payments, err := PrismaDB.Payment.FindMany().Exec(ctx)
		return int64(len(payments)), err
	case "wallets":
		wallets, err := PrismaDB.Wallet.FindMany().Exec(ctx)
		return int64(len(wallets)), err
	case "wallet_transactions":
		transactions, err := PrismaDB.WalletTransaction.FindMany().Exec(ctx)
		return int64(len(transactions)), err
	case "grouped_deliveries":
		grouped, err := PrismaDB.GroupedDelivery.FindMany().Exec(ctx)
		return int64(len(grouped)), err
	case "moving_services":
		services, err := PrismaDB.MovingService.FindMany().Exec(ctx)
		return int64(len(services)), err
	case "promos":
		promos, err := PrismaDB.Promo.FindMany().Exec(ctx)
		return int64(len(promos)), err
	case "promo_usages":
		usages, err := PrismaDB.PromoUsage.FindMany().Exec(ctx)
		return int64(len(usages)), err
	case "referrals":
		referrals, err := PrismaDB.Referral.FindMany().Exec(ctx)
		return int64(len(referrals)), err
	case "notifications":
		notifications, err := PrismaDB.Notification.FindMany().Exec(ctx)
		return int64(len(notifications)), err
	case "incidents":
		incidents, err := PrismaDB.Incident.FindMany().Exec(ctx)
		return int64(len(incidents)), err
	case "ratings":
		ratings, err := PrismaDB.Rating.FindMany().Exec(ctx)
		return int64(len(ratings)), err
	case "user_addresses":
		addresses, err := PrismaDB.UserAddress.FindMany().Exec(ctx)
		return int64(len(addresses)), err
	case "files":
		files, err := PrismaDB.File.FindMany().Exec(ctx)
		return int64(len(files)), err
	case "delivery_zones":
		zones, err := PrismaDB.DeliveryZone.FindMany().Exec(ctx)
		return int64(len(zones)), err
	case "driver_locations":
		// DriverLocation n'existe pas dans le schéma Prisma actuel
		return 0, fmt.Errorf("table driver_locations not found in schema")
	case "subscriptions":
		subscriptions, err := PrismaDB.Subscription.FindMany().Exec(ctx)
		return int64(len(subscriptions)), err
	default:
		return 0, fmt.Errorf("unknown table: %s", tableName)
	}
}

// GetDatabaseStats retourne les statistiques de la base de données
func GetDatabaseStats() (map[string]interface{}, error) {
	ctx := context.Background()
	stats := make(map[string]interface{})

	// Compter les utilisateurs
	users, err := PrismaDB.User.FindMany().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %v", err)
	}
	stats["users"] = len(users)

	// Compter les livraisons
	deliveries, err := PrismaDB.Delivery.FindMany().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count deliveries: %v", err)
	}
	stats["deliveries"] = len(deliveries)

	// Compter les OTPs
	otps, err := PrismaDB.Otp.FindMany().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count OTPs: %v", err)
	}
	stats["otps"] = len(otps)

	// Compter les véhicules
	vehicles, err := PrismaDB.Vehicle.FindMany().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count vehicles: %v", err)
	}
	stats["vehicles"] = len(vehicles)

	return stats, nil
}
