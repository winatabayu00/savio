package recurring_test

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
	"github.com/savio/savio/backend/internal/recurring"
	"github.com/savio/savio/backend/internal/seeds"
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
	wsID   uuid.UUID
	owner  uuid.UUID
	acctID uuid.UUID
	cat    uuid.UUID
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
	acctID := uuid.New()
	mustNil(t, db.Create(&accounts.Account{ID: acctID, WorkspaceID: wsID, Name: "Main", Type: "BANK",
		Currency: "IDR", OpeningBalance: 10000000, Status: "ACTIVE", Version: 1}).Error)
	var salary struct{ ID uuid.UUID }
	mustNil(t, db.Table("categories").Where("name = ? AND type = 'INCOME'", "Salary").First(&salary).Error)

	t.Cleanup(func() {
		db.Exec(`DELETE FROM transactions WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM recurring_occurrences WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM recurring_transactions WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM audit_logs WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM accounts WHERE id = $1`, acctID)
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		db.Exec(`DELETE FROM users WHERE id = $1`, owner)
	})
	return fx{wsID: wsID, owner: owner, acctID: acctID, cat: salary.ID}
}

func newRouter(t *testing.T, wsID uuid.UUID) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := recurring.NewHandler(recurring.NewService(db))
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
	g := r.Group("/api/v1/recurring-transactions", authMW)
	occs := r.Group("/api/v1/recurring-occurrences", authMW)
	recurring.RegisterRoutes(g, h, occs, auth.RequireWrite())
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

