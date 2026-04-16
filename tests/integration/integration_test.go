//go:build integration

// Package integration contains tests that run against the Token.io sandbox.
//
// Run with:
//
//	TOKEN_CLIENT_ID=your-id TOKEN_CLIENT_SECRET=your-secret \
//	  go test -tags=integration ./tests/integration/... -v
package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	tokenio "github.com/iamkanishka/tokenio-client-go"
	"github.com/iamkanishka/tokenio-client-go/pkg/banks"
	"github.com/iamkanishka/tokenio-client-go/pkg/payments"
)

func newSandboxClient(t *testing.T) *tokenio.Client {
	t.Helper()
	clientID := os.Getenv("TOKEN_CLIENT_ID")
	secret := os.Getenv("TOKEN_CLIENT_SECRET")
	if clientID == "" || secret == "" {
		t.Skip("TOKEN_CLIENT_ID / TOKEN_CLIENT_SECRET not set")
	}
	c, err := tokenio.NewClient(tokenio.Config{
		ClientID:     clientID,
		ClientSecret: secret,
	}, tokenio.WithTimeout(30*time.Second))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func TestIntegration_Ping(t *testing.T) {
	c := newSandboxClient(t)
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := c.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestIntegration_GetBanksV2(t *testing.T) {
	c := newSandboxClient(t)
	defer c.Close()
	ctx := context.Background()

	resp, err := c.Banks.GetBanksV2(ctx, banks.GetBanksV2Request{Limit: 10})
	if err != nil {
		t.Fatalf("GetBanksV2: %v", err)
	}
	if len(resp.Banks) == 0 {
		t.Error("expected at least one bank")
	}
	t.Logf("Got %d banks", len(resp.Banks))
}

func TestIntegration_InitiateAndGetPayment(t *testing.T) {
	c := newSandboxClient(t)
	defer c.Close()
	ctx := context.Background()

	ref := "integration-test-" + time.Now().Format("20060102150405")
	callbackURL := "https://example.com/return"

	p, err := c.Payments.InitiatePayment(ctx, payments.InitiatePaymentRequest{
		Initiation: payments.PaymentInitiation{
			BankID: "ob-modelo",
			RefID:  ref,
			Amount: &payments.Amount{Value: "1.00", Currency: "GBP"},
			Creditor: &payments.PartyAccount{
				AccountNumber: "12345678",
				SortCode:      "040004",
				Name:          "Acme Ltd",
			},
			RemittanceInformationPrimary: "integration test",
			CallbackURL:                  callbackURL,
		},
	})
	if err != nil {
		t.Fatalf("InitiatePayment: %v", err)
	}
	t.Logf("Payment created: id=%s status=%s", p.ID, p.Status)

	// Retrieve the payment.
	p2, err := c.Payments.GetPayment(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetPayment: %v", err)
	}
	if p2.ID != p.ID {
		t.Errorf("ID mismatch: got %q, want %q", p2.ID, p.ID)
	}
}

func TestIntegration_ListPayments(t *testing.T) {
	c := newSandboxClient(t)
	defer c.Close()
	ctx := context.Background()

	resp, err := c.Payments.GetPayments(ctx, payments.GetPaymentsRequest{Limit: 5})
	if err != nil {
		t.Fatalf("GetPayments: %v", err)
	}
	t.Logf("Listed %d payments", len(resp.Payments))
}
