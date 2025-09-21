package main

import (
	"fmt"
	"log"

	"github.com/ambroise1219/livraison_go/database"
	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	simpleCreationService "github.com/ambroise1219/livraison_go/services/delivery"
)

func main() {
	fmt.Println("📍 Test du module Location...")

	// Initialiser la connexion
	err := database.InitPrisma()
	if err != nil {
		log.Fatalf("❌ Erreur de connexion: %v", err)
	}
	defer database.ClosePrisma()

	err = db.InitializePrisma()
	if err != nil {
		log.Fatalf("❌ Erreur d'initialisation DB: %v", err)
	}
	defer db.ClosePrisma()

	fmt.Println("✅ Connexion établie!")

	// Test 1: Créer des locations de test
	fmt.Println("\n🔧 Test 1: Création de locations")
	
	service := simpleCreationService.NewSimpleCreationService()

	// Location de ramassage
	lat1 := 5.37158  // Abidjan - Plateau
	lng1 := -4.02814
	pickupLocation, err := service.createLocation("Boulevard de la République, Plateau, Abidjan", &lat1, &lng1)
	if err != nil {
		log.Printf("❌ Erreur création pickup: %v", err)
	} else {
		fmt.Printf("✅ Location pickup créée: ID=%s, Address=%s\n", pickupLocation.ID, pickupLocation.Address)
	}

	// Location de livraison
	lat2 := 5.34489  // Abidjan - Cocody
	lng2 := -3.98589
	dropoffLocation, err := service.createLocation("Riviera 2, Cocody, Abidjan", &lat2, &lng2)
	if err != nil {
		log.Printf("❌ Erreur création dropoff: %v", err)
	} else {
		fmt.Printf("✅ Location dropoff créée: ID=%s, Address=%s\n", dropoffLocation.ID, dropoffLocation.Address)
	}

	// Test 2: Calculer distance entre les locations
	if pickupLocation != nil && dropoffLocation != nil {
		fmt.Println("\n📏 Test 2: Calcul de distance")
		distance, duration, err := service.calculateDistanceAndDuration(pickupLocation, dropoffLocation)
		if err != nil {
			log.Printf("❌ Erreur calcul distance: %v", err)
		} else {
			fmt.Printf("✅ Distance: %.2f km, Durée: %.2f min\n", distance, duration)
		}
	}

	// Test 3: Vérifier les statistiques de la base
	fmt.Println("\n📊 Test 3: Statistiques après création")
	stats, err := db.GetDatabaseStats()
	if err != nil {
		log.Printf("❌ Erreur statistiques: %v", err)
	} else {
		fmt.Printf("✅ Statistiques: %+v\n", stats)
	}

	// Test 4: Tester les prix de livraison
	fmt.Println("\n💰 Test 4: Calcul des prix")
	
	distance := 8.5 // Distance test
	
	// Prix livraison simple
	simplePrice := service.calculateSimpleDeliveryPrice(models.VehicleTypeMoto, distance, 0)
	fmt.Printf("✅ Prix livraison simple (moto): %.0f FCFA\n", simplePrice)

	voiturePrice := service.calculateSimpleDeliveryPrice(models.VehicleTypeVoiture, distance, 5) // 5 min d'attente
	fmt.Printf("✅ Prix livraison simple (voiture + 5min attente): %.0f FCFA\n", voiturePrice)

	fmt.Println("\n🎉 Tests du module Location terminés avec succès!")
}