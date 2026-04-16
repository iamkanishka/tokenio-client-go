package vrp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
)

// Client exposes the Token.io Variable Recurring Payments API.
// Safe for concurrent use; obtain via the root tokenio.Client.
type Client struct {
	hc *httpclient.Client
}

// NewClient creates a VRP client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// CreateVrpConsent creates a new VRP consent and initiates PSU authorisation.
//
// POST /vrp-consents
func (c *Client) CreateVrpConsent(ctx context.Context, req CreateVrpConsentRequest) (*VrpConsent, error) {
	if req.Initiation.BankID == "" {
		return nil, fmt.Errorf("vrp: Initiation.BankID is required")
	}
	if req.Initiation.Creditor == nil {
		return nil, fmt.Errorf("vrp: Initiation.Creditor is required")
	}
	var out CreateVrpConsentResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, "/vrp-consents").WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out.VrpConsent, nil
}

// GetVrpConsents retrieves a paginated, filtered list of VRP consents.
//
// GET /vrp-consents
func (c *Client) GetVrpConsents(ctx context.Context, req GetVrpConsentsRequest) (*GetVrpConsentsResponse, error) {
	if req.Limit <= 0 || req.Limit > 200 {
		return nil, fmt.Errorf("vrp: Limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/vrp-consents").
		WithQuery("limit", strconv.Itoa(req.Limit)).
		WithQuery("offset", req.Offset).
		WithQuery("createdAfter", req.CreatedAfter).
		WithQuery("createdBefore", req.CreatedBefore).
		WithQuery("onBehalfOfId", req.OnBehalfOfID)
	if req.Scheme != "" {
		r.WithQuery("scheme", string(req.Scheme))
	}
	for _, s := range req.Statuses {
		r.Query().Add("statuses", string(s))
	}
	var out GetVrpConsentsResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetVrpConsent retrieves a single VRP consent by ID.
//
// GET /vrp-consents/{id}
func (c *Client) GetVrpConsent(ctx context.Context, consentID string) (*VrpConsent, error) {
	if consentID == "" {
		return nil, fmt.Errorf("vrp: consentID is required")
	}
	var out GetVrpConsentResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/vrp-consents/"+consentID), &out); err != nil {
		return nil, err
	}
	return &out.VrpConsent, nil
}

// RevokeVrpConsent revokes an active VRP consent.
//
// DELETE /vrp-consents/{id}
func (c *Client) RevokeVrpConsent(ctx context.Context, consentID string) (*VrpConsent, error) {
	if consentID == "" {
		return nil, fmt.Errorf("vrp: consentID is required")
	}
	var out RevokeVrpConsentResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodDelete, "/vrp-consents/"+consentID), &out); err != nil {
		return nil, err
	}
	return &out.VrpConsent, nil
}

// GetVrpConsentPayments retrieves all payments made under a VRP consent.
//
// GET /vrp-consents/{id}/payments
func (c *Client) GetVrpConsentPayments(ctx context.Context, req GetVrpConsentPaymentsRequest) (*GetVrpConsentPaymentsResponse, error) {
	if req.ConsentID == "" {
		return nil, fmt.Errorf("vrp: ConsentID is required")
	}
	if req.Limit <= 0 || req.Limit > 200 {
		return nil, fmt.Errorf("vrp: Limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/vrp-consents/"+req.ConsentID+"/payments").
		WithQuery("limit", strconv.Itoa(req.Limit)).
		WithQuery("offset", req.Offset)
	var out GetVrpConsentPaymentsResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateVrp initiates a single VRP payment against an authorised consent.
//
// POST /vrps
func (c *Client) CreateVrp(ctx context.Context, req CreateVrpRequest) (*Vrp, error) {
	if req.Initiation.ConsentID == "" {
		return nil, fmt.Errorf("vrp: Initiation.ConsentID is required")
	}
	if req.Initiation.Amount == nil {
		return nil, fmt.Errorf("vrp: Initiation.Amount is required")
	}
	var out CreateVrpResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, "/vrps").WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out.Vrp, nil
}

// GetVrps retrieves a paginated, filtered list of VRP payments.
//
// GET /vrps
func (c *Client) GetVrps(ctx context.Context, req GetVrpsRequest) (*GetVrpsResponse, error) {
	if req.Limit <= 0 || req.Limit > 200 {
		return nil, fmt.Errorf("vrp: Limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/vrps").
		WithQuery("limit", strconv.Itoa(req.Limit)).
		WithQuery("offset", req.Offset).
		WithQuery("createdAfter", req.CreatedAfter).
		WithQuery("createdBefore", req.CreatedBefore).
		WithQuery("vrpConsentId", req.VrpConsentID)
	for _, id := range req.IDs {
		r.Query().Add("ids", id)
	}
	for _, s := range req.Statuses {
		r.Query().Add("statuses", string(s))
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
	var out GetVrpsResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetVrp retrieves a single VRP payment by ID.
//
// GET /vrps/{id}
func (c *Client) GetVrp(ctx context.Context, vrpID string) (*Vrp, error) {
	if vrpID == "" {
		return nil, fmt.Errorf("vrp: vrpID is required")
	}
	var out GetVrpResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/vrps/"+vrpID), &out); err != nil {
		return nil, err
	}
	return &out.Vrp, nil
}

// ConfirmFunds checks whether sufficient funds are available for a VRP payment.
//
// GET /vrps/{id}/confirm-funds
func (c *Client) ConfirmFunds(ctx context.Context, consentID, amount string) (bool, error) {
	if consentID == "" {
		return false, fmt.Errorf("vrp: consentID is required")
	}
	if amount == "" {
		return false, fmt.Errorf("vrp: amount is required")
	}
	r := httpclient.NewRequest(http.MethodGet, "/vrps/"+consentID+"/confirm-funds").
		WithQuery("amount", amount)
	var out ConfirmFundsResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return false, err
	}
	return out.FundsAvailable, nil
}
