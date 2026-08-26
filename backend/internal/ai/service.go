package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/accounts"
	"github.com/savio/savio/backend/internal/analytics"
	"github.com/savio/savio/backend/internal/budgets"
	"github.com/savio/savio/backend/internal/forecast"
	"github.com/savio/savio/backend/internal/goals"
	"github.com/savio/savio/backend/internal/platform/ai"
	"github.com/savio/savio/backend/internal/platform/config"
	"github.com/savio/savio/backend/internal/platform/errs"
)

type Service struct {
	cfg       *config.Config
	analytics *analytics.Service
	db        *gorm.DB
}

func NewService(db *gorm.DB, cfg *config.Config) *Service {
	// Seed the singleton row from env defaults once so the Settings page can
	// take over as the source of truth (AGENTS: config replacement, #79).
	_ = seedSettings(db, cfg)
	return &Service{cfg: cfg, analytics: analytics.NewService(db), db: db}
}

// enabled reports whether AI endpoints answer or degrade, from runtime settings.
func (s *Service) enabled(ctx context.Context) (bool, error) {
	st, err := s.loadSettings(ctx)
	if err != nil {
		return false, err
	}
	return st.Enabled, nil
}

// complete builds the provider from current runtime settings on every call, so
// AI configuration changes take effect immediately without a restart.
func (s *Service) complete(ctx context.Context, system, prompt string) (map[string]any, error) {
	st, err := s.loadSettings(ctx)
	if err != nil {
		return nil, err
	}
	if !st.Enabled {
		return nil, errs.AIUnavailable("AI is disabled for this installation")
	}
	provider := ai.NewProvider(settingsAdapter{st: st})
	out, err := provider.Complete(ctx, system, prompt)
	if err != nil {
		return nil, errs.AIUnavailable("The AI provider is currently unavailable. Try again shortly.")
	}
	m, err := ai.ExtractJSON(out)
	if err != nil {
		return nil, errs.AIValidation("The AI returned an invalid response and it was rejected.")
	}
	return m, nil
}

// UsesMockProvider reports whether the current configuration resolves to the
// deterministic test/demo provider (no base URL or API key configured). Callers
// that need a trustworthy guess must fall back to deterministic rules instead
// of consuming the mock's canned output.
func (s *Service) UsesMockProvider(ctx context.Context) bool {
	st, err := s.loadSettings(ctx)
	if err != nil {
		return true
	}
	return ai.NewProvider(settingsAdapter{st: st}).Mock()
}

type CategorizeResult struct {
	CategoryGuess string  `json:"category_guess"`
	Confidence    float64 `json:"confidence"`
	MatchedRule   string  `json:"matched_rule"`
}

// Categorize suggests an expense category for raw description/merchant input.
// The AI sees only the candidate category names (context minimization) and
// returns a validated, bounded label — never financial amounts.
func (s *Service) Categorize(ctx context.Context, workspaceID uuid.UUID, description, merchant string) (*CategorizeResult, error) {
	candidates, err := s.workspaceExpenseCategories(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	system := "You classify personal finance transactions into one of the provided categories. Return JSON only: {\"category_guess\": string, \"confidence\": number 0-1}. Choose from the candidate list; if unsure pick the closest or \"Other Expense\"."
	prompt := "CATEGORIZE Candidates: " + strings.Join(candidates, ", ") + "\nDescription: " + truncate(description, 200) + "\nMerchant: " + truncate(merchant, 200)
	m, err := s.complete(ctx, system, prompt)
	if err != nil {
		return nil, err
	}
	guess, err := ai.RequireString(m, "category_guess")
	if err != nil {
		return nil, errs.AIValidation(err.Error())
	}
	conf, _ := ai.RequireFloat(m, "confidence")
	if conf < 0 || conf > 1 {
		conf = 0
	}
	return &CategorizeResult{CategoryGuess: guess, Confidence: conf, MatchedRule: "ai_suggestion"}, nil
}

type InsightResult struct {
	Headline     string   `json:"headline"`
	Detail       string   `json:"detail"`
	Signal       string   `json:"signal"`
	RelatedFacts []string `json:"related_facts"`
}

// Insight generates a bounded explanation from deterministic analytics facts.
func (s *Service) Insight(ctx context.Context, workspaceID uuid.UUID, from, to, prevFrom, prevTo string) (*InsightResult, error) {
	cash, err := s.analytics.Cashflow(ctx, workspaceID, from, to)
	if err != nil {
		return nil, err
	}
	comparison, err := s.analytics.PeriodComparison(ctx, workspaceID, from, to, prevFrom, prevTo)
	if err != nil {
		return nil, err
	}
	cats, err := s.analytics.Categories(ctx, workspaceID, from, to)
	if err != nil {
		return nil, err
	}
	fact := fmt.Sprintf(
		"income=%s expense=%s net=%s prev_income=%s prev_expense=%s income_delta_pct=%v expense_delta_pct=%v top_categories=%v",
		moneyStr(cash.Income), moneyStr(cash.Expense), moneyStr(cash.Net),
		moneyStr(comparison.Previous.Income), moneyStr(comparison.Previous.Expense),
		pct(comparison.IncomeDelta), pct(comparison.ExpenseDelta), topCats(cats),
	)
	system := "You write concise, factual finance insights. JSON only: {\"headline\": string, \"detail\": string, \"signal\": string one of [spending_increase, spending_decrease, income_change, low_balance_risk, stable, other], \"related_facts\": [string]}. Never invent numbers not provided."
	st, err := s.loadSettings(ctx)
	if err != nil {
		return nil, err
	}
	prompt := "INSIGHT Facts: " + fact
	m, err := s.complete(ctx, withPersona(system, st.Persona), prompt)
	if err != nil {
		return nil, err
	}
	headline, err := ai.RequireString(m, "headline")
	if err != nil {
		return nil, errs.AIValidation(err.Error())
	}
	detail, err := ai.RequireString(m, "detail")
	if err != nil {
		return nil, errs.AIValidation(err.Error())
	}
	signal, err := ai.RequireString(m, "signal", "spending_increase", "spending_decrease", "income_change", "low_balance_risk", "stable", "other")
	if err != nil {
		return nil, errs.AIValidation(err.Error())
	}
	var facts []string
	if raw, ok := m["related_facts"].([]any); ok {
		for _, f := range raw {
			if str, ok := f.(string); ok {
				facts = append(facts, str)
			}
		}
	}
	return &InsightResult{Headline: headline, Detail: detail, Signal: signal, RelatedFacts: facts}, nil
}

func (s *Service) forecastService() *forecast.Service { return forecast.NewService(s.db) }
func (s *Service) accountService() *accounts.Service  { return accounts.NewService(s.db) }
func (s *Service) budgetService() *budgets.Service    { return budgets.NewService(s.db) }
func (s *Service) goalService() *goals.Service        { return goals.NewService(s.db) }

type CategoryRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type AccountRef struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	Currency string    `json:"currency"`
}

