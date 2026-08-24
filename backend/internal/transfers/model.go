package transfers

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPosted Status = "POSTED"
	StatusVoided Status = "VOIDED"
)

// Transfer moves money between two internal accounts in the same workspace.
// It is a ledger record like a transaction but never counts as income or
// expense (INV-006, INV-007).
type Transfer struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	WorkspaceID     uuid.UUID  `gorm:"type:uuid;not null"`
	FromAccountID   uuid.UUID  `gorm:"type:uuid;not null"`
	ToAccountID     uuid.UUID  `gorm:"type:uuid;not null"`
	Amount          int64      `gorm:"not null"`
	TransferDate    time.Time  `gorm:"type:date;not null"`
	Description     *string    `gorm:"type:text"`
	Status          string     `gorm:"size:20;not null;default:POSTED"`
	Version         int64      `gorm:"not null;default:1"`
	CreatedByUserID *uuid.UUID `gorm:"type:uuid"`
	VoidReason      *string    `gorm:"type:text"`
	VoidedByUserID  *uuid.UUID `gorm:"type:uuid"`
	VoidedAt        *time.Time `gorm:"type:timestamptz"`
	CreatedAt       time.Time  `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt       time.Time  `gorm:"type:timestamptz;not null;default:now()"`

	FromAccountName string `gorm:"->;-:migration"`
	ToAccountName   string `gorm:"->;-:migration"`
}

func (Transfer) TableName() string { return "account_transfers" }
