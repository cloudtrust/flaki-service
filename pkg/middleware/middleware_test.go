package middleware

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/flaki"
	"github.com/cloudtrust/flaki-service/pkg/health"
	"github.com/cloudtrust/flaki-service/pkg/middleware/mock"
	"github.com/golang/mock/gomock"
	opentracing "github.com/opentracing/opentracing-go"
)

func TestEndpointCorrelationIDMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockFlakiComponent = mock.NewFlakiComponent(mockCtrl)
	var mockHealthComponent = mock.NewHealthComponent(mockCtrl)

	var flakiEndpoint = flaki.Endpoints{
		NextValidIDEndpoint: flaki.MakeNextValidIDEndpoint(mockFlakiComponent),
	}

	var m = MakeEndpointCorrelationIDMW(flakiEndpoint)(health.MakeInfluxHealthCheckEndpoint(mockHealthComponent))

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	var ctxFID = context.WithValue(context.Background(), "correlation_id", flakiID)

	// Context with correlation ID.
	mockFlakiComponent.EXPECT().NextValidID(gomock.Any()).Return(flakiID).Times(0)
	mockHealthComponent.EXPECT().InfluxHealthChecks(ctx).Return(health.HealthReports{}).Times(1)
	m(ctx, nil)

	// Without correlation ID.
	mockFlakiComponent.EXPECT().NextValidID(gomock.Any()).Return(flakiID).Times(1)
	mockHealthComponent.EXPECT().InfluxHealthChecks(ctxFID).Return(health.HealthReports{}).Times(1)
	m(context.Background(), nil)
}
func TestEndpointLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockLogger = mock.NewLogger(mockCtrl)
	var mockComponent = mock.NewComponent(mockCtrl)

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeEndpointLoggingMW(mockLogger)(flaki.MakeNextIDEndpoint(mockComponent))

	// With correlation ID.
	mockLogger.EXPECT().Log("correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	mockComponent.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
	m(ctx, nil)

	// Without correlation ID.
	mockLogger.EXPECT().Log("correlation_id", flakiID, "took", gomock.Any()).Return(nil).Times(1)
	mockComponent.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
	m(context.Background(), nil)
}

func TestEndpointInstrumentingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockHistogram = mock.NewHistogram(mockCtrl)
	var mockComponent = mock.NewComponent(mockCtrl)

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeEndpointInstrumentingMW(mockHistogram)(flaki.MakeNextIDEndpoint(mockComponent))

	// With correlation ID.
	mockHistogram.EXPECT().With("correlation_id", corrID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	mockComponent.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
	m(ctx, nil)

	// Without correlation ID.
	mockHistogram.EXPECT().With("correlation_id", flakiID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	mockComponent.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
	m(context.Background(), nil)
}

func TestEndpointTracingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)
	var mockTracer = mock.NewTracer(mockCtrl)
	var mockSpan = mock.NewSpan(mockCtrl)
	var mockSpanContext = mock.NewSpanContext(mockCtrl)

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	ctx = opentracing.ContextWithSpan(ctx, mockSpan)

	var m = MakeEndpointTracingMW(mockTracer, "operationName")(flaki.MakeNextIDEndpoint(mockComponent))

	// With correlation ID.
	mockTracer.EXPECT().StartSpan("operationName", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", corrID).Return(mockSpan).Times(1)
	mockComponent.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
	m(ctx, nil)

	// Without correlation ID.
	mockTracer.EXPECT().StartSpan("operationName", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", flakiID).Return(mockSpan).Times(1)
	mockComponent.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
	m(opentracing.ContextWithSpan(context.Background(), mockSpan), nil)

	// Without tracer.
	mockTracer.EXPECT().StartSpan("operationName", gomock.Any()).Times(0)
	mockSpan.EXPECT().Context().Times(0)
	mockSpan.EXPECT().Finish().Times(0)
	mockSpan.EXPECT().SetTag("correlation_id", flakiID).Times(0)
	mockComponent.EXPECT().NextID(gomock.Any()).Return(flakiID, nil).Times(1)
	m(context.Background(), nil)
}
