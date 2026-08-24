package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entry struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	WorkspaceID  *uuid.UUID     `gorm:"type:uuid"`
	ActorUserID  *uuid.UUID     `gorm:"type:uuid"`
	Action       string         `gorm:"size:50;not null"`
	ResourceType string         `gorm:"size:50;not null"`
	ResourceID   *uuid.UUID     `gorm:"type:uuid"`
	Metadata     map[string]any `gorm:"type:jsonb"`
	OccurredAt   time.Time      `gorm:"type:timestamptz;not null;default:now()"`
}

func (Entry) TableName() string { return "audit_logs" }

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// Record appends an audit entry. Audit writes are best-effort and never
// block the primary financial operation (INV not affected on failure).
func (r *Repository) Record(ctx context.Context, workspaceID, userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, metadata map[string]any) {
	entry := Entry{
		ID:           uuid.New(),
		WorkspaceID:  workspaceID,
		ActorUserID:  userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Metadata:     metadata,
		OccurredAt:   time.Now().UTC(),
	}
	_ = r.db.WithContext(ctx).Create(&entry).Error
}
