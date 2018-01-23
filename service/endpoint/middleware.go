package endpoint

import (
	"context"
	"time"

	"github.com/cloudtrust/flaki"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	opentracing "github.com/opentracing/opentracing-go"
)

// MakeCorrelationIDMiddleware makes a middleware that adds a correlation ID
// in the context if there is not already one.
func MakeCorrelationIDMiddleware(flaki flaki.Flaki) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var id = ctx.Value("correlation-id")

			if id == nil {
				ctx = context.WithValue(ctx, "correlation-id", flaki.NextValidIDString())
			}
			return next(ctx, req)
		}
	}
}

// MakeLoggingMiddleware makes a logging middleware.
func MakeLoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			defer func(begin time.Time) {
				logger.Log("correlation_id", ctx.Value("correlation-id").(string), "took", time.Since(begin))
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
			defer func(begin time.Time) {
				h.With("correlation-id", ctx.Value("correlation-id").(string)).Observe(time.Since(begin).Seconds())
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

				span.SetTag("correlation-id", ctx.Value("correlation-id").(string))

				ctx = opentracing.ContextWithSpan(ctx, span)
			}
			return next(ctx, request)
		}
	}
}
