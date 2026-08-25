package scenarios

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/audit"
	"github.com/savio/savio/backend/internal/forecast"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/money"
	"github.com/savio/savio/backend/internal/recurring"
)

const calcVersion = "1"

type Service struct {
	db    *gorm.DB
	repo  *Repository
	fc    *forecast.Service
	audit *audit.Repository
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db, repo: NewRepository(db), fc: forecast.NewService(db), audit: audit.NewRepository(db)}
}

type CreateInput struct {
	Name        string
	Description string
}

type UpdateInput struct {
	ID          uuid.UUID
	Name        string
	Description string
	Version     int64
}

type ModInput struct {
	Type       string
	Amount     int64
	Frequency  string
	Narrative  string
	VersionPtr *int64
}

type ResultView struct {
	BaselineEnding     string       `json:"baseline_ending_balance"`
	ScenarioEnding     string       `json:"scenario_ending_balance"`
	BaselineMinimum    string       `json:"baseline_minimum_balance"`
	ScenarioMinimum    string       `json:"scenario_minimum_balance"`
	BaselineIncome     string       `json:"baseline_income"`
	ScenarioIncome     string       `json:"scenario_income"`
	BaselineExpense    string       `json:"baseline_expense"`
	ScenarioExpense    string       `json:"scenario_expense"`
	CashflowDifference string       `json:"cashflow_difference"`
	ModifiedEvents     int          `json:"modified_events"`
	AssumptionNote     string       `json:"assumption_note"`
	CalculationVersion string       `json:"calculation_version"`
	Timeline           []DayBalance `json:"timeline"`
}

type DayBalance struct {
	Date            string `json:"date"`
	BaselineBalance string `json:"baseline_balance"`
	ScenarioBalance string `json:"scenario_balance"`
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, in *CreateInput) (*View, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ValidationFields(map[string]string{"name": "Scenario name is required"})
	}
	now := time.Now().UTC()
	sn := &Scenario{
		ID: uuid.New(), WorkspaceID: workspaceID, Name: strings.TrimSpace(in.Name),
		Status: StatusDraft, Version: 1, CreatedByUserID: &userID,
		CreatedAt: now, UpdatedAt: now,
	}
	if strings.TrimSpace(in.Description) != "" {
		d := strings.TrimSpace(in.Description)
		sn.Description = &d
	}
	if err := s.repo.CreateScenario(ctx, sn); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "scenario.create", "scenario", &sn.ID, nil)
	return s.view(ctx, workspaceID, sn)
}

func (s *Service) Update(ctx context.Context, workspaceID, userID uuid.UUID, in *UpdateInput) (*View, error) {
	sn, err := s.repo.FindScenario(ctx, workspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ValidationFields(map[string]string{"name": "Scenario name is required"})
	}
	sn.Name = strings.TrimSpace(in.Name)
	if strings.TrimSpace(in.Description) != "" {
		d := strings.TrimSpace(in.Description)
		sn.Description = &d
	} else {
		sn.Description = nil
	}
	sn.Version = in.Version
	if err := s.repo.UpdateScenario(ctx, sn); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "scenario.update", "scenario", &sn.ID, nil)
	return s.view(ctx, workspaceID, sn)
}

func (s *Service) Delete(ctx context.Context, workspaceID, userID uuid.UUID, id uuid.UUID) error {
	if err := s.repo.DeleteScenario(ctx, workspaceID, id); err != nil {
		return err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "scenario.delete", "scenario", &id, nil)
	return nil
}

