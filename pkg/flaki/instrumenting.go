package flaki

import (
	"context"

	"github.com/go-kit/kit/metrics"
)

// Instrumenting middleware at module level.
type moduleInstrumentingMW struct {
	counter metrics.Counter
	next    Module
}

// MakeModuleInstrumentingMW makes an instrumenting middleware (at module level) that counts the number
// of IDs generated.
func MakeModuleInstrumentingMW(counter metrics.Counter) func(Module) Module {
	return func(next Module) Module {
		return &moduleInstrumentingMW{
			counter: counter,
			next:    next,
		}
	}
}

// moduleInstrumentingMW implements Module.
func (m *moduleInstrumentingMW) NextID(ctx context.Context) (string, error) {
	var id, err = m.next.NextID(ctx)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = id
	if ctx.Value("correlation_id") != nil {
		corrID = ctx.Value("correlation_id").(string)
	}

	m.counter.With("correlation_id", corrID).Add(1)

	return id, err
}

// moduleInstrumentingMW implements Module.
func (m *moduleInstrumentingMW) NextValidID(ctx context.Context) string {
	var id = m.next.NextValidID(ctx)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = id
	if ctx.Value("correlation_id") != nil {
		corrID = ctx.Value("correlation_id").(string)
	}

	m.counter.With("correlation_id", corrID).Add(1)

	return id
}
