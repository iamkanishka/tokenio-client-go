package refunds

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
	"github.com/iamkanishka/tokenio-client-go/pkg/common"
)

// InitiateRefundRequest is the body for POST /refunds.
type InitiateRefundRequest struct {
	RefundInitiation RefundInitiation `json:"refundInitiation"`
}

// InitiateRefundResponse wraps the created refund.
type InitiateRefundResponse struct {
	Refund Refund `json:"refund"`
}

// GetRefundsRequest holds query params for GET /refunds.
type GetRefundsRequest struct {
	Limit         int
	Offset        string
	TransferID    string
	CreatedAfter  string
	CreatedBefore string
}

// GetRefundsResponse is returned by GET /refunds.
type GetRefundsResponse struct {
	Refunds  []Refund         `json:"refunds"`
	PageInfo *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetRefundResponse is returned by GET /refunds/{refundId}.
type GetRefundResponse struct {
	Refund Refund `json:"refund"`
}

// GetTransferRefundsResponse is returned by GET /refunds?transferId=...
type GetTransferRefundsResponse struct {
	Refunds  []Refund         `json:"refunds"`
	PageInfo *common.PageInfo `json:"pageInfo,omitempty"`
}

// Client exposes the Token.io Refunds API.
type Client struct {
	hc *httpclient.Client
}

// NewClient creates a Refunds client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// InitiateRefund creates a refund for a previously completed transfer.
//
// POST /refunds
func (c *Client) InitiateRefund(ctx context.Context, req InitiateRefundRequest) (*Refund, error) {
	if req.RefundInitiation.TransferID == "" {
		return nil, fmt.Errorf("refunds: RefundInitiation.TransferID is required")
	}
	if req.RefundInitiation.Amount == nil {
		return nil, fmt.Errorf("refunds: RefundInitiation.Amount is required")
	}
	var out InitiateRefundResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, "/refunds").WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out.Refund, nil
}

// GetRefunds retrieves a paginated list of refunds.
//
// GET /refunds
func (c *Client) GetRefunds(ctx context.Context, req GetRefundsRequest) (*GetRefundsResponse, error) {
	if req.Limit <= 0 || req.Limit > 200 {
		return nil, fmt.Errorf("refunds: Limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/refunds").
		WithQuery("limit", strconv.Itoa(req.Limit)).
		WithQuery("offset", req.Offset).
		WithQuery("transferId", req.TransferID).
		WithQuery("createdAfter", req.CreatedAfter).
		WithQuery("createdBefore", req.CreatedBefore)
	var out GetRefundsResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRefund retrieves a single refund by ID.
//
// GET /refunds/{refundId}
func (c *Client) GetRefund(ctx context.Context, refundID string) (*Refund, error) {
	if refundID == "" {
		return nil, fmt.Errorf("refunds: refundID is required")
	}
	var out GetRefundResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/refunds/"+refundID), &out); err != nil {
		return nil, err
	}
	return &out.Refund, nil
}

// GetTransferRefunds retrieves all refunds for a specific transfer.
//
// GET /refunds?transferId={transferId}
func (c *Client) GetTransferRefunds(ctx context.Context, transferID string, limit int, offset string) (*GetTransferRefundsResponse, error) {
	if transferID == "" {
		return nil, fmt.Errorf("refunds: transferID is required")
	}
	if limit <= 0 || limit > 200 {
		return nil, fmt.Errorf("refunds: limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/refunds").
		WithQuery("transferId", transferID).
		WithQuery("limit", strconv.Itoa(limit)).
		WithQuery("offset", offset)
	var out GetTransferRefundsResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
