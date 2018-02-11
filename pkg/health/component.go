package health

import (
	"context"
)

type HealthStatus int

const (
	OK HealthStatus = iota + 1
	KO
	Degraded
	Deactivated
)

func (s HealthStatus) String() string {
	var names = []string{"ok", "ko", "degraded", "deactivated"}

	if s < OK || s > Deactivated {
		return "nknown"
	}

	return names[s]
}

// Component is the health component interface.
type Component interface {
	InfluxHealthChecks(context.Context) HealthReports
	JaegerHealthChecks(context.Context) HealthReports
	RedisHealthChecks(context.Context) HealthReports
	SentryHealthChecks(context.Context) HealthReports
}

// Component is the Health component.
type component struct {
	influx InfluxModule
	jaeger JaegerModule
	redis  RedisModule
	sentry SentryModule
}

// NewComponent returns the health component.
func NewComponent(influx InfluxModule, jaeger JaegerModule, redis RedisModule, sentry SentryModule) Component {
	return &component{
		influx: influx,
		jaeger: jaeger,
		redis:  redis,
		sentry: sentry,
	}
}

// InfluxHealthChecks uses the health component to test the Influx health.
func (c *component) InfluxHealthChecks(ctx context.Context) HealthReports {
	var reports = c.influx.HealthChecks(ctx)
	var hr = HealthReports{}
	for _, r := range reports {
		hr.Reports = append(hr.Reports, HealthReport(r))
	}
	return hr
}

// JaegerHealthChecks uses the health component to test the Jaeger health.
func (c *component) JaegerHealthChecks(ctx context.Context) HealthReports {
	var reports = c.jaeger.HealthChecks(ctx)
	var hr = HealthReports{}
	for _, r := range reports {
		hr.Reports = append(hr.Reports, HealthReport(r))
	}
	return hr
}

// RedisHealthChecks uses the health component to test the Redis health.
func (c *component) RedisHealthChecks(ctx context.Context) HealthReports {
	var reports = c.redis.HealthChecks(ctx)
	var hr = HealthReports{}
	for _, r := range reports {
		hr.Reports = append(hr.Reports, HealthReport(r))
	}
	return hr
}

// SentryHealthChecks uses the health component to test the Sentry health.
func (c *component) SentryHealthChecks(ctx context.Context) HealthReports {
	var reports = c.sentry.HealthChecks(ctx)
	var hr = HealthReports{}
	for _, r := range reports {
		hr.Reports = append(hr.Reports, HealthReport(r))
	}
	return hr
}
