package errs

import (
	"errors"
	"fmt"
	"net/http"
)

// Stable domain error codes.
const (
	CodeValidationError       = "VALIDATION_ERROR"
	CodeInvalidUUID           = "INVALID_UUID"
	CodeInvalidDateRange      = "INVALID_DATE_RANGE"
	CodeInvalidEnumValue      = "INVALID_ENUM_VALUE"
	CodeResourceNotFound      = "RESOURCE_NOT_FOUND"
	CodeVersionConflict       = "VERSION_CONFLICT"
	CodePermissionDenied      = "PERMISSION_DENIED"
	CodeResourceAccessDenied  = "RESOURCE_ACCESS_DENIED"
	CodeUnauthenticated       = "AUTHENTICATION_REQUIRED"
	CodeInvalidCredentials    = "INVALID_CREDENTIALS"
	CodeSessionExpired        = "SESSION_EXPIRED"
	CodeSessionRevoked        = "SESSION_REVOKED"
	CodeRefreshTokenInvalid   = "REFRESH_TOKEN_INVALID"
	CodeCSRFTokenInvalid      = "CSRF_TOKEN_INVALID"
	CodeRateLimited           = "RATE_LIMITED"
	CodeDuplicate             = "DUPLICATE_RESOURCE"
	CodeBusinessConflict      = "BUSINESS_CONFLICT"
	CodeAIProviderUnavailable = "AI_PROVIDER_UNAVAILABLE"
	CodeAIValidationFailed    = "AI_VALIDATION_FAILED"
	CodeInternal              = "INTERNAL_ERROR"
)

// Error is the centralized application error returned through a stable
// envelope. Causes are never exposed to clients.
type Error struct {
	Code    string              `json:"code"`
	Message string              `json:"message"`
	Status  int                 `json:"-"`
	Details map[string][]string `json:"details,omitempty"`
	Cause   error               `json:"-"`
	Extras  map[string]any      `json:"-"`
}

func (e *Error) Error() string { return e.Message }

func (e *Error) Unwrap() error { return e.Cause }

func newErr(code, msg string, status int) *Error {
	return &Error{Code: code, Message: msg, Status: status}
}

func Validation(message string) *Error {
	return newErr(CodeValidationError, message, http.StatusBadRequest)
}

func ValidationFields(fields map[string]string) *Error {
	e := newErr(CodeValidationError, "Please correct the highlighted fields.", http.StatusUnprocessableEntity)
	e.Details = map[string][]string{}
	for f, m := range fields {
		e.Details[f] = []string{m}
	}
	return e
}

func InvalidUUID() *Error {
	return newErr(CodeInvalidUUID, "Invalid UUID format", http.StatusBadRequest)
}

func NotFound(message string) *Error {
	return newErr(CodeResourceNotFound, message, http.StatusNotFound)
}

func VersionConflict(message string) *Error {
	return newErr(CodeVersionConflict, message, http.StatusConflict)
}

func PermissionDenied(message string) *Error {
	return newErr(CodePermissionDenied, message, http.StatusForbidden)
}

// ResourceAccessDenied is used for cross-workspace resource access attempts.
func ResourceAccessDenied() *Error {
	return newErr(CodeResourceAccessDenied, "You do not have access to this resource", http.StatusNotFound)
}

func Unauthenticated(message string) *Error {
	return newErr(CodeUnauthenticated, message, http.StatusUnauthorized)
}

func InvalidCredentials() *Error {
	return newErr(CodeInvalidCredentials, "Email or password is incorrect", http.StatusUnauthorized)
}

func SessionExpired() *Error {
	return newErr(CodeSessionExpired, "Session has expired", http.StatusUnauthorized)
}

func SessionRevoked() *Error {
	return newErr(CodeSessionRevoked, "Session was revoked", http.StatusUnauthorized)
}

func RefreshTokenInvalid() *Error {
	return newErr(CodeRefreshTokenInvalid, "Refresh token is invalid", http.StatusUnauthorized)
}

func CSRFTokenInvalid(message string) *Error {
	return newErr(CodeCSRFTokenInvalid, message, http.StatusForbidden)
}

func RateLimited(message string) *Error {
	return newErr(CodeRateLimited, message, http.StatusTooManyRequests)
}

func Duplicate(message string) *Error {
	return newErr(CodeDuplicate, message, http.StatusConflict)
}

func BusinessConflict(code, message string) *Error {
	return newErr(code, message, http.StatusConflict)
}

func AIUnavailable(message string) *Error {
	return newErr(CodeAIProviderUnavailable, message, http.StatusServiceUnavailable)
}

func AIValidation(message string) *Error {
	return newErr(CodeAIValidationFailed, message, http.StatusUnprocessableEntity)
}

// Internal builds an internal error, wrapping the cause for logs only.
func Internal(cause error) *Error {
	return &Error{Code: CodeInternal, Message: "An unexpected error occurred", Status: http.StatusInternalServerError, Cause: cause}
}

func From(err error) *Error {
	if err == nil {
		return nil
	}
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal(err)
}

func WrapInternal(cause error, format string, args ...any) *Error {
	return &Error{
		Code:    CodeInternal,
		Message: "An unexpected error occurred",
		Status:  http.StatusInternalServerError,
		Cause:   fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), cause),
	}
}
