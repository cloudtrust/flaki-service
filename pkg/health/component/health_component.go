package healthc

import (
	"context"

	"github.com/cloudtrust/flaki-service/pkg/health/module"
)

type HealthReports struct {
	Reports []HealthReport
}

type HealthReport struct {
	Name     string
	Duration string
	Status   string
	Error    string
}

// HealthService contains the health checks.
type HealthService struct {
	influx *healthm.InfluxHealthModule
	jaeger *healthm.JaegerHealthModule
	redis  *healthm.RedisHealthModule
	sentry *healthm.SentryHealthModule
}

// NewHealthService returns the basic service.
func NewHealthService(influxM *healthm.InfluxHealthModule, jaegerM *healthm.JaegerHealthModule,
	redisM *healthm.RedisHealthModule, sentryM *healthm.SentryHealthModule) *HealthService {
	return &HealthService{
		influx: influxM,
		jaeger: jaegerM,
		redis:  redisM,
		sentry: sentryM,
	}
}

// InfluxHealthChecks uses the health component to test the Influx health.
func (s *HealthService) InfluxHealthChecks(ctx context.Context) HealthReports {
	var reports = s.influx.HealthChecks(ctx)
	var hr = HealthReports{}
	for _, r := range reports {
		hr.Reports = append(hr.Reports, HealthReport(r))
	}
	return hr
}

// JaegerHealthChecks uses the health component to test the Jaeger health.
func (s *HealthService) JaegerHealthChecks(ctx context.Context) HealthReports {
	var reports = s.jaeger.HealthChecks(ctx)
	var hr = HealthReports{}
	for _, r := range reports {
		hr.Reports = append(hr.Reports, HealthReport(r))
	}
	return hr
}

// RedisHealthChecks uses the health component to test the Redis health.
func (s *HealthService) RedisHealthChecks(ctx context.Context) HealthReports {
	var reports = s.redis.HealthChecks(ctx)
	var hr = HealthReports{}
	for _, r := range reports {
		hr.Reports = append(hr.Reports, HealthReport(r))
	}
	return hr
}

// SentryHealthChecks uses the health component to test the Sentry health.
func (s *HealthService) SentryHealthChecks(ctx context.Context) HealthReports {
	var reports = s.sentry.HealthChecks(ctx)
	var hr = HealthReports{}
	for _, r := range reports {
		hr.Reports = append(hr.Reports, HealthReport(r))
	}
	return hr
}
