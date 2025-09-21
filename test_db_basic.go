package main

import (
	"context"
	"fmt"
	"log"

	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

func main() {
	fmt.Println("🔍 Test de connectivité DB de base")
	
	// Initialisation directe du client Prisma
	client := prismadb.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		log.Fatalf("❌ Connexion échouée: %v", err)
	}
	defer client.Prisma.Disconnect()
	
	fmt.Println("✅ Connexion établie!")
	
	// Test de création d'un utilisateur simple
	ctx := context.Background()
	user, err := client.User.CreateOne(
		prismadb.User.Phone.Set("+2250789654321"),
		prismadb.User.FirstName.Set("Test"),
		prismadb.User.LastName.Set("User"),
	).Exec(ctx)
	
	if err != nil {
		log.Fatalf("❌ Erreur création utilisateur: %v", err)
	}
	
	fmt.Printf("✅ Utilisateur créé: %s (%s %s)\n", user.Phone, user.FirstName, user.LastName)
	
	// Test de récupération 
	users, err := client.User.FindMany().Exec(ctx)
	if err != nil {
		log.Fatalf("❌ Erreur récupération: %v", err)
	}
	
	fmt.Printf("✅ %d utilisateur(s) trouvé(s)\n", len(users))
}