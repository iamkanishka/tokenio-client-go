package httpclient

import (
	"net"
	"net/http"
	"time"
)

// newTransport builds a production-grade http.RoundTripper with connection
// pooling, keep-alive, HTTP/2 support, and optional pluggable tracing.
// The tracing parameter is preserved for API compatibility but tracing is
// wired via NewTracingTransport at a higher level.
func newTransport(tracing bool) http.RoundTripper {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	base := &http.Transport{
		DialContext:           dialer.DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		MaxConnsPerHost:       50,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}
	// Tracing is opt-in via WithTracing() + a custom Tracer implementation.
	// See Tracer interface in middleware.go.
	_ = tracing
	return base
}
