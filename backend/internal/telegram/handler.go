package telegram

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// configResponse never exposes the raw bot token.
type configResponse struct {
	Enabled      bool   `json:"enabled"`
	BotTokenMask string `json:"bot_token_masked"`
	ChatID       string `json:"chat_id"`
	WorkspaceID  string `json:"workspace_id"`
}

func maskToken(t string) string {
	if t == "" {
		return ""
	}
	if len(t) <= 8 {
		return "••••"
	}
	return "••••••••" + t[len(t)-4:]
}

func toConfigResponse(st *Settings) configResponse {
	return configResponse{
		Enabled:      st.Enabled,
		BotTokenMask: maskToken(st.BotToken),
		ChatID:       st.ChatID,
		WorkspaceID:  st.WorkspaceID.String(),
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
	Enabled  *bool   `json:"enabled"`
	BotToken *string `json:"bot_token"`
	ChatID   *string `json:"chat_id"`
}

func (h *Handler) UpdateConfig(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req updateConfigReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	if err := validateConfig(&req); err != nil {
		httpx.Fail(c, err)
		return
	}
	st, err := h.svc.UpdateSettings(c.Request.Context(), x.WorkspaceID, &UpdateInput{
		Enabled:  req.Enabled,
		BotToken: req.BotToken,
		ChatID:   req.ChatID,
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, toConfigResponse(st))
}

func validateConfig(req *updateConfigReq) error {
	fields := map[string]string{}
	if req.BotToken != nil {
		t := strings.TrimSpace(*req.BotToken)
		if t != "" && !strings.Contains(t, ":") {
			fields["bot_token"] = "bot_token must be the token from @BotFather (format <numeric>:<secret>)"
		}
	}
	if req.ChatID != nil {
		cid := strings.TrimSpace(*req.ChatID)
		if cid != "" && !strings.HasPrefix(cid, "-") && !allDigits(cid) && !strings.HasPrefix(cid, "@") {
			fields["chat_id"] = "chat_id must be a numeric id, a -number for groups, or a @username"
		}
	}
	if len(fields) > 0 {
		return errs.ValidationFields(fields)
	}
	return nil
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
