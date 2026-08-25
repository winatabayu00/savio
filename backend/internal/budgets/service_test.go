package budgets_test

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
	"github.com/savio/savio/backend/internal/auth"
	"github.com/savio/savio/backend/internal/budgets"
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
			testURL.Path = "/savio_test_budgets"
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
	_, _ = c.Exec(`CREATE DATABASE savio_test_budgets`)
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
	wsID    uuid.UUID
	owner   uuid.UUID
	acct    uuid.UUID
	acctB   uuid.UUID
	foodCat uuid.UUID
	txSvc   *transactions.Service
}

func fixture(t *testing.T) fx {
	t.Helper()
	mustNil(t, seeds.SeedSystemCategories(t.Context(), db))
	wsID := uuid.New()
	mustNil(t, db.Create(&workspaces.Workspace{ID: wsID, Name: "T", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)
	owner := uuid.New()
	mustNil(t, db.Create(&users.User{ID: owner, Name: "O", Email: "o-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: owner, Role: "OWNER", Status: "ACTIVE"}).Error)
	a := uuid.New()
	b := uuid.New()
	mustNil(t, db.Create(&accounts.Account{ID: a, WorkspaceID: wsID, Name: "A", Type: "CASH", Currency: "IDR",
		OpeningBalance: 10000000, Status: "ACTIVE", Version: 1}).Error)
	mustNil(t, db.Create(&accounts.Account{ID: b, WorkspaceID: wsID, Name: "B", Type: "BANK", Currency: "IDR",
		OpeningBalance: 0, Status: "ACTIVE", Version: 1}).Error)
	var food struct{ ID uuid.UUID }
	mustNil(t, db.Table("categories").Where("name = ? AND type = 'EXPENSE'", "Food & Dining").First(&food).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM transactions WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM account_transfers WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM audit_logs WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM budgets WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM accounts WHERE id IN ($1,$2)`, a, b)
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		db.Exec(`DELETE FROM users WHERE id = $1`, owner)
	})
	return fx{wsID: wsID, owner: owner, acct: a, acctB: b, foodCat: food.ID, txSvc: transactions.NewService(db)}
}

func (f *fx) expense(t *testing.T, amount string, day string, status string) *transactions.View {
	t.Helper()
	minor, err := money.ParseMinorUnits(amount)
	if err != nil {
		t.Fatalf("amount %q: %v", amount, err)
	}
	v, err := f.txSvc.Create(t.Context(), f.wsID, f.owner, &transactions.CreateInput{
		AccountID: f.acct, CategoryID: &f.foodCat, Type: "EXPENSE", AmountMinor: minor,
		TransactionDate: date(day), Status: status, Description: "e" + day,
	})
	if err != nil {
		t.Fatalf("expense: %v", err)
	}
	return v
}

func date(day string) time.Time {
	t, _ := time.Parse("2006-01-02", "2026-08-"+day)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func newRouter(t *testing.T, wsID uuid.UUID, h *budgets.Handler) *gin.Engine {
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
	g := r.Group("/api/v1/budgets", authMW)
	budgets.RegisterRoutes(g, h, auth.RequireWrite())
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

func TestBudgetDerivedSpendAndStatus(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	// spent 200 before budget edit covers 1000 threshold check; build budget
	f.expense(t, "200", "03", "POSTED") // 20000 minor
	f.expense(t, "300", "04", "POSTED") // 30000 minor

	r := newRouter(t, f.wsID, budgets.NewHandler(budgets.NewService(db)))
	w := doReq(t, r, "POST", "/api/v1/budgets", f.owner.String(),
		`{"category_id":"`+f.foodCat.String()+`","amount":"1000","period_start":"2026-08-01","period_end":"2026-08-31"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	data := decodeBody(t, w)["data"].(map[string]any)
	// spent 500 major, budget 1000 → ON_TRACK, utilization 50
	if data["spent"].(string) != "500.00" || data["computed_status"].(string) != "ON_TRACK" {
		t.Fatalf("derived = %v / %v", data["spent"], data["computed_status"])
	}
	util := data["utilization_percent"].(float64)
	if util < 49.9 || util > 50.1 {
		t.Fatalf("utilization = %v", util)
	}
}

func TestBudgetDuplicatesAndCloseUnblocks(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID, budgets.NewHandler(budgets.NewService(db)))
	mk := func() string {
		return `{"category_id":"` + f.foodCat.String() + `","amount":"500","period_start":"2026-08-01","period_end":"2026-08-31"}`
	}
	w := doReq(t, r, "POST", "/api/v1/budgets", f.owner.String(), mk())
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	id := decodeBody(t, w)["data"].(map[string]any)["id"].(string)
	version := int64(decodeBody(t, w)["data"].(map[string]any)["version"].(float64))
	// overlapping duplicate rejected
	w = doReq(t, r, "POST", "/api/v1/budgets", f.owner.String(), mk())
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate expected 409, got %d %s", w.Code, w.Body.String())
	}
	closeBody, _ := json.Marshal(map[string]any{"version": version})
	w = doReq(t, r, "POST", "/api/v1/budgets/"+id+"/close", f.owner.String(), string(closeBody))
	if w.Code != http.StatusOK || decodeBody(t, w)["data"].(map[string]any)["status"].(string) != "CLOSED" {
		t.Fatalf("close: %d %s", w.Code, w.Body.String())
	}
	w = doReq(t, r, "POST", "/api/v1/budgets", f.owner.String(), mk())
	if w.Code != http.StatusCreated {
		t.Fatalf("after close new budget should succeed, got %d %s", w.Code, w.Body.String())
	}
}

func TestBudgetVoidedAndTransferExcluded(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	exp := f.expense(t, "100", "03", "POSTED") // 10000 minor
	f.txSvc.Void(t.Context(), f.wsID, f.owner, &transactions.VoidInput{ID: exp.ID, Version: exp.Version, Reason: "x"})
	_, err := transfers.NewService(db).Create(t.Context(), f.wsID, f.owner, &transfers.CreateInput{
		FromAccountID: f.acct, ToAccountID: f.acctB, AmountMinor: 50000, TransferDate: date("05"),
	})
	if err != nil {
		t.Fatalf("transfer: %v", err)
	}
	r := newRouter(t, f.wsID, budgets.NewHandler(budgets.NewService(db)))
	w := doReq(t, r, "POST", "/api/v1/budgets", f.owner.String(),
		`{"category_id":"`+f.foodCat.String()+`","amount":"1000","period_start":"2026-08-01","period_end":"2026-08-31"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	spent := decodeBody(t, w)["data"].(map[string]any)["spent"].(string)
	if spent != "0.00" {
		t.Fatalf("spent should exclude voided + transfers, got %s", spent)
	}
}

func TestBudgetVersionConflict(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID, budgets.NewHandler(budgets.NewService(db)))
	w := doReq(t, r, "POST", "/api/v1/budgets", f.owner.String(),
		`{"category_id":"`+f.foodCat.String()+`","amount":"500","period_start":"2026-08-01","period_end":"2026-08-31"}`)
	id := decodeBody(t, w)["data"].(map[string]any)["id"].(string)
	_, v := decodedVersion(t, w)
	stale, _ := json.Marshal(map[string]any{"category_id": f.foodCat.String(), "amount": "600", "period_start": "2026-08-01", "period_end": "2026-08-31", "version": v})
	w = doReq(t, r, "PATCH", "/api/v1/budgets/"+id, f.owner.String(), string(stale))
	if w.Code != http.StatusOK {
		t.Fatalf("first update should succeed, got %d %s", w.Code, w.Body.String())
	}
	w = doReq(t, r, "PATCH", "/api/v1/budgets/"+id, f.owner.String(), string(stale))
	if w.Code != http.StatusConflict {
		t.Fatalf("stale version expected 409, got %d %s", w.Code, w.Body.String())
	}
}

func decodedVersion(t *testing.T, w *httptest.ResponseRecorder) (string, int64) {
	t.Helper()
	d := decodeBody(t, w)["data"].(map[string]any)
	return d["id"].(string), int64(d["version"].(float64))
}

// TestBudgetRejectsInvalidCategory proves budgets only target ACTIVE EXPENSE
// categories available to the workspace (system or custom). INCOME, foreign,
// or missing categories must be rejected with 422.
func TestBudgetRejectsInvalidCategory(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	seedSystem := func() {
		mustNil(t, seeds.SeedSystemCategories(t.Context(), db))
	}
	seedSystem()
	var incomeCat struct{ ID uuid.UUID }
	mustNil(t, db.Table("categories").Where("name = ? AND type = 'INCOME'", "Salary").First(&incomeCat).Error)
	r := newRouter(t, f.wsID, budgets.NewHandler(budgets.NewService(db)))
	mk := func(catID string) string {
		return `{"category_id":"` + catID + `","amount":"500","period_start":"2026-08-01","period_end":"2026-08-31"}`
	}
	// INCOME category on an expense budget.
	w := doReq(t, r, "POST", "/api/v1/budgets", f.owner.String(), mk(incomeCat.ID.String()))
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("income category expected 422, got %d %s", w.Code, w.Body.String())
	}
	// Nonexistent category.
	w = doReq(t, r, "POST", "/api/v1/budgets", f.owner.String(), mk(uuid.New().String()))
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("missing category expected 422, got %d %s", w.Code, w.Body.String())
	}
	// Foreign-workspace custom EXPENSE category.
	otherWS := uuid.New()
	mustNil(t, db.Create(&workspaces.Workspace{ID: otherWS, Name: "X", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)
	otherCat := uuid.New()
	mustNil(t, db.Exec(`INSERT INTO categories (id, workspace_id, name, type, is_system, status, created_at, updated_at)
		VALUES (?, ?, 'Private', 'EXPENSE', FALSE, 'ACTIVE', NOW(), NOW())`, otherCat, otherWS).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM categories WHERE id = $1`, otherCat)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, otherWS)
	})
	w = doReq(t, r, "POST", "/api/v1/budgets", f.owner.String(), mk(otherCat.String()))
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("foreign category expected 422, got %d %s", w.Code, w.Body.String())
	}
	// System EXPENSE category is valid.
	w = doReq(t, r, "POST", "/api/v1/budgets", f.owner.String(), mk(f.foodCat.String()))
	if w.Code != http.StatusCreated {
		t.Fatalf("system expense category expected 201, got %d %s", w.Code, w.Body.String())
	}
}
