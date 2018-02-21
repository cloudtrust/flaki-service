package middleware

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/flaki"
	"github.com/cloudtrust/flaki-service/pkg/middleware/mock"
	"github.com/go-kit/kit/endpoint"
	"github.com/golang/mock/gomock"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

func TestEndpointCorrelationIDMW(t *testing.T) {
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var flakiEndpoint = flaki.Endpoints{
		NextValidIDEndpoint: MakeMockEndpoint(flakiID, false),
	}

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var mockEndpoint = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		var id = ctx.Value("correlation_id")
		assert.Equal(t, corrID, id)
		return nil, nil
	}

	var m = MakeEndpointCorrelationIDMW(flakiEndpoint)(mockEndpoint)
	m(ctx, nil)

	// Without correlation ID.
	mockEndpoint = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		var id = ctx.Value("correlation_id")
		assert.Equal(t, flakiID, id)
		return nil, nil
	}

	m = MakeEndpointCorrelationIDMW(flakiEndpoint)(mockEndpoint)
	m(context.Background(), nil)

	// Flaki returns error.
	flakiEndpoint = flaki.Endpoints{
		NextValidIDEndpoint: MakeMockEndpoint(flakiID, true),
	}
	mockEndpoint = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		// Should not be called.
		assert.True(t, false)
		return nil, nil
	}

	m = MakeEndpointCorrelationIDMW(flakiEndpoint)(mockEndpoint)
	var i, err = m(context.Background(), nil)
	assert.Zero(t, i.(string))
	assert.NotNil(t, err)
}
func TestEndpointLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockLogger = mock.NewLogger(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockEndpoint = MakeMockEndpoint(flakiID, false)

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeEndpointLoggingMW(mockLogger)(mockEndpoint)

	// With correlation ID.
	mockLogger.EXPECT().Log("correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m(ctx, nil)

	// Without correlation ID.
	mockLogger.EXPECT().Log("correlation_id", flakiID, "took", gomock.Any()).Return(nil).Times(1)
	m(context.Background(), nil)
}

func TestEndpointInstrumentingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockHistogram = mock.NewHistogram(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockEndpoint = MakeMockEndpoint(flakiID, false)

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeEndpointInstrumentingMW(mockHistogram)(mockEndpoint)

	// With correlation ID.
	mockHistogram.EXPECT().With("correlation_id", corrID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m(ctx, nil)

	// Without correlation ID.
	mockHistogram.EXPECT().With("correlation_id", flakiID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m(context.Background(), nil)
}

func TestEndpointTracingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockTracer = mock.NewTracer(mockCtrl)
	var mockSpan = mock.NewSpan(mockCtrl)
	var mockSpanContext = mock.NewSpanContext(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockEndpoint = MakeMockEndpoint(flakiID, false)

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	ctx = opentracing.ContextWithSpan(ctx, mockSpan)

	var m = MakeEndpointTracingMW(mockTracer, "operationName")(mockEndpoint)

	// With correlation ID.
	mockTracer.EXPECT().StartSpan("operationName", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", corrID).Return(mockSpan).Times(1)
	m(ctx, nil)

	// Without correlation ID.
	mockTracer.EXPECT().StartSpan("operationName", gomock.Any()).Return(mockSpan).Times(1)
	mockSpan.EXPECT().Context().Return(mockSpanContext).Times(1)
	mockSpan.EXPECT().Finish().Return().Times(1)
	mockSpan.EXPECT().SetTag("correlation_id", flakiID).Return(mockSpan).Times(1)
	m(opentracing.ContextWithSpan(context.Background(), mockSpan), nil)

	// Without tracer.
	mockTracer.EXPECT().StartSpan("operationName", gomock.Any()).Times(0)
	mockSpan.EXPECT().Context().Times(0)
	mockSpan.EXPECT().Finish().Times(0)
	mockSpan.EXPECT().SetTag("correlation_id", flakiID).Times(0)
	m(context.Background(), nil)
}

func MakeMockEndpoint(id string, fail bool) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		if fail {
			return "", fmt.Errorf("fail")
		}
		return id, nil
	}
}
