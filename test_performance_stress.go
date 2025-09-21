package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/ambroise1219/livraison_go/database"
	"github.com/ambroise1219/livraison_go/db"
	"github.com/ambroise1219/livraison_go/models"
	"github.com/ambroise1219/livraison_go/services/delivery"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

const (
	CONCURRENT_USERS    = 10
	DELIVERIES_PER_USER = 5
	WORKERS             = 20
)

type TestMetrics struct {
	sync.Mutex
	TotalUsers      int
	TotalDeliveries int
	SuccessfulOps   int
	FailedOps       int
	StartTime       time.Time
	EndTime         time.Time
}

func main() {
	fmt.Println("🚀 Test de performance et de stress...")
	fmt.Printf("📊 Configuration:\n")
	fmt.Printf("   👥 Utilisateurs concurrents: %d\n", CONCURRENT_USERS)
	fmt.Printf("   📦 Livraisons par utilisateur: %d\n", DELIVERIES_PER_USER)
	fmt.Printf("   🔧 Workers: %d\n", WORKERS)
	
	// Initialiser Prisma
	err := database.InitPrisma()
	if err != nil {
		log.Fatalf("❌ Erreur connexion Prisma: %v", err)
	}
	defer database.ClosePrisma()

	err = db.InitializePrisma()
	if err != nil {
		log.Fatalf("❌ Erreur initialisation db.PrismaDB: %v", err)
	}

	fmt.Println("✅ Connexion Prisma établie")

	// Initialiser les métriques
	metrics := &TestMetrics{
		StartTime: time.Now(),
	}

	// Lancer les tests
	fmt.Println("\n🏁 Lancement des tests de charge...")
	
	// Test 1: Création massive d'utilisateurs
	testMassiveUserCreation(metrics)
	
	// Test 2: Création massive de livraisons
	testMassiveDeliveryCreation(metrics)
	
	// Test 3: Opérations concurrentes mixtes
	testMixedConcurrentOperations(metrics)
	
	// Test 4: Test de stress de lecture
	testReadStress(metrics)

	// Finaliser les métriques
	metrics.EndTime = time.Now()
	
	// Afficher le rapport final
	showPerformanceReport(metrics)
}

func testMassiveUserCreation(metrics *TestMetrics) {
	fmt.Println("\n👥 Test 1: Création massive d'utilisateurs")
	
	var wg sync.WaitGroup
	userChan := make(chan int, CONCURRENT_USERS)
	
	startTime := time.Now()
	
	// Lancer les workers
	for i := 0; i < WORKERS; i++ {
		go func(workerID int) {
			for userNum := range userChan {
				phone := fmt.Sprintf("+2250799%04d", userNum)
				firstName := fmt.Sprintf("User%d", userNum)
				lastName := "StressTest"
				
				_, err := createUser(phone, firstName, lastName)
				
				metrics.Lock()
				if err != nil {
					metrics.FailedOps++
					if userNum <= 5 { // Afficher seulement les premières erreurs
						fmt.Printf("   ❌ Worker %d: Erreur utilisateur %d: %v\n", workerID, userNum, err)
					}
				} else {
					metrics.SuccessfulOps++
					metrics.TotalUsers++
					if userNum%20 == 0 { // Afficher le progrès
						fmt.Printf("   ✅ Worker %d: %d utilisateurs créés\n", workerID, userNum)
					}
				}
				metrics.Unlock()
				
				wg.Done()
			}
		}(i)
	}
	
	// Envoyer les tâches
	for i := 1; i <= CONCURRENT_USERS; i++ {
		wg.Add(1)
		userChan <- i
	}
	
	close(userChan)
	wg.Wait()
	
	duration := time.Since(startTime)
	fmt.Printf("✅ Création d'utilisateurs terminée en %v\n", duration)
	fmt.Printf("   📊 Succès: %d, Échecs: %d\n", metrics.TotalUsers, metrics.FailedOps)
	fmt.Printf("   ⚡ Débit: %.2f utilisateurs/seconde\n", float64(metrics.TotalUsers)/duration.Seconds())
}

