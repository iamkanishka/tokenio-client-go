// Package tokenio provides a production-grade Go client SDK for the Token.io
// Open Banking platform. It covers all APIs documented at reference.token.io:
//
//   - Payments v2 (single immediate, future-dated, redirect/embedded/decoupled auth)
//   - Variable Recurring Payments (VRP — consents, payments, fund confirmation)
//   - Account on File (tokenized account storage)
//   - Token Requests — Payments v1 & AIS legacy flow
//   - Transfers — v1 token redemption
//   - Refunds
//   - Payouts
//   - Settlement Accounts (virtual accounts, rules, transactions)
//   - Accounts / AIS (balances, standing orders, transactions)
//   - Tokens
//   - Banks v1 + v2
//   - Sub-TPPs
//   - Authentication Keys
//   - Reports (bank status)
//   - Webhooks (config management + typed event parsing)
//   - Verification (account ownership checks)
//
// # Quick Start
//
//	client, err := tokenio.NewClient(tokenio.Config{
//	    ClientID:     "your-client-id",
//	    ClientSecret: "your-client-secret",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	payment, err := client.Payments.InitiatePayment(ctx, req)
package tokenio

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/iamkanishka/tokenio-client-go/internal/auth"
	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
	"github.com/iamkanishka/tokenio-client-go/pkg/accountonfile"
	"github.com/iamkanishka/tokenio-client-go/pkg/ais"
	"github.com/iamkanishka/tokenio-client-go/pkg/authkeys"
	"github.com/iamkanishka/tokenio-client-go/pkg/banks"
	"github.com/iamkanishka/tokenio-client-go/pkg/payments"
	"github.com/iamkanishka/tokenio-client-go/pkg/payouts"
	"github.com/iamkanishka/tokenio-client-go/pkg/refunds"
	"github.com/iamkanishka/tokenio-client-go/pkg/reports"
	"github.com/iamkanishka/tokenio-client-go/pkg/settlement"
	"github.com/iamkanishka/tokenio-client-go/pkg/subtpps"
	"github.com/iamkanishka/tokenio-client-go/pkg/tokenrequests"
	"github.com/iamkanishka/tokenio-client-go/pkg/tokens"
	"github.com/iamkanishka/tokenio-client-go/pkg/transfers"
	"github.com/iamkanishka/tokenio-client-go/pkg/verification"
	"github.com/iamkanishka/tokenio-client-go/pkg/vrp"
	"github.com/iamkanishka/tokenio-client-go/pkg/webhooks"
)

const (
	sandboxBaseURL    = "https://api.sandbox.token.io"
	productionBaseURL = "https://api.token.io"
	sdkVersion        = "2.0.0"
	sdkUserAgent      = "tokenio-go-sdk/" + sdkVersion

	defaultTimeout        = 30 * time.Second
	defaultMaxRetries     = 3
	defaultRateLimit      = 100.0
	defaultRateLimitBurst = 20
)

// Environment selects the Token.io deployment target.
type Environment string

const (
	// EnvironmentSandbox targets https://api.sandbox.token.io (default).
	EnvironmentSandbox Environment = "sandbox"
	// EnvironmentProduction targets https://api.token.io.
	EnvironmentProduction Environment = "production"
)

// Config holds all configuration for the Token.io SDK client.
type Config struct {
	// BaseURL overrides the API base URL. Takes precedence over Environment.
	BaseURL string

	// Environment selects sandbox or production. Defaults to sandbox.
	Environment Environment

	// ClientID is the OAuth2 client ID issued by Token.io.
	ClientID string

	// ClientSecret is the OAuth2 client secret issued by Token.io.
	ClientSecret string

	// StaticToken bypasses OAuth2 with a pre-obtained bearer token.
	// Useful for testing; not recommended for production.
	StaticToken string

	// Timeout is the per-request HTTP timeout. Defaults to 30s.
	Timeout time.Duration

	// MaxRetries is the maximum retry attempts on transient errors. Defaults to 3.
	MaxRetries int

	// RetryWaitMin is the minimum backoff between retries. Defaults to 500ms.
	RetryWaitMin time.Duration

	// RetryWaitMax is the maximum backoff between retries. Defaults to 5s.
	RetryWaitMax time.Duration

	// RateLimit is the sustained request rate cap (req/sec). Defaults to 100.
	RateLimit float64

	// RateLimitBurst is the token-bucket burst allowance. Defaults to 20.
	RateLimitBurst int

	// Logger is an optional slog.Logger. Uses slog.Default() if nil.
	Logger *slog.Logger

	// HTTPClient allows injecting a custom *http.Client (e.g. for testing).
	HTTPClient *http.Client

	// WebhookSecret is the shared secret for verifying incoming webhook payloads.
	WebhookSecret string

	// EnableTracing is reserved for future pluggable tracing support.
	// Wire tracing via a custom HTTPClient with a tracing RoundTripper instead.
	EnableTracing bool
}

