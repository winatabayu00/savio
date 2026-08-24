package ai

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/money"
	platformai "github.com/savio/savio/backend/internal/platform/ai"
)

type Fact struct {
	Tool  string `json:"tool"`
	Label string `json:"label"`
	Value string `json:"value"`
}

type CopilotResult struct {
	Answer        string   `json:"answer"`
	Facts         []Fact   `json:"facts"`
	ToolUsed      string   `json:"tool_used"`
	Sources       []string `json:"sources"`
	Actions       []string `json:"actions"`
	Clarification *string  `json:"clarification,omitempty"`
}

var amountRe = regexp.MustCompile(`(?i)(\d[\d.,]*)\s*(m|jt|juta|rb|k|ribu)?`)

// Copilot routes a natural-language question to bounded deterministic tools,
// builds minimal context, and asks the model for a grounded explanation.
// The LLM never chooses workspace/user identity or executes beyond these
// tools (AGENTS #68-71).
func (s *Service) Copilot(ctx context.Context, workspaceID uuid.UUID, question string, horizon int, now time.Time) (*CopilotResult, error) {
	q := strings.ToLower(strings.TrimSpace(question))
	if q == "" {
		return nil, errs.ValidationFields(map[string]string{"question": "Ask something about your finances."})
	}

	var facts []Fact
	tool := s.routeIntent(ctx, workspaceID, q, horizon, now, &facts)
	if tool.validation != "" {
		return nil, errs.ValidationFields(map[string]string{"question": tool.validation})
	}

	// minimal context: only aggregated facts, never raw ledger
	var sb strings.Builder
	sb.WriteString("COPILOT Facts:\n")
	for _, f := range facts {
		sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", f.Label, f.Tool, f.Value))
	}
	sb.WriteString("Question: ")
	sb.WriteString(truncate(question, 500))

	out, err := s.complete(ctx, "You are Savio Copilot, a grounded finance assistant. Answer using ONLY the provided facts. JSON only: {\"answer\": string, \"actions\": [string], \"clarification\": string or null}. Do not invent numbers.", sb.String())
	if err != nil {
		return nil, err
	}
	answer, err := platformai.RequireString(out, "answer")
	if err != nil {
		return nil, errs.AIValidation(err.Error())
	}
	var actions []string
	if raw, ok := out["actions"].([]any); ok {
		for _, a := range raw {
			if str, ok := a.(string); ok {
				actions = append(actions, str)
			}
		}
	}
	var clarification *string
	if raw, ok := out["clarification"].(string); ok && raw != "" {
		clarification = &raw
	}

	res := &CopilotResult{
		Answer:        answer,
		Facts:         facts,
		ToolUsed:      tool.name,
		Sources:       tool.sources,
		Actions:       actions,
		Clarification: clarification,
	}
	return res, nil
}

type intentResult struct {
	name       string
	sources    []string
	validation string
}

