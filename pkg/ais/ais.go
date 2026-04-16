// Package ais provides access to the Token.io Account Information Services API.
package ais

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
	"github.com/iamkanishka/tokenio-client-go/pkg/common"
)

// Account represents a bank account returned from AIS.
type Account struct {
	ID              string    `json:"id"`
	DisplayName     string    `json:"displayName,omitempty"`
	Type            string    `json:"type,omitempty"`
	Currency        string    `json:"currency,omitempty"`
	BankID          string    `json:"bankId,omitempty"`
	IBAN            string    `json:"iban,omitempty"`
	AccountNumber   string    `json:"accountNumber,omitempty"`
	SortCode        string    `json:"sortCode,omitempty"`
	BIC             string    `json:"bic,omitempty"`
	Status          string    `json:"status,omitempty"`
	CreatedDateTime time.Time `json:"createdDateTime,omitempty"`
	UpdatedDateTime time.Time `json:"updatedDateTime,omitempty"`
}

// Balance holds balance details for an account.
type Balance struct {
	AccountID       string         `json:"accountId"`
	Current         *common.Amount `json:"current,omitempty"`
	Available       *common.Amount `json:"available,omitempty"`
	CreditLimit     *common.Amount `json:"creditLimit,omitempty"`
	UpdatedDateTime time.Time      `json:"updatedDateTime,omitempty"`
}

// StandingOrder represents a recurring payment instruction.
type StandingOrder struct {
	ID              string         `json:"id"`
	AccountID       string         `json:"accountId"`
	Amount          *common.Amount `json:"amount,omitempty"`
	Frequency       string         `json:"frequency,omitempty"`
	NextPaymentDate string         `json:"nextPaymentDate,omitempty"`
	Status          string         `json:"status,omitempty"`
}

// Transaction represents an account transaction.
type Transaction struct {
	ID              string         `json:"id"`
	AccountID       string         `json:"accountId"`
	Amount          *common.Amount `json:"amount,omitempty"`
	Type            string         `json:"type,omitempty"`
	Status          string         `json:"status,omitempty"`
	Description     string         `json:"description,omitempty"`
	MerchantName    string         `json:"merchantName,omitempty"`
	BookingDateTime time.Time      `json:"bookingDateTime,omitempty"`
	ValueDateTime   time.Time      `json:"valueDateTime,omitempty"`
}

