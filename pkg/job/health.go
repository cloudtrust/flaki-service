package job

import (
	"context"
	"time"
	"encoding/json"

	common "github.com/cloudtrust/common-healthcheck"
	"github.com/cloudtrust/go-jobs/job"
	"github.com/go-kit/kit/log"
)

// Cockroach is the interface of the module that stores the health reports
// in the DB.
type Cockroach interface {
	Update(unit string, validity time.Duration, jsonReports json.RawMessage) error
	Clean() error
}

// Flaki is the interface of the IDs generator.
type Flaki interface {
	NextValidIDString() string
}

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

// MakeInfluxJob creates the job that periodically exectutes the health checks and save the result in DB.
func MakeInfluxJob(influx InfluxHealthChecker, healthCheckValidity time.Duration, cockroach Cockroach) (*job.Job, error) {
	var step1 = func(ctx context.Context, r interface{}) (interface{}, error) {
		return influx.HealthChecks(ctx), nil
	}

	var step2 = func(_ context.Context, r interface{}) (interface{}, error) {
		var jsonReports, _ = json.Marshal(r)
	
		var err = cockroach.Update("influx", healthCheckValidity, jsonReports)
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
		var jsonReports, _ = json.Marshal(r)
	
		var err = cockroach.Update("jaeger", healthCheckValidity, jsonReports)
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
		var jsonReports, _ = json.Marshal(r)
	
		var err = cockroach.Update("redis", healthCheckValidity, jsonReports)
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
		var jsonReports, _ = json.Marshal(r)
	
		var err = cockroach.Update("sentry", healthCheckValidity, jsonReports)
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
