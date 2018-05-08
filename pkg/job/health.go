package job

import (
	"context"
	"time"

	"github.com/cloudtrust/flaki-service/pkg/health"
	"github.com/cloudtrust/go-jobs/job"
	"github.com/go-kit/kit/log"
)

// Cockroach is the interface of the module that stores the health reports
// in the DB.
type Cockroach interface {
	Update(unit string, reports []health.StoredReport) error
	Clean() error
}

// Flaki is the interface of the IDs generator.
type Flaki interface {
	NextValidIDString() string
}

// InfluxHealthChecker is the interface of the influx health check module.
type InfluxHealthChecker interface {
	HealthChecks(context.Context) []health.InfluxReport
}

// JaegerHealthChecker is the interface of the jaeger health check module.
type JaegerHealthChecker interface {
	HealthChecks(context.Context) []health.JaegerReport
}

// RedisHealthChecker is the interface of the redis health check module.
type RedisHealthChecker interface {
	HealthChecks(context.Context) []health.RedisReport
}

// SentryHealthChecker is the interface of the sentry health check module.
type SentryHealthChecker interface {
	HealthChecks(context.Context) []health.SentryReport
}

// MakeInfluxJob creates the job that periodically exectutes the health checks and save the result in DB.
func MakeInfluxJob(influx InfluxHealthChecker, healthCheckValidity time.Duration, cockroach Cockroach) (*job.Job, error) {
	var step1 = func(ctx context.Context, r interface{}) (interface{}, error) {
		return influx.HealthChecks(ctx), nil
	}
	var step2 = func(_ context.Context, r interface{}) (interface{}, error) {
		var reports = []health.StoredReport{}
		var now = time.Now()
		for _, r := range r.([]health.InfluxReport) {
			reports = append(reports, health.StoredReport{
				Name:          r.Name,
				Duration:      r.Duration,
				Status:        r.Status,
				Error:         err(r.Error),
				LastExecution: now,
				ValidUntil:    now.Add(healthCheckValidity),
			})
		}
		var err = cockroach.Update("influx", reports)
		return nil, err
	}
	return job.NewJob("influx", job.Steps(step1, step2))
}

// MakeJaegerJob creates the job that periodically exectutes the health checks and save the result in DB.
func MakeJaegerJob(jaeger JaegerHealthChecker, healthCheckValidity time.Duration, cockroach Cockroach) (*job.Job, error) {
	var step1 = func(ctx context.Context, r interface{}) (interface{}, error) {
		return jaeger.HealthChecks(ctx), nil
	}
	var step2 = func(_ context.Context, r interface{}) (interface{}, error) {
		var reports = []health.StoredReport{}
		var now = time.Now()
		for _, r := range r.([]health.JaegerReport) {
			reports = append(reports, health.StoredReport{
				Name:          r.Name,
				Duration:      r.Duration,
				Status:        r.Status,
				Error:         err(r.Error),
				LastExecution: now,
				ValidUntil:    now.Add(healthCheckValidity),
			})
		}
		var err = cockroach.Update("jaeger", reports)
		return nil, err
	}
	return job.NewJob("jaeger", job.Steps(step1, step2))
}

// MakeRedisJob creates the job that periodically exectutes the health checks and save the result in DB.
func MakeRedisJob(redis RedisHealthChecker, healthCheckValidity time.Duration, cockroach Cockroach) (*job.Job, error) {
	var step1 = func(ctx context.Context, r interface{}) (interface{}, error) {
		return redis.HealthChecks(ctx), nil
	}
	var step2 = func(_ context.Context, r interface{}) (interface{}, error) {
		var reports = []health.StoredReport{}
		var now = time.Now()
		for _, r := range r.([]health.RedisReport) {
			reports = append(reports, health.StoredReport{
				Name:          r.Name,
				Duration:      r.Duration,
				Status:        r.Status,
				Error:         err(r.Error),
				LastExecution: now,
				ValidUntil:    now.Add(healthCheckValidity),
			})
		}
		var err = cockroach.Update("redis", reports)
		return nil, err
	}
	return job.NewJob("redis", job.Steps(step1, step2))
}

// MakeSentryJob creates the job that periodically exectutes the health checks and save the result in DB.
func MakeSentryJob(sentry SentryHealthChecker, healthCheckValidity time.Duration, cockroach Cockroach) (*job.Job, error) {
	var step1 = func(ctx context.Context, r interface{}) (interface{}, error) {
		return sentry.HealthChecks(ctx), nil
	}
	var step2 = func(_ context.Context, r interface{}) (interface{}, error) {
		var reports = []health.StoredReport{}
		var now = time.Now()
		for _, r := range r.([]health.SentryReport) {
			reports = append(reports, health.StoredReport{
				Name:          r.Name,
				Duration:      r.Duration,
				Status:        r.Status,
				Error:         err(r.Error),
				LastExecution: now,
				ValidUntil:    now.Add(healthCheckValidity),
			})
		}
		var err = cockroach.Update("sentry", reports)
		return nil, err
	}
	return job.NewJob("sentry", job.Steps(step1, step2))
}

// MakeCleanCockroachJob creates the job that periodically exectutes the health checks and save the result in DB.
func MakeCleanCockroachJob(cockroach Cockroach, logger log.Logger) (*job.Job, error) {
	var clean = func(context.Context, interface{}) (interface{}, error) {
		logger.Log("step", "clean")
		return nil, cockroach.Clean()
	}
	return job.NewJob("clean", job.Steps(clean))
}

// err return the string error that will be in the health report
func err(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
