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
	fb.FlakiRequestStart(b)
	b.Finish(fb.FlakiRequestEnd(b))

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
	var mockGRPCHandler = mock.NewHandler(mockCtrl)
	var mockTracer = mock.NewTracer(mockCtrl)
	var mockSpan = mock.NewSpan(mockCtrl)
	var mockSpanContext = mock.NewSpanContext(mockCtrl)

	var m = MakeGRPCTracingMW(mockTracer, "componentName", "operationName")(mockGRPCHandler)

	// Flatbuffer request.
	var b = flatbuffers.NewBuilder(0)
	fb.FlakiRequestStart(b)
	b.Finish(fb.FlakiRequestEnd(b))

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
func TestEndpointTracingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)
	var mockTracer = mock.NewTracer(mockCtrl)
	var mockSpan = mock.NewSpan(mockCtrl)
	var mockSpanContext = mock.NewSpanContext(mockCtrl)

	var m = MakeEndpointTracingMW(mockTracer, "operationName")(MakeNextIDEndpoint(mockComponent))

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	ctx = opentracing.ContextWithSpan(ctx, mockSpan)
	var req = createFlakiRequest()
	var reply = createFlakiReply(flakiID)

	// NextID.
	mockComponent.EXPECT().NextID(gomock.Any(), req).Return(reply, nil).Times(1)
	mockTracer.EXPECT().StartSpan("operationName", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", corrID).Return(mockSpan).Times(1)
	m(ctx, req)

	// NextID error.
	mockComponent.EXPECT().NextID(gomock.Any(), req).Return(nil, fmt.Errorf("fail")).Times(1)
	mockTracer.EXPECT().StartSpan("operationName", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", corrID).Return(mockSpan).Times(1)
	m(ctx, req)

	// NextID without correlation ID.
	mockComponent.EXPECT().NextID(gomock.Any(), req).Return(reply, nil).Times(1)
	mockTracer.EXPECT().StartSpan("operationName", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", flakiID).Return(mockSpan).Times(1)
	m(opentracing.ContextWithSpan(context.Background(), mockSpan), req)

	// NextID error without correlation ID.
	mockComponent.EXPECT().NextID(gomock.Any(), req).Return(nil, fmt.Errorf("fail")).Times(1)
	mockTracer.EXPECT().StartSpan("operationName", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", "").Return(mockSpan).Times(1)
	m(opentracing.ContextWithSpan(context.Background(), mockSpan), req)

	// Without tracer.
	mockComponent.EXPECT().NextID(gomock.Any(), req).Return(reply, nil).Times(1)
	m(context.Background(), req)
}
func TestComponentTracingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)
	var mockTracer = mock.NewTracer(mockCtrl)
	var mockSpan = mock.NewSpan(mockCtrl)
	var mockSpanContext = mock.NewSpanContext(mockCtrl)

	var m = MakeComponentTracingMW(mockTracer)(mockComponent)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	ctx = opentracing.ContextWithSpan(ctx, mockSpan)
	var req = createFlakiRequest()
	var reply = createFlakiReply(flakiID)

	// NextID.
	mockComponent.EXPECT().NextID(gomock.Any(), req).Return(reply, nil).Times(1)
	mockTracer.EXPECT().StartSpan("nextid_component", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", corrID).Return(mockSpan).Times(1)
	m.NextID(ctx, req)

	// NextID error.
	mockComponent.EXPECT().NextID(gomock.Any(), req).Return(nil, fmt.Errorf("fail")).Times(1)
	mockTracer.EXPECT().StartSpan("nextid_component", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", corrID).Return(mockSpan).Times(1)
	m.NextID(ctx, req)

	// NextID without correlation ID.
	mockComponent.EXPECT().NextID(gomock.Any(), req).Return(reply, nil).Times(1)
	mockTracer.EXPECT().StartSpan("nextid_component", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", flakiID).Return(mockSpan).Times(1)
	m.NextID(opentracing.ContextWithSpan(context.Background(), mockSpan), req)

	// NextID error without correlation ID.
	mockComponent.EXPECT().NextID(gomock.Any(), req).Return(nil, fmt.Errorf("fail")).Times(1)
	mockTracer.EXPECT().StartSpan("nextid_component", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", "").Return(mockSpan).Times(1)
	m.NextID(opentracing.ContextWithSpan(context.Background(), mockSpan), req)

	// NextID without tracer.
	mockComponent.EXPECT().NextID(gomock.Any(), req).Return(reply, nil).Times(1)
	m.NextID(context.Background(), req)

	// NextValidID.
	mockComponent.EXPECT().NextValidID(gomock.Any(), req).Return(reply).Times(1)
	mockTracer.EXPECT().StartSpan("nextvalidid_component", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", corrID).Return(mockSpan).Times(1)
	m.NextValidID(ctx, req)

	// NextValidID without correlation ID.
	mockComponent.EXPECT().NextValidID(gomock.Any(), req).Return(reply).Times(1)
	mockTracer.EXPECT().StartSpan("nextvalidid_component", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", flakiID).Return(mockSpan).Times(1)
	m.NextValidID(opentracing.ContextWithSpan(context.Background(), mockSpan), req)

	// NextValidID without tracer.
	mockComponent.EXPECT().NextValidID(gomock.Any(), req).Return(reply).Times(1)
	m.NextValidID(context.Background(), req)
}

func TestModuleTracingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockModule = mock.NewModule(mockCtrl)
	var mockTracer = mock.NewTracer(mockCtrl)
	var mockSpan = mock.NewSpan(mockCtrl)
	var mockSpanContext = mock.NewSpanContext(mockCtrl)

	var m = MakeModuleTracingMW(mockTracer)(mockModule)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	ctx = opentracing.ContextWithSpan(ctx, mockSpan)

	// NextID.
	mockModule.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
	mockTracer.EXPECT().StartSpan("nextid_module", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", corrID).Return(mockSpan).Times(1)
	m.NextID(ctx)

	// NextID error.
	mockModule.EXPECT().NextID(gomock.Any()).Return("", fmt.Errorf("fail")).Times(1)
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

	// NextID error without correlation ID.
	mockModule.EXPECT().NextID(gomock.Any()).Return("", fmt.Errorf("fail")).Times(1)
	mockTracer.EXPECT().StartSpan("nextid_module", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", "").Return(mockSpan).Times(1)
	m.NextID(opentracing.ContextWithSpan(context.Background(), mockSpan))

	// NextID without tracer.
	mockModule.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
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
	m.NextValidID(context.Background())
}
