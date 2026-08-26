package accounts

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/httpx"
	"github.com/savio/savio/backend/internal/platform/money"
	"github.com/savio/savio/backend/internal/transactions"
)

type Handler struct {
	svc *Service
	trx *transactions.Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// WithTransactions injects the ledger service for reconciliation.
func (h *Handler) WithTransactions(trx *transactions.Service) *Handler {
	h.trx = trx
	return h
}

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
	typ := c.Query("type")
	if typ != "" && !ValidType(typ) {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"type": "Unsupported account type"}))
		return
	}
	sortField := c.Query("sort")
	if sortField == "" {
		sortField = "created_at"
	}
	allowedSort := map[string]bool{"name": true, "type": true, "opening_balance": true, "created_at": true, "updated_at": true}
	if !allowedSort[sortField] {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"sort": "Unsupported sort field"}))
		return
	}
	order := c.Query("order")
	if order == "" {
		order = "asc"
	}
	if order != "asc" && order != "desc" {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"order": "order must be asc or desc"}))
		return
	}
	rows, total, err := h.svc.List(c.Request.Context(), ctx.WorkspaceID, status, typ, sortField, order == "desc", pg.Page, pg.Limit, pg.Offset())
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

type setStatusReq struct {
	Version *int64 `json:"version"`
}

func (h *Handler) setStatus(c *gin.Context, archived bool) {
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
	var req setStatusReq
	if raw, rerr := io.ReadAll(c.Request.Body); rerr == nil && len(raw) > 0 {
		if uerr := json.Unmarshal(raw, &req); uerr != nil {
			httpx.Fail(c, errs.Validation("Request body is not valid JSON"))
			return
		}
	}
	v, err := h.svc.SetStatus(c.Request.Context(), ctx.WorkspaceID, id, archived, req.Version)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) Archive(c *gin.Context) { h.setStatus(c, true) }
func (h *Handler) Restore(c *gin.Context) { h.setStatus(c, false) }

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

type reconcileReq struct {
	ActualBalance string `json:"actual_balance"`
	Reason        string `json:"reason"`
}

// Reconcile aligns the tracked balance with the physical account by creating
// a signed ADJUSTMENT (AGENTS #29: never rewrite history).
func (h *Handler) Reconcile(c *gin.Context) {
	if h.trx == nil {
		httpx.Fail(c, errs.Internal(nil))
		return
	}
	ax, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	id, err := httpx.ParseUUID(c, "id")
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req reconcileReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	actual, err := money.ParseMinorUnits(req.ActualBalance)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"actual_balance": "actual_balance must be a valid decimal"}))
		return
	}
	view, err := h.svc.Get(c.Request.Context(), ax.WorkspaceID, id)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	diff := actual - view.DerivedBalance
	if diff == 0 {
		httpx.Fail(c, errs.BusinessConflict("BUSINESS_CONFLICT", "The account already matches the stated balance. No adjustment needed."))
		return
	}
	reason := req.Reason
	if reason == "" {
		reason = "Reconciliation adjustment"
	}
	adj, err := h.trx.Create(c.Request.Context(), ax.WorkspaceID, ax.UserID, &transactions.CreateInput{
		AccountID:       id,
		Type:            string(transactions.TypeAdjustment),
		AmountMinor:     diff,
		TransactionDate: timeNowUTC(),
		Description:     reason,
		Source:          string(transactions.SourceSystem),
		Status:          string(transactions.StatusPosted),
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, gin.H{
		"adjustment": adj,
		"difference": money.FormatMinorUnits(diff),
	})
}

func timeNowUTC() time.Time { return time.Now().UTC() }

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
	g.POST("/:id/reconcile", writeOnly, h.Reconcile)
	g.DELETE("/:id", writeOnly, h.Delete)
}
