// Package webhooks provides webhook configuration management and typed event
// parsing for Token.io webhook deliveries.
package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
)

// ── Webhook Config Management ─────────────────────────────────────────────────

// WebhookConfig defines the TPP's webhook endpoint configuration.
type WebhookConfig struct {
	URL    string   `json:"url"`
	Events []string `json:"events,omitempty"` // event types to subscribe to
}

// SetWebhookConfigRequest is the body for PUT /webhook-config.
type SetWebhookConfigRequest struct {
	Config WebhookConfig `json:"config"`
}

// GetWebhookConfigResponse is returned by GET /webhook-config.
type GetWebhookConfigResponse struct {
	Config WebhookConfig `json:"config"`
}

// Client manages webhook configuration and parses incoming events.
type Client struct {
	hc     *httpclient.Client
	secret []byte
}

// NewClient creates a Webhooks client backed by hc.
// secret is the shared secret for verifying incoming webhook payloads.
func NewClient(hc *httpclient.Client, secret string) *Client {
	return &Client{hc: hc, secret: []byte(secret)}
}

// SetConfig registers or updates the TPP webhook endpoint.
//
// PUT /webhook-config
func (c *Client) SetConfig(ctx context.Context, req SetWebhookConfigRequest) error {
	if req.Config.URL == "" {
		return fmt.Errorf("webhooks: Config.URL is required")
	}
	return c.hc.Do(ctx, httpclient.NewRequest(http.MethodPut, "/webhook-config").WithBody(req), nil)
}

// GetConfig retrieves the current webhook configuration.
//
// GET /webhook-config
func (c *Client) GetConfig(ctx context.Context) (*WebhookConfig, error) {
	var out GetWebhookConfigResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/webhook-config"), &out); err != nil {
		return nil, err
	}
	return &out.Config, nil
}

// DeleteConfig removes the webhook configuration.
//
// DELETE /webhook-config
func (c *Client) DeleteConfig(ctx context.Context) error {
	return c.hc.Do(ctx, httpclient.NewRequest(http.MethodDelete, "/webhook-config"), nil)
}

// ── Event Parsing & Verification ─────────────────────────────────────────────

// EventType classifies incoming webhook payloads.
type EventType string

const (
	// Payment events
	EventTypePaymentCreated   EventType = "payment.created"
	EventTypePaymentUpdated   EventType = "payment.updated"
	EventTypePaymentCompleted EventType = "payment.completed"
	EventTypePaymentFailed    EventType = "payment.failed"
	EventTypePaymentCancelled EventType = "payment.cancelled"

	// VRP Consent events
	EventTypeVRPConsentCreated EventType = "vrp_consent.created"
	EventTypeVRPConsentUpdated EventType = "vrp_consent.updated"
	EventTypeVRPConsentRevoked EventType = "vrp_consent.revoked"

	// VRP Payment events
	EventTypeVRPCreated   EventType = "vrp.created"
	EventTypeVRPUpdated   EventType = "vrp.updated"
	EventTypeVRPCompleted EventType = "vrp.completed"
	EventTypeVRPFailed    EventType = "vrp.failed"

	// Refund events
	EventTypeRefundCreated   EventType = "refund.created"
	EventTypeRefundUpdated   EventType = "refund.updated"
	EventTypeRefundCompleted EventType = "refund.completed"
	EventTypeRefundFailed    EventType = "refund.failed"

	// Payout events
	EventTypePayoutCreated   EventType = "payout.created"
	EventTypePayoutUpdated   EventType = "payout.updated"
	EventTypePayoutCompleted EventType = "payout.completed"
	EventTypePayoutFailed    EventType = "payout.failed"
)

// Event is the envelope for all Token.io webhook deliveries.
type Event struct {
	ID         string          `json:"id"`
	Type       EventType       `json:"type"`
	APIVersion string          `json:"apiVersion,omitempty"`
	CreatedAt  time.Time       `json:"createdAt"`
	Data       json.RawMessage `json:"data"`
}

