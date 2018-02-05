package flaki

import (
	"context"

	"github.com/go-kit/kit/metrics"
)

// Metrics Middleware.
type metricMiddleware struct {
	counter metrics.Counter
	next    Module
}

// MakeModuleMetricMiddleware makes a metric middleware (at module level) that counts the number
// of IDs generated.
func MakeModuleMetricMiddleware(counter metrics.Counter) Middleware {
	return func(next Module) Module {
		return &metricMiddleware{
			counter: counter,
			next:    next,
		}
	}
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
