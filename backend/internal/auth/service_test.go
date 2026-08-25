package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gorm.io/driver/postgres"

	"github.com/savio/savio/backend/internal/auth/csrf"
	"github.com/savio/savio/backend/internal/migrations"
	"github.com/savio/savio/backend/internal/platform/config"
	"github.com/savio/savio/backend/internal/platform/errs"
)

// TestMain isolates integration tests in their own database so the
// destructive migration test can run against the dev DB in parallel.
func TestMain(m *testing.M) {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		dsn := v
		testURL, err := url.Parse(dsn)
		if err == nil {
			testURL.Path = "/savio_test_auth"
			ensureTestDB(dsn, testURL.String())
			os.Setenv("DATABASE_URL", testURL.String())
		}
	}
	os.Exit(m.Run())
}

func ensureTestDB(adminDSN, testDSN string) {
	db, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return
	}
	defer db.Close()
	_, _ = db.Exec(`CREATE DATABASE savio_test_auth`)
	if err := migrateTestDB(testDSN); err != nil {
		panic(err)
	}
}

func migrateTestDB(dsn string) error {
	src, err := iofs.New(migrations.FS, migrations.Dir)
	if err != nil {
		return err
	}
	defer src.Close()
	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return err
	}
	defer m.Close()
	// fresh state: roll everything down, then up
	for {
		_, _, verr := m.Version()
		if errors.Is(verr, migrate.ErrNilVersion) {
			break
		}
		if verr != nil {
			return verr
		}
		if err := m.Steps(-1); err != nil {
			return err
		}
	}
	return m.Up()
}

func dsn() string {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://savio:savio@localhost:5433/savio?sslmode=disable"
}

func testService(t *testing.T) (*Service, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(postgres.Open(dsn()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Skipf("postgres not reachable: %v", err)
	}
	cfg := &config.Config{
		JWTSecret:       "test-secret-test-secret-test-secret",
		CSRFSecret:      "test-csrf-secret-test-csrf-secret",
		AccessTokenTTL:  time.Minute * 5,
		RefreshTokenTTL: time.Hour,
		CookieSecure:    false,
	}
	return NewService(db, cfg), db
}

func randEmail() string {
	return "t-" + uuid.NewString()[:12] + "@savio.test"
}

func cleanupUser(t *testing.T, db *gorm.DB, userID uuid.UUID) {
	t.Helper()
	t.Cleanup(func() {
		db.Exec(`DELETE FROM auth_sessions WHERE user_id = $1`, userID)
		db.Exec(`DELETE FROM user_settings WHERE user_id = $1`, userID)
		db.Exec(`DELETE FROM workspace_memberships WHERE user_id = $1`, userID)
		db.Exec(`DELETE FROM workspaces WHERE created_by = $1`, userID)
		db.Exec(`DELETE FROM users WHERE id = $1`, userID)
	})
}

func errCode(err error) string {
	var e *errs.Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

func mustNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v (cause: %+v)", err, errs.From(err).Cause)
	}
}

// TestRegisterLoginRefreshRotation verifies the auth happy path plus the
// critical rotation invariant: replaying an already-rotated refresh token
// must be rejected (INV-018).
func TestRegisterLoginRefreshRotation(t *testing.T) {
	svc, db := testService(t)
	ctx := context.Background()
	email := randEmail()

	reg, err := svc.Register(ctx, "Test User", email, "strong-password-1", "test-agent", "127.0.0.1")
	mustNil(t, err)
	cleanupUser(t, db, reg.UserID)
	if reg.Role != "OWNER" || reg.RefreshToken == "" || reg.AccessToken == "" {
		t.Fatalf("unexpected register result: %+v", reg)
	}

	if _, err := svc.Login(ctx, email, "wrong", "agent", "127.0.0.1"); errCode(err) != errs.CodeInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	login, err := svc.Login(ctx, email, "strong-password-1", "test-agent", "127.0.0.1")
	mustNil(t, err)

	first := login.RefreshToken
	afterFirst, err := svc.Refresh(ctx, first, "test-agent")
	mustNil(t, err)

	if _, err := svc.Refresh(ctx, first, "test-agent"); err == nil {
		t.Fatal("refresh replay must fail (INV-018)")
	}

	if _, err := svc.Refresh(ctx, afterFirst.RefreshToken, "test-agent"); err != nil {
		t.Fatalf("second rotation must work: %v", err)
	}
}

