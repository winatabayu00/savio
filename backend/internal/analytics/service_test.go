package analytics_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/savio/savio/backend/internal/accounts"
	"github.com/savio/savio/backend/internal/analytics"
	"github.com/savio/savio/backend/internal/migrations"
	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/money"
	"github.com/savio/savio/backend/internal/seeds"
	"github.com/savio/savio/backend/internal/transactions"
	"github.com/savio/savio/backend/internal/transfers"
	"github.com/savio/savio/backend/internal/users"
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

func mustNil2(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type fx struct {
	wsID   uuid.UUID
	owner  uuid.UUID
	acctA  uuid.UUID
	acctB  uuid.UUID
	incCat uuid.UUID
	expCat uuid.UUID
	txSvc  *transactions.Service
}

func fixture(t *testing.T) fx {
	t.Helper()
	mustNil2(t, seeds.SeedSystemCategories(t.Context(), db))
	wsID := uuid.New()
	mustNil2(t, db.Create(&workspaces.Workspace{ID: wsID, Name: "T", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)
	owner := uuid.New()
	mustNil2(t, db.Create(&users.User{ID: owner, Name: "O", Email: "o-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil2(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: owner, Role: "OWNER", Status: "ACTIVE"}).Error)
	a := uuid.New()
	b := uuid.New()
	mustNil2(t, db.Create(&accounts.Account{ID: a, WorkspaceID: wsID, Name: "A", Type: "CASH", Currency: "IDR",
		OpeningBalance: 10000000, Status: "ACTIVE", Version: 1}).Error)
	mustNil2(t, db.Create(&accounts.Account{ID: b, WorkspaceID: wsID, Name: "B", Type: "BANK", Currency: "IDR",
		OpeningBalance: 0, Status: "ACTIVE", Version: 1}).Error)
	var inc, exp struct{ ID uuid.UUID }
	mustNil2(t, db.Table("categories").Where("name = ? AND type = 'INCOME'", "Salary").First(&inc).Error)
	mustNil2(t, db.Table("categories").Where("name = ? AND type = 'EXPENSE'", "Food & Dining").First(&exp).Error)

	t.Cleanup(func() {
		db.Exec(`DELETE FROM transactions WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM account_transfers WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM audit_logs WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM accounts WHERE id IN ($1,$2)`, a, b)
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		db.Exec(`DELETE FROM users WHERE id = $1`, owner)
	})
	return fx{wsID: wsID, owner: owner, acctA: a, acctB: b, incCat: inc.ID, expCat: exp.ID, txSvc: transactions.NewService(db)}
}

func (f *fx) add(t *testing.T, typ, amount string, day string, cat *uuid.UUID, status string) *transactions.View {
	t.Helper()
	minor, err := money.ParseMinorUnits(amount)
	if err != nil {
		t.Fatalf("amount %q: %v", amount, err)
	}
	var catPtr *uuid.UUID
	if cat != nil {
		catPtr = cat
	}
	v, err := f.txSvc.Create(t.Context(), f.wsID, f.owner, &transactions.CreateInput{
		AccountID: f.acctA, CategoryID: catPtr, Type: typ, AmountMinor: minor,
		TransactionDate: d("2026-08-" + day), Status: status, Description: typ + "-" + day,
	})
	if err != nil {
		t.Fatalf("add tx: %v", err)
	}
	return v
}

func newRouter(t *testing.T, wsID uuid.UUID, h *analytics.Handler) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authMW := func(c *gin.Context) {
		uid, err := uuid.Parse(c.GetHeader("X-Test-User"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false})
			return
		}
		var m workspaces.Membership
		if err := db.Where("workspace_id = ? AND user_id = ? AND status = 'ACTIVE'", wsID, uid).First(&m).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"success": false})
			return
		}
		authctx.Set(c, &authctx.Ctx{UserID: uid, WorkspaceID: wsID, WorkspaceRole: authctx.Role(m.Role), IsAuthenticated: true})
		c.Next()
	}
	g := r.Group("/api/v1/analytics", authMW)
	analytics.RegisterRoutes(g, h)
	return r
}

func doReq(t *testing.T, r *gin.Engine, method, path, userID, body string) *httptest.ResponseRecorder {
	t.Helper()
	var buf *bytes.Buffer
	if body == "" {
		buf = bytes.NewBuffer(nil)
	} else {
		buf = bytes.NewBufferString(body)
	}
	req := httptest.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", userID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return out
}

func d(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// TestAnalyticsExclusions verifies transfers, voided and adjustments are
// excluded from ordinary income/expense metrics (AGENTS #35).
func TestAnalyticsExclusions(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	// income + expense
	f.add(t, "INCOME", "5000", "02", &f.incCat, "POSTED")
	f.add(t, "EXPENSE", "1000", "03", &f.expCat, "POSTED")
	// voided expense must not count
	exp := f.add(t, "EXPENSE", "900", "04", &f.expCat, "POSTED")
	f.txSvc.Void(t.Context(), f.wsID, f.owner, &transactions.VoidInput{ID: exp.ID, Version: exp.Version, Reason: "x"})
	// adjustment must not count as income/expense
	f.add(t, "ADJUSTMENT", "777", "05", nil, "POSTED")

	r := newRouter(t, f.wsID, analytics.NewHandler(analytics.NewService(db)))
	w := doReq(t, r, "GET", "/api/v1/analytics/cashflow?from=2026-08-01&to=2026-08-31", f.owner.String(), "")
	cash := decodeBody(t, w)["data"].(map[string]any)
	if cash["income"].(string) != "5000.00" {
		t.Fatalf("income = %v, want 5000.00", cash["income"])
	}
	if cash["expense"].(string) != "1000.00" {
		t.Fatalf("expense = %v, want 1000.00", cash["expense"])
	}
	if cash["net"].(string) != "4000.00" {
		t.Fatalf("net = %v", cash["net"])
	}
}

// TestAnalyticsTransfersExcluded verifies transfers never appear as income or
// expense even though they move between accounts.
func TestAnalyticsTransfersExcluded(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	f.add(t, "INCOME", "5000", "02", &f.incCat, "POSTED")
	_, err := transfers.NewService(db).Create(t.Context(), f.wsID, f.owner, &transfers.CreateInput{
		FromAccountID: f.acctA, ToAccountID: f.acctB, AmountMinor: 100000, TransferDate: d("2026-08-06"),
	})
	if err != nil {
		t.Fatalf("transfer: %v", err)
	}
	r := newRouter(t, f.wsID, analytics.NewHandler(analytics.NewService(db)))
	w := doReq(t, r, "GET", "/api/v1/analytics/cashflow?from=2026-08-01&to=2026-08-31", f.owner.String(), "")
	cash := decodeBody(t, w)["data"].(map[string]any)
	if cash["income"].(string) != "5000.00" {
		t.Fatalf("income = %v, want 5000.00", cash["income"])
	}
	if cash["expense"].(string) != "0.00" {
		t.Fatalf("expense = %v, want 0.00", cash["expense"])
	}
}

func TestAnalyticsCategoryBreakdown(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	f.add(t, "EXPENSE", "200", "02", &f.expCat, "POSTED")
	f.add(t, "EXPENSE", "300", "03", &f.expCat, "POSTED")
	r := newRouter(t, f.wsID, analytics.NewHandler(analytics.NewService(db)))
	w := doReq(t, r, "GET", "/api/v1/analytics/categories?from=2026-08-01&to=2026-08-31", f.owner.String(), "")
	rows := decodeBody(t, w)["data"].([]any)
	if len(rows) != 1 {
		t.Fatalf("expected 1 category row, got %d", len(rows))
	}
	row := rows[0].(map[string]any)
	if row["total"].(string) != "500.00" {
		t.Fatalf("category total = %v", row["total"])
	}
}

func TestAnalyticsPeriodComparison(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	f.add(t, "INCOME", "1000", "05", &f.incCat, "POSTED")
	r := newRouter(t, f.wsID, analytics.NewHandler(analytics.NewService(db)))
	w := doReq(t, r, "GET", "/api/v1/analytics/period-comparison?from=2026-09-01&to=2026-09-30&compare_from=2026-08-01&compare_to=2026-08-31", f.owner.String(), "")
	body := decodeBody(t, w)["data"].(map[string]any)
	current := body["current"].(map[string]any)
	previous := body["previous"].(map[string]any)
	if previous["income"].(string) != "1000.00" || current["income"].(string) != "0.00" {
		t.Fatalf("comparison failed: %v / %v", current, previous)
	}
}

func TestAnalyticsDashboardTotal(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	f.add(t, "INCOME", "500", "02", &f.incCat, "POSTED")
	f.add(t, "EXPENSE", "150", "03", &f.expCat, "POSTED")
	v, err := analytics.NewService(db).Dashboard(t.Context(), f.wsID, d("2026-08-15"))
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	// opening 10000000 + 50000 - 15000
	want := int64(10000000) + 50000 - 15000
	if v.TotalBalance != want {
		t.Fatalf("total balance = %d, want %d", v.TotalBalance, want)
	}
	if len(v.Accounts) != 2 || len(v.Upcoming) > 0 || len(v.Recent) != 2 {
		t.Fatalf("dashboard shape off: acc=%d up=%d recent=%d", len(v.Accounts), len(v.Upcoming), len(v.Recent))
	}
}

