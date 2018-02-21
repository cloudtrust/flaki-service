package flaki

//go:generate mockgen -destination=./mock/logging.go -package=mock -mock_names=Logger=Logger github.com/go-kit/kit/log Logger

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
	var begin = time.Now()
	var id, err = m.next.NextID(ctx)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value("correlation_id")
	if corrID == nil {
		corrID = id
	}

	m.logger.Log("unit", "NextID", "correlation_id", corrID.(string), "took", time.Since(begin))

	return id, err
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) NextValidID(ctx context.Context) string {
	var begin = time.Now()
	var id = m.next.NextValidID(ctx)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value("correlation_id")
	if corrID == nil {
		corrID = id
	}

	m.logger.Log("unit", "NextValidID", "correlation_id", corrID.(string), "took", time.Since(begin))

	return id
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
	var begin = time.Now()
	var id, err = m.next.NextID(ctx)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value("correlation_id")
	if corrID == nil {
		corrID = id
	}

	m.logger.Log("unit", "NextID", "correlation_id", corrID.(string), "took", time.Since(begin))

	return id, err
}

// moduleLoggingMW implements Module.
func (m *moduleLoggingMW) NextValidID(ctx context.Context) string {
	var begin = time.Now()
	var id = m.next.NextValidID(ctx)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value("correlation_id")
	if corrID == nil {
		corrID = id
	}

	m.logger.Log("unit", "NextValidID", "correlation_id", corrID.(string), "took", time.Since(begin))

	return id
}
