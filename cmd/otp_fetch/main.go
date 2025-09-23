package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/ambroise1219/livraison_go/config"
	"github.com/ambroise1219/livraison_go/database"
	"github.com/ambroise1219/livraison_go/db"
	prismadb "github.com/ambroise1219/livraison_go/prisma/db"
)

func main() {
	phone := flag.String("phone", "", "Phone in +225... format or local")
	flag.Parse()
	if *phone == "" {
		log.Fatal("-phone est requis")
	}

	// Init config + prisma
	_ = config.LoadConfig()
	if err := database.InitPrisma(); err != nil {
		log.Fatalf("init prisma: %v", err)
	}
	defer database.ClosePrisma()
	if err := db.InitializePrisma(); err != nil {
		log.Fatalf("init prisma client: %v", err)
	}
	defer db.ClosePrisma()

	ctx := context.Background()
	otp, err := db.PrismaDB.Otp.FindFirst(
		prismadb.Otp.Phone.Equals(*phone),
	).OrderBy(
		prismadb.Otp.CreatedAt.Order(prismadb.SortOrderDesc),
	).Exec(ctx)
	if err != nil {
		log.Fatalf("fetch otp: %v", err)
	}
	fmt.Print(otp.Code)
}
