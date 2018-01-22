package grpc

import (
	"context"
	"fmt"

	fb "github.com/cloudtrust/flaki-service/service/transport/flatbuffer/flaki"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	grpc_transport "github.com/go-kit/kit/transport/grpc"
	"github.com/google/flatbuffers/go"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/metadata"
)

type grpcServer struct {
	nextID      grpc_transport.Handler
	nextValidID grpc_transport.Handler
}

func MakeNextIDHandler(e endpoint.Endpoint, log log.Logger, tracer opentracing.Tracer) *grpc_transport.Server {
	return grpc_transport.NewServer(
		e,
		decodeNextIDRequest,
		encodeNextIDResponse,
		grpc_transport.ServerBefore(fetchCorrelationID),
		grpc_transport.ServerBefore(makeTracerHandler(tracer, "nextID")),
	)
}

func MakeNextValidIDHandler(e endpoint.Endpoint, log log.Logger, tracer opentracing.Tracer) *grpc_transport.Server {
	return grpc_transport.NewServer(
		e,
		decodeNextValidIDRequest,
		encodeNextValidIDResponse,
		grpc_transport.ServerBefore(fetchCorrelationID),
		grpc_transport.ServerBefore(makeTracerHandler(tracer, "nextValidID")),
	)
}

// NewGRPCServer makes a set of endpoints available as a grpc server.
func NewGRPCServer(nextIDHandler, nextValidIDHandler grpc_transport.Handler) fb.FlakiServer {
	return &grpcServer{
		nextID:      nextIDHandler,
		nextValidID: nextValidIDHandler,
	}
}

// correlationIDToContext put the correlationID to the context.
func fetchCorrelationID(ctx context.Context, md metadata.MD) context.Context {
	var val = md["correlation-id"]

	// If there is no id in the metadata, return current context.
	if val == nil || val[0] == "" {
		return ctx
	}

	// If there is an id in the metadata, add it to the context.
	var id = val[0]
	return context.WithValue(ctx, "correlationID", id)
}

func makeTracerHandler(tracer opentracing.Tracer, operationName string) grpc_transport.ServerRequestFunc {
	return func(ctx context.Context, MD metadata.MD) context.Context {
		var m = make(opentracing.TextMapCarrier)
		for k, v := range MD {
			m.Set(k, v[0])
		}
		var sc, err = tracer.Extract(opentracing.TextMap, m)

		var span opentracing.Span
		if err != nil {
			span = tracer.StartSpan(operationName)
		} else {
			span = tracer.StartSpan(operationName, opentracing.ChildOf(sc))
		}
		defer span.Finish()

		sc.ForeachBaggageItem(func(k, v string) bool {
			fmt.Printf("key: %s, val: %s\n", k, v)
			return true
		})
		return opentracing.ContextWithSpan(ctx, span)
	}
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
