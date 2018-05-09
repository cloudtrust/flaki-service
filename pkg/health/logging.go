package health


import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	common "github.com/cloudtrust/common-healthcheck"
)

// MakeEndpointLoggingMW makes a logging middleware.
func MakeEndpointLoggingMW(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			defer func(begin time.Time) {
				logger.Log("correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
			}(time.Now())

			return next(ctx, req)
		}
	}
}

// Logging middleware at component level.
type componentLoggingMW struct {
	logger log.Logger
	next   HealthChecker
}

// MakeComponentLoggingMW makes a logging middleware at component level.
func MakeComponentLoggingMW(logger log.Logger) func(HealthChecker) HealthChecker {
	return func(next HealthChecker) HealthChecker {
		return &componentLoggingMW{
			logger: logger,
			next:   next,
		}
	}
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) ExecInfluxHealthChecks(ctx context.Context) []Report {
	defer func(begin time.Time) {
		m.logger.Log("unit", "ExecInfluxHealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.ExecInfluxHealthChecks(ctx)
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) ReadInfluxHealthChecks(ctx context.Context) []Report {
	defer func(begin time.Time) {
		m.logger.Log("unit", "ReadInfluxHealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.ReadInfluxHealthChecks(ctx)
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) ExecJaegerHealthChecks(ctx context.Context) []Report {
	defer func(begin time.Time) {
		m.logger.Log("unit", "ExecJaegerHealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.ExecJaegerHealthChecks(ctx)
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) ReadJaegerHealthChecks(ctx context.Context) []Report {
	defer func(begin time.Time) {
		m.logger.Log("unit", "ReadJaegerHealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.ReadJaegerHealthChecks(ctx)
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) ExecRedisHealthChecks(ctx context.Context) []Report {
	defer func(begin time.Time) {
		m.logger.Log("unit", "ExecRedisHealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.ExecRedisHealthChecks(ctx)
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) ReadRedisHealthChecks(ctx context.Context) []Report {
	defer func(begin time.Time) {
		m.logger.Log("unit", "ReadRedisHealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.ReadRedisHealthChecks(ctx)
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) ExecSentryHealthChecks(ctx context.Context) []Report {
	defer func(begin time.Time) {
		m.logger.Log("unit", "ExecSentryHealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.ExecSentryHealthChecks(ctx)
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) ReadSentryHealthChecks(ctx context.Context) []Report {
	defer func(begin time.Time) {
		m.logger.Log("unit", "ReadSentryHealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.ReadSentryHealthChecks(ctx)
}

// componentLoggingMW implements Component.
func (m *componentLoggingMW) AllHealthChecks(ctx context.Context) map[string]string {
	defer func(begin time.Time) {
		m.logger.Log("unit", "AllHealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.AllHealthChecks(ctx)
}

// Logging middleware at module level.
type influxModuleLoggingMW struct {
	logger log.Logger
	next   InfluxHealthChecker
}

// MakeInfluxModuleLoggingMW makes a logging middleware at module level.
func MakeInfluxModuleLoggingMW(logger log.Logger) func(InfluxHealthChecker) InfluxHealthChecker {
	return func(next InfluxHealthChecker) InfluxHealthChecker {
		return &influxModuleLoggingMW{
			logger: logger,
			next:   next,
		}
	}
}

// influxModuleLoggingMW implements Module.
func (m *influxModuleLoggingMW) HealthChecks(ctx context.Context) []common.InfluxReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}

// Logging middleware at module level.
type jaegerModuleLoggingMW struct {
	logger log.Logger
	next   JaegerHealthChecker
}

// MakeJaegerModuleLoggingMW makes a logging middleware at module level.
func MakeJaegerModuleLoggingMW(logger log.Logger) func(JaegerHealthChecker) JaegerHealthChecker {
	return func(next JaegerHealthChecker) JaegerHealthChecker {
		return &jaegerModuleLoggingMW{
			logger: logger,
			next:   next,
		}
	}
}

// jaegerModuleLoggingMW implements Module.
func (m *jaegerModuleLoggingMW) HealthChecks(ctx context.Context) []common.JaegerReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}

// Logging middleware at module level.
type redisModuleLoggingMW struct {
	logger log.Logger
	next   RedisHealthChecker
}

// MakeRedisModuleLoggingMW makes a logging middleware at module level.
func MakeRedisModuleLoggingMW(logger log.Logger) func(RedisHealthChecker) RedisHealthChecker {
	return func(next RedisHealthChecker) RedisHealthChecker {
		return &redisModuleLoggingMW{
			logger: logger,
			next:   next,
		}
	}
}

// redisModuleLoggingMW implements Module.
func (m *redisModuleLoggingMW) HealthChecks(ctx context.Context) []common.RedisReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}

// Logging middleware at module level.
type sentryModuleLoggingMW struct {
	logger log.Logger
	next   SentryHealthChecker
}

// MakeSentryModuleLoggingMW makes a logging middleware at module level.
func MakeSentryModuleLoggingMW(logger log.Logger) func(SentryHealthChecker) SentryHealthChecker {
	return func(next SentryHealthChecker) SentryHealthChecker {
		return &sentryModuleLoggingMW{
			logger: logger,
			next:   next,
		}
	}
}

// sentryModuleLoggingMW implements Module.
func (m *sentryModuleLoggingMW) HealthChecks(ctx context.Context) []common.SentryReport {
	defer func(begin time.Time) {
		m.logger.Log("unit", "HealthChecks", "correlation_id", ctx.Value("correlation_id").(string), "took", time.Since(begin))
	}(time.Now())

	return m.next.HealthChecks(ctx)
}
