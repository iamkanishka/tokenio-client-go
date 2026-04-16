package unit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	tokenio "github.com/iamkanishka/tokenio-client-go"
	"github.com/iamkanishka/tokenio-client-go/pkg/banks"
)

func encodeResponse(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("encode response: %v", err)
	}
}

func TestRetry_RetriesOn503ThenSucceeds(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			encodeResponse(t, w, map[string]string{"code": "UNAVAILABLE"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		encodeResponse(t, w, map[string]any{"banks": []any{}, "pageInfo": nil})
	}))
	defer srv.Close()

	client, err := tokenio.NewClient(
		tokenio.Config{StaticToken: "tok"},
		tokenio.WithBaseURL(srv.URL),
		tokenio.WithMaxRetries(3),
		tokenio.WithRetryWait(10*time.Millisecond, 50*time.Millisecond),
		tokenio.WithRateLimit(10000, 10000),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.Banks.GetBanksV2(context.Background(), banks.GetBanksV2Request{Limit: 1})
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if n := atomic.LoadInt32(&calls); n < 3 {
		t.Errorf("expected at least 3 calls (2 retries), got %d", n)
	}
}

func TestRetry_DoesNotRetryOn400(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
		encodeResponse(t, w, map[string]string{"code": "INVALID_ARGUMENT", "message": "bad"})
	}))
	defer srv.Close()

	client, _ := tokenio.NewClient(
		tokenio.Config{StaticToken: "tok"},
		tokenio.WithBaseURL(srv.URL),
		tokenio.WithMaxRetries(3),
		tokenio.WithRateLimit(10000, 10000),
	)
	_, err := client.Payments.GetPayment(context.Background(), "pm:bad")
	if err == nil {
		t.Fatal("expected error")
	}
	if n := atomic.LoadInt32(&calls); n != 1 {
		t.Errorf("expected exactly 1 call (no retry on 400), got %d", n)
	}
}

func TestRetry_DoesNotRetryOn401(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusUnauthorized)
		encodeResponse(t, w, map[string]string{"code": "UNAUTHORIZED"})
	}))
	defer srv.Close()

	client, _ := tokenio.NewClient(
		tokenio.Config{StaticToken: "tok"},
		tokenio.WithBaseURL(srv.URL),
		tokenio.WithMaxRetries(3),
		tokenio.WithRateLimit(10000, 10000),
	)
	_, err := client.Payments.GetPayment(context.Background(), "pm:auth")
	if err == nil {
		t.Fatal("expected error")
	}
	if n := atomic.LoadInt32(&calls); n != 1 {
		t.Errorf("expected 1 call, got %d", n)
	}
}
