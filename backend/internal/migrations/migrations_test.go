package migrations

import (
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// TestMigrateUpDownUp verifies the full migration lifecycle against a real
// PostgreSQL instance. It requires DATABASE_URL to point at an empty database;
// when the service is unreachable the test is skipped (it cannot run in-memory).
func TestMigrateUpDownUp(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://savio:savio@localhost:5433/savio?sslmode=disable"
	}
	if !ping(t, dsn) {
		t.Skip("postgres not reachable; skipping migration test (requires real PostgreSQL)")
	}

	newMigrator := func(t *testing.T) (*migrate.Migrate, func()) {
		t.Helper()
		src, err := iofs.New(FS, Dir)
		if err != nil {
			t.Fatal(err)
		}
		m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
		if err != nil {
			t.Fatal(err)
		}
		return m, func() { src.Close(); m.Close() }
	}

	// Ensure a clean starting state regardless of what the DB currently holds.
	m, close := newMigrator(t)
	for {
		v, _, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) || v == 0 {
			break
		}
		if err != nil {
			t.Fatalf("pre-clean version: %v", err)
		}
		if err := m.Steps(-1); err != nil {
			t.Fatalf("pre-clean: %v", err)
		}
	}
	close()

	// up
	m, close = newMigrator(t)
	if err := m.Up(); err != nil {
		t.Fatalf("up: %v", err)
	}
	close()

	// down
	m, close = newMigrator(t)
	if err := m.Steps(-1); err != nil {
		t.Fatalf("down: %v", err)
	}
	close()

	// up again
	m, close = newMigrator(t)
	if err := m.Up(); err != nil {
		t.Fatalf("up again: %v", err)
	}
	v, dirty, err := m.Version()
	if err != nil {
		t.Fatal(err)
	}
	close()

	if dirty {
		t.Fatal("schema is dirty after up")
	}
	if v == 0 {
		t.Fatal("expected non-zero schema version")
	}

	// verify core tables exist
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, table := range []string{"users", "workspaces", "workspace_memberships", "auth_sessions", "user_settings", "accounts", "categories"} {
		var exists bool
		if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)`, table).Scan(&exists); err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Errorf("expected table %s to exist", table)
		}
	}
}

func ping(t *testing.T, dsn string) bool {
	t.Helper()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return false
	}
	defer db.Close()
	return db.Ping() == nil
}