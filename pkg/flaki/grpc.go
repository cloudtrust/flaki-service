package flaki

//go:generate mockgen -destination=./mock/grpc.go -package=mock -mock_names=Handler=Handler github.com/go-kit/kit/transport/grpc Handler

import (
	"context"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	"github.com/go-kit/kit/endpoint"
	grpc_transport "github.com/go-kit/kit/transport/grpc"
	"github.com/google/flatbuffers/go"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

const (
	grpcCorrelationIDKey = "correlation_id"
)

type grpcServer struct {
	nextID      grpc_transport.Handler
	nextValidID grpc_transport.Handler
}

// MakeGRPCNextIDHandler makes a GRPC handler for the NextID endpoint.
func MakeGRPCNextIDHandler(e endpoint.Endpoint) *grpc_transport.Server {
	return grpc_transport.NewServer(
		e,
		decodeGRPCRequest,
		encodeGRPCReply,
		grpc_transport.ServerBefore(fetchGRPCCorrelationID),
	)
}

// MakeGRPCNextValidIDHandler makes a GRPC handler for the NextValidID endpoint.
func MakeGRPCNextValidIDHandler(e endpoint.Endpoint) *grpc_transport.Server {
	return grpc_transport.NewServer(
		e,
		decodeGRPCRequest,
		encodeGRPCReply,
		grpc_transport.ServerBefore(fetchGRPCCorrelationID),
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
	var val = md[grpcCorrelationIDKey]

	// If there is no id in the metadata, return current context.
	if val == nil || val[0] == "" {
		return ctx
	}

	// If there is an id in the metadata, add it to the context.
	var id = val[0]
	return context.WithValue(ctx, CorrelationIDKey, id)
}

// Implement the flatbuffer FlakiServer interface.
func (s *grpcServer) NextID(ctx context.Context, req *fb.FlakiRequest) (*flatbuffers.Builder, error) {
	var _, rep, err = s.nextID.ServeGRPC(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "grpc server could not return next ID")
	}

	var reply = rep.(*fb.FlakiReply)

	var b = flatbuffers.NewBuilder(0)
	var str = b.CreateString(string(reply.Id()))

	fb.FlakiReplyStart(b)
	fb.FlakiReplyAddId(b, str)
	b.Finish(fb.FlakiReplyEnd(b))

	return b, nil
}

// Implement the flatbuffer FlakiServer interface.
func (s *grpcServer) NextValidID(ctx context.Context, req *fb.FlakiRequest) (*flatbuffers.Builder, error) {
	var _, rep, err = s.nextValidID.ServeGRPC(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "grpc server could not return next valid ID")
	}

	var reply = rep.(*fb.FlakiReply)

	var b = flatbuffers.NewBuilder(0)
	var str = b.CreateString(string(reply.Id()))

	fb.FlakiReplyStart(b)
	fb.FlakiReplyAddId(b, str)
	b.Finish(fb.FlakiReplyEnd(b))

	return b, nil
}

// decodeGRPCRequest decodes the flatbuffer flaki request.
func decodeGRPCRequest(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

// encodeHTTPReply encodes the flatbuffer flaki reply.
func encodeGRPCReply(_ context.Context, rep interface{}) (interface{}, error) {
	return rep, nil
}
