package endpoint

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	opentracing_log "github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
)

func TestMakeCorrelationIDMiddleware(t *testing.T) {
}

func TestMakeLoggingMiddleware(t *testing.T) {

}
func TestMakeMetricMiddleware(t *testing.T) {

}
func TestMakeTracingMiddleware(t *testing.T) {
	var mockTracer = &mockTracer{}

	// Context with correlation ID and span.
	rand.Seed(time.Now().UnixNano())
	var id = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation-id", id)
	ctx = opentracing.ContextWithSpan(ctx, mockTracer.StartSpan("flaki"))

	var endpoints = NewEndpoints(MakeTracingMiddleware(mockTracer, "flaki"))

	// NextID.
	endpoints = endpoints.MakeNextIDEndpoint(&mockFlakiService{
		id:   id,
		fail: false,
	},
	)
	mockTracer.Called = false
	endpoints.NextID(ctx)
	assert.True(t, mockTracer.Called)

	// NextValidID.
	endpoints = endpoints.MakeNextValidIDEndpoint(&mockFlakiService{
		id:   id,
		fail: false,
	},
	)
	mockTracer.Called = false
	endpoints.NextValidID(ctx)
	assert.True(t, mockTracer.Called)
}

// Mock Service.
type mockFlakiService struct {
	id   string
	fail bool
}

func (s *mockFlakiService) NextID(context.Context) (string, error) {
	if s.fail {
		return "", fmt.Errorf("fail")
	}
	return s.id, nil
}

func (s *mockFlakiService) NextValidID(context.Context) string {
	return s.id
}

// Mock Tracer
type mockTracer struct {
	Called bool
}

func (t *mockTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	t.Called = true
	return &mockSpan{}
}
func (t *mockTracer) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	return nil
}
func (t *mockTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	return nil, nil
}

// Mock Span.
type mockSpan struct {
}

func (s *mockSpan) Finish()                                                     {}
func (s *mockSpan) FinishWithOptions(opts opentracing.FinishOptions)            {}
func (s *mockSpan) Context() opentracing.SpanContext                            { return nil }
func (s *mockSpan) SetOperationName(operationName string) opentracing.Span      { return s }
func (s *mockSpan) SetTag(key string, value interface{}) opentracing.Span       { return s }
func (s *mockSpan) LogFields(fields ...opentracing_log.Field)                   {}
func (s *mockSpan) LogKV(alternatingKeyValues ...interface{})                   {}
func (s *mockSpan) SetBaggageItem(restrictedKey, value string) opentracing.Span { return s }
func (s *mockSpan) BaggageItem(restrictedKey string) string                     { return "" }
func (s *mockSpan) Tracer() opentracing.Tracer                                  { return nil }
func (s *mockSpan) LogEvent(event string)                                       {}
func (s *mockSpan) LogEventWithPayload(event string, payload interface{})       {}
func (s *mockSpan) Log(data opentracing.LogData)                                {}
