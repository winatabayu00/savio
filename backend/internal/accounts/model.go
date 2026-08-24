package accounts

import (
	"time"

	"github.com/google/uuid"
)

type Type string

const (
	TypeCash    Type = "CASH"
	TypeBank    Type = "BANK"
	TypeEWallet Type = "EWALLET"
	TypeSavings Type = "SAVINGS"
	TypeOther   Type = "OTHER"
)

func ValidType(t string) bool {
	switch Type(t) {
	case TypeCash, TypeBank, TypeEWallet, TypeSavings, TypeOther:
		return true
	}
	return false
}

type Status string

const (
	StatusActive   Status = "ACTIVE"
	StatusArchived Status = "ARCHIVED"
)

// Account is the persistent account row. Monetary values are integer minor
// units aligned with the workspace base currency.
type Account struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	WorkspaceID     uuid.UUID  `gorm:"type:uuid;not null"`
	Name            string     `gorm:"size:120;not null"`
	Type            string     `gorm:"size:30;not null"`
	Currency        string     `gorm:"type:char(3);not null"`
	OpeningBalance  int64      `gorm:"not null"`
	InstitutionName *string    `gorm:"size:150"`
	Description     *string    `gorm:"type:text"`
	Status          string     `gorm:"size:20;not null;default:ACTIVE"`
	Version         int64      `gorm:"not null;default:1"`
	CreatedByUserID *uuid.UUID `gorm:"type:uuid"`
	ArchivedAt      *time.Time `gorm:"type:timestamptz"`
	CreatedAt       time.Time  `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt       time.Time  `gorm:"type:timestamptz;not null;default:now()"`
}

func (Account) TableName() string { return "accounts" }