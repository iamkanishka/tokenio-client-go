// Package common contains domain types shared across Token.io API packages.
package common

// Amount represents a monetary value with ISO 4217 currency.
type Amount struct {
	Value    string `json:"value"`    // decimal string, e.g. "10.23"
	Currency string `json:"currency"` // ISO 4217, e.g. "GBP"
}

// Address holds a structured postal address.
type Address struct {
	AddressLine    []string `json:"addressLine,omitempty"`
	StreetName     string   `json:"streetName,omitempty"`
	BuildingNumber string   `json:"buildingNumber,omitempty"`
	PostCode       string   `json:"postCode,omitempty"`
	TownName       string   `json:"townName,omitempty"`
	State          string   `json:"state,omitempty"`
	District       string   `json:"district,omitempty"`
	Country        string   `json:"country,omitempty"` // ISO 3166-1 alpha-2
}

// DeliveryAddress extends Address with additional fields for risk data.
type DeliveryAddress struct {
	AddressLine    []string `json:"addressLine,omitempty"`
	AddressType    string   `json:"addressType,omitempty"`
	BuildingNumber string   `json:"buildingNumber,omitempty"`
	Country        string   `json:"country,omitempty"`
	CountrySubDiv  []string `json:"countrySubDivision,omitempty"`
	Department     string   `json:"department,omitempty"`
	PostCode       string   `json:"postCode,omitempty"`
	StreetName     string   `json:"streetName,omitempty"`
	SubDepartment  string   `json:"subDepartment,omitempty"`
	TownName       string   `json:"townName,omitempty"`
}

// PartyAccount holds bank account details for a debtor or creditor.
// Fields are mutually exclusive depending on the bank scheme.
type PartyAccount struct {
	IBAN                  string   `json:"iban,omitempty"`
	BIC                   string   `json:"bic,omitempty"`
	AccountNumber         string   `json:"accountNumber,omitempty"`
	SortCode              string   `json:"sortCode,omitempty"`
	Name                  string   `json:"name,omitempty"`
	UltimateDebtorName    string   `json:"ultimateDebtorName,omitempty"`
	UltimateCreditorName  string   `json:"ultimateCreditorName,omitempty"`
	BankName              string   `json:"bankName,omitempty"`
	AccountVerificationID string   `json:"accountVerificationId,omitempty"`
	Address               *Address `json:"address,omitempty"`
}

// RiskData carries PSU risk and context information for payment initiation.
type RiskData struct {
	PsuID                            string           `json:"psuId,omitempty"`
	PaymentContextCode               string           `json:"paymentContextCode,omitempty"`
	PaymentPurposeCode               string           `json:"paymentPurposeCode,omitempty"`
	MerchantCategoryCode             string           `json:"merchantCategoryCode,omitempty"`
	BeneficiaryAccountType           string           `json:"beneficiaryAccountType,omitempty"`
	ContractPresentIndicator         bool             `json:"contractPresentIndicator,omitempty"`
	BeneficiaryPrepopulatedIndicator bool             `json:"beneficiaryPrepopulatedIndicator,omitempty"`
	DeliveryAddress                  *DeliveryAddress `json:"deliveryAddress,omitempty"`
}

// PageInfo holds cursor-based pagination details.
type PageInfo struct {
	Limit      int    `json:"limit"`
	Offset     string `json:"offset,omitempty"`
	NextOffset string `json:"nextOffset,omitempty"`
	HaveMore   bool   `json:"haveMore"`
}

// ErrorInfo holds extra error details returned by the API.
type ErrorInfo struct {
	HTTPErrorCode      int    `json:"httpErrorCode,omitempty"`
	Message            string `json:"message,omitempty"`
	TokenExternalError bool   `json:"tokenExternalError,omitempty"`
	TokenTraceID       string `json:"tokenTraceId,omitempty"`
}

// Authentication holds the redirect URL or embedded auth fields returned
// after payment or consent initiation.
type Authentication struct {
	RedirectURL  string          `json:"redirectUrl,omitempty"`
	EmbeddedAuth []EmbeddedField `json:"embeddedAuth,omitempty"`
}

// EmbeddedField describes a single field required for embedded authentication.
type EmbeddedField struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	DisplayName string `json:"displayName"`
	Mandatory   bool   `json:"mandatory"`
}

// RefundAccount holds account details for refund routing.
type RefundAccount struct {
	IBAN string `json:"iban,omitempty"`
	BIC  string `json:"bic,omitempty"`
	Name string `json:"name,omitempty"`
}

// RefundDetails describes the refund status of a payment.
type RefundDetails struct {
	RefundAccount         *RefundAccount `json:"refundAccount,omitempty"`
	PaymentRefundStatus   string         `json:"paymentRefundStatus,omitempty"`
	SettledRefundAmount   *Amount        `json:"settledRefundAmount,omitempty"`
	RemainingRefundAmount *Amount        `json:"remainingRefundAmount,omitempty"`
}

// FlowType enumerates the supported payment authorization UI flows.
type FlowType string

const (
	FlowTypeFullHostedPages FlowType = "FULL_HOSTED_PAGES"
	FlowTypeRedirect        FlowType = "REDIRECT"
	FlowTypeEmbedded        FlowType = "EMBEDDED"
	FlowTypePayByLink       FlowType = "PAY_BY_LINK"
)

// ChargeBearer enumerates who bears the transaction charges.
type ChargeBearer string

const (
	ChargeBearerCred ChargeBearer = "CRED"
	ChargeBearerDebt ChargeBearer = "DEBT"
	ChargeBearerShar ChargeBearer = "SHAR"
	ChargeBearerSlev ChargeBearer = "SLEV"
)
