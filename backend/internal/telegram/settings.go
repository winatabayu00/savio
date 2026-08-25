package telegram

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/savio/savio/backend/internal/platform/errs"
)

// Settings is the Telegram recap configuration for one workspace. Every
// workspace owns its own row (workspace_id PK), so one workspace can never
// block another from configuring its own bot (AGENTS #25).
type Settings struct {
	WorkspaceID   uuid.UUID `gorm:"primaryKey;column:workspace_id;type:uuid"`
	Enabled       bool      `gorm:"column:enabled"`
	BotToken      string    `gorm:"column:bot_token;type:text"`
	ChatID        string    `gorm:"column:chat_id;type:text"`
	LastUpdateID  int64     `gorm:"column:last_update_id"`
	WebhookURL    string    `gorm:"column:webhook_url;type:text"`
	WebhookSecret string    `gorm:"column:webhook_secret;type:text"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (Settings) TableName() string { return "telegram_settings" }

// settingsForWorkspace returns the workspace's own config row, creating an
// ephemeral default (not yet persisted) when none exists.
func (s *Service) settingsForWorkspace(ctx context.Context, workspaceID uuid.UUID) (*Settings, error) {
	var st Settings
	err := s.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).First(&st).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &Settings{WorkspaceID: workspaceID}, nil
	}
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// byWebhookSecret resolves the workspace owning a bot from its webhook secret,
// used by the unauthenticated push endpoint.
func (s *Service) byWebhookSecret(ctx context.Context, secret string) (*Settings, error) {
	var st Settings
	if err := s.db.WithContext(ctx).Where("webhook_secret = ?", secret).First(&st).Error; err != nil {
		return nil, err
	}
	return &st, nil
}

type UpdateInput struct {
	Enabled  *bool
	BotToken *string
	ChatID   *string
}

// UpdateSettings upserts partial configuration for the given workspace. An
// empty bot token leaves the stored token untouched (masked round-trip safety).
func (s *Service) UpdateSettings(ctx context.Context, workspaceID uuid.UUID, in *UpdateInput) (*Settings, error) {
	st, err := s.settingsForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if in.Enabled != nil {
		st.Enabled = *in.Enabled
	}
	if in.BotToken != nil && strings.TrimSpace(*in.BotToken) != "" {
		st.BotToken = strings.TrimSpace(*in.BotToken)
	}
	if in.ChatID != nil {
		st.ChatID = strings.TrimSpace(*in.ChatID)
	}
	st.UpdatedAt = time.Now().UTC()
	if err := s.upsert(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}

// Settings returns the workspace's own configuration for the Settings page.
func (s *Service) Settings(ctx context.Context, workspaceID uuid.UUID) (*Settings, error) {
	return s.settingsForWorkspace(ctx, workspaceID)
}

// upsert writes the workspace's settings row (create or update).
func (s *Service) upsert(ctx context.Context, st *Settings) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "workspace_id"}},
		UpdateAll: true,
	}).Create(st).Error
}

// WebhookPath is the unauthenticated Telegram webhook route suffix mounted
// outside the CSRF-protected group (Telegram has no session cookies).
const WebhookPath = "telegram/webhook/:secret"

// webhookURLFor is the exact URL Telegram pushes updates to. The secret lives
// inside the path so the endpoint needs no auth and no CSRF.
func webhookURLFor(base, secret string) string {
	if base == "" || secret == "" {
		return ""
	}
	return strings.TrimRight(base, "/") + "/api/v1/telegram/webhook/" + secret
}

func generateSecret() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return uuid.NewString() + uuid.NewString()
	}
	return hex.EncodeToString(b)
}

// RegisterWebhook (re)registers the workspace's bot webhook with Telegram at
// publicURL, or removes it when publicURL is empty. publicURL is the public
// HTTPS base (typically a tunnel to this API); the real endpoint path is
// derived here and the secret is generated once and stored.
func (s *Service) RegisterWebhook(ctx context.Context, workspaceID uuid.UUID, publicURL string) (*Settings, error) {
	st, err := s.settingsForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if st.BotToken == "" {
		return nil, errs.Validation("simpan bot token dahulu sebelum mendaftarkan webhook")
	}
	client := newBotClient(st.BotToken)
	if strings.TrimSpace(publicURL) == "" {
		if err := client.deleteWebhook(ctx); err != nil {
			return nil, errs.WrapInternal(err, "delete telegram webhook")
		}
		st.WebhookURL = ""
		st.WebhookSecret = ""
		st.UpdatedAt = time.Now().UTC()
		if err := s.upsert(ctx, st); err != nil {
			return nil, err
		}
		return st, nil
	}
	u, err := url.Parse(publicURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return nil, errs.Validation("webhook URL harus https:// publik, contoh https://xxxx.ngrok-free.app")
	}
	if st.WebhookSecret == "" {
		st.WebhookSecret = generateSecret()
	}
	full := webhookURLFor(u.String(), st.WebhookSecret)
	if err := client.setWebhook(ctx, full, st.WebhookSecret); err != nil {
		return nil, errs.WrapInternal(err, "register telegram webhook")
	}
	st.WebhookURL = strings.TrimRight(u.String(), "/")
	st.UpdatedAt = time.Now().UTC()
	if err := s.upsert(ctx, st); err != nil {
		return nil, err
	}
	return st, nil
}