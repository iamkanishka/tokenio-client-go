package unit_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/iamkanishka/tokenio-client-go/pkg/webhooks"
)

// testHMACSecret is a non-production test credential.
//
//nolint:gosec // test-only value, not a real secret
const testHMACSecret = "test-webhook-hmac-secret-for-unit-tests"

func makeTestSignature(secret string, payload []byte) string {
	ts := time.Now().Unix()
	signed := fmt.Sprintf("%d.%s", ts, payload)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signed))
	return fmt.Sprintf("t=%d,v1=%x", ts, mac.Sum(nil))
}

func TestParseEvent_ValidSignature(t *testing.T) {
	payload, _ := json.Marshal(map[string]any{
		"id":        "evt-001",
		"type":      "payment.updated",
		"createdAt": time.Now().Format(time.RFC3339),
		"data":      map[string]any{"paymentId": "pm:abc", "status": "INITIATION_COMPLETED"},
	})

	sig := makeTestSignature(testHMACSecret, payload)
	h := webhooks.NewClient(nil, testHMACSecret)
	event, err := h.Parse(payload, sig)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if event.ID != "evt-001" {
		t.Errorf("ID: got %q, want evt-001", event.ID)
	}
	if event.Type != webhooks.EventTypePaymentUpdated {
		t.Errorf("Type: got %q", event.Type)
	}
}

func TestParseEvent_InvalidSignature(t *testing.T) {
	h := webhooks.NewClient(nil, "correct-secret")
	payload := []byte(`{"id":"evt-1","type":"payment.updated","createdAt":"2024-01-01T00:00:00Z"}`)
	_, err := h.Parse(payload, "t=999999999,v1=badsignature")
	if err == nil {
		t.Fatal("expected signature error, got nil")
	}
}

func TestParseEvent_StaleTimestamp(t *testing.T) {
	payload := []byte(`{"id":"evt-stale"}`)
	oldTS := time.Now().Add(-10 * time.Minute).Unix()
	signed := fmt.Sprintf("%d.%s", oldTS, payload)
	mac := hmac.New(sha256.New, []byte(testHMACSecret))
	_, _ = mac.Write([]byte(signed))
	sig := fmt.Sprintf("t=%d,v1=%x", oldTS, mac.Sum(nil))

	h := webhooks.NewClient(nil, testHMACSecret)
	_, err := h.Parse(payload, sig)
	if err == nil {
		t.Fatal("expected stale timestamp error, got nil")
	}
}

func TestParseEvent_NoSecret_SkipsVerification(t *testing.T) {
	h := webhooks.NewClient(nil, "") // empty secret = skip verification
	payload, _ := json.Marshal(map[string]any{
		"id":        "evt-2",
		"type":      "vrp.completed",
		"createdAt": time.Now().Format(time.RFC3339),
	})
	event, err := h.Parse(payload, "t=0,v1=anysig")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if event.ID != "evt-2" {
		t.Errorf("ID: got %q, want evt-2", event.ID)
	}
}

func TestDecodePaymentEvent(t *testing.T) {
	data, _ := json.Marshal(webhooks.PaymentEventData{
		PaymentID: "pm:xyz",
		Status:    "INITIATION_COMPLETED",
	})
	event := &webhooks.Event{Type: webhooks.EventTypePaymentCompleted, Data: data}
	d, err := webhooks.DecodePaymentEvent(event)
	if err != nil {
		t.Fatalf("DecodePaymentEvent: %v", err)
	}
	if d.PaymentID != "pm:xyz" {
		t.Errorf("PaymentID: got %q, want pm:xyz", d.PaymentID)
	}
}

func TestDecodeVRPConsentEvent(t *testing.T) {
	data, _ := json.Marshal(webhooks.VRPConsentEventData{
		ConsentID: "vc:abc",
		Status:    "AUTHORIZED",
	})
	event := &webhooks.Event{Type: webhooks.EventTypeVRPConsentCreated, Data: data}
	d, err := webhooks.DecodeVRPConsentEvent(event)
	if err != nil {
		t.Fatalf("DecodeVRPConsentEvent: %v", err)
	}
	if d.ConsentID != "vc:abc" {
		t.Errorf("ConsentID: got %q, want vc:abc", d.ConsentID)
	}
}
