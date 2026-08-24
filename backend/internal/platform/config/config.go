package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	AppEnv          string
	AppPort         int
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	CSRFSecret      string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	FrontendOrigin  string
	CookieSecure    bool

	AIEnabled     bool
	AIProvider    string
	AIBaseURL     string
	AIAPIKey      string
	AIModel       string
	AITimeout     time.Duration
	AIToolTimeout time.Duration

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
}

// Load reads configuration from the environment. Missing critical values cause
// a fail-fast error. Optional AI configuration is only required when AI_ENABLED=true.
func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:          get("APP_ENV", "development"),
		AppPort:         getInt("APP_PORT", 8080),
		DatabaseURL:     get("DATABASE_URL", "postgres://savio:savio@localhost:5433/savio?sslmode=disable"),
		RedisURL:        get("REDIS_URL", "redis://localhost:6380/0"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		CSRFSecret:      os.Getenv("CSRF_SECRET"),
		AccessTokenTTL:  getDur("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: getDur("REFRESH_TOKEN_TTL", 720*time.Hour),
		FrontendOrigin:  get("FRONTEND_ORIGIN", "http://localhost:5173"),
		CookieSecure:    get("APP_ENV", "development") == "production",

		AIEnabled:     getBool("AI_ENABLED", false),
		AIProvider:    get("AI_PROVIDER", "mock"),
		AIBaseURL:     os.Getenv("AI_BASE_URL"),
		AIAPIKey:      os.Getenv("AI_API_KEY"),
		AIModel:       get("AI_MODEL", "gpt-4o-mini"),
		AITimeout:     getDur("AI_TIMEOUT_SECONDS", 20) * time.Second,
		AIToolTimeout: 10 * time.Second,

		MinioEndpoint:  get("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey: os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:    get("MINIO_BUCKET", "savio-files"),
	}

	var missing []string
	for _, kv := range []struct {
		name string
		val  string
	}{{"JWT_SECRET", cfg.JWTSecret}, {"CSRF_SECRET", cfg.CSRFSecret}} {
		if kv.val == "" || len(kv.val) < 16 {
			slog.Warn("weak or missing secret, using random ephemeral value (sessions will invalidate on restart)",
				"name", kv.name)
		}
	}
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.AIEnabled {
		if cfg.AIProvider != "mock" && (cfg.AIBaseURL == "" || cfg.AIAPIKey == "") {
			slog.Warn("AI_ENABLED=true but provider is not mock and credentials are missing; AI features will degrade",
				"provider", cfg.AIProvider)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing critical configuration: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getDur(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
