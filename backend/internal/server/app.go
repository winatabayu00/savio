package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"

	"github.com/savio/savio/backend/internal/platform/config"
	"github.com/savio/savio/backend/internal/platform/mw"
	"github.com/savio/savio/backend/internal/platform/redisclient"
)

// App is the shared application dependency container for the HTTP API process.
type App struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
	Engine *gin.Engine
}

// New builds the App and its fully wired router. Feature modules register
// their routes through RegisterRoutes hooks.
func New(cfg *config.Config, db *gorm.DB, rdb *redis.Client) *App {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(mw.RequestID(), mw.ErrorDetails(cfg.AppEnv != "production"), mw.Recovery(), mw.Logging(), mw.SecurityHeaders(), mw.CORS(cfg.FrontendOrigin))

	app := &App{Config: cfg, DB: db, Redis: rdb, Engine: engine}
	app.registerCoreRoutes()
	registerModules(app)
	return app
}

func (a *App) registerCoreRoutes() {
	a.Engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "savio-api", "time": time.Now().UTC()})
	})

	a.Engine.GET("/ready", func(c *gin.Context) {
		deps := map[string]string{}

		if err := pingDB(a.DB); err != nil {
			deps["postgresql"] = "down"
		} else {
			deps["postgresql"] = "ok"
		}

		if a.Redis != nil {
			if err := redisclient.Ping(c.Request.Context(), a.Redis); err != nil {
				deps["redis"] = "down"
			} else {
				deps["redis"] = "ok"
			}
		}

		status := http.StatusOK
		if deps["postgresql"] != "ok" {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"status": statusLabel(status), "dependencies": deps})
	})

	a.Engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"code": "RESOURCE_NOT_FOUND"},
			"message": "Route not found",
		})
	})
}

func statusLabel(s int) string {
	if s == http.StatusOK {
		return "ready"
	}
	return "not_ready"
}

// pingDB uses a lightweight conn+Ping rather than a heavy query.
func pingDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		return err
	}
	defer conn.Close()
	return conn.PingContext(context.Background())
}

func (a *App) Server() *http.Server {
	return &http.Server{
		Addr:              ":" + strconv.Itoa(a.Config.AppPort),
		Handler:           a.Engine,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
