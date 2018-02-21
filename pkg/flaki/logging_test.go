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

func TestComponentLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockLogger = mock.NewLogger(mockCtrl)
	var mockComponent = mock.NewComponent(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeComponentLoggingMW(mockLogger)(mockComponent)

	// NextID.
	mockComponent.EXPECT().NextID(ctx).Return(flakiID, nil).Times(1)
	mockLogger.EXPECT().Log("unit", "NextID", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextID(ctx)

	// NextID without correlation ID.
	mockComponent.EXPECT().NextID(context.Background()).Return(flakiID, nil).Times(1)
	mockLogger.EXPECT().Log("unit", "NextID", "correlation_id", flakiID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextID(context.Background())

	// NextValidID.
	mockComponent.EXPECT().NextValidID(ctx).Return(flakiID).Times(1)
	mockLogger.EXPECT().Log("unit", "NextValidID", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextValidID(ctx)

	// NextValidID without correlation ID.
	mockComponent.EXPECT().NextValidID(context.Background()).Return(flakiID).Times(1)
	mockLogger.EXPECT().Log("unit", "NextValidID", "correlation_id", flakiID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextValidID(context.Background())
}

func TestModuleLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockLogger = mock.NewLogger(mockCtrl)
	var mockModule = mock.NewModule(mockCtrl)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	var m = MakeModuleLoggingMW(mockLogger)(mockModule)

	// NextID.
	mockModule.EXPECT().NextID(ctx).Return(flakiID, nil).Times(1)
	mockLogger.EXPECT().Log("unit", "NextID", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextID(ctx)

	// NextID without correlation ID.
	mockModule.EXPECT().NextID(context.Background()).Return(flakiID, nil).Times(1)
	mockLogger.EXPECT().Log("unit", "NextID", "correlation_id", flakiID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextID(context.Background())

	// NextValidID.
	mockModule.EXPECT().NextValidID(ctx).Return(flakiID).Times(1)
	mockLogger.EXPECT().Log("unit", "NextValidID", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextValidID(ctx)

	// NextValidID without correlation ID.
	mockModule.EXPECT().NextValidID(context.Background()).Return(flakiID).Times(1)
	mockLogger.EXPECT().Log("unit", "NextValidID", "correlation_id", flakiID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextValidID(context.Background())
}
