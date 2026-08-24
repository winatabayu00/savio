package transactions

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
)

// Service owns authoritative ledger behavior: validation, lifecycle,
// financial-invariant guards and audit recording.
type Service struct {
	db    *gorm.DB
	repo  *Repository
	audit *audit.Repository
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db, repo: NewRepository(db), audit: audit.NewRepository(db)}
}

type CreateInput struct {
	AccountID       uuid.UUID
	CategoryID      *uuid.UUID
	Type            string
	AmountMinor     int64
	TransactionDate time.Time
	Description     string
	Merchant        string
	Notes           string
	Source          string
	Status          string
}

type UpdateInput struct {
	ID              uuid.UUID
	CategoryID      *uuid.UUID
	Type            string
	AmountMinor     int64
	TransactionDate time.Time
	Description     string
	Merchant        string
	Notes           string
	Version         int64
}

type VoidInput struct {
	ID      uuid.UUID
	Reason  string
	Version int64
}

type View struct {
	ID              uuid.UUID  `json:"id"`
	AccountID       uuid.UUID  `json:"account_id"`
	CategoryID      *uuid.UUID `json:"category_id"`
	Type            string     `json:"type"`
	Amount          string     `json:"amount"`
	TransactionDate string     `json:"transaction_date"`
	Description     *string    `json:"description"`
	Merchant        *string    `json:"merchant"`
	Notes           *string    `json:"notes"`
	Source          string     `json:"source"`
	Status          string     `json:"status"`
	Version         int64      `json:"version"`
	AccountName     string     `json:"account_name"`
	CategoryName    string     `json:"category_name"`
	CategoryType    string     `json:"category_type"`
	CreatedByName   string     `json:"created_by_name"`
	PostedAt        *time.Time `json:"posted_at"`
	VoidedAt        *time.Time `json:"voided_at"`
	VoidReason      *string    `json:"void_reason"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type accountRow struct {
	ID     uuid.UUID
	Status string
}

type categoryRow struct {
	ID   uuid.UUID
	Type string
	Ws   *uuid.UUID
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, in *CreateInput) (*View, error) {
	return s.create(ctx, nil, workspaceID, userID, in)
}

// CreateInTx runs the same creation inside an explicit transaction so foreign
// writers (e.g. recurring occurrence confirmation) stay atomic with their own
// state transition.
func (s *Service) CreateInTx(ctx context.Context, tx *gorm.DB, workspaceID, userID uuid.UUID, in *CreateInput) (*View, error) {
	return s.create(ctx, tx, workspaceID, userID, in)
}

func (s *Service) create(ctx context.Context, dbIn *gorm.DB, workspaceID, userID uuid.UUID, in *CreateInput) (*View, error) {
	q := dbIn
	if q == nil {
		q = s.db
	}
	if err := validateCreate(in); err != nil {
		return nil, err
	}
	acct, err := loadAccount(q, workspaceID, in.AccountID)
	if err != nil {
		return nil, err
	}
	if acct.Status != "ACTIVE" {
		return nil, errs.BusinessConflict("BUSINESS_CONFLICT", "This account is archived and cannot accept new activity.")
	}
	if in.CategoryID != nil {
		if err := validateCategorySave(q, workspaceID, *in.CategoryID, in.Type); err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	status := in.Status
	if status == "" {
		status = string(StatusDraft)
	}
	t := &Transaction{
		ID:              uuid.New(),
		WorkspaceID:     workspaceID,
		AccountID:       in.AccountID,
		CategoryID:      in.CategoryID,
		Type:            in.Type,
		Amount:          in.AmountMinor,
		TransactionDate: in.TransactionDate,
		Description:     nullableStr(in.Description),
		Merchant:        nullableStr(in.Merchant),
		Notes:           nullableStr(in.Notes),
		Source:          defaultSource(in.Source),
		Status:          status,
		Version:         1,
		CreatedByUserID: &userID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if status == string(StatusPosted) {
		t.PostedAt = &now
	}
	if err := q.WithContext(ctx).Create(t).Error; err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "transaction.create", "transaction", &t.ID, map[string]any{
		"type":   t.Type,
		"status": t.Status,
	})
	return toView(t), nil
}

func (s *Service) Update(ctx context.Context, workspaceID, userID uuid.UUID, in *UpdateInput) (*View, error) {
	if in.Version <= 0 {
		return nil, errs.ValidationFields(map[string]string{"version": "version is required"})
	}
	t, err := s.repo.FindByID(ctx, workspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	if t.Status != string(StatusDraft) {
		return nil, errs.BusinessConflict("BUSINESS_CONFLICT", "Only draft transactions can be edited. Void the posted transaction and create a replacement.")
	}
	if in.CategoryID != nil {
		if err := validateCategorySave(s.db, workspaceID, *in.CategoryID, in.Type); err != nil {
			return nil, err
		}
	}
	if in.AmountMinor <= 0 && in.Type != string(TypeAdjustment) {
		return nil, errs.ValidationFields(map[string]string{"amount": "Amount must be positive"})
	}
	t.Type = in.Type
	t.Amount = in.AmountMinor
	t.CategoryID = in.CategoryID
	t.TransactionDate = in.TransactionDate
	t.Description = nullableStr(in.Description)
	t.Merchant = nullableStr(in.Merchant)
	t.Notes = nullableStr(in.Notes)
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "transaction.update", "transaction", &t.ID, nil)
	updated, err := s.repo.FindByID(ctx, workspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	return toView(updated), nil
}

// Post transitions DRAFT → POSTED (takes financial effect). Safe to call on an
// already-posted row only when version matches the fetched version.
func (s *Service) Post(ctx context.Context, workspaceID, userID uuid.UUID, id uuid.UUID, version int64) (*View, error) {
	t, err := s.repo.FindByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	if t.Status == string(StatusPosted) {
		return toView(t), nil
	}
	if t.Status != string(StatusDraft) {
		return nil, errs.BusinessConflict("BUSINESS_CONFLICT", "Only draft transactions can be posted.")
	}
	acct, err := loadAccount(s.db, workspaceID, t.AccountID)
	if err != nil {
		return nil, err
	}
	if acct.Status != "ACTIVE" {
		return nil, errs.BusinessConflict("BUSINESS_CONFLICT", "This account is archived and cannot accept posted activity.")
	}
	if t.CategoryID != nil {
		if err := validateCategorySave(s.db, workspaceID, *t.CategoryID, t.Type); err != nil {
			return nil, err
		}
	}
	if err := s.repo.Post(ctx, workspaceID, id, version, time.Now().UTC()); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "transaction.post", "transaction", &id, nil)
	done, err := s.repo.FindByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	return toView(done), nil
}

func (s *Service) Void(ctx context.Context, workspaceID, userID uuid.UUID, in *VoidInput) (*View, error) {
	t, err := s.repo.FindByID(ctx, workspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	if t.Status == string(StatusVoided) {
		return nil, errs.BusinessConflict("BUSINESS_CONFLICT", "This transaction is already voided.")
	}
	if t.Status != string(StatusPosted) {
		return nil, errs.BusinessConflict("BUSINESS_CONFLICT", "Only posted transactions can be voided.")
	}
	if err := s.repo.Void(ctx, workspaceID, in.ID, in.Version, in.Reason, userID, time.Now().UTC()); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "transaction.void", "transaction", &t.ID, nil)
	done, err := s.repo.FindByID(ctx, workspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	return toView(done), nil
}

func (s *Service) List(ctx context.Context, workspaceID uuid.UUID, f ListFilter) ([]View, int64, error) {
	rows, total, err := s.repo.List(ctx, workspaceID, f)
	if err != nil {
		return nil, 0, err
	}
	out := make([]View, 0, len(rows))
	for i := range rows {
		out = append(out, *toView(&rows[i]))
	}
	return out, total, nil
}

func (s *Service) Get(ctx context.Context, workspaceID, id uuid.UUID) (*View, error) {
	t, err := s.repo.FindByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	return toView(t), nil
}

func loadAccount(q *gorm.DB, workspaceID, accountID uuid.UUID) (*accountRow, error) {
	var a accountRow
	err := q.WithContext(context.Background()).Table("accounts").
		Select("id, status").Where("id = ? AND workspace_id = ?", accountID, workspaceID).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.NotFound("Account not found")
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// validateCategorySave ensures the category is usable by this workspace and has the
// correct type for the transaction type.
func validateCategorySave(q *gorm.DB, workspaceID uuid.UUID, categoryID uuid.UUID, txType string) error {
	var c categoryRow
	err := q.WithContext(context.Background()).Table("categories").
		Select("id, type, workspace_id AS ws").
		Where("(id = ? AND (workspace_id = ? OR is_system = TRUE)) AND status = 'ACTIVE'", categoryID, workspaceID).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.ValidationFields(map[string]string{"category_id": "Category is not available in this workspace"})
	}
	if err != nil {
		return err
	}
	if txType != string(TypeAdjustment) && c.Type != txType {
		return errs.ValidationFields(map[string]string{"category_id": "Category type does not match the transaction type"})
	}
	return nil
}

func validateCreate(in *CreateInput) error {
	fields := map[string]string{}
	if !ValidType(in.Type) {
		fields["type"] = "Transaction type must be INCOME, EXPENSE or ADJUSTMENT"
	}
	if in.AccountID == uuid.Nil {
		fields["account_id"] = "Account is required"
	}
	if in.TransactionDate.IsZero() {
		fields["transaction_date"] = "Transaction date is required"
	}
	if in.Type == string(TypeAdjustment) {
		if in.AmountMinor == 0 {
			fields["amount"] = "Adjustment amount must not be zero"
		}
	} else if in.AmountMinor <= 0 {
		fields["amount"] = "Amount must be positive"
	}
	if in.Status != "" && in.Status != string(StatusDraft) && in.Status != string(StatusPosted) {
		fields["status"] = "status must be DRAFT or POSTED"
	}
	if len(fields) > 0 {
		return errs.ValidationFields(fields)
	}
	return nil
}

func defaultSource(s string) string {
	if s == "" {
		return string(SourceManual)
	}
	return strings.ToUpper(strings.TrimSpace(s))
}

func toView(t *Transaction) *View {
	return &View{
		ID:              t.ID,
		AccountID:       t.AccountID,
		CategoryID:      t.CategoryID,
		Type:            t.Type,
		Amount:          money.FormatMinorUnits(t.Amount),
		TransactionDate: t.TransactionDate.Format("2006-01-02"),
		Description:     t.Description,
		Merchant:        t.Merchant,
		Notes:           t.Notes,
		Source:          t.Source,
		Status:          t.Status,
		Version:         t.Version,
		AccountName:     t.AccountName,
		CategoryName:    t.CategoryName,
		CategoryType:    t.CategoryType,
		CreatedByName:   t.CreatedByName,
		PostedAt:        t.PostedAt,
		VoidedAt:        t.VoidedAt,
		VoidReason:      t.VoidReason,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}
