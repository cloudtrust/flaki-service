package endpoint

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudtrust/flaki"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	opentracing "github.com/opentracing/opentracing-go"
	tags "github.com/opentracing/opentracing-go/ext"
)

// MakeCorrelationIDMiddleware makes a middleware that adds a correlation id
// in the context if there is not already one.
func MakeCorrelationIDMiddleware(flaki flaki.Flaki) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var idStr = ctx.Value("correlationID")

			if idStr == nil {
				// If there is no correlation ID in the context, add one.
				ctx = context.WithValue(ctx, "correlationID", flaki.NextValidID())
			} else {
				// If there is already a correlation ID in the context, use it.
				var id, err = strconv.ParseUint(idStr.(string), 10, 64)
				if err != nil {
					panic("cannot convert to uint64")
				}
				ctx = context.WithValue(ctx, "correlationID", id)
			}
			return next(ctx, req)
		}
	}
}

func getIDFromContext(ctx context.Context) uint64 {
	var id = ctx.Value("correlationID")
	if id == nil {
		return 0
	}
	return id.(uint64)
}

// MakeLoggingMiddleware makes a logging middleware.
func MakeLoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			defer func(begin time.Time) {
				logger.Log("correlation_id", getIDFromContext(ctx), "took", time.Since(begin))
			}(time.Now())
			return next(ctx, req)
		}
	}
}

// MakeMetricMiddleware makes a middleware that measure the endpoints response time and
// send the metrics to influx DB.
func MakeMetricMiddleware(h metrics.Histogram) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var correlationID = getIDFromContext(ctx)
			defer func(begin time.Time) {
				h.With("correlationID", strconv.FormatUint(correlationID, 10)).Observe(time.Since(begin).Seconds())
			}(time.Now())
			return next(ctx, req)
		}
	}
}

// MakeTracingMiddleware makes a middleware that handle the tracing with jaeger.
func MakeTracingMiddleware(tracer opentracing.Tracer, operationName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			if span := opentracing.SpanFromContext(ctx); span != nil {
				span = tracer.StartSpan(operationName, opentracing.ChildOf(span.Context()))
				defer span.Finish()

				tags.SpanKindRPCServer.Set(span)

				ctx = opentracing.ContextWithSpan(ctx, span)
			}
			return next(ctx, request)
		}
	}
}
