package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// OAuthConfig holds OAuth2 client-credentials configuration.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
	Scopes       []string
}

type oauthToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	expiresAt   time.Time
}

func (t *oauthToken) isExpired() bool {
	return time.Now().After(t.expiresAt.Add(-30 * time.Second))
}

// oauthProvider implements the OAuth2 client-credentials flow with caching.
type oauthProvider struct {
	cfg    OAuthConfig
	hc     *http.Client
	mu     sync.RWMutex
	cached *oauthToken
}

// NewOAuthProvider returns a Provider using OAuth2 client credentials.
// Tokens are fetched lazily and refreshed automatically before expiry.
func NewOAuthProvider(cfg OAuthConfig) Provider {
	return &oauthProvider{
		cfg: cfg,
		hc:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (o *oauthProvider) GetToken(ctx context.Context) (string, error) {
	// Fast path: cached and valid.
	o.mu.RLock()
	if o.cached != nil && !o.cached.isExpired() {
		tok := o.cached.AccessToken
		o.mu.RUnlock()
		return tok, nil
	}
	o.mu.RUnlock()

	// Slow path: fetch new token.
	o.mu.Lock()
	defer o.mu.Unlock()

	// Double-check after acquiring write lock.
	if o.cached != nil && !o.cached.isExpired() {
		return o.cached.AccessToken, nil
	}

	tok, err := o.fetch(ctx)
	if err != nil {
		return "", err
	}
	o.cached = tok
	return tok.AccessToken, nil
}

func (o *oauthProvider) fetch(ctx context.Context) (*oauthToken, error) {
	vals := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {o.cfg.ClientID},
		"client_secret": {o.cfg.ClientSecret},
	}
	if len(o.cfg.Scopes) > 0 {
		vals.Set("scope", strings.Join(o.cfg.Scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.cfg.TokenURL,
		strings.NewReader(vals.Encode()))
	if err != nil {
		return nil, fmt.Errorf("auth: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := o.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth: fetch token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth: token endpoint returned HTTP %d", resp.StatusCode)
	}

	var tok oauthToken
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, fmt.Errorf("auth: decode token response: %w", err)
	}
	if tok.AccessToken == "" {
		return nil, fmt.Errorf("auth: empty access_token in response")
	}

	ttl := time.Duration(tok.ExpiresIn) * time.Second
	if ttl == 0 {
		ttl = 3600 * time.Second
	}
	tok.expiresAt = time.Now().Add(ttl)
	return &tok, nil
}
