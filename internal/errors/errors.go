// Package errors defines typed errors for the Token.io SDK.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Code is a machine-readable error code.
type Code string

const (
	CodeUnknown            Code = "UNKNOWN"
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeForbidden          Code = "FORBIDDEN"
	CodeNotFound           Code = "NOT_FOUND"
	CodeConflict           Code = "CONFLICT"
	CodeValidation         Code = "VALIDATION_ERROR"
	CodeRateLimit          Code = "RATE_LIMIT_EXCEEDED"
	CodeInternalServer     Code = "INTERNAL_SERVER_ERROR"
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
	CodeTimeout            Code = "TIMEOUT"
	CodeBadGateway         Code = "BAD_GATEWAY"
	CodeDeadlineExceeded   Code = "DEADLINE_EXCEEDED"
)

// APIError is a structured error returned by the Token.io API.
type APIError struct {
	Code      Code           `json:"code"`
	Message   string         `json:"message"`
	Status    int            `json:"status"`
	RequestID string         `json:"requestId,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("tokenio: [%s] %s (HTTP %d, traceId=%s)",
			e.Code, e.Message, e.Status, e.RequestID)
	}
	return fmt.Sprintf("tokenio: [%s] %s (HTTP %d)", e.Code, e.Message, e.Status)
}

// Is enables errors.Is matching by status and code.
func (e *APIError) Is(target error) bool {
	var t *APIError
	if errors.As(target, &t) {
		return (t.Status == 0 || t.Status == e.Status) &&
			(t.Code == "" || t.Code == e.Code)
	}
	return false
}

// IsNotFound returns true if err is a 404 API error.
func IsNotFound(err error) bool {
	var e *APIError
	return errors.As(err, &e) && e.Status == http.StatusNotFound
}

// IsUnauthorized returns true if err is a 401 API error.
func IsUnauthorized(err error) bool {
	var e *APIError
	return errors.As(err, &e) && e.Status == http.StatusUnauthorized
}

// IsRateLimit returns true if err is a 429 API error.
func IsRateLimit(err error) bool {
	var e *APIError
	return errors.As(err, &e) && e.Status == http.StatusTooManyRequests
}

// IsServerError returns true for any 5xx API error.
func IsServerError(err error) bool {
	var e *APIError
	return errors.As(err, &e) && e.Status >= 500
}

// IsRetryable returns true if the error is safe to retry automatically.
func IsRetryable(err error) bool {
	var e *APIError
	if !errors.As(err, &e) {
		return false
	}
	switch e.Status {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}
