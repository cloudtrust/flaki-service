package flaki

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/flaki/mock"
	"github.com/golang/mock/gomock"
)

func TestEndpointInstrumentingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)
	var mockHistogram = mock.NewHistogram(mockCtrl)

	var m = MakeEndpointInstrumentingMW(mockHistogram)(MakeNextIDEndpoint(mockComponent))

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), CorrelationIDKey, corrID)
	var req = createFlakiRequest()
	var reply = createFlakiReply(flakiID)

	// NextID.
	mockComponent.EXPECT().NextID(ctx, req).Return(reply, nil).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, corrID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m(ctx, req)

	// NextID error.
	mockComponent.EXPECT().NextID(ctx, req).Return(nil, fmt.Errorf("fail")).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, corrID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m(ctx, req)

	// NextID without correlation ID.
	mockComponent.EXPECT().NextID(context.Background(), req).Return(reply, nil).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, flakiID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m(context.Background(), req)

	// NextID error without correlation ID.
	mockComponent.EXPECT().NextID(context.Background(), req).Return(nil, fmt.Errorf("fail")).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, "").Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m(context.Background(), req)
}

func TestComponentInstrumentingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)
	var mockHistogram = mock.NewHistogram(mockCtrl)

	var m = MakeComponentInstrumentingMW(mockHistogram)(mockComponent)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), CorrelationIDKey, corrID)
	var req = createFlakiRequest()
	var reply = createFlakiReply(flakiID)

	// NextID.
	mockComponent.EXPECT().NextID(ctx, req).Return(reply, nil).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, corrID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextID(ctx, req)

	// NextID error.
	mockComponent.EXPECT().NextID(ctx, req).Return(nil, fmt.Errorf("fail")).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, corrID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextID(ctx, req)

	// NextID without correlation ID.
	mockComponent.EXPECT().NextID(context.Background(), req).Return(reply, nil).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, flakiID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextID(context.Background(), req)

	// NextID error without correlation ID.
	mockComponent.EXPECT().NextID(context.Background(), req).Return(nil, fmt.Errorf("fail")).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, "").Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextID(context.Background(), req)

	// NextValidID.
	mockComponent.EXPECT().NextValidID(ctx, req).Return(reply).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, corrID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextValidID(ctx, req)

	// NextValidID without correlation ID.
	mockComponent.EXPECT().NextValidID(context.Background(), req).Return(reply).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, flakiID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextValidID(context.Background(), req)
}

func TestModuleInstrumentingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockModule = mock.NewModule(mockCtrl)
	var mockHistogram = mock.NewHistogram(mockCtrl)

	var m = MakeModuleInstrumentingMW(mockHistogram)(mockModule)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), CorrelationIDKey, corrID)

	// NextID.
	mockModule.EXPECT().NextID(ctx).Return(flakiID, nil).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, corrID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextID(ctx)

	// NextID error.
	mockModule.EXPECT().NextID(ctx).Return("", fmt.Errorf("fail")).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, corrID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextID(ctx)

	// NextID without correlation ID.
	mockModule.EXPECT().NextID(context.Background()).Return(flakiID, nil).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, flakiID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextID(context.Background())

	// NextID error without correlation ID.
	mockModule.EXPECT().NextID(context.Background()).Return("", fmt.Errorf("fail")).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, "").Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextID(context.Background())

	// NextValidID.
	mockModule.EXPECT().NextValidID(ctx).Return(flakiID).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, corrID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextValidID(ctx)

	// NextValidID without correlation ID.
	mockModule.EXPECT().NextValidID(context.Background()).Return(flakiID).Times(1)
	mockHistogram.EXPECT().With(MetricCorrelationIDKey, flakiID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextValidID(context.Background())
}

func TestModuleInstrumentingCounterMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockModule = mock.NewModule(mockCtrl)
	var mockCounter = mock.NewCounter(mockCtrl)

	var m = MakeModuleInstrumentingCounterMW(mockCounter)(mockModule)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), CorrelationIDKey, corrID)

	// NextID.
	mockModule.EXPECT().NextID(ctx).Return(flakiID, nil).Times(1)
	mockCounter.EXPECT().With(MetricCorrelationIDKey, corrID).Return(mockCounter).Times(1)
	mockCounter.EXPECT().Add(float64(1)).Return().Times(1)
	m.NextID(ctx)

	// NextID error.
	mockModule.EXPECT().NextID(ctx).Return("", fmt.Errorf("fail")).Times(1)
	mockCounter.EXPECT().With(MetricCorrelationIDKey, corrID).Return(mockCounter).Times(1)
	mockCounter.EXPECT().Add(float64(1)).Return().Times(1)
	m.NextID(ctx)

	// NextID without correlation ID.
	mockModule.EXPECT().NextID(context.Background()).Return(flakiID, nil).Times(1)
	mockCounter.EXPECT().With(MetricCorrelationIDKey, flakiID).Return(mockCounter).Times(1)
	mockCounter.EXPECT().Add(float64(1)).Return().Times(1)
	m.NextID(context.Background())

	// NextID error without correlation ID.
	mockModule.EXPECT().NextID(context.Background()).Return("", fmt.Errorf("fail")).Times(1)
	mockCounter.EXPECT().With(MetricCorrelationIDKey, "").Return(mockCounter).Times(1)
	mockCounter.EXPECT().Add(float64(1)).Return().Times(1)
	m.NextID(context.Background())

	// NextValidID.
	mockModule.EXPECT().NextValidID(ctx).Return(flakiID).Times(1)
	mockCounter.EXPECT().With(MetricCorrelationIDKey, corrID).Return(mockCounter).Times(1)
	mockCounter.EXPECT().Add(float64(1)).Return().Times(1)
	m.NextValidID(ctx)

	// NextValidID without correlation ID.
	mockModule.EXPECT().NextValidID(context.Background()).Return(flakiID).Times(1)
	mockCounter.EXPECT().With(MetricCorrelationIDKey, flakiID).Return(mockCounter).Times(1)
	mockCounter.EXPECT().Add(float64(1)).Return().Times(1)
	m.NextValidID(context.Background())
}
