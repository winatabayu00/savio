package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/auth/csrf"
	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/config"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/httpx"
	"github.com/savio/savio/backend/internal/users"
	"github.com/savio/savio/backend/internal/workspaces"
)

// AuthRequired verifies the access cookie and loads the authenticated context
// (user + active workspace + role) from fresh database state. The client can
// never select authorization context: claims/user are server-resolved.
func AuthRequired(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(AccessCookieName)
		if err != nil || token == "" {
			httpx.Fail(c, errs.Unauthenticated("Authentication required"))
			c.Abort()
			return
		}
		claims, err := ParseAccessToken(cfg.JWTSecret, token)
		if err != nil {
			httpx.Fail(c, errs.Unauthenticated("Session expired"))
			c.Abort()
			return
		}
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			httpx.Fail(c, errs.Unauthenticated("Invalid session"))
			c.Abort()
			return
		}

		userRepo := users.NewRepository(db)
		user, userErr := userRepo.FindByID(c.Request.Context(), userID)
		if userErr != nil || user.Status != "ACTIVE" {
			httpx.Fail(c, errs.Unauthenticated("Account unavailable"))
			c.Abort()
			return
		}

		// A revoked session must not keep minting access (immediate logout).
		sessRepo := NewSessionRepository(db)
		sess, sessErr := sessRepo.FindByID(c.Request.Context(), claims.SessionID)
		if sessErr != nil || sess.IsRevoked() || sess.UserID != userID || sess.IsExpired(time.Now()) {
			httpx.Fail(c, errs.Unauthenticated("Session was ended"))
			c.Abort()
			return
		}

		wsRepo := workspaces.NewRepository(db)
		var membership *workspaces.Membership
		requested := c.GetHeader("X-Workspace-ID")
		if requested != "" {
			wsID, parseErr := uuid.Parse(requested)
			if parseErr != nil {
				httpx.Fail(c, errs.ValidationFields(map[string]string{"X-Workspace-ID": "Invalid workspace id"}))
				c.Abort()
				return
			}
			membership, _ = wsRepo.FindMembership(c.Request.Context(), wsID, userID)
		}
		if membership == nil {
			if m, mErr := wsRepo.FindDefaultByUser(c.Request.Context(), userID); mErr == nil {
				membership = m
			}
		}
		if membership == nil || membership.Status != "ACTIVE" {
			httpx.Fail(c, errs.ResourceAccessDenied())
			c.Abort()
			return
		}

		authctx.Set(c, &authctx.Ctx{
			UserID:          userID,
			WorkspaceID:     membership.WorkspaceID,
			WorkspaceRole:   authctx.Role(membership.Role),
			SessionID:       claims.SessionID,
			IsAuthenticated: true,
		})
		c.Next()
	}
}

// CSRFRequired protects state-changing requests via the signed double-submit
// pattern. Login/register also receive the CSRF cookie from GET /auth/csrf.
func CSRFRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}
		cookieVal, err := c.Cookie(CSRFCookieName)
		if err != nil {
			cookieVal = ""
		}
		headerVal := c.GetHeader("X-CSRF-Token")
		if err := csrf.Validate(secret, cookieVal, headerVal); err != nil {
			httpx.Fail(c, errs.CSRFTokenInvalid("CSRF validation failed"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireWrite denies VIEWER mutations from the handler level (backend enforced).
func RequireWrite() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, err := authctx.Get(c)
		if err != nil {
			httpx.Fail(c, err)
			c.Abort()
			return
		}
		if !ctx.CanWrite() {
			httpx.Fail(c, errs.PermissionDenied("Your role does not allow this action"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireOwner restricts a route to workspace OWNERs.
func RequireOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, err := authctx.Get(c)
		if err != nil {
			httpx.Fail(c, err)
			c.Abort()
			return
		}
		if !ctx.CanManageMembers() {
			httpx.Fail(c, errs.PermissionDenied("Only the workspace owner can do this"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// clientIP is a helper to extract a safe client ip.
func clientIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	return c.ClientIP()
}
