package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AccessTokenClaims are embedded in the short-lived JWT.
type AccessTokenClaims struct {
	SessionID uuid.UUID `json:"sid"`
	jwt.RegisteredClaims
}

// IssueAccessToken signs a short-lived JWT for the user+session.
func IssueAccessToken(secret string, ttl time.Duration, userID, sessionID uuid.UUID) (string, error) {
	now := time.Now()
	claims := AccessTokenClaims{
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// ParseAccessToken validates the signed JWT and returns claims.
func ParseAccessToken(secret, tokenString string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid access token")
	}
	if claims.Subject == "" {
		return nil, errors.New("missing subject")
	}
	return claims, nil
}

// NewRefreshToken returns a raw random token (sent via HttpOnly cookie) and
// its SHA-256 hash (the only value persisted).
func NewRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b)
	return raw, HashToken(raw), nil
}

// HashToken produces the storage hash for a raw token.
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// Session records an auth session row.
type Session struct {
	ID               uuid.UUID
	RefreshTokenHash string
	UserID           uuid.UUID
	ExpiresAt        time.Time
	LastUsedAt       *time.Time
	RevokedAt        *time.Time
	UserAgent        *string
	IPAddress        *string
}

func (s *Session) IsExpired(now time.Time) bool { return now.After(s.ExpiresAt) }
func (s *Session) IsRevoked() bool              { return s.RevokedAt != nil }
