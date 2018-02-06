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
	m.counter.With("correlation_id", ctx.Value("correlation_id").(string)).Add(1)
	return m.next.NextID(ctx)
}

// moduleInstrumentingMW implements Module.
func (m *moduleInstrumentingMW) NextValidID(ctx context.Context) string {
	m.counter.With("correlation_id", ctx.Value("correlation_id").(string)).Add(1)
	return m.next.NextValidID(ctx)
}
