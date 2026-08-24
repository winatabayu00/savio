package goals

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/errs"
)

const (
	StatusActive   = "ACTIVE"
	StatusPaused   = "PAUSED"
	StatusAchieved = "ACHIEVED"
	StatusCancelled = "CANCELLED"
)

type Goal struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	WorkspaceID     uuid.UUID  `gorm:"type:uuid;not null"`
	Name            string     `gorm:"size:120;not null"`
	TargetAmount    int64      `gorm:"not null"`
	CurrentAmount   int64      `gorm:"not null;default:0"`
	TargetDate      *time.Time `gorm:"type:date"`
	Priority        string     `gorm:"size:10;not null;default:MEDIUM"`
	LinkedAccountID *uuid.UUID `gorm:"type:uuid"`
	Status          string     `gorm:"size:20;not null;default:ACTIVE"`
	Version         int64      `gorm:"not null;default:1"`
	CreatedByUserID *uuid.UUID
	CreatedAt       time.Time `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt       time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (Goal) TableName() string { return "goals" }

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, g *Goal) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *Repository) FindByID(ctx context.Context, workspaceID, id uuid.UUID) (*Goal, error) {
	var g Goal
	err := r.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", id, workspaceID).First(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("Goal not found")
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID, status string) ([]Goal, error) {
	q := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var rows []Goal
	if err := q.Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) Update(ctx context.Context, g *Goal) error {
	res := r.db.WithContext(ctx).Model(&Goal{}).
		Where("id = ? AND workspace_id = ? AND version = ?", g.ID, g.WorkspaceID, g.Version).
		Updates(map[string]any{
			"name":            g.Name,
			"target_amount":   g.TargetAmount,
			"current_amount":  g.CurrentAmount,
			"target_date":     g.TargetDate,
			"priority":        g.Priority,
			"linked_account_id": g.LinkedAccountID,
			"version":         gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.VersionConflict("This goal was changed. Reload the latest version.")
	}
	return nil
}

func (r *Repository) SetStatus(ctx context.Context, workspaceID, id uuid.UUID, status string, version int64) error {
	res := r.db.WithContext(ctx).Model(&Goal{}).
		Where("id = ? AND workspace_id = ? AND version = ?", id, workspaceID, version).
		Updates(map[string]any{"status": status, "version": gorm.Expr("version + 1")})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.VersionConflict("This goal was changed. Reload the latest version.")
	}
	return nil
}

// AverageMonthlyNet estimates free cashflow from the last 90 days of posted
// activity (income − expense). Deterministic given the ledger.
func (r *Repository) AverageMonthlyNet(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	var net int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(CASE WHEN t.type = 'INCOME' THEN t.amount ELSE -t.amount END), 0) * 30 / 90
		FROM transactions t
		WHERE t.workspace_id = $1 AND t.status = 'POSTED'
			AND t.type IN ('INCOME', 'EXPENSE')
			AND t.transaction_date >= NOW() - INTERVAL '90 days'`,
		workspaceID).Scan(&net).Error
	return net, err
}