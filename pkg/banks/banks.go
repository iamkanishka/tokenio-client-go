// Package banks provides access to the Token.io Banks v1 and v2 APIs.
package banks

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
	"github.com/iamkanishka/tokenio-client-go/pkg/common"
)

// Bank represents a financial institution supported by Token.io.
type Bank struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	FullName    string   `json:"fullName,omitempty"`
	DisplayName string   `json:"displayName,omitempty"`
	Logo        string   `json:"logo,omitempty"`
	LogoURI     string   `json:"logoUri,omitempty"`
	IconURI     string   `json:"iconUri,omitempty"`
	Country     string   `json:"country,omitempty"` // ISO 3166-1 alpha-2
	Currencies  []string `json:"currencies,omitempty"`
	BIC         string   `json:"bic,omitempty"`
	// Provider identifies the bank connectivity provider.
	Provider string `json:"provider,omitempty"`
	// Capabilities lists supported features: "PIS", "AIS", "VRP".
	Capabilities        []string `json:"capabilities,omitempty"`
	RequiresCallbackURL bool     `json:"requiresCallbackUrl,omitempty"`
	OpenBankingStandard string   `json:"openBankingStandard,omitempty"`
	Enabled             bool     `json:"enabled,omitempty"`
}

// GetBanksV1Request holds query parameters for GET /banks (v1).
type GetBanksV1Request struct {
	IDs         []string
	Search      string
	Country     string
	Start       int
	Limit       int
	Provider    string
	Destination string
}

// GetBanksV1Response is returned by GET /banks (v1).
type GetBanksV1Response struct {
	Banks    []Bank           `json:"banks"`
	PageInfo *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetBanksV2Request holds query parameters for GET /v2/banks.
type GetBanksV2Request struct {
	IDs          []string
	Search       string
	Country      string
	Start        int
	Limit        int
	Provider     string
	Capabilities []string
	Sort         string
}

// GetBanksV2Response is returned by GET /v2/banks.
type GetBanksV2Response struct {
	Banks    []Bank           `json:"banks"`
	PageInfo *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetBankCountriesResponse is returned by GET /banks/countries.
type GetBankCountriesResponse struct {
	Countries []string `json:"countries"`
}

// Client exposes the Token.io Banks v1 and v2 APIs.
type Client struct {
	hc *httpclient.Client
}

// NewClient creates a Banks client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// GetBanksV1 retrieves a list of supported banks using the v1 API.
//
// GET /banks
func (c *Client) GetBanksV1(ctx context.Context, req GetBanksV1Request) (*GetBanksV1Response, error) {
	if req.Limit <= 0 || req.Limit > 200 {
		return nil, fmt.Errorf("banks: Limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/banks").
		WithQuery("search", req.Search).
		WithQuery("country", req.Country).
		WithQuery("provider", req.Provider).
		WithQuery("destination", req.Destination)
	if req.Start > 0 {
		r.WithQuery("start", strconv.Itoa(req.Start))
	}
	r.WithQuery("limit", strconv.Itoa(req.Limit))
	for _, id := range req.IDs {
		r.Query().Add("ids", id)
	}
	var out GetBanksV1Response
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetBankCountries retrieves all countries that have supported banks.
//
// GET /banks/countries
func (c *Client) GetBankCountries(ctx context.Context) (*GetBankCountriesResponse, error) {
	var out GetBankCountriesResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/banks/countries"), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetBanksV2 retrieves a list of supported banks using the v2 API.
//
// GET /v2/banks
func (c *Client) GetBanksV2(ctx context.Context, req GetBanksV2Request) (*GetBanksV2Response, error) {
	if req.Limit <= 0 || req.Limit > 200 {
		return nil, fmt.Errorf("banks: Limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/v2/banks").
		WithQuery("search", req.Search).
		WithQuery("country", req.Country).
		WithQuery("provider", req.Provider).
		WithQuery("sort", req.Sort)
	if req.Start > 0 {
		r.WithQuery("start", strconv.Itoa(req.Start))
	}
	r.WithQuery("limit", strconv.Itoa(req.Limit))
	for _, id := range req.IDs {
		r.Query().Add("ids", id)
	}
	for _, cap := range req.Capabilities {
		r.Query().Add("capabilities", cap)
	}
	var out GetBanksV2Response
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