// TestRefreshRotationRace ensures concurrent refreshes with the same token do
// not both succeed — the row lock serializes rotation so only one wins (INV-018).
func TestRefreshRotationRace(t *testing.T) {
	svc, db := testService(t)
	ctx := context.Background()
	email := randEmail()
	reg, err := svc.Register(ctx, "Race User", email, "strong-password-2", "agent", "127.0.0.1")
	mustNil(t, err)
	cleanupUser(t, db, reg.UserID)

	login, err := svc.Login(ctx, email, "strong-password-2", "agent", "127.0.0.1")
	mustNil(t, err)
	token := login.RefreshToken

	done := make(chan error, 16)
	for i := 0; i < 8; i++ {
		go func() {
			_, err := svc.Refresh(ctx, token, "agent")
			done <- err
		}()
	}
	successes := 0
	for i := 0; i < 8; i++ {
		if err := <-done; err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly one successful rotation, got %d", successes)
	}
}

// TestLogoutRevokesImmediately ensures a logged-out session cannot fetch a
// fresh token pair.
func TestLogoutRevokesImmediately(t *testing.T) {
	svc, db := testService(t)
	ctx := context.Background()
	email := randEmail()
	reg, err := svc.Register(ctx, "Out User", email, "strong-password-3", "agent", "127.0.0.1")
	mustNil(t, err)
	cleanupUser(t, db, reg.UserID)

	login, err := svc.Login(ctx, email, "strong-password-3", "agent", "127.0.0.1")
	mustNil(t, err)
	sess, err := svc.sessions.FindByHash(ctx, HashToken(login.RefreshToken))
	mustNil(t, err)
	mustNil(t, svc.Logout(ctx, sess.ID))

	if _, err := svc.Refresh(ctx, login.RefreshToken, "agent"); err == nil {
		t.Fatal("refresh after logout must fail")
	}
}

// TestDuplicateEmailRejected covers unique email enforcement.
func TestDuplicateEmailRejected(t *testing.T) {
	svc, db := testService(t)
	ctx := context.Background()
	email := randEmail()
	a, err := svc.Register(ctx, "Dup User", email, "strong-password-4", "agent", "127.0.0.1")
	mustNil(t, err)
	cleanupUser(t, db, a.UserID)
	if _, err := svc.Register(ctx, "Dup User Two", email, "strong-password-4", "agent", "127.0.0.1"); err == nil {
		t.Fatal("duplicate email must be rejected")
	}
}

// TestUpdateSettings verifies partial preference updates persist and invalid
// values are rejected.
func TestUpdateSettings(t *testing.T) {
	svc, db := testService(t)
	ctx := context.Background()
	email := randEmail()
	reg, err := svc.Register(ctx, "Settings User", email, "strong-password-5", "agent", "127.0.0.1")
	mustNil(t, err)
	cleanupUser(t, db, reg.UserID)

	threshold := 85.0
	low := int64(5000000)
	enabled := false
	upd, err := svc.UpdateSettings(ctx, reg.UserID, &UpdateSettingsInput{
		AIInsightsEnabled:      &enabled,
		BudgetWarningThreshold: &threshold,
		LowBalanceThreshold:    &low,
	})
	mustNil(t, err)
	if upd.AIInsightsEnabled || upd.BudgetWarningThreshold != 85 || upd.LowBalanceThreshold == nil || *upd.LowBalanceThreshold != 5000000 {
		t.Fatalf("settings not applied: %+v", upd)
	}

	bad := 150.0
	if _, err := svc.UpdateSettings(ctx, reg.UserID, &UpdateSettingsInput{BudgetWarningThreshold: &bad}); errCode(err) != errs.CodeValidationError {
		t.Fatalf("expected validation error, got %v", err)
	}

	neg := int64(-1)
	if _, err := svc.UpdateSettings(ctx, reg.UserID, &UpdateSettingsInput{LowBalanceThreshold: &neg}); errCode(err) != errs.CodeValidationError {
		t.Fatalf("expected validation error for negative, got %v", err)
	}
}

// TestCSRFValidate exercises signed double-submit validation invariants.
func TestCSRFValidate(t *testing.T) {
	secret := "csrf-test-secret"
	cookie, err := csrf.Generate(secret)
	if err != nil {
		t.Fatal(err)
	}
	if err := csrf.Validate(secret, cookie, cookie); err != nil {
		t.Fatalf("expected valid: %v", err)
	}
	if err := csrf.Validate(secret, cookie, "evil"); err == nil {
		t.Fatal("mismatched header must fail")
	}
	if err := csrf.Validate("other-secret", cookie, cookie); err == nil {
		t.Fatal("wrong secret signature must fail")
	}
}
