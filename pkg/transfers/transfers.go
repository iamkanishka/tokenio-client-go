// Package transfers provides access to the Token.io Transfers API (Payments v1).
// Use this for redeeming transfer tokens created via the Token Requests flow.
package transfers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
	"github.com/iamkanishka/tokenio-client-go/pkg/common"
)

// TransferStatus enumerates transfer lifecycle states.
type TransferStatus string

const (
	TransferStatusPending    TransferStatus = "PENDING"
	TransferStatusProcessing TransferStatus = "PROCESSING"
	TransferStatusSuccess    TransferStatus = "SUCCESS"
	TransferStatusFailed     TransferStatus = "FAILED"
	TransferStatusCanceled   TransferStatus = "CANCELED"
)

// IsFinal reports whether the transfer is in a terminal state.
func (s TransferStatus) IsFinal() bool {
	switch s {
	case TransferStatusSuccess, TransferStatusFailed, TransferStatusCanceled:
		return true
	default:
		return false
	}
}

// Transfer is the complete transfer resource returned by the API.
type Transfer struct {
	ID                      string                `json:"id"`
	MemberID                string                `json:"memberId,omitempty"`
	TokenID                 string                `json:"tokenId,omitempty"`
	TransactionID           string                `json:"transactionId,omitempty"`
	Status                  TransferStatus        `json:"status"`
	StatusReasonInformation string                `json:"statusReasonInformation,omitempty"`
	Amount                  *common.Amount        `json:"amount,omitempty"`
	Creditor                *common.PartyAccount  `json:"creditor,omitempty"`
	Debtor                  *common.PartyAccount  `json:"debtor,omitempty"`
	RefundDetails           *common.RefundDetails `json:"refundDetails,omitempty"`
	ErrorInfo               *common.ErrorInfo     `json:"errorInfo,omitempty"`
	CreatedDateTime         time.Time             `json:"createdDateTime"`
	UpdatedDateTime         time.Time             `json:"updatedDateTime"`
}

// IsFinal delegates to TransferStatus.
func (t Transfer) IsFinal() bool { return t.Status.IsFinal() }

// RedeemTransferTokenRequest is the body for POST /transfers.
type RedeemTransferTokenRequest struct {
	// TokenID is the transfer token to redeem.
	TokenID string         `json:"tokenId"`
	Amount  *common.Amount `json:"amount,omitempty"`
	RefID   string         `json:"refId,omitempty"`
}

// RedeemTransferTokenResponse wraps the created transfer.
type RedeemTransferTokenResponse struct {
	Transfer Transfer `json:"transfer"`
}

// GetTransfersRequest holds query parameters for GET /transfers.
type GetTransfersRequest struct {
	Limit         int
	Offset        string
	TokenID       string
	CreatedAfter  string // ISO 8601
	CreatedBefore string // ISO 8601
}

// GetTransfersResponse is returned by GET /transfers.
type GetTransfersResponse struct {
	Transfers []Transfer       `json:"transfers"`
	PageInfo  *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetTransferResponse is returned by GET /transfers/{transferId}.
type GetTransferResponse struct {
	Transfer Transfer `json:"transfer"`
}

// Client exposes the Token.io Transfers API (Payments v1).
type Client struct {
	hc *httpclient.Client
}

// NewClient creates a Transfers client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// RedeemTransferToken redeems an authorised transfer token to create a payment.
//
// POST /transfers
func (c *Client) RedeemTransferToken(ctx context.Context, req RedeemTransferTokenRequest) (*Transfer, error) {
	if req.TokenID == "" {
		return nil, fmt.Errorf("transfers: TokenID is required")
	}
	var out RedeemTransferTokenResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, "/transfers").WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out.Transfer, nil
}

// GetTransfers retrieves a paginated list of transfers.
//
// GET /transfers
func (c *Client) GetTransfers(ctx context.Context, req GetTransfersRequest) (*GetTransfersResponse, error) {
	if req.Limit <= 0 || req.Limit > 200 {
		return nil, fmt.Errorf("transfers: Limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/transfers").
		WithQuery("limit", strconv.Itoa(req.Limit)).
		WithQuery("offset", req.Offset).
		WithQuery("tokenId", req.TokenID).
		WithQuery("createdAfter", req.CreatedAfter).
		WithQuery("createdBefore", req.CreatedBefore)
	var out GetTransfersResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTransfer retrieves a single transfer by ID.
//
// GET /transfers/{transferId}
func (c *Client) GetTransfer(ctx context.Context, transferID string) (*Transfer, error) {
	if transferID == "" {
		return nil, fmt.Errorf("transfers: transferID is required")
	}
	var out GetTransferResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/transfers/"+transferID), &out); err != nil {
		return nil, err
	}
	return &out.Transfer, nil
}