type EntryContext struct {
	Categories []CategoryRef `json:"categories"`
	Accounts   []AccountRef  `json:"accounts"`
}

type EntryResult struct {
	CategoryGuess string  `json:"category_guess"`
	AccountGuess  string  `json:"account_guess"`
	Confidence    float64 `json:"confidence"`
}

// CategorizeEntry sends the candidate categories and the workspace's active
// accounts as one structured JSON reference so the model can pick both the
// category and the wallet/account an entry belongs to. Only names/types/currency
// leave the DB — no balances, no finance authority (AGENTS #64, #72).
func (s *Service) CategorizeEntry(ctx context.Context, workspaceID uuid.UUID, description, merchant string) (*EntryResult, error) {
	refs, err := s.entryReferences(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(refs)
	if err != nil {
		return nil, errs.AIValidation("failed to encode entry reference")
	}
	system := "You classify a personal finance entry. Pick exactly one category_name and one account_name from the REFERENCE JSON only. Return JSON only: {\"category_guess\": string, \"confidence\": number 0-1, \"account_guess\": string}."
	prompt := "ENTRY REFERENCE (JSON): " + string(raw) +
		"\nDescription: " + truncate(description, 200) +
		"\nMerchant: " + truncate(merchant, 200)
	m, err := s.complete(ctx, system, prompt)
	if err != nil {
		return nil, err
	}
	guess, err := ai.RequireString(m, "category_guess")
	if err != nil {
		return nil, errs.AIValidation(err.Error())
	}
	account, _ := ai.RequireString(m, "account_guess")
	if !refs.hasAccount(account) {
		account = "" // untrusted/nonexistent account → caller falls back to default
	}
	conf, _ := ai.RequireFloat(m, "confidence")
	if conf < 0 || conf > 1 {
		conf = 0
	}
	return &EntryResult{CategoryGuess: guess, AccountGuess: account, Confidence: conf}, nil
}

func (s *Service) entryReferences(ctx context.Context, workspaceID uuid.UUID) (*EntryContext, error) {
	var cats []CategoryRef
	if err := s.db.WithContext(ctx).Table("categories").
		Where("(workspace_id = ? OR is_system = TRUE) AND type = 'EXPENSE' AND status = 'ACTIVE'", workspaceID).
		Order("name ASC").
		Scan(&cats).Error; err != nil {
		return nil, err
	}
	if len(cats) == 0 {
		cats = []CategoryRef{{Name: "Other Expense"}}
	}
	// ponytail: first created account is the alloc-to default; AI may pick any
	// of these but a wrong pick only falls back, it never writes finance state.
	var accs []AccountRef
	if err := s.db.WithContext(ctx).Table("accounts").
		Where("workspace_id = ? AND status = 'ACTIVE'", workspaceID).
		Order("created_at").Scan(&accs).Error; err != nil {
		return nil, err
	}
	return &EntryContext{Categories: cats, Accounts: accs}, nil
}

func (r *EntryContext) hasAccount(name string) bool {
	for _, a := range r.Accounts {
		if a.Name == name {
			return true
		}
	}
	return false
}

func (s *Service) workspaceExpenseCategories(ctx context.Context, workspaceID uuid.UUID) ([]string, error) {
	var names []string
	if err := s.db.WithContext(ctx).Table("categories").
		Where("(workspace_id = ? OR is_system = TRUE) AND type = 'EXPENSE' AND status = 'ACTIVE'", workspaceID).
		Order("name ASC").Pluck("name", &names).Error; err != nil {
		return nil, err
	}
	if len(names) == 0 {
		names = []string{"Other Expense"}
	}
	return names, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func moneyStr(v int64) string { return fmt.Sprintf("%d", v) }

func pct(p *float64) string {
	if p == nil {
		return "n/a"
	}
	return fmt.Sprintf("%.1f%%", *p)
}

func topCats(cats []analytics.CategoryTotal) string {
	var b strings.Builder
	for i, c := range cats {
		if i >= 3 {
			break
		}
		b.WriteString(fmt.Sprintf("%s:%d ", c.CategoryName, c.Total))
	}
	return strings.TrimSpace(b.String())
}

var _ = uuid.Nil
