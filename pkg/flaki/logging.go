package flaki

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
)

// Logging middleware at component level.
type componentLoggingMW struct {
	logger log.Logger
	next   Component
}

// MakeComponentLoggingMW makes a logging middleware at component level.
func MakeComponentLoggingMW(log log.Logger) func(Component) Component {
	return func(next Component) Component {
		return &componentLoggingMW{
			logger: log,
			next:   next,
		}
	}
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) NextID(ctx context.Context) (string, error) {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextID", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextID(ctx)
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) NextValidID(ctx context.Context) string {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextValidID", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextValidID(ctx)
}

// Logging middleware at module level.
type moduleLoggingMW struct {
	logger log.Logger
	next   Module
}

// MakeModuleLoggingMW makes a logging middleware at module level.
func MakeModuleLoggingMW(log log.Logger) func(Module) Module {
	return func(next Module) Module {
		return &moduleLoggingMW{
			logger: log,
			next:   next,
		}
	}
}

// moduleLoggingMW implements Module.
func (m *moduleLoggingMW) NextID(ctx context.Context) (string, error) {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextID", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextID(ctx)
}

// moduleLoggingMW implements Module.
func (m *moduleLoggingMW) NextValidID(ctx context.Context) string {
	defer func(begin time.Time) {
		m.logger.Log("method", "NextValidID", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())
	return m.next.NextValidID(ctx)
}
