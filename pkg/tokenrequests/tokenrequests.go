// Package tokenrequests provides access to the Token.io Token Requests API.
// This is the legacy v1 flow used for Payments v1 and AIS (Account Information Services).
package tokenrequests

import (
	"context"
	"fmt"
	"net/http"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
)

// TokenRequestStatus describes the state of a token request.
type TokenRequestStatus string

const (
	TokenRequestStatusPending    TokenRequestStatus = "PENDING"
	TokenRequestStatusAuthorised TokenRequestStatus = "AUTHORISED"
	TokenRequestStatusDeclined   TokenRequestStatus = "DECLINED"
	TokenRequestStatusExpired    TokenRequestStatus = "EXPIRED"
	TokenRequestStatusRevoked    TokenRequestStatus = "REVOKED"
)

// IsFinal reports whether the token request is in a terminal state.
func (s TokenRequestStatus) IsFinal() bool {
	switch s {
	case TokenRequestStatusAuthorised, TokenRequestStatusDeclined,
		TokenRequestStatusExpired, TokenRequestStatusRevoked:
		return true
	default:
		return false
	}
}

// TokenRequest represents a stored token request resource.
type TokenRequest struct {
	TokenRequestID string             `json:"tokenRequestId"`
	Status         TokenRequestStatus `json:"status,omitempty"`
	RedirectURL    string             `json:"redirectUrl,omitempty"`
	TokenID        string             `json:"tokenId,omitempty"`
	RequestPayload any                `json:"requestPayload,omitempty"`
}

// StoreTokenRequestRequest is the body for POST /token-requests.
type StoreTokenRequestRequest struct {
	// RequestPayload is the token request payload (payment or access token spec).
	RequestPayload map[string]any `json:"requestPayload"`
	// Options are additional options for the token request.
	Options map[string]any `json:"options,omitempty"`
	// RedirectURL is the URL to redirect the PSU to after authorization.
	RedirectURL string `json:"redirectUrl,omitempty"`
}

// StoreTokenRequestResponse wraps the created token request.
type StoreTokenRequestResponse struct {
	TokenRequest TokenRequest `json:"tokenRequest"`
}

// GetTokenRequestResponse is returned by GET /token-requests/{requestId}.
type GetTokenRequestResponse struct {
	TokenRequest TokenRequest `json:"tokenRequest"`
}

// TokenRequestResult holds the authorisation result of a token request.
type TokenRequestResult struct {
	TokenRequestID string             `json:"tokenRequestId"`
	Status         TokenRequestStatus `json:"status"`
	TokenID        string             `json:"tokenId,omitempty"`
	TransferID     string             `json:"transferId,omitempty"`
}

// GetTokenRequestResultResponse is returned by GET /token-requests/{requestId}/result.
type GetTokenRequestResultResponse struct {
	TokenRequestResult TokenRequestResult `json:"tokenRequestResult"`
}

// InitiateBankAuthRequest is the body for POST /token-requests/{requestId}/authorize.
type InitiateBankAuthRequest struct {
	TokenRequestID string `json:"-"`
	BankID         string `json:"bankId"`
}

// InitiateBankAuthResponse wraps the authorization redirect URL.
type InitiateBankAuthResponse struct {
	RedirectURL string `json:"redirectUrl,omitempty"`
}

// Client exposes the Token.io Token Requests API.
type Client struct {
	hc *httpclient.Client
}

// NewClient creates a TokenRequests client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// StoreTokenRequest creates a new token request (v1 payment or AIS flow).
//
// POST /token-requests
func (c *Client) StoreTokenRequest(ctx context.Context, req StoreTokenRequestRequest) (*TokenRequest, error) {
	if len(req.RequestPayload) == 0 {
		return nil, fmt.Errorf("tokenrequests: RequestPayload is required")
	}
	var out StoreTokenRequestResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, "/token-requests").WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out.TokenRequest, nil
}

// GetTokenRequest retrieves a token request by its ID.
//
// GET /token-requests/{requestId}
func (c *Client) GetTokenRequest(ctx context.Context, requestID string) (*TokenRequest, error) {
	if requestID == "" {
		return nil, fmt.Errorf("tokenrequests: requestID is required")
	}
	var out GetTokenRequestResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/token-requests/"+requestID), &out); err != nil {
		return nil, err
	}
	return &out.TokenRequest, nil
}

// GetTokenRequestResult retrieves the authorisation result of a token request.
//
// GET /token-requests/{requestId}/result
func (c *Client) GetTokenRequestResult(ctx context.Context, requestID string) (*TokenRequestResult, error) {
	if requestID == "" {
		return nil, fmt.Errorf("tokenrequests: requestID is required")
	}
	var out GetTokenRequestResultResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/token-requests/"+requestID+"/result"), &out); err != nil {
		return nil, err
	}
	return &out.TokenRequestResult, nil
}

// InitiateBankAuthorization redirects the PSU to authorise the token request at their bank.
//
// POST /token-requests/{requestId}/authorize
func (c *Client) InitiateBankAuthorization(ctx context.Context, req InitiateBankAuthRequest) (*InitiateBankAuthResponse, error) {
	if req.TokenRequestID == "" {
		return nil, fmt.Errorf("tokenrequests: TokenRequestID is required")
	}
	if req.BankID == "" {
		return nil, fmt.Errorf("tokenrequests: BankID is required")
	}
	path := "/token-requests/" + req.TokenRequestID + "/authorize"
	var out InitiateBankAuthResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, path).WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
