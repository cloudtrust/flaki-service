package server

import (
	"context"

	flaki_endpoint "github.com/JohanDroz/flaki-service/service/endpoint"
	fb "github.com/JohanDroz/flaki-service/service/transport/flatbuffer/flaki"
	grpc_transport "github.com/go-kit/kit/transport/grpc"
	"github.com/google/flatbuffers/go"
	"google.golang.org/grpc/metadata"
)

type grpcServer struct {
	nextID      grpc_transport.Handler
	nextValidID grpc_transport.Handler
}

// NewGRPCServer makes a set of endpoints available as a grpc server.
func NewGRPCServer(endpoints *flaki_endpoint.Endpoints) fb.FlakiServer {
	return &grpcServer{
		nextID: grpc_transport.NewServer(
			endpoints.NextIDEndpoint,
			decodeNextIDRequest,
			encodeNextIDResponse,
			grpc_transport.ServerBefore(fetchCorrelationID),
		),
		nextValidID: grpc_transport.NewServer(
			endpoints.NextValidIDEndpoint,
			decodeNextValidIDRequest,
			encodeNextValidIDResponse,
			grpc_transport.ServerBefore(fetchCorrelationID),
		),
	}
}

// correlationIDToContext put the correlationID to the context.
func fetchCorrelationID(ctx context.Context, md metadata.MD) context.Context {
	var val = md["id"]

	// If there is no id in the metadata, return current context.
	if val == nil || val[0] == "" {
		return ctx
	}

	// If there is an id in the metadata, add it to the context.
	var id = val[0]
	return context.WithValue(ctx, "id", id)
}

// Implement the flatbuffer FlakiServer interface.
func (s *grpcServer) NextID(ctx context.Context, req *fb.EmptyRequest) (*flatbuffers.Builder, error) {

	var _, res, err = s.nextID.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	var b = res.(*flatbuffers.Builder)

	return b, nil
}

func (s *grpcServer) NextValidID(ctx context.Context, req *fb.EmptyRequest) (*flatbuffers.Builder, error) {
	var _, res, err = s.nextValidID.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	var b = res.(*flatbuffers.Builder)

	return b, nil
}

func encodeNextIDResponse(_ context.Context, res interface{}) (interface{}, error) {
	var id = res.(uint64)

	var b = flatbuffers.NewBuilder(0)
	fb.NextIDReplyStart(b)
	fb.NextIDReplyAddId(b, id)
	b.Finish(fb.NextIDReplyEnd(b))

	return b, nil
}

func encodeNextValidIDResponse(_ context.Context, res interface{}) (interface{}, error) {
	var id = res.(uint64)

	var b = flatbuffers.NewBuilder(0)
	fb.NextValidIDReplyStart(b)
	fb.NextValidIDReplyAddId(b, id)
	b.Finish(fb.NextValidIDReplyEnd(b))

	return b, nil
}

func decodeNextIDRequest(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func decodeNextValidIDRequest(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}
