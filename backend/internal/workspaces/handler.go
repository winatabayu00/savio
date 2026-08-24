package workspaces

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/httpx"
)

// Handler exposes the workspace/membership endpoints.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type memberReq struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func toMemberJSON(m *MemberSummary) gin.H {
	return gin.H{
		"id":         m.ID.String(),
		"user_id":    m.UserID.String(),
		"name":       m.UserName,
		"email":      m.UserEmail,
		"role":       m.Role,
		"created_at": m.CreatedAt,
	}
}

// Current returns the authenticated user's active workspace and role.
func (h *Handler) Current(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	res, err := h.svc.Current(c.Request.Context(), ctx.WorkspaceID, ctx.WorkspaceRole)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, gin.H{
		"workspace": gin.H{
			"id":            res.Workspace.ID.String(),
			"name":          res.Workspace.Name,
			"base_currency": res.Workspace.BaseCurrency,
			"timezone":      res.Workspace.Timezone,
		},
		"role":         string(res.Role),
		"member_count": res.MemberCount,
	})
}

// ListMembers lists the workspace's active members.
func (h *Handler) ListMembers(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	rows, err := h.svc.ListMembers(c.Request.Context(), ctx.WorkspaceID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for i := range rows {
		out = append(out, toMemberJSON(&rows[i]))
	}
	httpx.Success(c, http.StatusOK, out)
}

// AddMember invites an existing user by email (OWNER only, enforced upstream).
func (h *Handler) AddMember(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req memberReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	m, err := h.svc.AddMember(c.Request.Context(), ctx.WorkspaceID, req.Email, req.Role)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusCreated, toMemberJSON(m))
}

// UpdateMember changes a member's role (OWNER only, enforced upstream).
func (h *Handler) UpdateMember(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	memberID, err := httpx.ParseUUID(c, "memberId")
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	var req memberReq
	if err := httpx.Bind(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	if err := h.svc.UpdateRole(c.Request.Context(), ctx.WorkspaceID, memberID, req.Role); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, gin.H{})
}

// RemoveMember removes a member from the workspace (OWNER only, enforced upstream).
func (h *Handler) RemoveMember(c *gin.Context) {
	ctx, err := authctx.Get(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	memberID, err := httpx.ParseUUID(c, "memberId")
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if err := h.svc.RemoveMember(c.Request.Context(), ctx.WorkspaceID, memberID); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Success(c, http.StatusOK, gin.H{})
}

// RegisterRoutes wires the workspace endpoints. ownerOnly is the role
// middleware (AuthRequired + RequireOwner); passed in to avoid an
// auth↔workspaces import cycle.
func RegisterRoutes(g *gin.RouterGroup, h *Handler, ownerOnly gin.HandlerFunc) {
	g.GET("/current", h.Current)
	g.GET("/current/members", h.ListMembers)
	g.POST("/current/members", ownerOnly, h.AddMember)
	g.PATCH("/current/members/:memberId", ownerOnly, h.UpdateMember)
	g.DELETE("/current/members/:memberId", ownerOnly, h.RemoveMember)
}
