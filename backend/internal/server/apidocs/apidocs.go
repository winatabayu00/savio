// Package apidocs serves the OpenAPI specification and a Swagger UI viewer.
package apidocs

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.yaml
var spec []byte

//go:embed index.html
var index []byte

// Register mounts the docs viewer outside the versioned API surface, next to
// /health and /ready. Read-only infrastructure; no auth required.
func Register(engine *gin.Engine) {
	engine.GET("/api/v1/docs/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml", spec)
	})
	engine.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})
}
