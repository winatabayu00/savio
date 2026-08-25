package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/savio/savio/backend/internal/auth/csrf"
	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/config"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/httpx"
	"github.com/savio/savio/backend/internal/platform/ratelimit"
	"github.com/savio/savio/backend/internal/users"
	"github.com/savio/savio/backend/internal/workspaces"
)

// Handler exposes the authentication routes.
type Handler struct {
	svc        *Service
	cfg        *config.Config
	authLim    *ratelimit.Limiter
	refreshLim *ratelimit.Limiter
}

func NewHandler(svc *Service, cfg *config.Config) *Handler {
	return &Handler{
		svc:        svc,
		cfg:        cfg,
		authLim:    ratelimit.New(15, time.Minute),
		refreshLim: ratelimit.New(60, time.Minute),
	}
}

type registerReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Timezone        string `json:"timezone"`
	DefaultCurrency string `json:"default_currency"`
}

type authResponse struct {
	User      userResponse      `json:"user"`
	Workspace workspaceResponse `json:"workspace"`
	Role      string            `json:"role"`
}

type workspaceResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	BaseCurrency string `json:"base_currency,omitempty"`
	Timezone     string `json:"timezone,omitempty"`
}

type settingsResponse struct {
	AIInsightsEnabled      bool    `json:"ai_insights_enabled"`
	AICopilotEnabled       bool    `json:"ai_copilot_enabled"`
	NotificationsEnabled   bool    `json:"notifications_enabled"`
	BudgetWarningThreshold float64 `json:"budget_warning_threshold"`
	LowBalanceThreshold    *int64  `json:"low_balance_threshold"`
}

func toSettingsResponse(s *users.UserSettings) *settingsResponse {
	return &settingsResponse{
		AIInsightsEnabled:      s.AIInsightsEnabled,
		AICopilotEnabled:       s.AICopilotEnabled,
		NotificationsEnabled:   s.NotificationsEnabled,
		BudgetWarningThreshold: s.BudgetWarningThreshold,
		LowBalanceThreshold:    s.LowBalanceThreshold,
	}
}

func (h *Handler) GetCSRF(c *gin.Context) {
	token, err := csrf.Generate(h.cfg.CSRFSecret)
	if err != nil {
		httpx.Fail(c, errs.Internal(err))
		return
	}
	SetCSRFCookie(c, h.cfg, token)
	httpx.Success(c, http.StatusOK, gin.H{"csrf_token": token})
}

func (h *Handler) Register(c *gin.Context) {
	if !h.authLim.Allow("register:"+clientIP(c), time.Now()) {
		httpx.Fail(c, errs.RateLimited("Too many registration attempts. Try again later."))
		return
	}
	var req registerReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	res, err := h.svc.Register(c.Request.Context(), req.Name, req.Email, req.Password,
		c.Request.UserAgent(), clientIP(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	SetAuthCookies(c, h.cfg, res)
	httpx.Success(c, http.StatusCreated, toAuthResponse(res))
}

func (h *Handler) Login(c *gin.Context) {
	if !h.authLim.Allow("login:"+clientIP(c), time.Now()) {
		httpx.Fail(c, errs.RateLimited("Too many login attempts. Try again later."))
		return
	}
	var req loginReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	res, err := h.svc.Login(c.Request.Context(), req.Email, req.Password,
		c.Request.UserAgent(), clientIP(c))
	if err != nil {
		// generic error response; never leak why credentials were rejected
		httpx.Fail(c, errs.InvalidCredentials())
		return
	}
	SetAuthCookies(c, h.cfg, res)
	httpx.Success(c, http.StatusOK, toAuthResponse(res))
}

func (h *Handler) Refresh(c *gin.Context) {
	if !h.refreshLim.Allow("refresh:"+clientIP(c), time.Now()) {
		httpx.Fail(c, errs.RateLimited("Too many refresh attempts. Try again later."))
		return
	}
	raw, err := c.Cookie(RefreshCookieName)
	if err != nil {
		httpx.Fail(c, errs.RefreshTokenInvalid())
		return
	}
	res, err := h.svc.Refresh(c.Request.Context(), raw, c.Request.UserAgent())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	SetAuthCookies(c, h.cfg, res)
	httpx.Success(c, http.StatusOK, toAuthResponse(res))
}

func (h *Handler) Logout(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		// idempotent logout
		ClearAuthCookies(c)
		httpx.Success(c, http.StatusOK, gin.H{})
		return
	}
	_ = h.svc.Logout(c.Request.Context(), ctx.SessionID)
	ClearAuthCookies(c)
	httpx.Success(c, http.StatusOK, gin.H{})
}

func (h *Handler) LogoutAll(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		ClearAuthCookies(c)
		httpx.Success(c, http.StatusOK, gin.H{})
		return
	}
	if err := h.svc.LogoutAll(c.Request.Context(), ctx.UserID); err != nil {
		httpx.Fail(c, err)
		return
	}
	ClearAuthCookies(c)
	httpx.Success(c, http.StatusOK, gin.H{})
}

