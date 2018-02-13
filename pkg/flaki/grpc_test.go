package flaki

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestNewGRPCServer(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockEndpoint = MakeMockEndpoint(flakiID, false)

	var s = NewGRPCServer(MakeGRPCNextIDHandler(mockEndpoint), MakeGRPCNextValidIDHandler(mockEndpoint))

	// Flatbuffer Request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))
	var emptyReq = fb.GetRootAsEmptyRequest(b.FinishedBytes(), 0)

	// NextID.
	{
		var data, err = s.NextID(context.Background(), emptyReq)
		assert.Nil(t, err)
		// Decode and check reply.
		var r = fb.GetRootAsFlakiReply(data.FinishedBytes(), 0)
		assert.Equal(t, flakiID, string(r.Id()))
		assert.Zero(t, string(r.Error()))
	}
	// NextValidID.
	{
		var data, err = s.NextValidID(context.Background(), emptyReq)
		assert.Nil(t, err)
		// Decode and check reply.
		var r = fb.GetRootAsFlakiReply(data.FinishedBytes(), 0)
		assert.Equal(t, flakiID, string(r.Id()))
		assert.Zero(t, string(r.Error()))
	}
}

func TestGRPCErrorHandler(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	var flakiID = strconv.FormatUint(rand.Uint64(), 10)
	var mockEndpoint = MakeMockEndpoint(flakiID, true)

	var s = NewGRPCServer(MakeGRPCNextIDHandler(mockEndpoint), MakeGRPCNextValidIDHandler(mockEndpoint))

	// Flatbuffer Request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))
	var emptyReq = fb.GetRootAsEmptyRequest(b.FinishedBytes(), 0)

	// NextID.
	{
		var data, err = s.NextID(context.Background(), emptyReq)
		assert.Nil(t, err)
		// Decode and check reply.
		var r = fb.GetRootAsFlakiReply(data.FinishedBytes(), 0)
		assert.Zero(t, string(r.Id()))
		assert.NotZero(t, string(r.Error()))
	}
	// NextValidID.
	{
		var data, err = s.NextValidID(context.Background(), emptyReq)
		assert.Nil(t, err)
		// Decode and check reply.
		var r = fb.GetRootAsFlakiReply(data.FinishedBytes(), 0)
		assert.Zero(t, string(r.Id()))
		assert.NotZero(t, string(r.Error()))
	}
}

func TestFetchGRPCCorrelationID(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// Context with correlation ID.
	var corrID = strconv.FormatUint(rand.Uint64(), 10)
	var md = metadata.New(map[string]string{"correlation_id": corrID})
	var ctx = metadata.NewIncomingContext(context.Background(), md)

	var endpoint = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		var id = ctx.Value("correlation_id")
		assert.NotNil(t, id)
		assert.Equal(t, corrID, id.(string))

		return "", nil
	}

	var s = NewGRPCServer(MakeGRPCNextIDHandler(endpoint), MakeGRPCNextValidIDHandler(endpoint))

	// Flatbuffer Request.
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))
	var emptyReq = fb.GetRootAsEmptyRequest(b.FinishedBytes(), 0)

	s.NextID(ctx, emptyReq)
	s.NextValidID(ctx, emptyReq)
}
