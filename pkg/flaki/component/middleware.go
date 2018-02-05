package flakic

import (
	"context"
	"time"

	sentry "github.com/getsentry/raven-go"
	"github.com/go-kit/kit/log"
	opentracing "github.com/opentracing/opentracing-go"
)

// FlakiComponent is the interface of the flaki Component
type FlakiComponent interface {
	NextID(context.Context) (string, error)
	NextValidID(context.Context) string
}

// Middleware on FlakiComponent.
type Middleware func(FlakiComponent) FlakiComponent

// Logging Middleware.
type loggingMiddleware struct {
	logger log.Logger
	next   FlakiComponent
}

// MakeLoggingMiddleware makes a logging middleware.
func MakeLoggingMiddleware(log log.Logger) Middleware {
	return func(next FlakiComponent) FlakiComponent {
		return &loggingMiddleware{
			logger: log,
			next:   next,
		}
	}
}

// loggingMiddleware implements FlakiComponent.
func (m *loggingMiddleware) NextID(ctx context.Context) (string, error) {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextID", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextID(ctx)
}

// loggingMiddleware implements FlakiComponent.
func (m *loggingMiddleware) NextValidID(ctx context.Context) string {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextValidID", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextValidID(ctx)
}

// Tracing Middleware.
type tracingMiddleware struct {
	tracer opentracing.Tracer
	next   FlakiComponent
}

// MakeTracingMiddleware makes a tracing middleware.
func MakeTracingMiddleware(tracer opentracing.Tracer) Middleware {
	return func(next FlakiComponent) FlakiComponent {
		return &tracingMiddleware{
			tracer: tracer,
			next:   next,
		}
	}
}

// tracingMiddleware implements FlakiComponent.
func (m *tracingMiddleware) NextID(ctx context.Context) (string, error) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span = m.tracer.StartSpan("nextid_component", opentracing.ChildOf(span.Context()))
		defer span.Finish()
		span.SetTag("correlation_id", ctx.Value("correlation_id").(string))

		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	return m.next.NextID(ctx)
}

// tracingMiddleware implements FlakiComponent.
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

// Error Middleware.
type errorMiddleware struct {
	client Sentry
	next   FlakiComponent
}

// MakeErrorMiddleware makes an error handling middleware, where the errors are sent to Sentry.
func MakeErrorMiddleware(client Sentry) Middleware {
	return func(next FlakiComponent) FlakiComponent {
		return &errorMiddleware{
			client: client,
			next:   next,
		}
	}
}

// errorMiddleware implements FlakiComponent.
func (m *errorMiddleware) NextID(ctx context.Context) (string, error) {
	var id, err = m.next.NextID(ctx)
	if err != nil {
		m.client.CaptureError(err, map[string]string{"correlation_id": ctx.Value("correlation_id").(string)})
	}
	return id, err
}

// errorMiddleware implements FlakiComponent.
func (m *errorMiddleware) NextValidID(ctx context.Context) string {
	return m.next.NextValidID(ctx)
}
