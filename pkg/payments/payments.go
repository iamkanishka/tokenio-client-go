package payments

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
)

const basePath = "/v2/payments"

// Client exposes the Token.io Payments v2 API.
// Safe for concurrent use; obtain via the root tokenio.Client.
type Client struct {
	hc *httpclient.Client
}

// NewClient creates a Payments client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// InitiatePayment creates a new single immediate or future-dated payment.
//
// POST /v2/payments
func (c *Client) InitiatePayment(ctx context.Context, req InitiatePaymentRequest) (*Payment, error) {
	if req.Initiation.BankID == "" {
		return nil, fmt.Errorf("payments: Initiation.BankID is required")
	}
	var out InitiatePaymentResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, basePath).WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out.Payment, nil
}

// GetPayments retrieves a paginated, filtered list of payments.
//
// GET /v2/payments
func (c *Client) GetPayments(ctx context.Context, req GetPaymentsRequest) (*GetPaymentsResponse, error) {
	if req.Limit <= 0 || req.Limit > 200 {
		return nil, fmt.Errorf("payments: Limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, basePath).
		WithQuery("limit", strconv.Itoa(req.Limit)).
		WithQuery("offset", req.Offset).
		WithQuery("onBehalfOfId", req.OnBehalfOfID).
		WithQuery("createdAfter", req.CreatedAfter).
		WithQuery("createdBefore", req.CreatedBefore).
		WithQuery("externalPsuReference", req.ExternalPsuReference).
		WithQuery("vrpConsentId", req.VRPConsentID)

	if req.Type != "" {
		r.WithQuery("type", string(req.Type))
	}
	for _, id := range req.IDs {
		r.Query().Add("ids", id)
	}
	for _, s := range req.Statuses {
		r.Query().Add("statuses", string(s))
	}
	for _, s := range req.RefundStatuses {
		r.Query().Add("refundStatuses", s)
	}
	for _, id := range req.RefIDs {
		r.Query().Add("refIds", id)
	}
	if req.InvertIDs {
		r.WithQuery("invertIds", "true")
	}
	if req.InvertStatuses {
		r.WithQuery("invertStatuses", "true")
	}
	if req.Partial {
		r.WithQuery("partial", "true")
	}

	var out GetPaymentsResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPayment retrieves a single payment by ID.
//
// GET /v2/payments/{paymentId}
func (c *Client) GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("payments: paymentID is required")
	}
	var out GetPaymentResponse
	r := httpclient.NewRequest(http.MethodGet, basePath+"/"+paymentID)
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out.Payment, nil
}

// GetPaymentWithTimeout retrieves a payment with a server-side request timeout.
// requestTimeout is the number of seconds Token.io should wait before returning.
//
// GET /v2/payments/{paymentId}
func (c *Client) GetPaymentWithTimeout(ctx context.Context, paymentID string, requestTimeout int) (*Payment, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("payments: paymentID is required")
	}
	r := httpclient.NewRequest(http.MethodGet, basePath+"/"+paymentID).
		WithHeader("request-timeout", strconv.Itoa(requestTimeout))
	var out GetPaymentResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out.Payment, nil
}

// ProvideEmbeddedAuth submits PSU credentials during embedded authentication.
// Called when payment status is INITIATION_PENDING_EMBEDDED_AUTH.
//
// POST /v2/payments/{paymentId}/embedded-auth
func (c *Client) ProvideEmbeddedAuth(ctx context.Context, req ProvideEmbeddedAuthRequest) (*Payment, error) {
	if req.PaymentID == "" {
		return nil, fmt.Errorf("payments: PaymentID is required")
	}
	if len(req.EmbeddedAuth) == 0 {
		return nil, fmt.Errorf("payments: EmbeddedAuth must not be empty")
	}
	path := basePath + "/" + req.PaymentID + "/embedded-auth"
	var out ProvideEmbeddedAuthResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, path).WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out.Payment, nil
}

// GenerateQRCode generates a 240×240 SVG QR code for the given URL.
// Returns the raw SVG bytes.
//
// GET /qr-code
func (c *Client) GenerateQRCode(ctx context.Context, req GenerateQRCodeRequest) ([]byte, error) {
	if req.Data == "" {
		return nil, fmt.Errorf("payments: Data (URL) is required")
	}
	r := httpclient.NewRequest(http.MethodGet, "/qr-code").
		WithQuery("data", req.Data)
	body, status, err := c.hc.DoRaw(ctx, r)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("payments: QR code generation failed with HTTP %d: %s",
			status, strings.TrimSpace(string(body)))
	}
	return body, nil
}

// PollUntilFinal polls GetPayment every opts.Interval until the payment
// reaches a final state or ctx is cancelled. Prefer webhooks in production.
func (c *Client) PollUntilFinal(ctx context.Context, paymentID string, opts PollOptions) (*Payment, error) {
	if paymentID == "" {
		return nil, fmt.Errorf("payments: paymentID is required")
	}
	opts.applyDefaults()

	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			p, err := c.GetPayment(ctx, paymentID)
			if err != nil {
				return nil, err
			}
			if p.IsFinal() {
				return p, nil
			}
		}
	}
}
