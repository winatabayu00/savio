package categories

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type categoryReq struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
}

func (h *Handler) List(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	typ := c.Query("type")
	if typ != "" && !ValidType(typ) {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"type": "type must be INCOME or EXPENSE"}))
		return
	}
	includeArchived := c.Query("include_archived") == "true"
	rows, err := h.svc.List(c.Request.Context(), ctx.WorkspaceID, typ, includeArchived)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, rows)
}

func (h *Handler) Create(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req categoryReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	v, err := h.svc.Create(c.Request.Context(), ctx.WorkspaceID, req.Name, req.Type, req.Icon, req.Description)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusCreated, v)
}

func (h *Handler) Update(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	id, err := httpx.ParseUUID(c, "id")
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req categoryReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	v, err := h.svc.Update(c.Request.Context(), ctx.WorkspaceID, id, req.Name, req.Icon, req.Description)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) Archive(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	id, err := httpx.ParseUUID(c, "id")
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	v, err := h.svc.SetStatus(c.Request.Context(), ctx.WorkspaceID, id, true)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) Restore(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	id, err := httpx.ParseUUID(c, "id")
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	v, err := h.svc.SetStatus(c.Request.Context(), ctx.WorkspaceID, id, false)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func RegisterRoutes(g *gin.RouterGroup, h *Handler, writeOnly gin.HandlerFunc) {
	g.GET("", h.List)
	g.POST("", writeOnly, h.Create)
	g.PATCH("/:id", writeOnly, h.Update)
	g.POST("/:id/archive", writeOnly, h.Archive)
	g.POST("/:id/restore", writeOnly, h.Restore)
}