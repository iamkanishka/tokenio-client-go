// Package accountonfile provides access to the Token.io Account on File API.
// This API allows TPPs to tokenize and store bank account references securely.
package accountonfile

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
)

// TokenizedAccount represents a stored (tokenized) bank account reference.
type TokenizedAccount struct {
	ID              string    `json:"id"`
	MemberID        string    `json:"memberId,omitempty"`
	BankID          string    `json:"bankId,omitempty"`
	AccountRef      string    `json:"accountRef,omitempty"`
	DisplayName     string    `json:"displayName,omitempty"`
	Currency        string    `json:"currency,omitempty"`
	Status          string    `json:"status,omitempty"`
	CreatedDateTime time.Time `json:"createdDateTime"`
	UpdatedDateTime time.Time `json:"updatedDateTime"`
}

// CreateTokenizedAccountRequest is the body for POST /tokenized-accounts.
type CreateTokenizedAccountRequest struct {
	BankID        string `json:"bankId"`
	CallbackURL   string `json:"callbackUrl,omitempty"`
	CallbackState string `json:"callbackState,omitempty"`
}

// CreateTokenizedAccountResponse wraps the created tokenized account.
type CreateTokenizedAccountResponse struct {
	TokenizedAccount TokenizedAccount `json:"tokenizedAccount"`
}

// GetTokenizedAccountResponse wraps a retrieved tokenized account.
type GetTokenizedAccountResponse struct {
	TokenizedAccount TokenizedAccount `json:"tokenizedAccount"`
}

// Client exposes the Token.io Account on File API.
type Client struct {
	hc *httpclient.Client
}

// NewClient creates an AccountOnFile client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// CreateTokenizedAccount creates a new Account on File (tokenized account).
//
// POST /tokenized-accounts
func (c *Client) CreateTokenizedAccount(ctx context.Context, req CreateTokenizedAccountRequest) (*TokenizedAccount, error) {
	if req.BankID == "" {
		return nil, fmt.Errorf("accountonfile: BankID is required")
	}
	var out CreateTokenizedAccountResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, "/tokenized-accounts").WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out.TokenizedAccount, nil
}

// GetTokenizedAccount retrieves a tokenized account by ID.
//
// GET /tokenized-accounts/{id}
func (c *Client) GetTokenizedAccount(ctx context.Context, id string) (*TokenizedAccount, error) {
	if id == "" {
		return nil, fmt.Errorf("accountonfile: id is required")
	}
	var out GetTokenizedAccountResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/tokenized-accounts/"+id), &out); err != nil {
		return nil, err
	}
	return &out.TokenizedAccount, nil
}

// DeleteTokenizedAccount removes a tokenized account.
//
// DELETE /tokenized-accounts/{id}
func (c *Client) DeleteTokenizedAccount(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("accountonfile: id is required")
	}
	return c.hc.Do(ctx, httpclient.NewRequest(http.MethodDelete, "/tokenized-accounts/"+id), nil)
}
