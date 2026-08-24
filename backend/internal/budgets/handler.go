package budgets

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

type budgetReq struct {
	CategoryID  string `json:"category_id"`
	Amount      string `json:"amount"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
	Version     *int64 `json:"version"`
}

func parseDates(c *gin.Context, startStr, endStr string) (time.Time, time.Time, bool) {
	start, err1 := time.Parse("2006-01-02", startStr)
	end, err2 := time.Parse("2006-01-02", endStr)
	if err1 != nil || err2 != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"period": "period_start and period_end must use YYYY-MM-DD"}))
		return time.Time{}, time.Time{}, false
	}
	return time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC),
		time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC), true
}

func (h *Handler) List(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	status := c.Query("status")
	if status != "" && status != "ACTIVE" && status != "CLOSED" {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"status": "status must be ACTIVE or CLOSED"}))
		return
	}
	warn := h.svc.WarningThreshold(c.Request.Context(), x.UserID)
	rows, err := h.svc.List(c.Request.Context(), x.WorkspaceID, status, warn)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, rows)
}

func (h *Handler) Get(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	id, err := httpx.ParseUUID(c, "id")
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	v, err := h.svc.Get(c.Request.Context(), x.WorkspaceID, id, h.svc.WarningThreshold(c.Request.Context(), x.UserID))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) Create(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req budgetReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	catID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"category_id": "invalid category id"}))
		return
	}
	amount, err := money.ParseMinorUnits(req.Amount)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"amount": "amount must be a valid decimal"}))
		return
	}
	start, end, ok := parseDates(c, req.PeriodStart, req.PeriodEnd)
	if !ok {
		return
	}
	v, err := h.svc.Create(c.Request.Context(), x.WorkspaceID, x.UserID, &CreateInput{CategoryID: catID, AmountMinor: amount, PeriodStart: start, PeriodEnd: end})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusCreated, v)
}

func (h *Handler) Update(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	id, err := httpx.ParseUUID(c, "id")
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req budgetReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	catID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"category_id": "invalid category id"}))
		return
	}
	amount, err := money.ParseMinorUnits(req.Amount)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"amount": "amount must be a valid decimal"}))
		return
	}
	start, end, ok := parseDates(c, req.PeriodStart, req.PeriodEnd)
	if !ok {
		return
	}
	v, err := h.svc.Update(c.Request.Context(), x.WorkspaceID, x.UserID, &UpdateInput{
		ID: id, CategoryID: catID, AmountMinor: amount, PeriodStart: start, PeriodEnd: end, Version: ver(req.Version),
	}, h.svc.WarningThreshold(c.Request.Context(), x.UserID))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) Close(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	id, err := httpx.ParseUUID(c, "id")
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req struct {
		Version *int64 `json:"version"`
	}
	_ = httpx.Bind(c, &req)
	v, err := h.svc.Close(c.Request.Context(), x.WorkspaceID, x.UserID, id, ver(req.Version), h.svc.WarningThreshold(c.Request.Context(), x.UserID))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func ver(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func RegisterRoutes(g *gin.RouterGroup, h *Handler, writeOnly gin.HandlerFunc) {
	g.GET("", h.List)
	g.POST("", writeOnly, h.Create)
	g.GET("/:id", h.Get)
	g.PATCH("/:id", writeOnly, h.Update)
	g.POST("/:id/close", writeOnly, h.Close)
}