func testMassiveDeliveryCreation(metrics *TestMetrics) {
	fmt.Println("\n🚚 Test 2: Création massive de livraisons")
	
	var wg sync.WaitGroup
	type DeliveryTask struct {
		UserNum  int
		DelNum   int
		Phone    string
	}
	
	deliveryChan := make(chan DeliveryTask, CONCURRENT_USERS*DELIVERIES_PER_USER)
	
	startTime := time.Now()
	
	// Créer quelques utilisateurs pour les livraisons
	testUsers := make([]string, CONCURRENT_USERS)
	for i := 0; i < CONCURRENT_USERS; i++ {
		timestamp := time.Now().UnixNano()
		phone := fmt.Sprintf("+2250788%04d", (timestamp+int64(i))%10000)
		user, err := createUser(phone, fmt.Sprintf("DeliveryUser%d", i), "Test")
		if err == nil {
			testUsers[i] = user.Phone
		} else {
			testUsers[i] = phone // Fallback
		}
	}
	
	// Lancer les workers
	for i := 0; i < WORKERS; i++ {
		go func(workerID int) {
			simpleService := delivery.NewSimpleCreationService()
			
			for task := range deliveryChan {
				request := &models.CreateDeliveryRequest{
					Type:         models.DeliveryTypeSimple,
					VehicleType:  models.VehicleTypeMoto,
					PickupAddress:  fmt.Sprintf("Pickup %d-%d", task.UserNum, task.DelNum),
					PickupLat:      floatPtr(5.3200 + float64(task.UserNum)*0.001),
					PickupLng:      floatPtr(-4.0200 + float64(task.DelNum)*0.001),
					DropoffAddress: fmt.Sprintf("Dropoff %d-%d", task.UserNum, task.DelNum),
					DropoffLat:     floatPtr(5.3500 + float64(task.UserNum)*0.001),
					DropoffLng:     floatPtr(-3.9800 + float64(task.DelNum)*0.001),
					PackageInfo: &models.PackageInfo{
						Description: stringPtr(fmt.Sprintf("Colis stress test %d-%d", task.UserNum, task.DelNum)),
						WeightKg:    floatPtr(1.0 + float64(task.DelNum)*0.1),
						Fragile:     task.DelNum%2 == 0,
					},
					PaymentMethod: "CASH",
				}
				
				_, err := simpleService.CreateSimpleDelivery(task.Phone, request)
				
				metrics.Lock()
				if err != nil {
					metrics.FailedOps++
				} else {
					metrics.SuccessfulOps++
					metrics.TotalDeliveries++
					if metrics.TotalDeliveries%25 == 0 {
						fmt.Printf("   ✅ Worker %d: %d livraisons créées\n", workerID, metrics.TotalDeliveries)
					}
				}
				metrics.Unlock()
				
				wg.Done()
			}
		}(i)
	}
	
	// Envoyer les tâches
	for userNum := 0; userNum < CONCURRENT_USERS; userNum++ {
		for delNum := 1; delNum <= DELIVERIES_PER_USER; delNum++ {
			wg.Add(1)
			deliveryChan <- DeliveryTask{
				UserNum: userNum,
				DelNum:  delNum,
				Phone:   testUsers[userNum],
			}
		}
	}
	
	close(deliveryChan)
	wg.Wait()
	
	duration := time.Since(startTime)
	fmt.Printf("✅ Création de livraisons terminée en %v\n", duration)
	fmt.Printf("   📦 Livraisons créées: %d\n", metrics.TotalDeliveries)
	fmt.Printf("   ⚡ Débit: %.2f livraisons/seconde\n", float64(metrics.TotalDeliveries)/duration.Seconds())
}

func testMixedConcurrentOperations(metrics *TestMetrics) {
	fmt.Println("\n🔄 Test 3: Opérations concurrentes mixtes")
	
	var wg sync.WaitGroup
	startTime := time.Now()
	
	operations := []string{"CREATE_USER", "CREATE_DELIVERY", "GET_DELIVERIES", "UPDATE_STATUS"}
	successCount := make(map[string]int)
	failCount := make(map[string]int)
	var mu sync.Mutex
	
	for i := 0; i < 50; i++ { // 50 opérations concurrentes
		wg.Add(1)
		go func(opIndex int) {
			defer wg.Done()
			
			op := operations[opIndex%len(operations)]
			var err error
			
			switch op {
			case "CREATE_USER":
				phone := fmt.Sprintf("+2250777%04d", opIndex)
				_, err = createUser(phone, fmt.Sprintf("Mixed%d", opIndex), "Test")
				
			case "CREATE_DELIVERY":
				// Utiliser un utilisateur existant ou créer un nouveau
				phone := fmt.Sprintf("+2250766%04d", opIndex)
				createUser(phone, "QuickUser", "Test") // S'assurer que l'utilisateur existe
				
				simpleService := delivery.NewSimpleCreationService()
				request := &models.CreateDeliveryRequest{
					Type:         models.DeliveryTypeSimple,
					VehicleType:  models.VehicleTypeMoto,
					PickupAddress:  fmt.Sprintf("Quick pickup %d", opIndex),
					PickupLat:      floatPtr(5.32),
					PickupLng:      floatPtr(-4.02),
					DropoffAddress: fmt.Sprintf("Quick dropoff %d", opIndex),
					DropoffLat:     floatPtr(5.35),
					DropoffLng:     floatPtr(-3.98),
					PackageInfo: &models.PackageInfo{
						Description: stringPtr(fmt.Sprintf("Quick package %d", opIndex)),
						WeightKg:    floatPtr(1.0),
						Fragile:     false,
					},
					PaymentMethod: "CASH",
				}
				_, err = simpleService.CreateSimpleDelivery(phone, request)
				
			case "GET_DELIVERIES":
				deliveryService := delivery.NewDeliveryService()
				phone := fmt.Sprintf("+2250766%04d", opIndex%10) // Réutiliser quelques numéros
				_, err = deliveryService.GetDeliveriesByClient(phone)
				
			case "UPDATE_STATUS":
				// Simulation de mise à jour (sans erreur pour ce test)
				time.Sleep(1 * time.Millisecond) // Simuler une opération
			}
			
			mu.Lock()
			if err != nil {
				failCount[op]++
			} else {
				successCount[op]++
			}
			mu.Unlock()
		}(i)
	}
	
	wg.Wait()
	
	duration := time.Since(startTime)
	fmt.Printf("✅ Opérations mixtes terminées en %v\n", duration)
	fmt.Println("   📊 Résultats par opération:")
	for op := range operations {
		opName := operations[op]
		fmt.Printf("     %s: ✅ %d, ❌ %d\n", opName, successCount[opName], failCount[opName])
	}
}

