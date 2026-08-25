package budgets

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/errs"
)

type Status string

const (
	StatusActive Status = "ACTIVE"
	StatusClosed Status = "CLOSED"
)

// Budget is a monthly spending plan for one expense category.
type Budget struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	WorkspaceID     uuid.UUID `gorm:"type:uuid;not null"`
	CategoryID      uuid.UUID `gorm:"type:uuid;not null"`
	Amount          int64     `gorm:"not null"`
	PeriodStart     time.Time `gorm:"type:date;not null"`
	PeriodEnd       time.Time `gorm:"type:date;not null"`
	Status          string    `gorm:"size:20;not null;default:ACTIVE"`
	Version         int64     `gorm:"not null;default:1"`
	CreatedByUserID *uuid.UUID
	CreatedAt       time.Time `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt       time.Time `gorm:"type:timestamptz;not null;default:now()"`

	CategoryName string `gorm:"->;-:migration"`
}

func (Budget) TableName() string { return "budgets" }

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, b *Budget) error {
	return r.db.WithContext(ctx).Omit("CategoryName").Create(b).Error
}

func (r *Repository) FindByID(ctx context.Context, workspaceID, id uuid.UUID) (*Budget, error) {
	var b Budget
	err := r.db.WithContext(ctx).
		Table("budgets b").
		Joins("LEFT JOIN categories c ON c.id = b.category_id").
		Select("b.*, c.name AS category_name").
		Where("b.id = ? AND b.workspace_id = ?", id, workspaceID).
		First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("Budget not found")
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID, status string) ([]Budget, error) {
	q := r.db.WithContext(ctx).Table("budgets b").
		Joins("LEFT JOIN categories c ON c.id = b.category_id").
		Select("b.*, c.name AS category_name").
		Where("b.workspace_id = ?", workspaceID)
	if status != "" {
		q = q.Where("b.status = ?", status)
	}
	var rows []Budget
	if err := q.Order("b.period_start DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) Update(ctx context.Context, b *Budget) error {
	res := r.db.WithContext(ctx).Model(&Budget{}).
		Where("id = ? AND workspace_id = ? AND status = 'ACTIVE' AND version = ?", b.ID, b.WorkspaceID, b.Version).
		Updates(map[string]any{
			"category_id":  b.CategoryID,
			"amount":       b.Amount,
			"period_start": b.PeriodStart,
			"period_end":   b.PeriodEnd,
			"version":      gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.VersionConflict("This budget was changed. Reload the latest version.")
	}
	return nil
}

func (r *Repository) Close(ctx context.Context, workspaceID, id uuid.UUID, version int64) error {
	res := r.db.WithContext(ctx).Model(&Budget{}).
		Where("id = ? AND workspace_id = ? AND status = 'ACTIVE' AND version = ?", id, workspaceID, version).
		Updates(map[string]any{"status": string(StatusClosed), "version": gorm.Expr("version + 1")})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.VersionConflict("This budget was changed. Reload the latest version.")
	}
	return nil
}

// spent returns the POSTED EXPENSE total for the category within the period
// (excludes transfers, voided and adjustments — AGENTS #49).
func (r *Repository) spent(ctx context.Context, workspaceID, categoryID uuid.UUID, from, to time.Time) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(t.amount), 0) FROM transactions t
		WHERE t.workspace_id = $1 AND t.status = 'POSTED'
			AND t.type = 'EXPENSE' AND t.category_id = $2
			AND t.transaction_date BETWEEN $3 AND $4`,
		workspaceID, categoryID, from.Format("2006-01-02"), to.Format("2006-01-02")).
		Scan(&total).Error
	return total, err
}

// BudgetDue is an active in-period budget with its workspace-and-period-scoped
// posted EXPENSE spend, used by the worker notification sweep (AGENTS #102:
// the worker reuses repository logic instead of re-deriving it).
type BudgetDue struct {
	ID           uuid.UUID
	WorkspaceID  uuid.UUID
	CategoryID   uuid.UUID
	CategoryName string
	Amount       int64
	Spent        int64
}

// ActiveDue lists active budgets whose period includes now, with spend scoped
// to the budget's own workspace and period (fix: the previous worker sweep
// summed all-time, cross-workspace spend on shared system categories).
func (r *Repository) ActiveDue(ctx context.Context, now time.Time) ([]BudgetDue, error) {
	var rows []BudgetDue
	err := r.db.WithContext(ctx).Raw(`
		SELECT b.id, b.workspace_id, b.category_id, COALESCE(c.name, '') AS category_name, b.amount,
		       COALESCE((SELECT SUM(t.amount) FROM transactions t
		           WHERE t.workspace_id = b.workspace_id AND t.status = 'POSTED'
		             AND t.type = 'EXPENSE' AND t.category_id = b.category_id
		             AND t.transaction_date BETWEEN b.period_start AND b.period_end), 0) AS spent
		FROM budgets b
		LEFT JOIN categories c ON c.id = b.category_id
		WHERE b.status = 'ACTIVE' AND b.period_start <= $1 AND b.period_end >= $1`,
		now.Format("2006-01-02")).Scan(&rows).Error
	return rows, err
}

// ActiveForCategoryWithin checks for an overlapping active budget on the same
// category (calendar-overlap, not just exact match).
func (r *Repository) OverlappingActive(ctx context.Context, workspaceID, categoryID uuid.UUID, from, to time.Time, excludeID uuid.UUID) (bool, error) {
	var n int64
	q := r.db.WithContext(ctx).Model(&Budget{}).
		Where("workspace_id = ? AND category_id = ? AND status = 'ACTIVE' AND period_start <= ? AND period_end >= ?",
			workspaceID, categoryID, to.Format("2006-01-02"), from.Format("2006-01-02"))
	if excludeID != uuid.Nil {
		q = q.Where("id <> ?", excludeID)
	}
	if err := q.Count(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
}
