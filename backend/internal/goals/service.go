package goals

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/audit"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/money"
)

type Service struct {
	db    *gorm.DB
	repo  *Repository
	audit *audit.Repository
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db, repo: NewRepository(db), audit: audit.NewRepository(db)}
}

type CreateInput struct {
	Name            string
	TargetAmount    int64
	CurrentAmount   int64
	TargetDate      *time.Time
	Priority        string
	LinkedAccountID *uuid.UUID
}

type UpdateInput struct {
	ID              uuid.UUID
	Name            string
	TargetAmount    int64
	CurrentAmount   int64
	TargetDate      *time.Time
	Priority        string
	LinkedAccountID *uuid.UUID
	Version         int64
}

type View struct {
	ID                     uuid.UUID  `json:"id"`
	Name                   string     `json:"name"`
	TargetAmount           string     `json:"target_amount"`
	CurrentAmount          string     `json:"current_amount"`
	TargetDate             *string    `json:"target_date"`
	Priority               string     `json:"priority"`
	LinkedAccountID        *uuid.UUID `json:"linked_account_id"`
	Status                 string     `json:"status"`
	Version                int64      `json:"version"`
	ProgressPercent        float64    `json:"progress_percent"`
	Remaining              string     `json:"remaining"`
	MonthsRemaining        int        `json:"months_remaining"`
	RequiredMonthly        string     `json:"required_monthly"`
	EstimatedMonthlyIncome string     `json:"estimated_monthly_income"`
	Feasibility            string     `json:"feasibility"`
}

func (s *Service) List(ctx context.Context, workspaceID uuid.UUID, status string) ([]View, error) {
	rows, err := s.repo.List(ctx, workspaceID, status)
	if err != nil {
		return nil, err
	}
	free := s.freeCashflow(ctx, workspaceID)
	out := make([]View, 0, len(rows))
	for i := range rows {
		out = append(out, *s.derive(&rows[i], free))
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, workspaceID, id uuid.UUID) (*View, error) {
	g, err := s.repo.FindByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	return s.derive(g, s.freeCashflow(ctx, workspaceID)), nil
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, in *CreateInput) (*View, error) {
	if err := validate(in.Name, in.TargetAmount, in.CurrentAmount, in.Priority); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	g := &Goal{
		ID:              uuid.New(),
		WorkspaceID:     workspaceID,
		Name:            strings.TrimSpace(in.Name),
		TargetAmount:    in.TargetAmount,
		CurrentAmount:   in.CurrentAmount,
		TargetDate:      in.TargetDate,
		Priority:        strings.ToUpper(in.Priority),
		LinkedAccountID: in.LinkedAccountID,
		Status:          StatusActive,
		Version:         1,
		CreatedByUserID: &userID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.Create(ctx, g); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "goal.create", "goal", &g.ID, nil)
	return s.derive(g, s.freeCashflow(ctx, workspaceID)), nil
}

func (s *Service) Update(ctx context.Context, workspaceID, userID uuid.UUID, in *UpdateInput) (*View, error) {
	if err := validate(in.Name, in.TargetAmount, in.CurrentAmount, in.Priority); err != nil {
		return nil, err
	}
	g, err := s.repo.FindByID(ctx, workspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	if g.Status == StatusAchieved || g.Status == StatusCancelled {
		return nil, errs.BusinessConflict("BUSINESS_CONFLICT", "Achieved or cancelled goals cannot be edited.")
	}
	g.Name = strings.TrimSpace(in.Name)
	g.TargetAmount = in.TargetAmount
	g.CurrentAmount = in.CurrentAmount
	g.TargetDate = in.TargetDate
	g.Priority = strings.ToUpper(in.Priority)
	g.LinkedAccountID = in.LinkedAccountID
	g.Version = in.Version
	if err := s.repo.Update(ctx, g); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "goal.update", "goal", &g.ID, nil)
	return s.derive(g, s.freeCashflow(ctx, workspaceID)), nil
}

func (s *Service) SetStatus(ctx context.Context, workspaceID, userID uuid.UUID, id uuid.UUID, to string, version int64) (*View, error) {
	g, err := s.repo.FindByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	if g.Status == to {
		return s.derive(g, s.freeCashflow(ctx, workspaceID)), nil
	}
	if err := s.repo.SetStatus(ctx, workspaceID, id, to, version); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "goal.status", "goal", &id, map[string]any{"status": to})
	fresh, err := s.repo.FindByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	return s.derive(fresh, s.freeCashflow(ctx, workspaceID)), nil
}

