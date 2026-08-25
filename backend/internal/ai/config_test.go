package ai_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/savio/savio/backend/internal/ai"
	"github.com/savio/savio/backend/internal/users"
	"github.com/savio/savio/backend/internal/workspaces"
)

// newConfigRouter registers /config routes with the same auth + owner gate as the app.
func newConfigRouter(t *testing.T, wsID uuid.UUID, h *ai.Handler) *gin.Engine {
	t.Helper()
	r := newRouter(t, wsID, h)
	return r
}

func member(t *testing.T, wsID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	mustNil(t, db.Create(&users.User{ID: id, Name: "M", Email: "m-" + uuid.NewString()[:10] + "@savio.test",
		PasswordHash: "x", Timezone: "Asia/Jakarta", DefaultCurrency: "IDR", Locale: "id-ID", Status: "ACTIVE"}).Error)
	mustNil(t, db.Create(&workspaces.Membership{ID: uuid.New(), WorkspaceID: wsID, UserID: id, Role: "MEMBER", Status: "ACTIVE"}).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM workspace_memberships WHERE workspace_id = $1 AND user_id = $2`, wsID, id)
		db.Exec(`DELETE FROM users WHERE id = $1`, id)
	})
	return id
}

func TestConfigGetAndMaskedKey(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	wsID, owner := fixture(t)
	h := ai.NewHandler(ai.NewService(db, testCfg(true)))
	r := newConfigRouter(t, wsID, h)

	w := doReq(t, r, "PATCH", "/api/v1/ai/config", owner.String(),
		`{"api_key":"sk-secret1234","base_url":"https://api.openai.test/v1","model":"gpt-4o"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("patch config: %d %s", w.Code, w.Body.String())
	}

	w = doReq(t, r, "GET", "/api/v1/ai/config", owner.String(), "")
	if w.Code != http.StatusOK {
		t.Fatalf("get config: %d", w.Code)
	}
	raw := w.Body.String()
	if !json.Valid(w.Body.Bytes()) {
		t.Fatalf("invalid json: %s", raw)
	}
	data := decodeBody(t, w)["data"].(map[string]any)
	if data["api_key_masked"].(string) != "••••1234" {
		t.Fatalf("mask = %v", data["api_key_masked"])
	}
	if _, ok := data["api_key"]; ok {
		t.Fatalf("raw api_key leaked: %s", raw)
	}
	if data["base_url"].(string) != "https://api.openai.test/v1" {
		t.Fatalf("base_url = %v", data["base_url"])
	}
}

func TestConfigUpdatePersistsAcrossServices(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	wsID, owner := fixture(t)
	h := ai.NewHandler(ai.NewService(db, testCfg(true)))
	r := newConfigRouter(t, wsID, h)

	w := doReq(t, r, "PATCH", "/api/v1/ai/config", owner.String(),
		`{"enabled":false,"provider":"openai","base_url":"https://api.openai.test/v1","model":"gpt-4o","timeout_seconds":30}`)
	if w.Code != http.StatusOK {
		t.Fatalf("patch: %d %s", w.Code, w.Body.String())
	}

	// A fresh service reads the same row: user edits win over env defaults.
	h2 := ai.NewHandler(ai.NewService(db, testCfg(true)))
	r2 := newConfigRouter(t, wsID, h2)
	w = doReq(t, r2, "GET", "/api/v1/ai/config", owner.String(), "")
	data := decodeBody(t, w)["data"].(map[string]any)
	if data["enabled"].(bool) || data["provider"].(string) != "openai" || data["model"].(string) != "gpt-4o" {
		t.Fatalf("persisted config = %v", data)
	}
}

func TestConfigRequiresOwner(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	wsID, _ := fixture(t)
	m := member(t, wsID)
	h := ai.NewHandler(ai.NewService(db, testCfg(true)))
	r := newConfigRouter(t, wsID, h)

	w := doReq(t, r, "PATCH", "/api/v1/ai/config", m.String(), `{"enabled":true}`)
	if w.Code != http.StatusForbidden {
		t.Fatalf("member patch expected 403, got %d %s", w.Code, w.Body.String())
	}
	w = doReq(t, r, "GET", "/api/v1/ai/config", m.String(), "")
	if w.Code != http.StatusForbidden {
		t.Fatalf("member get expected 403, got %d", w.Code)
	}
}

func TestConfigValidation(t *testing.T) {
	if db == nil {
		t.Skip("DATABASE_URL not set")
	}
	wsID, owner := fixture(t)
	h := ai.NewHandler(ai.NewService(db, testCfg(true)))
	r := newConfigRouter(t, wsID, h)

	for _, body := range []string{
		`{"provider":"claude"}`,
		`{"timeout_seconds":0}`,
		`{"base_url":"not a url"}`,
	} {
		w := doReq(t, r, "PATCH", "/api/v1/ai/config", owner.String(), body)
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("%s expected 422, got %d %s", body, w.Code, w.Body.String())
		}
	}
}
