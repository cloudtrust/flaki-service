package flaki

//go:generate mockgen -destination=./mock/instrumenting.go -package=mock -mock_names=Counter=Counter,Histogram=Histogram github.com/go-kit/kit/metrics Counter,Histogram

import (
	"context"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
)

// MakeEndpointInstrumentingMW makes an Instrumenting middleware at endpoint level.
func MakeEndpointInstrumentingMW(h metrics.Histogram) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var begin = time.Now()
			var reply, err = next(ctx, req)
			var duration = time.Since(begin)

			// If there is no correlation ID, use the newly generated ID.
			var corrID = ctx.Value("correlation_id")
			if corrID == nil {
				if rep := reply.(*fb.FlakiReply); rep != nil {
					corrID = string(rep.Id())
				} else {
					corrID = ""
				}
			}

			h.With("correlation_id", corrID.(string)).Observe(duration.Seconds())
			return reply, err
		}
	}
}

// Instrumenting middleware at component level.
type componentInstrumentingMW struct {
	histogram metrics.Histogram
	next      Component
}

// MakeComponentInstrumentingMW makes an Instrumenting middleware at component level.
func MakeComponentInstrumentingMW(histogram metrics.Histogram) func(Component) Component {
	return func(next Component) Component {
		return &componentInstrumentingMW{
			histogram: histogram,
			next:      next,
		}
	}
}

// componentInstrumentingMW implements Component.
func (m *componentInstrumentingMW) NextID(ctx context.Context, req *fb.FlakiRequest) (*fb.FlakiReply, error) {
	var begin = time.Now()
	var reply, err = m.next.NextID(ctx, req)
	var duration = time.Since(begin)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value("correlation_id")
	if corrID == nil {
		if reply != nil {
			corrID = string(reply.Id())
		} else {
			corrID = ""
		}
	}

	m.histogram.With("correlation_id", corrID.(string)).Observe(duration.Seconds())
	return reply, err
}

// componentInstrumentingMW implements Component.
func (m *componentInstrumentingMW) NextValidID(ctx context.Context, req *fb.FlakiRequest) *fb.FlakiReply {
	var begin = time.Now()
	var reply = m.next.NextValidID(ctx, req)
	var duration = time.Since(begin)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value("correlation_id")
	if corrID == nil {
		corrID = string(reply.Id())
	}

	m.histogram.With("correlation_id", corrID.(string)).Observe(duration.Seconds())
	return reply
}

// Instrumenting middleware at module level.
type moduleInstrumentingMW struct {
	histogram metrics.Histogram
	next      Module
}

// MakeModuleInstrumentingMW makes an Instrumenting middleware at module level.
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
	var duration = time.Since(begin)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value("correlation_id")
	if corrID == nil {
		corrID = id
	}

	m.histogram.With("correlation_id", corrID.(string)).Observe(duration.Seconds())
	return id, err
}

// moduleInstrumentingMW implements Module.
func (m *moduleInstrumentingMW) NextValidID(ctx context.Context) string {
	var begin = time.Now()
	var id = m.next.NextValidID(ctx)
	var duration = time.Since(begin)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value("correlation_id")
	if corrID == nil {
		corrID = id
	}

	m.histogram.With("correlation_id", corrID.(string)).Observe(duration.Seconds())
	return id
}

// Instrumenting middleware at module level.
type moduleInstrumentingCounterMW struct {
	counter metrics.Counter
	next    Module
}

// MakeModuleInstrumentingCounterMW makes an Instrumenting middleware at module level.
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
