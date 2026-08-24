package csrf

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// Signed double-submit CSRF token. The cookie holds "value.signature" where
// signature = HMAC-SHA256(secret, value). The header must exactly match the
// cookie value, and the signature must verify — this prevents login CSRF and
// forgery since an attacker cannot reproduce a valid signature.
func Generate(secret string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	value := base64.RawURLEncoding.EncodeToString(b)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(value))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return value + "." + sig, nil
}

// Validate checks the header token against the signed cookie token.
func Validate(secret, cookieValue, headerToken string) error {
	if cookieValue == "" || headerToken == "" {
		return errors.New("missing CSRF token")
	}
	if !hmac.Equal([]byte(cookieValue), []byte(headerToken)) {
		return errors.New("CSRF token mismatch")
	}
	lastDot := strings.LastIndex(cookieValue, ".")
	if lastDot <= 0 || lastDot == len(cookieValue)-1 {
		return errors.New("malformed CSRF token")
	}
	value := cookieValue[:lastDot]
	sig := cookieValue[lastDot+1:]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(value))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return fmt.Errorf("CSRF signature verification failed")
	}
	return nil
}
