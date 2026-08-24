package categories

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/errs"
)

type Service struct {
	repo *Repository
}

func NewService(db *gorm.DB) *Service { return &Service{repo: NewRepository(db)} }

type View struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID *uuid.UUID `json:"workspace_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	IsSystem    bool      `json:"is_system"`
	Status      string    `json:"status"`
	Icon        *string   `json:"icon"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *Service) List(ctx context.Context, workspaceID uuid.UUID, typ string, includeArchived bool) ([]View, error) {
	rows, err := s.repo.List(ctx, workspaceID, typ, includeArchived)
	if err != nil {
		return nil, err
	}
	out := make([]View, 0, len(rows))
	for i := range rows {
		out = append(out, toView(&rows[i]))
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, workspaceID uuid.UUID, name, typ, icon, description string) (*View, error) {
	fields := map[string]string{}
	name = strings.TrimSpace(name)
	if name == "" {
		fields["name"] = "Category name is required"
	}
	if !ValidType(strings.ToUpper(typ)) {
		fields["type"] = "Category type must be INCOME or EXPENSE"
	}
	if len(fields) > 0 {
		return nil, errs.ValidationFields(fields)
	}
	ws := workspaceID
	var iconPtr, descPtr *string
	if icon != "" {
		iconPtr = &icon
	}
	if description != "" {
		descPtr = &description
	}
	c := &Category{
		ID:          uuid.New(),
		WorkspaceID: &ws,
		Name:        name,
		Type:        strings.ToUpper(typ),
		IsSystem:    false,
		Status:      string(StatusActive),
		Icon:        iconPtr,
		Description: descPtr,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.repo.Create(ctx, c); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil, errs.Duplicate("A category with this name and type already exists")
		}
		return nil, err
	}
	v := toView(c)
	return &v, nil
}

func (s *Service) Update(ctx context.Context, workspaceID uuid.UUID, id uuid.UUID, name, icon, description string) (*View, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errs.ValidationFields(map[string]string{"name": "Category name is required"})
	}
	custom, err := s.repo.FindCustomByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	if custom.IsSystem {
		return nil, errs.PermissionDenied("System categories cannot be modified")
	}
	custom.Name = name
	if icon != "" {
		custom.Icon = &icon
	} else {
		custom.Icon = nil
	}
	if description != "" {
		custom.Description = &description
	} else {
		custom.Description = nil
	}
	if err := s.repo.Update(ctx, custom); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil, errs.Duplicate("A category with this name and type already exists")
		}
		return nil, err
	}
	v := toView(custom)
	return &v, nil
}

func (s *Service) SetStatus(ctx context.Context, workspaceID uuid.UUID, id uuid.UUID, archived bool) (*View, error) {
	custom, err := s.repo.FindCustomByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	status := string(StatusActive)
	if archived {
		status = string(StatusArchived)
	}
	if err := s.repo.SetStatus(ctx, custom.ID, status); err != nil {
		return nil, err
	}
	custom.Status = status
	v := toView(custom)
	return &v, nil
}

func toView(c *Category) View {
	return View{
		ID:          c.ID,
		WorkspaceID: c.WorkspaceID,
		Name:        c.Name,
		Type:        c.Type,
		IsSystem:    c.IsSystem,
		Status:      c.Status,
		Icon:        c.Icon,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
	}
}