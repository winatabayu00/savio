package transfers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/errs"
)

type ListFilter struct {
	DateFrom string
	DateTo   string
	Page     int
	Limit    int
	Offset   int
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, tx *gorm.DB, t *Transfer) error {
	return tx.WithContext(ctx).Create(t).Error
}

func (r *Repository) FindByIDTx(ctx context.Context, tx *gorm.DB, workspaceID, id uuid.UUID) (*Transfer, error) {
	var t Transfer
	err := tx.WithContext(ctx).
		Table("account_transfers t").
		Joins("LEFT JOIN accounts fa ON fa.id = t.from_account_id").
		Joins("LEFT JOIN accounts ta ON ta.id = t.to_account_id").
		Select("t.*, fa.name AS from_account_name, ta.name AS to_account_name").
		Where("t.id = ? AND t.workspace_id = ?", id, workspaceID).
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("Transfer not found")
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID, f ListFilter) ([]Transfer, int64, error) {
	base := r.db.WithContext(ctx).Table("account_transfers t").Where("t.workspace_id = ?", workspaceID)
	if f.DateFrom != "" {
		base = base.Where("t.transfer_date >= ?", f.DateFrom)
	}
	if f.DateTo != "" {
		base = base.Where("t.transfer_date <= ?", f.DateTo)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []Transfer
	err := base.Session(&gorm.Session{}).
		Joins("LEFT JOIN accounts fa ON fa.id = t.from_account_id").
		Joins("LEFT JOIN accounts ta ON ta.id = t.to_account_id").
		Select("t.*, fa.name AS from_account_name, ta.name AS to_account_name").
		Order("t.transfer_date DESC, t.created_at DESC").
		Limit(f.Limit).
		Offset(f.Offset).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *Repository) Void(ctx context.Context, tx *gorm.DB, t *Transfer) error {
	res := tx.WithContext(ctx).Model(&Transfer{}).
		Where("id = ? AND workspace_id = ? AND status = 'POSTED' AND version = ?", t.ID, t.WorkspaceID, t.Version).
		Updates(map[string]any{
			"status":            string(StatusVoided),
			"void_reason":       t.VoidReason,
			"voided_by_user_id": t.VoidedByUserID,
			"voided_at":         t.VoidedAt,
			"version":           gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.VersionConflict("This transfer was changed. Reload the latest version.")
	}
	return nil
}
