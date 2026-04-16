// Package reports provides access to the Token.io Reports API.
// This API returns operational status information for connected banks.
package reports

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
)

// BankStatus represents the operational status of a bank connection.
type BankStatus struct {
	BankID        string    `json:"bankId"`
	Status        string    `json:"status,omitempty"` // UP, DOWN, DEGRADED
	StatusMessage string    `json:"statusMessage,omitempty"`
	LastChecked   time.Time `json:"lastChecked,omitempty"`
	AISAvailable  bool      `json:"aisAvailable,omitempty"`
	PISAvailable  bool      `json:"pisAvailable,omitempty"`
	VRPAvailable  bool      `json:"vrpAvailable,omitempty"`
}

// GetBankStatusesResponse is returned by GET /bank-statuses.
type GetBankStatusesResponse struct {
	BankStatuses []BankStatus `json:"bankStatuses"`
}

// GetBankStatusResponse is returned by GET /bank-statuses/{bankId}.
type GetBankStatusResponse struct {
	BankStatus BankStatus `json:"bankStatus"`
}

// Client exposes the Token.io Reports API.
type Client struct {
	hc *httpclient.Client
}

// NewClient creates a Reports client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// GetBankStatuses retrieves operational status for all connected banks.
//
// GET /bank-statuses
func (c *Client) GetBankStatuses(ctx context.Context) (*GetBankStatusesResponse, error) {
	var out GetBankStatusesResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/bank-statuses"), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetBankStatus retrieves operational status for a single bank.
//
// GET /bank-statuses/{bankId}
func (c *Client) GetBankStatus(ctx context.Context, bankID string) (*BankStatus, error) {
	if bankID == "" {
		return nil, fmt.Errorf("reports: bankID is required")
	}
	var out GetBankStatusResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/bank-statuses/"+bankID), &out); err != nil {
		return nil, err
	}
	return &out.BankStatus, nil
}
