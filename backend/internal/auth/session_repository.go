package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/savio/savio/backend/internal/platform/errs"
)

// sessionModel is the GORM mapping for auth_sessions.
type sessionModel struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID           uuid.UUID `gorm:"type:uuid;index"`
	RefreshTokenHash string    `gorm:"size:255"`
	UserAgent        *string
	IPAddress        *string
	DeviceName       *string
	ExpiresAt        time.Time
	LastUsedAt       *time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
}

func (sessionModel) TableName() string { return "auth_sessions" }

// SessionRepository owns auth_sessions persistence including row-locked rotation.
type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) toEntity(m *sessionModel) *Session {
	return &Session{
		ID:               m.ID,
		UserID:           m.UserID,
		RefreshTokenHash: m.RefreshTokenHash,
		ExpiresAt:        m.ExpiresAt,
		LastUsedAt:       m.LastUsedAt,
		RevokedAt:        m.RevokedAt,
		UserAgent:        m.UserAgent,
		IPAddress:        m.IPAddress,
	}
}

func (r *SessionRepository) Create(ctx context.Context, s *Session) error {
	m := &sessionModel{
		ID:               s.ID,
		UserID:           s.UserID,
		RefreshTokenHash: s.RefreshTokenHash,
		UserAgent:        s.UserAgent,
		IPAddress:        s.IPAddress,
		ExpiresAt:        s.ExpiresAt,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

// FindByHash locates a session by its refresh token hash.
func (r *SessionRepository) FindByHash(ctx context.Context, hash string) (*Session, error) {
	var m sessionModel
	err := r.db.WithContext(ctx).Where("refresh_token_hash = ?", hash).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errs.RefreshTokenInvalid()
	}
	if err != nil {
		return nil, err
	}
	return r.toEntity(&m), nil
}

// FindByID locates a session for logout.
func (r *SessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*Session, error) {
	var m sessionModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	return r.toEntity(&m), nil
}

// Rotate performs an atomic, row-locked refresh rotation: the session row is
// locked for update, verified still valid, then moved to the new token hash.
// Returns the locked session so the caller can enforce expiry/revocation rules.
func (r *SessionRepository) Rotate(ctx context.Context, id uuid.UUID, newHash string, now time.Time) (*Session, error) {
	var m sessionModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", id).First(&m).Error; err != nil {
			return err
		}
		if m.RevokedAt != nil {
			return errs.SessionRevoked()
		}
		if now.After(m.ExpiresAt) {
			return errs.SessionExpired()
		}
		lastUsed := now
		updates := map[string]any{
			"refresh_token_hash": newHash,
			"last_used_at":       lastUsed,
			"revoked_at":         gorm.Expr("NULL"),
		}
		return tx.Model(&m).Updates(updates).Error
	})
	if err != nil {
		return nil, err
	}
	return r.toEntity(&m), nil
}

// Revoke invalidates a single session (logout).
func (r *SessionRepository) Revoke(ctx context.Context, id uuid.UUID, now time.Time) error {
	return r.db.WithContext(ctx).Model(&sessionModel{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", now).Error
}

// RevokeAllByUser invalidates every session for a user (logout-all).
func (r *SessionRepository) RevokeAllByUser(ctx context.Context, userID uuid.UUID, now time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Model(&sessionModel{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now)
	return res.RowsAffected, res.Error
}

// ListByUser lists active sessions for the sessions endpoint.
func (r *SessionRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]sessionModel, error) {
	var rows []sessionModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&rows).Error
	return rows, err
}

// DeleteExpired removes expired sessions for housekeeping.
func (r *SessionRepository) DeleteExpired(ctx context.Context, now time.Time) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", now).Delete(&sessionModel{}).Error
}

// CountActive returns the number of active sessions for a user.
func (r *SessionRepository) CountActive(ctx context.Context, userID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&sessionModel{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).Count(&n).Error
	return n, err
}
