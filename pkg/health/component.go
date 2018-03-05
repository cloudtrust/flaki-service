package health

//go:generate mockgen -destination=./mock/component.go -package=mock -mock_names=Component=Component github.com/cloudtrust/flaki-service/pkg/health Component

import (
	"context"
)

// Status is the status of the health check.
type Status int

const (
	// OK is the status for a successful health check.
	OK Status = iota
	// KO is the status for an unsuccessful health check.
	KO
	// Degraded is the status for a degraded service, e.g. the service still works, but the metrics DB is KO.
	Degraded
	// Deactivated is the status for a service that is deactivated, e.g. we can disable error tracking, instrumenting, tracing,...
	Deactivated
)

type key int

const (
	// CorrelationIDKey is the key for the correlation ID in the context.
	CorrelationIDKey key = iota
)

func (s Status) String() string {
	var names = []string{"OK", "KO", "Degraded", "Deactivated"}

	if s < OK || s > Deactivated {
		return "Unknown"
	}

	return names[s]
}

// Component is the health component interface.
type Component interface {
	InfluxHealthChecks(context.Context) Reports
	JaegerHealthChecks(context.Context) Reports
	RedisHealthChecks(context.Context) Reports
	SentryHealthChecks(context.Context) Reports
}

// Reports contains the results of all health tests for a given module.
type Reports struct {
	Reports []Report
}

// Report contains the result of one health test.
type Report struct {
	Name     string
	Duration string
	Status   Status
	Error    string
}

// component is the Health component.
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
func (c *component) InfluxHealthChecks(ctx context.Context) Reports {
	var reports = c.influx.HealthChecks(ctx)
	var hr = Reports{}
	for _, r := range reports {
		hr.Reports = append(hr.Reports, Report(r))
	}
	return hr
}

// JaegerHealthChecks uses the health component to test the Jaeger health.
func (c *component) JaegerHealthChecks(ctx context.Context) Reports {
	var reports = c.jaeger.HealthChecks(ctx)
	var hr = Reports{}
	for _, r := range reports {
		hr.Reports = append(hr.Reports, Report(r))
	}
	return hr
}

// RedisHealthChecks uses the health component to test the Redis health.
func (c *component) RedisHealthChecks(ctx context.Context) Reports {
	var reports = c.redis.HealthChecks(ctx)
	var hr = Reports{}
	for _, r := range reports {
		hr.Reports = append(hr.Reports, Report(r))
	}
	return hr
}

// SentryHealthChecks uses the health component to test the Sentry health.
func (c *component) SentryHealthChecks(ctx context.Context) Reports {
	var reports = c.sentry.HealthChecks(ctx)
	var hr = Reports{}
	for _, r := range reports {
		hr.Reports = append(hr.Reports, Report(r))
	}
	return hr
}
