package accounts_test

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
	"github.com/savio/savio/backend/internal/transactions"
	"github.com/savio/savio/backend/internal/users"
	"github.com/savio/savio/backend/internal/workspaces"
)

func TestMain(m *testing.M) {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		dsn := v
		testURL, err := url.Parse(dsn)
		if err == nil {
			testURL.Path = "/savio_test_accounts"
			ensureTestDB(dsn, testURL.String())
			os.Setenv("DATABASE_URL", testURL.String())
		}
	}
	openDB()
	code := m.Run()
	os.Exit(code)
}

func ensureTestDB(adminDSN, testDSN string) {
	db, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return
	}
	defer db.Close()
	_, _ = db.Exec(`CREATE DATABASE savio_test_accounts`)
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

func mustNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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
		t.Fatalf("decode response: %v", err)
	}
	return out
}

func newRouter(t *testing.T, db *gorm.DB, wsID uuid.UUID) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := accounts.NewHandler(accounts.NewService(db)).WithTransactions(transactions.NewService(db))
	testAuth := func(c *gin.Context) {
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
		authctx.Set(c, &authctx.Ctx{
			UserID:          uid,
			WorkspaceID:     wsID,
			WorkspaceRole:   authctx.Role(m.Role),
			IsAuthenticated: true,
		})
		c.Next()
	}
	g := r.Group("/api/v1/accounts", testAuth)
	accounts.RegisterRoutes(g, h, auth.RequireWrite())
	return r
}

type fixtureResult struct {
	wsID  uuid.UUID
	owner uuid.UUID
	memb  uuid.UUID
	view  uuid.UUID
}

