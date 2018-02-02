package endpoint

import (
	"context"

	health "github.com/cloudtrust/flaki-service/pkg/health/module"
	"github.com/go-kit/kit/endpoint"
)

type HealthReport interface {
	Get() []byte
}

// Service is the interface that the service implements.
type Service interface {
	InfluxHealthChecks(context.Context) (HealthReport, error)
	JaegerHealthChecks(context.Context) (HealthReport, error)
	RedisHealthChecks(context.Context) (HealthReport, error)
	SentryHealthChecks(context.Context) (HealthReport, error)
}

// Endpoints wraps a service behind a set of endpoints.
type Endpoints struct {
	mids                      []endpoint.Middleware
	InfluxHealthCheckEndpoint endpoint.Endpoint
	JaegerHealthCheckEndpoint endpoint.Endpoint
	RedisHealthCheckEndpoint  endpoint.Endpoint
	SentryHealthCheckEndpoint endpoint.Endpoint
}

// NewEndpoints returns Endpoints with the middlware mids. Mids are used to apply middlware
// to all the endpoint in Endpoints.
func NewEndpoints(mids ...endpoint.Middleware) *Endpoints {
	var m = append([]endpoint.Middleware{}, mids...)
	return &Endpoints{
		mids: m,
	}
}

// MakeInfluxHealthCheckEndpoint makes the InfluxHealthCheck endpoint and apply the middelwares mids and Endpoints.mids.
func (es *Endpoints) MakeInfluxHealthCheckEndpoint(s Service, mids ...endpoint.Middleware) *Endpoints {
	var e endpoint.Endpoint = func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.InfluxHealthChecks(ctx)
	}
	e = es.applyMids(e, mids...)
	es.InfluxHealthCheckEndpoint = e
	return es
}

// MakeJaegerHealthCheckEndpoint makes the JaegerHealthCheck endpoint and apply the middelwares mids and Endpoints.mids.
func (es *Endpoints) MakeJaegerHealthCheckEndpoint(s Service, mids ...endpoint.Middleware) *Endpoints {
	var e endpoint.Endpoint = func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.JaegerHealthChecks(ctx)
	}
	e = es.applyMids(e, mids...)
	es.JaegerHealthCheckEndpoint = e
	return es
}

// MakeRedisHealthCheckEndpoint makes the RedisHealthCheck endpoint and apply the middelwares mids and Endpoints.mids.
func (es *Endpoints) MakeRedisHealthCheckEndpoint(s Service, mids ...endpoint.Middleware) *Endpoints {
	var e endpoint.Endpoint = func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.RedisHealthChecks(ctx)
	}
	e = es.applyMids(e, mids...)
	es.RedisHealthCheckEndpoint = e
	return es
}

// MakeSentryHealthCheckEndpoint makes the SentryHealthCheck endpoint and apply the middelwares mids and Endpoints.mids.
func (es *Endpoints) MakeSentryHealthCheckEndpoint(s Service, mids ...endpoint.Middleware) *Endpoints {
	var e endpoint.Endpoint = func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.SentryHealthChecks(ctx)
	}
	e = es.applyMids(e, mids...)
	es.SentryHealthCheckEndpoint = e
	return es
}

// applyMids apply first the middlware mids, then Endpoints.mids to the endpoint.
func (es *Endpoints) applyMids(e endpoint.Endpoint, mids ...endpoint.Middleware) endpoint.Endpoint {
	for _, m := range mids {
		e = m(e)
	}
	for _, m := range es.mids {
		e = m(e)
	}
	return e
}

// Implements Service.
func (es *Endpoints) InfluxHealthChecks(ctx context.Context) ([]health.TestReport, error) {
	var testReport []health.TestReport
	{
		var report interface{}
		var err error
		report, err = es.InfluxHealthCheckEndpoint(ctx, nil)
		if err != nil {
			return []health.TestReport{}, err
		}
		testReport = report.([]health.TestReport)
	}
	return testReport, nil
}

// Implements Service.
func (es *Endpoints) JaegerHealthChecks(ctx context.Context) ([]health.TestReport, error) {
	var testReport []health.TestReport
	{
		var report interface{}
		var err error
		report, err = es.JaegerHealthCheckEndpoint(ctx, nil)
		if err != nil {
			return []health.TestReport{}, err
		}
		testReport = report.([]health.TestReport)
	}
	return testReport, nil
}

// Implements Service.
func (es *Endpoints) RedisHealthChecks(ctx context.Context) ([]health.TestReport, error) {
	var testReport []health.TestReport
	{
		var report interface{}
		var err error
		report, err = es.RedisHealthCheckEndpoint(ctx, nil)
		if err != nil {
			return []health.TestReport{}, err
		}
		testReport = report.([]health.TestReport)
	}
	return testReport, nil
}

// Implements Service.
func (es *Endpoints) SentryHealthChecks(ctx context.Context) ([]health.TestReport, error) {
	var testReport []health.TestReport
	{
		var report interface{}
		var err error
		report, err = es.SentryHealthCheckEndpoint(ctx, nil)
		if err != nil {
			return []health.TestReport{}, err
		}
		testReport = report.([]health.TestReport)
	}
	return testReport, nil
}