// GetAccountsResponse is returned by GET /accounts.
type GetAccountsResponse struct {
	Accounts []Account        `json:"accounts"`
	PageInfo *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetAccountResponse is returned by GET /accounts/{accountId}.
type GetAccountResponse struct {
	Account Account `json:"account"`
}

// GetBalancesResponse is returned by GET /accounts/balances.
type GetBalancesResponse struct {
	Balances []Balance        `json:"balances"`
	PageInfo *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetBalanceResponse is returned by GET /accounts/{accountId}/balance.
type GetBalanceResponse struct {
	Balance Balance `json:"balance"`
}

// GetStandingOrdersResponse is returned by GET /accounts/standing-orders.
type GetStandingOrdersResponse struct {
	StandingOrders []StandingOrder  `json:"standingOrders"`
	PageInfo       *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetStandingOrderResponse is returned by GET /accounts/{id}/standing-orders/{soId}.
type GetStandingOrderResponse struct {
	StandingOrder StandingOrder `json:"standingOrder"`
}

// GetTransactionsResponse is returned by GET /accounts/{accountId}/transactions.
type GetTransactionsResponse struct {
	Transactions []Transaction    `json:"transactions"`
	PageInfo     *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetTransactionResponse is returned by GET /accounts/{id}/transactions/{txId}.
type GetTransactionResponse struct {
	Transaction Transaction `json:"transaction"`
}

// Client exposes the Token.io Account Information Services API.
type Client struct {
	hc *httpclient.Client
}

// NewClient creates an AIS client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// GetAccounts retrieves all accounts accessible under the current AIS consent.
//
// GET /accounts
func (c *Client) GetAccounts(ctx context.Context, limit int, offset string) (*GetAccountsResponse, error) {
	if limit <= 0 || limit > 200 {
		return nil, fmt.Errorf("ais: limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/accounts").
		WithQuery("limit", strconv.Itoa(limit)).
		WithQuery("offset", offset)
	var out GetAccountsResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAccount retrieves a single account by ID.
//
// GET /accounts/{accountId}
func (c *Client) GetAccount(ctx context.Context, accountID string) (*Account, error) {
	if accountID == "" {
		return nil, fmt.Errorf("ais: accountID is required")
	}
	var out GetAccountResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/accounts/"+accountID), &out); err != nil {
		return nil, err
	}
	return &out.Account, nil
}

// GetBalances retrieves balances for all accessible accounts.
//
// GET /accounts/balances
func (c *Client) GetBalances(ctx context.Context, limit int, offset string) (*GetBalancesResponse, error) {
	if limit <= 0 || limit > 200 {
		return nil, fmt.Errorf("ais: limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/accounts/balances").
		WithQuery("limit", strconv.Itoa(limit)).
		WithQuery("offset", offset)
	var out GetBalancesResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetBalance retrieves the balance for a single account.
//
// GET /accounts/{accountId}/balance
func (c *Client) GetBalance(ctx context.Context, accountID string) (*Balance, error) {
	if accountID == "" {
		return nil, fmt.Errorf("ais: accountID is required")
	}
	var out GetBalanceResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/accounts/"+accountID+"/balance"), &out); err != nil {
		return nil, err
	}
	return &out.Balance, nil
}

// GetStandingOrders retrieves all standing orders across accessible accounts.
//
// GET /accounts/standing-orders
func (c *Client) GetStandingOrders(ctx context.Context, limit int, offset string) (*GetStandingOrdersResponse, error) {
	if limit <= 0 || limit > 200 {
		return nil, fmt.Errorf("ais: limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/accounts/standing-orders").
		WithQuery("limit", strconv.Itoa(limit)).
		WithQuery("offset", offset)
	var out GetStandingOrdersResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetStandingOrder retrieves a single standing order.
//
// GET /accounts/{accountId}/standing-orders/{standingOrderId}
func (c *Client) GetStandingOrder(ctx context.Context, accountID, soID string) (*StandingOrder, error) {
	if accountID == "" {
		return nil, fmt.Errorf("ais: accountID is required")
	}
	if soID == "" {
		return nil, fmt.Errorf("ais: standingOrderID is required")
	}
	path := "/accounts/" + accountID + "/standing-orders/" + soID
	var out GetStandingOrderResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, path), &out); err != nil {
		return nil, err
	}
	return &out.StandingOrder, nil
}

// GetTransactions retrieves transactions for an account.
//
// GET /accounts/{accountId}/transactions
func (c *Client) GetTransactions(ctx context.Context, accountID string, limit int, offset string) (*GetTransactionsResponse, error) {
	if accountID == "" {
		return nil, fmt.Errorf("ais: accountID is required")
	}
	if limit <= 0 || limit > 200 {
		return nil, fmt.Errorf("ais: limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/accounts/"+accountID+"/transactions").
		WithQuery("limit", strconv.Itoa(limit)).
		WithQuery("offset", offset)
	var out GetTransactionsResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTransaction retrieves a single transaction.
//
// GET /accounts/{accountId}/transactions/{transactionId}
func (c *Client) GetTransaction(ctx context.Context, accountID, txID string) (*Transaction, error) {
	if accountID == "" {
		return nil, fmt.Errorf("ais: accountID is required")
	}
	if txID == "" {
		return nil, fmt.Errorf("ais: transactionID is required")
	}
	path := "/accounts/" + accountID + "/transactions/" + txID
	var out GetTransactionResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, path), &out); err != nil {
		return nil, err
	}
	return &out.Transaction, nil
}
