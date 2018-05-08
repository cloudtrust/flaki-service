package health

//go:generate mockgen -destination=./mock/module.go -package=mock -mock_names=InfluxHealthChecker=InfluxHealthChecker,JaegerHealthChecker=JaegerHealthChecker,RedisHealthChecker=RedisHealthChecker,SentryHealthChecker=SentryHealthChecker,StorageModule=StorageModule  github.com/cloudtrust/flaki-service/pkg/health InfluxHealthChecker,JaegerHealthChecker,RedisHealthChecker,SentryHealthChecker,StorageModule

import (
	"context"
	"fmt"
	"time"
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
	// Unknown is the status set when there is unexpected errors, e.g. parsing status from DB.
	Unknown
)

const (
	// Names of the units in the health check http response and in the DB.
	influxUnitName = "influx"
	jaegerUnitName = "jaeger"
	redisUnitName  = "redis"
	sentryUnitName = "sentry"
)

var statusName = []string{"OK", "KO", "Degraded", "Deactivated", "Unknown"}

func status(s string) Status {
	for i, n := range statusName {
		if n == s {
			return Status(i)
		}
	}
	return Unknown
}

func (s Status) String() string {
	if s >= 0 && int(s) <= len(statusName) {
		return statusName[s]
	}
	return Unknown.String()
}

// InfluxHealthChecker is the interface of the influx health check module.
type InfluxHealthChecker interface {
	HealthChecks(context.Context) []InfluxReport
}

// JaegerHealthChecker is the interface of the jaeger health check module.
type JaegerHealthChecker interface {
	HealthChecks(context.Context) []JaegerReport
}

// RedisHealthChecker is the interface of the redis health check module.
type RedisHealthChecker interface {
	HealthChecks(context.Context) []RedisReport
}

// SentryHealthChecker is the interface of the sentry health check module.
type SentryHealthChecker interface {
	HealthChecks(context.Context) []SentryReport
}

// StorageModule is the interface of the module that stores the health reports
// in the DB.
type StorageModule interface {
	Read(name string) ([]StoredReport, error)
	Update(unit string, reports []StoredReport) error
}

// Component is the Health component.
type Component struct {
	influx              InfluxHealthChecker
	jaeger              JaegerHealthChecker
	redis               RedisHealthChecker
	sentry              SentryHealthChecker
	storage             StorageModule
	healthCheckValidity map[string]time.Duration
}

// NewComponent returns the health component.
func NewComponent(influx InfluxHealthChecker, jaeger JaegerHealthChecker, redis RedisHealthChecker, sentry SentryHealthChecker, storage StorageModule, healthCheckValidity map[string]time.Duration) *Component {
	return &Component{
		influx:              influx,
		jaeger:              jaeger,
		redis:               redis,
		sentry:              sentry,
		storage:             storage,
		healthCheckValidity: healthCheckValidity,
	}
}

// Report contains the result of one health test.
type Report struct {
	Name     string
	Duration string
	Status   string
	Error    string
}

// ExecInfluxHealthChecks executes the health checks for Influx.
func (c *Component) ExecInfluxHealthChecks(ctx context.Context) []Report {
	var reports = c.influx.HealthChecks(ctx)

	var now = time.Now()
	var out = []Report{}
	var dbReport = []StoredReport{}
	for _, r := range reports {
		out = append(out, Report{
			Name:     r.Name,
			Duration: r.Duration.String(),
			Status:   r.Status.String(),
			Error:    err(r.Error),
		})
		dbReport = append(dbReport, StoredReport{
			Name:          r.Name,
			Duration:      r.Duration,
			Status:        r.Status,
			Error:         err(r.Error),
			LastExecution: now,
			ValidUntil:    now.Add(c.healthCheckValidity[influxUnitName]),
		})
	}

	c.storage.Update(influxUnitName, dbReport)
	return out
}

// ReadInfluxHealthChecks read the health checks status in DB.
func (c *Component) ReadInfluxHealthChecks(ctx context.Context) []Report {
	return c.readFromDB(influxUnitName)
}

