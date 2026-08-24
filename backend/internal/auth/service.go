package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/savio/savio/backend/internal/platform/config"
	"github.com/savio/savio/backend/internal/platform/errs"
	"github.com/savio/savio/backend/internal/platform/password"
	"github.com/savio/savio/backend/internal/users"
	"github.com/savio/savio/backend/internal/workspaces"
)

// Service implements cookie-based authentication flows.
type Service struct {
	db        *gorm.DB
	cfg       *config.Config
	users     *users.Repository
	workspace *workspaces.Repository
	sessions  *SessionRepository
	now       func() time.Time
}

func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{
		db:        db,
		cfg:       cfg,
		users:     users.NewRepository(db),
		workspace: workspaces.NewRepository(db),
		sessions:  NewSessionRepository(db),
		now:       time.Now,
	}
}

// RegResult is returned after successful register/login/refresh.
type RegResult struct {
	UserID         uuid.UUID
	Name           string
	Email          string
	WorkspaceID    uuid.UUID
	WorkspaceName  string
	Role           string
	AccessToken    string
	RefreshToken   string
	AccessTokenTTL time.Duration
	SessionTTL     time.Duration
}

// Register atomically creates user + default workspace + OWNER membership +
// settings, then establishes a session.
func (s *Service) Register(ctx context.Context, name, email, plainPassword, userAgent, ip string) (*RegResult, error) {
	validation := map[string]string{}
	if len(name) < 2 {
		validation["name"] = "Name must be at least 2 characters"
	}
	if len(email) < 5 || !contains(email, "@") {
		validation["email"] = "A valid email is required"
	}
	if len(plainPassword) < 8 {
		validation["password"] = "Password must be at least 8 characters"
	}
	if len(validation) > 0 {
		return nil, errs.ValidationFields(validation)
	}

	hash, err := password.Hash(plainPassword)
	if err != nil {
		return nil, errs.WrapInternal(err, "hash password")
	}

	var res *RegResult
	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		userID := uuid.New()
		user := &users.User{
			ID:              userID,
			Name:            name,
			Email:           email,
			PasswordHash:    hash,
			Timezone:        "Asia/Jakarta",
			DefaultCurrency: "IDR",
			Locale:          "id-ID",
			Status:          "ACTIVE",
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		ws := &workspaces.Workspace{
			ID:           uuid.New(),
			Name:         name + "'s Workspace",
			BaseCurrency: "IDR",
			Timezone:     "Asia/Jakarta",
			CreatedBy:    &userID,
		}
		if err := tx.Create(ws).Error; err != nil {
			return err
		}

		membership := &workspaces.Membership{
			ID:          uuid.New(),
			WorkspaceID: ws.ID,
			UserID:      userID,
			Role:        workspaces.RoleOwner,
			Status:      "ACTIVE",
			CreatedBy:   &userID,
		}
		if err := tx.Create(membership).Error; err != nil {
			return err
		}

		settings := &users.UserSettings{
			UserID:                 userID,
			AIInsightsEnabled:      true,
			AICopilotEnabled:       true,
			NotificationsEnabled:   true,
			BudgetWarningThreshold: 80,
		}
		if err := tx.Create(settings).Error; err != nil {
			return err
		}

		r, err := s.createSession(ctx, tx, userID, userAgent, ip)
		if err != nil {
			return err
		}
		res = &RegResult{
			UserID:         userID,
			Name:           name,
			Email:          email,
			WorkspaceID:    ws.ID,
			WorkspaceName:  ws.Name,
			Role:           workspaces.RoleOwner,
			AccessToken:    r.AccessToken,
			RefreshToken:   r.RefreshToken,
			AccessTokenTTL: r.AccessTokenTTL,
			SessionTTL:     r.SessionTTL,
		}
		return nil
	})
	if txErr != nil {
		if isUniqueViolation(txErr) {
			return nil, errs.Duplicate("An account with this email already exists")
		}
		return nil, errs.WrapInternal(txErr, "register")
	}
	return res, nil
}

