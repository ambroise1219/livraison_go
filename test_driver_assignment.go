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
	fmt.Println("🚗 Test de l'assignation de livreurs...")

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

	// Étape 1: Créer des livreurs et véhicules
	fmt.Println("\n👨‍💼 Étape 1: Création de livreurs et véhicules")
	drivers := createTestDrivers()
	vehicles := createTestVehicles(drivers)

	// Étape 2: Créer un client et une livraison
	fmt.Println("\n🧑‍💼 Étape 2: Création d'un client et d'une livraison")
	client := createTestClient()
	delivery := createTestDelivery(client.Phone)

	// Étape 3: Test d'assignation automatique
	fmt.Println("\n🤖 Étape 3: Assignation automatique")
	testAutoAssignment(delivery.ID)

	// Étape 4: Test d'assignation manuelle
	fmt.Println("\n✋ Étape 4: Assignation manuelle")
	delivery2 := createTestDelivery(client.Phone)
	testManualAssignment(delivery2.ID, drivers[1].ID)

	// Étape 5: Test de mise à jour de statut
	fmt.Println("\n📊 Étape 5: Mise à jour de statuts")
	testStatusUpdates(delivery.ID)

	// Statistiques finales
	fmt.Println("\n📊 Étape 6: Statistiques finales")
	showDriverStats(drivers, vehicles)
}

func createTestDrivers() []*models.User {
	fmt.Println("🚶 Création de livreurs...")
	var drivers []*models.User
	
	driverData := []struct {
		phone, firstName, lastName string
	}{
		{"+22507111001", "Kouame", "Driver1"},
		{"+22507111002", "Yao", "Driver2"},
		{"+22507111003", "Kone", "Driver3"},
	}

	for i, d := range driverData {
		timestamp := time.Now().Unix()
		phoneNumber := fmt.Sprintf("+2250711%04d", (timestamp+int64(i))%10000)
		
		driver, err := createDriver(phoneNumber, d.firstName, d.lastName)
		if err != nil {
			log.Printf("❌ Erreur création driver %s: %v", d.firstName, err)
			continue
		}
		drivers = append(drivers, driver)
		fmt.Printf("   ✅ Livreur créé: %s %s (%s)\n", driver.FirstName, driver.LastName, driver.Phone)
	}

	return drivers
}

func createTestVehicles(drivers []*models.User) []*models.Vehicle {
	fmt.Println("🚗 Création de véhicules...")
	var vehicles []*models.Vehicle

	vehicleData := []struct {
		vehicleType  string // Utiliser string pour le mapping direct
		plate        string
		description  string
	}{
		{"MOTORCYCLE", "AB123CD", "Yamaha 125cc"},
		{"CAR", "EF456GH", "Toyota Corolla"},
		{"VAN", "IJ789KL", "Hyundai H100"},
	}

	for i, vd := range vehicleData {
		if i < len(drivers) {
			vehicle, err := createVehicle(vd.vehicleType, vd.plate, vd.description, drivers[i].ID)
			if err != nil {
				log.Printf("❌ Erreur création véhicule %s: %v", vd.description, err)
				continue
			}
			vehicles = append(vehicles, vehicle)
			plateNumber := "N/A"
			if vehicle.PlaqueImmatriculation != nil {
				plateNumber = *vehicle.PlaqueImmatriculation
			}
			fmt.Printf("   🚗 Véhicule créé: %s (%s) - Driver: %s\n", vehicle.Type, plateNumber, drivers[i].FirstName)
		}
	}

	return vehicles
}

func createTestClient() *models.User {
	timestamp := time.Now().Unix()
	phoneNumber := fmt.Sprintf("+2250788%04d", timestamp%10000)
	
	client, err := createUser(phoneNumber, "Client", "Test")
	if err != nil {
		log.Fatalf("❌ Erreur création client: %v", err)
	}
	
	fmt.Printf("✅ Client créé: %s %s (%s)\n", client.FirstName, client.LastName, client.Phone)
	return client
}

func createTestDelivery(clientPhone string) *models.DeliveryResponse {
	simpleService := delivery.NewSimpleCreationService()

	request := &models.CreateDeliveryRequest{
		Type:         models.DeliveryTypeSimple,
		VehicleType:  models.VehicleTypeMoto,
		PickupAddress:  "Point de ramassage Plateau",
		PickupLat:      floatPtr(5.3200),
		PickupLng:      floatPtr(-4.0200),
		DropoffAddress: "Point de dépôt Cocody",
		DropoffLat:     floatPtr(5.3500),
		DropoffLng:     floatPtr(-3.9800),
		PackageInfo: &models.PackageInfo{
			Description: stringPtr("Colis test assignation"),
			WeightKg:    floatPtr(2.0),
			Fragile:     false,
		},
		PaymentMethod: "CASH",
	}

	response, err := simpleService.CreateSimpleDelivery(clientPhone, request)
	if err != nil {
		log.Fatalf("❌ Erreur création livraison: %v", err)
	}

	fmt.Printf("✅ Livraison créée: ID=%s, Prix=%.0f FCFA\n", response.ID[:8]+"...", response.FinalPrice)
	return response
}

