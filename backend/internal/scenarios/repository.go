package scenarios

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/errs"
)

const (
	StatusDraft      = "DRAFT"
	StatusCalculated = "CALCULATED"
)

const (
	ModOneTimeExpense   = "ONE_TIME_EXPENSE"
	ModOneTimeIncome    = "ONE_TIME_INCOME"
	ModRecurringExpense = "RECURRING_EXPENSE"
	ModRecurringIncome  = "RECURRING_INCOME"
	ModIncomeReduction  = "INCOME_REDUCTION"
	ModIncomeRemoval    = "INCOME_REMOVAL"
	ModExpenseReduction = "EXPENSE_REDUCTION"
)

var ValidModTypes = map[string]bool{
	ModOneTimeExpense: true, ModOneTimeIncome: true,
	ModRecurringExpense: true, ModRecurringIncome: true,
	ModIncomeReduction: true, ModIncomeRemoval: true,
	ModExpenseReduction: true,
}

type Scenario struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey"`
	WorkspaceID        uuid.UUID `gorm:"type:uuid;not null"`
	Name               string    `gorm:"size:120;not null"`
	Description        *string
	Status             string `gorm:"size:20;not null;default:DRAFT"`
	IsStale            bool   `gorm:"not null;default:false"`
	BaselineSnapshot   string `gorm:"type:text"`
	Result             string `gorm:"type:text"`
	CalculationVersion string `gorm:"size:10"`
	Version            int64  `gorm:"not null;default:1"`
	CreatedByUserID    *uuid.UUID
	CreatedAt          time.Time `gorm:"type:timestamptz;not null;default:now()"`
	UpdatedAt          time.Time `gorm:"type:timestamptz;not null;default:now()"`
}

func (Scenario) TableName() string { return "scenarios" }

type Modification struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	ScenarioID uuid.UUID `gorm:"type:uuid;not null"`
	Type       string    `gorm:"size:30;not null"`
	Amount     int64     `gorm:"not null"`
	Frequency  *string
	Narrative  *string
	Position   int   `gorm:"not null;default:0"`
	Version    int64 `gorm:"not null;default:1"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (Modification) TableName() string { return "scenario_modifications" }

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) CreateScenario(ctx context.Context, s *Scenario) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *Repository) FindScenario(ctx context.Context, workspaceID, id uuid.UUID) (*Scenario, error) {
	var s Scenario
	err := r.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", id, workspaceID).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("Scenario not found")
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) ListScenarios(ctx context.Context, workspaceID uuid.UUID) ([]Scenario, error) {
	var rows []Scenario
	if err := r.db.WithContext(ctx).Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) UpdateScenario(ctx context.Context, s *Scenario) error {
	res := r.db.WithContext(ctx).Model(&Scenario{}).
		Where("id = ? AND workspace_id = ? AND version = ?", s.ID, s.WorkspaceID, s.Version).
		Updates(map[string]any{
			"name": s.Name, "description": s.Description,
			"status": s.Status, "is_stale": s.IsStale,
			"baseline_snapshot": s.BaselineSnapshot, "result": s.Result,
			"calculation_version": s.CalculationVersion,
			"version":             gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.VersionConflict("This scenario was changed. Reload the latest version.")
	}
	return nil
}

func (r *Repository) DeleteScenario(ctx context.Context, workspaceID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", id, workspaceID).Delete(&Scenario{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.NotFound("Scenario not found")
	}
	return nil
}

func (r *Repository) CreateModification(ctx context.Context, m *Modification) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *Repository) ListModifications(ctx context.Context, scenarioID uuid.UUID) ([]Modification, error) {
	var rows []Modification
	if err := r.db.WithContext(ctx).Where("scenario_id = ?", scenarioID).
		Order("position ASC, created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *Repository) FindModification(ctx context.Context, scenarioID, id uuid.UUID) (*Modification, error) {
	var m Modification
	err := r.db.WithContext(ctx).Where("id = ? AND scenario_id = ?", id, scenarioID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("Modification not found")
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) UpdateModification(ctx context.Context, m *Modification) error {
	res := r.db.WithContext(ctx).Model(&Modification{}).
		Where("id = ? AND scenario_id = ? AND version = ?", m.ID, m.ScenarioID, m.Version).
		Updates(map[string]any{
			"type": m.Type, "amount": m.Amount, "frequency": m.Frequency,
			"narrative": m.Narrative, "version": gorm.Expr("version + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.VersionConflict("This modification was changed. Reload the latest version.")
	}
	return nil
}

func (r *Repository) DeleteModification(ctx context.Context, scenarioID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("id = ? AND scenario_id = ?", id, scenarioID).Delete(&Modification{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.NotFound("Modification not found")
	}
	return nil
}

func marshal(v any) (string, error) {
	b, err := json.Marshal(v)
	return string(b), err
}

func unmarshal(s string, v any) error {
	if s == "" {
		return nil
	}
	return json.Unmarshal([]byte(s), v)
}
