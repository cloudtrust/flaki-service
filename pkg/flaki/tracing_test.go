package flaki

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
)

func TestComponentTracingMW(t *testing.T) {
	var mockSpan = &mockSpan{}
	var mockTracer = &mockTracer{Span: mockSpan}
	var mockComponent = &mockComponent{}

	// Context with correlation ID and span.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", id)
	ctx = opentracing.ContextWithSpan(ctx, mockTracer.StartSpan("flaki"))

	var m = MakeComponentTracingMW(mockTracer)(mockComponent)

	// NextID.
	mockTracer.Called = false
	mockTracer.Span.CorrelationID = ""
	m.NextID(ctx)
	assert.True(t, mockTracer.Called)
	assert.Equal(t, id, mockTracer.Span.CorrelationID)

	// NextValidID.
	mockTracer.Called = false
	mockTracer.Span.CorrelationID = ""
	m.NextValidID(ctx)
	assert.True(t, mockTracer.Called)
	assert.Equal(t, id, mockTracer.Span.CorrelationID)

	// NextID without correlation ID.
	var f = func() {
		m.NextID(opentracing.ContextWithSpan(context.Background(), mockTracer.StartSpan("flaki")))
	}
	assert.Panics(t, f)

	// NextValidID without correlation ID.
	f = func() {
		m.NextValidID(opentracing.ContextWithSpan(context.Background(), mockTracer.StartSpan("flaki")))
	}
	assert.Panics(t, f)
}
func TestModuleTracingMW(t *testing.T) {
	var mockSpan = &mockSpan{}
	var mockTracer = &mockTracer{Span: mockSpan}
	var mockFlaki = &mockFlaki{}

	// Context with correlation ID and span.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", id)
	ctx = opentracing.ContextWithSpan(ctx, mockTracer.StartSpan("flaki"))

	var srv = NewModule(mockFlaki)
	srv = MakeModuleTracingMW(mockTracer)(srv)

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
func (s *mockSpan) LogFields(fields ...olog.Field)                              {}
func (s *mockSpan) LogKV(alternatingKeyValues ...interface{})                   {}
func (s *mockSpan) SetBaggageItem(restrictedKey, value string) opentracing.Span { return s }
func (s *mockSpan) BaggageItem(restrictedKey string) string                     { return "" }
func (s *mockSpan) Tracer() opentracing.Tracer                                  { return nil }
func (s *mockSpan) LogEvent(event string)                                       {}
func (s *mockSpan) LogEventWithPayload(event string, payload interface{})       {}
func (s *mockSpan) Log(data opentracing.LogData)                                {}
