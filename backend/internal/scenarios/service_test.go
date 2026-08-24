package scenarios_test

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
	"github.com/savio/savio/backend/internal/migrations"
	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/money"
	"github.com/savio/savio/backend/internal/scenarios"
	"github.com/savio/savio/backend/internal/transactions"
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

func mustNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type fx struct {
	wsID  uuid.UUID
	owner uuid.UUID
	acct  uuid.UUID
	now   time.Time
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
		OpeningBalance: 10000000, Status: "ACTIVE", Version: 1}).Error)
	now, _ := time.Parse("2006-01-02", "2026-08-25")
	t.Cleanup(func() {
		db.Exec(`DELETE FROM scenario_modifications WHERE scenario_id IN (SELECT id FROM scenarios WHERE workspace_id = $1)`, wsID)
		db.Exec(`DELETE FROM scenarios WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM transactions WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM recurring_transactions WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM audit_logs WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM accounts WHERE id = $1`, acct)
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		db.Exec(`DELETE FROM users WHERE id = $1`, owner)
	})
	return fx{wsID: wsID, owner: owner, acct: acct, now: now}
}

func newRouter(t *testing.T, wsID uuid.UUID, h *scenarios.Handler) *gin.Engine {
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
	g := r.Group("/api/v1/scenarios", authMW)
	scenarios.RegisterRoutes(g, h, auth.RequireWrite())
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

func txCount(t *testing.T, wsID uuid.UUID) int64 {
	t.Helper()
	var n int64
	if err := db.Table("transactions").Where("workspace_id = ?", wsID).Count(&n).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func TestScenarioOneTimeExpenseNonDestructive(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID, scenarios.NewHandler(scenarios.NewService(db)))

	w := doReq(t, r, "POST", "/api/v1/scenarios", f.owner.String(), `{"name":"New laptop"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	scenarioID := decodeBody(t, w)["data"].(map[string]any)["id"].(string)

	// baseline snapshot first
	w = doReq(t, r, "POST", "/api/v1/scenarios/"+scenarioID+"/calculate?horizon=90", f.owner.String(), "")
	if w.Code != http.StatusOK {
		t.Fatalf("calculate baseline: %d %s", w.Code, w.Body.String())
	}
	baselineEnding := decodeBody(t, w)["data"].(map[string]any)["result"].(map[string]any)["scenario_ending_balance"].(string)

	// add one-time expense and recalculate
	mod := doReq(t, r, "POST", "/api/v1/scenarios/"+scenarioID+"/modifications", f.owner.String(),
		`{"type":"ONE_TIME_EXPENSE","amount":"10000","narrative":"MacBook"}`)
	if mod.Code != http.StatusCreated {
		t.Fatalf("mod: %d %s", mod.Code, mod.Body.String())
	}
	w = doReq(t, r, "POST", "/api/v1/scenarios/"+scenarioID+"/calculate?horizon=90", f.owner.String(), "")
	if w.Code != http.StatusOK {
		t.Fatalf("calculate: %d %s", w.Code, w.Body.String())
	}
	data := decodeBody(t, w)["data"].(map[string]any)["result"].(map[string]any)
	if delta := dec(data["scenario_ending_balance"].(string)) - dec(baselineEnding); delta != -1000000 {
		t.Fatalf("ending %v vs baseline %v (delta %d)", data["scenario_ending_balance"], baselineEnding, delta)
	}
	if data["cashflow_difference"].(string) != "-10000.00" {
		t.Fatalf("cashflow diff %v", data["cashflow_difference"])
	}
	if got := txCount(t, f.wsID); got != 0 {
		t.Fatalf("scenario must not write ledger rows, found %d", got)
	}
}

func TestScenarioMultipleModificationsAndIncome(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	svc := scenarios.NewService(db)
	sc, err := svc.Create(t.Context(), f.wsID, f.owner, &scenarios.CreateInput{Name: "Multi"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = svc.AddModification(t.Context(), f.wsID, f.owner, sc.ID, &scenarios.ModInput{Type: "ONE_TIME_INCOME", Amount: 500000})
	if err != nil {
		t.Fatalf("mod income: %v", err)
	}
	_, err = svc.AddModification(t.Context(), f.wsID, f.owner, sc.ID, &scenarios.ModInput{Type: "ONE_TIME_EXPENSE", Amount: 100000})
	if err != nil {
		t.Fatalf("mod expense: %v", err)
	}
	v, err := svc.Calculate(t.Context(), f.wsID, f.owner, sc.ID, 90, f.now)
	if err != nil {
		t.Fatalf("calculate: %v", err)
	}
	inc := dec(v.Result.ScenarioIncome) - dec(v.Result.BaselineIncome)
	exp := dec(v.Result.ScenarioExpense) - dec(v.Result.BaselineExpense)
	if inc != 500000 || exp != 100000 {
		t.Fatalf("income delta %d expense delta %d", inc, exp)
	}
	if got := dec(v.Result.CashflowDifference); got != 400000 {
		t.Fatalf("cashflow diff %d", got)
	}
	if v.Result.ModifiedEvents != 2 {
		t.Fatalf("modified events %d", v.Result.ModifiedEvents)
	}
}

func TestScenarioRecurringModification(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID, scenarios.NewHandler(scenarios.NewService(db)))
	w := doReq(t, r, "POST", "/api/v1/scenarios", f.owner.String(), `{"name":"Home gym"}`)
	scID := decodeBody(t, w)["data"].(map[string]any)["id"].(string)
	// periodic expense MONTHLY 1000 — expect 3 occurrences in 90d from 08-26
	doReq(t, r, "POST", "/api/v1/scenarios/"+scID+"/modifications", f.owner.String(),
		`{"type":"RECURRING_EXPENSE","amount":"1000","frequency":"MONTHLY","narrative":"subscription"}`)
	w = doReq(t, r, "POST", "/api/v1/scenarios/"+scID+"/calculate?horizon=90", f.owner.String(), "")
	if w.Code != http.StatusOK {
		t.Fatalf("calculate: %d %s", w.Code, w.Body.String())
	}
	res := decodeBody(t, w)["data"].(map[string]any)["result"].(map[string]any)
	if got := dec(res["cashflow_difference"].(string)); got != -300000 {
		t.Fatalf("monthly recurring over 90d should add 3 expenses of 1000 each, diff %v", res["cashflow_difference"])
	}
}

func TestScenarioStaleDetection(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	svc := scenarios.NewService(db)
	sc, err := svc.Create(t.Context(), f.wsID, f.owner, &scenarios.CreateInput{Name: "Stale check"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.Calculate(t.Context(), f.wsID, f.owner, sc.ID, 90, f.now); err != nil {
		t.Fatalf("calculate: %v", err)
	}
	// add a real future income → baseline changes
	incomeAmt := int64(1500000)
	_, err = transactions.NewService(db).Create(t.Context(), f.wsID, f.owner, &transactions.CreateInput{
		AccountID: f.acct, Type: "INCOME", AmountMinor: incomeAmt,
		TransactionDate: f.now.AddDate(0, 0, 2), Status: "POSTED",
	})
	if err != nil {
		t.Fatalf("income: %v", err)
	}
	v, err := svc.Get(t.Context(), f.wsID, sc.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !v.IsStale {
		t.Fatalf("expected scenario to be stale after finance change")
	}
}

func dec(s string) int64 {
	m, _ := money.ParseMinorUnits(s)
	return m
}