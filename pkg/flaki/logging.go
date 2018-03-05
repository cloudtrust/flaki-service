package flaki

//go:generate mockgen -destination=./mock/logging.go -package=mock -mock_names=Logger=Logger github.com/go-kit/kit/log Logger

import (
	"context"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

const (
	// LoggingCorrelationIDKey is the key for the correlation ID in the trace.
	LoggingCorrelationIDKey = "correlation_id"
)

// MakeEndpointLoggingMW makes a logging middleware.
func MakeEndpointLoggingMW(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var begin = time.Now()
			var reply, err = next(ctx, req)
			var duration = time.Since(begin)

			// If there is no correlation ID, use the newly generated ID.
			var corrID = ctx.Value(CorrelationIDKey)
			if corrID == nil {
				if rep := reply.(*fb.FlakiReply); rep != nil {
					corrID = string(rep.Id())
				} else {
					corrID = ""
				}
			}

			logger.Log(LoggingCorrelationIDKey, corrID.(string), "took", duration)
			return reply, err
		}
	}
}

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
func (m *componentLoggingMW) NextID(ctx context.Context, req *fb.FlakiRequest) (*fb.FlakiReply, error) {
	var begin = time.Now()
	var reply, err = m.next.NextID(ctx, req)
	var duration = time.Since(begin)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value(CorrelationIDKey)
	if corrID == nil {
		if reply != nil {
			corrID = string(reply.Id())
		} else {
			corrID = ""
		}
	}

	m.logger.Log("unit", "NextID", LoggingCorrelationIDKey, corrID.(string), "took", duration)

	return reply, err
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) NextValidID(ctx context.Context, req *fb.FlakiRequest) *fb.FlakiReply {
	var begin = time.Now()
	var reply = m.next.NextValidID(ctx, req)
	var duration = time.Since(begin)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value(CorrelationIDKey)
	if corrID == nil {
		corrID = string(reply.Id())
	}

	m.logger.Log("unit", "NextValidID", LoggingCorrelationIDKey, corrID.(string), "took", duration)

	return reply
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
	var duration = time.Since(begin)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value(CorrelationIDKey)
	if corrID == nil {
		corrID = id
	}

	m.logger.Log("unit", "NextID", LoggingCorrelationIDKey, corrID.(string), "took", duration)

	return id, err
}

// moduleLoggingMW implements Module.
func (m *moduleLoggingMW) NextValidID(ctx context.Context) string {
	var begin = time.Now()
	var id = m.next.NextValidID(ctx)
	var duration = time.Since(begin)

	// If there is no correlation ID, use the newly generated ID.
	var corrID = ctx.Value(CorrelationIDKey)
	if corrID == nil {
		corrID = id
	}

	m.logger.Log("unit", "NextValidID", LoggingCorrelationIDKey, corrID.(string), "took", duration)

	return id
}
