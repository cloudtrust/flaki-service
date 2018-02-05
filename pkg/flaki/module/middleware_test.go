package flakim

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/go-kit/kit/metrics"
	opentracing "github.com/opentracing/opentracing-go"
	opentracing_log "github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	var mockLogger = &mockLogger{}
	var mockFlaki = &mockFlaki{}

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", id)

	var srv = New(mockFlaki)
	srv = MakeLoggingMiddleware(mockLogger)(srv)

	// NextID.
	mockLogger.Called = false
	mockLogger.CorrelationID = ""
	srv.NextID(ctx)
	assert.True(t, mockLogger.Called)
	assert.Equal(t, id, mockLogger.CorrelationID)

	// NextValidID.
	mockLogger.Called = false
	mockLogger.CorrelationID = ""
	srv.NextValidID(ctx)
	assert.True(t, mockLogger.Called)
	assert.Equal(t, id, mockLogger.CorrelationID)

	// NextID without correlation ID.
	var f = func() {
		srv.NextID(context.Background())
	}
	assert.Panics(t, f)

	// NextValidID without correlation ID.
	f = func() {
		srv.NextValidID(context.Background())
	}
	assert.Panics(t, f)
}

func TestMetricMiddleware(t *testing.T) {
	var mockCounter = &mockCounter{}
	var mockFlaki = &mockFlaki{}

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", id)

	var srv = New(mockFlaki)
	srv = MakeMetricMiddleware(mockCounter)(srv)

	// NextID.
	mockCounter.Called = false
	mockCounter.CorrelationID = ""
	srv.NextID(ctx)
	assert.True(t, mockCounter.Called)
	assert.Equal(t, id, mockCounter.CorrelationID)

	// NextValidID.
	mockCounter.Called = false
	mockCounter.CorrelationID = ""
	srv.NextValidID(ctx)
	assert.True(t, mockCounter.Called)
	assert.Equal(t, id, mockCounter.CorrelationID)

	// NextID without correlation ID.
	var f = func() {
		srv.NextID(context.Background())
	}
	assert.Panics(t, f)

	// NextValidID without correlation ID.
	f = func() {
		srv.NextValidID(context.Background())
	}
	assert.Panics(t, f)
}

func TestTracingMiddleware(t *testing.T) {
	var mockSpan = &mockSpan{}
	var mockTracer = &mockTracer{
		Span: mockSpan,
	}
	var mockFlaki = &mockFlaki{}

	// Context with correlation ID and span.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", id)
	ctx = opentracing.ContextWithSpan(ctx, mockTracer.StartSpan("flaki"))

	var srv = New(mockFlaki)
	srv = MakeTracingMiddleware(mockTracer)(srv)

	// NextID.
	mockTracer.Called = false
	mockTracer.Span.CorrelationID = ""
	srv.NextID(ctx)
	assert.True(t, mockTracer.Called)
	assert.Equal(t, id, mockTracer.Span.CorrelationID)

	// NextValidID.
	mockTracer.Called = false
	mockTracer.Span.CorrelationID = ""
	srv.NextValidID(ctx)
	assert.True(t, mockTracer.Called)
	assert.Equal(t, id, mockTracer.Span.CorrelationID)

	// NextID without correlation ID.
	var f = func() {
		srv.NextID(opentracing.ContextWithSpan(context.Background(), mockTracer.StartSpan("flaki")))
	}
	assert.Panics(t, f)

	// NextValidID without correlation ID.
	f = func() {
		srv.NextValidID(opentracing.ContextWithSpan(context.Background(), mockTracer.StartSpan("flaki")))
	}
	assert.Panics(t, f)
}

// Mock Logger.
type mockLogger struct {
	Called        bool
	CorrelationID string
}

func (l *mockLogger) Log(keyvals ...interface{}) error {
	l.Called = true

	for i, kv := range keyvals {
		if kv == "correlation_id" {
			l.CorrelationID = keyvals[i+1].(string)
		}
	}
	return nil
}

// Mock counter.
type mockCounter struct {
	Called        bool
	CorrelationID string
}

func (h *mockCounter) With(labelValues ...string) metrics.Counter {
	for i, kv := range labelValues {
		if kv == "correlation_id" {
			h.CorrelationID = labelValues[i+1]
		}
	}
	return h
}

func (h *mockCounter) Add(delta float64) {
	h.Called = true
}

// Mock Tracer.
type mockTracer struct {
	Called bool
	Span   *mockSpan
}

func (t *mockTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	t.Called = true
	return t.Span
}
func (t *mockTracer) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	return nil
}
func (t *mockTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	return nil, nil
}

// Mock Span.
type mockSpan struct {
	CorrelationID string
}

func (s *mockSpan) SetTag(key string, value interface{}) opentracing.Span {
	if key == "correlation_id" {
		s.CorrelationID = value.(string)
	}
	return s
}
func (s *mockSpan) Finish()                                                     {}
func (s *mockSpan) FinishWithOptions(opts opentracing.FinishOptions)            {}
func (s *mockSpan) Context() opentracing.SpanContext                            { return nil }
func (s *mockSpan) SetOperationName(operationName string) opentracing.Span      { return s }
func (s *mockSpan) LogFields(fields ...opentracing_log.Field)                   {}
func (s *mockSpan) LogKV(alternatingKeyValues ...interface{})                   {}
func (s *mockSpan) SetBaggageItem(restrictedKey, value string) opentracing.Span { return s }
func (s *mockSpan) BaggageItem(restrictedKey string) string                     { return "" }
func (s *mockSpan) Tracer() opentracing.Tracer                                  { return nil }
func (s *mockSpan) LogEvent(event string)                                       {}
func (s *mockSpan) LogEventWithPayload(event string, payload interface{})       {}
func (s *mockSpan) Log(data opentracing.LogData)                                {}
