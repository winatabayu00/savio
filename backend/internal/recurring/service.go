package recurring

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/audit"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/money"
	"github.com/savio/savio/backend/internal/transactions"
)

// ForecastHorizon bounds how far future occurrences are materialized.
const ForecastHorizon = 90 * 24 * time.Hour

// Service owns recurring-rule behavior: schedule generation, lifecycle and
// occurrence confirmation (the only path that writes actual ledger history).
type Service struct {
	db    *gorm.DB
	repo  *Repository
	tx    *transactions.Service
	audit *audit.Repository
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db, repo: NewRepository(db), tx: transactions.NewService(db), audit: audit.NewRepository(db)}
}

type CreateInput struct {
	AccountID   uuid.UUID
	CategoryID  *uuid.UUID
	Type        string
	AmountMinor int64
	Frequency   string
	StartDate   time.Time
	EndDate     *time.Time
	Description string
	Merchant    string
	Notes       string
	AutoPost    bool
}

type UpdateInput struct {
	ID          uuid.UUID
	AccountID   uuid.UUID
	CategoryID  *uuid.UUID
	Type        string
	AmountMinor int64
	Frequency   string
	StartDate   time.Time
	EndDate     *time.Time
	Description string
	Merchant    string
	Notes       string
	AutoPost    bool
	Version     int64
}

