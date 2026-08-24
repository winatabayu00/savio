package workspaces

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/savio/savio/backend/internal/platform/errs"
)

// Repository owns workspaces and workspace_memberships persistence.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateWorkspace(ctx context.Context, w *Workspace) error {
	return r.db.WithContext(ctx).Create(w).Error
}

func (r *Repository) CreateMembership(ctx context.Context, m *Membership) error {
	return r.db.WithContext(ctx).Create(m).Error
}

// FindMembership returns an active membership for user+workspace.
func (r *Repository) FindMembership(ctx context.Context, workspaceID, userID uuid.UUID) (*Membership, error) {
	var m Membership
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND user_id = ? AND status = 'ACTIVE'", workspaceID, userID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ResourceAccessDenied()
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// FindWorkspaceByID finds a workspace.
func (r *Repository) FindWorkspaceByID(ctx context.Context, id uuid.UUID) (*Workspace, error) {
	var w Workspace
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&w).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("workspace not found")
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// FindDefaultByUser returns the default selected workspace for a user: if the
// user requests a specific active workspace, that one; otherwise the first
// active membership. The personal-default behavior keeps single-workspace
// usage simple.
func (r *Repository) FindDefaultByUser(ctx context.Context, userID uuid.UUID) (*Membership, error) {
	var m Membership
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = 'ACTIVE'", userID).
		Order("created_at ASC").
		Limit(1).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.ResourceAccessDenied()
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// LockMembershipForUpdate locks a membership row for role-change operations
// (prevents concurrent owner demotions from removing the last owner).
func (r *Repository) LockMembershipForUpdate(tx *gorm.DB, workspaceID, userID uuid.UUID) (*Membership, error) {
	var m Membership
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("workspace_id = ? AND user_id = ? AND status = 'ACTIVE'", workspaceID, userID).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) GetDB() *gorm.DB { return r.db }

// ListActiveMembershipsForUser returns all active memberships.
func (r *Repository) ListActiveMembershipsForUser(ctx context.Context, userID uuid.UUID) ([]Membership, error) {
	var rows []Membership
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = 'ACTIVE'", userID).
		Order("created_at ASC").Find(&rows).Error
	return rows, err
}

func (r *Repository) TouchWorkspace(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&Workspace{}).Where("id = ?", id).
		Update("updated_at", time.Now()).Error
}
