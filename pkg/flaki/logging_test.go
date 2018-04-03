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

func TestEndpointLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewIDGeneratorComponent(mockCtrl)
	var mockLogger = mock.NewLogger(mockCtrl)

	var m = MakeEndpointLoggingMW(mockLogger)(MakeNextIDEndpoint(mockComponent))

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	var req = createFlakiRequest()
	var reply = createFlakiReply(flakiID)

	// NextID.
	mockComponent.EXPECT().NextID(ctx, req).Return(reply, nil).Times(1)
	mockLogger.EXPECT().Log("correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m(ctx, req)

	// NextID error.
	mockComponent.EXPECT().NextID(ctx, req).Return(nil, fmt.Errorf("fail")).Times(1)
	mockLogger.EXPECT().Log("correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m(ctx, req)

	// NextID without correlation ID.
	mockComponent.EXPECT().NextID(context.Background(), req).Return(reply, nil).Times(1)
	mockLogger.EXPECT().Log("correlation_id", flakiID, "took", gomock.Any()).Return(nil).Times(1)
	m(context.Background(), req)

	// NextID error without correlation ID.
	mockComponent.EXPECT().NextID(context.Background(), req).Return(nil, fmt.Errorf("fail")).Times(1)
	mockLogger.EXPECT().Log("correlation_id", "", "took", gomock.Any()).Return(nil).Times(1)
	m(context.Background(), req)
}

func TestComponentLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockLogger = mock.NewLogger(mockCtrl)
	var mockComponent = mock.NewIDGeneratorComponent(mockCtrl)

	var m = MakeComponentLoggingMW(mockLogger)(mockComponent)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)
	var req = createFlakiRequest()
	var reply = createFlakiReply(flakiID)

	// NextID.
	mockComponent.EXPECT().NextID(ctx, req).Return(reply, nil).Times(1)
	mockLogger.EXPECT().Log("unit", "NextID", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextID(ctx, req)

	// NextID error.
	mockComponent.EXPECT().NextID(ctx, req).Return(nil, fmt.Errorf("fail")).Times(1)
	mockLogger.EXPECT().Log("unit", "NextID", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextID(ctx, req)

	// NextID without correlation ID.
	mockComponent.EXPECT().NextID(context.Background(), req).Return(reply, nil).Times(1)
	mockLogger.EXPECT().Log("unit", "NextID", "correlation_id", flakiID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextID(context.Background(), req)

	// NextID error without correlation ID.
	mockComponent.EXPECT().NextID(context.Background(), req).Return(nil, fmt.Errorf("fail")).Times(1)
	mockLogger.EXPECT().Log("unit", "NextID", "correlation_id", "", "took", gomock.Any()).Return(nil).Times(1)
	m.NextID(context.Background(), req)

	// NextValidID.
	mockComponent.EXPECT().NextValidID(ctx, req).Return(reply).Times(1)
	mockLogger.EXPECT().Log("unit", "NextValidID", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextValidID(ctx, req)

	// NextValidID without correlation ID.
	mockComponent.EXPECT().NextValidID(context.Background(), req).Return(reply).Times(1)
	mockLogger.EXPECT().Log("unit", "NextValidID", "correlation_id", flakiID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextValidID(context.Background(), req)
}

func TestModuleLoggingMW(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockModule = mock.NewIDGeneratorModule(mockCtrl)
	var mockLogger = mock.NewLogger(mockCtrl)

	var m = MakeModuleLoggingMW(mockLogger)(mockModule)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var ctx = context.WithValue(context.Background(), "correlation_id", corrID)

	// NextID.
	mockModule.EXPECT().NextID(ctx).Return(flakiID, nil).Times(1)
	mockLogger.EXPECT().Log("unit", "NextID", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextID(ctx)

	// NextID error.
	mockModule.EXPECT().NextID(ctx).Return("", fmt.Errorf("fail")).Times(1)
	mockLogger.EXPECT().Log("unit", "NextID", "correlation_id", corrID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextID(ctx)

	// NextID without correlation ID.
	mockModule.EXPECT().NextID(context.Background()).Return(flakiID, nil).Times(1)
	mockLogger.EXPECT().Log("unit", "NextID", "correlation_id", flakiID, "took", gomock.Any()).Return(nil).Times(1)
	m.NextID(context.Background())

	// NextID error without correlation ID.
	mockModule.EXPECT().NextID(context.Background()).Return("", fmt.Errorf("fail")).Times(1)
	mockLogger.EXPECT().Log("unit", "NextID", "correlation_id", "", "took", gomock.Any()).Return(nil).Times(1)
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
