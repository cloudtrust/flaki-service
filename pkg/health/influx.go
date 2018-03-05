package health

//go:generate mockgen -destination=./mock/influx.go -package=mock -mock_names=InfluxModule=InfluxModule,Influx=Influx  github.com/cloudtrust/flaki-service/pkg/health InfluxModule,Influx

import (
	"context"
	"fmt"
	"time"
)

// InfluxModule is the health check module for influx.
type InfluxModule interface {
	HealthChecks(context.Context) []InfluxHealthReport
}

type influxModule struct {
	influx  Influx
	enabled bool
}

// InfluxHealthReport is the health report returned by the influx module.
type InfluxHealthReport struct {
	Name     string
	Duration string
	Status   Status
	Error    string
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
func (m *influxModule) HealthChecks(context.Context) []InfluxHealthReport {
	var reports = []InfluxHealthReport{}
	reports = append(reports, m.influxPing())
	return reports
}

func (m *influxModule) influxPing() InfluxHealthReport {
	var healthCheckName = "ping"

	if !m.enabled {
		return InfluxHealthReport{
			Name:     healthCheckName,
			Duration: "N/A",
			Status:   Deactivated,
		}
	}

	var d, _, err = m.influx.Ping(5 * time.Second)

	var error string
	var s Status
	switch {
	case err != nil:
		error = fmt.Sprintf("could not ping influx: %v", err.Error())
		s = KO
	default:
		s = OK
	}

	return InfluxHealthReport{
		Name:     healthCheckName,
		Duration: d.String(),
		Status:   s,
		Error:    error,
	}
}
