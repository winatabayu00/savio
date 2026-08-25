package ai

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/savio/savio/backend/internal/forecast"
	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/httpx"
	"github.com/savio/savio/backend/internal/platform/ratelimit"
)

type Handler struct {
	svc           *Service
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
	enabled, err := h.svc.enabled(c.Request.Context())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	state := "enabled"
	if !enabled {
		state = "disabled"
	}
	httpx.Success(c, http.StatusOK, gin.H{
		"enabled":      enabled,
		"state":        state,
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
	From     string `json:"from"`
	To       string `json:"to"`
	PrevFrom string `json:"compare_from"`
	PrevTo   string `json:"compare_to"`
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
	g.POST("/conversations", h.CreateConversation)
	g.GET("/conversations", h.ListConversations)
	g.GET("/conversations/:id", h.GetConversation)
	g.POST("/conversations/:id/messages", h.SendMessage)
	g.DELETE("/conversations/:id", h.DeleteConversation)
}

func conversationID(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, errs.ValidationFields(map[string]string{"id": "Invalid conversation ID."})
	}
	return id, nil
}

func (h *Handler) CreateConversation(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	row, err := h.svc.CreateConversation(c.Request.Context(), x.WorkspaceID, x.UserID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusCreated, row)
}

func (h *Handler) ListConversations(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	rows, err := h.svc.ListConversations(c.Request.Context(), x.WorkspaceID, x.UserID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if rows == nil {
		rows = []Conversation{}
	}
	httpx.Success(c, http.StatusOK, rows)
}

func (h *Handler) GetConversation(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	id, err := conversationID(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	row, err := h.svc.Conversation(c.Request.Context(), x.WorkspaceID, x.UserID, id)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, row)
}

func (h *Handler) SendMessage(c *gin.Context) {
	if !h.copilotLim.Allow("copilot:"+c.ClientIP(), time.Now()) {
		httpx.Fail(c, errs.RateLimited("Too many Copilot requests. Try again shortly."))
		return
	}
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	id, err := conversationID(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req copilotReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	if req.Horizon <= 0 || !forecast.AllowedHorizons[req.Horizon] {
		req.Horizon = 90
	}
	row, err := h.svc.SendMessage(c.Request.Context(), x.WorkspaceID, x.UserID, id, req.Question, req.Horizon, time.Now().UTC())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, row)
}

func (h *Handler) DeleteConversation(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	id, err := conversationID(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if err := h.svc.DeleteConversation(c.Request.Context(), x.WorkspaceID, x.UserID, id); err != nil {
		httpx.Fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// configResponse never exposes the raw API key.
type configResponse struct {
	Enabled        bool   `json:"enabled"`
	Provider       string `json:"provider"`
	BaseURL        string `json:"base_url"`
	APIKeyMasked   string `json:"api_key_masked"`
	Model          string `json:"model"`
	Persona        string `json:"persona"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

func maskKey(k string) string {
	if k == "" {
		return ""
	}
	if len(k) <= 4 {
		return "••••"
	}
	return "••••" + k[len(k)-4:]
}

func toConfigResponse(st *Settings) configResponse {
	return configResponse{
		Enabled:        st.Enabled,
		Provider:       st.Provider,
		BaseURL:        st.BaseURL,
		APIKeyMasked:   maskKey(st.APIKey),
		Model:          st.Model,
		Persona:        st.Persona,
		TimeoutSeconds: st.TimeoutSeconds,
	}
}

func (h *Handler) GetConfig(c *gin.Context) {
	st, err := h.svc.Settings(c.Request.Context())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, toConfigResponse(st))
}

type updateConfigReq struct {
	Enabled        *bool   `json:"enabled"`
	Provider       *string `json:"provider"`
	BaseURL        *string `json:"base_url"`
	APIKey         *string `json:"api_key"`
	Model          *string `json:"model"`
	Persona        *string `json:"persona"`
	TimeoutSeconds *int    `json:"timeout_seconds"`
}

func (h *Handler) UpdateConfig(c *gin.Context) {
	var req updateConfigReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	if err := validateConfigReq(&req); err != nil {
		httpx.Fail(c, err)
		return
	}
	in := &UpdateSettingsInput{
		Enabled:        req.Enabled,
		Provider:       req.Provider,
		BaseURL:        req.BaseURL,
		APIKey:         req.APIKey,
		Model:          req.Model,
		Persona:        req.Persona,
		TimeoutSeconds: req.TimeoutSeconds,
	}
	st, err := h.svc.UpdateSettings(c.Request.Context(), in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, toConfigResponse(st))
}

func validateConfigReq(req *updateConfigReq) error {
	fields := map[string]string{}
	if req.Provider != nil {
		p := strings.TrimSpace(*req.Provider)
		if p != "openai" && p != "mock" {
			fields["provider"] = "provider must be 'openai' or 'mock'"
		}
	}
	if req.TimeoutSeconds != nil && (*req.TimeoutSeconds < 1 || *req.TimeoutSeconds > 120) {
		fields["timeout_seconds"] = "timeout_seconds must be between 1 and 120"
	}
	if req.BaseURL != nil {
		b := strings.TrimSpace(*req.BaseURL)
		if b != "" {
			u, err := url.ParseRequestURI(b)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
				fields["base_url"] = "base_url must be a valid http(s) URL"
			}
		}
	}
	if req.Model != nil && strings.TrimSpace(*req.Model) == "" {
		fields["model"] = "model is required"
	}
	if req.Persona != nil {
		p := strings.TrimSpace(*req.Persona)
		if p != "balanced" && p != "lenna" {
			fields["persona"] = "persona must be 'balanced' or 'lenna'"
		}
	}
	if len(fields) > 0 {
		return errs.ValidationFields(fields)
	}
	return nil
}
