package vrp

import "github.com/iamkanishka/tokenio-client-go/pkg/common"

// CreateVrpConsentRequest is the body for POST /vrp-consents.
type CreateVrpConsentRequest struct {
	Initiation          VrpConsentInitiation `json:"initiation"`
	PispConsentAccepted bool                 `json:"pispConsentAccepted,omitempty"`
}

// CreateVrpConsentResponse wraps the created consent.
type CreateVrpConsentResponse struct {
	VrpConsent VrpConsent `json:"vrpConsent"`
}

// GetVrpConsentsRequest holds query params for GET /vrp-consents.
type GetVrpConsentsRequest struct {
	Limit         int
	Offset        string
	CreatedAfter  string // ISO 8601
	CreatedBefore string // ISO 8601
	Statuses      []ConsentStatus
	Scheme        VrpScheme
	OnBehalfOfID  string
}

// GetVrpConsentsResponse is returned by GET /vrp-consents.
type GetVrpConsentsResponse struct {
	VrpConsents []VrpConsent     `json:"vrpConsents"`
	PageInfo    *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetVrpConsentResponse is returned by GET /vrp-consents/{id}.
type GetVrpConsentResponse struct {
	VrpConsent VrpConsent `json:"vrpConsent"`
}

// RevokeVrpConsentResponse is returned by DELETE /vrp-consents/{id}.
type RevokeVrpConsentResponse struct {
	VrpConsent VrpConsent `json:"vrpConsent"`
}

// GetVrpConsentPaymentsRequest holds query params for GET /vrp-consents/{id}/payments.
type GetVrpConsentPaymentsRequest struct {
	ConsentID string
	Limit     int
	Offset    string
}

// GetVrpConsentPaymentsResponse is returned by GET /vrp-consents/{id}/payments.
type GetVrpConsentPaymentsResponse struct {
	Vrps     []Vrp            `json:"vrps"`
	PageInfo *common.PageInfo `json:"pageInfo,omitempty"`
}

// CreateVrpRequest is the body for POST /vrps.
type CreateVrpRequest struct {
	Initiation VrpInitiation `json:"initiation"`
}

// CreateVrpResponse wraps the created VRP payment.
type CreateVrpResponse struct {
	Vrp Vrp `json:"vrp"`
}

// GetVrpsRequest holds query params for GET /vrps.
type GetVrpsRequest struct {
	Limit          int
	Offset         string
	IDs            []string
	InvertIDs      bool
	Statuses       []VrpStatus
	InvertStatuses bool
	CreatedAfter   string // ISO 8601
	CreatedBefore  string // ISO 8601
	RefIDs         []string
	VrpConsentID   string
}

// GetVrpsResponse is returned by GET /vrps.
type GetVrpsResponse struct {
	Vrps     []Vrp            `json:"vrps"`
	PageInfo *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetVrpResponse is returned by GET /vrps/{id}.
type GetVrpResponse struct {
	Vrp Vrp `json:"vrp"`
}

// ConfirmFundsResponse is returned by GET /vrps/{id}/confirm-funds.
type ConfirmFundsResponse struct {
	FundsAvailable bool `json:"fundsAvailable"`
}
