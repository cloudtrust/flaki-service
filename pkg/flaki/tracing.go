package flaki

//go:generate mockgen -destination=./mock/tracing.go -package=mock -mock_names=Tracer=Tracer,Span=Span,SpanContext=SpanContext github.com/opentracing/opentracing-go Tracer,Span,SpanContext

import (
	"context"
	"net/http"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	"github.com/go-kit/kit/endpoint"
	grpc_transport "github.com/go-kit/kit/transport/grpc"
	opentracing "github.com/opentracing/opentracing-go"
	otag "github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc/metadata"
)

// MakeHTTPTracingMW try to extract an existing span from the HTTP headers. It it exists, we
// continue the span, if not we create a new one.
func MakeHTTPTracingMW(tracer opentracing.Tracer, componentName, operationName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var sc, err = tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))

			var span opentracing.Span
			if err != nil {
				span = tracer.StartSpan(operationName)
			} else {
				span = tracer.StartSpan(operationName, opentracing.ChildOf(sc))
			}
			defer span.Finish()

			// Set tags.
			otag.Component.Set(span, componentName)
			span.SetTag("transport", "http")
			otag.SpanKindRPCServer.Set(span)

			next.ServeHTTP(w, r.WithContext(opentracing.ContextWithSpan(r.Context(), span)))
		})
	}
}

type grpcTracingMW struct {
	tracer        opentracing.Tracer
	componentName string
	operationName string
	next          grpc_transport.Handler
}

// MakeGRPCTracingMW makes a tracing middleware at transport level.
func MakeGRPCTracingMW(tracer opentracing.Tracer, componentName, operationName string) func(grpc_transport.Handler) grpc_transport.Handler {
	return func(next grpc_transport.Handler) grpc_transport.Handler {
		return &grpcTracingMW{
			tracer:        tracer,
			componentName: componentName,
			operationName: operationName,
			next:          next,
		}
	}
}

// ServeGRPC try to extract an existing span from the GRPC metadata. It it exists, we
// continue the span, if not we create a new one.
func (m *grpcTracingMW) ServeGRPC(ctx context.Context, req interface{}) (context.Context, interface{}, error) {
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
	otag.Component.Set(span, m.componentName)
	span.SetTag("transport", "grpc")
	otag.SpanKindRPCServer.Set(span)

	return m.next.ServeGRPC(opentracing.ContextWithSpan(ctx, span), req)
}

// MakeEndpointTracingMW makes a tracing middleware at endpoint level.
func MakeEndpointTracingMW(tracer opentracing.Tracer, operationName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if span := opentracing.SpanFromContext(ctx); span != nil {
				span = tracer.StartSpan(operationName, opentracing.ChildOf(span.Context()))
				defer span.Finish()

				var reply, err = next(opentracing.ContextWithSpan(ctx, span), req)

				// If there is no correlation ID, use the newly generated ID.
				var corrID = ctx.Value("correlation_id")
				if corrID == nil {
					if rep := reply.(*fb.FlakiReply); rep != nil {
						corrID = string(rep.Id())
					} else {
						corrID = ""
					}
				}

				span.SetTag("correlation_id", corrID.(string))
				return reply, err
			}
			return next(ctx, req)
		}
	}
}

// Tracing middleware at component level.
type componentTracingMW struct {
	tracer opentracing.Tracer
	next   Component
}

// MakeComponentTracingMW makes a tracing middleware at component level.
func MakeComponentTracingMW(tracer opentracing.Tracer) func(Component) Component {
	return func(next Component) Component {
		return &componentTracingMW{
			tracer: tracer,
			next:   next,
		}
	}
}

// componentTracingMW implements Component.
func (m *componentTracingMW) NextID(ctx context.Context, req *fb.FlakiRequest) (*fb.FlakiReply, error) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span = m.tracer.StartSpan("nextid_component", opentracing.ChildOf(span.Context()))
		defer span.Finish()

		var reply, err = m.next.NextID(opentracing.ContextWithSpan(ctx, span), req)

		// If there is no correlation ID, use the newly generated ID.
		var corrID = ctx.Value("correlation_id")
		if corrID == nil {
			if reply != nil {
				corrID = string(reply.Id())
			} else {
				corrID = ""
			}
		}
		span.SetTag("correlation_id", corrID.(string))

		return reply, err
	}

	return m.next.NextID(ctx, req)
}

// componentTracingMW implements Component.
func (m *componentTracingMW) NextValidID(ctx context.Context, req *fb.FlakiRequest) *fb.FlakiReply {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span = m.tracer.StartSpan("nextvalidid_component", opentracing.ChildOf(span.Context()))
		defer span.Finish()

		var reply = m.next.NextValidID(opentracing.ContextWithSpan(ctx, span), req)

		// If there is no correlation ID, use the newly generated ID.
		var corrID = ctx.Value("correlation_id")
		if corrID == nil {
			corrID = string(reply.Id())
		}
		span.SetTag("correlation_id", corrID.(string))

		return reply
	}

	return m.next.NextValidID(ctx, req)
}

// Tracing middleware at module level.
type moduleTracingMW struct {
	tracer opentracing.Tracer
	next   Module
}

// MakeModuleTracingMW makes a tracing middleware at module level.
func MakeModuleTracingMW(tracer opentracing.Tracer) func(Module) Module {
	return func(next Module) Module {
		return &moduleTracingMW{
			tracer: tracer,
			next:   next,
		}
	}
}

// moduleTracingMW implements Module.
func (m *moduleTracingMW) NextID(ctx context.Context) (string, error) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span = m.tracer.StartSpan("nextid_module", opentracing.ChildOf(span.Context()))
		defer span.Finish()

		var id, err = m.next.NextID(opentracing.ContextWithSpan(ctx, span))

		// If there is no correlation ID, use the newly generated ID.
		var corrID = ctx.Value("correlation_id")
		if corrID == nil {
			corrID = id
		}
		span.SetTag("correlation_id", corrID.(string))

		return id, err
	}

	return m.next.NextID(ctx)
}

// moduleTracingMW implements Module.
func (m *moduleTracingMW) NextValidID(ctx context.Context) string {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span = m.tracer.StartSpan("nextvalidid_module", opentracing.ChildOf(span.Context()))
		defer span.Finish()

		var id = m.next.NextValidID(opentracing.ContextWithSpan(ctx, span))

		// If there is no correlation ID, use the newly generated ID.
		var corrID = ctx.Value("correlation_id")
		if corrID == nil {
			corrID = id
		}
		span.SetTag("correlation_id", corrID.(string))

		return id
	}

	return m.next.NextValidID(ctx)
}
