package unit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	tokenio "github.com/iamkanishka/tokenio-client-go"
	sdkerrors "github.com/iamkanishka/tokenio-client-go/internal/errors"
	"github.com/iamkanishka/tokenio-client-go/pkg/payments"
)

func newTestClient(t *testing.T, srv *httptest.Server) *tokenio.Client {
	t.Helper()
	c, err := tokenio.NewClient(
		tokenio.Config{StaticToken: "test-token"},
		tokenio.WithBaseURL(srv.URL),
		tokenio.WithMaxRetries(0),
		tokenio.WithTimeout(5*time.Second),
		tokenio.WithRateLimit(10000, 10000),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func respondJSON(t *testing.T, w http.ResponseWriter, status int, body any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		t.Errorf("encode response: %v", err)
	}
}

func TestInitiatePayment_Success(t *testing.T) {
	want := payments.Payment{
		ID:     "pm:abc123:def",
		Status: payments.StatusInitiationPendingRedirectAuth,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/payments" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing Bearer token")
		}
		respondJSON(t, w, http.StatusOK, map[string]any{"payment": want})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, err := client.Payments.InitiatePayment(context.Background(), payments.InitiatePaymentRequest{
		Initiation: payments.PaymentInitiation{
			BankID: "ob-modelo",
			Amount: &payments.Amount{Value: "10.00", Currency: "GBP"},
		},
	})
	if err != nil {
		t.Fatalf("InitiatePayment: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID: got %q want %q", got.ID, want.ID)
	}
	if got.Status != want.Status {
		t.Errorf("Status: got %q want %q", got.Status, want.Status)
	}
}

func TestInitiatePayment_MissingBankID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be reached")
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Payments.InitiatePayment(context.Background(), payments.InitiatePaymentRequest{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestInitiatePayment_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(t, w, http.StatusBadRequest, map[string]any{
			"code":    "VALIDATION_ERROR",
			"message": "invalid bank id",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Payments.InitiatePayment(context.Background(), payments.InitiatePaymentRequest{
		Initiation: payments.PaymentInitiation{BankID: "bad-bank"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ae *sdkerrors.APIError
	if !asAPIError(err, &ae) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
	if ae.Status != http.StatusBadRequest {
		t.Errorf("Status: got %d, want 400", ae.Status)
	}
}

func TestGetPayment_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(t, w, http.StatusNotFound, map[string]any{
			"code":    "NOT_FOUND",
			"message": "payment not found",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Payments.GetPayment(context.Background(), "pm:missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !sdkerrors.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got false; err=%v", err)
	}
}

func TestGetPayment_EmptyID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be reached")
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Payments.GetPayment(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty paymentID")
	}
}

func TestStatus_IsFinal(t *testing.T) {
	for _, s := range []payments.Status{
		payments.StatusInitiationCompleted,
		payments.StatusInitiationRejected,
		payments.StatusInitiationRejectedInsufficientFunds,
		payments.StatusInitiationFailed,
		payments.StatusInitiationDeclined,
		payments.StatusInitiationExpired,
		payments.StatusInitiationNoFinalStatusAvailable,
		payments.StatusSettlementCompleted,
		payments.StatusSettlementIncomplete,
		payments.StatusCanceled,
	} {
		if !s.IsFinal() {
			t.Errorf("expected %q to be final", s)
		}
	}
	for _, s := range []payments.Status{
		payments.StatusInitiationPending,
		payments.StatusInitiationPendingRedirectAuth,
		payments.StatusInitiationPendingEmbeddedAuth,
		payments.StatusInitiationPendingDecoupledAuth,
		payments.StatusInitiationProcessing,
		payments.StatusSettlementInProgress,
	} {
		if s.IsFinal() {
			t.Errorf("expected %q NOT to be final", s)
		}
	}
}

func TestStatus_RequiresRedirect(t *testing.T) {
	for _, s := range []payments.Status{
		payments.StatusInitiationPendingRedirectAuth,
		payments.StatusInitiationPendingRedirectAuthVerif,
		payments.StatusInitiationPendingRedirectHP,
		payments.StatusInitiationPendingRedirectPBL,
	} {
		if !s.RequiresRedirect() {
			t.Errorf("expected %q to RequiresRedirect", s)
		}
	}
}

func TestPayment_GetRedirectURL(t *testing.T) {
	p := payments.Payment{
		Authentication: &payments.Authentication{RedirectURL: "https://bank.example.com/auth"},
	}
	if got := p.GetRedirectURL(); got != "https://bank.example.com/auth" {
		t.Errorf("GetRedirectURL: got %q", got)
	}
	p2 := payments.Payment{}
	if got := p2.GetRedirectURL(); got != "" {
		t.Errorf("GetRedirectURL on nil auth: got %q", got)
	}
}

func TestGetPayments_QueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("limit") != "20" {
			t.Errorf("limit: got %q, want 20", q.Get("limit"))
		}
		if q.Get("offset") != "tok_abc" {
			t.Errorf("offset: got %q, want tok_abc", q.Get("offset"))
		}
		respondJSON(t, w, http.StatusOK, map[string]any{"payments": []any{}})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Payments.GetPayments(context.Background(), payments.GetPaymentsRequest{
		Limit:  20,
		Offset: "tok_abc",
	})
	if err != nil {
		t.Fatalf("GetPayments: %v", err)
	}
}

func TestGetPayments_InvalidLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be reached")
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Payments.GetPayments(context.Background(), payments.GetPaymentsRequest{Limit: 0})
	if err == nil {
		t.Fatal("expected error for limit=0")
	}
}

func TestProvideEmbeddedAuth_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/payments/pm:123/embedded-auth" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		respondJSON(t, w, http.StatusOK, map[string]any{
			"payment": map[string]any{
				"id":     "pm:123",
				"status": string(payments.StatusInitiationProcessing),
			},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	p, err := client.Payments.ProvideEmbeddedAuth(context.Background(), payments.ProvideEmbeddedAuthRequest{
		PaymentID:    "pm:123",
		EmbeddedAuth: map[string]string{"otp": "123456"},
	})
	if err != nil {
		t.Fatalf("ProvideEmbeddedAuth: %v", err)
	}
	if p.ID != "pm:123" {
		t.Errorf("ID: got %q, want pm:123", p.ID)
	}
}

func TestProvideEmbeddedAuth_EmptyPaymentID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be reached")
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Payments.ProvideEmbeddedAuth(context.Background(), payments.ProvideEmbeddedAuthRequest{
		EmbeddedAuth: map[string]string{"otp": "123456"},
	})
	if err == nil {
		t.Fatal("expected error for empty PaymentID")
	}
}

func asAPIError(err error, out **sdkerrors.APIError) bool {
	if ae, ok := err.(*sdkerrors.APIError); ok {
		if out != nil {
			*out = ae
		}
		return true
	}
	return false
}
