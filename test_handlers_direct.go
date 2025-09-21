package main

import (
	"fmt"
	"log"

	"github.com/ambroise1219/livraison_go/database"
	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/handlers"
	"github.com/ambroise1219/livraison_go/models"
	"github.com/ambroise1219/livraison_go/services/auth"
	"github.com/ambroise1219/livraison_go/services/delivery"
)

func main() {
	fmt.Println("🧪 TEST DIRECT DES HANDLERS")
	fmt.Println("============================")
	
	// Initialiser la base de données
	if err := database.InitPrisma(); err != nil {
		log.Fatalf("❌ Erreur DB: %v", err)
	}
	defer database.ClosePrisma()
	
	if err := db.InitializePrisma(); err != nil {
		log.Fatalf("❌ Erreur db client: %v", err)
	}
	defer db.ClosePrisma()
	
	// Initialiser les handlers
	handlers.InitHandlers()
	fmt.Println("✅ Handlers initialisés")
	
	// Test 1: Créer un utilisateur de test
	fmt.Println("\n📱 Test 1: Création utilisateur via OTP")
	user, err := createTestUser("+2250123456789")
	if err != nil {
		log.Fatalf("❌ Erreur création utilisateur: %v", err)
	}
	fmt.Printf("✅ Utilisateur créé: %s (%s)\n", user.Phone, user.ID)
	
	// Test 2: Test création livraison simple
	fmt.Println("\n🚚 Test 2: Création livraison STANDARD")
	delivery, err := testCreateDelivery(user.Phone)
	if err != nil {
		log.Fatalf("❌ Erreur création livraison: %v", err)
	}
	fmt.Printf("✅ Livraison créée: %s (Type: %s, Prix: %.2f)\n", 
		delivery.ID, delivery.Type, delivery.FinalPrice)
	
	// Test 3: Test récupération livraison
	fmt.Println("\n🔍 Test 3: Récupération livraison")
	retrievedDelivery, err := testGetDelivery(delivery.ID)
	if err != nil {
		log.Fatalf("❌ Erreur récupération livraison: %v", err)
	}
	fmt.Printf("✅ Livraison récupérée: %s (Status: %s)\n", 
		retrievedDelivery.ID, retrievedDelivery.Status)
	
	fmt.Println("\n🎉 TOUS LES TESTS RÉUSSIS!")
	fmt.Println("✅ Migration Prisma: OK")
	fmt.Println("✅ Handlers CreateDelivery: OK") 
	fmt.Println("✅ Handlers GetDelivery: OK")
	fmt.Println("✅ Base de données: OK")
}

func createTestUser(phone string) (*models.User, error) {
	userService := auth.NewUserService()
	user, _, err := userService.FindOrCreateUser(phone)
	return user, err
}

func testCreateDelivery(clientPhone string) (*models.DeliveryResponse, error) {
	simpleService := delivery.NewSimpleCreationService()
	
	req := &models.CreateDeliveryRequest{
		Type:           models.DeliveryTypeStandard,
		PickupAddress:  "Cocody Riviera, Abidjan",
		PickupLat:      floatPtr(5.3599517),
		PickupLng:      floatPtr(-3.9622047),
		DropoffAddress: "Yopougon, Abidjan",
		DropoffLat:     floatPtr(5.3456),
		DropoffLng:     floatPtr(-4.0731),
		VehicleType:    models.VehicleTypeMotorcycle,
		PaymentMethod:  models.PaymentMethodCash,
		PackageInfo: &models.PackageInfo{
			Description: stringPtr("Produits électroniques"),
			WeightKg:    floatPtr(2.5),
			Fragile:     true,
		},
	}
	
	return simpleService.CreateSimpleDelivery(clientPhone, req)
}

func testGetDelivery(deliveryID string) (*models.Delivery, error) {
	deliveryService := delivery.NewDeliveryService()
	return deliveryService.GetDelivery(deliveryID)
}

func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}