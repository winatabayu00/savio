package worker_test

import (
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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/savio/savio/backend/internal/accounts"
	"github.com/savio/savio/backend/internal/migrations"
	"github.com/savio/savio/backend/internal/recurring"
	"github.com/savio/savio/backend/internal/transactions"
	"github.com/savio/savio/backend/internal/users"
	"github.com/savio/savio/backend/internal/worker"
	"github.com/savio/savio/backend/internal/workspaces"
)

func TestMain(m *testing.M) {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		dsn := v
		testURL, err := url.Parse(dsn)
		if err == nil {
			testURL.Path = "/savio_test"
			ensureTestDB(dsn, testURL.String())
			os.Setenv("DATABASE_URL", testURL.String())
		}
	}
	openDB()
	code := m.Run()
	os.Exit(code)
}

var db *gorm.DB

func openDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return
	}
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		panic(err)
	}
}

func ensureTestDB(adminDSN, testDSN string) {
	c, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return
	}
	defer c.Close()
	_, _ = c.Exec(`CREATE DATABASE savio_test`)
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

func mustNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func derived(t *testing.T, wsID, acctID uuid.UUID) int64 {
	t.Helper()
	v, err := accounts.NewService(db).Get(t.Context(), wsID, acctID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	return v.DerivedBalance
}

func TestWorkerAutoPostIsIdempotent(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	wsID := uuid.New()
	mustNil(t, db.Create(&workspaces.Workspace{ID: wsID, Name: "T", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)
	owner := uuid.New()
	mustNil(t, db.Create(&users.User{ID: owner, Name: "O", Email: "o-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: owner, Role: "OWNER", Status: "ACTIVE"}).Error)
	acct := uuid.New()
	mustNil(t, db.Create(&accounts.Account{ID: acct, WorkspaceID: wsID, Name: "A", Type: "CASH", Currency: "IDR",
		OpeningBalance: 10000000, Status: "ACTIVE", Version: 1}).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM transactions WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM recurring_occurrences WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM recurring_transactions WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM notifications WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM audit_logs WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM accounts WHERE id = $1`, acct)
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		db.Exec(`DELETE FROM users WHERE id = $1`, owner)
	})

	start := time.Now().UTC().AddDate(0, 0, -3)
	_, err := recurring.NewService(db).Create(t.Context(), wsID, owner, &recurring.CreateInput{
		AccountID: acct, Type: "EXPENSE", AmountMinor: 100000, Frequency: "DAILY",
		StartDate: start, AutoPost: true, Description: "auto gym",
	})
	if err != nil {
		t.Fatalf("create rule: %v", err)
	}

	recSvc := recurring.NewService(db)
	var probe struct{ N int64 }
	db.Table("recurring_occurrences").Count(&probe.N)
	t.Logf("occurrences=%d", probe.N)
	var dueProbe struct{ N int64 }
	db.Raw(`SELECT COUNT(*) AS n FROM recurring_occurrences o JOIN recurring_transactions rt ON rt.id=o.recurring_id
		WHERE o.workspace_id=$1 AND o.status='PENDING' AND rt.auto_post AND rt.status='ACTIVE' AND o.due_date<=CURRENT_DATE`, wsID).Scan(&dueProbe)
	t.Logf("due+autopost=%d", dueProbe.N)
	posted, err := recSvc.AutoPostDue(t.Context(), time.Now().UTC())
	if err != nil {
		t.Fatalf("autopost: %v", err)
	}
	if posted < 1 {
		t.Fatalf("expected at least one due occurrence posted, got %d", posted)
	}
	got := derived(t, wsID, acct)
	if got >= 10000000 {
		t.Fatalf("expected posted expense to reduce balance, got %d", got)
	}
	// second run: nothing new
	posted2, err := recSvc.AutoPostDue(t.Context(), time.Now().UTC())
	if err != nil {
		t.Fatalf("autopost2: %v", err)
	}
	if posted2 != 0 {
		t.Fatalf("second run should post nothing, got %d", posted2)
	}
	if got2 := derived(t, wsID, acct); got2 != got {
		t.Fatalf("balance changed on second run: %d vs %d", got2, got)
	}
}

func TestWorkerNotificationDedup(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	wsID := uuid.New()
	mustNil(t, db.Create(&workspaces.Workspace{ID: wsID, Name: "T", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)
	owner := uuid.New()
	mustNil(t, db.Create(&users.User{ID: owner, Name: "O", Email: "o-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: owner, Role: "OWNER", Status: "ACTIVE"}).Error)
	threshold := int64(1) // 0.01 → every account below threshold
	mustNil(t, db.Create(&users.UserSettings{UserID: owner, LowBalanceThreshold: &threshold}).Error)
	acct := uuid.New()
	mustNil(t, db.Create(&accounts.Account{ID: acct, WorkspaceID: wsID, Name: "Empty", Type: "CASH", Currency: "IDR",
		OpeningBalance: 0, Status: "ACTIVE", Version: 1}).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM notifications WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM transactions WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM audit_logs WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM accounts WHERE id = $1`, acct)
		db.Exec(`DELETE FROM user_settings WHERE user_id = $1`, owner)
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		db.Exec(`DELETE FROM users WHERE id = $1`, owner)
	})

	svc := worker.NewService(db)
	if err := svc.SweepNotifications(t.Context(), time.Now().UTC()); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if err := svc.SweepNotifications(t.Context(), time.Now().UTC()); err != nil {
		t.Fatalf("sweep2: %v", err)
	}
	var n int64
	if err := db.Table("notifications").Where("workspace_id = ? AND type = 'LOW_BALANCE'", wsID).Count(&n).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected exactly 1 notification (dedup), got %d", n)
	}
}

var _ = transactions.NewService