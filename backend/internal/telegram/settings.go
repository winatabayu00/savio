package telegram

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Settings is the runtime Telegram recap configuration for one bot, stored as a
// singleton row. The bot is bound to the workspace of the OWNER who configured
// it, so messages it receives are written into that workspace.
type Settings struct {
	ID           int       `gorm:"primaryKey" json:"-"`
	Enabled      bool      `gorm:"column:enabled"`
	BotToken     string    `gorm:"column:bot_token;type:text"`
	ChatID       string    `gorm:"column:chat_id;type:text"`
	WorkspaceID  uuid.UUID `gorm:"column:workspace_id;type:uuid"`
	LastUpdateID int64     `gorm:"column:last_update_id"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
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
