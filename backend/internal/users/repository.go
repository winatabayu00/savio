package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/errs"
)

// Repository owns users and user_settings persistence.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, u *User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

// FindByEmail matches on lower(email) to align with the unique index.
func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).Where("LOWER(email) = LOWER(?)", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var u User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateSettings(ctx context.Context, s *UserSettings) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *Repository) GetSettings(ctx context.Context, userID uuid.UUID) (*UserSettings, error) {
	var s UserSettings
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("settings not found")
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) UpdateSettings(ctx context.Context, s *UserSettings) error {
	return r.db.WithContext(ctx).Model(&UserSettings{}).Where("user_id = ?", s.UserID).Updates(map[string]any{
		"ai_insights_enabled":      s.AIInsightsEnabled,
		"ai_copilot_enabled":       s.AICopilotEnabled,
		"notifications_enabled":    s.NotificationsEnabled,
		"budget_warning_threshold": s.BudgetWarningThreshold,
		"low_balance_threshold":    s.LowBalanceThreshold,
		"updated_at":               gorm.Expr("NOW()"),
	}).Error
}
