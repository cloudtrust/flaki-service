package flaki

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	"github.com/cloudtrust/flaki-service/pkg/flaki/mock"
	"github.com/golang/mock/gomock"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestNewGRPCServer(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)

	var s = NewGRPCServer(MakeGRPCNextIDHandler(MakeNextIDEndpoint(mockComponent)), MakeGRPCNextValidIDHandler(MakeNextValidIDEndpoint(mockComponent)))

	// Flatbuffer Request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))
	var emptyReq = fb.GetRootAsEmptyRequest(b.FinishedBytes(), 0)

	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)

	// NextID.
	{
		mockComponent.EXPECT().NextID(context.Background()).Return(flakiID, nil).Times(1)
		var data, err = s.NextID(context.Background(), emptyReq)
		assert.Nil(t, err)
		// Decode and check reply.
		var r = fb.GetRootAsFlakiReply(data.FinishedBytes(), 0)
		assert.Equal(t, flakiID, string(r.Id()))
		assert.Zero(t, string(r.Error()))
	}
	// NextValidID.
	{
		mockComponent.EXPECT().NextValidID(context.Background()).Return(flakiID).Times(1)
		var data, err = s.NextValidID(context.Background(), emptyReq)
		assert.Nil(t, err)
		// Decode and check reply.
		var r = fb.GetRootAsFlakiReply(data.FinishedBytes(), 0)
		assert.Equal(t, flakiID, string(r.Id()))
		assert.Zero(t, string(r.Error()))
	}
}

func TestGRPCErrorHandler(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)

	var s = NewGRPCServer(MakeGRPCNextIDHandler(MakeNextIDEndpoint(mockComponent)), MakeGRPCNextValidIDHandler(MakeNextValidIDEndpoint(mockComponent)))

	// Flatbuffer Request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))
	var emptyReq = fb.GetRootAsEmptyRequest(b.FinishedBytes(), 0)

	// NextID.
	{
		mockComponent.EXPECT().NextID(context.Background()).Return("", fmt.Errorf("fail")).Times(1)
		var data, err = s.NextID(context.Background(), emptyReq)
		assert.Nil(t, err)
		// Decode and check reply.
		var r = fb.GetRootAsFlakiReply(data.FinishedBytes(), 0)
		assert.Zero(t, string(r.Id()))
		assert.NotZero(t, string(r.Error()))
	}
}

func TestFetchGRPCCorrelationID(t *testing.T) {
	var mockCtrl = gomock.NewController(t)
	defer mockCtrl.Finish()
	var mockComponent = mock.NewComponent(mockCtrl)

	var s = NewGRPCServer(MakeGRPCNextIDHandler(MakeNextIDEndpoint(mockComponent)), MakeGRPCNextValidIDHandler(MakeNextValidIDEndpoint(mockComponent)))

	// Context with correlation ID.
	rand.Seed(time.Now().UnixNano())
	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var md = metadata.New(map[string]string{"correlation_id": corrID})
	var ctx = metadata.NewIncomingContext(context.Background(), md)

	// Flatbuffer Request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))
	var emptyReq = fb.GetRootAsEmptyRequest(b.FinishedBytes(), 0)

	mockComponent.EXPECT().NextID(context.WithValue(ctx, "correlation_id", corrID)).Return(flakiID, nil).Times(1)
	s.NextID(ctx, emptyReq)
	mockComponent.EXPECT().NextValidID(context.WithValue(ctx, "correlation_id", corrID)).Return(flakiID).Times(1)
	s.NextValidID(ctx, emptyReq)
}