func (h *Handler) Me(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	user, err := h.svc.Me(c.Request.Context(), ctx.UserID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	settings, err := h.svc.Settings(c.Request.Context(), ctx.UserID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	ws, err := h.svc.workspace.FindWorkspaceByID(c.Request.Context(), ctx.WorkspaceID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	sessions, _ := h.svc.CountActiveSessions(c.Request.Context(), ctx.UserID)

	httpx.Success(c, http.StatusOK, gin.H{
		"user":          toUserResponse(user),
		"workspace":     toWorkspaceResponse(ws),
		"role":          string(ctx.WorkspaceRole),
		"settings":      toSettingsResponse(settings),
		"session_count": sessions,
	})
}

func toUserResponse(u *users.User) userResponse {
	return userResponse{
		ID:              u.ID.String(),
		Name:            u.Name,
		Email:           u.Email,
		Timezone:        u.Timezone,
		DefaultCurrency: u.DefaultCurrency,
	}
}

func toWorkspaceResponse(w *workspaces.Workspace) workspaceResponse {
	return workspaceResponse{
		ID:           w.ID.String(),
		Name:         w.Name,
		BaseCurrency: w.BaseCurrency,
		Timezone:     w.Timezone,
	}
}

func toAuthResponse(res *RegResult) *authResponse {
	return &authResponse{
		User: userResponse{
			ID:    res.UserID.String(),
			Name:  res.Name,
			Email: res.Email,
		},
		Workspace: workspaceResponse{
			ID:   res.WorkspaceID.String(),
			Name: res.WorkspaceName,
		},
		Role: res.Role,
	}
}

// ListSessions returns the user's active sessions (GET /api/v1/sessions).
func (h *Handler) ListSessions(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	rows, err := h.svc.sessions.ListByUser(c.Request.Context(), ctx.UserID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		current := r.ID == ctx.SessionID
		status := "ACTIVE"
		if r.RevokedAt != nil {
			status = "REVOKED"
		}
		out = append(out, gin.H{
			"id":           r.ID.String(),
			"device_name":  r.DeviceName,
			"user_agent":   r.UserAgent,
			"ip_address":   r.IPAddress,
			"created_at":   r.CreatedAt,
			"last_used_at": r.LastUsedAt,
			"expires_at":   r.ExpiresAt,
			"status":       status,
			"current":      current,
		})
	}
	httpx.Success(c, http.StatusOK, out)
}

// DeleteSession revokes one session (DELETE /api/v1/sessions/:id).
func (h *Handler) DeleteSession(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	sid, err := httpx.ParseUUID(c, "id")
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	sess, err := h.svc.sessions.FindByID(c.Request.Context(), sid)
	if err != nil {
		httpx.Fail(c, errs.NotFound("session not found"))
		return
	}
	if sess.UserID != ctx.UserID {
		httpx.Fail(c, errs.NotFound("session not found"))
		return
	}
	if err := h.svc.sessions.Revoke(c.Request.Context(), sid, time.Now()); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, gin.H{})
}

// DeleteAllSessions revokes every session except the current one.
func (h *Handler) DeleteAllSessions(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	rows, err := h.svc.sessions.ListByUser(c.Request.Context(), ctx.UserID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	for _, r := range rows {
		if r.ID == ctx.SessionID {
			continue
		}
		_ = h.svc.sessions.Revoke(c.Request.Context(), r.ID, time.Now())
	}
	httpx.Success(c, http.StatusOK, gin.H{})
}
