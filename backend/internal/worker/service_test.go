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
	"github.com/savio/savio/backend/internal/budgets"
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
			testURL.Path = "/savio_test_worker"
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
	_, _ = c.Exec(`CREATE DATABASE savio_test_worker`)
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

// TestBudgetSweepIsWorkspaceScoped proves a workspace's budget warning is
// computed only from that workspace's spend, even for shared system
// categories (regression for the old cross-workspace, all-time sweep).
func TestBudgetSweepIsWorkspaceScoped(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	wsa, wsb := uuid.New(), uuid.New()
	mkWS := func(id uuid.UUID, name string) *users.User {
		owner := uuid.New()
		mustNil(t, db.Create(&workspaces.Workspace{ID: id, Name: name, BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)
		mustNil(t, db.Create(&users.User{ID: owner, Name: "O", Email: "o-" + uuid.NewString()[:10] + "@savio.test",
			PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
		mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: id, UserID: owner, Role: "OWNER", Status: "ACTIVE"}).Error)
		return &users.User{ID: owner}
	}
	ownerA := mkWS(wsa, "A")
	ownerB := mkWS(wsb, "B")
	cat := uuid.New()
	mustNil(t, db.Exec(`INSERT INTO categories (id, name, type, is_system, status, created_at, updated_at)
		VALUES (?, 'Food & Dining', 'EXPENSE', TRUE, 'ACTIVE', NOW(), NOW())`, cat).Error)
	leave := func() {
		for _, ws := range []uuid.UUID{wsa, wsb} {
			db.Exec(`DELETE FROM notifications WHERE workspace_id = $1`, ws)
			db.Exec(`DELETE FROM transactions WHERE workspace_id = $1`, ws)
			db.Exec(`DELETE FROM budgets WHERE workspace_id = $1`, ws)
			db.Exec(`DELETE FROM accounts WHERE workspace_id = $1`, ws)
			db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, ws)
			db.Exec(`DELETE FROM workspaces WHERE id = $1`, ws)
		}
		db.Exec(`DELETE FROM categories WHERE id = $1`, cat)
		db.Exec(`DELETE FROM users WHERE id IN (?, ?)`, ownerA.ID, ownerB.ID)
	}
	t.Cleanup(leave)

	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1)
	mkBudget := func(ws uuid.UUID, amount int64) uuid.UUID {
		id := uuid.New()
		mustNil(t, db.Create(&budgets.Budget{
			ID: id, WorkspaceID: ws, CategoryID: cat, Amount: amount,
			PeriodStart: start, PeriodEnd: end, Status: "ACTIVE", Version: 1,
			CreatedAt: now, UpdatedAt: now,
		}).Error)
		return id
	}
	mkAcct := func(ws uuid.UUID) uuid.UUID {
		id := uuid.New()
		mustNil(t, db.Create(&accounts.Account{ID: id, WorkspaceID: ws, Name: "A", Type: "CASH", Currency: "IDR",
			OpeningBalance: 0, Status: "ACTIVE", Version: 1}).Error)
		return id
	}
	acctA, acctB := mkAcct(wsa), mkAcct(wsb)
	// WS-A spends 120% of its 100k budget -> EXCEEDED.
	budgetA := mkBudget(wsa, 100000)
	mustNil(t, db.Create(&transactions.Transaction{
		ID: uuid.New(), WorkspaceID: wsa, AccountID: acctA, CategoryID: &cat,
		Amount: 120000, Type: "EXPENSE", Status: "POSTED", TransactionDate: start,
		CreatedAt: now, UpdatedAt: now,
	}).Error)
	// WS-B spends only 5% of its 100k budget.
	mkBudget(wsb, 100000)
	mustNil(t, db.Create(&transactions.Transaction{
		ID: uuid.New(), WorkspaceID: wsb, AccountID: acctB, CategoryID: &cat,
		Amount: 5000, Type: "EXPENSE", Status: "POSTED", TransactionDate: start,
		CreatedAt: now, UpdatedAt: now,
	}).Error)

	mustNil(t, db.Table("user_settings").Where("user_id = ?", ownerA.ID).Update("budget_warning_threshold", 80).Error)

	if err := worker.NewService(db).SweepNotifications(t.Context(), now); err != nil {
		t.Fatalf("sweep: %v", err)
	}

	var warnA, warnB int64
	db.Table("notifications").Where("workspace_id = ? AND type = 'BUDGET_EXCEEDED'", wsa).Count(&warnA)
	db.Table("notifications").Where("workspace_id = ? AND (type = 'BUDGET_EXCEEDED' OR type = 'BUDGET_WARNING')", wsb).Count(&warnB)
	if warnA != 1 {
		t.Fatalf("expected WS-A exceed notification, got %d (budget %v)", warnA, budgetA)
	}
	if warnB != 0 {
		t.Fatalf("cross-workspace contamination: WS-B notified %v times from WS-A spend", warnB)
	}
}

