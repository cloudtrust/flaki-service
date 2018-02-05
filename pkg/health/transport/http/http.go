package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	health "github.com/cloudtrust/flaki-service/pkg/health/endpoint"
	"github.com/go-kit/kit/endpoint"
	http_transport "github.com/go-kit/kit/transport/http"
)

type HealthReports struct {
	Reports []HealthReport `json:"health checks"`
}

type HealthReport struct {
	Name     string `json:"name"`
	Duration string `json:"duration"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

func MakeHealthChecksHandler(es *health.Endpoints) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var report = map[string]string{}

		// Make all tests
		var influxReport health.HealthReports
		{
			var err error
			influxReport, err = es.InfluxHealthChecks(context.Background())
			if err != nil {
				report["influx"] = "KO"
			} else {
				report["influx"] = reportsStatus(influxReport)
			}
		}
		var jaegerReport health.HealthReports
		{
			var err error
			jaegerReport, err = es.JaegerHealthChecks(context.Background())
			if err != nil {
				report["jaeger"] = "KO"
			} else {
				report["jaeger"] = reportsStatus(jaegerReport)
			}
		}
		var redisReport health.HealthReports
		{
			var err error
			redisReport, err = es.RedisHealthChecks(context.Background())
			if err != nil {
				report["redis"] = "KO"
			} else {
				report["redis"] = reportsStatus(redisReport)
			}
		}
		var sentryReport health.HealthReports
		{
			var err error
			sentryReport, err = es.SentryHealthChecks(context.Background())
			if err != nil {
				report["sentry"] = "KO"
			} else {
				report["sentry"] = reportsStatus(sentryReport)
			}
		}

		// Write report.
		var j, err = json.MarshalIndent(report, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(j)
		}
	}
}

func reportsStatus(reports health.HealthReports) string {
	for _, r := range reports.Reports {
		if r.Status != "OK" {
			return "KO"
		}
	}
	return "OK"
}

// MakeInfluxHealthCheckHandler makes a HTTP handler for the Influx HealthCheck endpoint.
func MakeInfluxHealthCheckHandler(e endpoint.Endpoint) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeHealthCheckRequest,
		encodeHealthCheckReply,
		http_transport.ServerErrorEncoder(healthCheckErrorHandler),
	)
}

// MakeJaegerHealthCheckHandler makes a HTTP handler for the Jaeger HealthCheck endpoint.
func MakeJaegerHealthCheckHandler(e endpoint.Endpoint) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeHealthCheckRequest,
		encodeHealthCheckReply,
		http_transport.ServerErrorEncoder(healthCheckErrorHandler),
	)
}

// MakeRedisHealthCheckHandler makes a HTTP handler for the Redis HealthCheck endpoint.
func MakeRedisHealthCheckHandler(e endpoint.Endpoint) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeHealthCheckRequest,
		encodeHealthCheckReply,
		http_transport.ServerErrorEncoder(healthCheckErrorHandler),
	)
}

// MakeSentryHealthCheckHandler makes a HTTP handler for the Sentry HealthCheck endpoint.
func MakeSentryHealthCheckHandler(e endpoint.Endpoint) *http_transport.Server {
	return http_transport.NewServer(e,
		decodeHealthCheckRequest,
		encodeHealthCheckReply,
		http_transport.ServerErrorEncoder(healthCheckErrorHandler),
	)
}

// decodeHealthRequest decodes the health check request.
func decodeHealthCheckRequest(_ context.Context, r *http.Request) (res interface{}, err error) {
	var data []byte

	data, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// encodeHealthCheckReply encodes the health check reply.
func encodeHealthCheckReply(_ context.Context, w http.ResponseWriter, res interface{}) error {
	w.Header().Set("Content-Type", "application/json")

	var reports = res.(health.HealthReports)
	var hr = HealthReports{}
	for _, r := range reports.Reports {
		hr.Reports = append(hr.Reports, HealthReport(r))
	}

	var d, err = json.MarshalIndent(hr, "", "  ")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(d)
	}

	return nil
}

// healthCheckErrorHandler encodes the flatbuffer flaki reply when there is an error.
func healthCheckErrorHandler(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	w.Write([]byte("TODO"))
}
