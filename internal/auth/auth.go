// Package auth provides authentication providers for the Token.io SDK.
package auth

import "context"

// Provider is the pluggable authentication interface.
// All implementations must be safe for concurrent use.
type Provider interface {
	// GetToken returns a valid bearer token, refreshing it automatically if needed.
	GetToken(ctx context.Context) (string, error)
}

// staticProvider always returns the same pre-issued token.
type staticProvider struct{ token string }

// NewStaticProvider returns a Provider that always returns the given token.
// Use for development or when managing token lifecycle externally.
func NewStaticProvider(token string) Provider {
	return &staticProvider{token: token}
}

func (s *staticProvider) GetToken(_ context.Context) (string, error) {
	return s.token, nil
}
