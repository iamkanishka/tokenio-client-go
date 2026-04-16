// Package settlement provides access to the Token.io Settlement Accounts API.
package settlement

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
	"github.com/iamkanishka/tokenio-client-go/pkg/common"
)

// Account represents a Token.io virtual settlement account.
type Account struct {
	ID              string    `json:"id"`
	Name            string    `json:"name,omitempty"`
	Currency        string    `json:"currency"`
	IBAN            string    `json:"iban,omitempty"`
	AccountNumber   string    `json:"accountNumber,omitempty"`
	SortCode        string    `json:"sortCode,omitempty"`
	BIC             string    `json:"bic,omitempty"`
	Balance         string    `json:"balance,omitempty"`
	CreatedDateTime time.Time `json:"createdDateTime"`
	UpdatedDateTime time.Time `json:"updatedDateTime"`
}

// Transaction represents a settlement account transaction.
type Transaction struct {
	ID              string         `json:"id"`
	AccountID       string         `json:"accountId"`
	Type            string         `json:"type,omitempty"`
	Amount          *common.Amount `json:"amount,omitempty"`
	Status          string         `json:"status,omitempty"`
	Reference       string         `json:"reference,omitempty"`
	CreatedDateTime time.Time      `json:"createdDateTime"`
}

// Rule defines an automated settlement rule for an account.
// Previously named SettlementRule; renamed to avoid revive stutter (settlement.SettlementRule).
type Rule struct {
	ID              string    `json:"id"`
	AccountID       string    `json:"accountId"`
	RuleType        string    `json:"ruleType,omitempty"`
	ThresholdAmount string    `json:"thresholdAmount,omitempty"`
	Currency        string    `json:"currency,omitempty"`
	Enabled         bool      `json:"enabled"`
	CreatedDateTime time.Time `json:"createdDateTime"`
}

// SettlementRule is an alias for Rule retained for backwards compatibility.
type SettlementRule = Rule

// CreateAccountRequest is the body for POST /virtual-accounts.
type CreateAccountRequest struct {
	Name     string `json:"name,omitempty"`
	Currency string `json:"currency"`
}

// CreateAccountResponse wraps the created account.
type CreateAccountResponse struct {
	VirtualAccount Account `json:"virtualAccount"`
}

