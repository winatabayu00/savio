package transfers

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

// Service owns transfer creation/voiding as a single atomic unit so a source
// debit is never observed without its destination credit.
type Service struct {
	db    *gorm.DB
	repo  *Repository
	audit *audit.Repository
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db, repo: NewRepository(db), audit: audit.NewRepository(db)}
}

type AccountLite struct {
	ID     uuid.UUID
	Status string
}

type CreateInput struct {
	FromAccountID uuid.UUID
	ToAccountID   uuid.UUID
	AmountMinor   int64
	TransferDate  time.Time
	Description   string
}

type VoidInput struct {
	ID      uuid.UUID
	Version int64
	Reason  string
}

type View struct {
	ID              uuid.UUID  `json:"id"`
	FromAccountID   uuid.UUID  `json:"from_account_id"`
	ToAccountID     uuid.UUID  `json:"to_account_id"`
	Amount          string     `json:"amount"`
	TransferDate    string     `json:"transfer_date"`
	Description     *string    `json:"description"`
	Status          string     `json:"status"`
	Version         int64      `json:"version"`
	FromAccountName string     `json:"from_account_name"`
	ToAccountName   string     `json:"to_account_name"`
	CreatedAt       time.Time  `json:"created_at"`
	VoidedAt        *time.Time `json:"voided_at"`
	VoidReason      *string    `json:"void_reason"`
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, in *CreateInput) (*View, error) {
	if err := validate(in); err != nil {
		return nil, err
	}
	if in.FromAccountID == in.ToAccountID {
		return nil, errs.ValidationFields(map[string]string{"to_account_id": "Source and destination must be different accounts"})
	}
	var created *Transfer
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		src, err := s.loadAccount(ctx, tx, workspaceID, in.FromAccountID)
		if err != nil {
			return err
		}
		dst, err := s.loadAccount(ctx, tx, workspaceID, in.ToAccountID)
		if err != nil {
			return err
		}
		if src.Status != "ACTIVE" || dst.Status != "ACTIVE" {
			return errs.BusinessConflict("BUSINESS_CONFLICT", "Both accounts must be active to transfer.")
		}
		now := time.Now().UTC()
		t := &Transfer{
			ID:              uuid.New(),
			WorkspaceID:     workspaceID,
			FromAccountID:   in.FromAccountID,
			ToAccountID:     in.ToAccountID,
			Amount:          in.AmountMinor,
			TransferDate:    in.TransferDate,
			Status:          string(StatusPosted),
			Version:         1,
			CreatedByUserID: &userID,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if strings.TrimSpace(in.Description) != "" {
			d := strings.TrimSpace(in.Description)
			t.Description = &d
		}
		created = t
		return s.repo.Create(ctx, tx, t)
	})
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "transfer.create", "transfer", &created.ID, nil)
	return toView(created), nil
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
	t, err := s.repo.FindByIDTx(ctx, s.db, workspaceID, id)
	if err != nil {
		return nil, err
	}
	return toView(t), nil
}

func (s *Service) Void(ctx context.Context, workspaceID, userID uuid.UUID, in *VoidInput) (*View, error) {
	t, err := s.repo.FindByIDTx(ctx, s.db, workspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	if t.Status == string(StatusVoided) {
		return nil, errs.BusinessConflict("BUSINESS_CONFLICT", "This transfer is already voided.")
	}
	now := time.Now().UTC()
	t.Version = in.Version
	t.VoidedByUserID = &userID
	t.VoidedAt = &now
	if strings.TrimSpace(in.Reason) != "" {
		r := strings.TrimSpace(in.Reason)
		t.VoidReason = &r
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.repo.Void(ctx, tx, t)
	}); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, &workspaceID, &userID, "transfer.void", "transfer", &t.ID, nil)
	done, err := s.repo.FindByIDTx(ctx, s.db, workspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	return toView(done), nil
}

func (s *Service) loadAccount(ctx context.Context, tx *gorm.DB, workspaceID, accountID uuid.UUID) (AccountLite, error) {
	var a AccountLite
	err := tx.WithContext(ctx).Table("accounts").
		Select("id, status").Where("id = ? AND workspace_id = ?", accountID, workspaceID).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return a, errs.NotFound("Account not found")
	}
	if err != nil {
		return a, err
	}
	return a, nil
}

func validate(in *CreateInput) error {
	fields := map[string]string{}
	if in.FromAccountID == uuid.Nil || in.ToAccountID == uuid.Nil {
		fields["accounts"] = "Both source and destination accounts are required"
	}
	if in.AmountMinor <= 0 {
		fields["amount"] = "Transfer amount must be positive"
	}
	if in.TransferDate.IsZero() {
		fields["transfer_date"] = "Transfer date is required"
	}
	if len(fields) > 0 {
		return errs.ValidationFields(fields)
	}
	return nil
}

func toView(t *Transfer) *View {
	return &View{
		ID:              t.ID,
		FromAccountID:   t.FromAccountID,
		ToAccountID:     t.ToAccountID,
		Amount:          money.FormatMinorUnits(t.Amount),
		TransferDate:    t.TransferDate.Format("2006-01-02"),
		Description:     t.Description,
		Status:          t.Status,
		Version:         t.Version,
		FromAccountName: t.FromAccountName,
		ToAccountName:   t.ToAccountName,
		CreatedAt:       t.CreatedAt,
		VoidedAt:        t.VoidedAt,
		VoidReason:      t.VoidReason,
	}
}
