package goals

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

type goalReq struct {
	Name            string  `json:"name"`
	TargetAmount    string  `json:"target_amount"`
	CurrentAmount   string  `json:"current_amount"`
	TargetDate      *string `json:"target_date"`
	Priority        string  `json:"priority"`
	LinkedAccountID *string `json:"linked_account_id"`
	Version         *int64  `json:"version"`
}

func (h *Handler) List(c *gin.Context) {
	x, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	rows, err := h.svc.List(c.Request.Context(), x.WorkspaceID, c.Query("status"))
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
	var req goalReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	in, ok := parseReq(c, &req)
	if !ok {
		return
	}
	v, err := h.svc.Create(c.Request.Context(), x.WorkspaceID, x.UserID, &CreateInput{
		Name: in.Name, TargetAmount: in.TargetAmount, CurrentAmount: in.CurrentAmount,
		TargetDate: in.TargetDate, Priority: in.Priority, LinkedAccountID: in.LinkedAccountID,
	})
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
	var req goalReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	in, ok := parseReq(c, &req)
	if !ok {
		return
	}
	v, err := h.svc.Update(c.Request.Context(), x.WorkspaceID, x.UserID, &UpdateInput{
		ID: id, Name: in.Name, TargetAmount: in.TargetAmount, CurrentAmount: in.CurrentAmount,
		TargetDate: in.TargetDate, Priority: in.Priority, LinkedAccountID: in.LinkedAccountID, Version: ver(req.Version),
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (h *Handler) Status(c *gin.Context) {
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
	v, err := h.svc.SetStatus(c.Request.Context(), x.WorkspaceID, x.UserID, id, stringsToUpperStatus(c.Param("action")), ver(req.Version))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func stringsToUpperStatus(a string) string {
	switch a {
	case "pause":
		return StatusPaused
	case "resume":
		return StatusActive
	case "achieve":
		return StatusAchieved
	case "cancel":
		return StatusCancelled
	}
	return StatusActive
}

func ver(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func parseReq(c *gin.Context, req *goalReq) (*CreateInput, bool) {
	target, err := money.ParseMinorUnits(req.TargetAmount)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"target_amount": "target_amount must be a valid decimal"}))
		return nil, false
	}
	current, err := money.ParseMinorUnits(req.CurrentAmount)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"current_amount": "current_amount must be a valid decimal"}))
		return nil, false
	}
	var targetDate *time.Time
	if req.TargetDate != nil && *req.TargetDate != "" {
		t, err := time.Parse("2006-01-02", *req.TargetDate)
		if err != nil {
			httpx.Fail(c, errs.ValidationFields(map[string]string{"target_date": "date must use YYYY-MM-DD"}))
			return nil, false
		}
		td := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		targetDate = &td
	}
	var linked *uuid.UUID
	if req.LinkedAccountID != nil && *req.LinkedAccountID != "" {
		id, err := uuid.Parse(*req.LinkedAccountID)
		if err != nil {
			httpx.Fail(c, errs.ValidationFields(map[string]string{"linked_account_id": "invalid account id"}))
			return nil, false
		}
		linked = &id
	}
	return &CreateInput{
		Name: req.Name, TargetAmount: target, CurrentAmount: current,
		TargetDate: targetDate, Priority: req.Priority, LinkedAccountID: linked,
	}, true
}

func RegisterRoutes(g *gin.RouterGroup, h *Handler, writeOnly gin.HandlerFunc) {
	g.GET("", h.List)
	g.POST("", writeOnly, h.Create)
	g.GET("/:id", h.Get)
	g.PATCH("/:id", writeOnly, h.Update)
	g.POST("/:id/:action", writeOnly, h.Status)
}