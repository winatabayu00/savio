package accounts

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

type createReq struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	Currency        string `json:"currency"`
	OpeningBalance  *int64 `json:"opening_balance"`
	InstitutionName string `json:"institution_name"`
	Description     string `json:"description"`
}

type updateReq struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	InstitutionName string `json:"institution_name"`
	Description     string `json:"description"`
	Version         *int64 `json:"version"`
}

func (h *Handler) List(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	pg := httpx.ParsePagination(c)
	status := c.Query("status")
	if status != "" && status != string(StatusActive) && status != string(StatusArchived) {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"status": "status must be ACTIVE or ARCHIVED"}))
		return
	}
	rows, total, err := h.svc.List(c.Request.Context(), ctx.WorkspaceID, status, pg.Page, pg.Limit, pg.Offset())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Collection(c, rows, pg.Page, pg.Limit, int(total))
}

func (h *Handler) Get(c *gin.Context) {
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
	v, err := h.svc.Get(c.Request.Context(), ctx.WorkspaceID, id)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) Create(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req createReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	balance := int64(0)
	if req.OpeningBalance != nil {
		balance = *req.OpeningBalance
	}
	v, err := h.svc.Create(c.Request.Context(), ctx.WorkspaceID, ctx.UserID, &CreateInput{
		Name:            req.Name,
		Type:            req.Type,
		Currency:        req.Currency,
		OpeningBalance:  balance,
		InstitutionName: req.InstitutionName,
		Description:     req.Description,
	})
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
	var req updateReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	v, err := h.svc.Update(c.Request.Context(), ctx.WorkspaceID, &UpdateInput{
		ID:              id,
		Name:            req.Name,
		Type:            req.Type,
		InstitutionName: req.InstitutionName,
		Description:     req.Description,
		Version:         derefVersion(req.Version),
	})
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

func (h *Handler) Delete(c *gin.Context) {
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
	if err := h.svc.Delete(c.Request.Context(), ctx.WorkspaceID, id); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, gin.H{})
}

func derefVersion(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

// RegisterRoutes wires the account endpoints. writeOnly enforces VIEWER
// cannot mutate; roles are passed in to keep the auth dependency out of the
// module.
func RegisterRoutes(g *gin.RouterGroup, h *Handler, writeOnly gin.HandlerFunc) {
	g.GET("", h.List)
	g.POST("", writeOnly, h.Create)
	g.GET("/:id", h.Get)
	g.PATCH("/:id", writeOnly, h.Update)
	g.POST("/:id/archive", writeOnly, h.Archive)
	g.POST("/:id/restore", writeOnly, h.Restore)
	g.DELETE("/:id", writeOnly, h.Delete)
}