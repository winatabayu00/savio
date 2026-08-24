package recurring

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/savio/savio/backend/internal/platform/errs"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, rt *RecurringTransaction) error {
	return r.db.WithContext(ctx).Create(rt).Error
}

func (r *Repository) FindByID(ctx context.Context, workspaceID, id uuid.UUID) (*RecurringTransaction, error) {
	var rt RecurringTransaction
	err := r.db.WithContext(ctx).
		Table("recurring_transactions t").
		Joins("LEFT JOIN accounts a ON a.id = t.account_id").
		Joins("LEFT JOIN categories c ON c.id = t.category_id").
		Select("t.*, a.name AS account_name, c.name AS category_name").
		Where("t.id = ? AND t.workspace_id = ?", id, workspaceID).
		First(&rt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("Recurring transaction not found")
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID) ([]RecurringTransaction, int64, error) {
	q := r.db.WithContext(ctx).Table("recurring_transactions t").
		Where("t.workspace_id = ?", workspaceID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []RecurringTransaction
	err := q.Session(&gorm.Session{}).
		Joins("LEFT JOIN accounts a ON a.id = t.account_id").
		Joins("LEFT JOIN categories c ON c.id = t.category_id").
		Select("t.*, a.name AS account_name, c.name AS category_name").
		Order("t.created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) Update(ctx context.Context, t *RecurringTransaction) error {
	res := r.db.WithContext(ctx).Model(&RecurringTransaction{}).
		Where("id = ? AND workspace_id = ? AND version = ?", t.ID, t.WorkspaceID, t.Version).
		Updates(map[string]any{
			"account_id":  t.AccountID,
			"category_id": t.CategoryID,
			"type":        t.Type,
			"amount":      t.Amount,
			"frequency":   t.Frequency,
			"start_date":  t.StartDate,
			"end_date":    t.EndDate,
			"description": t.Description,
			"merchant":    t.Merchant,
			"notes":       t.Notes,
			"auto_post":   t.AutoPost,
			"version":     gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.VersionConflict("This recurring transaction was changed. Reload the latest version.")
	}
	return nil
}

func (r *Repository) SetStatus(ctx context.Context, workspaceID, id uuid.UUID, status string, version int64) error {
	res := r.db.WithContext(ctx).Model(&RecurringTransaction{}).
		Where("id = ? AND workspace_id = ? AND version = ?", id, workspaceID, version).
		Updates(map[string]any{"status": status, "version": gorm.Expr("version + 1")})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.VersionConflict("This recurring transaction was changed. Reload the latest version.")
	}
	return nil
}

// UpsertOccurrences inserts generated occurrences, skipping dates already
// present for the rule (idempotent; INV-010 relies on the unique constraint).
func (r *Repository) UpsertOccurrences(ctx context.Context, occs []RecurringOccurrence) error {
	if len(occs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "recurring_id"}, {Name: "due_date"}},
		DoNothing: true,
	}).Create(&occs).Error
}

func (r *Repository) ListOccurrences(ctx context.Context, workspaceID, recurringID uuid.UUID, status string, from, to string, page, limit, offset int) ([]RecurringOccurrence, int64, error) {
	q := r.db.WithContext(ctx).Table("recurring_occurrences o").
		Where("o.workspace_id = ?", workspaceID)
	if recurringID != uuid.Nil {
		q = q.Where("o.recurring_id = ?", recurringID)
	}
	if status != "" {
		q = q.Where("o.status = ?", status)
	}
	if from != "" {
		q = q.Where("o.due_date >= ?", from)
	}
	if to != "" {
		q = q.Where("o.due_date <= ?", to)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []RecurringOccurrence
	err := q.Session(&gorm.Session{}).
		Joins("JOIN recurring_transactions rt ON rt.id = o.recurring_id").
		Joins("JOIN accounts a ON a.id = rt.account_id").
		Select("o.*, rt.type AS recurring_type, rt.amount AS recurring_amount, a.name AS recurring_account").
		Order("o.due_date ASC").
		Limit(limit).
		Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) GetOccurrence(ctx context.Context, workspaceID, id uuid.UUID) (*RecurringOccurrence, error) {
	var o RecurringOccurrence
	err := r.db.WithContext(ctx).
		Table("recurring_occurrences o").
		Joins("JOIN recurring_transactions rt ON rt.id = o.recurring_id").
		Joins("JOIN accounts a ON a.id = rt.account_id").
		Select("o.*, rt.type AS recurring_type, rt.amount AS recurring_amount, a.name AS recurring_account").
		Where("o.id = ? AND o.workspace_id = ?", id, workspaceID).
		First(&o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("Occurrence not found")
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// LockOccurrenceForUpdate locks an occurrence row and returns it if it
// belongs to the workspace. Used by confirmation to serialize concurrent
// confirms on the same occurrence.
func (r *Repository) LockOccurrenceForUpdate(tx *gorm.DB, workspaceID, id uuid.UUID) (*RecurringOccurrence, error) {
	var o RecurringOccurrence
	err := tx.WithContext(context.Background()).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND workspace_id = ?", id, workspaceID).
		First(&o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("Occurrence not found")
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *Repository) MarkOccurrence(tx *gorm.DB, id uuid.UUID, status string, postedTxID *uuid.UUID) error {
	return tx.WithContext(context.Background()).Model(&RecurringOccurrence{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":                status,
			"posted_transaction_id": postedTxID,
		}).Error
}

func (r *Repository) OccurrenceCount(ctx context.Context, recurringID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&RecurringOccurrence{}).
		Where("recurring_id = ? AND status = ?", recurringID, string(OccConfirmed)).Count(&n).Error
	return n, err
}

func (r *Repository) Rule(ctx context.Context, id uuid.UUID) (*RecurringTransaction, error) {
	var rt RecurringTransaction
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}
