package workspaces

import (
	"time"

	"github.com/google/uuid"
)

// Workspace is the scoping container for financial resources.
type Workspace struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name         string    `gorm:"size:120"`
	BaseCurrency string    `gorm:"type:char(3)"`
	Timezone     string    `gorm:"size:100"`
	CreatedBy    *uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Workspace) TableName() string { return "workspaces" }

// Membership links a user to a workspace with a role.
type Membership struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	WorkspaceID uuid.UUID `gorm:"type:uuid" index:"idx_membership_ws"`
	UserID      uuid.UUID `gorm:"type:uuid" index:"idx_membership_user"`
	Role        string    `gorm:"size:20"`
	Status      string    `gorm:"size:20"`
	CreatedBy   *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Membership) TableName() string { return "workspace_memberships" }

// Active roles.
const (
	RoleOwner  = "OWNER"
	RoleMember = "MEMBER"
	RoleViewer = "VIEWER"
)
