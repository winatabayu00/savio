package server

import (
	"github.com/gin-gonic/gin"

	"github.com/savio/savio/backend/internal/accounts"
	"github.com/savio/savio/backend/internal/analytics"
	"github.com/savio/savio/backend/internal/auth"
	"github.com/savio/savio/backend/internal/categories"
	"github.com/savio/savio/backend/internal/recurring"
	"github.com/savio/savio/backend/internal/transactions"
	"github.com/savio/savio/backend/internal/transfers"
	"github.com/savio/savio/backend/internal/workspaces"
)

// registerModules wires every installed feature module onto the router.
// This file grows as milestones add modules; each module exposes a
// RegisterRoutes-style function bound to the App's dependencies.
func registerModules(a *App) {
	api := a.Engine.Group("/api/v1")
	api.Use(auth.CSRFRequired(a.Config.CSRFSecret))

	registerAuthRoutes(api, a)
	registerWorkspaceRoutes(api, a)
	registerAccountRoutes(api, a)
	registerCategoryRoutes(api, a)
	registerTransactionRoutes(api, a)
	registerTransferRoutes(api, a)
	registerRecurringRoutes(api, a)
	registerAnalyticsRoutes(api, a)
}

func registerTransactionRoutes(api *gin.RouterGroup, a *App) {
	h := transactions.NewHandler(transactions.NewService(a.DB))
	g := api.Group("/transactions")
	g.Use(auth.AuthRequired(a.DB, a.Config))
	transactions.RegisterRoutes(g, h, auth.RequireWrite())
}

func registerWorkspaceRoutes(api *gin.RouterGroup, a *App) {
	h := workspaces.NewHandler(workspaces.NewService(a.DB))
	g := api.Group("/workspaces")
	g.Use(auth.AuthRequired(a.DB, a.Config))
	workspaces.RegisterRoutes(g, h, auth.RequireOwner())
}

func registerAccountRoutes(api *gin.RouterGroup, a *App) {
	h := accounts.NewHandler(accounts.NewService(a.DB)).WithTransactions(transactions.NewService(a.DB))
	g := api.Group("/accounts")
	g.Use(auth.AuthRequired(a.DB, a.Config))
	accounts.RegisterRoutes(g, h, auth.RequireWrite())
}

func registerTransferRoutes(api *gin.RouterGroup, a *App) {
	h := transfers.NewHandler(transfers.NewService(a.DB))
	g := api.Group("/transfers")
	g.Use(auth.AuthRequired(a.DB, a.Config))
	transfers.RegisterRoutes(g, h, auth.RequireWrite())
}

func registerAnalyticsRoutes(api *gin.RouterGroup, a *App) {
	h := analytics.NewHandler(analytics.NewService(a.DB))
	g := api.Group("/analytics")
	g.Use(auth.AuthRequired(a.DB, a.Config))
	analytics.RegisterRoutes(g, h)
	dash := api.Group("")
	dash.Use(auth.AuthRequired(a.DB, a.Config))
	analytics.RegisterDashboardRoute(dash, h)
}

func registerRecurringRoutes(api *gin.RouterGroup, a *App) {
	h := recurring.NewHandler(recurring.NewService(a.DB))
	g := api.Group("/recurring-transactions")
	g.Use(auth.AuthRequired(a.DB, a.Config))
	occs := api.Group("/recurring-occurrences")
	occs.Use(auth.AuthRequired(a.DB, a.Config))
	recurring.RegisterRoutes(g, h, occs, auth.RequireWrite())
}

func registerCategoryRoutes(api *gin.RouterGroup, a *App) {
	h := categories.NewHandler(categories.NewService(a.DB))
	g := api.Group("/categories")
	g.Use(auth.AuthRequired(a.DB, a.Config))
	categories.RegisterRoutes(g, h, auth.RequireWrite())
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
