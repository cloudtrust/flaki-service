package flakic

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	sentry "github.com/getsentry/raven-go"
	opentracing "github.com/opentracing/opentracing-go"
	opentracing_log "github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware(t *testing.T) {
	var mockLogger = &mockLogger{}

	var srv = MakeLoggingMiddleware(mockLogger)(&mockFlakiService{
		fail: false,
	})

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", id)

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

func TestTracingMiddleware(t *testing.T) {
	var mockSpan = &mockSpan{}
	var mockTracer = &mockTracer{
		Span: mockSpan,
	}
	var mockFlakiService = &mockFlakiService{}

	// Context with correlation ID and span.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", id)
	ctx = opentracing.ContextWithSpan(ctx, mockTracer.StartSpan("flaki"))

	var srv = MakeTracingMiddleware(mockTracer)(mockFlakiService)

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
func TestErrorMiddleware(t *testing.T) {
	var mockSentry = &mockSentry{}

	var srv = MakeErrorMiddleware(mockSentry)(&mockFlakiService{
		fail: true,
	})

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", id)

	// NextID.
	mockSentry.Called = false
	mockSentry.CorrelationID = ""
	srv.NextID(ctx)
	assert.True(t, mockSentry.Called)
	assert.Equal(t, id, mockSentry.CorrelationID)

	// NextValidID never returns an error.
	mockSentry.Called = false
	mockSentry.CorrelationID = ""
	srv.NextValidID(ctx)
	assert.False(t, mockSentry.Called)

	// NextID without correlation ID.
	var f = func() {
		srv.NextID(context.Background())
	}
	assert.Panics(t, f)
}

// Mock Flaki service. If fail is set to true, it returns an error.
type mockFlakiService struct {
	fail bool
}

func (s *mockFlakiService) NextID(context.Context) (string, error) {
	if s.fail {
		return "", fmt.Errorf("fail")
	}
	return "", nil
}

func (s *mockFlakiService) NextValidID(context.Context) string {
	return ""
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

// Mock Sentry.
type mockSentry struct {
	Called        bool
	CorrelationID string
}

func (client *mockSentry) CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string {
	client.Called = true
	client.CorrelationID = tags["correlation_id"]
	return ""
}
