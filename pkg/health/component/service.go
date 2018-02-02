package component

import (
	"context"

	health "github.com/cloudtrust/flaki-service/pkg/health/module"
)

type TestReport struct {
	Name     string `json:"name"`
	Duration string `json:"duration"`
	Status   string `json:"status"`
	Error    string `json:"status,omitempty"`
}
type HealthReport struct {
	reports []TestReport
}

// Service is the interface that the service implements.
type Service interface {
	InfluxHealthChecks(context.Context) (HealthReport, error)
	JaegerHealthChecks(context.Context) (HealthReport, error)
	RedisHealthChecks(context.Context) (HealthReport, error)
	SentryHealthChecks(context.Context) (HealthReport, error)
}
type InfluxService interface {
	InfluxHealthChecks(context.Context) (HealthReport, error)
}

// healthService contains the health checks.
type healthService struct {
	module health.Service
}

// NewHealthService returns the basic service.
func NewHealthService(healthModule health.Service) Service {
	return &healthService{
		module: healthModule,
	}
}

// InfluxHealthChecks uses the health component to test the Influx health.
func (s *healthService) InfluxHealthChecks(ctx context.Context) ([]health.TestReport, error) {
	return s.module.InfluxHealthChecks(ctx)
}

// JaegerHealthChecks uses the health component to test the Jaeger health.
func (s *healthService) JaegerHealthChecks(ctx context.Context) ([]health.TestReport, error) {
	return s.module.JaegerHealthChecks(ctx)
}

// RedisHealthChecks uses the health component to test the Redis health.
func (s *healthService) RedisHealthChecks(ctx context.Context) ([]health.TestReport, error) {
	return s.module.RedisHealthChecks(ctx)
}

// SentryHealthChecks uses the health component to test the Sentry health.
func (s *healthService) SentryHealthChecks(ctx context.Context) ([]health.TestReport, error) {
	return s.module.SentryHealthChecks(ctx)
}
