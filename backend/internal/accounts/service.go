package accounts

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/workspaces"
)

// Service owns account business rules: currency compatibility, validation,
// workspace scoping and derived balances.
type Service struct {
	db   *gorm.DB
	repo *Repository
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db, repo: NewRepository(db)}
}

type CreateInput struct {
	Name            string
	Type            string
	Currency        string
	OpeningBalance  int64
	InstitutionName string
	Description     string
}

type UpdateInput struct {
	ID              uuid.UUID
	Name            string
	Type            string
	InstitutionName string
	Description     string
	Version         int64
}

type View struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	Type           string     `json:"type"`
	Currency       string     `json:"currency"`
	OpeningBalance int64      `json:"opening_balance"`
	DerivedBalance int64      `json:"derived_balance"`
	Institution    *string    `json:"institution_name"`
	Description    *string    `json:"description"`
	Status         string     `json:"status"`
	Version        int64      `json:"version"`
	ArchivedAt     *time.Time `json:"archived_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (s *Service) workspaceBaseCurrency(ctx context.Context, workspaceID uuid.UUID) (string, error) {
	ws := &workspaces.Workspace{}
	if err := s.repo.db.WithContext(ctx).Where("id = ?", workspaceID).First(ws).Error; err != nil {
		return "", errs.NotFound("Workspace not found")
	}
	return ws.BaseCurrency, nil
}

func (s *Service) List(ctx context.Context, workspaceID uuid.UUID, status, typ, sortField string, sortDesc bool, page, limit, offset int) ([]View, int64, error) {
	rows, total, err := s.repo.List(ctx, workspaceID, status, typ, sortField, sortDesc, page, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	mods, err := s.repo.PostBalanceModifiers(ctx, workspaceID)
	if err != nil {
		return nil, 0, err
	}
	out := make([]View, 0, len(rows))
	for i := range rows {
		out = append(out, toView(&rows[i], mods[rows[i].ID]))
	}
	return out, total, nil
}

func (s *Service) Get(ctx context.Context, workspaceID, id uuid.UUID) (*View, error) {
	a, err := s.repo.FindByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	mods, err := s.repo.PostBalanceModifiers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	v := toView(a, mods[a.ID])
	return &v, nil
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, in *CreateInput) (*View, error) {
	if err := validateCreate(in); err != nil {
		return nil, err
	}
	base, err := s.workspaceBaseCurrency(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = base
	}
	if currency != base {
		return nil, errs.Validation("This account currency is not supported yet. Accounts use the workspace base currency.")
	}
	now := time.Now()
	a := &Account{
		ID:              uuid.New(),
		WorkspaceID:     workspaceID,
		Name:            strings.TrimSpace(in.Name),
		Type:            strings.ToUpper(in.Type),
		Currency:        currency,
		OpeningBalance:  in.OpeningBalance,
		Status:          string(StatusActive),
		Version:         1,
		CreatedByUserID: &userID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if in.InstitutionName != "" {
		a.InstitutionName = &in.InstitutionName
	}
	if in.Description != "" {
		a.Description = &in.Description
	}
	if err := s.repo.Create(ctx, a); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil, errs.Duplicate("An account with this name already exists in the workspace")
		}
		return nil, err
	}
	v := toView(a, 0)
	return &v, nil
}

func (s *Service) Update(ctx context.Context, workspaceID uuid.UUID, in *UpdateInput) (*View, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ValidationFields(map[string]string{"name": "Name is required"})
	}
	if !ValidType(strings.ToUpper(in.Type)) {
		return nil, errs.ValidationFields(map[string]string{"type": "Unsupported account type"})
	}
	a, err := s.repo.FindByID(ctx, workspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Version > 0 && in.Version != a.Version {
		return nil, errs.VersionConflict("This account was changed since you last opened it. Reload the latest version.")
	}
	a.Name = strings.TrimSpace(in.Name)
	a.Type = strings.ToUpper(in.Type)
	if in.InstitutionName != "" {
		a.InstitutionName = &in.InstitutionName
	} else {
		a.InstitutionName = nil
	}
	if in.Description != "" {
		a.Description = &in.Description
	} else {
		a.Description = nil
	}
	if err := s.repo.Update(ctx, a); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil, errs.Duplicate("An account with this name already exists in the workspace")
		}
		return nil, err
	}
	mods, err := s.repo.PostBalanceModifiers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	v := toView(a, mods[a.ID])
	return &v, nil
}

func (s *Service) SetStatus(ctx context.Context, workspaceID, id uuid.UUID, archived bool, expectVersion *int64) (*View, error) {
	status := string(StatusActive)
	var archivedAt *time.Time
	if archived {
		status = string(StatusArchived)
		t := time.Now()
		archivedAt = &t
	}
	if err := s.repo.SetStatus(ctx, workspaceID, id, status, archivedAt, expectVersion); err != nil {
		return nil, err
	}
	a, err := s.repo.FindByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	mods, err := s.repo.PostBalanceModifiers(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	v := toView(a, mods[a.ID])
	return &v, nil
}

func (s *Service) Delete(ctx context.Context, workspaceID, id uuid.UUID) error {
	return s.repo.Delete(ctx, workspaceID, id)
}

func validateCreate(in *CreateInput) error {
	fields := map[string]string{}
	if strings.TrimSpace(in.Name) == "" {
		fields["name"] = "Account name is required"
	}
	if !ValidType(strings.ToUpper(in.Type)) {
		fields["type"] = "Unsupported account type"
	}
	if in.OpeningBalance < 0 {
		fields["opening_balance"] = "Opening balance cannot be negative"
	}
	if len(fields) > 0 {
		return errs.ValidationFields(fields)
	}
	return nil
}

func toView(a *Account, modifier int64) View {
	return View{
		ID:             a.ID,
		Name:           a.Name,
		Type:           a.Type,
		Currency:       a.Currency,
		OpeningBalance: a.OpeningBalance,
		DerivedBalance: a.OpeningBalance + modifier,
		Institution:    a.InstitutionName,
		Description:    a.Description,
		Status:         a.Status,
		Version:        a.Version,
		ArchivedAt:     a.ArchivedAt,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}
