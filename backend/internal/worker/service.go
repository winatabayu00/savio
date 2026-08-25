package worker

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/accounts"
	"github.com/savio/savio/backend/internal/budgets"
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

// ownerRow is a workspace OWNER with their notification-relevant settings.
type ownerRow struct {
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	BudgetPct   float64
	LowBalance  *int64
}

// owners loads ACTIVE OWNER memberships joined to their settings so sweeps
// notify the right users with the right thresholds.
func owners(ctx context.Context, db *gorm.DB) ([]ownerRow, error) {
	var rows []ownerRow
	err := db.WithContext(ctx).Raw(`
		SELECT m.workspace_id, u.id AS user_id,
		       COALESCE(us.budget_warning_threshold, 80) AS budget_pct,
		       us.low_balance_threshold AS low_balance
		FROM workspace_memberships m
		JOIN users u ON u.id = m.user_id
		LEFT JOIN user_settings us ON us.user_id = u.id
		WHERE m.status = 'ACTIVE' AND m.role = 'OWNER'`).Scan(&rows).Error
	return rows, err
}

func sweepLowBalance(ctx context.Context, db *gorm.DB, now time.Time) {
	balances, err := accounts.NewRepository(db).ActiveBalanceMap(ctx)
	if err != nil {
		slog.Warn("worker: low balance sweep", "error", err)
		return
	}
	type acct struct {
		ID          uuid.UUID
		WorkspaceID uuid.UUID
		Name        string
	}
	var accts []acct
	if err := db.WithContext(ctx).Raw(`
		SELECT id, workspace_id, COALESCE(name, '') AS name
		FROM accounts WHERE status = 'ACTIVE'`).Scan(&accts).Error; err != nil {
		slog.Warn("worker: low balance sweep accounts", "error", err)
		return
	}
	owners, err := owners(ctx, db)
	if err != nil {
		slog.Warn("worker: low balance sweep owners", "error", err)
		return
	}
	nowStr := now.Format("2006-01-02")
	for _, a := range accts {
		bal, ok := balances[a.ID]
		if !ok {
			continue
		}
		for _, o := range owners {
			if o.WorkspaceID != a.WorkspaceID || o.LowBalance == nil {
				continue
			}
			if bal <= *o.LowBalance {
				notify(ctx, db, a.WorkspaceID, o.UserID, "LOW_BALANCE", "Account balance is low",
					"Your account balance is at or below your low-balance threshold.", nowStr)
			}
		}
	}
}

func sweepBudgetWarnings(ctx context.Context, db *gorm.DB, now time.Time) {
	due, err := budgets.NewRepository(db).ActiveDue(ctx, now)
	if err != nil {
		slog.Warn("worker: budget sweep", "error", err)
		return
	}
	owners, err := owners(ctx, db)
	if err != nil {
		slog.Warn("worker: budget sweep owners", "error", err)
		return
	}
	nowStr := now.Format("2006-01-02")
	for _, b := range due {
		for _, o := range owners {
			if o.WorkspaceID != b.WorkspaceID {
				continue
			}
			pct := 0.0
			if b.Amount > 0 {
				pct = (float64(b.Spent) / float64(b.Amount)) * 100
			}
			if pct < o.BudgetPct {
				continue
			}
			typ := "BUDGET_WARNING"
			if pct >= 100 {
				typ = "BUDGET_EXCEEDED"
			}
			notify(ctx, db, b.WorkspaceID, o.UserID, typ,
				"Budget "+b.CategoryName+" needs attention",
				"Spending is at "+strconv.FormatFloat(pct, 'f', 1, 64)+"% of the "+b.CategoryName+" budget.", nowStr)
		}
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