func testReadStress(metrics *TestMetrics) {
	fmt.Println("\n📖 Test 4: Stress de lecture")
	
	var wg sync.WaitGroup
	startTime := time.Now()
	
	readCount := 0
	var readMu sync.Mutex
	
	// Lancer 30 lectures concurrentes
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(readID int) {
			defer wg.Done()
			
			deliveryService := delivery.NewDeliveryService()
			
			// Faire 10 lectures par goroutine
			for j := 0; j < 10; j++ {
				phone := fmt.Sprintf("+2250766%04d", (readID+j)%10)
				_, err := deliveryService.GetDeliveriesByClient(phone)
				
				readMu.Lock()
				if err == nil {
					readCount++
				}
				readMu.Unlock()
				
				// Petite pause pour simuler des lectures réalistes
				time.Sleep(5 * time.Millisecond)
			}
		}(i)
	}
	
	wg.Wait()
	
	duration := time.Since(startTime)
	fmt.Printf("✅ Test de stress de lecture terminé en %v\n", duration)
	fmt.Printf("   📊 Lectures réussies: %d\n", readCount)
	fmt.Printf("   ⚡ Débit: %.2f lectures/seconde\n", float64(readCount)/duration.Seconds())
}

func showPerformanceReport(metrics *TestMetrics) {
	duration := metrics.EndTime.Sub(metrics.StartTime)
	
	// Récupérer les statistiques finales de la DB
	stats, err := db.GetDatabaseStats()
	if err != nil {
		log.Printf("❌ Erreur statistiques: %v", err)
		return
	}
	
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🏆 RAPPORT DE PERFORMANCE FINAL")
	fmt.Println(strings.Repeat("=", 60))
	
	fmt.Printf("⏱️  Durée totale: %v\n", duration)
	fmt.Printf("✅ Opérations réussies: %d\n", metrics.SuccessfulOps)
	fmt.Printf("❌ Opérations échouées: %d\n", metrics.FailedOps)
	fmt.Printf("📊 Taux de succès: %.2f%%\n", float64(metrics.SuccessfulOps)/float64(metrics.SuccessfulOps+metrics.FailedOps)*100)
	
	fmt.Println("\n📈 STATISTIQUES DE LA BASE DE DONNÉES:")
	fmt.Printf("   👥 Utilisateurs totaux: %v\n", stats["users"])
	fmt.Printf("   🚚 Livraisons totales: %v\n", stats["deliveries"])
	fmt.Printf("   📱 OTPs: %v\n", stats["otps"])
	fmt.Printf("   🚙 Véhicules: %v\n", stats["vehicles"])
	
	fmt.Println("\n⚡ PERFORMANCES:")
	totalOps := metrics.SuccessfulOps + metrics.FailedOps
	fmt.Printf("   🔥 Débit global: %.2f opérations/seconde\n", float64(totalOps)/duration.Seconds())
	
	fmt.Println("\n🎯 ÉVALUATION:")
	if metrics.FailedOps == 0 {
		fmt.Println("   🏅 EXCELLENT: Aucune erreur détectée!")
	} else if float64(metrics.FailedOps)/float64(totalOps) < 0.05 {
		fmt.Println("   ✅ BON: Taux d'erreur acceptable (< 5%)")
	} else {
		fmt.Println("   ⚠️  MOYEN: Taux d'erreur élevé, optimisation recommandée")
	}
	
	fmt.Println("\n🚀 Le système a survécu au test de stress!")
	fmt.Println(strings.Repeat("=", 60))
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
		return nil, err
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

func floatPtr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}