// Client is the root Token.io SDK client. All API sub-clients are accessible
// as exported fields. Client is safe for concurrent use by multiple goroutines.
type Client struct {
	// ── Payment Initiation ────────────────────────────────────────────────────
	// Payments exposes the Payments v2 API.
	Payments *payments.Client
	// VRP exposes the Variable Recurring Payments API.
	VRP *vrp.Client

	// ── Account Management ────────────────────────────────────────────────────
	// AccountOnFile exposes the Account on File (tokenized account storage) API.
	AccountOnFile *accountonfile.Client

	// ── Legacy v1 Flows ───────────────────────────────────────────────────────
	// TokenRequests exposes the Token Requests API (Payments v1 / AIS).
	TokenRequests *tokenrequests.Client
	// Transfers exposes the Transfers API (v1 token redemption).
	Transfers *transfers.Client
	// Tokens exposes the Tokens management API.
	Tokens *tokens.Client

	// ── Money Movement ────────────────────────────────────────────────────────
	// Refunds exposes the Refunds API.
	Refunds *refunds.Client
	// Payouts exposes the Payouts API.
	Payouts *payouts.Client
	// Settlement exposes the Settlement Accounts API.
	Settlement *settlement.Client

	// ── Account Information ───────────────────────────────────────────────────
	// AIS exposes the Account Information Services API.
	AIS *ais.Client

	// ── Infrastructure ────────────────────────────────────────────────────────
	// Banks exposes the Banks v1 and v2 APIs.
	Banks *banks.Client
	// SubTPPs exposes the Sub-TPPs management API.
	SubTPPs *subtpps.Client
	// AuthKeys exposes the Authentication Keys API.
	AuthKeys *authkeys.Client
	// Reports exposes the bank status Reports API.
	Reports *reports.Client
	// Webhooks exposes webhook config management and typed event parsing.
	Webhooks *webhooks.Client
	// Verification exposes the Account Verification API.
	Verification *verification.Client

	httpClient *httpclient.Client
}

// Option is a functional option for fine-grained Client configuration.
type Option func(*Config)

// WithBaseURL overrides the API base URL.
func WithBaseURL(u string) Option { return func(c *Config) { c.BaseURL = u } }

// WithEnvironment sets the deployment environment (sandbox or production).
func WithEnvironment(e Environment) Option { return func(c *Config) { c.Environment = e } }

// WithTimeout sets a custom per-request HTTP timeout.
func WithTimeout(d time.Duration) Option { return func(c *Config) { c.Timeout = d } }

// WithMaxRetries sets the maximum retry attempts on transient errors.
func WithMaxRetries(n int) Option { return func(c *Config) { c.MaxRetries = n } }

// WithRetryWait sets the minimum and maximum retry backoff durations.
func WithRetryWait(min, max time.Duration) Option {
	return func(c *Config) { c.RetryWaitMin = min; c.RetryWaitMax = max }
}

// WithLogger injects a custom slog.Logger.
func WithLogger(l *slog.Logger) Option { return func(c *Config) { c.Logger = l } }

// WithHTTPClient injects a custom *http.Client (e.g. with a tracing RoundTripper).
func WithHTTPClient(hc *http.Client) Option { return func(c *Config) { c.HTTPClient = hc } }

// WithTracing is a no-op option retained for API compatibility.
// To add tracing, inject a custom HTTPClient with a tracing RoundTripper:
//
//	client, err := tokenio.NewClient(cfg, tokenio.WithHTTPClient(&http.Client{
//	    Transport: myTracingRoundTripper,
//	}))
func WithTracing() Option { return func(c *Config) { c.EnableTracing = true } }