func (s *Service) List(ctx context.Context, workspaceID uuid.UUID) ([]View, error) {
	rows, err := s.repo.ListScenarios(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]View, 0, len(rows))
	for i := range rows {
		v, err := s.view(ctx, workspaceID, &rows[i])
		if err != nil {
			return nil, err
		}
		out = append(out, *v)
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, workspaceID, id uuid.UUID) (*View, error) {
	sn, err := s.repo.FindScenario(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	if sn.Status == StatusCalculated {
		if err := s.refreshStale(ctx, workspaceID, sn); err != nil {
			return nil, err
		}
	}
	return s.view(ctx, workspaceID, sn)
}

func (s *Service) view(ctx context.Context, workspaceID uuid.UUID, sn *Scenario) (*View, error) {
	mods, err := s.repo.ListModifications(ctx, sn.ID)
	if err != nil {
		return nil, err
	}
	out := &View{
		ID: sn.ID, Name: sn.Name, Description: sn.Description, Status: sn.Status,
		IsStale: sn.IsStale, Version: sn.Version, CreatedAt: sn.CreatedAt, UpdatedAt: sn.UpdatedAt,
		CalculationVersion: sn.CalculationVersion,
	}
	var result ResultView
	if sn.Result != "" {
		if err := unmarshal(sn.Result, &result); err == nil {
			out.Result = &result
		}
	}
	out.Modifications = make([]ModView, 0, len(mods))
	for i := range mods {
		out.Modifications = append(out.Modifications, toModView(&mods[i]))
	}
	return out, nil
}

func toModView(m *Modification) ModView {
	return ModView{
		ID: m.ID, Type: m.Type, Frequency: m.Frequency, Narrative: m.Narrative,
		Amount: money.FormatMinorUnits(m.Amount), Version: m.Version, UpdatedAt: m.UpdatedAt,
	}
}

// Calculate runs a non-destructive overlay over a fresh baseline and persists
// a snapshot result. Real finance state is never touched (INV-012).
func (s *Service) Calculate(ctx context.Context, workspaceID, userID uuid.UUID, id uuid.UUID, horizon int, now time.Time) (*View, error) {
	sn, err := s.repo.FindScenario(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	base, err := s.fc.Compute(ctx, workspaceID, horizon, now)
	if err != nil {
		return nil, err
	}
	mods, err := s.repo.ListModifications(ctx, sn.ID)
	if err != nil {
		return nil, err
	}
	result := s.apply(base, mods, horizon, now)

	baselineSnap := map[string]any{
		"ending": base.EndingBalance, "minimum": base.MinimumBalance,
		"income": base.ProjectedIncome, "expense": base.ProjectedExpense,
		"horizon": horizon, "as_of": now.Format("2006-01-02"),
	}
	bs, _ := marshal(baselineSnap)
	rs, _ := marshal(result)
	sn.BaselineSnapshot = bs
	sn.Result = rs
	sn.Status = StatusCalculated
	sn.CalculationVersion = calcVersion
	sn.IsStale = false
	if err := s.repo.UpdateScenario(ctx, sn); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "scenario.calculate", "scenario", &sn.ID, nil)
	fresh, err := s.repo.FindScenario(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	out, err := s.view(ctx, workspaceID, fresh)
	if err != nil {
		return nil, err
	}
	rt := result
	out.Result = &rt
	return out, nil
}

func (s *Service) AddModification(ctx context.Context, workspaceID, userID uuid.UUID, scenarioID uuid.UUID, in *ModInput) (*ModView, error) {
	if !ValidModTypes[in.Type] {
		return nil, errs.ValidationFields(map[string]string{"type": "Unsupported modification type"})
	}
	if in.Amount <= 0 {
		return nil, errs.ValidationFields(map[string]string{"amount": "amount must be positive"})
	}
	sn, err := s.repo.FindScenario(ctx, workspaceID, scenarioID)
	if err != nil {
		return nil, err
	}
	mods, _ := s.repo.ListModifications(ctx, scenarioID)
	m := &Modification{
		ID: uuid.New(), ScenarioID: scenarioID, Type: in.Type, Amount: in.Amount,
		Position: len(mods), Version: 1, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	if in.Frequency != "" {
		f := strings.ToUpper(in.Frequency)
		m.Frequency = &f
	}
	if strings.TrimSpace(in.Narrative) != "" {
		n := strings.TrimSpace(in.Narrative)
		m.Narrative = &n
	}
	if err := s.repo.CreateModification(ctx, m); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "scenario.modification", "scenario_modification", &m.ID, nil)
	s.invalidateResult(ctx, workspaceID, sn)
	v := toModView(m)
	return &v, nil
}

func (s *Service) UpdateModification(ctx context.Context, workspaceID, userID uuid.UUID, scenarioID, modID uuid.UUID, in *ModInput, version int64) (*ModView, error) {
	sn, err := s.repo.FindScenario(ctx, workspaceID, scenarioID)
	if err != nil {
		return nil, err
	}
	m, err := s.repo.FindModification(ctx, scenarioID, modID)
	if err != nil {
		return nil, err
	}
	if in.Amount <= 0 {
		return nil, errs.ValidationFields(map[string]string{"amount": "amount must be positive"})
	}
	m.Type = in.Type
	m.Amount = in.Amount
	m.Version = version
	if in.Frequency != "" {
		f := strings.ToUpper(in.Frequency)
		m.Frequency = &f
	}
	if strings.TrimSpace(in.Narrative) != "" {
		n := strings.TrimSpace(in.Narrative)
		m.Narrative = &n
	}
	if err := s.repo.UpdateModification(ctx, m); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "scenario.modification", "scenario_modification", &m.ID, nil)
	s.invalidateResult(ctx, workspaceID, sn)
	v := toModView(m)
	return &v, nil
}

func (s *Service) RemoveModification(ctx context.Context, workspaceID, userID uuid.UUID, scenarioID, modID uuid.UUID) error {
	if _, err := s.repo.FindScenario(ctx, workspaceID, scenarioID); err != nil {
		return err
	}
	if err := s.repo.DeleteModification(ctx, scenarioID, modID); err != nil {
		return err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "scenario.modification", "scenario_modification", &modID, nil)
	return nil
}

// invalidateResult marks a calculated scenario stale when its inputs change.
func (s *Service) invalidateResult(ctx context.Context, workspaceID uuid.UUID, sn *Scenario) {
	if sn.Status != StatusCalculated || sn.IsStale {
		return
	}
	sn.IsStale = true
	_ = s.repo.UpdateScenario(ctx, sn)
}

func (s *Service) refreshStale(ctx context.Context, workspaceID uuid.UUID, sn *Scenario) error {
	if sn.Status != StatusCalculated || sn.IsStale || sn.BaselineSnapshot == "" {
		return nil
	}
	now := time.Now().UTC()
	base, err := s.fc.Compute(ctx, workspaceID, 90, now)
	if err != nil {
		return nil
	}
	var prev map[string]any
	if err := unmarshal(sn.BaselineSnapshot, &prev); err != nil {
		return nil
	}
	if prevEnding, ok := prev["ending"].(string); ok && prevEnding != base.EndingBalance {
		sn.IsStale = true
		_ = s.repo.UpdateScenario(ctx, sn)
	}
	return nil
}

// apply computes scenario deltas deterministically over the baseline timeline.
func (s *Service) apply(base *forecast.Result, mods []Modification, horizon int, now time.Time) ResultView {
	asOf := atDate(now)
	delts := make([]int64, horizon)
	incomeAdd := int64(0)
	expenseAdd := int64(0)

	indexFor := func(dateStr string) int {
		d, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return 0
		}
		idx := int(atDate(d).Sub(asOf).Hours()/24 - 1)
		if idx < 0 || idx >= horizon {
			return 0
		}
		return idx
	}

	for _, m := range mods {
		switch m.Type {
		case ModOneTimeExpense:
			delts[0] -= m.Amount
			expenseAdd += m.Amount
		case ModOneTimeIncome:
			delts[0] += m.Amount
			incomeAdd += m.Amount
		case ModRecurringExpense, ModRecurringIncome:
			freq := "MONTHLY"
			if m.Frequency != nil && *m.Frequency != "" {
				freq = *m.Frequency
			}
			horizonEnd := asOf.AddDate(0, 0, horizon)
			for _, d := range recurring.OccurrenceDates(freq, asOf.AddDate(0, 0, 1), nil, horizonEnd) {
				idx := indexFor(d.Format("2006-01-02"))
				if m.Type == ModRecurringExpense {
					delts[idx] -= m.Amount
					expenseAdd += m.Amount
				} else {
					delts[idx] += m.Amount
					incomeAdd += m.Amount
				}
			}
		case ModIncomeReduction, ModIncomeRemoval:
			var idy int
			var amt int64
			if m.Type == ModIncomeRemoval {
				idy, amt = pickLargestEvent(base.Events, "INCOME", indexFor)
			} else {
				idy, _ = firstEvent(base.Events, "INCOME", indexFor)
			}
			if idy >= 0 {
				if m.Type == ModIncomeRemoval {
					delts[idy] -= amt
					incomeAdd += amt
				} else {
					delts[idy] -= m.Amount
					incomeAdd += m.Amount
				}
			}
		case ModExpenseReduction:
			if idy, _ := firstEvent(base.Events, "EXPENSE", indexFor); idy >= 0 {
				delts[idy] += m.Amount
				expenseAdd += m.Amount
			}
		}
	}

	scenarioTimeline := make([]DayBalance, len(base.Timeline))
	cum := int64(0)
	scenarioMin := int64(0)
	scenarioEnding := int64(0)
	first := true
	for i, p := range base.Timeline {
		if i < len(delts) {
			cum += delts[i]
		}
		bal := parseMinor(p.Balance) + cum
		scenarioTimeline[i] = DayBalance{Date: p.Date, BaselineBalance: p.Balance, ScenarioBalance: money.FormatMinorUnits(bal)}
		if first || bal < scenarioMin {
			scenarioMin = bal
			first = false
		}
		scenarioEnding = bal
	}
	cashDiff := incomeAdd - expenseAdd

	return ResultView{
		BaselineEnding:     base.EndingBalance,
		ScenarioEnding:     money.FormatMinorUnits(scenarioEnding),
		BaselineMinimum:    base.MinimumBalance,
		ScenarioMinimum:    money.FormatMinorUnits(scenarioMin),
		BaselineIncome:     base.ProjectedIncome,
		ScenarioIncome:     money.FormatMinorUnits(parseMinor(base.ProjectedIncome) + incomeAdd),
		BaselineExpense:    base.ProjectedExpense,
		ScenarioExpense:    money.FormatMinorUnits(parseMinor(base.ProjectedExpense) + expenseAdd),
		CashflowDifference: money.FormatMinorUnits(cashDiff),
		ModifiedEvents:     len(mods),
		AssumptionNote:     "Deltas apply to the baseline timeline: one-time at day 1, recurring on their scheduled dates, reductions on the first matching event.",
		CalculationVersion: calcVersion,
		Timeline:           scenarioTimeline,
	}
}

