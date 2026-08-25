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
	ActorType    string         `gorm:"size:20;not null"`
	Action       string         `gorm:"size:50;not null"`
	ResourceType string         `gorm:"size:50;not null"`
	ResourceID   *uuid.UUID     `gorm:"type:uuid"`
	Reason       *string        `gorm:"type:text"`
	BeforeData   map[string]any `gorm:"type:jsonb"`
	AfterData    map[string]any `gorm:"type:jsonb"`
	Metadata     map[string]any `gorm:"type:jsonb"`
	OccurredAt   time.Time      `gorm:"type:timestamptz;not null;default:now()"`
}

func (Entry) TableName() string { return "audit_logs" }

type Repository struct {
	db *gorm.DB
}

type View struct {
	ID           uuid.UUID      `json:"id"`
	ActorUserID  *uuid.UUID     `json:"actor_user_id"`
	ActorName    *string        `json:"actor_name"`
	ActorType    string         `json:"actor_type"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   *uuid.UUID     `json:"resource_id"`
	Reason       *string        `json:"reason"`
	BeforeData   map[string]any `json:"before_data"`
	AfterData    map[string]any `json:"after_data"`
	OccurredAt   time.Time      `json:"occurred_at"`
}

func (r *Repository) List(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]View, int64, error) {
	var entries []View
	var total int64
	q := r.db.WithContext(ctx).Table("audit_logs a").Where("a.workspace_id = ?", workspaceID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Select("a.id, a.actor_user_id, u.name AS actor_name, a.actor_type, a.action, a.resource_type, a.resource_id, a.reason, a.before_data, a.after_data, a.occurred_at").Joins("LEFT JOIN users u ON u.id = a.actor_user_id").Order("a.occurred_at DESC").Limit(limit).Offset(offset).Scan(&entries).Error
	return entries, total, err
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// Record appends an audit entry. Audit writes are best-effort and never
// block the primary financial operation (INV not affected on failure).
type Change struct {
	ActorType string
	Reason    string
	Before    map[string]any
	After     map[string]any
	Metadata  map[string]any
}

func (r *Repository) RecordChange(ctx context.Context, workspaceID, userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, change Change) {
	actorType := normalizeActorType(change.ActorType)
	entry := Entry{
		ID:           uuid.New(),
		WorkspaceID:  workspaceID,
		ActorUserID:  userID,
		ActorType:    actorType,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Reason:       nullable(change.Reason),
		BeforeData:   change.Before,
		AfterData:    change.After,
		Metadata:     change.Metadata,
		OccurredAt:   time.Now().UTC(),
	}
	_ = r.db.WithContext(ctx).Create(&entry).Error
}

func normalizeActorType(actorType string) string {
	if actorType == "" {
		return "USER"
	}
	return actorType
}

func (r *Repository) Record(ctx context.Context, workspaceID, userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, metadata map[string]any) {
	r.RecordChange(ctx, workspaceID, userID, action, resourceType, resourceID, Change{Metadata: metadata})
}

func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
