// Package vrp provides access to the Token.io Variable Recurring Payments API.
package vrp

import (
	"time"

	"github.com/iamkanishka/tokenio-client-go/pkg/common"
)

// ConsentStatus enumerates all VRP consent lifecycle states.
type ConsentStatus string

const (
	ConsentStatusPending                  ConsentStatus = "PENDING"
	ConsentStatusPendingMoreInfo          ConsentStatus = "PENDING_MORE_INFO"
	ConsentStatusPendingRedirectAuth      ConsentStatus = "PENDING_REDIRECT_AUTH"
	ConsentStatusPendingRedirectAuthVerif ConsentStatus = "PENDING_REDIRECT_AUTH_VERIFICATION"
	ConsentStatusAuthorized               ConsentStatus = "AUTHORIZED"
	ConsentStatusRejected                 ConsentStatus = "REJECTED"
	ConsentStatusRevoked                  ConsentStatus = "REVOKED"
	ConsentStatusFailed                   ConsentStatus = "FAILED"
)

// IsFinal reports whether the consent has reached a terminal state.
func (s ConsentStatus) IsFinal() bool {
	switch s {
	case ConsentStatusAuthorized,
		ConsentStatusRejected,
		ConsentStatusRevoked,
		ConsentStatusFailed:
		return true

	case ConsentStatusPending,
		ConsentStatusPendingMoreInfo,
		ConsentStatusPendingRedirectAuth,
		ConsentStatusPendingRedirectAuthVerif:
		return false
	}

	panic("unhandled consent.ConsentStatus: " + string(s))
}

// RequiresRedirect reports whether the PSU must be redirected for auth.
func (s ConsentStatus) RequiresRedirect() bool {
	return s == ConsentStatusPendingRedirectAuth ||
		s == ConsentStatusPendingRedirectAuthVerif
}

// Status enumerates lifecycle states of an individual VRP payment.
// Replaces VrpStatus to avoid revive stutter (vrp.VrpStatus → vrp.Status).
type Status string

const (
	VrpStatusInitiationPending              Status = "INITIATION_PENDING"
	VrpStatusInitiationProcessing           Status = "INITIATION_PROCESSING"
	VrpStatusInitiationCompleted            Status = "INITIATION_COMPLETED"
	VrpStatusInitiationRejected             Status = "INITIATION_REJECTED"
	VrpStatusInitiationRejectedInsufficient Status = "INITIATION_REJECTED_INSUFFICIENT_FUNDS"
	VrpStatusInitiationFailed               Status = "INITIATION_FAILED"
	VrpStatusNoFinalStatus                  Status = "INITIATION_NO_FINAL_STATUS_AVAILABLE"
)

// VrpStatus is a backwards-compatible alias for Status.
type VrpStatus = Status

// IsFinal reports whether the VRP payment is in a terminal state.
func (s Status) IsFinal() bool {
	switch s {
	case VrpStatusInitiationCompleted,
		VrpStatusInitiationRejected,
		VrpStatusInitiationRejectedInsufficient,
		VrpStatusInitiationFailed,
		VrpStatusNoFinalStatus:
		return true

	case VrpStatusInitiationPending,
		VrpStatusInitiationProcessing:
		return false
	}

	panic("unhandled vrp.Status: " + string(s))
}

// Scheme is the VRP consent scheme type.
// Replaces VrpScheme to avoid revive stutter (vrp.VrpScheme → vrp.Scheme).
type Scheme string

// VrpScheme is a backwards-compatible alias for Scheme.
type VrpScheme = Scheme

// VrpSchemeOBLSweeping is the OBL sweeping consent scheme.
const VrpSchemeOBLSweeping Scheme = "OBL_SWEEPING"

// PeriodicLimit defines a maximum spend in a recurring period.
type PeriodicLimit struct {
	MaximumAmount   string `json:"maximumAmount"`
	PeriodType      string `json:"periodType"`
	PeriodAlignment string `json:"periodAlignment"`
}