// WithRateLimit sets the sustained request rate cap and burst allowance.
func WithRateLimit(rps float64, burst int) Option {
	return func(c *Config) { c.RateLimit = rps; c.RateLimitBurst = burst }
}

// WithWebhookSecret sets the shared secret for webhook payload verification.
func WithWebhookSecret(secret string) Option {
	return func(c *Config) { c.WebhookSecret = secret }
}

// NewClient constructs a fully-initialised Token.io SDK Client.
// Returns an error if the configuration is invalid (e.g. missing credentials).
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	for _, o := range opts {
		o(&cfg)
	}
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}
	applyDefaults(&cfg)

	authProv := buildAuthProvider(cfg)

	var limiter *httpclient.TokenBucketLimiter
	if cfg.RateLimit > 0 {
		limiter = httpclient.NewTokenBucketLimiter(cfg.RateLimit, cfg.RateLimitBurst)
	}

	hc := httpclient.New(httpclient.Config{
		BaseURL:       cfg.BaseURL,
		Timeout:       cfg.Timeout,
		MaxRetries:    cfg.MaxRetries,
		RetryWaitMin:  cfg.RetryWaitMin,
		RetryWaitMax:  cfg.RetryWaitMax,
		UserAgent:     sdkUserAgent,
		Logger:        cfg.Logger,
		Auth:          authProv,
		RateLimiter:   limiter,
		HTTPClient:    cfg.HTTPClient,
		EnableTracing: cfg.EnableTracing,
	})

	return &Client{
		Payments:      payments.NewClient(hc),
		VRP:           vrp.NewClient(hc),
		AccountOnFile: accountonfile.NewClient(hc),
		TokenRequests: tokenrequests.NewClient(hc),
		Transfers:     transfers.NewClient(hc),
		Tokens:        tokens.NewClient(hc),
		Refunds:       refunds.NewClient(hc),
		Payouts:       payouts.NewClient(hc),
		Settlement:    settlement.NewClient(hc),
		AIS:           ais.NewClient(hc),
		Banks:         banks.NewClient(hc),
		SubTPPs:       subtpps.NewClient(hc),
		AuthKeys:      authkeys.NewClient(hc),
		Reports:       reports.NewClient(hc),
		Webhooks:      webhooks.NewClient(hc, cfg.WebhookSecret),
		Verification:  verification.NewClient(hc),
		httpClient:    hc,
	}, nil
}

// Ping verifies connectivity to the Token.io API with a minimal request.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.Banks.GetBanksV2(ctx, banks.GetBanksV2Request{Limit: 1})
	return err
}

// Version returns the SDK version string.
func (c *Client) Version() string { return sdkVersion }

// Close releases idle HTTP connections held by the underlying transport.
func (c *Client) Close() { c.httpClient.Close() }

func validateConfig(cfg *Config) error {
	if cfg.StaticToken == "" && (cfg.ClientID == "" || cfg.ClientSecret == "") {
		return errors.New("tokenio: provide StaticToken or both ClientID and ClientSecret")
	}
	return nil
}

func applyDefaults(cfg *Config) {
	if cfg.BaseURL == "" {
		if cfg.Environment == EnvironmentProduction {
			cfg.BaseURL = productionBaseURL
		} else {
			cfg.BaseURL = sandboxBaseURL
		}
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = defaultMaxRetries
	}
	if cfg.RetryWaitMin == 0 {
		cfg.RetryWaitMin = 500 * time.Millisecond
	}
	if cfg.RetryWaitMax == 0 {
		cfg.RetryWaitMax = 5 * time.Second
	}
	if cfg.RateLimit == 0 {
		cfg.RateLimit = defaultRateLimit
	}
	if cfg.RateLimitBurst == 0 {
		cfg.RateLimitBurst = defaultRateLimitBurst
	}
}

func buildAuthProvider(cfg Config) auth.Provider {
	if cfg.StaticToken != "" {
		return auth.NewStaticProvider(cfg.StaticToken)
	}
	return auth.NewOAuthProvider(auth.OAuthConfig{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.BaseURL + "/oauth2/token",
	})
}