// Login verifies credentials and establishes a session. Responses are generic
// to prevent account enumeration.
func (s *Service) Login(ctx context.Context, email, plainPassword, userAgent, ip string) (*RegResult, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, errs.InvalidCredentials()
	}
	ok, err := password.Verify(plainPassword, user.PasswordHash)
	if err != nil || !ok {
		return nil, errs.InvalidCredentials()
	}
	if user.Status != "ACTIVE" {
		return nil, errs.InvalidCredentials()
	}

	session, err := s.createSession(ctx, s.db, user.ID, userAgent, ip)
	if err != nil {
		return nil, err
	}
	// Refresh last_login_at (non-critical, best effort).
	_ = s.db.WithContext(ctx).Model(&users.User{}).Where("id = ?", user.ID).
		Update("last_login_at", s.now()).Error

	membership, err := s.workspace.FindDefaultByUser(ctx, user.ID)
	if err != nil {
		// Account exists but has no workspace; treat as invalid to stay secure.
		return nil, errs.InvalidCredentials()
	}
	ws, err := s.workspace.FindWorkspaceByID(ctx, membership.WorkspaceID)
	if err != nil {
		return nil, err
	}
	return &RegResult{
		UserID:         user.ID,
		Name:           user.Name,
		Email:          user.Email,
		WorkspaceID:    ws.ID,
		WorkspaceName:  ws.Name,
		Role:           membership.Role,
		AccessToken:    session.AccessToken,
		RefreshToken:   session.RefreshToken,
		AccessTokenTTL: session.AccessTokenTTL,
		SessionTTL:     session.SessionTTL,
	}, nil
}

type sessionIssue struct {
	AccessToken    string
	RefreshToken   string
	AccessTokenTTL time.Duration
	SessionTTL     time.Duration
}

func (s *Service) createSession(ctx context.Context, db *gorm.DB, userID uuid.UUID, userAgent, ip string) (*sessionIssue, error) {
	expires := s.now().Add(s.cfg.RefreshTokenTTL)
	raw, hash, err := NewRefreshToken()
	if err != nil {
		return nil, errs.WrapInternal(err, "new refresh token")
	}
	sid := uuid.New()
	sess := &Session{
		ID:               sid,
		UserID:           userID,
		RefreshTokenHash: hash,
		ExpiresAt:        expires,
		UserAgent:        strPtr(userAgent),
	}
	if ip != "" {
		sess.IPAddress = &ip
	}
	if err := NewSessionRepository(db).Create(ctx, sess); err != nil {
		return nil, err
	}
	access, err := IssueAccessToken(s.cfg.JWTSecret, s.cfg.AccessTokenTTL, userID, sid)
	if err != nil {
		return nil, errs.WrapInternal(err, "issue access token")
	}
	return &sessionIssue{AccessToken: access, RefreshToken: raw, AccessTokenTTL: s.cfg.AccessTokenTTL, SessionTTL: s.cfg.RefreshTokenTTL}, nil
}

// Refresh rotates the refresh token (atomic, row-locked) and issues a new
// access token. Replay of an already-rotated token is rejected because the
// stored hash no longer matches it and the session has moved on.
func (s *Service) Refresh(ctx context.Context, rawRefresh string, userAgent string) (*RegResult, error) {
	if rawRefresh == "" {
		return nil, errs.RefreshTokenInvalid()
	}
	hash := HashToken(rawRefresh)
	sess, err := s.sessions.FindByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	newRaw, newHash, err := NewRefreshToken()
	if err != nil {
		return nil, errs.WrapInternal(err, "refresh token")
	}
	finalSess, err := s.sessions.Rotate(ctx, sess.ID, newHash, s.now())
	if err != nil {
		return nil, err
	}

	user, err := s.users.FindByID(ctx, finalSess.UserID)
	if err != nil {
		return nil, err
	}
	membership, err := s.workspace.FindDefaultByUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	ws, err := s.workspace.FindWorkspaceByID(ctx, membership.WorkspaceID)
	if err != nil {
		return nil, err
	}
	access, err := IssueAccessToken(s.cfg.JWTSecret, s.cfg.AccessTokenTTL, user.ID, finalSess.ID)
	if err != nil {
		return nil, errs.WrapInternal(err, "issue access token")
	}
	return &RegResult{
		UserID:         user.ID,
		Name:           user.Name,
		Email:          user.Email,
		WorkspaceID:    ws.ID,
		WorkspaceName:  ws.Name,
		Role:           membership.Role,
		AccessToken:    access,
		RefreshToken:   newRaw,
		AccessTokenTTL: s.cfg.AccessTokenTTL,
		SessionTTL:     s.cfg.RefreshTokenTTL,
	}, nil
}

// Logout revokes the current session.
func (s *Service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessions.Revoke(ctx, sessionID, s.now())
}

// LogoutAll revokes every session for the user.
func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	_, err := s.sessions.RevokeAllByUser(ctx, userID, s.now())
	return err
}

// Me returns the current user for bootstrapping.
func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*users.User, error) {
	return s.users.FindByID(ctx, userID)
}

func (s *Service) Settings(ctx context.Context, userID uuid.UUID) (*users.UserSettings, error) {
	return s.users.GetSettings(ctx, userID)
}

func (s *Service) CountActiveSessions(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.sessions.CountActive(ctx, userID)
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	// Postgres SQLSTATE 23505 (unique_violation) surfaced through GORM.
	return contains(msg, "23505") || contains(msg, "duplicate key")
}
