package health

//go:generate mockgen -destination=./mock/logging.go -package=mock -mock_names=Logger=Logger github.com/go-kit/kit/log Logger

import (
	"context"
	"time"

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

			logger.Log(LoggingCorrelationIDKey, ctx.Value(CorrelationIDKey).(string), "took", time.Since(begin))
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
func (m *componentLoggingMW) InfluxHealthChecks(ctx context.Context) Reports {
	defer func(begin time.Time) {
		m.logger.Log("unit", "InfluxHealthChecks", LoggingCorrelationIDKey, ctx.Value(CorrelationIDKey).(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.InfluxHealthChecks(ctx)
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) JaegerHealthChecks(ctx context.Context) Reports {
	defer func(begin time.Time) {
		m.logger.Log("unit", "JaegerHealthChecks", LoggingCorrelationIDKey, ctx.Value(CorrelationIDKey).(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.JaegerHealthChecks(ctx)
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) RedisHealthChecks(ctx context.Context) Reports {
	defer func(begin time.Time) {
		m.logger.Log("unit", "RedisHealthChecks", LoggingCorrelationIDKey, ctx.Value(CorrelationIDKey).(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.RedisHealthChecks(ctx)
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) SentryHealthChecks(ctx context.Context) Reports {
	defer func(begin time.Time) {
		m.logger.Log("unit", "SentryHealthChecks", LoggingCorrelationIDKey, ctx.Value(CorrelationIDKey).(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.SentryHealthChecks(ctx)
}

// Logging middleware at module level.
type influxModuleLoggingMW struct {
	logger log.Logger
	next   InfluxModule
}

// MakeInfluxModuleLoggingMW makes a logging middleware at module level.
func MakeInfluxModuleLoggingMW(log log.Logger) func(InfluxModule) InfluxModule {
	return func(next InfluxModule) InfluxModule {
		return &influxModuleLoggingMW{
			logger: log,
			next:   next,
		}
	}
}

// influxModuleLoggingMW implements Module.
func (m *influxModuleLoggingMW) HealthChecks(ctx context.Context) []InfluxReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", LoggingCorrelationIDKey, ctx.Value(CorrelationIDKey).(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}

// Logging middleware at module level.
type jaegerModuleLoggingMW struct {
	logger log.Logger
	next   JaegerModule
}

// MakeJaegerModuleLoggingMW makes a logging middleware at module level.
func MakeJaegerModuleLoggingMW(log log.Logger) func(JaegerModule) JaegerModule {
	return func(next JaegerModule) JaegerModule {
		return &jaegerModuleLoggingMW{
			logger: log,
			next:   next,
		}
	}
}

// jaegerModuleLoggingMW implements Module.
func (m *jaegerModuleLoggingMW) HealthChecks(ctx context.Context) []JaegerReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", LoggingCorrelationIDKey, ctx.Value(CorrelationIDKey).(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}

// Logging middleware at module level.
type redisModuleLoggingMW struct {
	logger log.Logger
	next   RedisModule
}

// MakeRedisModuleLoggingMW makes a logging middleware at module level.
func MakeRedisModuleLoggingMW(log log.Logger) func(RedisModule) RedisModule {
	return func(next RedisModule) RedisModule {
		return &redisModuleLoggingMW{
			logger: log,
			next:   next,
		}
	}
}

// redisModuleLoggingMW implements Module.
func (m *redisModuleLoggingMW) HealthChecks(ctx context.Context) []RedisReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", LoggingCorrelationIDKey, ctx.Value(CorrelationIDKey).(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}

// Logging middleware at module level.
type sentryModuleLoggingMW struct {
	logger log.Logger
	next   SentryModule
}

// MakeSentryModuleLoggingMW makes a logging middleware at module level.
func MakeSentryModuleLoggingMW(log log.Logger) func(SentryModule) SentryModule {
	return func(next SentryModule) SentryModule {
		return &sentryModuleLoggingMW{
			logger: log,
			next:   next,
		}
	}
}

// sentryModuleLoggingMW implements Module.
func (m *sentryModuleLoggingMW) HealthChecks(ctx context.Context) []SentryReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", LoggingCorrelationIDKey, ctx.Value(CorrelationIDKey).(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}
