package ai

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/savio/savio/backend/internal/forecast"
	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/httpx"
	"github.com/savio/savio/backend/internal/platform/ratelimit"
)

type Handler struct {
	svc         *Service
	categorizeLim *ratelimit.Limiter
	insightLim    *ratelimit.Limiter
	copilotLim    *ratelimit.Limiter
}

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc:           svc,
		categorizeLim: ratelimit.New(20, time.Minute),
		insightLim:    ratelimit.New(10, time.Minute),
		copilotLim:    ratelimit.New(10, time.Minute),
	}
}

func (h *Handler) Status(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	enabled := h.svc.Enabled()
	state := "enabled"
	if !enabled {
		state = "disabled"
	}
	httpx.Success(c, http.StatusOK, gin.H{
		"enabled": enabled,
		"state":   state,
		"workspace_id": x.WorkspaceID.String(),
	})
}

type categorizeReq struct {
	Description string `json:"description"`
	Merchant    string `json:"merchant"`
}

func (h *Handler) Categorize(c *gin.Context) {
	if !h.categorizeLim.Allow("categorize:"+c.ClientIP(), time.Now()) {
		httpx.Fail(c, errs.RateLimited("Too many categorization requests. Try again shortly."))
		return
	}
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req categorizeReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	if len(req.Description) < 2 && len(req.Merchant) < 2 {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"description": "A description or merchant is required"}))
		return
	}
	res, err := h.svc.Categorize(c.Request.Context(), x.WorkspaceID, req.Description, req.Merchant)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, res)
}

type insightReq struct {
	From    string `json:"from"`
	To      string `json:"to"`
	PrevFrom string `json:"compare_from"`
	PrevTo  string `json:"compare_to"`
}

func (h *Handler) Insight(c *gin.Context) {
	if !h.insightLim.Allow("insight:"+c.ClientIP(), time.Now()) {
		httpx.Fail(c, errs.RateLimited("Too many insight requests. Try again shortly."))
		return
	}
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req insightReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	if req.From == "" || req.To == "" {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"from": "from and to are required (YYYY-MM-DD)"}))
		return
	}
	if req.PrevFrom == "" {
		req.PrevFrom = req.From
	}
	if req.PrevTo == "" {
		req.PrevTo = req.To
	}
	res, err := h.svc.Insight(c.Request.Context(), x.WorkspaceID, req.From, req.To, req.PrevFrom, req.PrevTo)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, res)
}

type copilotReq struct {
	Question string `json:"question"`
	Horizon  int    `json:"horizon"`
}

func (h *Handler) Copilot(c *gin.Context) {
	if !h.copilotLim.Allow("copilot:"+c.ClientIP(), time.Now()) {
		httpx.Fail(c, errs.RateLimited("Too many Copilot requests. Try again shortly."))
		return
	}
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req copilotReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	if req.Horizon <= 0 {
		req.Horizon = 90
	}
	if !forecast.AllowedHorizons[req.Horizon] {
		req.Horizon = 90
	}
	res, err := h.svc.Copilot(c.Request.Context(), x.WorkspaceID, req.Question, req.Horizon, time.Now().UTC())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, res)
}

func RegisterRoutes(g *gin.RouterGroup, h *Handler) {
	g.GET("/status", h.Status)
	g.POST("/categorize", h.Categorize)
	g.POST("/insight", h.Insight)
	g.POST("/copilot", h.Copilot)
}