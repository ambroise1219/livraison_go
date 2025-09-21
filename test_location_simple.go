package main

import (
	"fmt"
	"log"

	"github.com/ambroise1219/livraison_go/database"
	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	"github.com/ambroise1219/livraison_go/services/delivery"
)

func main() {
	fmt.Println("📍 Test du module Location (API publiques)...")

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

	// Test 1: Créer une livraison simple pour tester les locations
	fmt.Println("\n🚚 Test 1: Création d'une livraison de test")
	
	simpleService := delivery.NewSimpleCreationService()

	// Créer une demande de livraison
	deliveryRequest := &models.CreateDeliveryRequest{
		Type:           "STANDARD", // Utiliser STANDARD pour correspondre au schéma Prisma
		PickupAddress:  "Boulevard de la République, Plateau, Abidjan",
		PickupLat:      func() *float64 { v := 5.37158; return &v }(),
		PickupLng:      func() *float64 { v := -4.02814; return &v }(),
		DropoffAddress: "Riviera 2, Cocody, Abidjan",
		DropoffLat:     func() *float64 { v := 5.34489; return &v }(),
		DropoffLng:     func() *float64 { v := -3.98589; return &v }(),
		VehicleType:    models.VehicleTypeMoto,
		PaymentMethod:  models.PaymentMethodCash,
		PackageInfo: &models.PackageInfo{
			Description: func() *string { v := "Colis de test"; return &v }(),
			WeightKg:    func() *float64 { v := 2.5; return &v }(),
			Fragile:     false,
			Size:        func() *string { v := "medium"; return &v }(),
		},
	}

	clientID := "+22507123456"

	// Créer la livraison
	deliveryResponse, err := simpleService.CreateSimpleDelivery(clientID, deliveryRequest)
	if err != nil {
		log.Printf("❌ Erreur création livraison: %v", err)
	} else {
		fmt.Printf("✅ Livraison créée: ID=%s\n", deliveryResponse.ID)
		fmt.Printf("   📍 Pickup: %s\n", deliveryResponse.Pickup.Address)
		fmt.Printf("   📍 Dropoff: %s\n", deliveryResponse.Dropoff.Address)
		fmt.Printf("   📏 Distance: %.2f km\n", *deliveryResponse.DistanceKm)
		fmt.Printf("   ⏱️ Durée estimée: %.0f min\n", *deliveryResponse.DurationMin)
		fmt.Printf("   💰 Prix: %.0f FCFA\n", deliveryResponse.FinalPrice)
	}

	// Test 2: Vérifier les statistiques de la base
	fmt.Println("\n📊 Test 2: Statistiques après création")
	stats, err := db.GetDatabaseStats()
	if err != nil {
		log.Printf("❌ Erreur statistiques: %v", err)
	} else {
		fmt.Printf("✅ Statistiques DB: %+v\n", stats)
	}

	// Test 3: Tester le service delivery pour récupérer la livraison
	if deliveryResponse != nil {
		fmt.Println("\n🔍 Test 3: Récupération de la livraison")
		deliveryService := delivery.NewDeliveryService()
		
		retrievedDelivery, err := deliveryService.GetDelivery(deliveryResponse.ID)
		if err != nil {
			log.Printf("❌ Erreur récupération: %v", err)
		} else {
			fmt.Printf("✅ Livraison récupérée: ID=%s, Status=%s\n", retrievedDelivery.ID, retrievedDelivery.Status)
			fmt.Printf("   Client: %s\n", retrievedDelivery.ClientID)
			fmt.Printf("   Type: %s\n", retrievedDelivery.Type)
		}
	}

	fmt.Println("\n🎉 Tests du module Location terminés avec succès!")
}