package httpx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestFailIncludesTraceButNotCauseByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(requestIDKey, "req_test")
	Fail(c, errors.New("database password leaked"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", w.Code)
	}
	if got := w.Body.String(); !strings.Contains(got, `"trace_id":"req_test"`) || !strings.Contains(got, `"endpoint":"/"`) || strings.Contains(got, "database password leaked") {
		t.Fatalf("unsafe error response: %s", got)
	}
}

func TestFailIncludesCauseWhenErrorDetailsEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	c.Set(requestIDKey, "req_test")
	c.Set("error_details", true)
	Fail(c, errors.New("database unavailable"))

	if got := w.Body.String(); !strings.Contains(got, `"reason":"database unavailable"`) {
		t.Fatalf("missing development error detail: %s", got)
	}
}
