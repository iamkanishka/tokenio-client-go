// Package refunds provides access to the Token.io Refunds API.
package refunds

import (
	"time"

	"github.com/iamkanishka/tokenio-client-go/pkg/common"
)

// RefundStatus enumerates all refund lifecycle states.
type RefundStatus string

const (
	RefundStatusInitiationPending    RefundStatus = "INITIATION_PENDING"
	RefundStatusInitiationProcessing RefundStatus = "INITIATION_PROCESSING"
	RefundStatusInitiationCompleted  RefundStatus = "INITIATION_COMPLETED"
	RefundStatusInitiationRejected   RefundStatus = "INITIATION_REJECTED"
	RefundStatusInitiationFailed     RefundStatus = "INITIATION_FAILED"
)

// IsFinal reports whether the refund has reached a terminal state.
func (s RefundStatus) IsFinal() bool {
	switch s {
	case RefundStatusInitiationCompleted,
		RefundStatusInitiationRejected,
		RefundStatusInitiationFailed:
		return true

	case RefundStatusInitiationPending,
		RefundStatusInitiationProcessing:
		return false
	}

	panic("unhandled refunds.RefundStatus: " + string(s))
}

// Refund is the complete refund resource returned by the API.
type Refund struct {
	ID                      string            `json:"id"`
	MemberID                string            `json:"memberId,omitempty"`
	TransferID              string            `json:"transferId,omitempty"`
	RefundInitiation        *RefundInitiation `json:"refundInitiation,omitempty"`
	Status                  RefundStatus      `json:"status"`
	StatusReasonInformation string            `json:"statusReasonInformation,omitempty"`
	ErrorInfo               *common.ErrorInfo `json:"errorInfo,omitempty"`
	CreatedDateTime         time.Time         `json:"createdDateTime"`
	UpdatedDateTime         time.Time         `json:"updatedDateTime"`
}

// IsFinal delegates to RefundStatus.
func (r Refund) IsFinal() bool { return r.Status.IsFinal() }

// RefundInitiation holds the initiation details for a refund.
type RefundInitiation struct {
	TransferID                     string               `json:"transferId"`
	RefID                          string               `json:"refId,omitempty"`
	Amount                         *common.Amount       `json:"amount"`
	RemittanceInformationPrimary   string               `json:"remittanceInformationPrimary,omitempty"`
	RemittanceInformationSecondary string               `json:"remittanceInformationSecondary,omitempty"`
	Debtor                         *common.PartyAccount `json:"debtor,omitempty"`
	Creditor                       *common.PartyAccount `json:"creditor,omitempty"`
}
