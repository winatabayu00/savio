package recurring

import (
	"time"

	"github.com/google/uuid"
)

type Frequency string

const (
	FreqDaily    Frequency = "DAILY"
	FreqWeekly   Frequency = "WEEKLY"
	FreqMonthly  Frequency = "MONTHLY"
	FreqMonthEnd Frequency = "MONTH_END"
)

func ValidFrequency(f string) bool {
	switch Frequency(f) {
	case FreqDaily, FreqWeekly, FreqMonthly, FreqMonthEnd:
		return true
	}
	return false
}

type Status string

const (
	StatusActive Status = "ACTIVE"
	StatusPaused Status = "PAUSED"
	StatusEnded  Status = "ENDED"
)

type OccurrenceStatus string

const (
	OccPending   OccurrenceStatus = "PENDING"
	OccConfirmed OccurrenceStatus = "CONFIRMED"
	OccSkipped   OccurrenceStatus = "SKIPPED"
	OccFailed    OccurrenceStatus = "FAILED"
)

// RecurringTransaction is a planned/expected activity rule. It never becomes
// actual ledger history by itself; occurrences are confirmed by the user.
type RecurringTransaction struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	WorkspaceID     uuid.UUID  `gorm:"type:uuid;not null"`
	AccountID       uuid.UUID  `gorm:"type:uuid;not null"`
	CategoryID      *uuid.UUID `gorm:"type:uuid"`
	Type            string     `gorm:"size:20;not null"`
	Amount          int64      `gorm:"not null"`
	Frequency       string     `gorm:"size:20;not null"`
	DayOfMonth      *int       `gorm:"type:int"`
	DayOfWeek       *int       `gorm:"type:int"`
	StartDate       time.Time  `gorm:"type:date;not null"`
	EndDate         *time.Time `gorm:"type:date"`
	Description     *string    `gorm:"type:text"`
	Merchant        *string    `gorm:"size:200"`
	Notes           *string    `gorm:"type:text"`
	Status          string     `gorm:"size:20;not null;default:ACTIVE"`
	AutoPost        bool       `gorm:"not null;default:false"`
	Version         int64      `gorm:"not null;default:1"`
	CreatedByUserID *uuid.UUID `gorm:"type:uuid"`
	CreatedAt       time.Time  `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt       time.Time  `gorm:"type:timestamptz;not null;default:now()"`

	AccountName  string `gorm:"->;-:migration"`
	CategoryName string `gorm:"->;-:migration"`
}

func (RecurringTransaction) TableName() string { return "recurring_transactions" }

// RecurringOccurrence materializes one scheduled instance of a rule.
// The UNIQUE (recurring_id, due_date) constraint guarantees a rule–date pair
// can become actual at most once (INV-010).
type RecurringOccurrence struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey"`
	RecurringID         uuid.UUID  `gorm:"type:uuid;not null"`
	WorkspaceID         uuid.UUID  `gorm:"type:uuid;not null"`
	DueDate             time.Time  `gorm:"type:date;not null"`
	Status              string     `gorm:"size:20;not null;default:PENDING;index:idx_occurrence_status"`
	PostedTransactionID *uuid.UUID `gorm:"type:uuid"`
	Version             int64      `gorm:"not null;default:1"`
	CreatedAt           time.Time  `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt           time.Time  `gorm:"type:timestamptz;not null;default:now()"`

	RecurringType    string `gorm:"->;-:migration"`
	RecurringAmount  int64  `gorm:"->;-:migration"`
	RecurringAccount string `gorm:"->;-:migration"`
}

func (RecurringOccurrence) TableName() string { return "recurring_occurrences" }
