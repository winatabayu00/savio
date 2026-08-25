package httpx

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/savio/savio/backend/internal/platform/errs"
)

const requestIDKey = "request_id"

// Success renders the standard success envelope.
func Success(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"success": true, "data": data})
}

// Collection renders a paginated collection envelope.
func Collection(c *gin.Context, data any, page, limit, total int) {
	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// Fail renders the standard error envelope. Internal causes stay in logs in production.
func Fail(c *gin.Context, err error) {
	appErr := errs.From(err)
	if appErr.Cause != nil {
		slog.Error("request failed",
			"request_id", c.GetString(requestIDKey),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", appErr.Status,
			"code", appErr.Code,
			"error", appErr.Cause.Error())
	}
	trace := gin.H{
		"trace_id": c.GetString(requestIDKey),
		"endpoint": c.Request.URL.Path,
		"method":   c.Request.Method,
		"status":   appErr.Status,
		"code":     appErr.Code,
	}
	if enabled, _ := c.Get("error_details"); enabled == true && appErr.Cause != nil {
		trace["reason"] = appErr.Cause.Error()
	}
	c.JSON(appErr.Status, gin.H{
		"success": false,
		"error": gin.H{
			"code":    appErr.Code,
			"details": appErr.Details,
		},
		"message":    appErr.Message,
		"request_id": c.GetString(requestIDKey),
		"trace":      trace,
	})
}

// Bind validates and binds a JSON body. Returns parsed request or nil.
func Bind(c *gin.Context, dst any) error {
	if err := c.ShouldBindJSON(dst); err != nil {
		return errs.Validation("Request body is not valid JSON")
	}
	return nil
}

// ParseUUID parses a path parameter as a UUID.
func ParseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errs.InvalidUUID()
	}
	return id, nil
}

// Pagination carries validated page/limit values.
type Pagination struct {
	Page  int
	Limit int
}

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

func ParsePagination(c *gin.Context) Pagination {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(DefaultLimit)))
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return Pagination{Page: page, Limit: limit}
}

func (p Pagination) Offset() int { return (p.Page - 1) * p.Limit }

// SortSpec is a validated sort field/order pair.
type SortSpec struct {
	Field string
	Desc  bool
}

// ParseSort validates order and returns a SortSpec. Allowed fields are
// checked per-endpoint via the allowlist set.
func ParseSort(c *gin.Context, defaultField string, allow map[string]bool) (SortSpec, error) {
	field := c.DefaultQuery("sort", defaultField)
	if !allow[field] {
		return SortSpec{}, errs.ValidationFields(map[string]string{"sort": "Unsupported sort field"})
	}
	order := strings.ToLower(c.DefaultQuery("order", "desc"))
	desc := true
	if order == "asc" {
		desc = false
	} else if order != "desc" {
		return SortSpec{}, errs.ValidationFields(map[string]string{"order": "order must be asc or desc"})
	}
	return SortSpec{Field: field, Desc: desc}, nil
}

// ParseOptionalDate validates an optional date in YYYY-MM-DD format.
func ParseOptionalDate(c *gin.Context, key string) (string, error) {
	raw := c.Query(key)
	if raw == "" {
		return "", nil
	}
	if len(raw) != 10 || raw[4] != '-' || raw[7] != '-' {
		return "", errs.ValidationFields(map[string]string{key: "Date must use YYYY-MM-DD format"})
	}
	if _, err := timeParse(raw); err != nil {
		return "", errs.ValidationFields(map[string]string{key: "Date must be a valid calendar date"})
	}
	return raw, nil
}

func timeParse(s string) (any, error) {
	t := strings.Split(s, "-")
	if len(t) != 3 {
		return nil, errors.New("bad")
	}
	y, err1 := strconv.Atoi(t[0])
	m, err2 := strconv.Atoi(t[1])
	d, err3 := strconv.Atoi(t[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return nil, errors.New("bad")
	}
	if m < 1 || m > 12 || d < 1 || d > 31 {
		return nil, errors.New("bad")
	}
	_ = y
	return nil, nil
}
