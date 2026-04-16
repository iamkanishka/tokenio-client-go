package unit_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	sdkerrors "github.com/iamkanishka/tokenio-client-go/internal/errors"
)

// writeBody writes a string body to a ResponseRecorder, failing the test on error.
func writeBody(t *testing.T, rec *httptest.ResponseRecorder, body string) {
	t.Helper()
	if _, err := fmt.Fprint(rec, body); err != nil {
		t.Fatalf("write body: %v", err)
	}
}

func TestFromResponse_ParsesCode(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("Content-Type", "application/json")
	rec.Header().Set("X-Request-ID", "trace-abc")
	rec.WriteHeader(http.StatusNotFound)
	writeBody(t, rec, `{"code":"NOT_FOUND","message":"payment not found"}`)

	err := sdkerrors.FromResponse(rec.Result())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	ae, ok := err.(*sdkerrors.APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if ae.Status != http.StatusNotFound {
		t.Errorf("Status: got %d, want 404", ae.Status)
	}
	if ae.Code != sdkerrors.CodeNotFound {
		t.Errorf("Code: got %q, want NOT_FOUND", ae.Code)
	}
	if ae.Message != "payment not found" {
		t.Errorf("Message: got %q", ae.Message)
	}
	if ae.RequestID != "trace-abc" {
		t.Errorf("RequestID: got %q", ae.RequestID)
	}
}

func TestFromResponse_DefaultCode429(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusTooManyRequests)
	writeBody(t, rec, `{}`)

	err := sdkerrors.FromResponse(rec.Result())
	ae, _ := err.(*sdkerrors.APIError)
	if ae.Code != sdkerrors.CodeRateLimit {
		t.Errorf("Code: got %q, want RATE_LIMIT_EXCEEDED", ae.Code)
	}
}

func TestIsNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusNotFound)
	writeBody(t, rec, `{"code":"NOT_FOUND","message":"missing"}`)
	err := sdkerrors.FromResponse(rec.Result())
	if !sdkerrors.IsNotFound(err) {
		t.Error("expected IsNotFound=true")
	}
}

func TestIsRateLimit(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusTooManyRequests)
	writeBody(t, rec, `{}`)
	err := sdkerrors.FromResponse(rec.Result())
	if !sdkerrors.IsRateLimit(err) {
		t.Error("expected IsRateLimit=true")
	}
}

func TestIsServerError(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusInternalServerError)
	writeBody(t, rec, `{}`)
	err := sdkerrors.FromResponse(rec.Result())
	if !sdkerrors.IsServerError(err) {
		t.Error("expected IsServerError=true")
	}
}

func TestIsRetryable(t *testing.T) {
	for _, code := range []int{429, 500, 502, 503, 504} {
		rec := httptest.NewRecorder()
		rec.WriteHeader(code)
		writeBody(t, rec, `{}`)
		if !sdkerrors.IsRetryable(sdkerrors.FromResponse(rec.Result())) {
			t.Errorf("expected HTTP %d to be retryable", code)
		}
	}
	for _, code := range []int{400, 401, 403, 404, 409, 422} {
		rec := httptest.NewRecorder()
		rec.WriteHeader(code)
		writeBody(t, rec, `{}`)
		if sdkerrors.IsRetryable(sdkerrors.FromResponse(rec.Result())) {
			t.Errorf("expected HTTP %d NOT to be retryable", code)
		}
	}
}

func TestAPIError_ErrorString(t *testing.T) {
	ae := &sdkerrors.APIError{
		Code:      "NOT_FOUND",
		Message:   "payment not found",
		Status:    404,
		RequestID: "req-abc",
	}
	if ae.Error() == "" {
		t.Error("Error() returned empty string")
	}
}