// ExecJaegerHealthChecks executes the health checks for Jaeger.
func (c *Component) ExecJaegerHealthChecks(ctx context.Context) []Report {
	var reports = c.jaeger.HealthChecks(ctx)

	var now = time.Now()
	var out = []Report{}
	var dbReport = []StoredReport{}
	for _, r := range reports {
		out = append(out, Report{
			Name:     r.Name,
			Duration: r.Duration.String(),
			Status:   r.Status.String(),
			Error:    err(r.Error),
		})
		dbReport = append(dbReport, StoredReport{
			Name:          r.Name,
			Duration:      r.Duration,
			Status:        r.Status,
			Error:         err(r.Error),
			LastExecution: now,
			ValidUntil:    now.Add(c.healthCheckValidity[jaegerUnitName]),
		})
	}

	c.storage.Update(jaegerUnitName, dbReport)
	return out
}

// ReadJaegerHealthChecks read the health checks status in DB.
func (c *Component) ReadJaegerHealthChecks(ctx context.Context) []Report {
	return c.readFromDB(jaegerUnitName)
}

// ExecRedisHealthChecks executes the health checks for Redis.
func (c *Component) ExecRedisHealthChecks(ctx context.Context) []Report {
	var reports = c.redis.HealthChecks(ctx)

	var now = time.Now()
	var out = []Report{}
	var dbReport = []StoredReport{}
	for _, r := range reports {
		out = append(out, Report{
			Name:     r.Name,
			Duration: r.Duration.String(),
			Status:   r.Status.String(),
			Error:    err(r.Error),
		})
		dbReport = append(dbReport, StoredReport{
			Name:          r.Name,
			Duration:      r.Duration,
			Status:        r.Status,
			Error:         err(r.Error),
			LastExecution: now,
			ValidUntil:    now.Add(c.healthCheckValidity[redisUnitName]),
		})
	}

	c.storage.Update(redisUnitName, dbReport)
	return out
}

// ReadRedisHealthChecks read the health checks status in DB.
func (c *Component) ReadRedisHealthChecks(ctx context.Context) []Report {
	return c.readFromDB(redisUnitName)
}

// ExecSentryHealthChecks executes the health checks for Sentry.
func (c *Component) ExecSentryHealthChecks(ctx context.Context) []Report {
	var reports = c.sentry.HealthChecks(ctx)

	var now = time.Now()
	var out = []Report{}
	var dbReport = []StoredReport{}
	for _, r := range reports {
		out = append(out, Report{
			Name:     r.Name,
			Duration: r.Duration.String(),
			Status:   r.Status.String(),
			Error:    err(r.Error),
		})
		dbReport = append(dbReport, StoredReport{
			Name:          r.Name,
			Duration:      r.Duration,
			Status:        r.Status,
			Error:         err(r.Error),
			LastExecution: now,
			ValidUntil:    now.Add(c.healthCheckValidity[sentryUnitName]),
		})
	}

	c.storage.Update(sentryUnitName, dbReport)
	return out
}

// ReadSentryHealthChecks read the health checks status in DB.
func (c *Component) ReadSentryHealthChecks(ctx context.Context) []Report {
	return c.readFromDB(sentryUnitName)
}

// AllHealthChecks call all component checks and build a general health report.
func (c *Component) AllHealthChecks(ctx context.Context) map[string]string {
	var reports = map[string]string{}

	reports[influxUnitName] = determineStatus(c.ReadInfluxHealthChecks(ctx))
	reports[jaegerUnitName] = determineStatus(c.ReadJaegerHealthChecks(ctx))
	reports[redisUnitName] = determineStatus(c.ReadRedisHealthChecks(ctx))
	reports[sentryUnitName] = determineStatus(c.ReadSentryHealthChecks(ctx))

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

func (c *Component) readFromDB(unit string) []Report {
	var reports, err = c.storage.Read(unit)

	switch {
	case err != nil:
		return []Report{{
			Name:   unit,
			Status: Unknown.String(),
			Error:  fmt.Sprintf("could not read reports from DB: %v", err),
		}}
	case len(reports) == 0:
		return []Report{{
			Name:   unit,
			Status: Unknown.String(),
			Error:  fmt.Sprintf("no reports stored in DB"),
		}}
	}

	var out = []Report{}
	for _, r := range reports {
		// If the health check was executed too long ago, the health check report
		// is considered not pertinant and an error is returned.
		if time.Now().After(r.ValidUntil) {
			out = append(out, Report{
				Name:  r.Name,
				Error: fmt.Sprintf("the health check results are stale because the test was not executed in the last %s", c.healthCheckValidity[r.Name]),
			})
		} else {
			out = append(out, Report{
				Name:     r.Name,
				Duration: r.Duration.String(),
				Status:   r.Status.String(),
				Error:    r.Error,
			})
		}
	}

	return out
}
