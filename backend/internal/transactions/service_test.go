package transactions_test

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
	"github.com/savio/savio/backend/internal/seeds"
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


func init() {
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

type fx struct {
	wsID      uuid.UUID
	owner     uuid.UUID
	acctID    uuid.UUID
	category  uuid.UUID
	expCat    uuid.UUID
	accRouter *gin.Engine
}

func fixture(t *testing.T) fx {
	t.Helper()
	mustNil(t, seeds.SeedSystemCategories(t.Context(), db))
	wsID := uuid.New()
	mustNil(t, db.Create(&workspaces.Workspace{ID: wsID, Name: "T WS", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)
	owner := uuid.New()
	mustNil(t, db.Create(&users.User{ID: owner, Name: "O", Email: "o-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: owner, Role: "OWNER", Status: "ACTIVE"}).Error)

	acctID := uuid.New()
	mustNil(t, db.Create(&accounts.Account{ID: acctID, WorkspaceID: wsID, Name: "Main", Type: "BANK",
		Currency: "IDR", OpeningBalance: 100000, Status: "ACTIVE", Version: 1}).Error)

	var salary, food struct {
		ID uuid.UUID
	}
	mustNil(t, db.Table("categories").Where("name = ? AND type = 'INCOME'", "Salary").First(&salary).Error)
	mustNil(t, db.Table("categories").Where("name = ? AND type = 'EXPENSE'", "Food & Dining").First(&food).Error)

	t.Cleanup(func() {
		db.Exec(`DELETE FROM transactions WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM audit_logs WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM accounts WHERE id = $1`, acctID)
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		db.Exec(`DELETE FROM users WHERE id = $1`, owner)
	})

	return fx{wsID: wsID, owner: owner, acctID: acctID, category: salary.ID, expCat: food.ID}
}

func newRouter(t *testing.T, wsID uuid.UUID) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := transactions.NewHandler(transactions.NewService(db))
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
	g := r.Group("/api/v1/transactions", testAuth)
	transactions.RegisterRoutes(g, h, auth.RequireWrite())
	return r
}

func accountBalance(t *testing.T, wsID, acctID uuid.UUID) int64 {
	t.Helper()
	v, err := accounts.NewService(db).Get(t.Context(), wsID, acctID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	return v.DerivedBalance
}

func TestPostedTransactionMovesBalance(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)

	w := doReq(t, r, "POST", "/api/v1/transactions", f.owner.String(),
		`{"account_id":"`+f.acctID.String()+`","category_id":"`+f.category.String()+`","type":"INCOME","amount":"1500000.00","transaction_date":"2026-08-01","status":"POSTED","description":"gaji"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create income: %d %s", w.Code, w.Body.String())
	}
	w = doReq(t, r, "POST", "/api/v1/transactions", f.owner.String(),
		`{"account_id":"`+f.acctID.String()+`","category_id":"`+f.expCat.String()+`","type":"EXPENSE","amount":"25000.00","transaction_date":"2026-08-02","status":"POSTED","description":"makan"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create expense: %d %s", w.Code, w.Body.String())
	}

	// opening 100000 + 150000000 - 2500000 (minor units)
	want := int64(100000) + 150000000 - 2500000
	if got := accountBalance(t, f.wsID, f.acctID); got != want {
		t.Fatalf("derived balance = %d, want %d", got, want)
	}
}

func TestCategoryMismatchRejected(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	w := doReq(t, r, "POST", "/api/v1/transactions", f.owner.String(),
		`{"account_id":"`+f.acctID.String()+`","category_id":"`+f.expCat.String()+`","type":"INCOME","amount":"1000","transaction_date":"2026-08-01"}`)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("category mismatch expected 422, got %d %s", w.Code, w.Body.String())
	}
}

func TestForeignAccountRejected(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	other := uuid.New()
	w := doReq(t, r, "POST", "/api/v1/transactions", f.owner.String(),
		`{"account_id":"`+other.String()+`","type":"INCOME","amount":"1000","transaction_date":"2026-08-01","status":"POSTED"}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("foreign account expected 404, got %d %s", w.Code, w.Body.String())
	}
}

func TestArchivedAccountRejected(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	mustNil(t, db.Model(&accounts.Account{}).Where("id = ?", f.acctID).Update("status", "ARCHIVED").Error)
	w := doReq(t, r, "POST", "/api/v1/transactions", f.owner.String(),
		`{"account_id":"`+f.acctID.String()+`","type":"INCOME","amount":"1000","transaction_date":"2026-08-01","status":"POSTED"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("archived account expected 409, got %d %s", w.Code, w.Body.String())
	}
}

func TestDraftEditAndPost(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	before := accountBalance(t, f.wsID, f.acctID)

	w := doReq(t, r, "POST", "/api/v1/transactions", f.owner.String(),
		`{"account_id":"`+f.acctID.String()+`","type":"EXPENSE","amount":"5000","transaction_date":"2026-08-01"}`)
	id := decodeBody(t, w)["data"].(map[string]any)["id"].(string)
	if w.Code != http.StatusCreated {
		t.Fatalf("create draft: %d", w.Code)
	}
	if got := accountBalance(t, f.wsID, f.acctID); got != before {
		t.Fatalf("draft must not move balance: %d vs %d", got, before)
	}

	// edit draft amount
	bodyV, _ := json.Marshal(map[string]any{"version": 1, "type": "EXPENSE", "amount": "7000", "transaction_date": "2026-08-01"})
	w = doReq(t, r, "PATCH", "/api/v1/transactions/"+id, f.owner.String(), string(bodyV))
	if w.Code != http.StatusOK {
		t.Fatalf("edit draft: %d %s", w.Code, w.Body.String())
	}

	// post it
	postV, _ := json.Marshal(map[string]any{"version": 2})
	w = doReq(t, r, "POST", "/api/v1/transactions/"+id+"/post", f.owner.String(), string(postV))
	if w.Code != http.StatusOK {
		t.Fatalf("post: %d %s", w.Code, w.Body.String())
	}
	if got := accountBalance(t, f.wsID, f.acctID); got != before-700000 {
		t.Fatalf("balance after post = %d, want %d", got, before-700000)
	}

	// posted transaction immutable
	bodyV2, _ := json.Marshal(map[string]any{"version": 3, "type": "EXPENSE", "amount": "9999", "transaction_date": "2026-08-01"})
	w = doReq(t, r, "PATCH", "/api/v1/transactions/"+id, f.owner.String(), string(bodyV2))
	if w.Code != http.StatusConflict {
		t.Fatalf("posted edit expected conflict, got %d %s", w.Code, w.Body.String())
	}
	w = doReq(t, r, "GET", "/api/v1/transactions/"+id, f.owner.String(), "")
	status := decodeBody(t, w)["data"].(map[string]any)["status"].(string)
	if status != "POSTED" {
		t.Fatalf("expected POSTED, got %s", status)
	}
}

func TestVoidReversesBalanceAndRejectsDoubleVoid(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	before := accountBalance(t, f.wsID, f.acctID)
	w := doReq(t, r, "POST", "/api/v1/transactions", f.owner.String(),
		`{"account_id":"`+f.acctID.String()+`","type":"EXPENSE","amount":"8000","transaction_date":"2026-08-01","status":"POSTED"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	tx := decodeBody(t, w)["data"].(map[string]any)
	id := tx["id"].(string)
	version := int64(tx["version"].(float64))

	if got := accountBalance(t, f.wsID, f.acctID); got != before-800000 {
		t.Fatalf("balance after expense = %d", got)
	}

	vBody, _ := json.Marshal(map[string]any{"version": version, "reason": "duplicate entry"})
	w = doReq(t, r, "POST", "/api/v1/transactions/"+id+"/void", f.owner.String(), string(vBody))
	if w.Code != http.StatusOK {
		t.Fatalf("void: %d %s", w.Code, w.Body.String())
	}
	if got := accountBalance(t, f.wsID, f.acctID); got != before {
		t.Fatalf("balance after void = %d, want %d", got, before)
	}

	// double void rejects
	vBody2, _ := json.Marshal(map[string]any{"version": 3})
	w = doReq(t, r, "POST", "/api/v1/transactions/"+id+"/void", f.owner.String(), string(vBody2))
	if w.Code != http.StatusConflict {
		t.Fatalf("double void expected conflict, got %d %s", w.Code, w.Body.String())
	}
}

func TestListSearchTypeFiltersAndPagination(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	create := func(typ, desc, amt, day string) {
		w := doReq(t, r, "POST", "/api/v1/transactions", f.owner.String(),
			`{"account_id":"`+f.acctID.String()+`","type":"`+typ+`","amount":"`+amt+`","transaction_date":"2026-08-0`+day+`","status":"POSTED","description":"`+desc+`"}`)
		if w.Code != http.StatusCreated {
			t.Fatalf("seed %s: %d", desc, w.Code)
		}
	}
	create("INCOME", "Marketplace payout", "500000", "1")
	create("EXPENSE", "Coffee run", "20000", "2")
	create("EXPENSE", "Lunch meeting", "75000", "3")

	w := doReq(t, r, "GET", "/api/v1/transactions?search=coffee", f.owner.String(), "")
	list := decodeBody(t, w)["data"].([]any)
	if len(list) != 1 {
		t.Fatalf("search coffee expected 1, got %d", len(list))
	}
	w = doReq(t, r, "GET", "/api/v1/transactions?type=EXPENSE", f.owner.String(), "")
	if len(decodeBody(t, w)["data"].([]any)) != 2 {
		t.Fatalf("type filter failed")
	}
	w = doReq(t, r, "GET", "/api/v1/transactions?status=POSTED&limit=2&page=1", f.owner.String(), "")
	body := decodeBody(t, w)
	if len(body["data"].([]any)) != 2 || body["meta"].(map[string]any)["total"].(float64) != 3 {
		t.Fatalf("pagination failed: %v", body["meta"])
	}
}

func TestWorkspaceIsolation(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	f2 := fixture(t)
	r2 := newRouter(t, f2.wsID)
	w := doReq(t, r2, "POST", "/api/v1/transactions", f2.owner.String(),
		`{"account_id":"`+f2.acctID.String()+`","type":"INCOME","amount":"1111","transaction_date":"2026-08-01","status":"POSTED","description":"arbitrary"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("seed other ws: %d", w.Code)
	}
	otherID := decodeBody(t, w)["data"].(map[string]any)["id"].(string)

	r1 := newRouter(t, f.wsID)
	w = doReq(t, r1, "GET", "/api/v1/transactions/"+otherID, f.owner.String(), "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("cross-workspace tx visible: %d %s", w.Code, w.Body.String())
	}
	w = doReq(t, r1, "GET", "/api/v1/transactions", f.owner.String(), "")
	if len(decodeBody(t, w)["data"].([]any)) != 0 {
		t.Fatalf("cross-workspace tx listed: %v", w.Body.String())
	}
	// fx2 account reflects only its own transaction (opening 100000 + 111100)
	if got := accountBalance(t, f2.wsID, f2.acctID); got != 100000+111100 {
		t.Fatalf("fx2 balance polluted: %d", got)
	}
}

func TestViewerCannotMutateTransactions(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	viewUser := uuid.New()
	mustNil(t, db.Create(&users.User{ID: viewUser, Name: "V", Email: "v-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: f.wsID, UserID: viewUser, Role: "VIEWER", Status: "ACTIVE"}).Error)
	r := newRouter(t, f.wsID)
	w := doReq(t, r, "POST", "/api/v1/transactions", viewUser.String(),
		`{"account_id":"`+f.acctID.String()+`","type":"INCOME","amount":"1","transaction_date":"2026-08-01"}`)
	if w.Code != http.StatusForbidden {
		t.Fatalf("VIEWER create expected 403, got %d", w.Code)
	}
}