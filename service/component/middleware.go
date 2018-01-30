package component

import (
	"context"
	"time"

	sentry "github.com/getsentry/raven-go"
	"github.com/go-kit/kit/log"
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
		span = m.tracer.StartSpan("nextid_component", opentracing.ChildOf(span.Context()))
		defer span.Finish()
		span.SetTag("correlation_id", ctx.Value("correlation_id").(string))

		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	return m.next.NextID(ctx)
}

// tracingMiddleware implements Service.
func (m *tracingMiddleware) NextValidID(ctx context.Context) string {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span = m.tracer.StartSpan("nextvalidid_component", opentracing.ChildOf(span.Context()))
		defer span.Finish()
		span.SetTag("correlation_id", ctx.Value("correlation_id").(string))

		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	return m.next.NextValidID(ctx)
}

// Sentry interface.
type Sentry interface {
	CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string
}

// MakeErrorMiddleware makes an error handling middleware, where the errors are sent to Sentry.
func MakeErrorMiddleware(client Sentry) Middleware {
	return func(next Service) Service {
		return &errorMiddleware{
			client: client,
			next:   next,
		}
	}
}

// Error Middleware.
type errorMiddleware struct {
	client Sentry
	next   Service
}

// errorMiddleware implements Service.
func (m *errorMiddleware) NextID(ctx context.Context) (string, error) {
	var id, err = m.next.NextID(ctx)
	if err != nil {
		m.client.CaptureError(err, map[string]string{"correlation_id": ctx.Value("correlation_id").(string)})
	}
	return id, err
}

// errorMiddleware implements Service.
func (m *errorMiddleware) NextValidID(ctx context.Context) string {
	return m.next.NextValidID(ctx)
}
