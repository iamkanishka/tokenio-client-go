// Package verification provides access to the Token.io Verification API.
package verification

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
	"github.com/iamkanishka/tokenio-client-go/pkg/common"
)

// Status enumerates account verification check states.
// Replaces VerificationStatus to avoid revive stutter.
type Status string

// VerificationStatus is a backwards-compatible alias for Status.
type VerificationStatus = Status

const (
	StatusPending   Status = "PENDING"
	StatusCompleted Status = "COMPLETED"
	StatusFailed    Status = "FAILED"
)

// Check represents an account verification result.
// Replaces VerificationCheck to avoid revive stutter.
type Check struct {
	ID              string            `json:"id"`
	Status          Status            `json:"status"`
	AccountVerified bool              `json:"accountVerified,omitempty"`
	NameMatched     *bool             `json:"nameMatched,omitempty"`
	ErrorInfo       *common.ErrorInfo `json:"errorInfo,omitempty"`
	CreatedDateTime time.Time         `json:"createdDateTime"`
	UpdatedDateTime time.Time         `json:"updatedDateTime"`
}

// VerificationCheck is a backwards-compatible alias for Check.
type VerificationCheck = Check

// IsFinal reports whether the verification check is complete.
func (v Check) IsFinal() bool {
	return v.Status == StatusCompleted || v.Status == StatusFailed
}

// InitiateVerificationRequest is the body for POST /verification.
type InitiateVerificationRequest struct {
	BankID        string               `json:"bankId"`
	Account       *common.PartyAccount `json:"account"`
	CallbackURL   string               `json:"callbackUrl,omitempty"`
	CallbackState string               `json:"callbackState,omitempty"`
}

// InitiateVerificationResponse wraps the created verification check.
type InitiateVerificationResponse struct {
	Verification   Check                  `json:"verification"`
	Authentication *common.Authentication `json:"authentication,omitempty"`
}

// Client exposes the Token.io Verification API.
type Client struct {
	hc *httpclient.Client
}

// NewClient creates a Verification client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// InitiateVerification starts an account ownership verification check.
//
// POST /verification
func (c *Client) InitiateVerification(ctx context.Context, req InitiateVerificationRequest) (*InitiateVerificationResponse, error) {
	if req.BankID == "" {
		return nil, fmt.Errorf("verification: BankID is required")
	}
	if req.Account == nil {
		return nil, fmt.Errorf("verification: Account is required")
	}
	var out InitiateVerificationResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, "/verification").WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out, nil
}
