package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/savio/savio/backend/internal/migrations"
)

// Savio schema is controlled exclusively by these explicit migrations.
// GORM AutoMigrate is never used as a production schema source of truth.

func main() {
	action := flag.String("action", "up", "up | down | down-all | force | version")
	version := flag.Int("version", 1, "version for force")
	flag.Parse()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://savio:savio@localhost:5433/savio?sslmode=disable"
	}

	src, err := iofs.New(migrations.FS, migrations.Dir)
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		log.Fatalf("migrate init: %v", err)
	}

	apply := func(fn func() error, verbose string) {
		if err := fn(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("migrate: no change")
				return
			}
			log.Fatalf("%s: %v", verbose, err)
		}
		fmt.Printf("migrate: %s ok\n", verbose)
	}

	switch *action {
	case "up":
		apply(m.Up, "up")
	case "down":
		apply(func() error { return m.Steps(-1) }, "down 1 step")
	case "down-all":
		apply(func() error {
			for {
				err := m.Steps(-1)
				if err == nil || errors.Is(err, migrate.ErrNoChange) {
					return nil
				}
				return err
			}
		}, "down-all")
	case "force":
		if err := m.Force(*version); err != nil {
			log.Fatalf("force: %v", err)
		}
		fmt.Printf("migrate: forced version %d\n", *version)
	case "version":
		v, dirty, err := m.Version()
		if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
			log.Fatalf("version: %v", err)
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
	default:
		log.Fatalf("unknown action %q", *action)
	}
}
