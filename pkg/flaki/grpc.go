package flaki

import (
	"context"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	"github.com/go-kit/kit/endpoint"
	grpc_transport "github.com/go-kit/kit/transport/grpc"
	"github.com/google/flatbuffers/go"
	"google.golang.org/grpc/metadata"
)

type grpcServer struct {
	nextID      grpc_transport.Handler
	nextValidID grpc_transport.Handler
}

// MakeNextIDGRPCHandler makes a GRPC handler for the NextID endpoint.
func MakeNextIDGRPCHandler(e endpoint.Endpoint) *grpc_transport.Server {
	return grpc_transport.NewServer(
		e,
		decodeFlakiRequest,
		encodeFlakiReply,
		grpc_transport.ServerBefore(fetchCorrelationID),
	)
}

// MakeNextValidIDGRPCHandler makes a GRPC handler for the NextValidID endpoint.
func MakeNextValidIDGRPCHandler(e endpoint.Endpoint) *grpc_transport.Server {
	return grpc_transport.NewServer(
		e,
		decodeFlakiRequest,
		encodeFlakiReply,
		grpc_transport.ServerBefore(fetchCorrelationID),
	)
}

// NewGRPCServer makes a set of handler available as a FlakiServer.
func NewGRPCServer(nextIDHandler, nextValidIDHandler grpc_transport.Handler) fb.FlakiServer {
	return &grpcServer{
		nextID:      nextIDHandler,
		nextValidID: nextValidIDHandler,
	}
}

// fetchGRPCCorrelationID reads the correlation ID from the GRPC metadata.
// If the id is not zero, we put it in the context.
func fetchGRPCCorrelationID(ctx context.Context, md metadata.MD) context.Context {
	var val = md["correlation_id"]

	// If there is no id in the metadata, return current context.
	if val == nil || val[0] == "" {
		return ctx
	}

	// If there is an id in the metadata, add it to the context.
	var id = val[0]
	return context.WithValue(ctx, "correlation_id", id)
}

// Implement the flatbuffer FlakiServer interface.
func (s *grpcServer) NextID(ctx context.Context, req *fb.EmptyRequest) (*flatbuffers.Builder, error) {
	var _, res, err = s.nextID.ServeGRPC(ctx, req)
	if err != nil {
		return flakiErrorHandler(err), nil
	}

	var b = res.(*flatbuffers.Builder)

	return b, nil
}

// Implement the flatbuffer FlakiServer interface.
func (s *grpcServer) NextValidID(ctx context.Context, req *fb.EmptyRequest) (*flatbuffers.Builder, error) {
	var _, res, err = s.nextValidID.ServeGRPC(ctx, req)
	if err != nil {
		return flakiErrorHandler(err), nil
	}

	var b = res.(*flatbuffers.Builder)

	return b, nil
}

// encodeFlakiReply encodes the flatbuffer flaki reply.
func encodeFlakiReply(_ context.Context, res interface{}) (interface{}, error) {
	var b = flatbuffers.NewBuilder(0)
	var id = b.CreateString(res.(string))

	fb.FlakiReplyStart(b)
	fb.FlakiReplyAddId(b, id)
	b.Finish(fb.FlakiReplyEnd(b))

	return b, nil
}

// decodeFlakiRequest decodes the flatbuffer flaki request.
func decodeFlakiRequest(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

// flakiErrorHandler encodes the flatbuffer flaki reply when there is an error.
func flakiErrorHandler(err error) *flatbuffers.Builder {

	var b = flatbuffers.NewBuilder(0)
	var errStr = b.CreateString(err.Error())

	fb.FlakiReplyStart(b)
	fb.FlakiReplyAddId(b, 0)
	fb.FlakiReplyAddError(b, errStr)
	b.Finish(fb.FlakiReplyEnd(b))

	return b
}
