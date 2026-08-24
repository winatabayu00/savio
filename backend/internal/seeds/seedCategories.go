package seeds

import (
	"context"
	"log/slog"
	"time"

	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SystemCategory is a globally available category (workspace_id IS NULL).
type SystemCategory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string
	Type        string // INCOME | EXPENSE
	Icon        *string
	Description *string
	IsSystem    bool `gorm:"default:true"`
	Status      string
	CreatedAt   time.Time `gorm:"type:timestamptz"`
	UpdatedAt   time.Time `gorm:"type:timestamptz"`
}

func (SystemCategory) TableName() string { return "categories" }

var systemCategories = []struct {
	Name string
	Type string
	Icon string
}{
	{"Salary", "INCOME", "briefcase"},
	{"Business Income", "INCOME", "store"},
	{"Freelance", "INCOME", "laptop"},
	{"Investment Income", "INCOME", "trending-up"},
	{"Gift", "INCOME", "gift"},
	{"Other Income", "INCOME", "plus-circle"},

	{"Food & Dining", "EXPENSE", "utensils"},
	{"Transportation", "EXPENSE", "car"},
	{"Housing & Rent", "EXPENSE", "home"},
	{"Utilities", "EXPENSE", "zap"},
	{"Groceries", "EXPENSE", "shopping-cart"},
	{"Health", "EXPENSE", "heart"},
	{"Shopping", "EXPENSE", "bag"},
	{"Entertainment", "EXPENSE", "film"},
	{"Travel", "EXPENSE", "plane"},
	{"Education", "EXPENSE", "book-open"},
	{"Subscriptions", "EXPENSE", "repeat"},
	{"Debt & Loans", "EXPENSE", "percent"},
	{"Insurance", "EXPENSE", "shield"},
	{"Personal Care", "EXPENSE", "smile"},
	{"Other Expense", "EXPENSE", "more-horizontal"},
}

// SeedSystemCategories idempotently inserts system categories.
func SeedSystemCategories(ctx context.Context, db *gorm.DB) error {
	for _, c := range systemCategories {
		icon := c.Icon
		var existing SystemCategory
		err := db.WithContext(ctx).
			Where("name = ? AND type = ? AND is_system = TRUE", c.Name, c.Type).
			First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		cat := SystemCategory{
			ID:       uuid.New(),
			Name:     c.Name,
			Type:     c.Type,
			Icon:     &icon,
			IsSystem: true,
			Status:   "ACTIVE",
		}
		if err := db.WithContext(ctx).Create(&cat).Error; err != nil {
			return err
		}
	}
	slog.Info("system categories seeded", "count", len(systemCategories))
	return nil
}
