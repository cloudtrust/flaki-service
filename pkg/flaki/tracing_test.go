package flaki

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	"github.com/cloudtrust/flaki-service/pkg/flaki/mock"
	"github.com/golang/mock/gomock"
	flatbuffers "github.com/google/flatbuffers/go"
	opentracing "github.com/opentracing/opentracing-go"
)

func TestHTTPTracingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockTracer = mock.NewTracer(mockCtrl)
	var mockSpan = mock.NewSpan(mockCtrl)
	var mockSpanContext = mock.NewSpanContext(mockCtrl)

	var m = MakeHTTPTracingMW(mockTracer, "componentName", "operationName")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// HTTP request.
	var req = httptest.NewRequest("POST", "http://cloudtrust.io/nextid", bytes.NewReader(b.FinishedBytes()))
	var w = httptest.NewRecorder()

	// With existing tracer.
	mockTracer.EXPECT().Extract(opentracing.HTTPHeaders, gomock.Any()).Return(mockSpanContext, nil).Times(1)
	mockTracer.EXPECT().StartSpan("operationName", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag(gomock.Any(), gomock.Any()).Return(mockSpan).Times(3)
	m.ServeHTTP(w, req)

	// Without existing tracer.
	mockTracer.EXPECT().Extract(opentracing.HTTPHeaders, gomock.Any()).Return(nil, fmt.Errorf("fail")).Times(1)
	mockTracer.EXPECT().StartSpan("operationName").Return(mockSpan).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag(gomock.Any(), gomock.Any()).Return(mockSpan).Times(3)
	m.ServeHTTP(w, req)
}

func TestGRPCTracingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockTracer = mock.NewTracer(mockCtrl)
	var mockSpan = mock.NewSpan(mockCtrl)
	var mockSpanContext = mock.NewSpanContext(mockCtrl)
	var mockGRPCHandler = mock.NewHandler(mockCtrl)

	var m = MakeGRPCTracingMW(mockTracer, "componentName", "operationName")(mockGRPCHandler)

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// With existing tracer.
	mockGRPCHandler.EXPECT().ServeGRPC(gomock.Any(), b.FinishedBytes()).Return(context.Background(), nil, nil).Times(1)
	mockTracer.EXPECT().Extract(opentracing.TextMap, gomock.Any()).Return(mockSpanContext, nil).Times(1)
	mockTracer.EXPECT().StartSpan("operationName", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag(gomock.Any(), gomock.Any()).Return(mockSpan).Times(3)
	m.ServeGRPC(context.Background(), b.FinishedBytes())

	// Without existing tracer.
	mockGRPCHandler.EXPECT().ServeGRPC(gomock.Any(), b.FinishedBytes()).Return(context.Background(), nil, nil).Times(1)
	mockTracer.EXPECT().Extract(opentracing.TextMap, gomock.Any()).Return(nil, fmt.Errorf("fail")).Times(1)
	mockTracer.EXPECT().StartSpan("operationName").Return(mockSpan).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag(gomock.Any(), gomock.Any()).Return(mockSpan).Times(3)
	m.ServeGRPC(context.Background(), b.FinishedBytes())
}

func TestComponentTracingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)
	var mockTracer = mock.NewTracer(mockCtrl)
	var mockSpan = mock.NewSpan(mockCtrl)
	var mockSpanContext = mock.NewSpanContext(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	ctx = opentracing.ContextWithSpan(ctx, mockSpan)

	var m = MakeComponentTracingMW(mockTracer)(mockComponent)

	// NextID.
	mockComponent.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
	mockTracer.EXPECT().StartSpan("nextid_component", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", corrID).Return(mockSpan).Times(1)
	m.NextID(ctx)

	// NextID without correlation ID.
	mockComponent.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
	mockTracer.EXPECT().StartSpan("nextid_component", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", flakiID).Return(mockSpan).Times(1)
	m.NextID(opentracing.ContextWithSpan(context.Background(), mockSpan))

	// NextID without tracer.
	mockComponent.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).Times(0)
	mockSpan.EXPECT().Context().Times(0)
	mockSpan.EXPECT().Finish().Times(0)
	mockSpan.EXPECT().SetTag(gomock.Any(), gomock.Any()).Times(0)
	m.NextID(context.Background())

	// NextValidID.
	mockComponent.EXPECT().NextValidID(gomock.Any()).Return(flakiID).Times(1)
	mockTracer.EXPECT().StartSpan("nextvalidid_component", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", corrID).Return(mockSpan).Times(1)
	m.NextValidID(ctx)

	// NextValidID without correlation ID.
	mockComponent.EXPECT().NextValidID(gomock.Any()).Return(flakiID).Times(1)
	mockTracer.EXPECT().StartSpan("nextvalidid_component", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", flakiID).Return(mockSpan).Times(1)
	m.NextValidID(opentracing.ContextWithSpan(context.Background(), mockSpan))

	// NextValidID without tracer.
	mockComponent.EXPECT().NextValidID(gomock.Any()).Return(flakiID).Times(1)
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).Times(0)
	mockSpan.EXPECT().Context().Times(0)
	mockSpan.EXPECT().Finish().Times(0)
	mockSpan.EXPECT().SetTag(gomock.Any(), gomock.Any()).Times(0)
	m.NextValidID(context.Background())
}

func TestModuleTracingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockModule = mock.NewModule(mockCtrl)
	var mockTracer = mock.NewTracer(mockCtrl)
	var mockSpan = mock.NewSpan(mockCtrl)
	var mockSpanContext = mock.NewSpanContext(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	ctx = opentracing.ContextWithSpan(ctx, mockSpan)

	var m = MakeModuleTracingMW(mockTracer)(mockModule)

	// NextID.
	mockModule.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
	mockTracer.EXPECT().StartSpan("nextid_module", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", corrID).Return(mockSpan).Times(1)
	m.NextID(ctx)

	// NextID without correlation ID.
	mockModule.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
	mockTracer.EXPECT().StartSpan("nextid_module", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", flakiID).Return(mockSpan).Times(1)
	m.NextID(opentracing.ContextWithSpan(context.Background(), mockSpan))

	// NextID without tracer.
	mockModule.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).Times(0)
	mockSpan.EXPECT().Context().Times(0)
	mockSpan.EXPECT().Finish().Times(0)
	mockSpan.EXPECT().SetTag(gomock.Any(), gomock.Any()).Times(0)
	m.NextID(context.Background())

	// NextValidID.
	mockModule.EXPECT().NextValidID(gomock.Any()).Return(flakiID).Times(1)
	mockTracer.EXPECT().StartSpan("nextvalidid_module", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", corrID).Return(mockSpan).Times(1)
	m.NextValidID(ctx)

	// NextValidID without correlation ID.
	mockModule.EXPECT().NextValidID(gomock.Any()).Return(flakiID).Times(1)
	mockTracer.EXPECT().StartSpan("nextvalidid_module", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", flakiID).Return(mockSpan).Times(1)
	m.NextValidID(opentracing.ContextWithSpan(context.Background(), mockSpan))

	// NextValidID without tracer.
	mockModule.EXPECT().NextValidID(gomock.Any()).Return(flakiID).Times(1)
	mockTracer.EXPECT().StartSpan(gomock.Any(), gomock.Any()).Times(0)
	mockSpan.EXPECT().Context().Times(0)
	mockSpan.EXPECT().Finish().Times(0)
	mockSpan.EXPECT().SetTag(gomock.Any(), gomock.Any()).Times(0)
	m.NextValidID(context.Background())
}