// GetAccountsResponse wraps a list of settlement accounts.
type GetAccountsResponse struct {
	VirtualAccounts []Account        `json:"virtualAccounts"`
	PageInfo        *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetAccountResponse wraps a single settlement account.
type GetAccountResponse struct {
	VirtualAccount Account `json:"virtualAccount"`
}

// GetTransactionsResponse wraps a list of transactions.
type GetTransactionsResponse struct {
	Transactions []Transaction    `json:"transactions"`
	PageInfo     *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetTransactionResponse wraps a single transaction.
type GetTransactionResponse struct {
	Transaction Transaction `json:"transaction"`
}

// CreateRuleRequest is the body for POST /virtual-accounts/{id}/settlement-rules.
type CreateRuleRequest struct {
	AccountID       string `json:"-"`
	RuleType        string `json:"ruleType"`
	ThresholdAmount string `json:"thresholdAmount,omitempty"`
	Currency        string `json:"currency,omitempty"`
}

// CreateRuleResponse wraps the created rule.
type CreateRuleResponse struct {
	SettlementRule Rule `json:"settlementRule"`
}

// GetRulesResponse wraps a list of settlement rules.
type GetRulesResponse struct {
	SettlementRules []Rule           `json:"settlementRules"`
	PageInfo        *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetRuleResponse wraps a single settlement rule.
type GetRuleResponse struct {
	SettlementRule Rule `json:"settlementRule"`
}

// Client exposes the Token.io Settlement Accounts API.
type Client struct {
	hc *httpclient.Client
}

// NewClient creates a Settlement client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// CreateAccount creates a new virtual settlement account.
//
// POST /virtual-accounts
func (c *Client) CreateAccount(ctx context.Context, req CreateAccountRequest) (*Account, error) {
	if req.Currency == "" {
		return nil, fmt.Errorf("settlement: Currency is required")
	}
	var out CreateAccountResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, "/virtual-accounts").WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out.VirtualAccount, nil
}

// GetAccounts retrieves all virtual settlement accounts.
//
// GET /virtual-accounts
func (c *Client) GetAccounts(ctx context.Context, limit int, offset string) (*GetAccountsResponse, error) {
	if limit <= 0 || limit > 200 {
		return nil, fmt.Errorf("settlement: limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/virtual-accounts").
		WithQuery("limit", strconv.Itoa(limit)).
		WithQuery("offset", offset)
	var out GetAccountsResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAccount retrieves a single settlement account by ID.
//
// GET /virtual-accounts/{id}
func (c *Client) GetAccount(ctx context.Context, accountID string) (*Account, error) {
	if accountID == "" {
		return nil, fmt.Errorf("settlement: accountID is required")
	}
	var out GetAccountResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/virtual-accounts/"+accountID), &out); err != nil {
		return nil, err
	}
	return &out.VirtualAccount, nil
}

// GetTransactions retrieves transactions for a settlement account.
//
// GET /virtual-accounts/{id}/transactions
func (c *Client) GetTransactions(ctx context.Context, accountID string, limit int, offset string) (*GetTransactionsResponse, error) {
	if accountID == "" {
		return nil, fmt.Errorf("settlement: accountID is required")
	}
	if limit <= 0 || limit > 200 {
		return nil, fmt.Errorf("settlement: limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/virtual-accounts/"+accountID+"/transactions").
		WithQuery("limit", strconv.Itoa(limit)).
		WithQuery("offset", offset)
	var out GetTransactionsResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTransaction retrieves a single transaction from a settlement account.
//
// GET /virtual-accounts/{id}/transactions/{transactionId}
func (c *Client) GetTransaction(ctx context.Context, accountID, transactionID string) (*Transaction, error) {
	if accountID == "" {
		return nil, fmt.Errorf("settlement: accountID is required")
	}
	if transactionID == "" {
		return nil, fmt.Errorf("settlement: transactionID is required")
	}
	path := "/virtual-accounts/" + accountID + "/transactions/" + transactionID
	var out GetTransactionResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, path), &out); err != nil {
		return nil, err
	}
	return &out.Transaction, nil
}

// CreateRule creates a settlement rule for an account.
//
// POST /virtual-accounts/{id}/settlement-rules
func (c *Client) CreateRule(ctx context.Context, req CreateRuleRequest) (*SettlementRule, error) {
	if req.AccountID == "" {
		return nil, fmt.Errorf("settlement: AccountID is required")
	}
	path := "/virtual-accounts/" + req.AccountID + "/settlement-rules"
	var out CreateRuleResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, path).WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out.SettlementRule, nil
}

// GetRules retrieves all settlement rules for an account.
//
// GET /virtual-accounts/{id}/settlement-rules
func (c *Client) GetRules(ctx context.Context, accountID string, limit int, offset string) (*GetRulesResponse, error) {
	if accountID == "" {
		return nil, fmt.Errorf("settlement: accountID is required")
	}
	r := httpclient.NewRequest(http.MethodGet, "/virtual-accounts/"+accountID+"/settlement-rules").
		WithQuery("limit", strconv.Itoa(limit)).
		WithQuery("offset", offset)
	var out GetRulesResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetRule retrieves a single settlement rule by ID.
//
// GET /virtual-accounts/{id}/settlement-rules/{ruleId}
func (c *Client) GetRule(ctx context.Context, accountID, ruleID string) (*SettlementRule, error) {
	if accountID == "" {
		return nil, fmt.Errorf("settlement: accountID is required")
	}
	if ruleID == "" {
		return nil, fmt.Errorf("settlement: ruleID is required")
	}
	path := "/virtual-accounts/" + accountID + "/settlement-rules/" + ruleID
	var out GetRuleResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, path), &out); err != nil {
		return nil, err
	}
	return &out.SettlementRule, nil
}

// DeleteRule deletes a settlement rule.
//
// DELETE /virtual-accounts/{id}/settlement-rules/{ruleId}
func (c *Client) DeleteRule(ctx context.Context, accountID, ruleID string) error {
	if accountID == "" {
		return fmt.Errorf("settlement: accountID is required")
	}
	if ruleID == "" {
		return fmt.Errorf("settlement: ruleID is required")
	}
	path := "/virtual-accounts/" + accountID + "/settlement-rules/" + ruleID
	return c.hc.Do(ctx, httpclient.NewRequest(http.MethodDelete, path), nil)
}

// GetSettlementPayouts retrieves payouts associated with a settlement account.
//
// GET /virtual-accounts/{id}/settlement-payouts
func (c *Client) GetSettlementPayouts(ctx context.Context, accountID string, limit int, offset string) (*GetTransactionsResponse, error) {
	if accountID == "" {
		return nil, fmt.Errorf("settlement: accountID is required")
	}
	r := httpclient.NewRequest(http.MethodGet, "/virtual-accounts/"+accountID+"/settlement-payouts").
		WithQuery("limit", strconv.Itoa(limit)).
		WithQuery("offset", offset)
	var out GetTransactionsResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
