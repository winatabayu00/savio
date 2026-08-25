package httpx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestFailIncludesRequestIDButNotCause(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(requestIDKey, "req_test")
	Fail(c, errors.New("database password leaked"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", w.Code)
	}
	if got := w.Body.String(); !strings.Contains(got, "request_id") || strings.Contains(got, "database password leaked") {
		t.Fatalf("unsafe error response: %s", got)
	}
}