// TestLowBalanceSweepIsPerAccountAndIncludesTransfers proves the sweep uses
// the authoritative per-account derived balance (opening + posted
// transactions + transfers), not a workspace-wide transaction sum.
func TestLowBalanceSweepIsPerAccountAndIncludesTransfers(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	wsID := uuid.New()
	mustNil(t, db.Create(&workspaces.Workspace{ID: wsID, Name: "T", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)
	owner := uuid.New()
	mustNil(t, db.Create(&users.User{ID: owner, Name: "O", Email: "o-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: owner, Role: "OWNER", Status: "ACTIVE"}).Error)
	acctA, acctB := uuid.New(), uuid.New()
	for _, a := range []struct {
		id      uuid.UUID
		name    string
		opening int64
	}{{acctA, "A", 5000}, {acctB, "B", 5000}} {
		mustNil(t, db.Create(&accounts.Account{ID: a.id, WorkspaceID: wsID, Name: a.name, Type: "CASH", Currency: "IDR",
			OpeningBalance: a.opening, Status: "ACTIVE", Version: 1}).Error)
	}
	// EXPENSE 3000 on B only: B=2000, A=5000.
	mustNil(t, db.Create(&transactions.Transaction{
		ID: uuid.New(), WorkspaceID: wsID, AccountID: acctB, Amount: 3000, Type: "EXPENSE",
		Status: "POSTED", TransactionDate: time.Now().UTC(), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}).Error)
	mkCat := uuid.New()
	mustNil(t, db.Exec(`INSERT INTO categories (id, name, type, is_system, workspace_id, status, created_at, updated_at)
		VALUES (?, 'Transfer', 'EXPENSE', FALSE, ?, 'ACTIVE', NOW(), NOW())`, mkCat, wsID).Error)
	// Transfer 1000 B->A: B=1000, A=6000.
	mustNil(t, db.Exec(`INSERT INTO account_transfers (id, workspace_id, from_account_id, to_account_id, amount, transfer_date, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, 1000, ?, 'POSTED', NOW(), NOW())`,
		uuid.New(), wsID, acctB, acctA, time.Now().UTC()).Error)
	threshold := int64(1500)
	mustNil(t, db.Create(&users.UserSettings{UserID: owner, LowBalanceThreshold: &threshold}).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM notifications WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM account_transfers WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM transactions WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM categories WHERE id = $1`, mkCat)
		db.Exec(`DELETE FROM accounts WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM user_settings WHERE user_id = $1`, owner)
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		db.Exec(`DELETE FROM users WHERE id = $1`, owner)
	})

	if err := worker.NewService(db).SweepNotifications(t.Context(), time.Now().UTC()); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	var n int64
	if err := db.Table("notifications").Where("workspace_id = ? AND type = 'LOW_BALANCE'", wsID).Count(&n).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	// B is at 1000 <= 1500; A at 6000. Exactly one account should notify.
	if n != 1 {
		t.Fatalf("expected 1 low-balance notification (account B), got %d", n)
	}
}

var _ = transactions.NewService
