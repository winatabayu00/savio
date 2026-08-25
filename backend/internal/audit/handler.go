package audit

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/httpx"
)

type Handler struct{ repo *Repository }

func NewHandler(repo *Repository) *Handler { return &Handler{repo: repo} }

func (h *Handler) List(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	pg := httpx.ParsePagination(c)
	entries, total, err := h.repo.List(c.Request.Context(), ctx.WorkspaceID, pg.Limit, pg.Offset())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": entries, "meta": gin.H{"page": pg.Page, "limit": pg.Limit, "total": total}})
}

func RegisterRoutes(g *gin.RouterGroup, h *Handler) { g.GET("", h.List) }
