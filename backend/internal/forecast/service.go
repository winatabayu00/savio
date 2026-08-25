package forecast

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/money"
	"github.com/savio/savio/backend/internal/recurring"
)

// Allowed horizons in days.
var AllowedHorizons = map[int]bool{30: true, 60: true, 90: true, 180: true, 365: true}

// CalculationVersion changes when the methodology materially changes.
const CalculationVersion = "1"

type Event struct {
	Date        string `json:"date"`
	Type        string `json:"type"` // KNOWN | SCHEDULED | ESTIMATED | ASSUMED
	Kind        string `json:"kind"` // INCOME | EXPENSE
	Amount      string `json:"amount"`
	Description string `json:"description"`
}

type DayPoint struct {
	Date    string `json:"date"`
	Balance string `json:"balance"`
}

type Assumptions struct {
	VariableExpenseDaily string `json:"variable_expense_daily"`
	BaselineDays         int    `json:"baseline_days"`
	ActiveRecurringRules int    `json:"active_recurring_rules"`
	ConfidenceBasis      string `json:"confidence_basis"`
}

type Result struct {
	OpeningBalance     string       `json:"opening_balance"`
	EndingBalance      string       `json:"ending_balance"`
	MinimumBalance     string       `json:"minimum_balance"`
	MinimumBalanceDate string       `json:"minimum_balance_date"`
	ProjectedIncome    string       `json:"projected_income"`
	ProjectedExpense   string       `json:"projected_expense"`
	Timeline           []DayPoint   `json:"timeline"`
	Events             []Event      `json:"events"`
	Confidence         string       `json:"confidence"`
	Assumptions        Assumptions  `json:"assumptions"`
	CalculationVersion string       `json:"calculation_version"`
	Stale              bool         `json:"stale"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// Compute generates a deterministic cashflow forecast for the next horizon.
// Inputs: current liquid balance, active recurring rules (SCHEDULED),
// future-dated posted transactions (KNOWN) and trailing-90-day average
// variable expense (ESTIMATED). No LLM involvement (AGENTS #54).
func (s *Service) Compute(ctx context.Context, workspaceID uuid.UUID, horizonDays int, now time.Time) (*Result, error) {
	now = now.UTC()
	asOf := atDate(now)
	horizon := asOf.AddDate(0, 0, horizonDays)

	opening, err := s.liquidBalance(ctx, workspaceID, asOf)
	if err != nil {
		return nil, err
	}

	var events []Event

	// SCHEDULED: active recurring rules within the horizon.
	var rules []struct {
		ID          uuid.UUID
		Type        string
		Amount      int64
		Frequency   string
		StartDate   string
		EndDate     string
		HasEnd      bool
		Description *string
	}
	if err := s.db.WithContext(ctx).Table("recurring_transactions").
		Select("id, type, amount, frequency, start_date::text, COALESCE(end_date::text,'') AS end_date, end_date IS NOT NULL AS has_end, description").
		Where("workspace_id = ? AND status = 'ACTIVE'", workspaceID).
		Scan(&rules).Error; err != nil {
		return nil, err
	}
	for _, r := range rules {
		start, _ := time.Parse("2006-01-02", r.StartDate)
		var end *time.Time
		if r.HasEnd {
			if e, err := time.Parse("2006-01-02", r.EndDate); err == nil {
				ee := time.Date(e.Year(), e.Month(), e.Day(), 0, 0, 0, 0, time.UTC)
				end = &ee
			}
		}
		for _, d := range recurring.OccurrenceDates(r.Frequency, start, end, horizon) {
			if d.Before(asOf) || d.After(horizon) {
				continue
			}
			desc := "Recurring " + strings.ToLower(r.Type)
			if r.Description != nil && *r.Description != "" {
				desc = *r.Description
			}
			events = append(events, Event{
				Date: d.Format("2006-01-02"), Type: "SCHEDULED", Kind: r.Type,
				Amount: money.FormatMinorUnits(r.Amount), Description: desc,
			})
		}
	}

	// KNOWN: future-dated posted transactions already on the ledger.
	type knownEvent struct {
		Date   string
		Type   string
		Amount int64
		Desc   *string
	}
	var known []knownEvent
	if err := s.db.WithContext(ctx).Raw(`
		SELECT transaction_date::text AS date, type, amount, description
		FROM transactions
		WHERE workspace_id = $1 AND status = 'POSTED'
			AND transaction_date > $2 AND transaction_date <= $3
			AND type IN ('INCOME', 'EXPENSE')`,
		workspaceID, asOf.Format("2006-01-02"), horizon.Format("2006-01-02")).Scan(&known).Error; err != nil {
		return nil, err
	}
	for _, k := range known {
		desc := ""
		if k.Desc != nil {
			desc = *k.Desc
		}
		events = append(events, Event{
			Date: k.Date, Type: "KNOWN", Kind: k.Type,
			Amount: money.FormatMinorUnits(k.Amount), Description: desc,
		})
	}

	// ESTIMATED: variable expense baseline from trailing history.
	variableDaily, baselineDays, err := s.variableBaseline(ctx, workspaceID, now)
	if err != nil {
		return nil, err
	}
	if variableDaily > 0 {
		for i := 0; i < horizonDays; i++ {
			d := asOf.AddDate(0, 0, i+1)
			events = append(events, Event{
				Date: d.Format("2006-01-02"), Type: "ESTIMATED", Kind: "EXPENSE",
				Amount: money.FormatMinorUnits(variableDaily), Description: "Estimated daily spending",
			})
		}
	}

	// Roll the balance forward day-by-day.
	balance := opening
	minBalance := opening
	minDate := asOf
	dayActuals := map[string]int64{} // date -> net effect in minor units
	projectedIncome := int64(0)
	projectedExpense := int64(0)
	for _, e := range events {
		amt := parseMinor(e.Amount)
		net := amt
		if e.Kind == "EXPENSE" {
			net = -amt
		}
		dayActuals[e.Date] += net
		if e.Kind == "EXPENSE" {
			projectedExpense += amt
		} else {
			projectedIncome += amt
		}
	}

	timeline := make([]DayPoint, 0, horizonDays)
	for i := 0; i < horizonDays; i++ {
		d := asOf.AddDate(0, 0, i+1)
		key := d.Format("2006-01-02")
		balance += dayActuals[key]
		if balance < minBalance {
			minBalance = balance
			minDate = d
		}
		timeline = append(timeline, DayPoint{Date: key, Balance: money.FormatMinorUnits(balance)})
	}

	confidence := confidenceFor(baselineDays)

	res := &Result{
		OpeningBalance:     money.FormatMinorUnits(opening),
		EndingBalance:      money.FormatMinorUnits(balance),
		MinimumBalance:     money.FormatMinorUnits(minBalance),
		MinimumBalanceDate: minDate.Format("2006-01-02"),
		ProjectedIncome:    money.FormatMinorUnits(projectedIncome),
		ProjectedExpense:   money.FormatMinorUnits(projectedExpense),
		Timeline:           timeline,
		Events:             events,
		Confidence:         confidence,
		Assumptions: Assumptions{
			VariableExpenseDaily: money.FormatMinorUnits(variableDaily),
			BaselineDays:         baselineDays,
			ActiveRecurringRules: len(rules),
			ConfidenceBasis:      "trailing history + recurring coverage",
		},
		CalculationVersion: CalculationVersion,
	}
	return res, nil
}

// liquidBalance is the realized liquid balance as of today: opening balanced
// against POSTED activity dated up to asOf. Future-dated POSTED transactions
// are intentionally excluded here and surfaced instead as KNOWN events, so
// they are never double-counted.
func (s *Service) liquidBalance(ctx context.Context, workspaceID uuid.UUID, asOf time.Time) (int64, error) {
	var total int64
	err := s.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(a.opening_balance), 0)
			+ COALESCE((SELECT SUM(CASE WHEN t.type = 'EXPENSE' THEN -t.amount ELSE t.amount END)
				FROM transactions t
				WHERE t.workspace_id = $1 AND t.status = 'POSTED' AND t.transaction_date <= $2), 0)
		FROM accounts a WHERE a.workspace_id = $1 AND a.status = 'ACTIVE'`,
		workspaceID, asOf.Format("2006-01-02")).Scan(&total).Error
	return total, err
}