func firstEvent(events []forecast.Event, kind string, indexFor func(string) int) (int, int64) {
	for _, e := range events {
		if e.Kind == kind {
			return indexFor(e.Date), parseMinor(e.Amount)
		}
	}
	return -1, 0
}

func pickLargestEvent(events []forecast.Event, kind string, indexFor func(string) int) (int, int64) {
	bestIdx, bestAmt := -1, int64(0)
	for _, e := range events {
		if e.Kind != kind {
			continue
		}
		if amt := parseMinor(e.Amount); amt > bestAmt {
			bestAmt = amt
			bestIdx = indexFor(e.Date)
		}
	}
	return bestIdx, bestAmt
}

func parseMinor(s string) int64 {
	m, _ := money.ParseMinorUnits(s)
	return m
}

func atDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

type View struct {
	ID                 uuid.UUID   `json:"id"`
	Name               string      `json:"name"`
	Description        *string     `json:"description"`
	Status             string      `json:"status"`
	IsStale            bool        `json:"is_stale"`
	Version            int64       `json:"version"`
	CalculationVersion string      `json:"calculation_version"`
	Modifications      []ModView   `json:"modifications"`
	Result             *ResultView `json:"result,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

type ModView struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	Amount    string    `json:"amount"`
	Frequency *string   `json:"frequency"`
	Narrative *string   `json:"narrative"`
	Version   int64     `json:"version"`
	UpdatedAt time.Time `json:"updated_at"`
}
