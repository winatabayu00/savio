package categories

import (
	"time"

	"github.com/google/uuid"
)

type Type string

const (
	TypeIncome  Type = "INCOME"
	TypeExpense Type = "EXPENSE"
)

func ValidType(t string) bool {
	return Type(t) == TypeIncome || Type(t) == TypeExpense
}

type Status string

const (
	StatusActive   Status = "ACTIVE"
	StatusArchived Status = "ARCHIVED"
)

// Category is a classification resource. System categories have NULL
// workspace_id and are globally available; custom categories are
// workspace-scoped.
type Category struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	WorkspaceID *uuid.UUID `gorm:"type:uuid"`
	Name        string     `gorm:"size:120;not null"`
	Type        string     `gorm:"size:20;not null"`
	ParentID    *uuid.UUID `gorm:"type:uuid"`
	IsSystem    bool       `gorm:"not null;default:false"`
	Status      string     `gorm:"size:20;not null;default:ACTIVE"`
	Icon        *string    `gorm:"size:100"`
	Description *string    `gorm:"type:text"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt   time.Time  `gorm:"type:timestamptz;not null;default:now()"`
}

func (Category) TableName() string { return "categories" }