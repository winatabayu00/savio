package authctx

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/savio/savio/backend/internal/platform/errs"
)

// ContextKey is the gin context key for the authenticated request context.
const ContextKey = "authctx"

// Role is a workspace membership role.
type Role string

const (
	RoleOwner  Role = "OWNER"
	RoleMember Role = "MEMBER"
	RoleViewer Role = "VIEWER"
)

// Ctx is the authenticated request context injected by auth middleware and
// refreshed with workspace scope. It is always built server-side from DB
// state; the client can never choose it.
type Ctx struct {
	UserID          uuid.UUID
	WorkspaceID     uuid.UUID
	WorkspaceRole   Role
	SessionID       uuid.UUID
	IsAuthenticated bool
}

// CanWrite reports whether the role may mutate financial state.
func (c *Ctx) CanWrite() bool {
	return c.WorkspaceRole == RoleOwner || c.WorkspaceRole == RoleMember
}

// CanManageMembers reports whether the role may manage workspace memberships.
func (c *Ctx) CanManageMembers() bool {
	return c.WorkspaceRole == RoleOwner
}

// Set stores the context on the gin request.
func Set(c *gin.Context, ctx *Ctx) { c.Set(ContextKey, ctx) }

// Get retrieves the authenticated context. Returns error when unauthenticated.
func Get(c *gin.Context) (*Ctx, error) {
	v, ok := c.Get(ContextKey)
	if !ok {
		return nil, errs.Unauthenticated("Authentication required")
	}
	ctx, ok := v.(*Ctx)
	if !ok || ctx == nil || !ctx.IsAuthenticated {
		return nil, errs.Unauthenticated("Authentication required")
	}
	if ctx.UserID == uuid.Nil {
		return nil, errors.New("authctx: missing user_id")
	}
	return ctx, nil
}

// MustGet returns the context; panics only on programmer error (handler misuse).
func MustGet(c *gin.Context) *Ctx {
	ctx, err := Get(c)
	if err != nil {
		panic(err)
	}
	return ctx
}
