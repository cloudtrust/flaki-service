package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudtrust/go-jobs/job"
	"github.com/go-kit/kit/log"
)

// Storage is the interface of the module that stores the health reports
// in the DB.
type Storage interface {
	Update(unit string, validity time.Duration, jsonReports json.RawMessage) error
	Clean() error
}

// Flaki is the interface of the IDs generator.
type Flaki interface {
	NextValidIDString() string
}

// HealthChecker is the interface of the health check modules.
type HealthChecker interface {
	HealthCheck(context.Context, string) (json.RawMessage, error)
}

// MakeHealthJob creates the job that periodically executes the health checks and save the result in DB.
func MakeHealthJob(module HealthChecker, moduleName string, healthCheckValidity time.Duration, storage Storage) (*job.Job, error) {
	var step1 = func(_ context.Context, _ interface{}) (interface{}, error) {
		return module.HealthCheck(context.Background(), "")
	}

	var step2 = func(_ context.Context, r interface{}) (interface{}, error) {
		var jsonReports, ok = r.(json.RawMessage)
		if !ok {
			return nil, fmt.Errorf("health report should be a json.Rawmessage not %T", r)
		}

		var err = storage.Update("influx", healthCheckValidity, jsonReports)
		return nil, err
	}
	return job.NewJob("influx", job.Steps(step1, step2))
}

// MakeStorageCleaningJob creates the job that periodically clean the DB from the outdated health check reports.
func MakeStorageCleaningJob(storage Storage, logger log.Logger) (*job.Job, error) {
	var clean = func(context.Context, interface{}) (interface{}, error) {
		logger.Log("step", "clean")
		return nil, storage.Clean()
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
