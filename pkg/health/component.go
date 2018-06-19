package health

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	common "github.com/cloudtrust/common-healthcheck"
)

const (
	// Names of the units in the health check http response and in the DB.
	influxUnitName = "influx"
	jaegerUnitName = "jaeger"
	redisUnitName  = "redis"
	sentryUnitName = "sentry"
)

// InfluxHealthChecker is the interface of the influx health check module.
type InfluxHealthChecker interface {
	HealthChecks(context.Context) []common.InfluxReport
}

// JaegerHealthChecker is the interface of the jaeger health check module.
type JaegerHealthChecker interface {
	HealthChecks(context.Context) []common.JaegerReport
}

// RedisHealthChecker is the interface of the redis health check module.
type RedisHealthChecker interface {
	HealthChecks(context.Context) []common.RedisReport
}

// SentryHealthChecker is the interface of the sentry health check module.
type SentryHealthChecker interface {
	HealthChecks(context.Context) []common.SentryReport
}

// StoreModule is the interface of the module that stores the health reports
// in the DB.
type StoreModule interface {
	Read(name string) (StoredReport, error)
	Update(unit string, validity time.Duration, reports json.RawMessage) error
}

// Component is the Health component.
type Component struct {
	influx              InfluxHealthChecker
	jaeger              JaegerHealthChecker
	redis               RedisHealthChecker
	sentry              SentryHealthChecker
	storage             StoreModule
	healthCheckValidity map[string]time.Duration
}

// NewComponent returns the health component.
func NewComponent(influx InfluxHealthChecker, jaeger JaegerHealthChecker, redis RedisHealthChecker, sentry SentryHealthChecker, storage StoreModule, healthCheckValidity map[string]time.Duration) *Component {
	return &Component{
		influx:              influx,
		jaeger:              jaeger,
		redis:               redis,
		sentry:              sentry,
		storage:             storage,
		healthCheckValidity: healthCheckValidity,
	}
}

// ExecInfluxHealthChecks executes the health checks for Influx.
func (c *Component) ExecInfluxHealthChecks(ctx context.Context) json.RawMessage {
	var reports = c.influx.HealthChecks(ctx)
	var jsonReports, _ = json.Marshal(reports)

	c.storage.Update(influxUnitName, c.healthCheckValidity[influxUnitName], jsonReports)
	return json.RawMessage(jsonReports)
}

// ReadInfluxHealthChecks read the health checks status in DB.
func (c *Component) ReadInfluxHealthChecks(ctx context.Context) json.RawMessage {
	return c.readFromDB(influxUnitName)
}

// ExecJaegerHealthChecks executes the health checks for Jaeger.
func (c *Component) ExecJaegerHealthChecks(ctx context.Context) json.RawMessage {
	var reports = c.jaeger.HealthChecks(ctx)
	var jsonReports, _ = json.Marshal(reports)

	c.storage.Update(jaegerUnitName, c.healthCheckValidity[jaegerUnitName], jsonReports)
	return json.RawMessage(jsonReports)
}

// ReadJaegerHealthChecks read the health checks status in DB.
func (c *Component) ReadJaegerHealthChecks(ctx context.Context) json.RawMessage {
	return c.readFromDB(jaegerUnitName)
}

// ExecRedisHealthChecks executes the health checks for Redis.
func (c *Component) ExecRedisHealthChecks(ctx context.Context) json.RawMessage {
	var reports = c.redis.HealthChecks(ctx)
	var jsonReports, _ = json.Marshal(reports)

	c.storage.Update(redisUnitName, c.healthCheckValidity[redisUnitName], jsonReports)
	return json.RawMessage(jsonReports)

}

// ReadRedisHealthChecks read the health checks status in DB.
func (c *Component) ReadRedisHealthChecks(ctx context.Context) json.RawMessage {
	return c.readFromDB(redisUnitName)
}

// ExecSentryHealthChecks executes the health checks for Sentry.
func (c *Component) ExecSentryHealthChecks(ctx context.Context) json.RawMessage {
	var reports = c.sentry.HealthChecks(ctx)
	var jsonReports, _ = json.Marshal(reports)

	c.storage.Update(sentryUnitName, c.healthCheckValidity[sentryUnitName], jsonReports)
	return json.RawMessage(jsonReports)
}

// ReadSentryHealthChecks read the health checks status in DB.
func (c *Component) ReadSentryHealthChecks(ctx context.Context) json.RawMessage {
	return c.readFromDB(sentryUnitName)
}

// AllHealthChecks call all component checks and build a general health report.
func (c *Component) AllHealthChecks(ctx context.Context) json.RawMessage {
	var reports = map[string]json.RawMessage{}

	reports[influxUnitName] = c.ReadInfluxHealthChecks(ctx)
	reports[jaegerUnitName] = c.ReadJaegerHealthChecks(ctx)
	reports[redisUnitName] = c.ReadRedisHealthChecks(ctx)
	reports[sentryUnitName] = c.ReadSentryHealthChecks(ctx)

	var jsonReports, _ = json.Marshal(reports)
	return json.RawMessage(jsonReports)
}

func (c *Component) readFromDB(unit string) json.RawMessage {
	var storedReport, err = c.storage.Read(unit)

	type report struct {
		Name   string `json:"name"`
		Status string `json:"status"`
		Error  string `json:"error"`
	}

	if err != nil {
		var jsonReport, _ = json.Marshal(report{
			Name:   unit,
			Status: common.KO.String(),
			Error:  fmt.Sprintf("could not read reports from DB: %v", err),
		})

		return json.RawMessage(jsonReport)
	}

	if storedReport.ComponentID == "" {
		var jsonReport, _ = json.Marshal(report{
			Name:   unit,
			Status: common.KO.String(),
			Error:  "no reports stored in DB",
		})

		return json.RawMessage(jsonReport)
	}

	// If the health check was executed too long ago, the health check report
	// is considered not pertinant and an error is returned.
	if time.Now().After(storedReport.ValidUntil) {
		var jsonReport, _ = json.Marshal(report{
			Name:   unit,
			Status: common.KO.String(),
			Error:  fmt.Sprintf("the health check results are stale because the test was not executed in the last %s", c.healthCheckValidity[storedReport.HealthcheckUnit]),
		})

		return json.RawMessage(jsonReport)
	}

	return storedReport.Reports
}
