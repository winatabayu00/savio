package transactions

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
	AccountID       string  `json:"account_id"`
	CategoryID      *string `json:"category_id"`
	Type            string  `json:"type"`
	Amount          string  `json:"amount"`
	TransactionDate string  `json:"transaction_date"`
	Description     string  `json:"description"`
	Merchant        string  `json:"merchant"`
	Notes           string  `json:"notes"`
	Source          string  `json:"source"`
	Status          string  `json:"status"`
}

type updateReq struct {
	CategoryID      *string `json:"category_id"`
	Type            string  `json:"type"`
	Amount          string  `json:"amount"`
	TransactionDate string  `json:"transaction_date"`
	Description     string  `json:"description"`
	Merchant        string  `json:"merchant"`
	Notes           string  `json:"notes"`
	Version         *int64  `json:"version"`
}

type voidReq struct {
	Reason  string `json:"reason"`
	Version *int64 `json:"version"`
}

func (h *Handler) List(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	pg := httpx.ParsePagination(c)

	filter := ListFilter{
		Search:    c.Query("search"),
		Status:    c.Query("status"),
		DateFrom:  c.Query("from"),
		DateTo:    c.Query("to"),
		Page:      pg.Page,
		Limit:     pg.Limit,
		Offset:    pg.Offset(),
		SortField: c.Query("sort"),
	}
	if filter.SortField == "" {
		filter.SortField = "t.transaction_date"
		filter.SortDesc = true
	}
	if t := c.Query("type"); t != "" {
		if !ValidType(t) {
			httpx.Fail(c, errs.ValidationFields(map[string]string{"type": "type must be INCOME, EXPENSE or ADJUSTMENT"}))
			return
		}
		filter.Type = t
	}
	if s := c.Query("status"); s != "" && s != string(StatusDraft) && s != string(StatusPosted) && s != string(StatusVoided) {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"status": "status must be DRAFT, POSTED or VOIDED"}))
		return
	}
	if a := c.Query("account_id"); a != "" {
		filter.AccountID, err = uuid.Parse(a)
		if err != nil {
			httpx.Fail(c, errs.ValidationFields(map[string]string{"account_id": "invalid account id"}))
			return
		}
	}
	if cat := c.Query("category_id"); cat != "" {
		filter.CategoryID, err = uuid.Parse(cat)
		if err != nil {
			httpx.Fail(c, errs.ValidationFields(map[string]string{"category_id": "invalid category id"}))
			return
		}
	}
	if filter.DateFrom != "" {
		if _, err := parseDate(filter.DateFrom); err != nil {
			httpx.Fail(c, errs.ValidationFields(map[string]string{"from": "date must use YYYY-MM-DD"}))
			return
		}
	}
	if filter.DateTo != "" {
		if _, err := parseDate(filter.DateTo); err != nil {
			httpx.Fail(c, errs.ValidationFields(map[string]string{"to": "date must use YYYY-MM-DD"}))
			return
		}
	}

	allowedSort := map[string]bool{
		"t.transaction_date": true,
		"t.amount":           true,
		"t.created_at":       true,
	}
	if !allowedSort[filter.SortField] {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"sort": "Unsupported sort field"}))
		return
	}
	order := c.DefaultQuery("order", "desc")
	filter.SortDesc = !(order == "asc")

	rows, total, err := h.svc.List(c.Request.Context(), ctx.WorkspaceID, filter)
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
	amount, err := money.ParseMinorUnits(req.Amount)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"amount": "amount must be a valid decimal"}))
		return
	}
	date, err := parseDate(req.TransactionDate)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"transaction_date": "date must use YYYY-MM-DD"}))
		return
	}
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"account_id": "invalid account id"}))
		return
	}
	var categoryID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		parsed, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			httpx.Fail(c, errs.ValidationFields(map[string]string{"category_id": "invalid category id"}))
			return
		}
		categoryID = &parsed
	}
	v, err := h.svc.Create(c.Request.Context(), ctx.WorkspaceID, ctx.UserID, &CreateInput{
		AccountID:       accountID,
		CategoryID:      categoryID,
		Type:            req.Type,
		AmountMinor:     amount,
		TransactionDate: date,
		Description:     req.Description,
		Merchant:        req.Merchant,
		Notes:           req.Notes,
		Source:          req.Source,
		Status:          req.Status,
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
	amount, err := money.ParseMinorUnits(req.Amount)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"amount": "amount must be a valid decimal"}))
		return
	}
	date, err := parseDate(req.TransactionDate)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"transaction_date": "date must use YYYY-MM-DD"}))
		return
	}
	var categoryID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		parsed, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			httpx.Fail(c, errs.ValidationFields(map[string]string{"category_id": "invalid category id"}))
			return
		}
		categoryID = &parsed
	}
	v, err := h.svc.Update(c.Request.Context(), ctx.WorkspaceID, ctx.UserID, &UpdateInput{
		ID:              id,
		CategoryID:      categoryID,
		Type:            req.Type,
		AmountMinor:     amount,
		TransactionDate: date,
		Description:     req.Description,
		Merchant:        req.Merchant,
		Notes:           req.Notes,
		Version:         req.VersionOrDefault(),
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) Post(c *gin.Context) {
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
	var req voidReq
	_ = httpx.Bind(c, &req)
	v, err := h.svc.Post(c.Request.Context(), ctx.WorkspaceID, ctx.UserID, id, req.VersionOrDefault())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) Void(c *gin.Context) {
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
	var req voidReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	v, err := h.svc.Void(c.Request.Context(), ctx.WorkspaceID, ctx.UserID, &VoidInput{
		ID:      id,
		Reason:  req.Reason,
		Version: req.VersionOrDefault(),
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errs.Validation("date required")
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, err
	}
	// normalize to midnight UTC to keep date-only fields timezone-stable
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

func (v *voidReq) VersionOrDefault() int64 {
	if v == nil || v.Version == nil {
		return 0
	}
	return *v.Version
}

func (u *updateReq) VersionOrDefault() int64 {
	if u == nil || u.Version == nil {
		return 0
	}
	return *u.Version
}

func RegisterRoutes(g *gin.RouterGroup, h *Handler, writeOnly gin.HandlerFunc) {
	g.GET("", h.List)
	g.POST("", writeOnly, h.Create)
	g.GET("/:id", h.Get)
	g.PATCH("/:id", writeOnly, h.Update)
	g.POST("/:id/post", writeOnly, h.Post)
	g.POST("/:id/void", writeOnly, h.Void)
}
