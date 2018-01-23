package grpc

import (
	"context"

	"github.com/cloudtrust/flaki-service/service/transport/flatbuffer/fb"
	"github.com/go-kit/kit/endpoint"
	grpc_transport "github.com/go-kit/kit/transport/grpc"
	"github.com/google/flatbuffers/go"
	opentracing "github.com/opentracing/opentracing-go"
	otag "github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc/metadata"
)

type grpcServer struct {
	nextID      grpc_transport.Handler
	nextValidID grpc_transport.Handler
}

// MakeNextIDHandler makes a GRPC handler for the NextID endpoint.
func MakeNextIDHandler(e endpoint.Endpoint, tracer opentracing.Tracer) *grpc_transport.Server {
	return grpc_transport.NewServer(
		e,
		decodeFlakiRequest,
		encodeFlakiReply,
		grpc_transport.ServerBefore(fetchCorrelationID),
		grpc_transport.ServerBefore(makeTracerHandler(tracer, "nextID")),
	)
}

// MakeNextValidIDHandler makes a GRPC handler for the NextValidID endpoint.
func MakeNextValidIDHandler(e endpoint.Endpoint, tracer opentracing.Tracer) *grpc_transport.Server {
	return grpc_transport.NewServer(
		e,
		decodeFlakiRequest,
		encodeFlakiReply,
		grpc_transport.ServerBefore(fetchCorrelationID),
		grpc_transport.ServerBefore(makeTracerHandler(tracer, "nextValidID")),
	)
}

// NewGRPCServer makes a set of handler available as a FlakiServer.
func NewGRPCServer(nextIDHandler, nextValidIDHandler grpc_transport.Handler) fb.FlakiServer {
	return &grpcServer{
		nextID:      nextIDHandler,
		nextValidID: nextValidIDHandler,
	}
}

// fetchCorrelationID reads the correlation ID from the GRPC metadata.
// If the id is not zero, we put it in the context.
func fetchCorrelationID(ctx context.Context, md metadata.MD) context.Context {
	var val = md["correlation-id"]

	// If there is no id in the metadata, return current context.
	if val == nil || val[0] == "" {
		return ctx
	}

	// If there is an id in the metadata, add it to the context.
	var id = val[0]
	return context.WithValue(ctx, "correlation-id", id)
}

// makeTracerHandler try to extract an existing span from the GRPC metadata. It it exists, we
// continue the span, if not we create a new one.
func makeTracerHandler(tracer opentracing.Tracer, operationName string) grpc_transport.ServerRequestFunc {
	return func(ctx context.Context, MD metadata.MD) context.Context {
		// Extract metadata.
		var carrier = make(opentracing.TextMapCarrier)
		for k, v := range MD {
			carrier.Set(k, v[0])
		}

		var sc, err = tracer.Extract(opentracing.TextMap, carrier)
		var span opentracing.Span
		if err != nil {
			span = tracer.StartSpan(operationName)
		} else {
			span = tracer.StartSpan(operationName, opentracing.ChildOf(sc))
		}
		defer span.Finish()

		// Set tags.
		otag.Component.Set(span, "flaki-service")
		span.SetTag("transport", "grpc")
		otag.SpanKindRPCServer.Set(span)

		return opentracing.ContextWithSpan(ctx, span)
	}
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
