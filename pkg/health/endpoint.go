package health

//go:generate mockgen -destination=./mock/component.go -package=mock -mock_names=HealthChecker=HealthChecker github.com/cloudtrust/flaki-service/pkg/health HealthChecker

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints wraps a service behind a set of endpoints.
type Endpoints struct {
	InfluxHealthCheck endpoint.Endpoint
	JaegerHealthCheck endpoint.Endpoint
	RedisHealthCheck  endpoint.Endpoint
	SentryHealthCheck endpoint.Endpoint
	AllHealthChecks   endpoint.Endpoint
}

// HealthChecker is the health component interface.
type HealthChecker interface {
	InfluxHealthChecks(context.Context) []Report
	JaegerHealthChecks(context.Context) []Report
	RedisHealthChecks(context.Context) []Report
	SentryHealthChecks(context.Context) []Report
	AllHealthChecks(context.Context) map[string]string
}

// MakeInfluxHealthCheckEndpoint makes the InfluxHealthCheck endpoint.
func MakeInfluxHealthCheckEndpoint(hc HealthChecker) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return hc.InfluxHealthChecks(ctx), nil
	}
}

// MakeJaegerHealthCheckEndpoint makes the JaegerHealthCheck endpoint.
func MakeJaegerHealthCheckEndpoint(hc HealthChecker) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return hc.JaegerHealthChecks(ctx), nil
	}
}

// MakeRedisHealthCheckEndpoint makes the RedisHealthCheck endpoint.
func MakeRedisHealthCheckEndpoint(hc HealthChecker) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return hc.RedisHealthChecks(ctx), nil
	}
}

// MakeSentryHealthCheckEndpoint makes the SentryHealthCheck endpoint.
func MakeSentryHealthCheckEndpoint(hc HealthChecker) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return hc.SentryHealthChecks(ctx), nil
	}
}

// MakeAllHealthChecksEndpoint makes an endpoint that does all health checks.
func MakeAllHealthChecksEndpoint(hc HealthChecker) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return hc.AllHealthChecks(ctx), nil
	}
}
