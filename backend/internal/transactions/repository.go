package transactions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/errs"
)

// ListFilter carries validated list query options.
type ListFilter struct {
	Search     string
	Type       string
	AccountID  uuid.UUID
	CategoryID uuid.UUID
	Status     string
	DateFrom   string // YYYY-MM-DD
	DateTo     string // YYYY-MM-DD
	Page       int
	Limit      int
	Offset     int
	SortField  string
	SortDesc   bool
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, tx *Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *Repository) FindByID(ctx context.Context, workspaceID, id uuid.UUID) (*Transaction, error) {
	var t Transaction
	err := r.db.WithContext(ctx).
		Table("transactions t").
		Joins("LEFT JOIN accounts a ON a.id = t.account_id").
		Joins("LEFT JOIN categories c ON c.id = t.category_id").
		Joins("LEFT JOIN users u ON u.id = t.created_by_user_id").
		Select("t.*, a.name AS account_name, c.name AS category_name, c.type AS category_type, u.name AS created_by_name").
		Where("t.id = ? AND t.workspace_id = ?", id, workspaceID).
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("Transaction not found")
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID, f ListFilter) ([]Transaction, int64, error) {
	base := r.db.WithContext(ctx).Table("transactions t").Where("t.workspace_id = ?", workspaceID)
	if f.Search != "" {
		like := "%" + f.Search + "%"
		base = base.Where("(t.description ILIKE ? OR t.merchant ILIKE ? OR t.notes ILIKE ?)", like, like, like)
	}
	if f.Type != "" {
		base = base.Where("t.type = ?", f.Type)
	}
	if f.AccountID != uuid.Nil {
		base = base.Where("t.account_id = ?", f.AccountID)
	}
	if f.CategoryID != uuid.Nil {
		base = base.Where("t.category_id = ?", f.CategoryID)
	}
	if f.Status != "" {
		base = base.Where("t.status = ?", f.Status)
	}
	if f.DateFrom != "" {
		base = base.Where("t.transaction_date >= ?", f.DateFrom)
	}
	if f.DateTo != "" {
		base = base.Where("t.transaction_date <= ?", f.DateTo)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "t.transaction_date DESC, t.created_at DESC"
	if f.SortField != "" {
		dir := "DESC"
		if !f.SortDesc {
			dir = "ASC"
		}
		order = fmt.Sprintf("%s %s", f.SortField, dir)
	}
	var rows []Transaction
	err := base.Session(&gorm.Session{}).
		Joins("LEFT JOIN accounts a ON a.id = t.account_id").
		Joins("LEFT JOIN categories c ON c.id = t.category_id").
		Joins("LEFT JOIN users u ON u.id = t.created_by_user_id").
		Select("t.*, a.name AS account_name, c.name AS category_name, c.type AS category_type, u.name AS created_by_name").
		Order(order).
		Limit(f.Limit).
		Offset(f.Offset).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// Update persists draft editable fields with optimistic versioning.
func (r *Repository) Update(ctx context.Context, t *Transaction) error {
	res := r.db.WithContext(ctx).Model(&Transaction{}).
		Where("id = ? AND workspace_id = ? AND status = 'DRAFT' AND version = ?", t.ID, t.WorkspaceID, t.Version).
		Updates(map[string]any{
			"amount":           t.Amount,
			"category_id":      t.CategoryID,
			"transaction_date": t.TransactionDate,
			"description":      t.Description,
			"merchant":         t.Merchant,
			"notes":            t.Notes,
			"type":             t.Type,
			"version":          gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.VersionConflict("This transaction was changed since you last opened it. Reload the latest version.")
	}
	return nil
}

// Post transitions a DRAFT transaction to POSTED (takes financial effect).
func (r *Repository) Post(ctx context.Context, workspaceID, id uuid.UUID, version int64, now time.Time) error {
	res := r.db.WithContext(ctx).Model(&Transaction{}).
		Where("id = ? AND workspace_id = ? AND status = 'DRAFT' AND version = ?", id, workspaceID, version).
		Updates(map[string]any{
			"status":    string(StatusPosted),
			"posted_at": now,
			"version":   gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.VersionConflict("This transaction cannot be posted because it changed. Reload the latest version.")
	}
	return nil
}

// Void marks a POSTED transaction as VOIDED (idempotency: only POSTED rows can
// transition, so a double void is a no-op that reports a conflict).
func (r *Repository) Void(ctx context.Context, workspaceID, id uuid.UUID, version int64, reason string, userID uuid.UUID, now time.Time) error {
	res := r.db.WithContext(ctx).Model(&Transaction{}).
		Where("id = ? AND workspace_id = ? AND status = 'POSTED' AND version = ?", id, workspaceID, version).
		Updates(map[string]any{
			"status":            string(StatusVoided),
			"voided_at":         now,
			"voided_by_user_id": userID,
			"void_reason":       nullableStr(reason),
			"version":           gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.VersionConflict("This transaction cannot be voided because it changed. Reload the latest version.")
	}
	return nil
}

func nullableStr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
