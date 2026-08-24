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

// LockMembershipByID locks a membership within a workspace by its id. The
// workspace scope makes ids from another workspace invisible (IDOR guard).
func (r *Repository) LockMembershipByID(tx *gorm.DB, workspaceID, memberID uuid.UUID) (*Membership, error) {
	var m Membership
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND workspace_id = ? AND status = 'ACTIVE'", memberID, workspaceID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("member not found")
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// FindMembershipByID finds an active membership by id within a workspace.
func (r *Repository) FindMembershipByID(ctx context.Context, workspaceID, memberID uuid.UUID) (*Membership, error) {
	var m Membership
	err := r.db.WithContext(ctx).
		Where("id = ? AND workspace_id = ? AND status = 'ACTIVE'", memberID, workspaceID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("member not found")
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// MemberSummary pairs a membership with the user's public identity for listing.
type MemberSummary struct {
	Membership
	UserName  string
	UserEmail string
}

// ListMembers returns every active member of a workspace with user identity.
func (r *Repository) ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]MemberSummary, error) {
	rows := []MemberSummary{}
	err := r.db.WithContext(ctx).
		Table("workspace_memberships AS m").
		Select("m.*, u.name AS user_name, u.email AS user_email").
		Joins("JOIN users AS u ON u.id = m.user_id").
		Where("m.workspace_id = ? AND m.status = 'ACTIVE'", workspaceID).
		Order("m.created_at ASC").
		Scan(&rows).Error
	return rows, err
}

// LockActiveOwners row-locks every active OWNER membership so concurrent
// demotions/removals serialize against the last-owner invariant, and returns
// the current owner count. Postgres forbids FOR UPDATE on aggregates, so the
// rows are locked and counted in Go.
func (r *Repository) LockActiveOwners(tx *gorm.DB, workspaceID uuid.UUID) (int64, error) {
	var ids []uuid.UUID
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Model(&Membership{}).
		Where("workspace_id = ? AND role = ? AND status = 'ACTIVE'", workspaceID, RoleOwner).
		Pluck("id", &ids).Error
	return int64(len(ids)), err
}

func (r *Repository) UpdateMembershipRole(tx *gorm.DB, id uuid.UUID, role string) error {
	return tx.Model(&Membership{}).Where("id = ?", id).Update("role", role).Error
}

func (r *Repository) DeleteMembership(tx *gorm.DB, id uuid.UUID) error {
	return tx.Where("id = ?", id).Delete(&Membership{}).Error
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
