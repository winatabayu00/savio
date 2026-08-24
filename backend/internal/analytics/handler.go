package analytics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/httpx"
	"github.com/savio/savio/backend/internal/platform/money"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type moneyInt struct {
	income  int64
	expense int64
	net     int64
}

func (h *Handler) bounds(c *gin.Context) (from, to string, ok bool) {
	from = c.Query("from")
	to = c.Query("to")
	if from == "" || to == "" {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"from/to": "from and to are required (YYYY-MM-DD)"}))
		return "", "", false
	}
	if _, err := time.Parse("2006-01-02", from); err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"from": "date must use YYYY-MM-DD"}))
		return "", "", false
	}
	if _, err := time.Parse("2006-01-02", to); err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"to": "date must use YYYY-MM-DD"}))
		return "", "", false
	}
	return from, to, true
}

func (h *Handler) Cashflow(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	from, to, ok := h.bounds(c)
	if !ok {
		return
	}
	v, err := h.svc.Cashflow(c.Request.Context(), x.WorkspaceID, from, to)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, cashflowJSON(v))
}

func (h *Handler) Categories(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	from, to, ok := h.bounds(c)
	if !ok {
		return
	}
	rows, err := h.svc.Categories(c.Request.Context(), x.WorkspaceID, from, to)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"category_id":   r.CategoryID,
			"category_name": r.CategoryName,
			"total":         money.FormatMinorUnits(r.Total),
			"items":         r.Items,
		})
	}
	httpx.Success(c, http.StatusOK, out)
}

func (h *Handler) PeriodComparison(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	from, to, ok := h.bounds(c)
	if !ok {
		return
	}
	prevFrom := c.Query("compare_from")
	prevTo := c.Query("compare_to")
	if prevFrom == "" || prevTo == "" {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"compare_from": "compare_from and compare_to are required"}))
		return
	}
	v, err := h.svc.PeriodComparison(c.Request.Context(), x.WorkspaceID, from, to, prevFrom, prevTo)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, gin.H{
		"current":             cashflowJSON(&v.Current),
		"previous":            cashflowJSON(&v.Previous),
		"income_delta_percent":  v.IncomeDelta,
		"expense_delta_percent": v.ExpenseDelta,
	})
}

func (h *Handler) SpendingChanges(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	from, to, ok := h.bounds(c)
	if !ok {
		return
	}
	prevFrom := c.Query("compare_from")
	prevTo := c.Query("compare_to")
	if prevFrom == "" || prevTo == "" {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"compare_from": "compare_from and compare_to are required"}))
		return
	}
	rows, err := h.svc.SpendingChanges(c.Request.Context(), x.WorkspaceID, from, to, prevFrom, prevTo)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		if r.Current == 0 && r.Previous == 0 {
			continue
		}
		out = append(out, gin.H{
			"category_id":    r.CategoryID,
			"category_name":  r.CategoryName,
			"current":        money.FormatMinorUnits(r.Current),
			"previous":       money.FormatMinorUnits(r.Previous),
			"delta":          money.FormatMinorUnits(r.Delta),
			"delta_percent":  r.DeltaPercent,
		})
	}
	httpx.Success(c, http.StatusOK, out)
}

func (h *Handler) RecurringExpenses(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	v, err := h.svc.RecurringExpenses(c.Request.Context(), x.WorkspaceID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, gin.H{
		"active_rules":      v.ActiveRules,
		"estimated_monthly": money.FormatMinorUnits(v.MonthlyEstimate),
	})
}

func (h *Handler) Dashboard(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	v, err := h.svc.Dashboard(c.Request.Context(), x.WorkspaceID, time.Now().UTC())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func cashflowJSON(v *CashflowView) gin.H {
	return gin.H{
		"income":  money.FormatMinorUnits(v.Income),
		"expense": money.FormatMinorUnits(v.Expense),
		"net":     money.FormatMinorUnits(v.Net),
	}
}

func RegisterRoutes(g *gin.RouterGroup, h *Handler) {
	g.GET("/cashflow", h.Cashflow)
	g.GET("/categories", h.Categories)
	g.GET("/period-comparison", h.PeriodComparison)
	g.GET("/recurring-expenses", h.RecurringExpenses)
	g.GET("/spending-changes", h.SpendingChanges)
}

// RegisterDashboardRoute wires the /dashboard endpoint onto its own group.
func RegisterDashboardRoute(g *gin.RouterGroup, h *Handler) {
	g.GET("/dashboard", h.Dashboard)
}