// Package payouts provides access to the Token.io Payouts API.
package payouts

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
	"github.com/iamkanishka/tokenio-client-go/pkg/common"
)

// PayoutStatus enumerates payout lifecycle states.
type PayoutStatus string

const (
	PayoutStatusPending    PayoutStatus = "INITIATION_PENDING"
	PayoutStatusProcessing PayoutStatus = "INITIATION_PROCESSING"
	PayoutStatusCompleted  PayoutStatus = "INITIATION_COMPLETED"
	PayoutStatusRejected   PayoutStatus = "INITIATION_REJECTED"
	PayoutStatusFailed     PayoutStatus = "INITIATION_FAILED"
)

// IsFinal reports whether the payout is in a terminal state.
func (s PayoutStatus) IsFinal() bool {
	switch s {
	case PayoutStatusCompleted,
		PayoutStatusRejected,
		PayoutStatusFailed:
		return true

	case PayoutStatusPending,
		PayoutStatusProcessing:
		return false
	}

	panic("unhandled payouts.PayoutStatus: " + string(s))
}

// PayoutInitiation is the initiation payload for a payout.
type PayoutInitiation struct {
	RefID                          string               `json:"refId,omitempty"`
	Amount                         *common.Amount       `json:"amount"`
	RemittanceInformationPrimary   string               `json:"remittanceInformationPrimary,omitempty"`
	RemittanceInformationSecondary string               `json:"remittanceInformationSecondary,omitempty"`
	Creditor                       *common.PartyAccount `json:"creditor"`
	Debtor                         *common.PartyAccount `json:"debtor,omitempty"`
	ExecutionDate                  string               `json:"executionDate,omitempty"` // YYYY-MM-DD
	LocalInstrument                string               `json:"localInstrument,omitempty"`
}

// Payout is the complete payout resource.
type Payout struct {
	ID                      string            `json:"id"`
	MemberID                string            `json:"memberId,omitempty"`
	Initiation              *PayoutInitiation `json:"initiation,omitempty"`
	Status                  PayoutStatus      `json:"status"`
	StatusReasonInformation string            `json:"statusReasonInformation,omitempty"`
	ErrorInfo               *common.ErrorInfo `json:"errorInfo,omitempty"`
	CreatedDateTime         time.Time         `json:"createdDateTime"`
	UpdatedDateTime         time.Time         `json:"updatedDateTime"`
}

// IsFinal delegates to PayoutStatus.
func (p Payout) IsFinal() bool { return p.Status.IsFinal() }

// InitiatePayoutRequest is the body for POST /payouts.
type InitiatePayoutRequest struct {
	PayoutInitiation PayoutInitiation `json:"payoutInitiation"`
}

// InitiatePayoutResponse wraps the created payout.
type InitiatePayoutResponse struct {
	Payout Payout `json:"payout"`
}

// GetPayoutsRequest holds query params for GET /payouts.
type GetPayoutsRequest struct {
	Limit         int
	Offset        string
	CreatedAfter  string
	CreatedBefore string
}

// GetPayoutsResponse is returned by GET /payouts.
type GetPayoutsResponse struct {
	Payouts  []Payout         `json:"payouts"`
	PageInfo *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetPayoutResponse is returned by GET /payouts/{payoutId}.
type GetPayoutResponse struct {
	Payout Payout `json:"payout"`
}

// Client exposes the Token.io Payouts API.
type Client struct {
	hc *httpclient.Client
}

// NewClient creates a Payouts client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// InitiatePayout creates a new payout from a settlement account.
//
// POST /payouts
func (c *Client) InitiatePayout(ctx context.Context, req InitiatePayoutRequest) (*Payout, error) {
	if req.PayoutInitiation.Amount == nil {
		return nil, fmt.Errorf("payouts: PayoutInitiation.Amount is required")
	}
	if req.PayoutInitiation.Creditor == nil {
		return nil, fmt.Errorf("payouts: PayoutInitiation.Creditor is required")
	}
	var out InitiatePayoutResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, "/payouts").WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out.Payout, nil
}

// GetPayouts retrieves a paginated list of payouts.
//
// GET /payouts
func (c *Client) GetPayouts(ctx context.Context, req GetPayoutsRequest) (*GetPayoutsResponse, error) {
	if req.Limit <= 0 || req.Limit > 200 {
		return nil, fmt.Errorf("payouts: Limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/payouts").
		WithQuery("limit", strconv.Itoa(req.Limit)).
		WithQuery("offset", req.Offset).
		WithQuery("createdAfter", req.CreatedAfter).
		WithQuery("createdBefore", req.CreatedBefore)
	var out GetPayoutsResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetPayout retrieves a single payout by ID.
//
// GET /payouts/{payoutId}
func (c *Client) GetPayout(ctx context.Context, payoutID string) (*Payout, error) {
	if payoutID == "" {
		return nil, fmt.Errorf("payouts: payoutID is required")
	}
	var out GetPayoutResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/payouts/"+payoutID), &out); err != nil {
		return nil, err
	}
	return &out.Payout, nil
}
