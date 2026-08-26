package analytics

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/accounts"
	"github.com/savio/savio/backend/internal/platform/money"
	"github.com/savio/savio/backend/internal/recurring"
)

// CashflowView is income/expense for one period.
type CashflowView struct {
	Income  int64 `json:"income"`
	Expense int64 `json:"expense"`
	Net     int64 `json:"net"`
}

type CategoryTotal struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	Total        int64  `json:"total"`
	Items        int64  `json:"items"`
}

type PeriodComparisonView struct {
	Current       CashflowView `json:"current"`
	Previous      CashflowView `json:"previous"`
	IncomeDelta   *float64     `json:"income_delta_percent"`
	ExpenseDelta  *float64     `json:"expense_delta_percent"`
}

type SpendingChange struct {
	CategoryID   string   `json:"category_id"`
	CategoryName string   `json:"category_name"`
	Current      int64    `json:"current"`
	Previous     int64    `json:"previous"`
	Delta        int64    `json:"delta"`
	DeltaPercent *float64 `json:"delta_percent"`
}

type RecurringExpenseSummary struct {
	ActiveRules     int64 `json:"active_rules"`
	MonthlyEstimate int64 `json:"estimated_monthly"`
}

// UpcomingOccurrence is a PENDING scheduled instance within the lookahead.
type UpcomingOccurrence struct {
	ID              uuid.UUID `json:"id"`
	RecurringID     uuid.UUID `json:"recurring_id"`
	DueDate         string    `json:"due_date"`
	Type            string    `json:"type"`
	Amount          int64     `json:"-"`
	AmountString    string    `json:"amount"`
	AccountName     string    `json:"account_name"`
	Description     *string   `json:"description"`
}

// RecentRow is a ledger row for the dashboard feed.
type RecentRow struct {
	ID           uuid.UUID `json:"id"`
	Type         string    `json:"type"`
	Amount       int64     `json:"-"`
	AmountString string    `json:"amount"`
	TxDate       string    `json:"transaction_date"`
	Description  *string   `json:"description"`
	AccountName  string    `json:"account_name"`
	CategoryName string    `json:"category_name"`
	Status       string    `json:"status"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// cashflow aggregates POSTED INCOME/EXPENSE in a date range (excludes
// transfers, VOIDED and ADJUSTMENT per AGENTS #35).
func (s *Service) cashflow(ctx context.Context, workspaceID uuid.UUID, from, to string) (*CashflowView, error) {
	var row struct {
		Income  int64
		Expense int64
	}
	err := s.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN type = 'INCOME' THEN amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN type = 'EXPENSE' THEN amount ELSE 0 END), 0) AS expense
		FROM transactions
		WHERE workspace_id = $1 AND status = 'POSTED'
			AND type IN ('INCOME', 'EXPENSE')
			AND transaction_date BETWEEN $2 AND $3`,
		workspaceID, from, to).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	return &CashflowView{Income: row.Income, Expense: row.Expense, Net: row.Income - row.Expense}, nil
}

func (s *Service) Cashflow(ctx context.Context, workspaceID uuid.UUID, from, to string) (*CashflowView, error) {
	return s.cashflow(ctx, workspaceID, from, to)
}

func (s *Service) Categories(ctx context.Context, workspaceID uuid.UUID, from, to string) ([]CategoryTotal, error) {
	var rows []CategoryTotal
	err := s.db.WithContext(ctx).Raw(`
		SELECT COALESCE(c.id::text, '') AS category_id,
		       COALESCE(c.name, 'Uncategorized') AS category_name,
		       SUM(t.amount) AS total,
		       COUNT(*) AS items
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.workspace_id = $1 AND t.status = 'POSTED'
			AND t.type IN ('INCOME', 'EXPENSE')
			AND t.transaction_date BETWEEN $2 AND $3
		GROUP BY c.id, c.name
		ORDER BY total DESC`,
		workspaceID, from, to).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) PeriodComparison(ctx context.Context, workspaceID uuid.UUID, from, to, prevFrom, prevTo string) (*PeriodComparisonView, error) {
	cur, err := s.cashflow(ctx, workspaceID, from, to)
	if err != nil {
		return nil, err
	}
	prev, err := s.cashflow(ctx, workspaceID, prevFrom, prevTo)
	if err != nil {
		return nil, err
	}
	out := &PeriodComparisonView{Current: *cur, Previous: *prev}
	out.IncomeDelta = deltaPercent(prev.Income, cur.Income)
	out.ExpenseDelta = deltaPercent(prev.Expense, cur.Expense)
	return out, nil
}