// variableBaseline returns the trailing daily average EXPENSE and how many
// calendar days of history exist (capped at 90) for confidence scoring.
func (s *Service) variableBaseline(ctx context.Context, workspaceID uuid.UUID, now time.Time) (int64, int, error) {
	from := now.AddDate(0, 0, -90)
	var total int64
	if err := s.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(amount), 0) FROM transactions
		WHERE workspace_id = $1 AND status = 'POSTED' AND type = 'EXPENSE'
			AND transaction_date >= $2 AND transaction_date <= $3`,
		workspaceID, from.Format("2006-01-02"), now.Format("2006-01-02")).Scan(&total).Error; err != nil {
		return 0, 0, err
	}
	var firstStr string
	if err := s.db.WithContext(ctx).Raw(`
		SELECT COALESCE(MIN(transaction_date)::text, '') FROM transactions
		WHERE workspace_id = $1 AND status = 'POSTED' AND type = 'EXPENSE'`,
		workspaceID).Scan(&firstStr).Error; err != nil {
		return 0, 0, err
	}
	days := 0
	if firstStr != "" {
		if first, err := time.Parse("2006-01-02", firstStr); err == nil {
			days = int(now.Sub(atDate(first)).Hours() / 24)
			if days < 0 {
				days = 0
			}
			if days > 90 {
				days = 90
			}
		}
	}
	if days < 1 {
		days = 1
	}
	return total / int64(days), days, nil
}

func confidenceFor(baselineDays int) string {
	switch {
	case baselineDays < 30:
		return "LOW"
	case baselineDays < 90:
		return "MEDIUM"
	default:
		return "HIGH"
	}
}

func atDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func parseMinor(s string) int64 {
	m, _ := money.ParseMinorUnits(s)
	return m
}