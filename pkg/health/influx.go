package health

import (
	"context"
	"time"
)

// InfluxModule is the health check module for influx.
type InfluxModule interface {
	HealthChecks(context.Context) []influxHealthReport
}

type influxModule struct {
	influx Influx
}

type influxHealthReport struct {
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
func NewInfluxModule(influx Influx) InfluxModule {
	return &influxModule{influx: influx}
}

// HealthChecks executes all health checks for influx.
func (m *influxModule) HealthChecks(context.Context) []influxHealthReport {
	var reports = []influxHealthReport{}
	reports = append(reports, influxPing(m.influx))
	return reports
}

func influxPing(influx Influx) influxHealthReport {
	var d, s, err = influx.Ping(5 * time.Second)

	// If influx is deactivated.
	if s == "NOOP" {
		return influxHealthReport{
			Name:     "ping",
			Duration: "N/A",
			Status:   Deactivated,
		}
	}

	var status = OK
	var error = ""
	if err != nil {
		status = KO
		error = err.Error()
	}

	return influxHealthReport{
		Name:     "ping",
		Duration: d.String(),
		Status:   status,
		Error:    error,
	}
}
