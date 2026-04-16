package unit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	tokenio "github.com/iamkanishka/tokenio-client-go"
	"github.com/iamkanishka/tokenio-client-go/pkg/common"
	"github.com/iamkanishka/tokenio-client-go/pkg/vrp"
)

func newVRPClient(t *testing.T, srv *httptest.Server) *tokenio.Client {
	t.Helper()
	c, err := tokenio.NewClient(
		tokenio.Config{StaticToken: "tok"},
		tokenio.WithBaseURL(srv.URL),
		tokenio.WithMaxRetries(0),
		tokenio.WithRateLimit(10000, 10000),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func jsonEncode(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Errorf("encode: %v", err)
	}
}

func TestCreateVrpConsent_Success(t *testing.T) {
	want := vrp.VrpConsent{ID: "vc:abc123:def", Status: vrp.ConsentStatusPending}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/vrp-consents" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		jsonEncode(t, w, map[string]any{"vrpConsent": want})
	}))
	defer srv.Close()

	client := newVRPClient(t, srv)
	got, err := client.VRP.CreateVrpConsent(context.Background(), vrp.CreateVrpConsentRequest{
		Initiation: vrp.VrpConsentInitiation{
			BankID:   "ob-modelo",
			Currency: "GBP",
			Creditor: &common.PartyAccount{
				AccountNumber: "12345678",
				SortCode:      "040004",
				Name:          "Acme Ltd",
			},
			PeriodicLimits: []vrp.PeriodicLimit{
				{MaximumAmount: "1000.00", PeriodType: "MONTH", PeriodAlignment: "CALENDAR"},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateVrpConsent: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID: got %q, want %q", got.ID, want.ID)
	}
}

func TestCreateVrpConsent_MissingBankID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be reached")
	}))
	defer srv.Close()

	client := newVRPClient(t, srv)
	_, err := client.VRP.CreateVrpConsent(context.Background(), vrp.CreateVrpConsentRequest{
		Initiation: vrp.VrpConsentInitiation{Creditor: &common.PartyAccount{Name: "Acme"}},
	})
	if err == nil {
		t.Fatal("expected validation error for missing BankID")
	}
}

func TestConsentStatus_IsFinal(t *testing.T) {
	for _, s := range []vrp.ConsentStatus{
		vrp.ConsentStatusAuthorized, vrp.ConsentStatusRejected,
		vrp.ConsentStatusRevoked, vrp.ConsentStatusFailed,
	} {
		if !s.IsFinal() {
			t.Errorf("expected %q to be final", s)
		}
	}
	for _, s := range []vrp.ConsentStatus{
		vrp.ConsentStatusPending, vrp.ConsentStatusPendingMoreInfo,
		vrp.ConsentStatusPendingRedirectAuth,
	} {
		if s.IsFinal() {
			t.Errorf("expected %q NOT to be final", s)
		}
	}
}

func TestVrpStatus_IsFinal(t *testing.T) {
	for _, s := range []vrp.VrpStatus{
		vrp.VrpStatusInitiationCompleted, vrp.VrpStatusInitiationRejected,
		vrp.VrpStatusInitiationRejectedInsufficient, vrp.VrpStatusInitiationFailed,
		vrp.VrpStatusNoFinalStatus,
	} {
		if !s.IsFinal() {
			t.Errorf("expected %q to be final", s)
		}
	}
}

func TestRevokeVrpConsent_MissingID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be reached")
	}))
	defer srv.Close()

	client := newVRPClient(t, srv)
	_, err := client.VRP.RevokeVrpConsent(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty consentID")
	}
}

func TestCreateVrp_RoutesCorrectly(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/vrps" {
			t.Errorf("path: got %q, want /vrps", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		jsonEncode(t, w, map[string]any{
			"vrp": vrp.Vrp{ID: "vrp:abc", Status: vrp.VrpStatusInitiationPending},
		})
	}))
	defer srv.Close()

	client := newVRPClient(t, srv)
	_, err := client.VRP.CreateVrp(context.Background(), vrp.CreateVrpRequest{
		Initiation: vrp.VrpInitiation{
			ConsentID: "vc:abc",
			Amount:    &common.Amount{Value: "49.99", Currency: "GBP"},
		},
	})
	if err != nil {
		t.Fatalf("CreateVrp: %v", err)
	}
	if !called {
		t.Error("server handler was never called")
	}
}

func TestConfirmFunds_ReturnsResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/vrps/vc:abc/confirm-funds" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("amount") != "100.00" {
			t.Errorf("amount: %s", r.URL.Query().Get("amount"))
		}
		w.Header().Set("Content-Type", "application/json")
		jsonEncode(t, w, map[string]any{"fundsAvailable": true})
	}))
	defer srv.Close()

	client := newVRPClient(t, srv)
	available, err := client.VRP.ConfirmFunds(context.Background(), "vc:abc", "100.00")
	if err != nil {
		t.Fatalf("ConfirmFunds: %v", err)
	}
	if !available {
		t.Error("expected fundsAvailable=true")
	}
}