func TestRecurringGenerateOccurrencesAndList(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	w := doReq(t, r, "POST", "/api/v1/recurring-transactions", f.owner.String(),
		`{"account_id":"`+f.acctID.String()+`","category_id":"`+f.cat.String()+`","type":"INCOME","amount":"3000000.00","frequency":"MONTH_END","start_date":"2026-07-15","description":"monthly salary"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	ruleID := decodeBody(t, w)["data"].(map[string]any)["id"].(string)

	w = doReq(t, r, "GET", "/api/v1/recurring-transactions/"+ruleID+"/occurrences", f.owner.String(), "")
	if w.Code != http.StatusOK {
		t.Fatalf("occurrences: %d %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	list := body["data"].([]any)
	if len(list) < 3 {
		t.Fatalf("expected 3+ month-end occurrences, got %d", len(list))
	}
	if list[0].(map[string]any)["due_date"].(string) != "2026-07-31" {
		t.Fatalf("first month-end occurrence should be 2026-07-31, got %v", list[0])
	}
	// draft rule generation must not move the balance
	if got := derived(t, f.wsID, f.acctID); got != 10000000 {
		t.Fatalf("balance must be untouched by generation: %d", got)
	}
}

func TestOccurrenceConfirmPostsTransactionIdempotently(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	w := doReq(t, r, "POST", "/api/v1/recurring-transactions", f.owner.String(),
		`{"account_id":"`+f.acctID.String()+`","type":"INCOME","amount":"750000.00","frequency":"MONTHLY","start_date":"2026-08-01","description":"rent income"}`)
	ruleID := decodeBody(t, w)["data"].(map[string]any)["id"].(string)
	w = doReq(t, r, "GET", "/api/v1/recurring-transactions/"+ruleID+"/occurrences?from=2026-08-01&to=2026-08-31", f.owner.String(), "")
	occ := decodeBody(t, w)["data"].([]any)[0].(map[string]any)
	occID := occ["id"].(string)
	version := int64(occ["version"].(float64))

	before := derived(t, f.wsID, f.acctID)
	confirmBody, _ := json.Marshal(map[string]any{"version": version})
	w = doReq(t, r, "POST", "/api/v1/recurring-occurrences/"+occID+"/confirm", f.owner.String(), string(confirmBody))
	if w.Code != http.StatusOK {
		t.Fatalf("confirm: %d %s", w.Code, w.Body.String())
	}
	confirmed := decodeBody(t, w)["data"].(map[string]any)
	if confirmed["status"].(string) != "CONFIRMED" {
		t.Fatalf("status %v", confirmed["status"])
	}
	if confirmed["posted_transaction_id"] == nil {
		t.Fatalf("missing posted transaction link")
	}
	// balance moved by the posted transaction
	want := before + 75000000
	if got := derived(t, f.wsID, f.acctID); got != want {
		t.Fatalf("balance after confirm = %d, want %d", got, want)
	}
	// confirming again must be rejected
	w = doReq(t, r, "POST", "/api/v1/recurring-occurrences/"+occID+"/confirm", f.owner.String(), string(confirmBody))
	if w.Code != http.StatusConflict {
		t.Fatalf("double confirm expected 409, got %d %s", w.Code, w.Body.String())
	}
}

func TestOccurrenceSkipCreatesNoLedger(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	w := doReq(t, r, "POST", "/api/v1/recurring-transactions", f.owner.String(),
		`{"account_id":"`+f.acctID.String()+`","type":"EXPENSE","amount":"100000.00","frequency":"WEEKLY","start_date":"2026-08-03","end_date":"2026-08-31","description":"weekly friday"}`)
	ruleID := decodeBody(t, w)["data"].(map[string]any)["id"].(string)
	w = doReq(t, r, "GET", "/api/v1/recurring-transactions/"+ruleID+"/occurrences?from=2026-08-03&to=2026-08-09", f.owner.String(), "")
	occ := decodeBody(t, w)["data"].([]any)[0].(map[string]any)
	occID := occ["id"].(string)
	version := int64(occ["version"].(float64))

	before := derived(t, f.wsID, f.acctID)
	skipBody, _ := json.Marshal(map[string]any{"version": version})
	w = doReq(t, r, "POST", "/api/v1/recurring-occurrences/"+occID+"/skip", f.owner.String(), string(skipBody))
	if w.Code != http.StatusOK || decodeBody(t, w)["data"].(map[string]any)["status"].(string) != "SKIPPED" {
		t.Fatalf("skip: %d %s", w.Code, w.Body.String())
	}
	if got := derived(t, f.wsID, f.acctID); got != before {
		t.Fatalf("skip must not move balance: %d vs %d", got, before)
	}
}

func TestRecurringLifecyclePauseResumeEnd(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	w := doReq(t, r, "POST", "/api/v1/recurring-transactions", f.owner.String(),
		`{"account_id":"`+f.acctID.String()+`","type":"EXPENSE","amount":"500000.00","frequency":"MONTHLY","start_date":"2026-08-01"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	ruleID := decodeBody(t, w)["data"].(map[string]any)["id"].(string)

	// PATCH bumps version 1→2; capture the version for status transitions
	w = doReq(t, r, "PATCH", "/api/v1/recurring-transactions/"+ruleID, f.owner.String(),
		`{"account_id":"`+f.acctID.String()+`","type":"EXPENSE","amount":"550000.00","frequency":"MONTHLY","start_date":"2026-08-01","version":1}`)
	if w.Code != http.StatusOK {
		t.Fatalf("patch: %d %s", w.Code, w.Body.String())
	}
	rule := decodeBody(t, w)["data"].(map[string]any)
	if rule["amount"].(string) != "550000.00" {
		t.Fatalf("patched amount %v", rule["amount"])
	}
	version := int64(rule["version"].(float64))

	pause, _ := json.Marshal(map[string]any{"version": version})
	w = doReq(t, r, "POST", "/api/v1/recurring-transactions/"+ruleID+"/pause", f.owner.String(), string(pause))
	if w.Code != http.StatusOK || decodeBody(t, w)["data"].(map[string]any)["status"].(string) != "PAUSED" {
		t.Fatalf("pause: %d %s", w.Code, w.Body.String())
	}
	version = int64(decodeBody(t, w)["data"].(map[string]any)["version"].(float64))
	resumeBody, _ := json.Marshal(map[string]any{"version": version})
	w = doReq(t, r, "POST", "/api/v1/recurring-transactions/"+ruleID+"/resume", f.owner.String(), string(resumeBody))
	if w.Code != http.StatusOK || decodeBody(t, w)["data"].(map[string]any)["status"].(string) != "ACTIVE" {
		t.Fatalf("resume: %d %s", w.Code, w.Body.String())
	}
	version = int64(decodeBody(t, w)["data"].(map[string]any)["version"].(float64))
	endBody, _ := json.Marshal(map[string]any{"version": version})
	w = doReq(t, r, "POST", "/api/v1/recurring-transactions/"+ruleID+"/end", f.owner.String(), string(endBody))
	if w.Code != http.StatusOK || decodeBody(t, w)["data"].(map[string]any)["status"].(string) != "ENDED" {
		t.Fatalf("end: %d %s", w.Code, w.Body.String())
	}
	// won't resume after end
	version = int64(decodeBody(t, w)["data"].(map[string]any)["version"].(float64))
	again, _ := json.Marshal(map[string]any{"version": version})
	w = doReq(t, r, "POST", "/api/v1/recurring-transactions/"+ruleID+"/resume", f.owner.String(), string(again))
	if w.Code != http.StatusConflict {
		t.Fatalf("resume after end expected 409, got %d", w.Code)
	}
}

func TestRecurringConfirmVersionConflict(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	w := doReq(t, r, "POST", "/api/v1/recurring-transactions", f.owner.String(),
		`{"account_id":"`+f.acctID.String()+`","type":"EXPENSE","amount":"100000.00","frequency":"MONTHLY","start_date":"2026-08-01"}`)
	ruleID := decodeBody(t, w)["data"].(map[string]any)["id"].(string)
	w = doReq(t, r, "GET", "/api/v1/recurring-transactions/"+ruleID+"/occurrences", f.owner.String(), "")
	occID := decodeBody(t, w)["data"].([]any)[0].(map[string]any)["id"].(string)
	badVersion, _ := json.Marshal(map[string]any{"version": 999})
	w = doReq(t, r, "POST", "/api/v1/recurring-occurrences/"+occID+"/confirm", f.owner.String(), string(badVersion))
	if w.Code != http.StatusConflict {
		t.Fatalf("version conflict expected 409, got %d %s", w.Code, w.Body.String())
	}
}
func TestConcurrentOccurrenceConfirmPostsOnce(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	f := fixture(t)
	r := newRouter(t, f.wsID)
	w := doReq(t, r, "POST", "/api/v1/recurring-transactions", f.owner.String(),
		`{"account_id":"`+f.acctID.String()+`","type":"EXPENSE","amount":"50000.00","frequency":"MONTHLY","start_date":"2026-08-01"}`)
	ruleID := decodeBody(t, w)["data"].(map[string]any)["id"].(string)
	w = doReq(t, r, "GET", "/api/v1/recurring-transactions/"+ruleID+"/occurrences", f.owner.String(), "")
	occ := decodeBody(t, w)["data"].([]any)[0].(map[string]any)
	occID := occ["id"].(string)
	version := int64(occ["version"].(float64))

	svc := recurring.NewService(db)
	const workers = 8
	results := make(chan bool, workers)
	for i := 0; i < workers; i++ {
		go func() {
			_, err := svc.Confirm(t.Context(), f.wsID, f.owner, mustParse(t, occID), version)
			results <- err == nil
		}()
	}
	successes := 0
	for i := 0; i < workers; i++ {
		if <-results {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly 1 successful confirm, got %d", successes)
	}
	var txCount int64
	if err := db.Table("transactions").Where("workspace_id = ? AND source = 'RECURRING'", f.wsID).Count(&txCount).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if txCount != 1 {
		t.Fatalf("expected exactly 1 posted transaction, got %d", txCount)
	}
	var confirmed int64
	if err := db.Table("recurring_occurrences").Where("id = ? AND status = 'CONFIRMED'", mustParse(t, occID)).Count(&confirmed).Error; err != nil {
		t.Fatalf("confirmed count: %v", err)
	}
	if confirmed != 1 {
		t.Fatalf("occurrence must be CONFIRMED exactly once")
	}
}

func mustParse(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse uuid: %v", err)
	}
	return id
}
