package server

import (
	"context"
	"github.com/JohanDroz/flaki-service/service/endpoint"
	fb "github.com/JohanDroz/flaki-service/service/transport/flatbuffer/flaki"
	"github.com/google/flatbuffers/go"
)

// A grpcServer is essentially a set of endpoints that can be called via gRPC.
type grpcServer struct {
	endpoints endpoint.Endpoints
}

func NewGrpcServer(endpoints endpoint.Endpoints) fb.FlakiServer {
	return &grpcServer{
		endpoints: endpoints,
	}
}

// Implement the flatbuffer FlakiServer interface
func (s *grpcServer) NextID(ctx context.Context, r *fb.EmptyRequest) (*flatbuffers.Builder, error) {
	var id, err = s.endpoints.NextIDEndpoint(ctx, nil)

	var b = flatbuffers.NewBuilder(0)

	if err != nil {
		var errPos = b.CreateString(err.Error())
		fb.NextIDReplyStart(b)
		fb.NextIDReplyAddError(b, errPos)
		b.Finish(fb.NextIDReplyEnd(b))
		return b, nil
	}

	fb.NextIDReplyStart(b)
	fb.NextIDReplyAddId(b, id.(uint64))
	b.Finish(fb.NextIDReplyEnd(b))
	return b, nil
}

// Implement the protobuf.FlakiServer interface
func (s *grpcServer) NextValidID(ctx context.Context, r *fb.EmptyRequest) (*flatbuffers.Builder, error) {
	var id, _ = s.endpoints.NextValidIDEndpoint(ctx, nil)

	var b = flatbuffers.NewBuilder(0)
	fb.NextValidIDReplyStart(b)
	fb.NextValidIDReplyAddId(b, id.(uint64))
	b.Finish(fb.NextValidIDReplyEnd(b))

	return b, nil
}
