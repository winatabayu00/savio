package categories_test

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

	"github.com/savio/savio/backend/internal/auth"
	"github.com/savio/savio/backend/internal/categories"
	"github.com/savio/savio/backend/internal/migrations"
	"github.com/savio/savio/backend/internal/platform/authctx"
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
	os.Exit(m.Run())
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

func newRouter(t *testing.T, db *gorm.DB, wsID uuid.UUID) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := categories.NewHandler(categories.NewService(db))
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
	g := r.Group("/api/v1/categories", testAuth)
	categories.RegisterRoutes(g, h, auth.RequireWrite())
	return r
}

type fixtureResult struct {
	wsID  uuid.UUID
	owner uuid.UUID
	view  uuid.UUID
}

func fixture(t *testing.T, db *gorm.DB) fixtureResult {
	t.Helper()
	mustNil(t, seeds.SeedSystemCategories(t.Context(), db))
	wsID := uuid.New()
	mustNil(t, db.Create(&workspaces.Workspace{ID: wsID, Name: "Test WS", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)
	owner := uuid.New()
	mustNil(t, db.Create(&users.User{ID: owner, Name: "Owner", Email: "o-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: owner, Role: "OWNER", Status: "ACTIVE"}).Error)
	view := uuid.New()
	mustNil(t, db.Create(&users.User{ID: view, Name: "Viewer", Email: "v-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: view, Role: "VIEWER", Status: "ACTIVE"}).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM categories WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		db.Exec(`DELETE FROM users WHERE id IN ($1, $2)`, owner, view)
	})
	return fixtureResult{wsID: wsID, owner: owner, view: view}
}

func TestCategoryListIncludesSystemAndCustom(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)

	w := doReq(t, r, "POST", "/api/v1/categories", fx.owner.String(), `{"name":"Groceries","type":"EXPENSE"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}

	w = doReq(t, r, "GET", "/api/v1/categories", fx.owner.String(), "")
	if w.Code != http.StatusOK {
		t.Fatalf("list: %d", w.Code)
	}
	list := decodeBody(t, w)["data"].([]any)
	customFound, systemFound := false, false
	for _, item := range list {
		cat := item.(map[string]any)
		if cat["name"].(string) == "Groceries" && cat["is_system"].(bool) == false {
			customFound = true
		}
		if cat["name"].(string) == "Salary" && cat["is_system"].(bool) == true {
			systemFound = true
		}
	}
	if !customFound || !systemFound {
		t.Fatalf("expected system + custom categories in list %v", list)
	}
}

func TestCategoryTypeFilter(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	w := doReq(t, r, "GET", "/api/v1/categories?type=INCOME", fx.owner.String(), "")
	inc := decodeBody(t, w)["data"].([]any)
	for _, item := range inc {
		if item.(map[string]any)["type"].(string) != "INCOME" {
			t.Fatalf("income filter leaked expense category: %v", item)
		}
	}
}

func TestCategoryValidationAndDuplicate(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	w := doReq(t, r, "POST", "/api/v1/categories", fx.owner.String(), `{"name":"","type":"EXPENSE"}`)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 empty name, got %d", w.Code)
	}
	w = doReq(t, r, "POST", "/api/v1/categories", fx.owner.String(), `{"name":"x","type":"WEIRD"}`)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 bad type, got %d", w.Code)
	}
	doReq(t, r, "POST", "/api/v1/categories", fx.owner.String(), `{"name":"Dup","type":"EXPENSE"}`)
	w = doReq(t, r, "POST", "/api/v1/categories", fx.owner.String(), `{"name":"Dup","type":"EXPENSE"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 duplicate, got %d %s", w.Code, w.Body.String())
	}
}

func TestSystemCategoryCannotBeModified(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	w := doReq(t, r, "GET", "/api/v1/categories", fx.owner.String(), "")
	var systemID string
	for _, item := range decodeBody(t, w)["data"].([]any) {
		cat := item.(map[string]any)
		if cat["is_system"].(bool) {
			systemID = cat["id"].(string)
			break
		}
	}
	w = doReq(t, r, "PATCH", "/api/v1/categories/"+systemID, fx.owner.String(), `{"name":"Hacked"}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("system category must be immutable, got %d %s", w.Code, w.Body.String())
	}
}

func TestCategoryCustomIsolation(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	fx2 := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	w := doReq(t, r, "POST", "/api/v1/categories", fx.owner.String(), `{"name":"Private","type":"EXPENSE"}`)
	id := decodeBody(t, w)["data"].(map[string]any)["id"].(string)

	// foreign custom category inaccessible from fx2's own workspace
	r2 := newRouter(t, db, fx2.wsID)
	w = doReq(t, r2, "GET", "/api/v1/categories?include_archived=true", fx2.owner.String(), "")
	list := decodeBody(t, w)["data"].([]any)
	for _, item := range list {
		if item.(map[string]any)["id"].(string) == id {
			t.Fatalf("foreign custom category leaked into list")
		}
	}
}

func TestCategoryViewerCannotMutate(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	fx := fixture(t, db)
	r := newRouter(t, db, fx.wsID)
	w := doReq(t, r, "POST", "/api/v1/categories", fx.view.String(), `{"name":"x","type":"EXPENSE"}`)
	if w.Code != http.StatusForbidden {
		t.Fatalf("VIEWER create expected 403, got %d", w.Code)
	}
}