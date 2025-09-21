package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ambroise1219/livraison_go/database"
	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
	"github.com/ambroise1219/livraison_go/services/delivery"
)

func main() {
	fmt.Println("👤 Test du parcours utilisateur complet...")

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

	// Étape 1: Créer un utilisateur
	fmt.Println("\n👤 Étape 1: Création d'un utilisateur")
	phone := "+22507123456"
	user, err := createUser(phone, "Jean", "Dupont")
	if err != nil {
		log.Printf("❌ Erreur création utilisateur: %v", err)
		return
	}
	fmt.Printf("✅ Utilisateur créé: ID=%s, Phone=%s, Name=%s %s\n", user.ID, user.Phone, user.FirstName, user.LastName)

	// Étape 2: Générer et vérifier un OTP  
	fmt.Println("\n📱 Étape 2: Génération et vérification OTP")
	otpCode, err := generateOTP(phone)
	if err != nil {
		log.Printf("❌ Erreur génération OTP: %v", err)
		return
	}
	fmt.Printf("✅ OTP généré: %s pour %s\n", otpCode, phone)

	// Vérifier l'OTP
	valid, err := verifyOTP(phone, otpCode)
	if err != nil {
		log.Printf("❌ Erreur vérification OTP: %v", err)
		return
	}
	if valid {
		fmt.Println("✅ OTP vérifié avec succès!")
	} else {
		fmt.Println("❌ OTP invalide!")
		return
	}

	// Étape 3: Créer une livraison avec l'utilisateur authentifié
	fmt.Println("\n🚚 Étape 3: Création d'une livraison authentifiée")
	
	simpleService := delivery.NewSimpleCreationService()

	// Demande de livraison
	deliveryRequest := &models.CreateDeliveryRequest{
		Type:           "STANDARD",
		PickupAddress:  "Boulevard de la République, Plateau, Abidjan",
		PickupLat:      func() *float64 { v := 5.37158; return &v }(),
		PickupLng:      func() *float64 { v := -4.02814; return &v }(),
		DropoffAddress: "Riviera 2, Cocody, Abidjan",
		DropoffLat:     func() *float64 { v := 5.34489; return &v }(),
		DropoffLng:     func() *float64 { v := -3.98589; return &v }(),
		VehicleType:    models.VehicleTypeMoto,
		PaymentMethod:  models.PaymentMethodCash,
		PackageInfo: &models.PackageInfo{
			Description: func() *string { v := "Documents importants"; return &v }(),
			WeightKg:    func() *float64 { v := 1.5; return &v }(),
			Fragile:     true,
			Size:        func() *string { v := "small"; return &v }(),
		},
	}

	// Créer la livraison avec l'ID utilisateur
	deliveryResponse, err := simpleService.CreateSimpleDelivery(user.Phone, deliveryRequest)
	if err != nil {
		log.Printf("❌ Erreur création livraison: %v", err)
		return
	}
	
	fmt.Printf("✅ Livraison créée pour l'utilisateur authentifié!\n")
	fmt.Printf("   📋 Livraison ID: %s\n", deliveryResponse.ID)
	fmt.Printf("   👤 Client: %s (%s)\n", user.Phone, user.FirstName+" "+user.LastName)
	fmt.Printf("   📍 De: %s\n", deliveryResponse.Pickup.Address)
	fmt.Printf("   📍 Vers: %s\n", deliveryResponse.Dropoff.Address)
	fmt.Printf("   📏 Distance: %.2f km\n", *deliveryResponse.DistanceKm)
	fmt.Printf("   💰 Prix: %.0f FCFA\n", deliveryResponse.FinalPrice)
	fmt.Printf("   📦 Colis: %s (%.1f kg) - Fragile: %t\n", 
		*deliveryRequest.PackageInfo.Description, 
		*deliveryRequest.PackageInfo.WeightKg,
		deliveryRequest.PackageInfo.Fragile)

	// Étape 4: Récupérer l'historique des livraisons de l'utilisateur
	fmt.Println("\n📋 Étape 4: Historique des livraisons")
	deliveryService := delivery.NewDeliveryService()
	userDeliveries, err := deliveryService.GetDeliveriesByClient(user.Phone)
	if err != nil {
		log.Printf("❌ Erreur récupération historique: %v", err)
	} else {
		fmt.Printf("✅ L'utilisateur a %d livraison(s) dans son historique\n", len(userDeliveries))
		for i, d := range userDeliveries {
			fmt.Printf("   %d. %s - %s - %.0f FCFA\n", i+1, d.ID[:8]+"...", d.Status, d.FinalPrice)
		}
	}

	// Étape 5: Statistiques finales
	fmt.Println("\n📊 Étape 5: Statistiques du système")
	stats, err := db.GetDatabaseStats()
	if err != nil {
		log.Printf("❌ Erreur statistiques: %v", err)
	} else {
		fmt.Printf("✅ Statistiques système:\n")
		fmt.Printf("   👥 Utilisateurs: %v\n", stats["users"])
		fmt.Printf("   🚚 Livraisons: %v\n", stats["deliveries"])
		fmt.Printf("   📱 OTPs: %v\n", stats["otps"])
		fmt.Printf("   🚗 Véhicules: %v\n", stats["vehicles"])
	}

	fmt.Println("\n🎉 Parcours utilisateur complet terminé avec succès!")
	fmt.Println("🏆 Authentification → OTP → Livraison → Historique ✅")
}

// createUser crée un nouvel utilisateur
func createUser(phone, firstName, lastName string) (*models.User, error) {
	// Ici on devrait utiliser un service utilisateur, 
	// pour l'instant on crée directement avec Prisma
	ctx := context.Background()
	user, err := db.PrismaDB.User.CreateOne(
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

// generateOTP génère un OTP pour un numéro de téléphone
func generateOTP(phone string) (string, error) {
	// Code OTP simple pour les tests
	otpCode := "1234"
	ctx := context.Background()
	
	// Créer l'enregistrement OTP en base avec une expiration dans 5 minutes
	expirationTime := time.Now().Add(5 * time.Minute)
	_, err := db.PrismaDB.Otp.CreateOne(
		prismadb.Otp.Phone.Set(phone),
		prismadb.Otp.Code.Set(otpCode),
		prismadb.Otp.ExpiresAt.Set(expirationTime),
	).Exec(ctx)

	if err != nil {
		return "", fmt.Errorf("failed to create OTP: %v", err)
	}

	return otpCode, nil
}

// verifyOTP vérifie un code OTP
func verifyOTP(phone, code string) (bool, error) {
	ctx := context.Background()
	
	// Chercher l'OTP en base
	otp, err := db.PrismaDB.Otp.FindFirst(
		prismadb.Otp.Phone.Equals(phone),
		prismadb.Otp.Code.Equals(code),
		prismadb.Otp.ExpiresAt.Gt(time.Now()),
	).Exec(ctx)

	if err != nil {
		if err == prismadb.ErrNotFound {
			return false, nil // OTP non trouvé ou expiré
		}
		return false, fmt.Errorf("failed to verify OTP: %v", err)
	}

	// OTP trouvé et valide
	return otp != nil, nil
}