package payments

import "github.com/iamkanishka/tokenio-client-go/pkg/common"

// InitiatePaymentRequest is the body for POST /v2/payments.
type InitiatePaymentRequest struct {
	// Initiation is required.
	Initiation PaymentInitiation `json:"initiation"`
	// PispConsentAccepted indicates the PSU granted PISP consent in the TPP UI.
	PispConsentAccepted bool `json:"pispConsentAccepted,omitempty"`
	// InitialEmbeddedAuth supplies initial embedded auth fields.
	// Key is the field id from bank metadata; value is the field value.
	InitialEmbeddedAuth map[string]string `json:"initialEmbeddedAuth,omitempty"`
}

// InitiatePaymentResponse is returned by POST /v2/payments.
type InitiatePaymentResponse struct {
	Payment Payment `json:"payment"`
}

// GetPaymentsRequest holds query parameters for GET /v2/payments.
type GetPaymentsRequest struct {
	Limit                int
	Offset               string
	IDs                  []string
	InvertIDs            bool
	Statuses             []Status
	InvertStatuses       bool
	CreatedAfter         string // ISO 8601
	CreatedBefore        string // ISO 8601
	RefIDs               []string
	OnBehalfOfID         string
	RefundStatuses       []string
	Partial              bool
	ExternalPsuReference string
	Type                 PaymentType
	VRPConsentID         string
}

// GetPaymentsResponse is returned by GET /v2/payments.
type GetPaymentsResponse struct {
	Payments []Payment        `json:"payments"`
	PageInfo *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetPaymentResponse is returned by GET /v2/payments/{paymentId}.
type GetPaymentResponse struct {
	Payment Payment `json:"payment"`
}

// ProvideEmbeddedAuthRequest is the body for POST /v2/payments/{id}/embedded-auth.
type ProvideEmbeddedAuthRequest struct {
	// PaymentID is the payment being authenticated (path param, not in body).
	PaymentID string `json:"-"`
	// EmbeddedAuth maps field IDs to their PSU-supplied values.
	EmbeddedAuth map[string]string `json:"embeddedAuth"`
}

// ProvideEmbeddedAuthResponse wraps the updated payment.
type ProvideEmbeddedAuthResponse struct {
	Payment Payment `json:"payment"`
}

// GenerateQRCodeRequest holds the query params for GET /qr-code.
type GenerateQRCodeRequest struct {
	// Data is the URL-encoded URL to encode into the QR code.
	Data string
}
