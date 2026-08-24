package scenarios

import (
	"net/http"
	"strconv"
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

type scenarioReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     *int64 `json:"version"`
}

type modReq struct {
	Type      string `json:"type"`
	Amount    string `json:"amount"`
	Frequency string `json:"frequency"`
	Narrative string `json:"narrative"`
	Version   *int64 `json:"version"`
}

func (h *Handler) List(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	rows, err := h.svc.List(c.Request.Context(), x.WorkspaceID)
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
	v, err := h.svc.Get(c.Request.Context(), x.WorkspaceID, id)
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
	var req scenarioReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	v, err := h.svc.Create(c.Request.Context(), x.WorkspaceID, x.UserID, &CreateInput{Name: req.Name, Description: req.Description})
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
	var req scenarioReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	in := &UpdateInput{ID: id, Name: req.Name, Description: req.Description, Version: ver(req.Version)}
	v, err := h.svc.Update(c.Request.Context(), x.WorkspaceID, x.UserID, in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) Delete(c *gin.Context) {
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
	if err := h.svc.Delete(c.Request.Context(), x.WorkspaceID, x.UserID, id); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, gin.H{})
}

func (h *Handler) Calculate(c *gin.Context) {
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
	horizon := 90
	if raw := c.Query("horizon"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || !forecastHorizons[v] {
			httpx.Fail(c, errs.ValidationFields(map[string]string{"horizon": "horizon must be 30, 60, 90, 180 or 365"}))
			return
		}
		horizon = v
	}
	v, err := h.svc.Calculate(c.Request.Context(), x.WorkspaceID, x.UserID, id, horizon, time.Now().UTC())
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) Snapshots(c *gin.Context) {
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
	sn, err := h.svc.repo.FindScenario(c.Request.Context(), x.WorkspaceID, id)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var baseline map[string]any
	_ = unmarshal(sn.BaselineSnapshot, &baseline)
	if baseline == nil {
		baseline = map[string]any{}
	}
	httpx.Success(c, http.StatusOK, gin.H{
		"baseline": baseline,
		"is_stale": sn.IsStale,
	})
}

func (h *Handler) AddModification(c *gin.Context) {
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
	in, ok := parseMod(c)
	if !ok {
		return
	}
	v, err := h.svc.AddModification(c.Request.Context(), x.WorkspaceID, x.UserID, id, in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusCreated, v)
}

func (h *Handler) UpdateModification(c *gin.Context) {
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
	modID, err := httpx.ParseUUID(c, "modId")
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	in, ok := parseMod(c)
	if !ok {
		return
	}
	v, err := h.svc.UpdateModification(c.Request.Context(), x.WorkspaceID, x.UserID, id, modID, in, ver(in.VersionPtr))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) DeleteModification(c *gin.Context) {
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
	modID, err := httpx.ParseUUID(c, "modId")
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if err := h.svc.RemoveModification(c.Request.Context(), x.WorkspaceID, x.UserID, id, modID); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, gin.H{})
}

func parseMod(c *gin.Context) (*ModInput, bool) {
	var req modReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return nil, false
	}
	amount, err := money.ParseMinorUnits(req.Amount)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"amount": "amount must be a valid decimal"}))
		return nil, false
	}
	return &ModInput{Type: req.Type, Amount: amount, Frequency: req.Frequency, Narrative: req.Narrative, VersionPtr: req.Version}, true
}

func ver(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

var forecastHorizons = map[int]bool{30: true, 60: true, 90: true, 180: true, 365: true}

func RegisterRoutes(g *gin.RouterGroup, h *Handler, writeOnly gin.HandlerFunc) {
	g.GET("", h.List)
	g.POST("", writeOnly, h.Create)
	g.GET("/:id", h.Get)
	g.PATCH("/:id", writeOnly, h.Update)
	g.DELETE("/:id", writeOnly, h.Delete)
	g.POST("/:id/modifications", writeOnly, h.AddModification)
	g.PATCH("/:id/modifications/:modId", writeOnly, h.UpdateModification)
	g.DELETE("/:id/modifications/:modId", writeOnly, h.DeleteModification)
	g.POST("/:id/calculate", writeOnly, h.Calculate)
	g.GET("/:id/snapshots", h.Snapshots)
}