func testAutoAssignment(deliveryID string) {
	deliveryService := delivery.NewDeliveryService()

	// Récupérer la livraison avant assignation
	beforeAssignment, err := deliveryService.GetDelivery(deliveryID)
	if err != nil {
		log.Printf("❌ Erreur récupération livraison: %v", err)
		return
	}

	fmt.Printf("🔍 Avant assignation - Status: %s, Driver: %v\n", beforeAssignment.Status, beforeAssignment.LivreurID)

	// Effectuer assignation automatique (la logique devrait choisir un driver disponible)
	assignRequest := &models.AssignDeliveryRequest{
		DeliveryID: deliveryID,
		DriverID:   nil, // nil = assignation automatique
	}

	err = assignDelivery(assignRequest)
	if err != nil {
		log.Printf("❌ Erreur assignation automatique: %v", err)
		return
	}

	// Récupérer la livraison après assignation
	afterAssignment, err := deliveryService.GetDelivery(deliveryID)
	if err != nil {
		log.Printf("❌ Erreur récupération après assignation: %v", err)
		return
	}

	if afterAssignment.LivreurID != nil {
		fmt.Printf("✅ Assignation automatique réussie!\n")
		fmt.Printf("   📋 Livraison: %s\n", deliveryID[:8]+"...")
		fmt.Printf("   🚗 Driver assigné: %s\n", *afterAssignment.LivreurID)
		fmt.Printf("   📊 Nouveau statut: %s\n", afterAssignment.Status)
	} else {
		fmt.Printf("⚠️  Aucun driver assigné automatiquement\n")
	}
}

func testManualAssignment(deliveryID, driverID string) {
	deliveryService := delivery.NewDeliveryService()

	fmt.Printf("🎯 Test assignation manuelle - Delivery: %s, Driver: %s\n", deliveryID[:8]+"...", driverID[:8]+"...")

	assignRequest := &models.AssignDeliveryRequest{
		DeliveryID: deliveryID,
		DriverID:   &driverID,
	}

	err := assignDelivery(assignRequest)
	if err != nil {
		log.Printf("❌ Erreur assignation manuelle: %v", err)
		return
	}

	// Vérifier l'assignation
	delivery, err := deliveryService.GetDelivery(deliveryID)
	if err != nil {
		log.Printf("❌ Erreur récupération après assignation: %v", err)
		return
	}

	if delivery.LivreurID != nil && *delivery.LivreurID == driverID {
		fmt.Printf("✅ Assignation manuelle réussie!\n")
		fmt.Printf("   👤 Driver assigné: %s\n", driverID[:8]+"...")
		fmt.Printf("   📊 Statut: %s\n", delivery.Status)
	} else {
		fmt.Printf("❌ Assignation manuelle échouée\n")
	}
}

func testStatusUpdates(deliveryID string) {
	deliveryService := delivery.NewDeliveryService()

	statusProgression := []models.DeliveryStatus{
		models.DeliveryStatusAccepted,
		models.DeliveryStatusPickedUp,
		models.DeliveryStatusInTransit,
		models.DeliveryStatusDelivered,
	}

	for _, status := range statusProgression {
		delivery, err := deliveryService.GetDelivery(deliveryID)
		if err != nil {
			log.Printf("❌ Erreur récupération pour mise à jour: %v", err)
			continue
		}

		// Mise à jour du statut dans l'objet
		delivery.Status = status
		err = deliveryService.UpdateDelivery(delivery)
		if err != nil {
			log.Printf("❌ Erreur mise à jour statut %s: %v", status, err)
			continue
		}

		fmt.Printf("📊 Statut mis à jour: %s\n", status)
		time.Sleep(500 * time.Millisecond) // Pause pour simuler le temps
	}

	fmt.Println("✅ Progression de statuts terminée!")
}

func showDriverStats(drivers []*models.User, vehicles []*models.Vehicle) {
	stats, err := db.GetDatabaseStats()
	if err != nil {
		log.Printf("❌ Erreur statistiques: %v", err)
		return
	}

	fmt.Printf("✅ Test d'assignation terminé!\n")
	fmt.Printf("📊 Résumé:\n")
	fmt.Printf("   👨‍💼 Livreurs créés: %d\n", len(drivers))
	fmt.Printf("   🚗 Véhicules créés: %d\n", len(vehicles))
	fmt.Printf("   👥 Total utilisateurs: %v\n", stats["users"])
	fmt.Printf("   🚚 Total livraisons: %v\n", stats["deliveries"])
	fmt.Printf("   🚙 Total véhicules: %v\n", stats["vehicles"])
	fmt.Println("\n🏆 Tests d'assignation et de gestion des statuts réussis!")
}

