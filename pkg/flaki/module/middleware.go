package flaki

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	opentracing "github.com/opentracing/opentracing-go"
)

// Middleware on Service.
type Middleware func(Service) Service

// MakeLoggingMiddleware makes a logging middleware.
func MakeLoggingMiddleware(log log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			logger: log,
			next:   next,
		}
	}
}

// Logging Middleware.
type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

// loggingMiddleware implements Service.
func (m *loggingMiddleware) NextID(ctx context.Context) (string, error) {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextID", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextID(ctx)
}

// loggingMiddleware implements Service.
func (m *loggingMiddleware) NextValidID(ctx context.Context) string {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextValidID", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextValidID(ctx)
}

// MakeMetricMiddleware makes a metric middleware that counts the number
// of IDs generated.
func MakeMetricMiddleware(counter metrics.Counter) Middleware {
	return func(next Service) Service {
		return &metricMiddleware{
			counter: counter,
			next:    next,
		}
	}
}

// Metrics Middleware.
type metricMiddleware struct {
	counter metrics.Counter
	next    Service
}

// metricMiddleware implements Service.
func (m *metricMiddleware) NextID(ctx context.Context) (string, error) {
	m.counter.With("correlation_id", ctx.Value("correlation_id").(string)).Add(1)
	return m.next.NextID(ctx)
}

// metricMiddleware implements Service.
func (m *metricMiddleware) NextValidID(ctx context.Context) string {
	m.counter.With("correlation_id", ctx.Value("correlation_id").(string)).Add(1)
	return m.next.NextValidID(ctx)
}

// MakeTracingMiddleware makes a tracing middleware.
func MakeTracingMiddleware(tracer opentracing.Tracer) Middleware {
	return func(next Service) Service {
		return &tracingMiddleware{
			tracer: tracer,
			next:   next,
		}
	}
}

// Tracing Middleware.
type tracingMiddleware struct {
	tracer opentracing.Tracer
	next   Service
}

// tracingMiddleware implements Service.
func (m *tracingMiddleware) NextID(ctx context.Context) (string, error) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span = m.tracer.StartSpan("nextid_module", opentracing.ChildOf(span.Context()))
		defer span.Finish()
		span.SetTag("correlation_id", ctx.Value("correlation_id").(string))

		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	return m.next.NextID(ctx)
}

// tracingMiddleware implements Service.
func (m *tracingMiddleware) NextValidID(ctx context.Context) string {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span = m.tracer.StartSpan("nextvalidid_module", opentracing.ChildOf(span.Context()))
		defer span.Finish()
		span.SetTag("correlation_id", ctx.Value("correlation_id").(string))

		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	return m.next.NextValidID(ctx)
}
