package recurring

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/httpx"
	"github.com/savio/savio/backend/internal/platform/money"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type createReq struct {
	AccountID   string  `json:"account_id"`
	CategoryID  *string `json:"category_id"`
	Type        string  `json:"type"`
	Amount      string  `json:"amount"`
	Frequency   string  `json:"frequency"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date"`
	Description string  `json:"description"`
	Merchant    string  `json:"merchant"`
	Notes       string  `json:"notes"`
	AutoPost    bool    `json:"auto_post"`
}

type updateReq struct {
	AccountID   string  `json:"account_id"`
	CategoryID  *string `json:"category_id"`
	Type        string  `json:"type"`
	Amount      string  `json:"amount"`
	Frequency   string  `json:"frequency"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date"`
	Description string  `json:"description"`
	Merchant    string  `json:"merchant"`
	Notes       string  `json:"notes"`
	AutoPost    bool    `json:"auto_post"`
	Version     *int64  `json:"version"`
}

type statusReq struct {
	Version *int64 `json:"version"`
}

func (r *statusReq) versionOrZero() int64 {
	if r == nil || r.Version == nil {
		return 0
	}
	return *r.Version
}

func (u *updateReq) versionOrZero() int64 {
	if u == nil || u.Version == nil {
		return 0
	}
	return *u.Version
}

func (h *Handler) List(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	rows, err := h.svc.List(c.Request.Context(), ctx.WorkspaceID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, rows)
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
	in, err := toCreateInput(&req)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	v, err := h.svc.Create(c.Request.Context(), ctx.WorkspaceID, ctx.UserID, in)
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
	in, err := toUpdateInput(id, &req)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	v, err := h.svc.Update(c.Request.Context(), ctx.WorkspaceID, ctx.UserID, in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) Pause(c *gin.Context) { h.changeStatus(c, StatusPaused) }
func (h *Handler) Resume(c *gin.Context) { h.changeStatus(c, StatusActive) }
func (h *Handler) End(c *gin.Context)    { h.changeStatus(c, StatusEnded) }

func (h *Handler) changeStatus(c *gin.Context, to Status) {
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
	var req statusReq
	_ = httpx.Bind(c, &req)
	v, err := h.svc.SetStatus(c.Request.Context(), ctx.WorkspaceID, ctx.UserID, id, to, req.versionOrZero())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) Occurrences(c *gin.Context) {
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
	pg := httpx.ParsePagination(c)
	rows, total, err := h.svc.Occurrences(c.Request.Context(), ctx.WorkspaceID, id,
		c.Query("status"), c.Query("from"), c.Query("to"), pg.Page, pg.Limit, pg.Offset())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Collection(c, rows, pg.Page, pg.Limit, int(total))
}

func (h *Handler) Confirm(c *gin.Context) {
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
	var req statusReq
	_ = httpx.Bind(c, &req)
	v, err := h.svc.Confirm(c.Request.Context(), ctx.WorkspaceID, ctx.UserID, id, req.versionOrZero())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) Skip(c *gin.Context) {
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
	var req statusReq
	_ = httpx.Bind(c, &req)
	v, err := h.svc.Skip(c.Request.Context(), ctx.WorkspaceID, ctx.UserID, id, req.versionOrZero())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func toCreateInput(req *createReq) (*CreateInput, error) {
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, errs.ValidationFields(map[string]string{"account_id": "invalid account id"})
	}
	var categoryID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		parsed, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, errs.ValidationFields(map[string]string{"category_id": "invalid category id"})
		}
		categoryID = &parsed
	}
	amount, err := money.ParseMinorUnits(req.Amount)
	if err != nil {
		return nil, errs.ValidationFields(map[string]string{"amount": "amount must be a valid decimal"})
	}
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errs.ValidationFields(map[string]string{"start_date": "date must use YYYY-MM-DD"})
	}
	var end *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		e, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, errs.ValidationFields(map[string]string{"end_date": "date must use YYYY-MM-DD"})
		}
		end = &e
	}
	return &CreateInput{
		AccountID: accountID, CategoryID: categoryID, Type: req.Type, AmountMinor: amount,
		Frequency: req.Frequency, StartDate: start, EndDate: end, Description: req.Description,
		Merchant: req.Merchant, Notes: req.Notes, AutoPost: req.AutoPost,
	}, nil
}

func toUpdateInput(id uuid.UUID, req *updateReq) (*UpdateInput, error) {
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, errs.ValidationFields(map[string]string{"account_id": "invalid account id"})
	}
	var categoryID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		parsed, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, errs.ValidationFields(map[string]string{"category_id": "invalid category id"})
		}
		categoryID = &parsed
	}
	amount, err := money.ParseMinorUnits(req.Amount)
	if err != nil {
		return nil, errs.ValidationFields(map[string]string{"amount": "amount must be a valid decimal"})
	}
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errs.ValidationFields(map[string]string{"start_date": "date must use YYYY-MM-DD"})
	}
	var end *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		e, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, errs.ValidationFields(map[string]string{"end_date": "date must use YYYY-MM-DD"})
		}
		end = &e
	}
	return &UpdateInput{
		ID: id, AccountID: accountID, CategoryID: categoryID, Type: req.Type, AmountMinor: amount,
		Frequency: req.Frequency, StartDate: start, EndDate: end, Description: req.Description,
		Merchant: req.Merchant, Notes: req.Notes, AutoPost: req.AutoPost, Version: req.versionOrZero(),
	}, nil
}

func RegisterRoutes(g *gin.RouterGroup, h *Handler, occurrences *gin.RouterGroup, writeOnly gin.HandlerFunc) {
	g.GET("", h.List)
	g.POST("", writeOnly, h.Create)
	g.GET("/:id", h.Get)
	g.PATCH("/:id", writeOnly, h.Update)
	g.POST("/:id/pause", writeOnly, h.Pause)
	g.POST("/:id/resume", writeOnly, h.Resume)
	g.POST("/:id/end", writeOnly, h.End)
	g.GET("/:id/occurrences", h.Occurrences)
	occurrences.POST("/:id/confirm", writeOnly, h.Confirm)
	occurrences.POST("/:id/skip", writeOnly, h.Skip)
}