func fixture(t *testing.T, db *gorm.DB) fixtureResult {
	t.Helper()
	wsID := uuid.New()
	mustNil(t, db.Create(&workspaces.Workspace{ID: wsID, Name: "Test WS", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)
	owner := uuid.New()
	mustNil(t, db.Create(&users.User{ID: owner, Name: "Owner", Email: "o-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	memb := uuid.New()
	mustNil(t, db.Create(&workspaces.Membership{ID: memb, WorkspaceID: wsID, UserID: owner, Role: "OWNER", Status: "ACTIVE"}).Error)
	view := uuid.New()
	mustNil(t, db.Create(&users.User{ID: view, Name: "Viewer", Email: "v-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: view, Role: "VIEWER", Status: "ACTIVE"}).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM audit_logs WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM transactions WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM account_transfers WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM accounts WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		db.Exec(`DELETE FROM users WHERE id IN ($1, $2)`, owner, view)
	})
	return fixtureResult{wsID: wsID, owner: owner, memb: memb, view: view}
}

func TestAccountReconcileCreatesSignedAdjustment(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	w := doReq(t, r, "POST", "/api/v1/accounts", fx.owner.String(),
		`{"name":"Cashbox","type":"CASH","currency":"IDR","opening_balance":100000}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create account: %d %s", w.Code, w.Body.String())
	}
	id := decodeBody(t, w)["data"].(map[string]any)["id"].(string)

	// actual balance "1200.00" = 120000 minor → positive ADJUSTMENT of 20000
	w = doReq(t, r, "POST", "/api/v1/accounts/"+id+"/reconcile", fx.owner.String(),
		`{"actual_balance":"1200.00","reason":"cash count"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("reconcile: %d %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)["data"].(map[string]any)
	if body["difference"].(string) != "200.00" {
		t.Fatalf("difference = %v", body["difference"])
	}
	adj := body["adjustment"].(map[string]any)
	if adj["type"].(string) != "ADJUSTMENT" || adj["status"].(string) != "POSTED" {
		t.Fatalf("unexpected adjustment: %v", adj)
	}
	// derived balance now matches the stated actual
	w = doReq(t, r, "GET", "/api/v1/accounts/"+id, fx.owner.String(), "")
	if got := decodeBody(t, w)["data"].(map[string]any)["derived_balance"].(float64); got != 120000 {
		t.Fatalf("derived after reconcile = %v, want 120000", got)
	}
	// reconciling to the same value now conflicts
	w = doReq(t, r, "POST", "/api/v1/accounts/"+id+"/reconcile", fx.owner.String(),
		`{"actual_balance":"1200.00","reason":"again"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("no-op reconcile expected 409, got %d %s", w.Code, w.Body.String())
	}
}

func TestAccountReconcileDownward(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	w := doReq(t, r, "POST", "/api/v1/accounts", fx.owner.String(),
		`{"name":"Wallet","type":"EWALLET","currency":"IDR","opening_balance":500000}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create account: %d", w.Code)
	}
	id := decodeBody(t, w)["data"].(map[string]any)["id"].(string)
	// actual "4000.00" = 400000 minor → negative signed adjustment
	w = doReq(t, r, "POST", "/api/v1/accounts/"+id+"/reconcile", fx.owner.String(),
		`{"actual_balance":"4000.00","reason":"spent"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("reconcile: %d %s", w.Code, w.Body.String())
	}
	if got := decodeBody(t, w)["data"].(map[string]any)["difference"].(string); got != "-1000.00" {
		t.Fatalf("difference = %s", got)
	}
	w = doReq(t, r, "GET", "/api/v1/accounts/"+id, fx.owner.String(), "")
	if got := decodeBody(t, w)["data"].(map[string]any)["derived_balance"].(float64); got != 400000 {
		t.Fatalf("derived = %v, want 400000", got)
	}
}

func TestAccountCreateListGet(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)

	w := doReq(t, r, "POST", "/api/v1/accounts", fx.owner.String(),
		`{"name":"Cash Wallet","type":"EWALLET","currency":"IDR","opening_balance":5000000}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: status %d body %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	data := body["data"].(map[string]any)
	if data["derived_balance"].(float64) != 5000000 || data["status"].(string) != "ACTIVE" {
		t.Fatalf("unexpected create payload: %v", data)
	}
	id := data["id"].(string)

	w = doReq(t, r, "GET", "/api/v1/accounts/"+id, fx.owner.String(), "")
	if w.Code != http.StatusOK {
		t.Fatalf("get: status %d", w.Code)
	}
	w = doReq(t, r, "GET", "/api/v1/accounts", fx.owner.String(), "")
	if w.Code != http.StatusOK {
		t.Fatalf("list: status %d", w.Code)
	}
	list := decodeBody(t, w)
	if list["meta"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("expected 1 account, got %v", list["meta"])
	}
}

func TestAccountCurrencyRuleRejectsNonBase(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	w := doReq(t, r, "POST", "/api/v1/accounts", fx.owner.String(),
		`{"name":"USD Acct","type":"BANK","currency":"USD","opening_balance":0}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for foreign currency, got %d %s", w.Code, w.Body.String())
	}
}

func TestAccountValidation(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	for _, body := range []string{
		`{"name":"","type":"BANK","currency":"IDR","opening_balance":0}`,
		`{"name":"x","type":"STOCKS","currency":"IDR","opening_balance":0}`,
		`{"name":"x","type":"BANK","currency":"IDR","opening_balance":-5}`,
	} {
		w := doReq(t, r, "POST", "/api/v1/accounts", fx.owner.String(), body)
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("body %s expected 422, got %d", body, w.Code)
		}
	}
}

func TestAccountDuplicateNameConflict(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	doReq(t, r, "POST", "/api/v1/accounts", fx.owner.String(),
		`{"name":"Wallet","type":"CASH","currency":"IDR","opening_balance":0}`)
	w := doReq(t, r, "POST", "/api/v1/accounts", fx.owner.String(),
		`{"name":"Wallet","type":"CASH","currency":"IDR","opening_balance":0}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 duplicate, got %d %s", w.Code, w.Body.String())
	}
}

func TestAccountUpdateVersionConflict(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	w := doReq(t, r, "POST", "/api/v1/accounts", fx.owner.String(),
		`{"name":"Wallet","type":"CASH","currency":"IDR","opening_balance":0}`)
	id := decodeBody(t, w)["data"].(map[string]any)["id"].(string)
	// stale version 1 after update bumped to 2
	doReq(t, r, "PATCH", "/api/v1/accounts/"+id, fx.owner.String(),
		`{"name":"Renamed","type":"CASH","institution_name":"BCA","description":"","version":1}`)
	w = doReq(t, r, "PATCH", "/api/v1/accounts/"+id, fx.owner.String(),
		`{"name":"Stale","type":"CASH","institution_name":"","description":"","version":1}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 version conflict, got %d %s", w.Code, w.Body.String())
	}
}

func TestAccountArchiveRestore(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	w := doReq(t, r, "POST", "/api/v1/accounts", fx.owner.String(),
		`{"name":"Wallet","type":"CASH","currency":"IDR","opening_balance":0}`)
	id := decodeBody(t, w)["data"].(map[string]any)["id"].(string)

	w = doReq(t, r, "POST", "/api/v1/accounts/"+id+"/archive", fx.owner.String(), "")
	if w.Code != http.StatusOK || decodeBody(t, w)["data"].(map[string]any)["status"].(string) != "ARCHIVED" {
		t.Fatalf("archive failed: %d %s", w.Code, w.Body.String())
	}
	w = doReq(t, r, "POST", "/api/v1/accounts/"+id+"/restore", fx.owner.String(), "")
	if w.Code != http.StatusOK || decodeBody(t, w)["data"].(map[string]any)["status"].(string) != "ACTIVE" {
		t.Fatalf("restore failed: %d %s", w.Code, w.Body.String())
	}
}

func TestAccountViewerCannotMutate(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	for _, body := range []struct{ method, path, b string }{
		{"POST", "/api/v1/accounts", `{"name":"x","type":"CASH","currency":"IDR","opening_balance":0}`},
		{"DELETE", "/api/v1/accounts/" + uuid.NewString(), ""},
	} {
		w := doReq(t, r, body.method, body.path, fx.view.String(), body.b)
		if w.Code != http.StatusForbidden {
			t.Fatalf("VIEWER mutation expected 403, got %d", w.Code)
		}
	}
}

func TestAccountCrossWorkspaceIsolation(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	fx2 := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	w := doReq(t, r, "POST", "/api/v1/accounts", fx.owner.String(),
		`{"name":"OnlyHere","type":"CASH","currency":"IDR","opening_balance":0}`)
	id := decodeBody(t, w)["data"].(map[string]any)["id"].(string)

	// foreign account requested through fx2's own workspace router must be
	// invisible (workspace-scoped lookup).
	r2 := newRouter(t, db, fx2.wsID)
	w = doReq(t, r2, "GET", "/api/v1/accounts/"+id, fx2.owner.String(), "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("foreign account must be invisible, got %d %s", w.Code, w.Body.String())
	}
}

func TestAccountDeleteWhenNoLedger(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	w := doReq(t, r, "POST", "/api/v1/accounts", fx.owner.String(),
		`{"name":"Temp","type":"CASH","currency":"IDR","opening_balance":0}`)
	id := decodeBody(t, w)["data"].(map[string]any)["id"].(string)
	w = doReq(t, r, "DELETE", "/api/v1/accounts/"+id, fx.owner.String(), "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected delete ok, got %d %s", w.Code, w.Body.String())
	}
	w = doReq(t, r, "GET", "/api/v1/accounts/"+id, fx.owner.String(), "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("deleted account should be gone, got %d", w.Code)
	}
}
