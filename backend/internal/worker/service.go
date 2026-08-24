package worker

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/recurring"
)

// Service runs non-critical background jobs. It calls shared domain services
// rather than duplicating business logic (AGENTS #102) and every job is
// idempotent (AGENTS #105).
type Service struct {
	db        *gorm.DB
	recurring *recurring.Service
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db, recurring: recurring.NewService(db)}
}

// AutoPostDue uses the shared recurring service (guarded by DB constraints).
func (s *Service) AutoPostDue(ctx context.Context, now time.Time) error {
	posted, err := s.recurring.AutoPostDue(ctx, now)
	if err != nil {
		return err
	}
	if posted > 0 {
		slog.Info("worker: auto-posted recurring occurrences", "count", posted)
	}
	return nil
}

// SweepNotifications creates low-balance and budget-warning notifications.
// Dedup is enforced by a per-workspace-per-type-per-day unique index, so
// concurrent sweeps cannot double-notify.
func (s *Service) SweepNotifications(ctx context.Context, now time.Time) error {
	sweepLowBalance(ctx, s.db, now)
	sweepBudgetWarnings(ctx, s.db, now)
	return nil
}

func sweepLowBalance(ctx context.Context, db *gorm.DB, now time.Time) {
	type row struct {
		WorkspaceID uuid.UUID
		UserID      uuid.UUID
		Name        string
		Balance     int64
	}
	var rows []row
	err := db.WithContext(ctx).Raw(`
		SELECT a.workspace_id, u.id AS user_id,
		       COALESCE(a.name, '') AS name,
		       a.opening_balance
			+ COALESCE((SELECT SUM(CASE WHEN t.type='EXPENSE' THEN -t.amount ELSE t.amount END)
				FROM transactions t WHERE t.workspace_id = a.workspace_id AND t.status='POSTED'), 0) AS balance
		FROM accounts a
		JOIN workspace_memberships m ON m.workspace_id = a.workspace_id AND m.status = 'ACTIVE' AND m.role = 'OWNER'
		JOIN users u ON u.id = m.user_id
		JOIN user_settings us ON us.user_id = u.id
		WHERE a.status = 'ACTIVE' AND us.low_balance_threshold IS NOT NULL
		  AND a.opening_balance
			+ COALESCE((SELECT SUM(CASE WHEN t.type='EXPENSE' THEN -t.amount ELSE t.amount END)
				FROM transactions t WHERE t.workspace_id = a.workspace_id AND t.status='POSTED'), 0) <= us.low_balance_threshold
	`).Scan(&rows).Error
	if err != nil {
		slog.Warn("worker: low balance sweep", "error", err)
		return
	}
	nowStr := now.Format("2006-01-02")
	for _, r := range rows {
		notify(ctx, db, r.WorkspaceID, r.UserID, "LOW_BALANCE", "Account balance is low",
			"Your account balance is at or below your low-balance threshold.", nowStr)
	}
}

func sweepBudgetWarnings(ctx context.Context, db *gorm.DB, now time.Time) {
	type row struct {
		WorkspaceID uuid.UUID
		UserID      uuid.UUID
		Category    string
		Status      string
		Pct         float64
	}
	var rows []row
	err := db.WithContext(ctx).Raw(`
		SELECT b.workspace_id, m.user_id AS user_id, c.name AS category,
		       CASE WHEN b.amount > 0 AND COALESCE(sp.spent,0) * 100.0 / b.amount >= 100 THEN 'EXCEEDED'
		            WHEN b.amount > 0 AND COALESCE(sp.spent,0) * 100.0 / b.amount >= us.budget_warning_threshold THEN 'WARNING'
		            ELSE 'ON_TRACK' END AS status,
		       COALESCE(sp.spent,0) * 100.0 / NULLIF(b.amount,0) AS pct
		FROM budgets b
		JOIN categories c ON c.id = b.category_id
		JOIN workspace_memberships m ON m.workspace_id = b.workspace_id AND m.status='ACTIVE' AND m.role='OWNER'
		JOIN users u ON u.id = m.user_id
		JOIN user_settings us ON us.user_id = u.id
		LEFT JOIN (
			SELECT category_id, SUM(amount) AS spent FROM transactions
			WHERE status='POSTED' AND type='EXPENSE' GROUP BY category_id
		) sp ON sp.category_id = b.category_id
		WHERE b.status='ACTIVE' AND (b.period_start <= CURRENT_DATE AND b.period_end >= CURRENT_DATE)
	`).Scan(&rows).Error
	if err != nil {
		slog.Warn("worker: budget sweep", "error", err)
		return
	}
	nowStr := now.Format("2006-01-02")
	for _, r := range rows {
		if r.Status == "ON_TRACK" {
			continue
		}
		typ := "BUDGET_WARNING"
		if r.Status == "EXCEEDED" {
			typ = "BUDGET_EXCEEDED"
		}
		notify(ctx, db, r.WorkspaceID, r.UserID, typ,
			"Budget "+r.Category+" needs attention",
			"Spending is at "+strconv.FormatFloat(r.Pct, 'f', 1, 64)+"% of the "+r.Category+" budget.", nowStr)
	}
}

func notify(ctx context.Context, db *gorm.DB, wsID, userID uuid.UUID, typ, title, body, dateKey string) {
	if wsID == uuid.Nil || userID == uuid.Nil {
		return
	}
	var n int64
	if err := db.WithContext(ctx).Table("notifications").
		Where("workspace_id = ? AND type = ? AND day = ?", wsID, typ, dateKey).
		Count(&n).Error; err != nil || n > 0 {
		return // already notified today; the unique index is the hard guard
	}
	_ = db.WithContext(ctx).Exec(`
		INSERT INTO notifications (id, workspace_id, user_id, type, title, body, day, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (workspace_id, type, day) DO NOTHING`,
		uuid.New(), wsID, userID, typ, title, body, dateKey)
}
