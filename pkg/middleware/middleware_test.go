package middleware

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
)

func TestEndpointLoggingMW(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockEndpoint = MakeMockEndpoint(flakiID, false)
	var mockLogger = &mockLogger{}

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeEndpointLoggingMW(mockLogger)(mockEndpoint)

	// With correlation ID.
	mockLogger.called = false
	mockLogger.correlationID = ""
	m(ctx, nil)
	assert.True(t, mockLogger.called)
	assert.Equal(t, corrID, mockLogger.correlationID)

	// Without correlation ID.
	mockLogger.called = false
	mockLogger.correlationID = ""
	m(context.Background(), nil)
	assert.True(t, mockLogger.called)
	assert.Equal(t, flakiID, mockLogger.correlationID)
}

func TestEndpointInstrumentingMW(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockEndpoint = MakeMockEndpoint(flakiID, false)
	var mockHistogram = &mockHistogram{}

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeEndpointInstrumentingMW(mockHistogram)(mockEndpoint)

	// With correlation ID.
	mockHistogram.called = false
	mockHistogram.correlationID = ""
	m(ctx, nil)
	assert.True(t, mockHistogram.called)
	assert.Equal(t, corrID, mockHistogram.correlationID)

	// Without correlation ID.
	mockHistogram.called = false
	mockHistogram.correlationID = ""
	m(context.Background(), nil)
	assert.True(t, mockHistogram.called)
	assert.Equal(t, flakiID, mockHistogram.correlationID)
}

func TestEndpointTracingMW(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockEndpoint = MakeMockEndpoint(flakiID, false)
	var mockSpan = &mockSpan{}
	var mockTracer = &mockTracer{span: mockSpan}

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	ctx = opentracing.ContextWithSpan(ctx, mockTracer.StartSpan("flaki"))

	var m = MakeEndpointTracingMW(mockTracer, "flaki")(mockEndpoint)

	// With correlation ID.
	mockTracer.called = false
	mockTracer.span.correlationID = ""
	m(ctx, nil)
	assert.True(t, mockTracer.called)
	assert.Equal(t, corrID, mockTracer.span.correlationID)

	// Without correlation ID.
	mockTracer.called = false
	mockTracer.span.correlationID = ""
	m(opentracing.ContextWithSpan(context.Background(), mockTracer.StartSpan("flaki")), nil)
	assert.True(t, mockTracer.called)
	assert.Equal(t, flakiID, mockTracer.span.correlationID)
}

func MakeMockEndpoint(id string, fail bool) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		if fail {
			return "", fmt.Errorf("fail")
		}
		return id, nil
	}
}

// Mock Logger.
type mockLogger struct {
	called        bool
	correlationID string
}

func (l *mockLogger) Log(keyvals ...interface{}) error {
	l.called = true

	for i, kv := range keyvals {
		if kv == "correlation_id" {
			l.correlationID = keyvals[i+1].(string)
		}
	}
	return nil
}

// Mock histogram.
type mockHistogram struct {
	called        bool
	correlationID string
}

func (h *mockHistogram) With(labelValues ...string) metrics.Histogram {
	for i, kv := range labelValues {
		if kv == "correlation_id" {
			h.correlationID = labelValues[i+1]
		}
	}
	return h
}
func (h *mockHistogram) Observe(value float64) {
	h.called = true
}

// Mock Tracer.
type mockTracer struct {
	called bool
	span   *mockSpan
}

func (t *mockTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	t.called = true
	return t.span
}
func (t *mockTracer) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	return nil
}
func (t *mockTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	return nil, nil
}

// Mock Span.
type mockSpan struct {
	correlationID string
}

func (s *mockSpan) SetTag(key string, value interface{}) opentracing.Span {
	if key == "correlation_id" {
		s.correlationID = value.(string)
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
