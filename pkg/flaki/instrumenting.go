package flaki

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
)

// Instrumenting middleware at component level.
type componentInstrumentingMW struct {
	histogram metrics.Histogram
	next      Module
}

// MakeComponentInstrumentingMW makes an instrumenting middleware (at component level) that counts the number
// of IDs generated.
func MakeComponentInstrumentingMW(histogram metrics.Histogram) func(Component) Component {
	return func(next Component) Component {
		return &componentInstrumentingMW{
			histogram: histogram,
			next:      next,
		}
	}
}

// componentInstrumentingMW implements Component.
func (m *componentInstrumentingMW) NextID(ctx context.Context) (string, error) {
	var begin = time.Now()
	var id, err = m.next.NextID(ctx)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value("correlation_id")
	if corrID == nil {
		corrID = id
	}

	m.histogram.With("correlation_id", corrID.(string)).Observe(time.Since(begin).Seconds())
	return id, err
}

// componentInstrumentingMW implements Component.
func (m *componentInstrumentingMW) NextValidID(ctx context.Context) string {
	var begin = time.Now()
	var id = m.next.NextValidID(ctx)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value("correlation_id")
	if corrID == nil {
		corrID = id
	}

	m.histogram.With("correlation_id", corrID.(string)).Observe(time.Since(begin).Seconds())
	return id
}

// Instrumenting middleware at module level.
type moduleInstrumentingMW struct {
	histogram metrics.Histogram
	next      Module
}

// MakeModuleInstrumentingMW makes an instrumenting middleware (at module level) that counts the number
// of IDs generated.
func MakeModuleInstrumentingMW(histogram metrics.Histogram) func(Module) Module {
	return func(next Module) Module {
		return &moduleInstrumentingMW{
			histogram: histogram,
			next:      next,
		}
	}
}

// moduleInstrumentingMW implements Module.
func (m *moduleInstrumentingMW) NextID(ctx context.Context) (string, error) {
	var begin = time.Now()
	var id, err = m.next.NextID(ctx)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value("correlation_id")
	if corrID == nil {
		corrID = id
	}

	m.histogram.With("correlation_id", corrID.(string)).Observe(time.Since(begin).Seconds())
	return id, err
}

// moduleInstrumentingMW implements Module.
func (m *moduleInstrumentingMW) NextValidID(ctx context.Context) string {
	var begin = time.Now()
	var id = m.next.NextValidID(ctx)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value("correlation_id")
	if corrID == nil {
		corrID = id
	}

	m.histogram.With("correlation_id", corrID.(string)).Observe(time.Since(begin).Seconds())
	return id
}

// Instrumenting middleware at module level.
type moduleInstrumentingCounterMW struct {
	counter metrics.Counter
	next    Module
}

// MakeModuleInstrumentingCounterMW makes an instrumenting middleware (at module level) that counts the number
// of IDs generated.
func MakeModuleInstrumentingCounterMW(counter metrics.Counter) func(Module) Module {
	return func(next Module) Module {
		return &moduleInstrumentingCounterMW{
			counter: counter,
			next:    next,
		}
	}
}

// moduleInstrumentingCounterMW implements Module.
func (m *moduleInstrumentingCounterMW) NextID(ctx context.Context) (string, error) {
	var id, err = m.next.NextID(ctx)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value("correlation_id")
	if corrID == nil {
		corrID = id
	}

	m.counter.With("correlation_id", corrID.(string)).Add(1)
	return id, err
}

// moduleInstrumentingCounterMW implements Module.
func (m *moduleInstrumentingCounterMW) NextValidID(ctx context.Context) string {
	var id = m.next.NextValidID(ctx)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value("correlation_id")
	if corrID == nil {
		corrID = id
	}

	m.counter.With("correlation_id", corrID.(string)).Add(1)
	return id
}
