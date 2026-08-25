package budgets

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
	CategoryID  uuid.UUID
	AmountMinor int64
	PeriodStart time.Time
	PeriodEnd   time.Time
}

type UpdateInput struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	AmountMinor int64
	PeriodStart time.Time
	PeriodEnd   time.Time
	Version     int64
}

type ComputedStatus string

const (
	ComputedOnTrack   ComputedStatus = "ON_TRACK"
	ComputedWarning   ComputedStatus = "WARNING"
	ComputedExceeded  ComputedStatus = "EXCEEDED"
)

type View struct {
	ID               uuid.UUID `json:"id"`
	CategoryID       uuid.UUID `json:"category_id"`
	CategoryName     string    `json:"category_name"`
	Amount           string    `json:"amount"`
	PeriodStart      string    `json:"period_start"`
	PeriodEnd        string    `json:"period_end"`
	Status           string    `json:"status"`
	Version          int64     `json:"version"`
	Spent            string    `json:"spent"`
	Remaining        string    `json:"remaining"`
	Utilization      float64   `json:"utilization_percent"`
	ComputedStatus   string    `json:"computed_status"`
	ProjectedSpend   string    `json:"projected_spend"`
	ProjectedOverspend *string `json:"projected_overspend"`
}

func (s *Service) List(ctx context.Context, workspaceID uuid.UUID, status string, warning int64) ([]View, error) {
	rows, err := s.repo.List(ctx, workspaceID, status)
	if err != nil {
		return nil, err
	}
	out := make([]View, 0, len(rows))
	for i := range rows {
		v, err := s.derive(ctx, workspaceID, &rows[i], warning)
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, workspaceID, id uuid.UUID, warning int64) (*View, error) {
	b, err := s.repo.FindByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	return s.derive(ctx, workspaceID, b, warning)
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, in *CreateInput) (*View, error) {
	if err := validate(in.AmountMinor, in.PeriodStart, in.PeriodEnd); err != nil {
		return nil, err
	}
	if err := s.validateExpenseCategory(ctx, workspaceID, in.CategoryID); err != nil {
		return nil, err
	}
	overlap, err := s.repo.OverlappingActive(ctx, workspaceID, in.CategoryID, in.PeriodStart, in.PeriodEnd, uuid.Nil)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, errs.Duplicate("A budget for this category already covers this period. Close it first or use a different period.")
	}
	b := &Budget{
		ID:              uuid.New(),
		WorkspaceID:     workspaceID,
		CategoryID:      in.CategoryID,
		Amount:          in.AmountMinor,
		PeriodStart:     in.PeriodStart,
		PeriodEnd:       in.PeriodEnd,
		Status:          string(StatusActive),
		Version:         1,
		CreatedByUserID: &userID,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, b); err != nil {
		if containsDuplicate(err) {
			return nil, errs.Duplicate("A budget for this category already covers this period.")
		}
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "budget.create", "budget", &b.ID, nil)
	return s.Get(ctx, workspaceID, b.ID, warningThreshold(ctx, s.db, userID))
}

func (s *Service) Update(ctx context.Context, workspaceID, userID uuid.UUID, in *UpdateInput, warning int64) (*View, error) {
	if err := validate(in.AmountMinor, in.PeriodStart, in.PeriodEnd); err != nil {
		return nil, err
	}
	if err := s.validateExpenseCategory(ctx, workspaceID, in.CategoryID); err != nil {
		return nil, err
	}
	overlap, err := s.repo.OverlappingActive(ctx, workspaceID, in.CategoryID, in.PeriodStart, in.PeriodEnd, in.ID)
	if err != nil {
		return nil, err
	}
	if overlap {
		return nil, errs.Duplicate("A budget for this category already covers this period.")
	}
	b, err := s.repo.FindByID(ctx, workspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	if b.Status != string(StatusActive) {
		return nil, errs.BusinessConflict("BUSINESS_CONFLICT", "Closed budgets cannot be edited.")
	}
	b.CategoryID = in.CategoryID
	b.Amount = in.AmountMinor
	b.PeriodStart = in.PeriodStart
	b.PeriodEnd = in.PeriodEnd
	b.Version = in.Version
	if err := s.repo.Update(ctx, b); err != nil {
		if containsDuplicate(err) {
			return nil, errs.Duplicate("A budget for this category already covers this period.")
		}
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "budget.update", "budget", &b.ID, nil)
	return s.Get(ctx, workspaceID, in.ID, warning)
}

func (s *Service) Close(ctx context.Context, workspaceID, userID uuid.UUID, id uuid.UUID, version int64, warning int64) (*View, error) {
	if err := s.repo.Close(ctx, workspaceID, id, version); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "budget.close", "budget", &id, nil)
	return s.Get(ctx, workspaceID, id, warning)
}

