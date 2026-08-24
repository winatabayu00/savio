package workspaces

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/authctx"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/users"
)

// Service owns workspace-scoped authorization and membership management.
type Service struct {
	db    *gorm.DB
	users *users.Repository
	repo  *Repository
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db, users: users.NewRepository(db), repo: NewRepository(db)}
}

// CurrentResult is the requester's view of the active workspace.
type CurrentResult struct {
	Workspace   *Workspace
	Role        authctx.Role
	MemberCount int64
}

// Current summarizes the active workspace for the authenticated requester.
func (s *Service) Current(ctx context.Context, workspaceID uuid.UUID, role authctx.Role) (*CurrentResult, error) {
	ws, err := s.repo.FindWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	var count int64
	s.db.WithContext(ctx).Model(&Membership{}).
		Where("workspace_id = ? AND status = 'ACTIVE'", workspaceID).Count(&count)
	return &CurrentResult{Workspace: ws, Role: role, MemberCount: count}, nil
}

// ListMembers returns active members with user identity for the workspace.
func (s *Service) ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]MemberSummary, error) {
	return s.repo.ListMembers(ctx, workspaceID)
}

// AddMember adds an existing Savio user by email as MEMBER or VIEWER. New
// members can never be granted OWNER: the workspace owner invites, owners are
// only produced through ownership transfer on an existing member.
func (s *Service) AddMember(ctx context.Context, workspaceID uuid.UUID, email, role string) (*MemberSummary, error) {
	fields := map[string]string{}
	if len(email) < 3 || !strings.Contains(email, "@") {
		fields["email"] = "A valid email is required"
	}
	if role == "" {
		role = RoleMember
	}
	if role != RoleMember && role != RoleViewer {
		fields["role"] = "Role must be MEMBER or VIEWER"
	}
	if len(fields) > 0 {
		return nil, errs.ValidationFields(fields)
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, errs.ValidationFields(map[string]string{"email": "No Savio account found for this email"})
	}

	m := &Membership{
		ID:          uuid.New(),
		WorkspaceID: workspaceID,
		UserID:      user.ID,
		Role:        role,
		Status:      "ACTIVE",
	}
	if err := s.repo.CreateMembership(ctx, m); err != nil {
		if isUniqueViolation(err) {
			return nil, errs.Duplicate("This user is already a member of the workspace")
		}
		return nil, errs.Internal(err)
	}
	return &MemberSummary{Membership: *m, UserName: user.Name, UserEmail: user.Email}, nil
}

// UpdateRole changes a member's role. Demoting the last OWNER is rejected
// (INV-003); the owner-count read is row-locked inside the transaction so
// concurrent demotions cannot both pass.
func (s *Service) UpdateRole(ctx context.Context, workspaceID, memberID uuid.UUID, role string) error {
	if role != RoleOwner && role != RoleMember && role != RoleViewer {
		return errs.ValidationFields(map[string]string{"role": "Role must be OWNER, MEMBER or VIEWER"})
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m, err := s.repo.LockMembershipByID(tx, workspaceID, memberID)
		if err != nil {
			return err
		}
		if m.Role == RoleOwner && role != RoleOwner {
			n, err := s.repo.LockActiveOwners(tx, workspaceID)
			if err != nil {
				return err
			}
			if n < 2 {
				return errs.BusinessConflict(errs.CodeBusinessConflict, "The workspace must always keep at least one owner")
			}
		}
		return s.repo.UpdateMembershipRole(tx, memberID, role)
	})
}

// RemoveMember revokes a membership. The last OWNER cannot be removed
// (INV-003); the check is row-locked like UpdateRole.
func (s *Service) RemoveMember(ctx context.Context, workspaceID, memberID uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m, err := s.repo.LockMembershipByID(tx, workspaceID, memberID)
		if err != nil {
			return err
		}
		if m.Role == RoleOwner {
			n, err := s.repo.LockActiveOwners(tx, workspaceID)
			if err != nil {
				return err
			}
			if n < 2 {
				return errs.BusinessConflict(errs.CodeBusinessConflict, "The workspace must always keep at least one owner")
			}
		}
		return s.repo.DeleteMembership(tx, memberID)
	})
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23505") || strings.Contains(msg, "duplicate key")
}
