package health

import (
	"context"

	health_cmp "github.com/cloudtrust/flaki-service/pkg/health/component"
	"github.com/go-kit/kit/endpoint"
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

// Service is the interface that the service implements.
type Service interface {
	InfluxHealthChecks(context.Context) health_cmp.HealthReports
	JaegerHealthChecks(context.Context) health_cmp.HealthReports
	RedisHealthChecks(context.Context) health_cmp.HealthReports
	SentryHealthChecks(context.Context) health_cmp.HealthReports
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
		var reports = s.InfluxHealthChecks(ctx)
		var hr = HealthReports{}
		for _, r := range reports.Reports {
			hr.Reports = append(hr.Reports, HealthReport(r))
		}
		return hr, nil
	}
	e = es.applyMids(e, mids...)
	es.InfluxHealthCheckEndpoint = e
	return es
}

// MakeJaegerHealthCheckEndpoint makes the JaegerHealthCheck endpoint and apply the middelwares mids and Endpoints.mids.
func (es *Endpoints) MakeJaegerHealthCheckEndpoint(s Service, mids ...endpoint.Middleware) *Endpoints {
	var e endpoint.Endpoint = func(ctx context.Context, req interface{}) (interface{}, error) {
		var reports = s.JaegerHealthChecks(ctx)
		var hr = HealthReports{}
		for _, r := range reports.Reports {
			hr.Reports = append(hr.Reports, HealthReport(r))
		}
		return hr, nil
	}
	e = es.applyMids(e, mids...)
	es.JaegerHealthCheckEndpoint = e
	return es
}

// MakeRedisHealthCheckEndpoint makes the RedisHealthCheck endpoint and apply the middelwares mids and Endpoints.mids.
func (es *Endpoints) MakeRedisHealthCheckEndpoint(s Service, mids ...endpoint.Middleware) *Endpoints {
	var e endpoint.Endpoint = func(ctx context.Context, req interface{}) (interface{}, error) {
		var reports = s.RedisHealthChecks(ctx)
		var hr = HealthReports{}
		for _, r := range reports.Reports {
			hr.Reports = append(hr.Reports, HealthReport(r))
		}
		return hr, nil
	}
	e = es.applyMids(e, mids...)
	es.RedisHealthCheckEndpoint = e
	return es
}

// MakeSentryHealthCheckEndpoint makes the SentryHealthCheck endpoint and apply the middelwares mids and Endpoints.mids.
func (es *Endpoints) MakeSentryHealthCheckEndpoint(s Service, mids ...endpoint.Middleware) *Endpoints {
	var e endpoint.Endpoint = func(ctx context.Context, req interface{}) (interface{}, error) {
		var reports = s.SentryHealthChecks(ctx)
		var hr = HealthReports{}
		for _, r := range reports.Reports {
			hr.Reports = append(hr.Reports, HealthReport(r))
		}
		return hr, nil
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
func (es *Endpoints) InfluxHealthChecks(ctx context.Context) (HealthReports, error) {
	var reports HealthReports
	{
		var report interface{}
		var err error
		report, err = es.InfluxHealthCheckEndpoint(ctx, nil)
		if err != nil {
			return HealthReports{}, err
		}
		reports = report.(HealthReports)
	}
	return reports, nil
}

// Implements Service.
func (es *Endpoints) JaegerHealthChecks(ctx context.Context) (HealthReports, error) {
	var reports HealthReports
	{
		var report interface{}
		var err error
		report, err = es.JaegerHealthCheckEndpoint(ctx, nil)
		if err != nil {
			return HealthReports{}, err
		}
		reports = report.(HealthReports)
	}
	return reports, nil
}

// Implements Service.
func (es *Endpoints) RedisHealthChecks(ctx context.Context) (HealthReports, error) {
	var reports HealthReports
	{
		var report interface{}
		var err error
		report, err = es.RedisHealthCheckEndpoint(ctx, nil)
		if err != nil {
			return HealthReports{}, err
		}
		reports = report.(HealthReports)
	}
	return reports, nil
}

// Implements Service.
func (es *Endpoints) SentryHealthChecks(ctx context.Context) (HealthReports, error) {
	var reports HealthReports
	{
		var report interface{}
		var err error
		report, err = es.SentryHealthCheckEndpoint(ctx, nil)
		if err != nil {
			return HealthReports{}, err
		}
		reports = report.(HealthReports)
	}
	return reports, nil
}
