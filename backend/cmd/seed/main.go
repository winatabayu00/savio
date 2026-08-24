package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/savio/savio/backend/internal/platform/database"
	"github.com/savio/savio/backend/internal/seeds"
)

func main() {
	action := flag.String("action", "categories", "categories | demo")
	flag.Parse()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://savio:savio@localhost:5433/savio?sslmode=disable"
	}
	db, err := database.Connect(dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	ctx := context.Background()

	switch *action {
	case "categories":
		if err := seeds.SeedSystemCategories(ctx, db); err != nil {
			log.Fatalf("seed categories: %v", err)
		}
	case "demo":
		if err := seeds.SeedDemo(ctx, db); err != nil {
			log.Fatalf("seed demo: %v", err)
		}
	default:
		log.Fatalf("unknown action %q", *action)
	}
}
