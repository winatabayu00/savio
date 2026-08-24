package transfers_test

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
	db, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return
	}
	defer db.Close()
	_, _ = db.Exec(`CREATE DATABASE savio_test`)
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

type fx struct {
	wsID  uuid.UUID
	owner uuid.UUID
	from  uuid.UUID
	to    uuid.UUID
}

func fixture(t *testing.T) fx {
	t.Helper()
	wsID := uuid.New()
	mustNil(t, db.Create(&workspaces.Workspace{ID: wsID, Name: "T WS", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)
	owner := uuid.New()
	mustNil(t, db.Create(&users.User{ID: owner, Name: "O", Email: "o-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: owner, Role: "OWNER", Status: "ACTIVE"}).Error)

	fromID := uuid.New()
	toID := uuid.New()
	mustNil(t, db.Create(&accounts.Account{ID: fromID, WorkspaceID: wsID, Name: "Cash", Type: "CASH",
		Currency: "IDR", OpeningBalance: 10000000, Status: "ACTIVE", Version: 1}).Error)
	mustNil(t, db.Create(&accounts.Account{ID: toID, WorkspaceID: wsID, Name: "Bank", Type: "BANK",
		Currency: "IDR", OpeningBalance: 20000000, Status: "ACTIVE", Version: 1}).Error)

	t.Cleanup(func() {
		db.Exec(`DELETE FROM account_transfers WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM audit_logs WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM accounts WHERE id IN ($1, $2)`, fromID, toID)
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		db.Exec(`DELETE FROM users WHERE id = $1`, owner)
	})
	return fx{wsID: wsID, owner: owner, from: fromID, to: toID}
}

func newRouter(t *testing.T, wsID uuid.UUID) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := transfers.NewHandler(transfers.NewService(db))
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
	g := r.Group("/api/v1/transfers", authMW)
	transfers.RegisterRoutes(g, h, auth.RequireWrite())
	return r
}

func derived(t *testing.T, wsID, acctID uuid.UUID) int64 {
	t.Helper()
	v, err := accounts.NewService(db).Get(t.Context(), wsID, acctID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	return v.DerivedBalance
}

func TestTransferMovesBalancesAndKeepsTotalFixed(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	fromBefore := derived(t, f.wsID, f.from)
	toBefore := derived(t, f.wsID, f.to)

	w := doReq(t, r, "POST", "/api/v1/transfers", f.owner.String(),
		`{"from_account_id":"`+f.from.String()+`","to_account_id":"`+f.to.String()+`","amount":"500000.00","transfer_date":"2026-08-10","description":"saving"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	// from -50000000, to +50000000 (minor)
	if got := derived(t, f.wsID, f.from); got != fromBefore-50000000 {
		t.Fatalf("from balance = %d, want %d", got, fromBefore-50000000)
	}
	if got := derived(t, f.wsID, f.to); got != toBefore+50000000 {
		t.Fatalf("to balance = %d, want %d", got, toBefore+50000000)
	}
	// total portfolio unchanged (INV-006)
	if got := derived(t, f.wsID, f.from) + derived(t, f.wsID, f.to); got != fromBefore+toBefore {
		t.Fatalf("portfolio changed: %d vs %d", got, fromBefore+toBefore)
	}
}

func TestTransferToSameAccountRejected(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	w := doReq(t, r, "POST", "/api/v1/transfers", f.owner.String(),
		`{"from_account_id":"`+f.from.String()+`","to_account_id":"`+f.from.String()+`","amount":"1000","transfer_date":"2026-08-10"}`)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("same-account transfer expected 422, got %d %s", w.Code, w.Body.String())
	}
}

func TestTransferAmountValidation(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	w := doReq(t, r, "POST", "/api/v1/transfers", f.owner.String(),
		`{"from_account_id":"`+f.from.String()+`","to_account_id":"`+f.to.String()+`","amount":"-5","transfer_date":"2026-08-10"}`)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("negative amount expected 422, got %d", w.Code)
	}
}

func TestTransferArchivedAccountRejected(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	mustNil(t, db.Model(&accounts.Account{}).Where("id = ?", f.to).Update("status", "ARCHIVED").Error)
	w := doReq(t, r, "POST", "/api/v1/transfers", f.owner.String(),
		`{"from_account_id":"`+f.from.String()+`","to_account_id":"`+f.to.String()+`","amount":"1000","transfer_date":"2026-08-10"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("archived destination expected 409, got %d %s", w.Code, w.Body.String())
	}
}

func TestTransferForeignAccountRejected(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	other := uuid.New()
	w := doReq(t, r, "POST", "/api/v1/transfers", f.owner.String(),
		`{"from_account_id":"`+f.from.String()+`","to_account_id":"`+other.String()+`","amount":"1000","transfer_date":"2026-08-10"}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("foreign to-account expected 404, got %d %s", w.Code, w.Body.String())
	}
}

func TestTransferVoidReversesAndRejectsDoubleVoid(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	fromBefore := derived(t, f.wsID, f.from)
	toBefore := derived(t, f.wsID, f.to)

	w := doReq(t, r, "POST", "/api/v1/transfers", f.owner.String(),
		`{"from_account_id":"`+f.from.String()+`","to_account_id":"`+f.to.String()+`","amount":"10000.00","transfer_date":"2026-08-10"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	tx := decodeBody(t, w)["data"].(map[string]any)
	id := tx["id"].(string)
	version := int64(tx["version"].(float64))

	vBody, _ := json.Marshal(map[string]any{"version": version, "reason": "oops"})
	w = doReq(t, r, "POST", "/api/v1/transfers/"+id+"/void", f.owner.String(), string(vBody))
	if w.Code != http.StatusOK {
		t.Fatalf("void: %d %s", w.Code, w.Body.String())
	}
	if got := derived(t, f.wsID, f.from); got != fromBefore {
		t.Fatalf("from balance after void = %d, want %d", got, fromBefore)
	}
	if got := derived(t, f.wsID, f.to); got != toBefore {
		t.Fatalf("to balance after void = %d, want %d", got, toBefore)
	}
	vBody2, _ := json.Marshal(map[string]any{"version": version + 1})
	w = doReq(t, r, "POST", "/api/v1/transfers/"+id+"/void", f.owner.String(), string(vBody2))
	if w.Code != http.StatusConflict {
		t.Fatalf("double void expected 409, got %d", w.Code)
	}
}

func TestTransferListViewWorksOnlyInWorkspace(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	f2 := fixture(t)
	r := newRouter(t, f.wsID)
	doReq(t, r, "POST", "/api/v1/transfers", f.owner.String(),
		`{"from_account_id":"`+f.from.String()+`","to_account_id":"`+f.to.String()+`","amount":"1000","transfer_date":"2026-08-10"}`)
	r2 := newRouter(t, f2.wsID)
	doReq(t, r2, "POST", "/api/v1/transfers", f2.owner.String(),
		`{"from_account_id":"`+f2.from.String()+`","to_account_id":"`+f2.to.String()+`","amount":"999","transfer_date":"2026-08-11"}`)
	w := doReq(t, r2, "GET", "/api/v1/transfers", f2.owner.String(), "")
	list := decodeBody(t, w)["data"].([]any)
	if len(list) != 1 {
		t.Fatalf("expected 1 transfer in f2 workspace, got %d", len(list))
	}
	if list[0].(map[string]any)["amount"].(string) != "999.00" {
		t.Fatalf("unexpected amount: %v", list)
	}
}