// PaymentEventData is the payload for payment-related events.
type PaymentEventData struct {
	PaymentID string `json:"paymentId"`
	Status    string `json:"status"`
	MemberID  string `json:"memberId,omitempty"`
}

// VRPConsentEventData is the payload for VRP consent events.
type VRPConsentEventData struct {
	ConsentID string `json:"consentId"`
	Status    string `json:"status"`
}

// VRPEventData is the payload for VRP payment events.
type VRPEventData struct {
	VrpID     string `json:"vrpId"`
	ConsentID string `json:"consentId"`
	Status    string `json:"status"`
}

// RefundEventData is the payload for refund events.
type RefundEventData struct {
	RefundID   string `json:"refundId"`
	TransferID string `json:"transferId,omitempty"`
	Status     string `json:"status"`
}

// Parse verifies the signature of an incoming webhook and parses the payload.
// signature is the raw value of the X-Token-Signature header.
//
// When no secret is configured (empty string passed to NewClient), signature
// verification is skipped — useful for local development only.
func (c *Client) Parse(payload []byte, signature string) (*Event, error) {
	if len(c.secret) > 0 {
		if err := c.verify(payload, signature); err != nil {
			return nil, fmt.Errorf("webhooks: signature verification failed: %w", err)
		}
	}
	var event Event
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("webhooks: unmarshal event: %w", err)
	}
	return &event, nil
}

// verify checks the HMAC-SHA256 webhook signature.
// Expected format: "t={timestamp},v1={hex_digest}"
func (c *Client) verify(payload []byte, signature string) error {
	tsPart, sigPart, ok := strings.Cut(signature, ",")
	if !ok {
		return fmt.Errorf("malformed signature header (expected t=...,v1=...)")
	}

	tsStr := strings.TrimPrefix(tsPart, "t=")
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp in signature: %w", err)
	}
	// Reject payloads older than 5 minutes to prevent replay attacks.
	if time.Now().Unix()-ts > 300 {
		return fmt.Errorf("webhook timestamp too old (possible replay attack)")
	}

	sigHex := strings.TrimPrefix(sigPart, "v1=")
	signed := fmt.Sprintf("%d.%s", ts, payload)

	mac := hmac.New(sha256.New, c.secret)
	_, _ = mac.Write([]byte(signed)) // hash.Hash.Write never errors
	expected := fmt.Sprintf("%x", mac.Sum(nil))

	if !hmac.Equal([]byte(sigHex), []byte(expected)) {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}

// DecodePaymentEvent decodes a payment event's data field.
func DecodePaymentEvent(e *Event) (*PaymentEventData, error) {
	var d PaymentEventData
	if err := json.Unmarshal(e.Data, &d); err != nil {
		return nil, fmt.Errorf("webhooks: decode payment event: %w", err)
	}
	return &d, nil
}

// DecodeVRPConsentEvent decodes a VRP consent event's data field.
func DecodeVRPConsentEvent(e *Event) (*VRPConsentEventData, error) {
	var d VRPConsentEventData
	if err := json.Unmarshal(e.Data, &d); err != nil {
		return nil, fmt.Errorf("webhooks: decode VRP consent event: %w", err)
	}
	return &d, nil
}

// DecodeVRPEvent decodes a VRP payment event's data field.
func DecodeVRPEvent(e *Event) (*VRPEventData, error) {
	var d VRPEventData
	if err := json.Unmarshal(e.Data, &d); err != nil {
		return nil, fmt.Errorf("webhooks: decode VRP event: %w", err)
	}
	return &d, nil
}

// DecodeRefundEvent decodes a refund event's data field.
func DecodeRefundEvent(e *Event) (*RefundEventData, error) {
	var d RefundEventData
	if err := json.Unmarshal(e.Data, &d); err != nil {
		return nil, fmt.Errorf("webhooks: decode refund event: %w", err)
	}
	return &d, nil
}
