package goals_test

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
	"github.com/savio/savio/backend/internal/goals"
	"github.com/savio/savio/backend/internal/migrations"
	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/users"
	"github.com/savio/savio/backend/internal/workspaces"
)

func TestMain(m *testing.M) {
	if v := os.Getenv("DATABASE_URL"); v != "" {
		dsn := v
		testURL, err := url.Parse(dsn)
		if err == nil {
			testURL.Path = "/savio_test_goals"
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
	_, _ = c.Exec(`CREATE DATABASE savio_test_goals`)
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

func fixture(t *testing.T) (uuid.UUID, uuid.UUID) {
	t.Helper()
	wsID := uuid.New()
	mustNil(t, db.Create(&workspaces.Workspace{ID: wsID, Name: "T", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)
	owner := uuid.New()
	mustNil(t, db.Create(&users.User{ID: owner, Name: "O", Email: "o-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: owner, Role: "OWNER", Status: "ACTIVE"}).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM audit_logs WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM goals WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		db.Exec(`DELETE FROM users WHERE id = $1`, owner)
	})
	return wsID, owner
}

func newRouter(t *testing.T, wsID uuid.UUID, h *goals.Handler) *gin.Engine {
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
	g := r.Group("/api/v1/goals", authMW)
	goals.RegisterRoutes(g, h, auth.RequireWrite())
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

func TestGoalMetricsProgressAndRequiredContribution(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	wsID, owner := fixture(t)
	r := newRouter(t, wsID, goals.NewHandler(goals.NewService(db)))
	// 24 months out, 25% saved, target higher than 0 contributions needed
	w := doReq(t, r, "POST", "/api/v1/goals", owner.String(),
		`{"name":"Emergency Fund","target_amount":"1000000","current_amount":"250000","target_date":"2028-08-01","priority":"HIGH"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	data := decodeBody(t, w)["data"].(map[string]any)
	if data["progress_percent"].(float64) < 24.9 || data["progress_percent"].(float64) > 25.1 {
		t.Fatalf("progress = %v", data["progress_percent"])
	}
	if data["months_remaining"].(float64) < 23 || data["months_remaining"].(float64) > 24 {
		t.Fatalf("months = %v", data["months_remaining"])
	}
	if data["progress_percent"].(float64) > 100 {
		t.Fatalf("progress must cap at 100")
	}
	if data["feasibility"].(string) != "AT_RISK" && data["feasibility"].(string) != "ON_TRACK" {
		t.Fatalf("feasibility = %v", data["feasibility"])
	}
	if data["remaining"].(string) != "750000.00" {
		t.Fatalf("remaining = %v", data["remaining"])
	}
}

func TestGoalProgressCapsAt100(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	wsID, owner := fixture(t)
	svc := goals.NewService(db)
	v, err := svc.Create(t.Context(), wsID, owner, &goals.CreateInput{
		Name: "Done", TargetAmount: 100000, CurrentAmount: 250000, Priority: "MEDIUM",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if v.ProgressPercent != 100 {
		t.Fatalf("progress = %v, want 100", v.ProgressPercent)
	}
	if v.Remaining != "0.00" {
		t.Fatalf("remaining = %v", v.Remaining)
	}
}

func TestGoalLifecycleStatusFlows(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	wsID, owner := fixture(t)
	r := newRouter(t, wsID, goals.NewHandler(goals.NewService(db)))
	w := doReq(t, r, "POST", "/api/v1/goals", owner.String(),
		`{"name":"Trip","target_amount":"500000","current_amount":"0","target_date":"2027-01-01","priority":"LOW"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d", w.Code)
	}
	body := decodeBody(t, w)["data"].(map[string]any)
	id := body["id"].(string)
	version := int64(body["version"].(float64))

	pauseBody, _ := json.Marshal(map[string]any{"version": version})
	w = doReq(t, r, "POST", "/api/v1/goals/"+id+"/pause", owner.String(), string(pauseBody))
	if w.Code != http.StatusOK || decodeBody(t, w)["data"].(map[string]any)["status"].(string) != "PAUSED" {
		t.Fatalf("pause: %d %s", w.Code, w.Body.String())
	}
	version = int64(decodeBody(t, w)["data"].(map[string]any)["version"].(float64))
	resumeBody, _ := json.Marshal(map[string]any{"version": version})
	w = doReq(t, r, "POST", "/api/v1/goals/"+id+"/resume", owner.String(), string(resumeBody))
	if w.Code != http.StatusOK || decodeBody(t, w)["data"].(map[string]any)["status"].(string) != "ACTIVE" {
		t.Fatalf("resume: %d %s", w.Code, w.Body.String())
	}
	version = int64(decodeBody(t, w)["data"].(map[string]any)["version"].(float64))
	achieveBody, _ := json.Marshal(map[string]any{"version": version})
	w = doReq(t, r, "POST", "/api/v1/goals/"+id+"/achieve", owner.String(), string(achieveBody))
	if w.Code != http.StatusOK || decodeBody(t, w)["data"].(map[string]any)["status"].(string) != "ACHIEVED" {
		t.Fatalf("achieve: %d %s", w.Code, w.Body.String())
	}
}

func TestGoalVersionConflict(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	wsID, owner := fixture(t)
	r := newRouter(t, wsID, goals.NewHandler(goals.NewService(db)))
	w := doReq(t, r, "POST", "/api/v1/goals", owner.String(),
		`{"name":"X","target_amount":"100000","current_amount":"0","priority":"MEDIUM"}`)
	id := decodeBody(t, w)["data"].(map[string]any)["id"].(string)
	stale, _ := json.Marshal(map[string]any{"name": "X", "target_amount": "100000", "current_amount": "0", "priority": "MEDIUM", "version": 1})
	doReq(t, r, "PATCH", "/api/v1/goals/"+id, owner.String(), string(stale))
	w = doReq(t, r, "PATCH", "/api/v1/goals/"+id, owner.String(), string(stale))
	if w.Code != http.StatusConflict {
		t.Fatalf("stale PATCH expected 409, got %d %s", w.Code, w.Body.String())
	}
}

