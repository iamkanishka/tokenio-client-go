//go:build ignore

// payment_flow.go demonstrates the complete Token.io Payments v2 lifecycle:
// initiate → redirect auth → poll to final state → QR code generation.
//
// Run: go run examples/payment_flow.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tokenio "github.com/iamkanishka/tokenio-client-go"
	"github.com/iamkanishka/tokenio-client-go/pkg/payments"
)

func main() {
	// ── 1. Create the client ──────────────────────────────────────────────────
	client, err := tokenio.NewClient(tokenio.Config{
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
		// Defaults to EnvironmentSandbox → https://api.sandbox.token.io
	})
	if err != nil {
		log.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	fmt.Printf("SDK version: %s\n", client.Version())
	ctx := context.Background()

	// ── 2. Initiate a payment ─────────────────────────────────────────────────
	callbackURL := "https://example.com/payment-return"
	payment, err := client.Payments.InitiatePayment(ctx, payments.InitiatePaymentRequest{
		Initiation: payments.PaymentInitiation{
			BankID: "ob-modelo", // sandbox test bank
			RefID:  fmt.Sprintf("ref-%d", time.Now().UnixMilli()),
			Amount: &payments.Amount{Value: "10.50", Currency: "GBP"},
			Creditor: &payments.PartyAccount{
				AccountNumber: "12345678",
				SortCode:      "040004",
				Name:          "Acme Ltd",
			},
			RemittanceInformationPrimary: "Invoice INV-2024-001",
			CallbackURL:                  callbackURL,
			CallbackState:                "csrf-token-abc",
			FlowType:                     payments.FlowTypeFullHostedPages,
			ReturnRefundAccount:          true,
		},
		PispConsentAccepted: true,
	})
	if err != nil {
		log.Fatalf("InitiatePayment: %v", err)
	}
	fmt.Printf("Payment created: id=%s status=%s\n", payment.ID, payment.Status)

	// ── 3. Handle the auth flow ───────────────────────────────────────────────
	switch {
	case payment.Status.RequiresRedirect():
		fmt.Printf("→ Redirect PSU to: %s\n", payment.GetRedirectURL())
		fmt.Println("  (In production: HTTP 302 the PSU's browser to this URL)")

	case payment.Status.RequiresEmbeddedAuth():
		fmt.Println("→ Embedded auth required. Collect fields and call ProvideEmbeddedAuth.")
		if payment.Authentication != nil {
			for _, f := range payment.Authentication.EmbeddedAuth {
				fmt.Printf("  Field: id=%s type=%s mandatory=%v\n",
					f.ID, f.Type, f.Mandatory)
			}
		}

	case payment.Status.IsDecoupledAuth():
		fmt.Println("→ Decoupled auth in progress (bank is contacting PSU). Poll for status.")
	}

	// ── 4. Poll until final (demo only — prefer webhooks in production) ───────
	fmt.Println("\nPolling for final status...")
	final, err := client.Payments.PollUntilFinal(ctx, payment.ID, payments.PollOptions{
		Interval: 2 * time.Second,
	})
	if err != nil {
		log.Fatalf("poll: %v", err)
	}
	fmt.Printf("Final status: %s\n", final.Status)
	if final.RefundDetails != nil {
		fmt.Printf("Refund status: %s\n", final.RefundDetails.PaymentRefundStatus)
	}

	// ── 5. QR code (alternative auth UX) ─────────────────────────────────────
	if payment.GetRedirectURL() != "" {
		svgBytes, err := client.Payments.GenerateQRCode(ctx, payments.GenerateQRCodeRequest{
			Data: payment.GetRedirectURL(),
		})
		if err != nil {
			log.Printf("GenerateQRCode (non-fatal): %v", err)
		} else {
			fmt.Printf("QR code SVG: %d bytes\n", len(svgBytes))
		}
	}
}
