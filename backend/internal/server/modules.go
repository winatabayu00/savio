package server

import (
	"github.com/gin-gonic/gin"

	"github.com/savio/savio/backend/internal/auth"
)

// registerModules wires every installed feature module onto the router.
// This file grows as milestones add modules; each module exposes a
// RegisterRoutes-style function bound to the App's dependencies.
func registerModules(a *App) {
	api := a.Engine.Group("/api/v1")
	api.Use(auth.CSRFRequired(a.Config.CSRFSecret))

	registerAuthRoutes(api, a)
}

func registerAuthRoutes(api *gin.RouterGroup, a *App) {
	h := auth.NewHandler(auth.NewService(a.DB, a.Config), a.Config)
	authed := auth.AuthRequired(a.DB, a.Config)

	g := api.Group("/auth")
	g.GET("/csrf", h.GetCSRF)
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.POST("/refresh", h.Refresh)
	g.POST("/logout", authed, h.Logout)
	g.POST("/logout-all", authed, h.LogoutAll)
	g.GET("/me", authed, h.Me)

	sg := api.Group("/sessions")
	sg.Use(authed)
	sg.GET("", h.ListSessions)
	sg.DELETE("/:id", h.DeleteSession)
	sg.DELETE("", h.DeleteAllSessions)
}
