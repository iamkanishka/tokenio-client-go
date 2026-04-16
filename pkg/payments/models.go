// Package payments provides access to the Token.io Payments v2 API.
package payments

import (
	"time"

	"github.com/iamkanishka/tokenio-client-go/pkg/common"
)

// Status enumerates all payment lifecycle states from the Token.io API.
type Status string

const (
	StatusInitiationPending	Status = "INITIATION_PENDING"
	StatusInitiationPendingRedirectAuth	Status = "INITIATION_PENDING_REDIRECT_AUTH"
	StatusInitiationPendingRedirectAuthVerif	Status = "INITIATION_PENDING_REDIRECT_AUTH_VERIFICATION"
	StatusInitiationPendingRedirectHP	Status = "INITIATION_PENDING_REDIRECT_HP"
	StatusInitiationPendingRedirectPBL	Status = "INITIATION_PENDING_REDIRECT_PBL"
	StatusInitiationPendingEmbeddedAuth	Status = "INITIATION_PENDING_EMBEDDED_AUTH"
	StatusInitiationPendingEmbeddedAuthVerif	Status = "INITIATION_PENDING_EMBEDDED_AUTH_VERIFICATION"
	StatusInitiationPendingDecoupledAuth	Status = "INITIATION_PENDING_DECOUPLED_AUTH"
	StatusInitiationPendingRedemption	Status = "INITIATION_PENDING_REDEMPTION"
	StatusInitiationPendingRedemptionVerif	Status = "INITIATION_PENDING_REDEMPTION_VERIFICATION"
	StatusInitiationProcessing	Status = "INITIATION_PROCESSING"
	StatusInitiationCompleted	Status = "INITIATION_COMPLETED"
	StatusInitiationRejected	Status = "INITIATION_REJECTED"
	StatusInitiationRejectedInsufficientFunds	Status = "INITIATION_REJECTED_INSUFFICIENT_FUNDS"
	StatusInitiationFailed	Status = "INITIATION_FAILED"
	StatusInitiationDeclined	Status = "INITIATION_DECLINED"
	StatusInitiationExpired	Status = "INITIATION_EXPIRED"
	StatusInitiationNoFinalStatusAvailable	Status = "INITIATION_NO_FINAL_STATUS_AVAILABLE"
	StatusSettlementInProgress	Status = "SETTLEMENT_IN_PROGRESS"
	StatusSettlementCompleted	Status = "SETTLEMENT_COMPLETED"
	StatusSettlementIncomplete	Status = "SETTLEMENT_INCOMPLETE"
	StatusCanceled	Status = "CANCELED"
)

// IsFinal reports whether the payment has reached a terminal state.
func (s Status) IsFinal() bool {
	switch s {
	case StatusInitiationCompleted,
		StatusInitiationRejected,
		StatusInitiationRejectedInsufficientFunds,
		StatusInitiationFailed,
		StatusInitiationDeclined,
		StatusInitiationExpired,
		StatusInitiationNoFinalStatusAvailable,
		StatusSettlementCompleted,
		StatusSettlementIncomplete,
		StatusCanceled:
		return true
	default:
		return false
	}
}

// RequiresRedirect reports whether the PSU must be redirected for auth.
func (s Status) RequiresRedirect() bool {
	switch s {
	case StatusInitiationPendingRedirectAuth,
		StatusInitiationPendingRedirectAuthVerif,
		StatusInitiationPendingRedirectHP,
		StatusInitiationPendingRedirectPBL:
		return true
	default:
		return false
	}
}

// RequiresEmbeddedAuth reports whether embedded auth fields are needed.
func (s Status) RequiresEmbeddedAuth() bool {
	return s == StatusInitiationPendingEmbeddedAuth ||
		s == StatusInitiationPendingEmbeddedAuthVerif
}

// IsDecoupledAuth reports whether the bank is doing decoupled (push) auth.
func (s Status) IsDecoupledAuth() bool {
	return s == StatusInitiationPendingDecoupledAuth
}

// PaymentType enumerates payment kinds.
type PaymentType string

const (
	PaymentTypeSingleImmediate     PaymentType = "SINGLE_IMMEDIATE_PAYMENT"
	PaymentTypeVariableRecurring   PaymentType = "VARIABLE_RECURRING_PAYMENT"
)

// PaymentLinkStatus enumerates statuses for pay-by-link payments.
type PaymentLinkStatus string

const (
	PaymentLinkStatusActive   PaymentLinkStatus = "LINK_ACTIVE"
	PaymentLinkStatusExpired  PaymentLinkStatus = "LINK_EXPIRED"
	PaymentLinkStatusUsed     PaymentLinkStatus = "LINK_USED"
)

