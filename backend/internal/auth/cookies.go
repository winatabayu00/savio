package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/savio/savio/backend/internal/platform/config"
)

// Cookie names (API contract §4).
const (
	AccessCookieName  = "access_token"
	RefreshCookieName = "refresh_token"
	CSRFCookieName    = "csrf_token"
)

// SetAuthCookies writes the HttpOnly access and refresh cookies.
func SetAuthCookies(c *gin.Context, cfg *config.Config, r *RegResult) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     AccessCookieName,
		Value:    r.AccessToken,
		Path:     "/api",
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(r.AccessTokenTTL.Seconds()),
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    r.RefreshToken,
		Path:     "/api/v1/auth/refresh",
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(r.SessionTTL.Seconds()),
	})
}

// SetCSRFCookie writes the JS-readable CSRF double-submit cookie.
func SetCSRFCookie(c *gin.Context, cfg *config.Config, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24,
	})
}

// ClearAuthCookies expires both auth cookies under every path they were set with.
func ClearAuthCookies(c *gin.Context) {
	paths := map[string][]string{
		AccessCookieName:  {"/", "/api", "/api/v1/auth"},
		RefreshCookieName: {"/", "/api/v1/auth/refresh"},
	}
	for name, ps := range paths {
		for _, p := range ps {
			http.SetCookie(c.Writer, &http.Cookie{
				Name: name, Value: "", Path: p, HttpOnly: true,
				SameSite: http.SameSiteLaxMode, MaxAge: -1,
			})
		}
	}
}
