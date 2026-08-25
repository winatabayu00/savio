package forecast_test

import (
	"database/sql"
	"encoding/json"
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
	"github.com/savio/savio/backend/internal/forecast"
	"github.com/savio/savio/backend/internal/migrations"
	"github.com/savio/savio/backend/internal/platform/money"
	"github.com/savio/savio/backend/internal/recurring"
	"github.com/savio/savio/backend/internal/transactions"
	"github.com/savio/savio/backend/internal/users"
	"github.com/savio/savio/backend/internal/workspaces"
)

func TestMain(m *testing.M) {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		dsn := v
		testURL, err := url.Parse(dsn)
		if err == nil {
			testURL.Path = "/savio_test_forecast"
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
	_, _ = c.Exec(`CREATE DATABASE savio_test_forecast`)
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

type fx struct {
	wsID   uuid.UUID
	owner  uuid.UUID
	acct   uuid.UUID
	now    time.Time
	txSvc  *transactions.Service
	recSvc *recurring.Service
}

func fixture(t *testing.T) fx {
	t.Helper()
	wsID := uuid.New()
	mustNil(t, db.Create(&workspaces.Workspace{ID: wsID, Name: "T", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)
	owner := uuid.New()
	mustNil(t, db.Create(&users.User{ID: owner, Name: "O", Email: "o-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: owner, Role: "OWNER", Status: "ACTIVE"}).Error)
	acct := uuid.New()
	mustNil(t, db.Create(&accounts.Account{ID: acct, WorkspaceID: wsID, Name: "Main", Type: "CASH", Currency: "IDR",
		OpeningBalance: 1000000, Status: "ACTIVE", Version: 1}).Error)
	now, _ := time.Parse("2006-01-02", "2026-08-25")
	t.Cleanup(func() {
		db.Exec(`DELETE FROM transactions WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM account_transfers WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM recurring_occurrences WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM recurring_transactions WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM audit_logs WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM accounts WHERE id = $1`, acct)
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		db.Exec(`DELETE FROM users WHERE id = $1`, owner)
	})
	return fx{wsID: wsID, owner: owner, acct: acct, now: now, txSvc: transactions.NewService(db), recSvc: recurring.NewService(db)}
}

func (f *fx) addRecurring(t *testing.T, typ, amount string, freq string, startDay string) {
	t.Helper()
	am, _ := moneyMinor(amount)
	start, _ := time.Parse("2006-01-02", startDay)
	_, err := f.recSvc.Create(t.Context(), f.wsID, f.owner, &recurring.CreateInput{
		AccountID: f.acct, Type: typ, AmountMinor: am, Frequency: freq,
		StartDate: start, Description: "rule",
	})
	if err != nil {
		t.Fatalf("recurring: %v", err)
	}
}

func (f *fx) addFutureKnown(t *testing.T, typ, amount string, day string) {
	t.Helper()
	am, _ := moneyMinor(amount)
	dd, _ := time.Parse("2006-01-02", day)
	_, err := f.txSvc.Create(t.Context(), f.wsID, f.owner, &transactions.CreateInput{
		AccountID: f.acct, Type: typ, AmountMinor: am, TransactionDate: dd, Status: "POSTED",
	})
	if err != nil {
		t.Fatalf("known: %v", err)
	}
}

func moneyMinor(s string) (int64, error) { return money.ParseMinorUnits(s) }

func TestForecastDeterministicScheduledEvents(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	// MONTH_END expense starting 08-25 → Aug31, Sep30, Oct31 within 90 days
	// (waves at next 90 days, from-asOf Aug25 through Nov23 -> 3 events)
	f.addRecurring(t, "EXPENSE", "5000", "MONTH_END", "2026-08-25")

	svc := forecast.NewService(db)
	r1, err := svc.Compute(t.Context(), f.wsID, 90, f.now)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	r2, err := svc.Compute(t.Context(), f.wsID, 90, f.now)
	if err != nil {
		t.Fatalf("compute2: %v", err)
	}
	if !jsonEq(r1, r2) {
		t.Fatalf("forecast must be deterministic")
	}
	if r1.OpeningBalance != "10000.00" {
		t.Fatalf("opening = %s", r1.OpeningBalance)
	}
	if r1.ProjectedIncome != "0.00" || r1.ProjectedExpense != "15000.00" {
		t.Fatalf("projections = %s / %s", r1.ProjectedIncome, r1.ProjectedExpense)
	}
	if r1.MinimumBalance != "-5000.00" {
		t.Fatalf("min = %s", r1.MinimumBalance)
	}
	if r1.EndingBalance != "-5000.00" {
		t.Fatalf("ending = %s", r1.EndingBalance)
	}
	if len(r1.Timeline) != 90 {
		t.Fatalf("timeline len = %d", len(r1.Timeline))
	}
	scheduled := 0
	for _, e := range r1.Events {
		if e.Type == "SCHEDULED" {
			scheduled++
		}
		if e.Date < "2026-08-25" {
			t.Fatalf("event before as-of: %s", e.Date)
		}
	}
	if scheduled != 3 {
		t.Fatalf("expected 3 scheduled events, got %d", scheduled)
	}
	if r1.Confidence != "LOW" {
		t.Fatalf("confidence = %s", r1.Confidence)
	}
}

func TestForecastKnownAndEstimatedEventsWithConfidence(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	// future KNOWN income inside horizon (2026-09-10)
	f.addFutureKnown(t, "INCOME", "1000", "2026-09-10")
	// historical expenses for the ESTIMATED baseline (90 days ago)
	am := int64(9000000) // 90 days x 1000
	old := f.now.AddDate(0, 0, -90)
	_, err := f.txSvc.Create(t.Context(), f.wsID, f.owner, &transactions.CreateInput{
		AccountID: f.acct, Type: "EXPENSE", AmountMinor: am, TransactionDate: old, Status: "POSTED",
	})
	if err != nil {
		t.Fatalf("history: %v", err)
	}

	r, err := forecast.NewService(db).Compute(t.Context(), f.wsID, 30, f.now)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if r.ProjectedIncome != "1000.00" {
		t.Fatalf("known income = %s", r.ProjectedIncome)
	}
	// baseline: 90-day window total = 9000000 minor over 90 days → 100000/day (1000.00)
	if r.Assumptions.VariableExpenseDaily != "1000.00" {
		t.Fatalf("variable daily = %s", r.Assumptions.VariableExpenseDaily)
	}
	if r.Confidence != "HIGH" {
		t.Fatalf("confidence = %s", r.Confidence)
	}
}

func jsonEq(a, b *forecast.Result) bool {
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	return string(ja) == string(jb)
}
