//go:build ignore

// vrp_flow.go demonstrates the Variable Recurring Payments lifecycle:
// create consent → PSU authorises → initiate payments → confirm funds.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	tokenio "github.com/iamkanishka/tokenio-client-go"
	"github.com/iamkanishka/tokenio-client-go/pkg/common"
	"github.com/iamkanishka/tokenio-client-go/pkg/vrp"
)

func main() {
	client, err := tokenio.NewClient(tokenio.Config{
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
	})
	if err != nil {
		log.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	endDate := time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339)

	// ── 1. Create a VRP consent ───────────────────────────────────────────────
	consent, err := client.VRP.CreateVrpConsent(ctx, vrp.CreateVrpConsentRequest{
		Initiation: vrp.VrpConsentInitiation{
			BankID:   "ob-modelo",
			RefID:    fmt.Sprintf("vrp-ref-%d", time.Now().UnixMilli()),
			Scheme:   vrp.VrpSchemeOBLSweeping,
			Currency: "GBP",
			Creditor: &common.PartyAccount{
				AccountNumber: "12345678",
				SortCode:      "040004",
				Name:          "Acme Subscription Service",
			},
			RemittanceInformationPrimary: "Monthly subscription",
			MaximumIndividualAmount:      "500.00",
			PeriodicLimits: []vrp.PeriodicLimit{
				{
					MaximumAmount:   "1000.00",
					PeriodType:      "MONTH",
					PeriodAlignment: "CALENDAR",
				},
			},
			MaximumOccurrences:  12,
			EndDateTime:         endDate,
			CallbackURL:         "https://example.com/vrp-return",
			ReturnRefundAccount: true,
		},
		PispConsentAccepted: true,
	})
	if err != nil {
		log.Fatalf("CreateVrpConsent: %v", err)
	}
	fmt.Printf("Consent created: id=%s status=%s\n", consent.ID, consent.Status)

	if consent.Status.RequiresRedirect() {
		fmt.Printf("→ Redirect PSU to: %s\n", consent.GetRedirectURL())
	}

	// ── 2. Poll for AUTHORIZED status (sandbox only; use webhooks in production)
	fmt.Println("Polling for consent authorization...")
	for !consent.Status.IsFinal() {
		time.Sleep(2 * time.Second)
		consent, err = client.VRP.GetVrpConsent(ctx, consent.ID)
		if err != nil {
			log.Fatalf("GetVrpConsent: %v", err)
		}
		fmt.Printf("  status=%s\n", consent.Status)
	}

	if consent.Status != vrp.ConsentStatusAuthorized {
		log.Fatalf("Consent ended with status %s (expected AUTHORIZED)", consent.Status)
	}
	fmt.Printf("Consent authorized: id=%s\n", consent.ID)

	// ── 3. Confirm funds before initiating a payment ──────────────────────────
	available, err := client.VRP.ConfirmFunds(ctx, consent.ID, "49.99")
	if err != nil {
		log.Printf("ConfirmFunds (non-fatal): %v", err)
	} else {
		fmt.Printf("Funds available (£49.99): %v\n", available)
	}

	// ── 4. Initiate a VRP payment under the consent ───────────────────────────
	payment, err := client.VRP.CreateVrp(ctx, vrp.CreateVrpRequest{
		Initiation: vrp.VrpInitiation{
			ConsentID:                    consent.ID,
			RefID:                        fmt.Sprintf("vrp-pay-%d", time.Now().UnixMilli()),
			RemittanceInformationPrimary: "Subscription Jan 2025",
			Amount:                       &common.Amount{Value: "49.99", Currency: "GBP"},
		},
	})
	if err != nil {
		log.Fatalf("CreateVrp: %v", err)
	}
	fmt.Printf("VRP payment initiated: id=%s status=%s\n", payment.ID, payment.Status)

	// ── 5. Poll to final ──────────────────────────────────────────────────────
	for !payment.IsFinal() {
		time.Sleep(2 * time.Second)
		payment, err = client.VRP.GetVrp(ctx, payment.ID)
		if err != nil {
			log.Fatalf("GetVrp: %v", err)
		}
		fmt.Printf("  status=%s\n", payment.Status)
	}
	fmt.Printf("VRP payment final: id=%s status=%s\n", payment.ID, payment.Status)

	// ── 6. List all payments under the consent ────────────────────────────────
	payments, err := client.VRP.GetVrpConsentPayments(ctx, vrp.GetVrpConsentPaymentsRequest{
		ConsentID: consent.ID,
		Limit:     10,
	})
	if err != nil {
		log.Fatalf("GetVrpConsentPayments: %v", err)
	}
	fmt.Printf("Payments under consent: %d\n", len(payments.Vrps))

	// ── 7. Revoke the consent ─────────────────────────────────────────────────
	if _, err := client.VRP.RevokeVrpConsent(ctx, consent.ID); err != nil {
		log.Printf("RevokeVrpConsent (non-fatal): %v", err)
	} else {
		fmt.Println("Consent revoked")
	}
}
