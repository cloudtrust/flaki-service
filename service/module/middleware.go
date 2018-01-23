package flaki

import (
	"context"
	"time"

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
