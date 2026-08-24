package accounts

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/errs"
)

// Repository owns accounts persistence. Every lookup is workspace-scoped so
// resource IDs alone are never treated as authorization (INV-019).
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, a *Account) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *Repository) FindByID(ctx context.Context, workspaceID, id uuid.UUID) (*Account, error) {
	var a Account
	err := r.db.WithContext(ctx).
		Where("id = ? AND workspace_id = ?", id, workspaceID).
		First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("Account not found")
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID, status string, page, limit, offset int) ([]Account, int64, error) {
	q := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Model(&Account{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []Account
	if err := q.Order("created_at ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// Update persists editable fields with optimistic versioning (INV via 409
// version conflict on stale submissions).
func (r *Repository) Update(ctx context.Context, a *Account) error {
	res := r.db.WithContext(ctx).Model(&Account{}).
		Where("id = ? AND workspace_id = ? AND version = ?", a.ID, a.WorkspaceID, a.Version).
		Updates(map[string]any{
			"name":             a.Name,
			"type":             a.Type,
			"institution_name": a.InstitutionName,
			"description":      a.Description,
			"version":          gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.VersionConflict("This account was changed since you last opened it. Reload the latest version.")
	}
	a.Version++
	return nil
}

// UpdateOpeningBalance changes the opening balance only when no ledger
// history exists; its presence is detected via the transactions table.
func (r *Repository) UpdateOpeningBalance(ctx context.Context, workspaceID, id uuid.UUID, balance int64, version int64) error {
	if err := r.ensureNoLedger(ctx, workspaceID, id); err != nil {
		return err
	}
	res := r.db.WithContext(ctx).Model(&Account{}).
		Where("id = ? AND workspace_id = ? AND version = ?", id, workspaceID, version).
		Updates(map[string]any{"opening_balance": balance, "version": gorm.Expr("version + 1")})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.VersionConflict("This account was changed since you last opened it. Reload the latest version.")
	}
	return nil
}

func (r *Repository) SetStatus(ctx context.Context, workspaceID, id uuid.UUID, status string, archivedAt *time.Time) error {
	res := r.db.WithContext(ctx).Model(&Account{}).
		Where("id = ? AND workspace_id = ?", id, workspaceID).
		Updates(map[string]any{"status": status, "archived_at": archivedAt})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.NotFound("Account not found")
	}
	return nil
}

// ensureNoLedger blocks edits that would corrupt history once the ledger
// exists. Before M08 the transactions tables do not exist, so the check is a
// no-op; when they appear it is enforced automatically.
func (r *Repository) ensureNoLedger(ctx context.Context, workspaceID, id uuid.UUID) error {
	// columns differ per ledger table; keep joins out of the count
	checks := []struct {
		table string
		where string
		args  []any
	}{
		{"transactions", "workspace_id = ? AND account_id = ?", []any{workspaceID, id}},
		{"account_transfers", "workspace_id = ? AND (from_account_id = ? OR to_account_id = ?)", []any{workspaceID, id, id}},
	}
	for _, chk := range checks {
		if !r.db.Migrator().HasTable(chk.table) {
			continue
		}
		var n int64
		if err := r.db.WithContext(ctx).Table(chk.table).Where(chk.where, chk.args...).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return errs.BusinessConflict("BUSINESS_CONFLICT", "This account has financial history and cannot be modified this way. Void or adjust instead.")
		}
	}
	return nil
}

// Delete removes an account only when it has no ledger history.
func (r *Repository) Delete(ctx context.Context, workspaceID, id uuid.UUID) error {
	if err := r.ensureNoLedger(ctx, workspaceID, id); err != nil {
		return err
	}
	res := r.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", id, workspaceID).Delete(&Account{}).Limit(1)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.NotFound("Account not found")
	}
	return nil
}

// PostBalanceModifiers computes per-account posted ledger effects keyed by
// account ID. Empty until the ledger tables exist (M08+) and derived balance
// becomes opening + modifiers (INV-005).
func (r *Repository) PostBalanceModifiers(ctx context.Context, workspaceID uuid.UUID) (map[uuid.UUID]int64, error) {
	mods := map[uuid.UUID]int64{}
	if r.db.Migrator().HasTable("transactions") {
		var rows []struct {
			AccountID uuid.UUID
			Sum       int64
		}
		// EXPENSE reduces the balance; INCOME/ADJUSTMENT increase it. ADJUSTMENT
		// carries a signed amount (negative = reduction).
		if err := r.db.WithContext(ctx).Raw(`
			SELECT account_id, COALESCE(SUM(CASE
				WHEN type = 'EXPENSE' THEN -amount
				ELSE amount
			END), 0) AS sum
			FROM transactions
			WHERE workspace_id = ? AND status = 'POSTED'
			GROUP BY account_id`, workspaceID).Scan(&rows).Error; err != nil {
			return nil, err
		}
		for i := range rows {
			mods[rows[i].AccountID] = rows[i].Sum
		}
	}
	if r.db.Migrator().HasTable("account_transfers") {
		var fromes []struct {
			AccountID uuid.UUID
			Sum       int64
		}
		var toes []struct {
			AccountID uuid.UUID
			Sum       int64
		}
		if err := r.db.WithContext(ctx).Raw(
			"SELECT from_account_id AS account_id, COALESCE(SUM(amount), 0) AS sum FROM account_transfers WHERE workspace_id = ? AND status = 'POSTED' GROUP BY from_account_id",
			workspaceID).Scan(&fromes).Error; err != nil {
			return nil, err
		}
		if err := r.db.WithContext(ctx).Raw(
			"SELECT to_account_id AS account_id, COALESCE(SUM(amount), 0) AS sum FROM account_transfers WHERE workspace_id = ? AND status = 'POSTED' GROUP BY to_account_id",
			workspaceID).Scan(&toes).Error; err != nil {
			return nil, err
		}
		for i := range fromes {
			mods[fromes[i].AccountID] -= fromes[i].Sum
		}
		for i := range toes {
			mods[toes[i].AccountID] += toes[i].Sum
		}
	}
	return mods, nil
}
