package users

import (
	"time"

	"github.com/google/uuid"
)

// User is the Savio identity record.
type User struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name            string    `gorm:"size:120"`
	Email           string    `gorm:"size:255"`
	PasswordHash    string    `gorm:"size:255"`
	Timezone        string    `gorm:"size:100"`
	DefaultCurrency string    `gorm:"type:char(3)"`
	Locale          string    `gorm:"size:20"`
	Status          string    `gorm:"size:20"`
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (User) TableName() string { return "users" }

// UserSettings holds per-user preferences.
type UserSettings struct {
	UserID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	AIInsightsEnabled      bool
	AICopilotEnabled       bool
	NotificationsEnabled   bool
	BudgetWarningThreshold float64 `gorm:"type:numeric(5,2)"`
	LowBalanceThreshold    *int64
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func (UserSettings) TableName() string { return "user_settings" }
