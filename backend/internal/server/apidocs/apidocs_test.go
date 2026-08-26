package apidocs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDocsRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r)

	cases := map[string]struct {
		contentType string
		contains    string
	}{
		"/api/v1/docs/openapi.yaml": {"application/yaml", "openapi: 3.0.3"},
		"/docs":                     {"text/html", "swagger-ui"},
	}
	for path, want := range cases {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("%s: status %d", path, w.Code)
		}
		if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, want.contentType) {
			t.Errorf("%s: content type %q", path, ct)
		}
		if !strings.Contains(w.Body.String(), want.contains) {
			t.Errorf("%s: body missing %q", path, want.contains)
		}
	}
}
