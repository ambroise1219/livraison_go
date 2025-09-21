package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ambroise1219/livraison_go/database"
	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	"github.com/ambroise1219/livraison_go/services/delivery"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

func main() {
	fmt.Println("🚚 Test des types de livraisons avancées...")

	// Initialiser Prisma
	err := database.InitPrisma()
	if err != nil {
		log.Fatalf("❌ Erreur connexion Prisma: %v", err)
	}
	defer database.ClosePrisma()

	// Initialiser db.PrismaDB pour les services
	err = db.InitializePrisma()
	if err != nil {
		log.Fatalf("❌ Erreur initialisation db.PrismaDB: %v", err)
	}

	fmt.Println("✅ Connexion Prisma établie")

	// Créer un utilisateur test pour toutes les livraisons
	// Utiliser un timestamp pour éviter les doublons
	timestamp := time.Now().Unix()
	phoneNumber := fmt.Sprintf("+2250788%04d", timestamp%10000)
	user, err := createTestUser(phoneNumber, "Marie", "Express")
	if err != nil {
		log.Fatalf("❌ Erreur création utilisateur: %v", err)
	}
	fmt.Printf("✅ Utilisateur test créé: %s\n", user.Phone)

	// Test 1: Livraison Express
	fmt.Println("\n🏃 Test 1: Livraison Express")
	testExpressDelivery(user.Phone)

	// Test 2: Livraison Groupée
	fmt.Println("\n👥 Test 2: Livraison Groupée")
	testGroupDelivery(user.Phone)

	// Test 3: Déménagement
	fmt.Println("\n🏠 Test 3: Déménagement")
	testMovingDelivery(user.Phone)

	// Statistiques finales
	fmt.Println("\n📊 Statistiques finales")
	showFinalStats()
}