func (s *Service) routeIntent(ctx context.Context, workspaceID uuid.UUID, q string, horizon int, now time.Time, facts *[]Fact) intentResult {
	appendFact := func(tool, label, value string) {
		*facts = append(*facts, Fact{Tool: tool, Label: label, Value: value})
	}

	monthFrom := now.AddDate(0, 0, -now.Day()+1).Format("2006-01-02")
	monthTo := now.AddDate(0, 1, -now.Day()).Format("2006-01-02")
	prevFrom := now.AddDate(0, -1, -now.Day()+1).Format("2006-01-02")
	prevTo := now.AddDate(0, 0, -now.Day()).Format("2006-01-02")

	switch {
	case strings.Contains(q, "forecast") || strings.Contains(q, "projection") || strings.Contains(q, "future"):
		f, err := s.forecastService().Compute(ctx, workspaceID, horizon, now)
		if err != nil {
			return intentResult{validation: "Could not compute forecast."}
		}
		appendFact("get_forecast", "Ending balance ("+strconv.Itoa(horizon)+"d)", f.EndingBalance)
		appendFact("get_forecast", "Minimum balance", f.MinimumBalance+" on "+f.MinimumBalanceDate)
		appendFact("get_forecast", "Projected income", f.ProjectedIncome)
		appendFact("get_forecast", "Projected expense", f.ProjectedExpense)
		appendFact("get_forecast", "Confidence", f.Confidence)
		return intentResult{name: "get_forecast", sources: []string{"get_forecast"}}

	case strings.Contains(q, "budget"):
		rows, err := s.budgetService().List(ctx, workspaceID, "", 80)
		if err != nil {
			return intentResult{validation: "Could not load budgets."}
		}
		for _, b := range rows {
			appendFact("get_budget_status", b.CategoryName, "spent "+b.Spent+" of "+b.Amount+" ("+fmt.Sprintf("%.0f", b.Utilization)+"%, "+b.ComputedStatus+")")
		}
		return intentResult{name: "get_budget_status", sources: []string{"get_budget_status"}}

	case strings.Contains(q, "goal") || strings.Contains(q, "on track"):
		rows, err := s.goalService().List(ctx, workspaceID, "")
		if err != nil {
			return intentResult{validation: "Could not load goals."}
		}
		for _, g := range rows {
			appendFact("get_goal_status", g.Name, "progress "+fmt.Sprintf("%.0f", g.ProgressPercent)+"%, required "+g.RequiredMonthly+"/month, "+g.Feasibility)
		}
		return intentResult{name: "get_goal_status", sources: []string{"get_goal_status"}}

	case strings.Contains(q, "recurring") || strings.Contains(q, "subscription"):
		summary, err := s.analytics.RecurringExpenses(ctx, workspaceID)
		if err != nil {
			return intentResult{validation: "Could not analyze recurring expenses."}
		}
		appendFact("get_recurring_expenses", "Active recurring expense rules", strconv.FormatInt(summary.ActiveRules, 10))
		appendFact("get_recurring_expenses", "Estimated monthly expense", fmt.Sprintf("%d", summary.MonthlyEstimate))
		return intentResult{name: "get_recurring_expenses", sources: []string{"get_recurring_expenses"}}

	case strings.Contains(q, "compare") || strings.Contains(q, "more this month") || strings.Contains(q, "than last month"):
		c, err := s.analytics.PeriodComparison(ctx, workspaceID, monthFrom, monthTo, prevFrom, prevTo)
		if err != nil {
			return intentResult{validation: "Could not compare periods."}
		}
		appendFact("compare_periods", "Income this month", money.FormatMinorUnits(c.Current.Income))
		appendFact("compare_periods", "Income last month", money.FormatMinorUnits(c.Previous.Income))
		appendFact("compare_periods", "Expense this month", money.FormatMinorUnits(c.Current.Expense))
		appendFact("compare_periods", "Expense last month", money.FormatMinorUnits(c.Previous.Expense))
		return intentResult{name: "compare_periods", sources: []string{"compare_periods"}}

	case strings.Contains(q, "afford") || strings.Contains(q, "if i buy") || strings.Contains(q, "if i ") || strings.Contains(q, "what if"):
		if amt := extractAmount(q); amt > 0 {
			f, err := s.forecastService().Compute(ctx, workspaceID, horizon, now)
			if err != nil {
				return intentResult{validation: "Could not compute forecast."}
			}
			end := parseMinorInt(f.EndingBalance)
			appendFact("calculate_scenario", "Projected ending balance (baseline)", f.EndingBalance)
			appendFact("calculate_scenario", "One-time impact", money.FormatMinorUnits(amt))
			appendFact("calculate_scenario", "Projected ending balance (after impact)", money.FormatMinorUnits(end-amt))
			return intentResult{name: "calculate_scenario", sources: []string{"get_forecast", "calculate_scenario"}}
		}
		fallthrough

	default:
		cash, err := s.analytics.Cashflow(ctx, workspaceID, monthFrom, monthTo)
		if err != nil {
			return intentResult{validation: "Could not compute cashflow."}
		}
		appendFact("get_cashflow_summary", "Income", money.FormatMinorUnits(cash.Income))
		appendFact("get_cashflow_summary", "Expenses", money.FormatMinorUnits(cash.Expense))
		appendFact("get_cashflow_summary", "Net", money.FormatMinorUnits(cash.Net))
		cats, err := s.analytics.Categories(ctx, workspaceID, monthFrom, monthTo)
		if err == nil {
			for i, c := range cats {
				if i >= 3 {
					break
				}
				appendFact("get_category_breakdown", c.CategoryName, money.FormatMinorUnits(c.Total))
			}
		}
		if strings.Contains(q, "where did") || strings.Contains(q, "money go") {
			return intentResult{name: "get_cashflow_summary", sources: []string{"get_cashflow_summary", "get_category_breakdown"}}
		}
		return intentResult{name: "get_cashflow_summary", sources: []string{"get_cashflow_summary", "get_category_breakdown"}}
	}
}

// extractAmount pulls a rough major-unit amount from a question ("15M laptop").
func extractAmount(q string) int64 {
	m := amountRe.FindStringSubmatch(q)
	if m == nil {
		return 0
	}
	var num float64
	clean := strings.ReplaceAll(strings.ReplaceAll(m[1], ",", ""), ".", "")
	if parts := strings.SplitN(m[1], ".", 2); len(parts) == 2 {
		natural, _ := strconv.ParseFloat(strings.ReplaceAll(parts[0], ",", ""), 64)
		frac, _ := strconv.ParseFloat(parts[1], 64)
		num = natural + frac/100
	} else {
		num, _ = strconv.ParseFloat(clean, 64)
	}
	switch strings.ToLower(m[2]) {
	case "m":
		num *= 1_000_000
	case "jt", "juta":
		num *= 1_000_000
	case "k", "rb", "ribu":
		num *= 1_000
	}
	return int64(num * 100)
}

func parseMinorInt(s string) int64 {
	v, _ := money.ParseMinorUnits(s)
	return v
}