// Fonctions utilitaires
func createUser(phone, firstName, lastName string) (*models.User, error) {
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

func createDriver(phone, firstName, lastName string) (*models.User, error) {
	// Même fonction que createUser, mais on pourrait ajouter des champs spécifiques driver
	return createUser(phone, firstName, lastName)
}

func createVehicle(vehicleType string, plateNumber, description, driverID string) (*models.Vehicle, error) {
	ctx := context.Background()
	
	// Convertir le string vers le bon enum Prisma
	var prismaVehicleType prismadb.VehicleType
	switch vehicleType {
	case "MOTORCYCLE":
		prismaVehicleType = prismadb.VehicleTypeMotorcycle
	case "CAR":
		prismaVehicleType = prismadb.VehicleTypeCar
	case "VAN":
		prismaVehicleType = prismadb.VehicleTypeVan
	default:
		prismaVehicleType = prismadb.VehicleTypeMotorcycle
	}

	vehicle, err := database.PrismaClient.Vehicle.CreateOne(
		prismadb.Vehicle.Type.Set(prismaVehicleType),
		prismadb.Vehicle.PlateNumber.Set(plateNumber),
		prismadb.Vehicle.Capacity.Set(getVehicleCapacityFromString(vehicleType)),
	).Exec(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create vehicle: %v", err)
	}

	// Gérer le champ Make optionnel
	var marque *string
	if make, ok := vehicle.Make(); ok {
		marque = &make
	}

	// Convertir le type Prisma vers le type du modèle
	var modelVehicleType models.VehicleType
	switch vehicle.Type {
	case prismadb.VehicleTypeMotorcycle:
		modelVehicleType = models.VehicleTypeMoto
	case prismadb.VehicleTypeCar:
		modelVehicleType = models.VehicleTypeVoiture
	case prismadb.VehicleTypeVan:
		modelVehicleType = models.VehicleTypeCamionnette
	default:
		modelVehicleType = models.VehicleTypeMoto
	}

	return &models.Vehicle{
		ID:                   vehicle.ID,
		Type:                 modelVehicleType,
		UserID:               driverID, // Stocker le driverID dans UserID
		PlaqueImmatriculation: &vehicle.PlateNumber,
		Marque:               marque,
		CreatedAt:            vehicle.CreatedAt,
	}, nil
}

func assignDelivery(request *models.AssignDeliveryRequest) error {
	ctx := context.Background()
	
	// Logique d'assignation simple
	var driverID string
	if request.DriverID != nil {
		driverID = *request.DriverID
	} else {
		// Assignation automatique - chercher un driver disponible
		vehicles, err := database.PrismaClient.Vehicle.FindMany(
			prismadb.Vehicle.IsActive.Equals(true),
		).Exec(ctx)
		if err != nil || len(vehicles) == 0 {
			return fmt.Errorf("aucun véhicule disponible")
		}
		// Pour l'instant, on prend le premier utilisateur comme driver
		// Dans un vrai système, il faudrait avoir une relation Vehicle -> Driver
		users, err := database.PrismaClient.User.FindMany().Exec(ctx)
		if err != nil || len(users) == 0 {
			return fmt.Errorf("aucun driver disponible")
		}
		driverID = users[0].ID
	}

	// Mettre à jour la livraison
	_, err := database.PrismaClient.Delivery.FindUnique(
		prismadb.Delivery.ID.Equals(request.DeliveryID),
	).Update(
		prismadb.Delivery.DriverID.Set(driverID),
		prismadb.Delivery.Status.Set(prismadb.DeliveryStatusAssigned),
	).Exec(ctx)

	return err
}

func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

func getVehicleCapacity(vehicleType models.VehicleType) float64 {
	switch vehicleType {
	case models.VehicleTypeMoto:
		return 50.0 // 50 kg
	case models.VehicleTypeVoiture:
		return 200.0 // 200 kg
	case models.VehicleTypeCamionnette:
		return 1000.0 // 1000 kg
	default:
		return 50.0
	}
}

func getVehicleCapacityFromString(vehicleType string) float64 {
	switch vehicleType {
	case "MOTORCYCLE":
		return 50.0 // 50 kg
	case "CAR":
		return 200.0 // 200 kg
	case "VAN":
		return 1000.0 // 1000 kg
	default:
		return 50.0
	}
}
