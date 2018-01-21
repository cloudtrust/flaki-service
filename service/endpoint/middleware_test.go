package endpoint

import (
	"context"
	"fmt"
	"math/rand"
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
	rand.Seed(time.Now().UnixNano())
	var id = rand.Uint64()

	var mockTracer = &mockTracer{Called: false}

	var endpoints = NewEndpoints()
	endpoints = endpoints.MakeNextIDEndpoint(&mockFlakiService{
		id:   id,
		fail: false,
	},
		MakeTracingMiddleware(mockTracer, "flaki"),
	)

	assert.False(t, mockTracer.Called)
	endpoints.NextID(context.Background())
	assert.True(t, mockTracer.Called)

	// Test with already existing span
	mockTracer.Called = false
	var span = mockTracer.StartSpan("flaki")
	var ctx = opentracing.ContextWithSpan(context.Background(), span)

	endpoints = endpoints.MakeNextIDEndpoint(&mockFlakiService{
		id:   id,
		fail: false,
	},
		MakeTracingMiddleware(mockTracer, "flaki"),
	)

	endpoints.NextID(ctx)
	assert.True(t, mockTracer.Called)
}

// Mock Service.
type mockFlakiService struct {
	id   uint64
	fail bool
}

func (s *mockFlakiService) NextID(context.Context) (uint64, error) {
	if s.fail {
		return 0, fmt.Errorf("fail")
	}
	return s.id, nil
}

func (s *mockFlakiService) NextValidID(context.Context) uint64 {
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

// Mock Span
type mockSpan struct {
}

func (s *mockSpan) Finish()                                          {}
func (s *mockSpan) FinishWithOptions(opts opentracing.FinishOptions) {}
func (s *mockSpan) Context() opentracing.SpanContext                 { return nil }

func (s *mockSpan) SetOperationName(operationName string) opentracing.Span { return s }

func (s *mockSpan) SetTag(key string, value interface{}) opentracing.Span       { return s }
func (s *mockSpan) LogFields(fields ...opentracing_log.Field)                   {}
func (s *mockSpan) LogKV(alternatingKeyValues ...interface{})                   {}
func (s *mockSpan) SetBaggageItem(restrictedKey, value string) opentracing.Span { return s }
func (s *mockSpan) BaggageItem(restrictedKey string) string                     { return "" }
func (s *mockSpan) Tracer() opentracing.Tracer                                  { return nil }
func (s *mockSpan) LogEvent(event string)                                       {}
func (s *mockSpan) LogEventWithPayload(event string, payload interface{})       {}
func (s *mockSpan) Log(data opentracing.LogData)                                {}
