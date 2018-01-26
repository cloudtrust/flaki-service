package component

import (
	"context"
	"time"

	sentry "github.com/getsentry/raven-go"
	"github.com/go-kit/kit/log"
)

// Middleware on Service.
type Middleware func(Service) Service

// Logging Middleware.
type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

// loggingMiddleware implements Service.
func (m *loggingMiddleware) NextID(ctx context.Context) (string, error) {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextID", "correlation_id", ctx.Value("correlation-id").(string), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextID(ctx)
}

// loggingMiddleware implements Service.
func (m *loggingMiddleware) NextValidID(ctx context.Context) string {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextValidID", "correlation_id", ctx.Value("correlation-id").(string), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextValidID(ctx)
}

// MakeLoggingMiddleware makes a logging middleware.
func MakeLoggingMiddleware(log log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			logger: log,
			next:   next,
		}
	}
}

// Sentry interface.
type Sentry interface {
	CaptureError(err error, tags map[string]string, interfaces ...sentry.Interface) string
	CaptureErrorAndWait(err error, tags map[string]string, interfaces ...sentry.Interface) string
	Close()
}

// Error Middleware.
type errorMiddleware struct {
	client Sentry
	next   Service
}

// errorMiddleware implements Service.
func (s *errorMiddleware) NextID(ctx context.Context) (string, error) {
	var id, err = s.next.NextID(ctx)
	if err != nil {
		s.client.CaptureErrorAndWait(err, map[string]string{"correlation-id": ctx.Value("correlation-id").(string)})
	}
	return id, err
}

// errorMiddleware implements Service.
func (s *errorMiddleware) NextValidID(ctx context.Context) string {
	return s.next.NextValidID(ctx)
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
