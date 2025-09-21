package database

import (
	"context"
	"fmt"
	"log"

	"github.com/ambroise1219/livraison_go/prisma/db"
)

var PrismaClient *db.PrismaClient

// InitPrisma initialise la connexion Prisma
func InitPrisma() error {
	PrismaClient = db.NewClient()

	if err := PrismaClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test de connexion
	ctx := context.Background()
	_, err := PrismaClient.User.FindMany().Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to test database connection: %w", err)
	}

	log.Println("✅ Connexion PostgreSQL via Prisma établie")
	return nil
}

// ClosePrisma ferme la connexion Prisma
func ClosePrisma() error {
	if PrismaClient != nil {
		return PrismaClient.Disconnect()
	}
	return nil
}

// GetPrismaClient retourne le client Prisma
func GetPrismaClient() *db.PrismaClient {
	return PrismaClient
}

// GetDatabaseStats retourne les statistiques de la base de données
func GetDatabaseStats() (map[string]interface{}, error) {
	ctx := context.Background()
	stats := make(map[string]interface{})

	// Compter les utilisateurs
	users, err := PrismaClient.User.FindMany().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}
	stats["users"] = len(users)

	// Compter les livraisons
	deliveries, err := PrismaClient.Delivery.FindMany().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count deliveries: %w", err)
	}
	stats["deliveries"] = len(deliveries)

	// Compter les OTPs
	otps, err := PrismaClient.Otp.FindMany().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count otps: %w", err)
	}
	stats["otps"] = len(otps)

	// Compter les véhicules
	vehicles, err := PrismaClient.Vehicle.FindMany().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count vehicles: %w", err)
	}
	stats["vehicles"] = len(vehicles)

	return stats, nil
}
