package httpclient

import (
	"net/http"
)

// Tracer is the interface for plugging in distributed tracing.
// Implement this to integrate OpenTelemetry, Datadog, or any other tracer.
//
// Example with OpenTelemetry:
//
//	type otelTracer struct{}
//	func (o *otelTracer) TraceRequest(req *http.Request) (*http.Request, func(resp *http.Response, err error)) {
//	    ctx, span := otel.Tracer("tokenio").Start(req.Context(), "http."+req.Method)
//	    return req.WithContext(ctx), func(resp *http.Response, err error) { span.End() }
//	}
type Tracer interface {
	// TraceRequest wraps an outgoing request with tracing.
	// It returns the (possibly modified) request and a finish func called after the response.
	TraceRequest(req *http.Request) (*http.Request, func(resp *http.Response, err error))
}

// tracingTransport wraps an http.RoundTripper with pluggable tracing.
type tracingTransport struct {
	base   http.RoundTripper
	tracer Tracer
}

// NewTracingTransport wraps base with the given Tracer.
// If tracer is nil, base is returned unchanged.
func NewTracingTransport(base http.RoundTripper, tracer Tracer) http.RoundTripper {
	if tracer == nil || base == nil {
		return base
	}
	return &tracingTransport{base: base, tracer: tracer}
}

func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req, finish := t.tracer.TraceRequest(req)
	resp, err := t.base.RoundTrip(req)
	finish(resp, err)
	return resp, err
}
