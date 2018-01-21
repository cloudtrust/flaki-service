package component

import (
	"context"
	"github.com/go-kit/kit/log"
	"time"
)

// Service Middleware declarations
type Middleware func(Service) Service

// Logging Middleware
type loggingMiddleware struct {
	logger log.Logger
	next   Service
}

// loggingMiddleware implements Service
func (m *loggingMiddleware) NextID(ctx context.Context) (uint64, error) {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextID", "took", time.Since(begin))
	}(time.Now())
	return m.next.NextID(ctx)
}

func (m *loggingMiddleware) NextValidID(ctx context.Context) uint64 {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextValidID", "took", time.Since(begin))
	}(time.Now())
	return m.next.NextValidID(ctx)
}

// Logging middleware for backend services.
func MakeLoggingMiddleware(log log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			logger: log,
			next:   next,
		}
	}
}
