package workspaces_test

import (
	"bytes"
	"context"
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

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return connect(t, v)
	}
	return connect(t, "postgres://savio:savio@localhost:5433/savio?sslmode=disable")
}

func connect(t *testing.T, dsn string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Skipf("postgres not reachable: %v", err)
	}
	return db
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

// newRouter builds a gin engine that mirrors the production wiring for the
// workspace endpoints: an auth middleware resolves the requester role from
// DB membership (like AuthRequired) and RequireOwner gates mutations, so the
// tests prove backend enforcement, not just handler behavior.
func newRouter(t *testing.T, db *gorm.DB, wsID uuid.UUID) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := workspaces.NewHandler(workspaces.NewService(db))
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
	g := r.Group("/api/v1/workspaces", testAuth)
	workspaces.RegisterRoutes(g, h, auth.RequireOwner())
	return r
}

// fixture creates a workspace owned by owner, plus optional extra users.
// Returns workspace id, owner user id, and a cleanup removing all created rows.
func fixture(t *testing.T, db *gorm.DB, extra int) (uuid.UUID, uuid.UUID, []uuid.UUID) {
	t.Helper()
	wsID := uuid.New()
	mustNil(t, db.Create(&workspaces.Workspace{ID: wsID, Name: "Test WS", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)

	owner := uuid.New()
	mustNil(t, db.Create(&users.User{ID: owner, Name: "Owner", Email: "o-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: owner, Role: "OWNER", Status: "ACTIVE"}).Error)

	ids := []uuid.UUID{}
	for i := 0; i < extra; i++ {
		uid := uuid.New()
		mustNil(t, db.Create(&users.User{ID: uid, Name: "User", Email: "u-" + uuid.NewString()[:10] + "@savio.test",
			PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
		mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: uid,
			Role: "MEMBER", Status: "ACTIVE"}).Error)
		ids = append(ids, uid)
	}

	t.Cleanup(func() {
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		for _, id := range append([]uuid.UUID{owner}, ids...) {
			db.Exec(`DELETE FROM users WHERE id = $1`, id)
		}
	})
	return wsID, owner, ids
}

func membershipIDFor(t *testing.T, db *gorm.DB, wsID, userID uuid.UUID) string {
	t.Helper()
	var m workspaces.Membership
	mustNil(t, db.Where("workspace_id = ? AND user_id = ?", wsID, userID).First(&m).Error)
	return m.ID.String()
}

func TestOwnerFullAccessDrive(t *testing.T) {
	db := testDB(t)
	wsID, owner, members := fixture(t, db, 1)
	r := newRouter(t, db, wsID)

	if w := doReq(t, r, http.MethodGet, "/api/v1/workspaces/current", owner.String(), ""); w.Code != http.StatusOK {
		t.Fatalf("owner get current: %d", w.Code)
	}
	if w := doReq(t, r, http.MethodGet, "/api/v1/workspaces/current/members", owner.String(), ""); w.Code != http.StatusOK {
		t.Fatalf("owner list members: %d", w.Code)
	}
	if w := doReq(t, r, http.MethodPost, "/api/v1/workspaces/current/members",
		owner.String(), `{"email":"nobody@savio.test","role":"VIEWER"}`); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("add unknown email should 422, got %d", w.Code)
	}
	if w := doReq(t, r, http.MethodPost, "/api/v1/workspaces/current/members",
		owner.String(), `{"email":"u-nope@savio.test","role":"OWNER"}`); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("assigning OWNER on invite should 422, got %d", w.Code)
	}
	if w := doReq(t, r, http.MethodPost, "/api/v1/workspaces/current/members",
		owner.String(), `{"email":"`+dbEmail(t, db, members[0])+`","role":"VIEWER"}`); w.Code != http.StatusConflict {
		t.Fatalf("re-inviting existing member should 409, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestViewerReadOnlyAllowedMutationMidcommittee(t *testing.T) {
	db := testDB(t)
	wsID, owner, _ := fixture(t, db, 0)

	viewer := uuid.New()
	mustNil(t, db.Create(&users.User{ID: viewer, Name: "Viewer", Email: "v-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: viewer, Role: "VIEWER", Status: "ACTIVE"}).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM workspace_memberships WHERE user_id = $1`, viewer)
		db.Exec(`DELETE FROM users WHERE id = $1`, viewer)
	})

	r := newRouter(t, db, wsID)
	if w := doReq(t, r, http.MethodGet, "/api/v1/workspaces/current", viewer.String(), ""); w.Code != http.StatusOK {
		t.Fatalf("viewer read should 200, got %d", w.Code)
	}
	if w := doReq(t, r, http.MethodGet, "/api/v1/workspaces/current/members", viewer.String(), ""); w.Code != http.StatusOK {
		t.Fatalf("viewer list should 200, got %d", w.Code)
	}
	if w := doReq(t, r, http.MethodPost, "/api/v1/workspaces/current/members", viewer.String(), `{"email":"x@savio.test","role":"VIEWER"}`); w.Code != http.StatusForbidden {
		t.Fatalf("viewer mutation should 403, got %d", w.Code)
	}
	if w := doReq(t, r, http.MethodPatch, "/api/v1/workspaces/current/members/"+membershipIDFor(t, db, wsID, owner), viewer.String(), `{"role":"MEMBER"}`); w.Code != http.StatusForbidden {
		t.Fatalf("viewer role change should 403, got %d", w.Code)
	}
	if w := doReq(t, r, http.MethodDelete, "/api/v1/workspaces/current/members/"+membershipIDFor(t, db, wsID, owner), viewer.String(), ""); w.Code != http.StatusForbidden {
		t.Fatalf("viewer removal should 403, got %d", w.Code)
	}
}

func TestMemberCannotManageMembers(t *testing.T) {
	db := testDB(t)
	wsID, owner, members := fixture(t, db, 1)
	r := newRouter(t, db, wsID)

	if w := doReq(t, r, http.MethodPost, "/api/v1/workspaces/current/members", members[0].String(), `{"email":"x@savio.test"}`); w.Code != http.StatusForbidden {
		t.Fatalf("member add should 403, got %d", w.Code)
	}
	// always a target member: the owner
	if w := doReq(t, r, http.MethodPatch, "/api/v1/workspaces/current/members/"+membershipIDFor(t, db, wsID, owner), members[0].String(), `{"role":"OWNER"}`); w.Code != http.StatusForbidden {
		t.Fatalf("member role change should 403, got %d", w.Code)
	}
}

// TestForeignWorkspaceMemberIDInvisible verifies IDOR protection: a member id
// from another workspace is untouchable (INV-019).
func TestForeignWorkspaceMemberIDInvisible(t *testing.T) {
	db := testDB(t)
	wsID, owner, _ := fixture(t, db, 0)

	otherWS, otherOwner, _ := fixture(t, db, 0)
	foreignID := membershipIDFor(t, db, otherWS, otherOwner)

	r := newRouter(t, db, wsID)
	if w := doReq(t, r, http.MethodPatch, "/api/v1/workspaces/current/members/"+foreignID, owner.String(), `{"role":"VIEWER"}`); w.Code != http.StatusNotFound {
		t.Fatalf("foreign member id should 404, got %d", w.Code)
	}
	if w := doReq(t, r, http.MethodDelete, "/api/v1/workspaces/current/members/"+foreignID, owner.String(), ""); w.Code != http.StatusNotFound {
		t.Fatalf("foreign member removal should 404, got %d", w.Code)
	}
}

func TestLastOwnerCannotBeDemotedOrRemoved(t *testing.T) {
	db := testDB(t)
	wsID, owner, _ := fixture(t, db, 0)
	r := newRouter(t, db, wsID)
	ownerMid := membershipIDFor(t, db, wsID, owner)

	if w := doReq(t, r, http.MethodPatch, "/api/v1/workspaces/current/members/"+ownerMid, owner.String(), `{"role":"MEMBER"}`); w.Code != http.StatusConflict {
		t.Fatalf("demoting last owner should 409, got %d (%s)", w.Code, w.Body.String())
	}
	if w := doReq(t, r, http.MethodDelete, "/api/v1/workspaces/current/members/"+ownerMid, owner.String(), ""); w.Code != http.StatusConflict {
		t.Fatalf("removing last owner should 409, got %d", w.Code)
	}
}

// TestOwnershipTransferHappyPath: promote a member to OWNER, then the old
// owner can demote/remove themselves. Matches the M05/RBAC demo flow.
func TestOwnershipTransferHappyPath(t *testing.T) {
	db := testDB(t)
	wsID, owner, members := fixture(t, db, 1)
	r := newRouter(t, db, wsID)
	memberMid := membershipIDFor(t, db, wsID, members[0])

	if w := doReq(t, r, http.MethodPatch, "/api/v1/workspaces/current/members/"+memberMid, owner.String(), `{"role":"OWNER"}`); w.Code != http.StatusOK {
		t.Fatalf("promote to OWNER should 200, got %d (%s)", w.Code, w.Body.String())
	}
	ownerMid := membershipIDFor(t, db, wsID, owner)
	if w := doReq(t, r, http.MethodPatch, "/api/v1/workspaces/current/members/"+ownerMid, owner.String(), `{"role":"MEMBER"}`); w.Code != http.StatusOK {
		t.Fatalf("demote old owner with transfer should 200, got %d (%s)", w.Code, w.Body.String())
	}
	if w := doReq(t, r, http.MethodDelete, "/api/v1/workspaces/current/members/"+ownerMid, members[0].String(), ""); w.Code != http.StatusOK {
		t.Fatalf("new owner removes old member should 200, got %d", w.Code)
	}
}

// TestConcurrentLastOwnerDemotionSerialized: two owners demote each other
// concurrently; exactly one may succeed (INV-003 concurrency).
func TestConcurrentLastOwnerDemotionSerialized(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	wsID := uuid.New()
	mustNil(t, db.Create(&workspaces.Workspace{ID: wsID, Name: "Concurrent WS", BaseCurrency: "IDR", Timezone: "Asia/Jakarta"}).Error)

	mk := func(name string) *workspaces.Membership {
		uid := uuid.New()
		mustNil(t, db.Create(&users.User{ID: uid, Name: name, Email: "c-" + name + "-" + uuid.NewString()[:6] + "@savio.test",
			PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
		m := &workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: uid, Role: "OWNER", Status: "ACTIVE"}
		mustNil(t, db.Create(m).Error)
		return m
	}
	a := mk("A")
	b := mk("B")
	svc := workspaces.NewService(db)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1`, wsID)
		db.Exec(`DELETE FROM workspaces WHERE id = $1`, wsID)
		for _, uid := range []uuid.UUID{a.UserID, b.UserID} {
			db.Exec(`DELETE FROM users WHERE id = $1`, uid)
		}
	})

	// Both owners demote the other. The last-owner count is row-locked inside
	// each transaction, so exactly one demotion may commit.
	done := make(chan error, 2)
	go func() { done <- svc.UpdateRole(ctx, wsID, b.ID, "MEMBER") }()
	go func() { done <- svc.UpdateRole(ctx, wsID, a.ID, "MEMBER") }()
	ok := 0
	for i := 0; i < 2; i++ {
		if err := <-done; err == nil {
			ok++
		}
	}
	if ok != 1 {
		t.Fatalf("expected exactly one demotion to succeed, got %d", ok)
	}
	// sanity: at least one OWNER remains
	var owners int64
	mustNil(t, db.Model(&workspaces.Membership{}).Where("workspace_id = ? AND role = 'OWNER' AND status = 'ACTIVE'", wsID).Count(&owners).Error)
	if owners != 1 {
		t.Fatalf("expected exactly one owner left, got %d", owners)
	}
}

func dbEmail(t *testing.T, db *gorm.DB, userID uuid.UUID) string {
	t.Helper()
	var u users.User
	mustNil(t, db.First(&u, "id = ?", userID).Error)
	return u.Email
}
