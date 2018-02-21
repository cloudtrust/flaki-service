package flaki

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/flaki/mock"
	"github.com/golang/mock/gomock"
)

func TestComponentInstrumentingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)
	var mockHistogram = mock.NewHistogram(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeComponentInstrumentingMW(mockHistogram)(mockComponent)

	// NextID.
	mockComponent.EXPECT().NextID(ctx).Return(flakiID, nil).Times(1)
	mockHistogram.EXPECT().With("correlation_id", corrID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextID(ctx)

	// NextID without correlation ID.
	mockComponent.EXPECT().NextID(context.Background()).Return(flakiID, nil).Times(1)
	mockHistogram.EXPECT().With("correlation_id", flakiID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextID(context.Background())

	// NextValidID.
	mockComponent.EXPECT().NextValidID(ctx).Return(flakiID).Times(1)
	mockHistogram.EXPECT().With("correlation_id", corrID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextValidID(ctx)

	// NextValidID without correlation ID.
	mockComponent.EXPECT().NextValidID(context.Background()).Return(flakiID).Times(1)
	mockHistogram.EXPECT().With("correlation_id", flakiID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextValidID(context.Background())
}

func TestModuleInstrumentingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockModule = mock.NewModule(mockCtrl)
	var mockHistogram = mock.NewHistogram(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeModuleInstrumentingMW(mockHistogram)(mockModule)

	// NextID.
	mockModule.EXPECT().NextID(ctx).Return(flakiID, nil).Times(1)
	mockHistogram.EXPECT().With("correlation_id", corrID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextID(ctx)

	// NextID without correlation ID.
	mockModule.EXPECT().NextID(context.Background()).Return(flakiID, nil).Times(1)
	mockHistogram.EXPECT().With("correlation_id", flakiID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextID(context.Background())

	// NextValidID.
	mockModule.EXPECT().NextValidID(ctx).Return(flakiID).Times(1)
	mockHistogram.EXPECT().With("correlation_id", corrID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextValidID(ctx)

	// NextValidID without correlation ID.
	mockModule.EXPECT().NextValidID(context.Background()).Return(flakiID).Times(1)
	mockHistogram.EXPECT().With("correlation_id", flakiID).Return(mockHistogram).Times(1)
	mockHistogram.EXPECT().Observe(gomock.Any()).Return().Times(1)
	m.NextValidID(context.Background())
}

func TestModuleInstrumentingCounterMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockModule = mock.NewModule(mockCtrl)
	var mockCounter = mock.NewCounter(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeModuleInstrumentingCounterMW(mockCounter)(mockModule)

	// NextID.
	mockModule.EXPECT().NextID(ctx).Return(flakiID, nil).Times(1)
	mockCounter.EXPECT().With("correlation_id", corrID).Return(mockCounter).Times(1)
	mockCounter.EXPECT().Add(float64(1)).Return().Times(1)
	m.NextID(ctx)

	// NextID without correlation ID.
	mockModule.EXPECT().NextID(context.Background()).Return(flakiID, nil).Times(1)
	mockCounter.EXPECT().With("correlation_id", flakiID).Return(mockCounter).Times(1)
	mockCounter.EXPECT().Add(float64(1)).Return().Times(1)
	m.NextID(context.Background())

	// NextValidID.
	mockModule.EXPECT().NextValidID(ctx).Return(flakiID).Times(1)
	mockCounter.EXPECT().With("correlation_id", corrID).Return(mockCounter).Times(1)
	mockCounter.EXPECT().Add(float64(1)).Return().Times(1)
	m.NextValidID(ctx)

	// NextValidID without correlation ID.
	mockModule.EXPECT().NextValidID(context.Background()).Return(flakiID).Times(1)
	mockCounter.EXPECT().With("correlation_id", flakiID).Return(mockCounter).Times(1)
	mockCounter.EXPECT().Add(float64(1)).Return().Times(1)
	m.NextValidID(context.Background())
}