func createTestUser(phone, firstName, lastName string) (*models.User, error) {
	ctx := context.Background()
	user, err := database.PrismaClient.User.CreateOne(
		prismadb.User.Phone.Set(phone),
		prismadb.User.FirstName.Set(firstName),
		prismadb.User.LastName.Set(lastName),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return &models.User{
		ID:        user.ID,
		Phone:     user.Phone,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func testExpressDelivery(clientPhone string) {
	expressService := delivery.NewExpressCreationService()

	// Création d'une livraison express avec les vrais modèles
	request := &models.CreateDeliveryRequest{
		Type:         models.DeliveryTypeExpress,
		VehicleType:  models.VehicleTypeMoto,
		PickupAddress:  "Aéroport Félix Houphouët-Boigny",
		PickupLat:      floatPtr(5.2614),
		PickupLng:      floatPtr(-3.9263),
		DropoffAddress: "Hôtel Ivoire, Plateau",
		DropoffLat:     floatPtr(5.3200),
		DropoffLng:     floatPtr(-4.0200),
		PackageInfo: &models.PackageInfo{
			Description: stringPtr("Documents urgents"),
			WeightKg:    floatPtr(0.5),
			Fragile:     true,
		},
		PaymentMethod: "CASH",
	}

	response, err := expressService.CreateExpressDelivery(clientPhone, request)
	if err != nil {
		log.Printf("❌ Erreur Express: %v", err)
		return
	}

	fmt.Printf("✅ Livraison Express créée!\n")
	fmt.Printf("   📋 ID: %s\n", response.ID[:10]+"...")
	fmt.Printf("   ⚡ Type: %s\n", response.Type)
	fmt.Printf("   🏍️ Véhicule: %s\n", response.VehicleType)
	fmt.Printf("   📏 Distance: %.2f km\n", *response.DistanceKm)
	fmt.Printf("   💰 Prix Express: %.0f FCFA\n", response.FinalPrice)
	fmt.Printf("   🚀 Multiplier Express appliqué!\n")
}

func testGroupDelivery(clientPhone string) {
	groupService := delivery.NewGroupedCreationService()

	// Créer une livraison groupée avec les vrais modèles
	request := &models.CreateDeliveryRequest{
		Type:         models.DeliveryTypeGroupee,
		VehicleType:  models.VehicleTypeCamionnette,
		PickupAddress:  "Marché de Cocody",
		PickupLat:      floatPtr(5.3500),
		PickupLng:      floatPtr(-3.9800),
		DropoffAddress: "Riviera Golf, Cocody",
		DropoffLat:     floatPtr(5.3600),
		DropoffLng:     floatPtr(-3.9700),
		PackageInfo: &models.PackageInfo{
			Description: stringPtr("Courses alimentaires multiple"),
			WeightKg:    floatPtr(15.0), // Poids total groupé
			Fragile:     false,
		},
		GroupedInfo: &models.GroupedInfo{
			Zones: []models.GroupedZone{
				{
					ZoneNumber: 1,
					RecipientName: "Famille Kouassi",
					RecipientPhone: "+22507111111",
					PickupAddress: "Marché de Cocody",
					DeliveryAddress: "Riviera 1",
				},
				{
					ZoneNumber: 2,
					RecipientName: "Famille Traoré",
					RecipientPhone: "+22507222222",
					PickupAddress: "Marché de Cocody",
					DeliveryAddress: "Riviera 2",
				},
			},
		},
		PaymentMethod: "CASH",
	}

	response, err := groupService.CreateGroupedDelivery(clientPhone, request)
	if err != nil {
		log.Printf("❌ Erreur Groupée: %v", err)
		return
	}

	fmt.Printf("✅ Livraison Groupée créée!\n")
	fmt.Printf("   📋 ID: %s\n", response.ID[:10]+"...")
	fmt.Printf("   👥 Type: %s\n", response.Type)
	fmt.Printf("   🚚 Véhicule: %s\n", response.VehicleType)
	fmt.Printf("   📏 Distance: %.2f km\n", *response.DistanceKm)
	fmt.Printf("   💰 Prix Groupé: %.0f FCFA\n", response.FinalPrice)
	fmt.Printf("   💵 Économies sur livraisons groupées!\n")
}

func testMovingDelivery(clientPhone string) {
	movingService := delivery.NewMovingCreationService()

	request := &models.CreateDeliveryRequest{
		Type:         models.DeliveryTypeDemenagement,
		VehicleType:  models.VehicleTypeCamionnette,
		PickupAddress:  "Appartement Plateau, Rue des Jardins",
		PickupLat:      floatPtr(5.3200),
		PickupLng:      floatPtr(-4.0300),
		DropoffAddress: "Villa Cocody, Boulevard Lagunaire",
		DropoffLat:     floatPtr(5.3500),
		DropoffLng:     floatPtr(-3.9800),
		PackageInfo: &models.PackageInfo{
			Description: stringPtr("Mobilier complet 3 pièces"),
			WeightKg:    floatPtr(500.0), // 500kg de mobilier
			Fragile:     true,
		},
		MovingInfo: &models.MovingInfo{
			VehicleSize:    "LARGE",
			HelpersCount:   3,
			Floors:        2,
			HasElevator:   false,
			NeedsDisassembly: true,
			HasFragileItems: true,
			EstimatedVolume: floatPtr(25.0),
			AdditionalServices: []string{"packing", "assembly"},
			SpecialInstructions: stringPtr("Fragile: piano et œuvres d'art"),
		},
		PaymentMethod: "CASH",
	}

	response, err := movingService.CreateMovingDelivery(clientPhone, request)
	if err != nil {
		log.Printf("❌ Erreur Déménagement: %v", err)
		return
	}

	fmt.Printf("✅ Déménagement créé!\n")
	fmt.Printf("   📋 ID: %s\n", response.ID[:10]+"...")
	fmt.Printf("   🏠 Type: %s\n", response.Type)
	fmt.Printf("   📦 Volume: %.1f m³\n", *request.MovingInfo.EstimatedVolume)
	fmt.Printf("   👷 Aides: %d\n", request.MovingInfo.HelpersCount)
	fmt.Printf("   📏 Distance: %.2f km\n", *response.DistanceKm)
	fmt.Printf("   💰 Prix Déménagement: %.0f FCFA\n", response.FinalPrice)
	fmt.Printf("   ⬆️ Étages: %d (sans ascenseur)\n", request.MovingInfo.Floors)
	fmt.Printf("   📦 Services: %v\n", request.MovingInfo.AdditionalServices)
}

func showFinalStats() {
	stats, err := db.GetDatabaseStats()
	if err != nil {
		log.Printf("❌ Erreur statistiques: %v", err)
		return
	}

	fmt.Printf("✅ Tests des types de livraisons terminés!\n")
	fmt.Printf("   👥 Utilisateurs: %v\n", stats["users"])
	fmt.Printf("   🚚 Total livraisons: %v\n", stats["deliveries"])
	fmt.Printf("   📱 OTPs: %v\n", stats["otps"])
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}