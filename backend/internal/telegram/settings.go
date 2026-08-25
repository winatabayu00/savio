package telegram

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/savio/savio/backend/internal/platform/errs"
)

// Settings is the runtime Telegram recap configuration for one bot, stored as a
// singleton row. The bot is bound to the workspace of the OWNER who configured
// it, so messages it receives are written into that workspace.
type Settings struct {
	ID            int       `gorm:"primaryKey" json:"-"`
	Enabled       bool      `gorm:"column:enabled"`
	BotToken      string    `gorm:"column:bot_token;type:text"`
	ChatID        string    `gorm:"column:chat_id;type:text"`
	WorkspaceID   uuid.UUID `gorm:"column:workspace_id;type:uuid"`
	LastUpdateID  int64     `gorm:"column:last_update_id"`
	WebhookURL    string    `gorm:"column:webhook_url;type:text"`
	WebhookSecret string    `gorm:"column:webhook_secret;type:text"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (Settings) TableName() string { return "telegram_settings" }

func (s *Service) load(ctx context.Context) (*Settings, error) {
	var st Settings
	if err := s.db.WithContext(ctx).Where("id = 1").First(&st).Error; err != nil {
		return nil, err
	}
	return &st, nil
}

type UpdateInput struct {
	Enabled  *bool
	BotToken *string
	ChatID   *string
}

// UpdateSettings persists partial configuration. An empty bot token leaves the
// stored token untouched (masked round-trip safety); WorkspaceID is bound to the
// workspace of the configuring user, never taken from the client.
func (s *Service) UpdateSettings(ctx context.Context, workspaceID uuid.UUID, in *UpdateInput) (*Settings, error) {
	st, err := s.load(ctx)
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
	st.WorkspaceID = workspaceID
	st.UpdatedAt = time.Now().UTC()
	if err := s.db.WithContext(ctx).Save(st).Error; err != nil {
		return nil, err
	}
	return st, nil
}

// Settings returns the current configuration for the Settings page.
func (s *Service) Settings(ctx context.Context) (*Settings, error) {
	return s.load(ctx)
}

// WebhookPath is the unauthenticated Telegram webhook route suffix mounted
// outside the CSRF-protected group (Telegram has no session cookies).
const WebhookPath = "telegram/webhook/:secret"

// fullWebhookURL is the exact URL Telegram pushes updates to; the secret lives
// inside the path so the endpoint needs no auth and no CSRF.
func (s *Settings) fullWebhookURL() string {
	if s.WebhookURL == "" || s.WebhookSecret == "" {
		return ""
	}
	return strings.TrimRight(s.WebhookURL, "/") + "/telegram/webhook/" + s.WebhookSecret
}

func generateSecret() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return uuid.NewString() + uuid.NewString()
	}
	return hex.EncodeToString(b)
}

// RegisterWebhook (re)registers the bot webhook with Telegram at publicURL, or
// removes it when publicURL is empty. publicURL is the user's public HTTPS base
// (typically a tunnel to this API); the real endpoint path is derived here and
// the secret is generated once and stored.
func (s *Service) RegisterWebhook(ctx context.Context, workspaceID uuid.UUID, publicURL string) (*Settings, error) {
	st, err := s.load(ctx)
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
		if err := s.db.WithContext(ctx).Save(st).Error; err != nil {
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
	full := strings.TrimRight(u.String(), "/") + "/telegram/webhook/" + st.WebhookSecret
	if err := client.setWebhook(ctx, full); err != nil {
		return nil, errs.WrapInternal(err, "register telegram webhook")
	}
	st.WebhookURL = strings.TrimRight(u.String(), "/")
	st.WorkspaceID = workspaceID
	st.UpdatedAt = time.Now().UTC()
	if err := s.db.WithContext(ctx).Save(st).Error; err != nil {
		return nil, err
	}
	return st, nil
}
