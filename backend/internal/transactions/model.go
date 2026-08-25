package transactions

import (
	"time"

	"github.com/google/uuid"
)

type Type string

const (
	TypeIncome     Type = "INCOME"
	TypeExpense    Type = "EXPENSE"
	TypeAdjustment Type = "ADJUSTMENT"
)

func ValidType(t string) bool {
	switch Type(t) {
	case TypeIncome, TypeExpense, TypeAdjustment:
		return true
	}
	return false
}

type Status string

const (
	StatusDraft  Status = "DRAFT"
	StatusPosted Status = "POSTED"
	StatusVoided Status = "VOIDED"
)

type Source string

const (
	SourceManual    Source = "MANUAL"
	SourceAI        Source = "AI"
	SourceImport    Source = "IMPORT"
	SourceRecurring Source = "RECURRING"
	SourceTelegram  Source = "TELEGRAM"
	SourceSystem    Source = "SYSTEM"
)

// Transaction is the authoritative ledger record. amount is integer minor
// units; INCOME/EXPENSE are positive with direction from type, ADJUSTMENT is
// signed (negative reduces the account balance).
type Transaction struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	WorkspaceID     uuid.UUID  `gorm:"type:uuid;not null"`
	AccountID       uuid.UUID  `gorm:"type:uuid;not null"`
	CategoryID      *uuid.UUID `gorm:"type:uuid"`
	Type            string     `gorm:"size:20;not null"`
	Amount          int64      `gorm:"not null"`
	TransactionDate time.Time  `gorm:"type:date;not null"`
	Description     *string    `gorm:"type:text"`
	Merchant        *string    `gorm:"size:200"`
	Notes           *string    `gorm:"type:text"`
	Source          string     `gorm:"size:20;not null;default:MANUAL"`
	Status          string     `gorm:"size:20;not null;default:DRAFT"`
	Version         int64      `gorm:"not null;default:1"`
	CreatedByUserID *uuid.UUID `gorm:"type:uuid"`
	VoidReason      *string    `gorm:"type:text"`
	VoidedByUserID  *uuid.UUID `gorm:"type:uuid"`
	PostedAt        *time.Time `gorm:"type:timestamptz"`
	VoidedAt        *time.Time `gorm:"type:timestamptz"`
	CreatedAt       time.Time  `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt       time.Time  `gorm:"type:timestamptz;not null;default:now()"`

	// Joined display fields populated by the listing query.
	AccountName   string `gorm:"->;-:migration"`
	CategoryName  string `gorm:"->;-:migration"`
	CategoryType  string `gorm:"->;-:migration"`
	CreatedByName string `gorm:"->;-:migration"`
}

func (Transaction) TableName() string { return "transactions" }
