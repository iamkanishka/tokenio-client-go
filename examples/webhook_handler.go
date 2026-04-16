//go:build ignore

// webhook_handler.go demonstrates a production-ready HTTP webhook handler
// that verifies signatures and dispatches typed events.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	tokenio "github.com/iamkanishka/tokenio-client-go"
	"github.com/iamkanishka/tokenio-client-go/pkg/webhooks"
)

func main() {
	client, err := tokenio.NewClient(tokenio.Config{
		ClientID:      "your-client-id",
		ClientSecret:  "your-client-secret",
		WebhookSecret: "your-webhook-secret", // from Token.io dashboard
	})
	if err != nil {
		log.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// ── Register webhook endpoint ─────────────────────────────────────────────
	if err := client.Webhooks.SetConfig(ctx, webhooks.SetWebhookConfigRequest{
		Config: webhooks.WebhookConfig{
			URL: "https://your-tpp.com/webhooks/token",
			Events: []string{
				string(webhooks.EventTypePaymentUpdated),
				string(webhooks.EventTypeVRPConsentCreated),
				string(webhooks.EventTypeVRPCompleted),
				string(webhooks.EventTypeRefundCompleted),
			},
		},
	}); err != nil {
		log.Printf("SetWebhookConfig (non-fatal): %v", err)
	}

	// ── Start webhook HTTP server ─────────────────────────────────────────────
	mux := http.NewServeMux()
	mux.HandleFunc("/webhooks/token", makeWebhookHandler(client))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Println("Webhook server listening on :8080")
	log.Fatal(srv.ListenAndServe())
}

func makeWebhookHandler(client *tokenio.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MB limit
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		sig := r.Header.Get("X-Token-Signature")
		event, err := client.Webhooks.Parse(body, sig)
		if err != nil {
			log.Printf("webhook: invalid signature: %v", err)
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		log.Printf("webhook: received event id=%s type=%s", event.ID, event.Type)

		// Dispatch by event type.
		switch event.Type {
		case webhooks.EventTypePaymentUpdated, webhooks.EventTypePaymentCompleted,
			webhooks.EventTypePaymentFailed:
			data, err := webhooks.DecodePaymentEvent(event)
			if err != nil {
				log.Printf("webhook: decode payment event: %v", err)
				break
			}
			fmt.Printf("Payment %s → status=%s\n", data.PaymentID, data.Status)

		case webhooks.EventTypeVRPConsentCreated, webhooks.EventTypeVRPConsentUpdated,
			webhooks.EventTypeVRPConsentRevoked:
			data, err := webhooks.DecodeVRPConsentEvent(event)
			if err != nil {
				log.Printf("webhook: decode VRP consent event: %v", err)
				break
			}
			fmt.Printf("VRP Consent %s → status=%s\n", data.ConsentID, data.Status)

		case webhooks.EventTypeVRPCreated, webhooks.EventTypeVRPCompleted,
			webhooks.EventTypeVRPFailed:
			data, err := webhooks.DecodeVRPEvent(event)
			if err != nil {
				log.Printf("webhook: decode VRP event: %v", err)
				break
			}
			fmt.Printf("VRP %s (consent=%s) → status=%s\n", data.VrpID, data.ConsentID, data.Status)

		case webhooks.EventTypeRefundCompleted, webhooks.EventTypeRefundFailed:
			data, err := webhooks.DecodeRefundEvent(event)
			if err != nil {
				log.Printf("webhook: decode refund event: %v", err)
				break
			}
			fmt.Printf("Refund %s → status=%s\n", data.RefundID, data.Status)

		default:
			log.Printf("webhook: unhandled event type %q", event.Type)
		}

		w.WriteHeader(http.StatusOK)
	}
}
