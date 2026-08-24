package forecast

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func timeNow() time.Time { return time.Now().UTC() }

func (h *Handler) Compute(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	horizon := 90
	if raw := c.Query("horizon"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || !AllowedHorizons[v] {
			httpx.Fail(c, errs.ValidationFields(map[string]string{"horizon": "horizon must be one of 30, 60, 90, 180, 365"}))
			return
		}
		horizon = v
	}
	res, err := h.svc.Compute(c.Request.Context(), x.WorkspaceID, horizon, timeNow())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, res)
}

func RegisterRoutes(g *gin.RouterGroup, h *Handler) {
	g.GET("", h.Compute)
}