package health

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type TestReport struct {
	Name     string `json:"name"`
	Duration string `json:"duration"`
	Status   string `json:"status"`
	Error    string `json:"status,omitempty"`
}

// Service is the interface that the service implements.
type Service interface {
	InfluxHealthChecks(context.Context) ([]TestReport, error)
	JaegerHealthChecks(context.Context) ([]TestReport, error)
	RedisHealthChecks(context.Context) ([]TestReport, error)
	SentryHealthChecks(context.Context) ([]TestReport, error)
}

type Influx interface {
	Ping(timeout time.Duration) (time.Duration, string, error)
}

type Jaeger interface {
	//Ping(timeout time.Duration) (time.Duration, string, error)
}

type Redis interface {
	Do(cmd string, args ...interface{}) (interface{}, error)
}

type Sentry interface {
	URL() string
}

type healthService struct {
	influx Influx
	redis  Redis
	jaeger Jaeger
	sentry Sentry
}

// NewHealthService returns the health service.
func NewHealthService(influx Influx, redis Redis, jaeger Jaeger, sentry Sentry) Service {
	return &healthService{
		influx: influx,
		redis:  redis,
		jaeger: jaeger,
		sentry: sentry,
	}
}

// InfluxHealthChecks executes all health checks for Influx.
func (s *healthService) InfluxHealthChecks(context.Context) ([]TestReport, error) {
	var reports = []TestReport{}
	reports = append(reports, influxPingCheck(s.influx))
	return reports, nil
}

func influxPingCheck(influx Influx) TestReport {
	var d, _, err = influx.Ping(time.Duration(5 * time.Second))
	var status = "OK"
	if err != nil {
		status = "KO"
	}

	return TestReport{
		Name:     "ping",
		Duration: d.String(),
		Status:   status,
	}
}

// JaegerHealthChecks executes all health checks for Jaeger.
func (s *healthService) JaegerHealthChecks(context.Context) ([]TestReport, error) {
	var reports = []TestReport{}
	reports = append(reports, jaegerPingCheck(s.jaeger))
	return reports, nil
}

func jaegerPingCheck(jaeger Jaeger) TestReport {
	var duration = time.Duration(1 * time.Second)
	var status = "Not yet implemented"
	return TestReport{
		Name:     "ping",
		Duration: duration.String(),
		Status:   status,
	}
}

// RedisHealthChecks executes all health checks for Redis.
func (s *healthService) RedisHealthChecks(context.Context) ([]TestReport, error) {
	var reports = []TestReport{}
	reports = append(reports, redisPingCheck(s.redis))
	return reports, nil
}

func redisPingCheck(redis Redis) TestReport {
	var now = time.Now()
	var _, err = redis.Do("PING")
	var duration = time.Since(now)

	var status = "OK"
	if err != nil {
		status = "KO"
	}

	return TestReport{
		Name:     "ping",
		Duration: duration.String(),
		Status:   status,
	}
}

// SentryHealthChecks executes all health checks for Sentry.
func (s *healthService) SentryHealthChecks(context.Context) ([]TestReport, error) {
	var reports = []TestReport{}
	reports = append(reports, sentryPingCheck(s.sentry))
	return reports, nil
}

func sentryPingCheck(sentry Sentry) TestReport {
	// Build sentry health url from sentry dsn. The health url is <sentryURL>/_health
	var dsn = sentry.URL()
	var healthURL string
	if idx := strings.LastIndex(dsn, "/api/"); idx != -1 {
		healthURL = fmt.Sprintf("%s/_health", dsn[:idx])
	}

	// Get Sentry health status.
	var now = time.Now()
	var status = getSentryStatus(healthURL)
	var duration = time.Since(now)

	return TestReport{
		Name:     "ping",
		Duration: duration.String(),
		Status:   status,
	}
}

func getSentryStatus(url string) string {
	// Query sentry health endpoint.
	var res *http.Response
	{
		var err error
		res, err = http.DefaultClient.Get(url)
		if err != nil {
			return "KO"
		}
		if res != nil {
			defer res.Body.Close()
		}
	}

	// Chesk response status.
	if res.StatusCode != http.StatusOK {
		return "KO"
	}

	// Chesk response body. The sentry health endpoint returns "ok" when there is no issue.
	var response []byte
	{
		var err error
		response, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return "KO"
		}
	}

	if strings.Compare(string(response), "ok") == 0 {
		return "OK"
	}

	return "KO"
}
