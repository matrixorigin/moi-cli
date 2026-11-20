package trace

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestBeginDoesNothingWithoutParent(t *testing.T) {
	t.Setenv(envTracesEnabled, "1")
	t.Setenv(envOTLPEndpoint, "127.0.0.1:4317")
	t.Setenv(envTraceParent, "")
	stop := BeginRuntimeChild()
	stop()
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/newmoi/api", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("traceparent", "00-0123456789abcdef0123456789abcdef-0123456789abcdef-01")
	Inject(req)
	if req.Header.Get("traceparent") != "" {
		t.Fatalf("traceparent = %q", req.Header.Get("traceparent"))
	}
}

func TestBeginFailureDoesNotStartSpan(t *testing.T) {
	t.Setenv(envTracesEnabled, "1")
	t.Setenv(envOTLPEndpoint, "127.0.0.1:4317")
	t.Setenv(envTraceParent, "00-0123456789abcdef0123456789abcdef-0123456789abcdef-01")
	previous := newExporter
	newExporter = func(context.Context) (sdktrace.SpanExporter, error) {
		return nil, errors.New("exporter unavailable")
	}
	t.Cleanup(func() { newExporter = previous })
	stop := BeginRuntimeChild()
	stop()
	if active.Load() != nil {
		t.Fatal("span stayed active after init failure")
	}
}

func TestRequestSpanInjectsParentAndOmitsRawError(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	mu.Lock()
	shutdownFn = func(context.Context) error { return errors.New("flush failed") }
	mu.Unlock()
	t.Cleanup(func() {
		mu.Lock()
		shutdownFn = nil
		mu.Unlock()
		active.Store(nil)
	})

	ctx, span := provider.Tracer("moi-cli").Start(context.Background(), spanName)
	active.Store(&holder{ctx: ctx})
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/newmoi/api", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("traceparent", "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01")
	Inject(req)
	if req.Header.Get("traceparent") == "" || req.Header.Get("traceparent") == "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01" {
		t.Fatalf("traceparent = %q", req.Header.Get("traceparent"))
	}
	MarkError()
	span.End()
	active.Store(nil)
	ended := recorder.Ended()
	if len(ended) != 1 || ended[0].Name() != spanName || ended[0].Status().Description != "error" {
		t.Fatalf("spans = %#v", ended)
	}
	if !slices.ContainsFunc(ended[0].Attributes(), func(item attribute.KeyValue) bool {
		return item.Key == "error.code" && item.Value.AsString() == "error"
	}) {
		t.Fatalf("attributes = %#v", ended[0].Attributes())
	}
}

func TestForbiddenTraceParentDoesNotStartATrace(t *testing.T) {
	t.Setenv(envTracesEnabled, "1")
	t.Setenv(envOTLPEndpoint, "127.0.0.1:4317")
	t.Setenv(envTraceParent, "ff-0123456789abcdef0123456789abcdef-0123456789abcdef-01")
	called := false
	previous := newExporter
	newExporter = func(context.Context) (sdktrace.SpanExporter, error) {
		called = true
		return nil, errors.New("should not install")
	}
	t.Cleanup(func() { newExporter = previous })
	BeginRuntimeChild()()
	if called || active.Load() != nil {
		t.Fatal("forbidden traceparent started a trace")
	}
}

type nopExporter struct{}

func (nopExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error { return nil }
func (nopExporter) Shutdown(context.Context) error                             { return nil }

func TestFlushHasADeadline(t *testing.T) {
	t.Setenv(envTracesEnabled, "1")
	t.Setenv(envOTLPEndpoint, "127.0.0.1:4317")
	t.Setenv(envTraceParent, "00-0123456789abcdef0123456789abcdef-0123456789abcdef-01")
	previous := newExporter
	newExporter = func(context.Context) (sdktrace.SpanExporter, error) {
		return nopExporter{}, nil
	}
	t.Cleanup(func() {
		newExporter = previous
		active.Store(nil)
	})
	stop := BeginRuntimeChild()
	hasDeadline := false
	mu.Lock()
	shutdownFn = func(ctx context.Context) error {
		_, hasDeadline = ctx.Deadline()
		return nil
	}
	mu.Unlock()
	stop()
	if !hasDeadline {
		t.Fatal("flush has no deadline")
	}
}

func TestShutdownFailureIsNotACommandError(t *testing.T) {
	mu.Lock()
	shutdownFn = func(context.Context) error { return errors.New("flush failed") }
	mu.Unlock()
	ctx, span := sdktrace.NewTracerProvider().Tracer("moi-cli").Start(context.Background(), spanName)
	active.Store(&holder{ctx: ctx})
	finish := func() {
		span.End()
		active.Store(nil)
		mu.Lock()
		fn := shutdownFn
		shutdownFn = nil
		mu.Unlock()
		_ = fn(context.Background())
	}
	finish()
	if active.Load() != nil {
		t.Fatal("span still active")
	}
}
