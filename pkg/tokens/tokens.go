// Package tokens provides access to the Token.io Tokens API.
package tokens

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
	"github.com/iamkanishka/tokenio-client-go/pkg/common"
)

// Token represents a Token.io authorization token.
type Token struct {
	ID              string     `json:"id"`
	MemberID        string     `json:"memberId,omitempty"`
	Type            string     `json:"type,omitempty"`
	Status          string     `json:"status,omitempty"`
	CreatedDateTime time.Time  `json:"createdDateTime"`
	UpdatedDateTime time.Time  `json:"updatedDateTime"`
	ExpiresDateTime *time.Time `json:"expiresDateTime,omitempty"`
	Payload         any        `json:"payload,omitempty"`
}

// GetTokensResponse is returned by GET /tokens.
type GetTokensResponse struct {
	Tokens   []Token          `json:"tokens"`
	PageInfo *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetTokenResponse is returned by GET /tokens/{tokenId}.
type GetTokenResponse struct {
	Token Token `json:"token"`
}

// CancelTokenResponse is returned by PUT /tokens/{tokenId}/cancel.
type CancelTokenResponse struct {
	Token Token `json:"token"`
}

// Client exposes the Token.io Tokens API.
type Client struct {
	hc *httpclient.Client
}

// NewClient creates a Tokens client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// GetTokens retrieves a paginated list of tokens.
//
// GET /tokens
func (c *Client) GetTokens(ctx context.Context, limit int, offset string) (*GetTokensResponse, error) {
	if limit <= 0 || limit > 200 {
		return nil, fmt.Errorf("tokens: limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/tokens").
		WithQuery("limit", strconv.Itoa(limit)).
		WithQuery("offset", offset)
	var out GetTokensResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetToken retrieves a single token by ID.
//
// GET /tokens/{tokenId}
func (c *Client) GetToken(ctx context.Context, tokenID string) (*Token, error) {
	if tokenID == "" {
		return nil, fmt.Errorf("tokens: tokenID is required")
	}
	var out GetTokenResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/tokens/"+tokenID), &out); err != nil {
		return nil, err
	}
	return &out.Token, nil
}

// CancelToken cancels an active token.
//
// PUT /tokens/{tokenId}/cancel
func (c *Client) CancelToken(ctx context.Context, tokenID string) (*Token, error) {
	if tokenID == "" {
		return nil, fmt.Errorf("tokens: tokenID is required")
	}
	var out CancelTokenResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPut, "/tokens/"+tokenID+"/cancel"), &out); err != nil {
		return nil, err
	}
	return &out.Token, nil
}
