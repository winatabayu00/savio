package ai

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/config"
)

// Settings is the runtime AI provider configuration, stored as a singleton row.
// On first startup it is seeded from environment defaults; afterwards the values
// entered on the Settings page win (they replace the AI_* env vars).
type Settings struct {
	ID             int       `gorm:"primaryKey" json:"-"`
	Enabled        bool      `gorm:"column:enabled"`
	Provider       string    `gorm:"column:provider;type:varchar(30)"`
	BaseURL        string    `gorm:"column:base_url;type:text"`
	APIKey         string    `gorm:"column:api_key;type:text"`
	Model          string    `gorm:"column:model;type:varchar(120)"`
	TimeoutSeconds int       `gorm:"column:timeout_seconds"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (Settings) TableName() string { return "ai_settings" }

// seedSettings populates the singleton row from environment defaults when it has
// not been configured yet (provider = 'pending'). Never overwrites user edits.
func seedSettings(db *gorm.DB, cfg *config.Config) error {
	var s Settings
	err := db.Where("id = 1").First(&s).Error
	if gorm.ErrRecordNotFound == err {
		return db.Exec(
			`INSERT INTO ai_settings (id, enabled, provider, base_url, api_key, model, timeout_seconds, updated_at)
			 VALUES (1, ?, ?, ?, ?, ?, ?, NOW())
			 ON CONFLICT (id) DO NOTHING`,
			cfg.AIEnabled, cfg.AIProvider, cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel,
			int(cfg.AITimeout/time.Second)).Error
	}
	if err != nil {
		return err
	}
	if s.Provider == "pending" {
		return db.Model(&Settings{ID: 1}).Updates(map[string]any{
			"enabled":         cfg.AIEnabled,
			"provider":        cfg.AIProvider,
			"base_url":        cfg.AIBaseURL,
			"api_key":         cfg.AIAPIKey,
			"model":           cfg.AIModel,
			"timeout_seconds": int(cfg.AITimeout / time.Second),
			"updated_at":      time.Now().UTC(),
		}).Error
	}
	return nil
}

func (s *Service) loadSettings(ctx context.Context) (*Settings, error) {
	var st Settings
	if err := s.db.WithContext(ctx).Where("id = 1").First(&st).Error; err != nil {
		return nil, err
	}
	return &st, nil
}

// settingsAdapter adapts the stored Settings row to the platform ai.Config
// interface used to build a provider (AGENTS #76).
type settingsAdapter struct{ st *Settings }

func (a settingsAdapter) AIBaseURL() string { return a.st.BaseURL }
func (a settingsAdapter) AIAPIKey() string  { return a.st.APIKey }
func (a settingsAdapter) AIModel() string   { return a.st.Model }
func (a settingsAdapter) AITimeout() time.Duration {
	return time.Duration(a.st.TimeoutSeconds) * time.Second
}

// Settings returns the runtime AI configuration.
func (s *Service) Settings(ctx context.Context) (*Settings, error) {
	return s.loadSettings(ctx)
}

// UpdateSettingsInput carries optional fields; nil means "keep current value".
type UpdateSettingsInput struct {
	Enabled        *bool
	Provider       *string
	BaseURL        *string
	APIKey         *string
	Model          *string
	TimeoutSeconds *int
}

// UpdateSettings persists partial AI configuration changes and returns the new
// state. An empty APIKey leaves the stored key untouched (so the masked value
// can be round-tripped safely); provider/mock + empty base URL degrades to the
// built-in mock provider.
func (s *Service) UpdateSettings(ctx context.Context, in *UpdateSettingsInput) (*Settings, error) {
	st, err := s.loadSettings(ctx)
	if err != nil {
		return nil, err
	}
	if in.Enabled != nil {
		st.Enabled = *in.Enabled
	}
	if in.Provider != nil {
		st.Provider = strings.TrimSpace(*in.Provider)
	}
	if in.BaseURL != nil {
		st.BaseURL = strings.TrimSpace(*in.BaseURL)
	}
	if in.APIKey != nil && strings.TrimSpace(*in.APIKey) != "" {
		st.APIKey = strings.TrimSpace(*in.APIKey)
	}
	if in.Model != nil {
		st.Model = strings.TrimSpace(*in.Model)
	}
	if in.TimeoutSeconds != nil {
		st.TimeoutSeconds = *in.TimeoutSeconds
	}
	st.UpdatedAt = time.Now().UTC()
	if err := s.db.WithContext(ctx).Save(st).Error; err != nil {
		return nil, err
	}
	return st, nil
}
