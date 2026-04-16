package unit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iamkanishka/tokenio-client-go/internal/auth"
)

func TestStaticProvider_AlwaysReturnsToken(t *testing.T) {
	p := auth.NewStaticProvider("static-bearer-token")
	for i := 0; i < 3; i++ {
		tok, err := p.GetToken(context.Background())
		if err != nil {
			t.Fatalf("[%d] GetToken: %v", i, err)
		}
		if tok != "static-bearer-token" {
			t.Errorf("[%d] got %q, want static-bearer-token", i, tok)
		}
	}
}

func TestOAuthProvider_FetchesAndCachesToken(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		if r.FormValue("grant_type") != "client_credentials" {
			t.Errorf("grant_type: got %q", r.FormValue("grant_type"))
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"access_token": "bearer-xyz",
			"expires_in":   3600,
			"token_type":   "Bearer",
		}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	p := auth.NewOAuthProvider(auth.OAuthConfig{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		TokenURL:     srv.URL,
	})

	tok, err := p.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	if tok != "bearer-xyz" {
		t.Errorf("token: got %q, want bearer-xyz", tok)
	}

	tok2, err := p.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken (cached): %v", err)
	}
	if tok2 != "bearer-xyz" {
		t.Errorf("cached token: got %q", tok2)
	}
	if calls != 1 {
		t.Errorf("expected exactly 1 server call, got %d", calls)
	}
}

func TestOAuthProvider_BadCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		if err := json.NewEncoder(w).Encode(map[string]any{
			"error":             "invalid_client",
			"error_description": "Client authentication failed",
		}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	p := auth.NewOAuthProvider(auth.OAuthConfig{
		ClientID:     "bad-id",
		ClientSecret: "bad-secret",
		TokenURL:     srv.URL,
	})
	_, err := p.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error for bad credentials, got nil")
	}
}

func TestOAuthProvider_EmptyAccessToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"access_token": "",
			"expires_in":   3600,
		}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	p := auth.NewOAuthProvider(auth.OAuthConfig{
		ClientID:     "id",
		ClientSecret: "secret",
		TokenURL:     srv.URL,
	})
	_, err := p.GetToken(context.Background())
	if err == nil {
		t.Fatal("expected error for empty access_token")
	}
}