func (s *Service) freeCashflow(ctx context.Context, workspaceID uuid.UUID) int64 {
	net, err := s.repo.AverageMonthlyNet(ctx, workspaceID)
	if err != nil {
		return 0
	}
	return net
}

// derive computes deterministic goal metrics. progress is capped at 100%;
// feasibility compares required monthly contribution to estimated free
// cashflow and only yields ON_TRACK / AT_RISK (never overclaims).
func (s *Service) derive(g *Goal, freeCashflow int64) *View {
	pct := 0.0
	if g.TargetAmount > 0 {
		pct = (float64(g.CurrentAmount) / float64(g.TargetAmount)) * 100
		if pct > 100 {
			pct = 100
		}
	}
	remaining := g.TargetAmount - g.CurrentAmount
	if remaining < 0 {
		remaining = 0
	}

	months := 0
	if g.TargetDate != nil {
		now := time.Now().UTC()
		if !g.TargetDate.Before(atDate(now)) {
			months = monthsBetween(now, *g.TargetDate)
		} else if remaining > 0 {
			months = -1 // past due
		}
	}

	required := int64(0)
	if months > 0 {
		required = (remaining + int64(months) - 1) / int64(months)
	}

	feasibility := "ON_TRACK"
	switch {
	case remaining == 0:
		feasibility = "ON_TRACK"
	case months <= 0:
		feasibility = "AT_RISK"
	case required > freeCashflow:
		feasibility = "AT_RISK"
	}

	v := &View{
		ID:                     g.ID,
		Name:                   g.Name,
		TargetAmount:           money.FormatMinorUnits(g.TargetAmount),
		CurrentAmount:          money.FormatMinorUnits(g.CurrentAmount),
		Priority:               g.Priority,
		LinkedAccountID:        g.LinkedAccountID,
		Status:                 g.Status,
		Version:                g.Version,
		ProgressPercent:        pct,
		Remaining:              money.FormatMinorUnits(remaining),
		MonthsRemaining:        months,
		RequiredMonthly:        money.FormatMinorUnits(required),
		EstimatedMonthlyIncome: money.FormatMinorUnits(freeCashflow),
		Feasibility:            feasibility,
	}
	if g.TargetDate != nil {
		s := g.TargetDate.Format("2006-01-02")
		v.TargetDate = &s
	}
	return v
}

func monthsBetween(from, to time.Time) int {
	from, to = atDate(from), atDate(to)
	months := (to.Year()-from.Year())*12 + int(to.Month()-from.Month())
	if to.Day() < from.Day() {
		months--
	}
	if months < 0 {
		months = 0
	}
	return months
}

func atDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func validate(name string, target, current int64, priority string) error {
	fields := map[string]string{}
	if strings.TrimSpace(name) == "" {
		fields["name"] = "Goal name is required"
	}
	if target <= 0 {
		fields["target_amount"] = "Target amount must be positive"
	}
	if current < 0 {
		fields["current_amount"] = "Current amount cannot be negative"
	}
	p := strings.ToUpper(priority)
	if p != "" && p != "LOW" && p != "MEDIUM" && p != "HIGH" {
		fields["priority"] = "priority must be LOW, MEDIUM or HIGH"
	}
	if len(fields) > 0 {
		return errs.ValidationFields(fields)
	}
	return nil
}