// derive computes deterministic budget health (AGENTS #48: derived, not
// stored).
func (s *Service) derive(ctx context.Context, workspaceID uuid.UUID, b *Budget, warningPct int64) (*View, error) {
	spent, err := s.repo.spent(ctx, workspaceID, b.CategoryID, b.PeriodStart, b.PeriodEnd)
	if err != nil {
		return nil, err
	}
	remaining := b.Amount - spent
	utilization := 0.0
	if b.Amount > 0 {
		utilization = (float64(spent) / float64(b.Amount)) * 100
	}
	cs := ComputedOnTrack
	if utilization >= 100 {
		cs = ComputedExceeded
	} else if utilization >= float64(warningPct) {
		cs = ComputedWarning
	}
	v := &View{
		ID:             b.ID,
		CategoryID:     b.CategoryID,
		CategoryName:   b.CategoryName,
		Amount:         money.FormatMinorUnits(b.Amount),
		PeriodStart:    b.PeriodStart.Format("2006-01-02"),
		PeriodEnd:      b.PeriodEnd.Format("2006-01-02"),
		Status:         b.Status,
		Version:        b.Version,
		Spent:          money.FormatMinorUnits(spent),
		Remaining:      money.FormatMinorUnits(remaining),
		Utilization:    utilization,
		ComputedStatus: string(cs),
	}
	if b.Status == string(StatusActive) {
		s.projection(v, spent, b)
	}
	return v, nil
}

// projection estimates end-of-period spend from pace-to-date and flags a
// projected overspend. Deterministic, assumption-based (AGENTS #56).
func (s *Service) projection(v *View, spent int64, b *Budget) {
	days := float64(b.PeriodEnd.Sub(b.PeriodStart).Hours()/24) + 1
	elapsed := float64(time.Now().UTC().Sub(b.PeriodStart).Hours()/24) + 1
	if elapsed < 1 {
		elapsed = 1
	}
	if days < 1 {
		days = 1
	}
	projected := int64(float64(spent) / elapsed * days)
	v.ProjectedSpend = money.FormatMinorUnits(projected)
	if projected > b.Amount {
		diff := projected - b.Amount
		v.ProjectedOverspend = stringPtr(money.FormatMinorUnits(diff))
	}
}

// WarningThreshold returns the user's configured budget warning percentage.
func (s *Service) WarningThreshold(ctx context.Context, userID uuid.UUID) int64 {
	return warningThreshold(ctx, s.db, userID)
}

func warningThreshold(ctx context.Context, db *gorm.DB, userID uuid.UUID) int64 {
	var t float64
	if err := db.WithContext(ctx).Table("user_settings").
		Select("budget_warning_threshold").Where("user_id = ?", userID).Scan(&t).Error; err != nil || t <= 0 {
		return 80
	}
	return int64(t)
}

// validateExpenseCategory ensures the budget references an ACTIVE EXPENSE
// category available to the workspace (system or custom). Without this a
// budget could silently target an INCOME category or another workspace.
func (s *Service) validateExpenseCategory(ctx context.Context, workspaceID, categoryID uuid.UUID) error {
	var n int64
	if err := s.db.WithContext(ctx).Table("categories").
		Where("id = ? AND type = 'EXPENSE' AND status = 'ACTIVE' AND (is_system = TRUE OR workspace_id = ?)",
			categoryID, workspaceID).Count(&n).Error; err != nil {
		return errs.WrapInternal(err, "validate budget category")
	}
	if n == 0 {
		return errs.ValidationFields(map[string]string{"category_id": "Category must be an active expense category in this workspace"})
	}
	return nil
}

func validate(amount int64, start, end time.Time) error {
	fields := map[string]string{}
	if amount <= 0 {
		fields["amount"] = "Budget amount must be positive"
	}
	if start.IsZero() || end.IsZero() {
		fields["period"] = "Period start and end are required"
	} else if end.Before(start) {
		fields["period"] = "Period end must be after period start"
	}
	if len(fields) > 0 {
		return errs.ValidationFields(fields)
	}
	return nil
}

func containsDuplicate(err error) bool {
	return err != nil && (strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(err.Error(), "23505"))
}

func stringPtr(s string) *string { return &s }