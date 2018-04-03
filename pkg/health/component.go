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

func (s Status) String() string {
	var names = []string{"OK", "KO", "Degraded", "Deactivated"}

	if s < OK || s > Deactivated {
		return "Unknown"
	}

	return names[s]
}

// Component is the health component interface.
type Component interface {
	InfluxHealthChecks(context.Context) []Report
	JaegerHealthChecks(context.Context) []Report
	RedisHealthChecks(context.Context) []Report
	SentryHealthChecks(context.Context) []Report
	AllHealthChecks(context.Context) map[string]string
}

// Report contains the result of one health test.
type Report struct {
	Name     string
	Duration string
	Status   string
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
func (c *component) InfluxHealthChecks(ctx context.Context) []Report {
	var reports = c.influx.HealthChecks(ctx)
	var out = []Report{}
	for _, r := range reports {
		out = append(out, Report{
			Name:     r.Name,
			Duration: r.Duration.String(),
			Status:   r.Status.String(),
			Error:    err(r.Error),
		})
	}
	return out
}

// JaegerHealthChecks uses the health component to test the Jaeger health.
func (c *component) JaegerHealthChecks(ctx context.Context) []Report {
	var reports = c.jaeger.HealthChecks(ctx)
	var out = []Report{}
	for _, r := range reports {
		out = append(out, Report{
			Name:     r.Name,
			Duration: r.Duration.String(),
			Status:   r.Status.String(),
			Error:    err(r.Error),
		})
	}
	return out
}

// RedisHealthChecks uses the health component to test the Redis health.
func (c *component) RedisHealthChecks(ctx context.Context) []Report {
	var reports = c.redis.HealthChecks(ctx)
	var out = []Report{}
	for _, r := range reports {
		out = append(out, Report{
			Name:     r.Name,
			Duration: r.Duration.String(),
			Status:   r.Status.String(),
			Error:    err(r.Error),
		})
	}
	return out
}

// SentryHealthChecks uses the health component to test the Sentry health.
func (c *component) SentryHealthChecks(ctx context.Context) []Report {
	var reports = c.sentry.HealthChecks(ctx)
	var out = []Report{}
	for _, r := range reports {
		out = append(out, Report{
			Name:     r.Name,
			Duration: r.Duration.String(),
			Status:   r.Status.String(),
			Error:    err(r.Error),
		})
	}
	return out
}

// AllChecks call all component checks and build a general health report.
func (c *component) AllHealthChecks(ctx context.Context) map[string]string {
	var reports = map[string]string{}

	reports["influx"] = determineStatus(c.InfluxHealthChecks(ctx))
	reports["jaeger"] = determineStatus(c.JaegerHealthChecks(ctx))
	reports["redis"] = determineStatus(c.RedisHealthChecks(ctx))
	reports["sentry"] = determineStatus(c.SentryHealthChecks(ctx))

	return reports
}

// err return the string error that will be in the health report
func err(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// determineStatus parse all the tests reports and output a global status.
func determineStatus(reports []Report) string {
	var degraded = false
	for _, r := range reports {
		switch r.Status {
		case Deactivated.String():
			// If the status is Deactivated, we do not need to go through all tests reports, all
			// status will be the same.
			return Deactivated.String()
		case KO.String():
			return KO.String()
		case Degraded.String():
			degraded = true
		}
	}
	if degraded {
		return Degraded.String()
	}
	return OK.String()
}