// ConsentInitiation is the initiation payload for a VRP consent.
// Replaces VrpConsentInitiation to avoid revive stutter.
type ConsentInitiation struct {
	BankID                         string               `json:"bankId"`
	RefID                          string               `json:"refId,omitempty"`
	RemittanceInformationPrimary   string               `json:"remittanceInformationPrimary,omitempty"`
	RemittanceInformationSecondary string               `json:"remittanceInformationSecondary,omitempty"`
	StartDateTime                  string               `json:"startDateTime,omitempty"`
	EndDateTime                    string               `json:"endDateTime,omitempty"`
	OnBehalfOfID                   string               `json:"onBehalfOfId,omitempty"`
	Scheme                         Scheme               `json:"scheme,omitempty"`
	LocalInstrument                string               `json:"localInstrument,omitempty"`
	Debtor                         *common.PartyAccount `json:"debtor,omitempty"`
	Creditor                       *common.PartyAccount `json:"creditor,omitempty"`
	Currency                       string               `json:"currency,omitempty"`
	MinimumIndividualAmount        string               `json:"minimumIndividualAmount,omitempty"`
	MaximumIndividualAmount        string               `json:"maximumIndividualAmount,omitempty"`
	PeriodicLimits                 []PeriodicLimit      `json:"periodicLimits,omitempty"`
	MaximumOccurrences             int                  `json:"maximumOccurrences,omitempty"`
	CallbackURL                    string               `json:"callbackUrl,omitempty"`
	CallbackState                  string               `json:"callbackState,omitempty"`
	ReturnRefundAccount            bool                 `json:"returnRefundAccount,omitempty"`
	Risk                           *common.RiskData     `json:"risk,omitempty"`
}

// VrpConsentInitiation is a backwards-compatible alias for ConsentInitiation.
type VrpConsentInitiation = ConsentInitiation

// Consent is the full VRP consent resource returned by the API.
// Replaces VrpConsent to avoid revive stutter (vrp.VrpConsent → vrp.Consent).
type Consent struct {
	ID                      string                 `json:"id"`
	MemberID                string                 `json:"memberId,omitempty"`
	Initiation              *ConsentInitiation     `json:"initiation,omitempty"`
	CreatedDateTime         time.Time              `json:"createdDateTime"`
	UpdatedDateTime         time.Time              `json:"updatedDateTime"`
	Status                  ConsentStatus          `json:"status"`
	BankVrpConsentID        string                 `json:"bankVrpConsentId,omitempty"`
	BankVrpConsentStatus    string                 `json:"bankVrpConsentStatus,omitempty"`
	StatusReasonInformation string                 `json:"statusReasonInformation,omitempty"`
	Authentication          *common.Authentication `json:"authentication,omitempty"`
}

// VrpConsent is a backwards-compatible alias for Consent.
type VrpConsent = Consent

// IsFinal delegates to ConsentStatus.
func (v Consent) IsFinal() bool { return v.Status.IsFinal() }

// GetRedirectURL safely returns the redirect URL or empty string.
func (v Consent) GetRedirectURL() string {
	if v.Authentication != nil {
		return v.Authentication.RedirectURL
	}
	return ""
}

// Initiation is the initiation payload for a single VRP payment.
// Replaces VrpInitiation to avoid revive stutter (vrp.VrpInitiation → vrp.Initiation).
type Initiation struct {
	ConsentID                      string           `json:"consentId"`
	RefID                          string           `json:"refId,omitempty"`
	RemittanceInformationPrimary   string           `json:"remittanceInformationPrimary,omitempty"`
	RemittanceInformationSecondary string           `json:"remittanceInformationSecondary,omitempty"`
	Amount                         *common.Amount   `json:"amount"`
	ConfirmFunds                   bool             `json:"confirmFunds,omitempty"`
	Risk                           *common.RiskData `json:"risk,omitempty"`
}

// VrpInitiation is a backwards-compatible alias for Initiation.
type VrpInitiation = Initiation

// Vrp is the complete VRP payment resource returned by the API.
type Vrp struct {
	ID                      string                `json:"id"`
	MemberID                string                `json:"memberId,omitempty"`
	Initiation              *Initiation           `json:"initiation,omitempty"`
	CreatedDateTime         time.Time             `json:"createdDateTime"`
	UpdatedDateTime         time.Time             `json:"updatedDateTime"`
	Status                  Status                `json:"status"`
	BankVrpID               string                `json:"bankVrpId,omitempty"`
	BankVrpStatus           string                `json:"bankVrpStatus,omitempty"`
	StatusReasonInformation string                `json:"statusReasonInformation,omitempty"`
	RefundDetails           *common.RefundDetails `json:"refundDetails,omitempty"`
}

// IsFinal delegates to the Status type.
func (v Vrp) IsFinal() bool { return v.Status.IsFinal() }
