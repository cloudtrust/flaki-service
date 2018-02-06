package health

import (
	"context"
	"time"
)

type InfluxHealthReport struct {
	Name     string
	Duration string
	Status   string
	Error    string
}

type Influx interface {
	Ping(timeout time.Duration) (time.Duration, string, error)
}

type InfluxHealthModule struct {
	influx Influx
}

// NewInfluxHealthModule returns the influx health module.
func NewInfluxHealthModule(influx Influx) *InfluxHealthModule {
	return &InfluxHealthModule{influx: influx}
}

// HealthChecks executes all health checks for Influx.
func (s *InfluxHealthModule) HealthChecks(context.Context) []InfluxHealthReport {
	var reports = []InfluxHealthReport{}
	reports = append(reports, influxPingCheck(s.influx))
	return reports
}

func influxPingCheck(influx Influx) InfluxHealthReport {
	var d, _, err = influx.Ping(time.Duration(5 * time.Second))

	var status = "OK"
	var error = ""
	if err != nil {
		status = "KO"
		error = err.Error()
	}

	return InfluxHealthReport{
		Name:     "ping",
		Duration: d.String(),
		Status:   status,
		Error:    error,
	}
}
