package grpc

import (
	"context"

	grpc_transport "github.com/go-kit/kit/transport/grpc"
	opentracing "github.com/opentracing/opentracing-go"
	otag "github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc/metadata"
)

// Middleware on http transport.
type Middleware func(grpc_transport.Handler) grpc_transport.Handler

type tracingMiddleware struct {
	next          grpc_transport.Handler
	tracer        opentracing.Tracer
	operationName string
}

func (m *tracingMiddleware) ServeGRPC(ctx context.Context, request interface{}) (context.Context, interface{}, error) {
	var md, _ = metadata.FromIncomingContext(ctx)

	// Extract metadata.
	var carrier = make(opentracing.TextMapCarrier)
	for k, v := range md {
		carrier.Set(k, v[0])
	}

	var sc, err = m.tracer.Extract(opentracing.TextMap, carrier)
	var span opentracing.Span
	if err != nil {
		span = m.tracer.StartSpan(m.operationName)
	} else {
		span = m.tracer.StartSpan(m.operationName, opentracing.ChildOf(sc))
	}
	defer span.Finish()

	// Set tags.
	otag.Component.Set(span, "flaki-service")
	span.SetTag("transport", "grpc")
	otag.SpanKindRPCServer.Set(span)

	return m.next.ServeGRPC(opentracing.ContextWithSpan(ctx, span), request)
}

// MakeTracingMiddleware try to extract an existing span from the HTTP headers. It it exists, we
// continue the span, if not we create a new one.
func MakeTracingMiddleware(tracer opentracing.Tracer, operationName string) Middleware {
	return func(next grpc_transport.Handler) grpc_transport.Handler {
		return &tracingMiddleware{
			next:          next,
			tracer:        tracer,
			operationName: operationName,
		}
	}
}
