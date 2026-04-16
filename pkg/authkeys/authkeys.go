// Package authkeys provides access to the Token.io Authentication Keys API.
// This API manages the RSA/EC public keys used for request signing.
package authkeys

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
)

// MemberKey represents a public authentication key.
type MemberKey struct {
	ID              string    `json:"id"`
	Algorithm       string    `json:"algorithm,omitempty"` // RSA, EC
	PublicKey       string    `json:"publicKey,omitempty"` // PEM or JWK
	Level           string    `json:"level,omitempty"`
	Status          string    `json:"status,omitempty"`
	CreatedDateTime time.Time `json:"createdDateTime"`
	UpdatedDateTime time.Time `json:"updatedDateTime"`
}

// SubmitKeyRequest is the body for POST /member-keys.
type SubmitKeyRequest struct {
	// PublicKey is the PEM or JWK-encoded public key.
	PublicKey string `json:"publicKey"`
	Algorithm string `json:"algorithm,omitempty"`
	Level     string `json:"level,omitempty"`
}

// SubmitKeyResponse wraps the submitted key.
type SubmitKeyResponse struct {
	MemberKey MemberKey `json:"memberKey"`
}

// GetKeysResponse is returned by GET /member-keys.
type GetKeysResponse struct {
	MemberKeys []MemberKey `json:"memberKeys"`
}

// GetKeyResponse is returned by GET /member-keys/{keyId}.
type GetKeyResponse struct {
	MemberKey MemberKey `json:"memberKey"`
}

// Client exposes the Token.io Authentication Keys API.
type Client struct {
	hc *httpclient.Client
}

// NewClient creates an AuthKeys client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// SubmitKey registers a new public key for request signing.
//
// POST /member-keys
func (c *Client) SubmitKey(ctx context.Context, req SubmitKeyRequest) (*MemberKey, error) {
	if req.PublicKey == "" {
		return nil, fmt.Errorf("authkeys: PublicKey is required")
	}
	var out SubmitKeyResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, "/member-keys").WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out.MemberKey, nil
}

// GetKeys retrieves all registered public keys for the calling member.
//
// GET /member-keys
func (c *Client) GetKeys(ctx context.Context) (*GetKeysResponse, error) {
	var out GetKeysResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/member-keys"), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetKey retrieves a single key by ID.
//
// GET /member-keys/{keyId}
func (c *Client) GetKey(ctx context.Context, keyID string) (*MemberKey, error) {
	if keyID == "" {
		return nil, fmt.Errorf("authkeys: keyID is required")
	}
	var out GetKeyResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/member-keys/"+keyID), &out); err != nil {
		return nil, err
	}
	return &out.MemberKey, nil
}

// DeleteKey removes a public key.
//
// DELETE /member-keys/{keyId}
func (c *Client) DeleteKey(ctx context.Context, keyID string) error {
	if keyID == "" {
		return fmt.Errorf("authkeys: keyID is required")
	}
	return c.hc.Do(ctx, httpclient.NewRequest(http.MethodDelete, "/member-keys/"+keyID), nil)
}
