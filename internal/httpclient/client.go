// Package httpclient provides the internal HTTP transport for all SDK modules.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/iamkanishka/tokenio-client-go/internal/auth"
	sdkerrors "github.com/iamkanishka/tokenio-client-go/internal/errors"
)

// Config holds all configuration for the HTTP client.
type Config struct {
	BaseURL       string
	Timeout       time.Duration
	MaxRetries    int
	RetryWaitMin  time.Duration
	RetryWaitMax  time.Duration
	UserAgent     string
	Logger        *slog.Logger
	Auth          auth.Provider
	RateLimiter   *TokenBucketLimiter
	HTTPClient    *http.Client
	EnableTracing bool
}

// Client is a thread-safe HTTP client with auth injection, retry, rate
// limiting, and structured error mapping.
type Client struct {
	baseURL     string
	userAgent   string
	authProv    auth.Provider
	rateLimiter *TokenBucketLimiter
	underlying  *http.Client
	logger      *slog.Logger
	retry       *retryPolicy
}

// New constructs an HTTP Client from the given Config.
func New(cfg Config) *Client {
	underlying := cfg.HTTPClient
	if underlying == nil {
		underlying = &http.Client{
			Timeout:   cfg.Timeout,
			Transport: newTransport(cfg.EnableTracing),
		}
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	return &Client{
		baseURL:     cfg.BaseURL,
		userAgent:   cfg.UserAgent,
		authProv:    cfg.Auth,
		rateLimiter: cfg.RateLimiter,
		underlying:  underlying,
		logger:      logger,
		retry: &retryPolicy{
			maxRetries: cfg.MaxRetries,
			waitMin:    cfg.RetryWaitMin,
			waitMax:    cfg.RetryWaitMax,
		},
	}
}

// Request is a fluent builder for HTTP requests.
type Request struct {
	method  string
	path    string
	query   url.Values
	headers map[string]string
	body    any
}

// NewRequest creates a Request for the given method and path.
func NewRequest(method, path string) *Request {
	return &Request{
		method:  method,
		path:    path,
		query:   url.Values{},
		headers: map[string]string{},
	}
}

// WithBody sets the JSON-encoded request body.
func (r *Request) WithBody(v any) *Request { r.body = v; return r }

// WithQuery adds a query parameter (no-op if value is empty).
func (r *Request) WithQuery(key, value string) *Request {
	if value != "" {
		r.query.Set(key, value)
	}
	return r
}

// WithQueryBool adds a boolean query parameter when true.
func (r *Request) WithQueryBool(key string, value bool) *Request {
	if value {
		r.query.Set(key, "true")
	}
	return r
}

// WithHeader sets a custom request header.
func (r *Request) WithHeader(key, value string) *Request {
	r.headers[key] = value
	return r
}

// Query returns the underlying url.Values for direct multi-value manipulation,
// e.g.: r.Query().Add("ids", "id1"); r.Query().Add("ids", "id2")
func (r *Request) Query() url.Values { return r.query }

// Do executes the Request with retry/backoff and decodes the JSON response
// into out. Pass nil for out to discard the body.
func (c *Client) Do(ctx context.Context, req *Request, out any) error {
	if c.rateLimiter != nil {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt <= c.retry.maxRetries; attempt++ {
		if attempt > 0 {
			wait := c.retry.backoff(attempt)
			c.logger.Warn("retrying request",
				"method", req.method,
				"path", req.path,
				"attempt", attempt,
				"wait", wait,
			)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		resp, err := c.execute(ctx, req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out == nil {
				resp.Body.Close()
				return nil
			}
			decErr := json.NewDecoder(resp.Body).Decode(out)
			resp.Body.Close()
			if decErr != nil && !errors.Is(decErr, io.EOF) {
				return fmt.Errorf("decode response: %w", decErr)
			}
			return nil
		}

		// Non-2xx: FromResponse closes the body.
		apiErr := sdkerrors.FromResponse(resp)
		if !isRetryable(resp.StatusCode) {
			return apiErr
		}
		lastErr = apiErr
	}

	return fmt.Errorf("tokenio: max retries exceeded: %w", lastErr)
}

// DoRaw executes the Request and returns raw bytes + HTTP status.
func (c *Client) DoRaw(ctx context.Context, req *Request) ([]byte, int, error) {
	resp, err := c.execute(ctx, req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	return b, resp.StatusCode, err
}

func (c *Client) execute(ctx context.Context, req *Request) (*http.Response, error) {
	endpoint := c.baseURL + req.path
	if len(req.query) > 0 {
		endpoint += "?" + req.query.Encode()
	}

	var bodyReader io.Reader
	if req.body != nil {
		b, err := json.Marshal(req.body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.method, endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")
	if req.body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if c.userAgent != "" {
		httpReq.Header.Set("User-Agent", c.userAgent)
	}
	for k, v := range req.headers {
		httpReq.Header.Set(k, v)
	}

	if c.authProv != nil {
		token, err := c.authProv.GetToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("auth token: %w", err)
		}
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	start := time.Now()
	resp, err := c.underlying.Do(httpReq)
	elapsed := time.Since(start)

	if err != nil {
		c.logger.Error("request failed",
			"method", req.method,
			"path", req.path,
			"elapsed_ms", elapsed.Milliseconds(),
			"error", err.Error(),
		)
		return nil, err
	}

	c.logger.Info("request completed",
		"method", req.method,
		"path", req.path,
		"status", resp.StatusCode,
		"elapsed_ms", elapsed.Milliseconds(),
	)
	return resp, nil
}

// Close releases idle HTTP connections.
func (c *Client) Close() {
	if t, ok := c.underlying.Transport.(*http.Transport); ok {
		t.CloseIdleConnections()
	}
}

func isRetryable(status int) bool {
	switch status {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}

// ── Token bucket rate limiter (stdlib-only replacement for golang.org/x/time/rate) ──

// TokenBucketLimiter is a simple token-bucket rate limiter using only stdlib.
type TokenBucketLimiter struct {
	mu       sync.Mutex
	tokens   float64
	maxBurst float64
	ratePerS float64
	last     time.Time
}

// NewTokenBucketLimiter creates a limiter allowing rps requests/sec with
// a burst allowance of burst.
func NewTokenBucketLimiter(rps float64, burst int) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		tokens:   float64(burst),
		maxBurst: float64(burst),
		ratePerS: rps,
		last:     time.Now(),
	}
}

// Wait blocks until a token is available or ctx is cancelled.
func (l *TokenBucketLimiter) Wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(l.last).Seconds()
		l.tokens += elapsed * l.ratePerS
		if l.tokens > l.maxBurst {
			l.tokens = l.maxBurst
		}
		l.last = now

		if l.tokens >= 1 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		// Calculate wait time for next token.
		wait := time.Duration((1-l.tokens)/l.ratePerS*1000) * time.Millisecond
		l.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}