func (s *Service) SpendingChanges(ctx context.Context, workspaceID uuid.UUID, from, to, prevFrom, prevTo string) ([]SpendingChange, error) {
	var rows []SpendingChange
	err := s.db.WithContext(ctx).Raw(`
		SELECT COALESCE(c.id::text, '') AS category_id,
		       COALESCE(c.name, 'Uncategorized') AS category_name,
		       COALESCE(SUM(cur.amount) FILTER (WHERE cur.period = 'cur'), 0) AS current,
		       COALESCE(SUM(cur.amount) FILTER (WHERE cur.period = 'prev'), 0) AS previous
		FROM (
			SELECT t.category_id, t.amount,
			       CASE WHEN t.transaction_date BETWEEN $4 AND $5 THEN 'cur' ELSE 'prev' END AS period
			FROM transactions t
			WHERE t.workspace_id = $1 AND t.status = 'POSTED'
				AND t.type IN ('INCOME', 'EXPENSE')
				AND (t.transaction_date BETWEEN $2 AND $3 OR t.transaction_date BETWEEN $4 AND $5)
		) cur
		LEFT JOIN categories c ON c.id = cur.category_id
		GROUP BY c.id, c.name
		ORDER BY current DESC`,
		workspaceID, from, to, prevFrom, prevTo).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].Delta = rows[i].Current - rows[i].Previous
		rows[i].DeltaPercent = deltaPercent(rows[i].Previous, rows[i].Current)
	}
	return rows, nil
}

func (s *Service) RecurringExpenses(ctx context.Context, workspaceID uuid.UUID) (*RecurringExpenseSummary, error) {
	var rules []struct {
		Amount    int64
		Frequency string
	}
	if err := s.db.WithContext(ctx).Table("recurring_transactions").
		Select("amount, frequency").
		Where("workspace_id = ? AND type = 'EXPENSE' AND status = 'ACTIVE'", workspaceID).
		Scan(&rules).Error; err != nil {
		return nil, err
	}
	out := &RecurringExpenseSummary{ActiveRules: int64(len(rules))}
	for _, r := range rules {
		switch r.Frequency {
		case string(recurring.FreqDaily):
			out.MonthlyEstimate += r.Amount * 30
		case string(recurring.FreqWeekly):
			out.MonthlyEstimate += r.Amount * 4
		default:
			out.MonthlyEstimate += r.Amount
		}
	}
	return out, nil
}

func deltaPercent(prev, cur int64) *float64 {
	if prev == 0 {
		return nil
	}
	d := (float64(cur-prev) / float64(prev)) * 100
	return &d
}

// DashboardView assembles the front-page financial picture.
type DashboardView struct {
	TotalBalance int64                  `json:"total_balance"`
	Accounts     []accounts.View        `json:"accounts"`
	Cashflow     *CashflowView          `json:"cashflow"`
	Upcoming     []UpcomingOccurrence   `json:"upcoming"`
	Recent       []RecentRow            `json:"recent"`
}

// Dashboard computes the balance summary, current cashflow and upcoming
// scheduled activity. Modules added later (budgets, goals, forecast, AI) plug
// in separately.
func (s *Service) Dashboard(ctx context.Context, workspaceID uuid.UUID, now time.Time) (*DashboardView, error) {
	var total int64
	err := s.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(a.opening_balance), 0)
			+ COALESCE((SELECT SUM(CASE WHEN t.type = 'EXPENSE' THEN -t.amount ELSE t.amount END)
				FROM transactions t WHERE t.workspace_id = $1 AND t.status = 'POSTED'), 0)
		FROM accounts a WHERE a.workspace_id = $1`, workspaceID).Scan(&total).Error
	if err != nil {
		return nil, err
	}

	accViews, _, err := accounts.NewService(s.db).List(ctx, workspaceID, "", "", "created_at", false, 1, 100, 0)
	if err != nil {
		return nil, err
	}

	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, -1)
	cash, err := s.cashflow(ctx, workspaceID, monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	var upcoming []UpcomingOccurrence
	err = s.db.WithContext(ctx).Raw(`
		SELECT o.id, o.recurring_id, o.due_date::text, rt.type, rt.amount, a.name AS account_name, rt.description
		FROM recurring_occurrences o
		JOIN recurring_transactions rt ON rt.id = o.recurring_id
		JOIN accounts a ON a.id = rt.account_id
		WHERE o.workspace_id = $1 AND o.status = 'PENDING'
			AND o.due_date BETWEEN $2 AND $3
		ORDER BY o.due_date ASC
		LIMIT 20`, workspaceID, now.Format("2006-01-02"), now.AddDate(0, 0, 30).Format("2006-01-02")).
		Scan(&upcoming).Error
	if err != nil {
		return nil, err
	}
	for i := range upcoming {
		upcoming[i].AmountString = money.FormatMinorUnits(upcoming[i].Amount)
	}

	var recent []RecentRow
	err = s.db.WithContext(ctx).Raw(`
		SELECT t.id, t.type, t.amount, t.transaction_date::text, t.description,
		       a.name AS account_name, COALESCE(c.name, '') AS category_name, t.status
		FROM transactions t
		JOIN accounts a ON a.id = t.account_id
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.workspace_id = $1
		ORDER BY t.transaction_date DESC, t.created_at DESC
		LIMIT 8`, workspaceID).Scan(&recent).Error
	if err != nil {
		return nil, err
	}
	for i := range recent {
		recent[i].AmountString = money.FormatMinorUnits(recent[i].Amount)
	}

	return &DashboardView{
		TotalBalance: total,
		Accounts:     accViews,
		Cashflow:     cash,
		Upcoming:     upcoming,
		Recent:       recent,
	}, nil
}