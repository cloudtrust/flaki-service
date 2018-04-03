package flaki

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/cloudtrust/flaki-service/api/fb"
	"github.com/cloudtrust/flaki-service/pkg/flaki/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestNewGRPCServer(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewIDGeneratorComponent(mockCtrl)

	var s = NewGRPCServer(MakeGRPCNextIDHandler(MakeNextIDEndpoint(mockComponent)), MakeGRPCNextValidIDHandler(MakeNextValidIDEndpoint(mockComponent)))

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var req = createFlakiRequest()

	// NextID.
	{
		mockComponent.EXPECT().NextID(context.Background(), req).Return(createFlakiReply(flakiID), nil).Times(1)
		var data, err = s.NextID(context.Background(), req)
		assert.Nil(t, err)
		// Decode and check reply.
		var r = fb.GetRootAsFlakiReply(data.FinishedBytes(), 0)
		assert.Equal(t, flakiID, string(r.Id()))
	}

	// NextValidID.
	{
		mockComponent.EXPECT().NextValidID(context.Background(), req).Return(createFlakiReply(flakiID)).Times(1)
		var data, err = s.NextValidID(context.Background(), req)
		assert.Nil(t, err)
		// Decode and check reply.
		var r = fb.GetRootAsFlakiReply(data.FinishedBytes(), 0)
		assert.Equal(t, flakiID, string(r.Id()))
	}
}

func TestGRPCErrorHandler(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewIDGeneratorComponent(mockCtrl)

	var s = NewGRPCServer(MakeGRPCNextIDHandler(MakeNextIDEndpoint(mockComponent)), MakeGRPCNextValidIDHandler(MakeNextValidIDEndpoint(mockComponent)))

	var req = createFlakiRequest()

	// NextID.
	mockComponent.EXPECT().NextID(context.Background(), req).Return(nil, fmt.Errorf("fail")).Times(1)
	var reply, err = s.NextID(context.Background(), req)
	assert.NotNil(t, err)
	assert.Nil(t, reply)
}

func TestFetchGRPCCorrelationID(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewIDGeneratorComponent(mockCtrl)

	var s = NewGRPCServer(MakeGRPCNextIDHandler(MakeNextIDEndpoint(mockComponent)), MakeGRPCNextValidIDHandler(MakeNextValidIDEndpoint(mockComponent)))

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var md = metadata.New(map[string]string{"correlation_id": corrID})
	var ctx = metadata.NewIncomingContext(context.Background(), md)
	var req = createFlakiRequest()
	var rep = createFlakiReply(flakiID)

	// NextID.
	mockComponent.EXPECT().NextID(context.WithValue(ctx, "correlation_id", corrID), req).Return(rep, nil).Times(1)
	s.NextID(ctx, req)

	// NextID without correlation ID.
	mockComponent.EXPECT().NextID(context.Background(), req).Return(rep, nil).Times(1)
	s.NextID(context.Background(), req)

	// NextValidID.
	mockComponent.EXPECT().NextValidID(context.WithValue(ctx, "correlation_id", corrID), req).Return(rep).Times(1)
	s.NextValidID(ctx, req)

	// NextValidID without correlation ID.
	mockComponent.EXPECT().NextValidID(context.Background(), req).Return(rep).Times(1)
	s.NextValidID(context.Background(), req)
}
