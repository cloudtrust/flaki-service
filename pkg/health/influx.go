package health

//go:generate mockgen -destination=./mock/influx.go -package=mock -mock_names=InfluxModule=InfluxModule,Influx=Influx  github.com/cloudtrust/flaki-service/pkg/health InfluxModule,Influx

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// InfluxModule is the health check module for influx.
type InfluxModule interface {
	HealthChecks(context.Context) []InfluxReport
}

type influxModule struct {
	influx  Influx
	enabled bool
}

// InfluxReport is the health report returned by the influx module.
type InfluxReport struct {
	Name     string
	Duration time.Duration
	Status   Status
	Error    error
}

// Influx is the interface of the influx client.
type Influx interface {
	Ping(timeout time.Duration) (time.Duration, string, error)
}

// NewInfluxModule returns the influx health module.
func NewInfluxModule(influx Influx, enabled bool) InfluxModule {
	return &influxModule{
		influx:  influx,
		enabled: enabled,
	}
}

// HealthChecks executes all health checks for influx.
func (m *influxModule) HealthChecks(context.Context) []InfluxReport {
	var reports = []InfluxReport{}
	reports = append(reports, m.influxPing())
	return reports
}

func (m *influxModule) influxPing() InfluxReport {
	var healthCheckName = "ping"

	if !m.enabled {
		return InfluxReport{
			Name:   healthCheckName,
			Status: Deactivated,
		}
	}

	var now = time.Now()
	var _, _, err = m.influx.Ping(5 * time.Second)
	var duration = time.Since(now)

	var hcErr error
	var s Status
	switch {
	case err != nil:
		hcErr = errors.Wrap(err, "could not ping influx")
		s = KO
	default:
		s = OK
	}

	return InfluxReport{
		Name:     healthCheckName,
		Duration: duration,
		Status:   s,
		Error:    hcErr,
	}
}
