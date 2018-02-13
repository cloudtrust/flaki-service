package flaki

import (
	"bytes"
	"context"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	flatbuffers "github.com/google/flatbuffers/go"
	opentracing "github.com/opentracing/opentracing-go"
	olog "github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
)

func TestHTTPTracingMW(t *testing.T) {
	var mockSpan = &mockSpan{}
	var mockTracer = &mockTracer{span: mockSpan}
	var handler = func(w http.ResponseWriter, r *http.Request) {
	}

	var m = MakeHTTPTracingMW(mockTracer, "component", "operation")(http.HandlerFunc(handler))

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// HTTP request.
	var req = httptest.NewRequest("POST", "http://cloudtrust.io/nextid", bytes.NewReader(b.FinishedBytes()))
	var w = httptest.NewRecorder()

	mockTracer.called = false
	m.ServeHTTP(w, req)
	assert.True(t, mockTracer.called)
}

func TestGRPCTracingMW(t *testing.T) {
	var mockSpan = &mockSpan{}
	var mockTracer = &mockTracer{called: false, span: mockSpan}
	var mockGRPCHandler = &mockGRPCHandler{}

	var m = MakeGRPCTracingMW(mockTracer, "component", "operation")(mockGRPCHandler)

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	mockTracer.called = false
	m.ServeGRPC(context.Background(), b.FinishedBytes())
	assert.True(t, mockTracer.called)
}

func TestComponentTracingMW(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockComponent = &mockComponent{fail: false, id: flakiID}
	var mockSpan = &mockSpan{}
	var mockTracer = &mockTracer{span: mockSpan}

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	ctx = opentracing.ContextWithSpan(ctx, mockTracer.StartSpan("flaki"))

	var m = MakeComponentTracingMW(mockTracer)(mockComponent)

	// NextID.
	mockTracer.called = false
	mockTracer.span.correlationID = ""
	m.NextID(ctx)
	assert.True(t, mockTracer.called)
	assert.Equal(t, corrID, mockTracer.span.correlationID)

	// NextValidID.
	mockTracer.called = false
	mockTracer.span.correlationID = ""
	m.NextValidID(ctx)
	assert.True(t, mockTracer.called)
	assert.Equal(t, corrID, mockTracer.span.correlationID)

	// NextID without correlation ID.
	mockTracer.called = false
	mockTracer.span.correlationID = ""
	m.NextID(opentracing.ContextWithSpan(context.Background(), mockTracer.StartSpan("flaki")))
	assert.True(t, mockTracer.called)
	assert.Equal(t, flakiID, mockTracer.span.correlationID)

	// NextValidID without correlation ID.
	mockTracer.called = false
	mockTracer.span.correlationID = ""
	m.NextValidID(opentracing.ContextWithSpan(context.Background(), mockTracer.StartSpan("flaki")))
	assert.True(t, mockTracer.called)
	assert.Equal(t, flakiID, mockTracer.span.correlationID)

	// NextID without tracer.
	mockTracer.called = false
	m.NextID(context.Background())
	assert.False(t, mockTracer.called)

	// NextValidID without tracer.
	mockTracer.called = false
	m.NextValidID(context.Background())
	assert.False(t, mockTracer.called)
}

func TestModuleTracingMW(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockModule = &mockModule{fail: false, id: flakiID}
	var mockSpan = &mockSpan{}
	var mockTracer = &mockTracer{span: mockSpan}

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	ctx = opentracing.ContextWithSpan(ctx, mockTracer.StartSpan("flaki"))

	var m = MakeModuleTracingMW(mockTracer)(mockModule)

	// NextID.
	mockTracer.called = false
	mockTracer.span.correlationID = ""
	m.NextID(ctx)
	assert.True(t, mockTracer.called)
	assert.Equal(t, corrID, mockTracer.span.correlationID)

	// NextValidID.
	mockTracer.called = false
	mockTracer.span.correlationID = ""
	m.NextValidID(ctx)
	assert.True(t, mockTracer.called)
	assert.Equal(t, corrID, mockTracer.span.correlationID)

	// NextID without correlation ID.
	mockTracer.called = false
	mockTracer.span.correlationID = ""
	m.NextID(opentracing.ContextWithSpan(context.Background(), mockTracer.StartSpan("flaki")))
	assert.True(t, mockTracer.called)
	assert.Equal(t, flakiID, mockTracer.span.correlationID)

	// NextValidID without correlation ID.
	mockTracer.called = false
	mockTracer.span.correlationID = ""
	m.NextValidID(opentracing.ContextWithSpan(context.Background(), mockTracer.StartSpan("flaki")))
	assert.True(t, mockTracer.called)
	assert.Equal(t, flakiID, mockTracer.span.correlationID)

	// NextID without tracer.
	mockTracer.called = false
	m.NextID(context.Background())
	assert.False(t, mockTracer.called)

	// NextValidID without tracer.
	mockTracer.called = false
	m.NextValidID(context.Background())
	assert.False(t, mockTracer.called)
}

// Mock GRPC Handler.

type mockGRPCHandler struct {
	called bool
}

func (h *mockGRPCHandler) ServeGRPC(ctx context.Context, request interface{}) (context.Context, interface{}, error) {
	h.called = true
	return ctx, nil, nil
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
