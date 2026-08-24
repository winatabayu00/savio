package categories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/errs"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, c *Category) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *Repository) FindCustomByID(ctx context.Context, workspaceID, id uuid.UUID) (*Category, error) {
	var c Category
	err := r.db.WithContext(ctx).
		Where("id = ? AND workspace_id = ? AND is_system = FALSE", id, workspaceID).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("Category not found")
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// List returns the union of system categories and the workspace's custom
// active categories, optionally filtered by type.
func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID, typ string, includeArchived bool) ([]Category, error) {
	q := r.db.WithContext(ctx).
		Where("(is_system = TRUE OR workspace_id = ?)", workspaceID)
	if !includeArchived {
		q = q.Where("status = ?", string(StatusActive))
	}
	if typ != "" {
		q = q.Where("type = ?", typ)
	}
	var rows []Category
	if err := q.Order("is_system DESC, type ASC, name ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) Update(ctx context.Context, c *Category) error {
	res := r.db.WithContext(ctx).Model(&Category{}).
		Where("id = ? AND is_system = FALSE", c.ID).
		Updates(map[string]any{
			"name":        c.Name,
			"icon":        c.Icon,
			"description": c.Description,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.NotFound("Category not found")
	}
	return nil
}

func (r *Repository) SetStatus(ctx context.Context, id uuid.UUID, status string) error {
	res := r.db.WithContext(ctx).Model(&Category{}).
		Where("id = ? AND is_system = FALSE", id).
		Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.NotFound("Category not found")
	}
	return nil
}