package middleware

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	opentracing "github.com/opentracing/opentracing-go"
)

// Flaki is the interface of the distributed unique IDs generator.
type Flaki interface {
	NextValidIDString() string
}

// MakeEndpointLoggingMW makes a logging middleware.
func MakeEndpointLoggingMW(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var begin = time.Now()
			var id, err = next(ctx, req)

			// If there is no correlation ID, use the newly generated ID.
			var corrID = id.(string)
			if ctx.Value("correlation_id") != nil {
				corrID = ctx.Value("correlation_id").(string)
			}

			logger.Log("correlation_id", corrID, "took", time.Since(begin))
			return id, err
		}
	}
}

// MakeEndpointInstrumentingMW makes a middleware that measure the endpoints response time and
// send the metrics to influx DB.
func MakeEndpointInstrumentingMW(h metrics.Histogram) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var begin = time.Now()
			var id, err = next(ctx, req)

			// If there is no correlation ID, use the newly generated ID.
			var corrID = id.(string)
			if ctx.Value("correlation_id") != nil {
				corrID = ctx.Value("correlation_id").(string)
			}

			h.With("correlation_id", corrID).Observe(time.Since(begin).Seconds())
			return id, err
		}
	}
}

// MakeEndpointTracingMW makes a middleware that handle the tracing with jaeger.
func MakeEndpointTracingMW(tracer opentracing.Tracer, operationName string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if span := opentracing.SpanFromContext(ctx); span != nil {
				span = tracer.StartSpan(operationName, opentracing.ChildOf(span.Context()))
				defer span.Finish()

				// If there is no correlation ID, use the newly generated ID.
				var id, err = next(opentracing.ContextWithSpan(ctx, span), req)
				var corrID = id.(string)
				if ctx.Value("correlation_id") != nil {
					corrID = ctx.Value("correlation_id").(string)
				}

				span.SetTag("correlation_id", corrID)
				return id, err
			}
			return next(ctx, req)
		}
	}
}
