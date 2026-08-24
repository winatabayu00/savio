package transfers

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
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Amount        string `json:"amount"`
	TransferDate  string `json:"transfer_date"`
	Description   string `json:"description"`
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
	filter := ListFilter{DateFrom: c.Query("from"), DateTo: c.Query("to"), Page: pg.Page, Limit: pg.Limit, Offset: pg.Offset()}
	if filter.DateFrom != "" {
		if _, err := time.Parse("2006-01-02", filter.DateFrom); err != nil {
			httpx.Fail(c, errs.ValidationFields(map[string]string{"from": "date must use YYYY-MM-DD"}))
			return
		}
	}
	if filter.DateTo != "" {
		if _, err := time.Parse("2006-01-02", filter.DateTo); err != nil {
			httpx.Fail(c, errs.ValidationFields(map[string]string{"to": "date must use YYYY-MM-DD"}))
			return
		}
	}
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
	fromID, err := uuid.Parse(req.FromAccountID)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"from_account_id": "invalid account id"}))
		return
	}
	toID, err := uuid.Parse(req.ToAccountID)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"to_account_id": "invalid account id"}))
		return
	}
	amount, err := money.ParseMinorUnits(req.Amount)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"amount": "amount must be a valid decimal"}))
		return
	}
	tdate, err := time.Parse("2006-01-02", req.TransferDate)
	if err != nil {
		httpx.Fail(c, errs.ValidationFields(map[string]string{"transfer_date": "date must use YYYY-MM-DD"}))
		return
	}
	v, err := h.svc.Create(c.Request.Context(), ctx.WorkspaceID, ctx.UserID, &CreateInput{
		FromAccountID: fromID,
		ToAccountID:   toID,
		AmountMinor:   amount,
		TransferDate:  time.Date(tdate.Year(), tdate.Month(), tdate.Day(), 0, 0, 0, 0, time.UTC),
		Description:   req.Description,
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusCreated, v)
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
		Version: req.versionOrZero(),
		Reason:  req.Reason,
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, v)
}

func (r *voidReq) versionOrZero() int64 {
	if r == nil || r.Version == nil {
		return 0
	}
	return *r.Version
}

func RegisterRoutes(g *gin.RouterGroup, h *Handler, writeOnly gin.HandlerFunc) {
	g.GET("", h.List)
	g.POST("", writeOnly, h.Create)
	g.GET("/:id", h.Get)
	g.POST("/:id/void", writeOnly, h.Void)
}