type View struct {
	ID           uuid.UUID  `json:"id"`
	AccountID    uuid.UUID  `json:"account_id"`
	CategoryID   *uuid.UUID `json:"category_id"`
	Type         string     `json:"type"`
	Amount       string     `json:"amount"`
	Frequency    string     `json:"frequency"`
	StartDate    string     `json:"start_date"`
	EndDate      *string    `json:"end_date"`
	Description  *string    `json:"description"`
	Merchant     *string    `json:"merchant"`
	Notes        *string    `json:"notes"`
	Status       string     `json:"status"`
	AutoPost     bool       `json:"auto_post"`
	Version      int64      `json:"version"`
	AccountName  string     `json:"account_name"`
	CategoryName string     `json:"category_name"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, in *CreateInput) (*View, error) {
	if err := validate(in.Frequency, in.Type, in.AmountMinor); err != nil {
		return nil, err
	}
	src, err := s.loadAccount(ctx, workspaceID, in.AccountID)
	if err != nil {
		return nil, err
	}
	if src != "ACTIVE" {
		return nil, errs.BusinessConflict("BUSINESS_CONFLICT", "This account is archived and cannot accept recurring activity.")
	}
	if in.CategoryID != nil {
		if err := s.validateCategory(ctx, workspaceID, *in.CategoryID, in.Type); err != nil {
			return nil, err
		}
	}
	now := time.Now().UTC()
	rt := &RecurringTransaction{
		ID:              uuid.New(),
		WorkspaceID:     workspaceID,
		AccountID:       in.AccountID,
		CategoryID:      in.CategoryID,
		Type:            strings.ToUpper(in.Type),
		Amount:          in.AmountMinor,
		Frequency:       strings.ToUpper(in.Frequency),
		StartDate:       atDate(in.StartDate),
		EndDate:         datePtr(in.EndDate),
		Description:     nullableStr(in.Description),
		Merchant:        nullableStr(in.Merchant),
		Notes:           nullableStr(in.Notes),
		Status:          string(StatusActive),
		AutoPost:        in.AutoPost,
		Version:         1,
		CreatedByUserID: &userID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.Create(ctx, rt); err != nil {
		return nil, err
	}
	if err := s.generateUnbound(ctx, workspaceID, rt); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "recurring.create", "recurring", &rt.ID, nil)
	out, err := s.repo.FindByID(ctx, workspaceID, rt.ID)
	if err != nil {
		return nil, err
	}
	return toView(out), nil
}

func (s *Service) Update(ctx context.Context, workspaceID, userID uuid.UUID, in *UpdateInput) (*View, error) {
	existing, err := s.repo.FindByID(ctx, workspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	if existing.Status == string(StatusEnded) {
		return nil, errs.BusinessConflict("BUSINESS_CONFLICT", "Ended recurring transactions cannot be edited.")
	}
	if err := validate(in.Frequency, in.Type, in.AmountMinor); err != nil {
		return nil, err
	}
	if srcStatus, err := s.loadAccount(ctx, workspaceID, in.AccountID); err != nil {
		return nil, err
	} else if srcStatus != "ACTIVE" {
		return nil, errs.BusinessConflict("BUSINESS_CONFLICT", "This account is archived and cannot accept recurring activity.")
	}
	if in.CategoryID != nil {
		if err := s.validateCategory(ctx, workspaceID, *in.CategoryID, in.Type); err != nil {
			return nil, err
		}
	}
	existing.AccountID = in.AccountID
	existing.CategoryID = in.CategoryID
	existing.Type = strings.ToUpper(in.Type)
	existing.Amount = in.AmountMinor
	existing.Frequency = strings.ToUpper(in.Frequency)
	existing.StartDate = atDate(in.StartDate)
	existing.EndDate = datePtr(in.EndDate)
	existing.Description = nullableStr(in.Description)
	existing.Merchant = nullableStr(in.Merchant)
	existing.Notes = nullableStr(in.Notes)
	existing.AutoPost = in.AutoPost
	existing.Version = in.Version
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	if err := s.generateUnbound(ctx, workspaceID, existing); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "recurring.update", "recurring", &in.ID, nil)
	out, err := s.repo.FindByID(ctx, workspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	return toView(out), nil
}

func (s *Service) SetStatus(ctx context.Context, workspaceID, userID uuid.UUID, id uuid.UUID, to Status, version int64) (*View, error) {
	existing, err := s.repo.FindByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	if existing.Status == string(to) {
		return toView(existing), nil
	}
	if existing.Status == string(StatusEnded) {
		return nil, errs.BusinessConflict("BUSINESS_CONFLICT", "Ended recurring transactions cannot be reactivated.")
	}
	if err := s.repo.SetStatus(ctx, workspaceID, id, string(to), version); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "recurring.status", "recurring", &id, map[string]any{"status": string(to)})
	out, err := s.repo.FindByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	return toView(out), nil
}

func (s *Service) List(ctx context.Context, workspaceID uuid.UUID) ([]View, error) {
	rows, _, err := s.repo.List(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make([]View, 0, len(rows))
	for i := range rows {
		out = append(out, *toView(&rows[i]))
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, workspaceID, id uuid.UUID) (*View, error) {
	rt, err := s.repo.FindByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	return toView(rt), nil
}

func (s *Service) Occurrences(ctx context.Context, workspaceID, recurringID uuid.UUID, status, from, to string, page, limit, offset int) ([]OccurrenceView, int64, error) {
	rows, total, err := s.repo.ListOccurrences(ctx, workspaceID, recurringID, status, from, to, page, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	out := make([]OccurrenceView, 0, len(rows))
	for i := range rows {
		out = append(out, *toOccurrenceView(&rows[i]))
	}
	return out, total, nil
}

// Confirm takes a PENDING occurrence and turns it into a POSTED transaction
// atomically (occurrence + ledger update in one DB transaction). The unique
// (recurring_id, due_date) constraint plus the row lock make double-posting
// impossible (INV-010).
func (s *Service) Confirm(ctx context.Context, workspaceID, userID uuid.UUID, id uuid.UUID, version int64) (*OccurrenceView, error) {
	var view *OccurrenceView
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		occ, err := s.repo.LockOccurrenceForUpdate(tx, workspaceID, id)
		if err != nil {
			return err
		}
		if occ.Version != version {
			return errs.VersionConflict("This occurrence changed. Reload the latest version.")
		}
		if occ.Status != string(OccPending) {
			if occ.Status == string(OccConfirmed) {
				return errs.BusinessConflict("BUSINESS_CONFLICT", "This occurrence is already confirmed.")
			}
			return errs.BusinessConflict("BUSINESS_CONFLICT", "This occurrence cannot be confirmed in its current state.")
		}
		rule, err := s.repo.Rule(ctx, occ.RecurringID)
		if err != nil {
			return err
		}
		created, err := s.tx.CreateInTx(ctx, tx, workspaceID, userID, &transactions.CreateInput{
			AccountID:       rule.AccountID,
			CategoryID:      rule.CategoryID,
			Type:            rule.Type,
			AmountMinor:     rule.Amount,
			TransactionDate: occ.DueDate,
			Description:     ruleDescription(rule),
			Source:          "RECURRING",
			Status:          string(transactions.StatusPosted),
		})
		if err != nil {
			return err
		}
		if err := s.repo.MarkOccurrence(tx, occ.ID, string(OccConfirmed), &created.ID); err != nil {
			return err
		}
		view = toOccurrenceView(occ)
		view.Status = string(OccConfirmed)
		view.PostedTransactionID = &created.ID
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "recurring.confirm", "occurrence", &id, nil)
	return view, nil
}

func (s *Service) Skip(ctx context.Context, workspaceID, userID uuid.UUID, id uuid.UUID, version int64) (*OccurrenceView, error) {
	var view *OccurrenceView
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		occ, err := s.repo.LockOccurrenceForUpdate(tx, workspaceID, id)
		if err != nil {
			return err
		}
		if occ.Version != version {
			return errs.VersionConflict("This occurrence changed. Reload the latest version.")
		}
		if occ.Status != string(OccPending) {
			return errs.BusinessConflict("BUSINESS_CONFLICT", "This occurrence can only be skipped while pending.")
		}
		if err := s.repo.MarkOccurrence(tx, occ.ID, string(OccSkipped), nil); err != nil {
			return err
		}
		occ.Status = string(OccSkipped)
		view = toOccurrenceView(occ)
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "recurring.skip", "occurrence", &id, nil)
	return view, nil
}

// generateUnbound materializes upcoming occurrences from the rule's start
// date through today+horizon (or the end date). Existing dates are kept.
func (s *Service) generateUnbound(ctx context.Context, workspaceID uuid.UUID, rt *RecurringTransaction) error {
	horizon := time.Now().UTC().Add(ForecastHorizon)
	dates := nextDates(rt.Frequency, rt.StartDate, rt.EndDate, horizon)
	occs := make([]RecurringOccurrence, 0, len(dates))
	for _, d := range dates {
		occs = append(occs, RecurringOccurrence{
			ID:          uuid.New(),
			RecurringID: rt.ID,
			WorkspaceID: workspaceID,
			DueDate:     d,
			Status:      string(OccPending),
			Version:     1,
		})
	}
	return s.repo.UpsertOccurrences(ctx, occs)
}

// AutoPostDue confirms PENDING due occurrences of ACTIVE auto_post rules.
// Each confirm is atomic and guarded by the (rule, due_date) unique
// constraint, so concurrent workers can never post an occurrence twice
// (INV-010). Returns the number of newly posted occurrences.
func (s *Service) AutoPostDue(ctx context.Context, now time.Time) (int, error) {
	ctx = context.WithoutCancel(ctx)
	var occs []struct {
		ID          uuid.UUID
		WorkspaceID uuid.UUID
		Version     int64
	}
	err := s.db.WithContext(ctx).Raw(`
		SELECT o.id, o.workspace_id, o.version
		FROM recurring_occurrences o
		JOIN recurring_transactions rt ON rt.id = o.recurring_id
		WHERE o.status = 'PENDING'
			AND rt.status = 'ACTIVE' AND rt.auto_post = TRUE
			AND o.due_date <= $1
		ORDER BY o.due_date ASC
		LIMIT 200`, now.Format("2006-01-02")).Scan(&occs).Error
	if err != nil {
		return 0, err
	}
	posted := 0
	var firstErr error
	for _, o := range occs {
		if _, err := s.Confirm(ctx, o.WorkspaceID, uuid.Nil, o.ID, o.Version); err != nil {
			if !isConflict(err) {
				if firstErr == nil {
					firstErr = err
				}
			}
			continue
		}
		posted++
	}
	return posted, firstErr
}

// isConflict reports whether an error is an expected concurrent-race outcome.
func isConflict(err error) bool {
	if err == nil {
		return false
	}
	var appErr *errs.Error
	if errors.As(err, &appErr) {
		return appErr.Code == errs.CodeVersionConflict || appErr.Code == errs.CodeBusinessConflict
	}
	return false
}

func (s *Service) loadAccount(ctx context.Context, workspaceID, accountID uuid.UUID) (string, error) {
	var status string
	err := s.db.WithContext(ctx).Table("accounts").
		Select("status").Where("id = ? AND workspace_id = ?", accountID, workspaceID).Scan(&status).Error
	if err != nil || status == "" {
		return "", errs.NotFound("Account not found")
	}
	return status, nil
}

func (s *Service) validateCategory(ctx context.Context, workspaceID uuid.UUID, categoryID uuid.UUID, txType string) error {
	var c string
	err := s.db.WithContext(ctx).Table("categories").
		Select("type").
		Where("id = ? AND (workspace_id = ? OR is_system = TRUE) AND status = 'ACTIVE'", categoryID, workspaceID).
		Scan(&c).Error
	if err != nil || c == "" {
		return errs.ValidationFields(map[string]string{"category_id": "Category is not available in this workspace"})
	}
	if strings.ToUpper(txType) != c {
		return errs.ValidationFields(map[string]string{"category_id": "Category type does not match"})
	}
	return nil
}

func validate(frequency, typ string, amount int64) error {
	fields := map[string]string{}
	if !ValidFrequency(strings.ToUpper(frequency)) {
		fields["frequency"] = "frequency must be DAILY, WEEKLY, MONTHLY or MONTH_END"
	}
	if typ != "INCOME" && typ != "EXPENSE" {
		fields["type"] = "type must be INCOME or EXPENSE"
	}
	if amount <= 0 {
		fields["amount"] = "Amount must be positive"
	}
	if len(fields) > 0 {
		return errs.ValidationFields(fields)
	}
	return nil
}

func nullableStr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func datePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	u := atDate(*t)
	return &u
}

func ruleDescription(rt *RecurringTransaction) string {
	if rt.Description != nil && strings.TrimSpace(*rt.Description) != "" {
		return *rt.Description
	}
	return "Recurring " + strings.ToLower(rt.Type)
}

func toView(rt *RecurringTransaction) *View {
	out := &View{
		ID:           rt.ID,
		AccountID:    rt.AccountID,
		CategoryID:   rt.CategoryID,
		Type:         rt.Type,
		Amount:       canonicalMoney(rt.Amount),
		Frequency:    rt.Frequency,
		StartDate:    rt.StartDate.Format("2006-01-02"),
		Description:  rt.Description,
		Merchant:     rt.Merchant,
		Notes:        rt.Notes,
		Status:       rt.Status,
		AutoPost:     rt.AutoPost,
		Version:      rt.Version,
		AccountName:  rt.AccountName,
		CategoryName: rt.CategoryName,
		CreatedAt:    rt.CreatedAt,
		UpdatedAt:    rt.UpdatedAt,
	}
	if rt.EndDate != nil {
		s := rt.EndDate.Format("2006-01-02")
		out.EndDate = &s
	}
	return out
}

type OccurrenceView struct {
	ID                  uuid.UUID  `json:"id"`
	RecurringID         uuid.UUID  `json:"recurring_id"`
	DueDate             string     `json:"due_date"`
	Status              string     `json:"status"`
	Version             int64      `json:"version"`
	PostedTransactionID *uuid.UUID `json:"posted_transaction_id"`
	RecurringType       string     `json:"recurring_type"`
	RecurringAmount     string     `json:"recurring_amount"`
	RecurringAccount    string     `json:"recurring_account"`
}

func toOccurrenceView(o *RecurringOccurrence) *OccurrenceView {
	return &OccurrenceView{
		ID:                  o.ID,
		RecurringID:         o.RecurringID,
		DueDate:             o.DueDate.Format("2006-01-02"),
		Status:              o.Status,
		Version:             o.Version,
		PostedTransactionID: o.PostedTransactionID,
		RecurringType:       o.RecurringType,
		RecurringAmount:     canonicalMoney(o.RecurringAmount),
		RecurringAccount:    o.RecurringAccount,
	}
}

func canonicalMoney(minor int64) string {
	return money.FormatMinorUnits(minor)
}