// PaymentInitiation is the full initiation payload sent to the API.
type PaymentInitiation struct {
	BankID                              string               `json:"bankId"`
	RefID                               string               `json:"refId,omitempty"`
	RemittanceInformationPrimary        string               `json:"remittanceInformationPrimary,omitempty"`
	RemittanceInformationSecondary      string               `json:"remittanceInformationSecondary,omitempty"`
	OnBehalfOfID                        string               `json:"onBehalfOfId,omitempty"`
	VRPConsentID                        string               `json:"vrpConsentId,omitempty"`
	Amount                              *common.Amount       `json:"amount,omitempty"`
	LocalInstrument                     string               `json:"localInstrument,omitempty"`
	Debtor                              *common.PartyAccount `json:"debtor,omitempty"`
	Creditor                            *common.PartyAccount `json:"creditor,omitempty"`
	ExecutionDate                       string               `json:"executionDate,omitempty"` // YYYY-MM-DD
	ConfirmFunds                        bool                 `json:"confirmFunds,omitempty"`
	ReturnRefundAccount                 bool                 `json:"returnRefundAccount,omitempty"`
	DisableFutureDatedPaymentConversion bool                 `json:"disableFutureDatedPaymentConversion,omitempty"`
	ReturnTokenizedAccount              bool                 `json:"returnTokenizedAccount,omitempty"`
	CallbackURL                         string               `json:"callbackUrl,omitempty"`
	CallbackState                       string               `json:"callbackState,omitempty"`
	ChargeBearer                        common.ChargeBearer  `json:"chargeBearer,omitempty"`
	Risk                                *common.RiskData     `json:"risk,omitempty"`
	FlowType                            common.FlowType      `json:"flowType,omitempty"`
	ExternalPsuReference                string               `json:"externalPsuReference,omitempty"`
}

// Payment is the complete payment resource as returned by the API.
type Payment struct {
	ID                            string                 `json:"id"`
	MemberID                      string                 `json:"memberId,omitempty"`
	Initiation                    *PaymentInitiation     `json:"initiation,omitempty"`
	Status                        Status                 `json:"status"`
	StatusReasonInformation       string                 `json:"statusReasonInformation,omitempty"`
	BankPaymentStatus             string                 `json:"bankPaymentStatus,omitempty"`
	BankPaymentID                 string                 `json:"bankPaymentId,omitempty"`
	BankTransactionID             string                 `json:"bankTransactionId,omitempty"`
	BankVrpID                     string                 `json:"bankVrpId,omitempty"`
	BankVrpStatus                 string                 `json:"bankVrpStatus,omitempty"`
	Authentication                *common.Authentication `json:"authentication,omitempty"`
	RefundDetails                 *common.RefundDetails  `json:"refundDetails,omitempty"`
	ConvertedToFutureDatedPayment bool                   `json:"convertedToFutureDatedPayment,omitempty"`
	PaymentLinkStatus             PaymentLinkStatus      `json:"paymentLinkStatus,omitempty"`
	ErrorInfo                     *common.ErrorInfo      `json:"errorInfo,omitempty"`
	CreatedDateTime               time.Time              `json:"createdDateTime"`
	UpdatedDateTime               time.Time              `json:"updatedDateTime"`
}

// IsFinal delegates to the Status type.
func (p Payment) IsFinal() bool { return p.Status.IsFinal() }

// RequiresRedirect delegates to the Status type.
func (p Payment) RequiresRedirect() bool { return p.Status.RequiresRedirect() }

// RequiresEmbeddedAuth delegates to the Status type.
func (p Payment) RequiresEmbeddedAuth() bool { return p.Status.RequiresEmbeddedAuth() }

// GetRedirectURL safely returns the redirect URL or empty string.
func (p Payment) GetRedirectURL() string {
	if p.Authentication != nil {
		return p.Authentication.RedirectURL
	}
	return ""
}

// Type aliases for common types, making the payments API ergonomic without
// requiring callers to import the common package separately.

// Amount is an alias for common.Amount.
type Amount = common.Amount

// PartyAccount is an alias for common.PartyAccount.
type PartyAccount = common.PartyAccount

// Address is an alias for common.Address.
type Address = common.Address

// RiskData is an alias for common.RiskData.
type RiskData = common.RiskData

// Authentication is an alias for common.Authentication.
type Authentication = common.Authentication

// EmbeddedField is an alias for common.EmbeddedField.
type EmbeddedField = common.EmbeddedField

// FlowType is an alias for common.FlowType.
type FlowType = common.FlowType

// ChargeBearer is an alias for common.ChargeBearer.
type ChargeBearer = common.ChargeBearer

// Re-export FlowType constants for ergonomic use without importing common.
const (
	FlowTypeFullHostedPages = common.FlowTypeFullHostedPages
	FlowTypeRedirect        = common.FlowTypeRedirect
	FlowTypeEmbedded        = common.FlowTypeEmbedded
	FlowTypePayByLink       = common.FlowTypePayByLink
)
