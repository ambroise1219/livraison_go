package main

import (
	"fmt"
	"log"

	"github.com/ambroise1219/livraison_go/database"
	"github.com/ambroise1219/livraison_go/db"
)

func main() {
	fmt.Println("🔌 Test de connexion à PostgreSQL...")

	// Initialiser la connexion Prisma
	err := database.InitPrisma()
	if err != nil {
		log.Fatalf("❌ Erreur de connexion à PostgreSQL: %v", err)
	}
	
	fmt.Println("✅ Connexion PostgreSQL réussie!")

	// Initialiser le wrapper DB
	err = db.InitializePrisma()
	if err != nil {
		log.Fatalf("❌ Erreur d'initialisation du wrapper DB: %v", err)
	}
	
	fmt.Println("✅ Wrapper DB initialisé!")

	// Test des statistiques de la base
	stats, err := db.GetDatabaseStats()
	if err != nil {
		log.Printf("⚠️ Erreur lors du test de connexion: %v", err)
	} else {
		fmt.Printf("✅ Statistiques DB: %+v\n", stats)
	}

	// Fermeture propre
	db.ClosePrisma()
	database.ClosePrisma()
	fmt.Println("✅ Connexions fermées proprement!")

	fmt.Println("🎉 Test de connexion terminé avec succès!")
}