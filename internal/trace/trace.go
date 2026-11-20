// Package trace exports moi_cli.request for a sandbox runtime child.
//
// The span starts only when OTEL_TRACES_ENABLED=1, an OTLP endpoint is set,
// and TRACEPARENT is present. A missing carrier does not start a new trace.
// Export failure does not change the command result.
package trace

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	envTracesEnabled = "OTEL_TRACES_ENABLED"
	envOTLPEndpoint  = "OTEL_EXPORTER_OTLP_ENDPOINT"
	envTraceParent   = "TRACEPARENT"
	envTraceState    = "TRACESTATE"
	spanName         = "moi_cli.request"
	flushTimeout     = 5 * time.Second
)

var (
	mu          sync.Mutex
	shutdownFn  func(context.Context) error
	prop        = propagation.TraceContext{}
	active      atomic.Pointer[holder]
	newExporter = func(ctx context.Context) (sdktrace.SpanExporter, error) {
		return otlptracegrpc.New(ctx)
	}
)

type holder struct {
	ctx context.Context
}

// BeginRuntimeChild starts moi_cli.request when this process is the sandbox
// child and a parent carrier was supplied. The returned function ends the span
// and flushes. Flush failure is written to stderr and is not an error result.
func BeginRuntimeChild() func() {
	if os.Getenv(envTracesEnabled) != "1" {
		return func() {}
	}
	endpoint := os.Getenv(envOTLPEndpoint)
	if endpoint == "" || endpoint != strings.TrimSpace(endpoint) {
		return func() {}
	}
	carrier := propagation.MapCarrier{}
	if parent := strings.TrimSpace(os.Getenv(envTraceParent)); parent != "" {
		carrier["traceparent"] = parent
	}
	if state := strings.TrimSpace(os.Getenv(envTraceState)); state != "" {
		carrier["tracestate"] = state
	}
	parentCtx := prop.Extract(context.Background(), carrier)
	if !trace.SpanFromContext(parentCtx).SpanContext().IsValid() {
		return func() {}
	}
	if err := install(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "opentelemetry export disabled: %v\n", err)
		return func() {}
	}
	ctx, span := otel.Tracer("moi-cli").Start(parentCtx, spanName)
	active.Store(&holder{ctx: ctx})
	return func() {
		span.End()
		active.Store(nil)
		mu.Lock()
		fn := shutdownFn
		shutdownFn = nil
		mu.Unlock()
		if fn == nil {
			return
		}
		flushCtx, cancel := context.WithTimeout(context.Background(), flushTimeout)
		defer cancel()
		if err := fn(flushCtx); err != nil {
			fmt.Fprintf(os.Stderr, "opentelemetry flush failed: %v\n", err)
		}
	}
}

// Inject writes the active span onto an outbound product request.
// With no active span, stale W3C headers are removed.
func Inject(req *http.Request) {
	if req == nil {
		return
	}
	req.Header.Del("traceparent")
	req.Header.Del("tracestate")
	current := active.Load()
	if current == nil {
		return
	}
	prop.Inject(current.ctx, propagation.HeaderCarrier(req.Header))
}

// MarkError records a stable code on the active span. The command error text
// stays off the exported span.
func MarkError() {
	current := active.Load()
	if current == nil {
		return
	}
	span := trace.SpanFromContext(current.ctx)
	span.SetStatus(codes.Error, "error")
	span.SetAttributes(attribute.String("error.code", "error"))
}

func install(ctx context.Context) error {
	mu.Lock()
	defer mu.Unlock()
	exporter, err := newExporter(ctx)
	if err != nil {
		return err
	}
	res, err := resource.New(ctx, resource.WithAttributes(attribute.String("service.name", "moi-cli")))
	if err != nil {
		return err
	}
	provider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter), sdktrace.WithResource(res))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(prop)
	shutdownFn = provider.Shutdown
	return nil
}
