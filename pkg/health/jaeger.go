package health

import (
	"context"
	"time"
)

type JaegerModule interface {
	HealthChecks(context.Context) []JaegerHealthReport
}

type JaegerHealthReport struct {
	Name     string
	Duration string
	Status   HealthStatus
	Error    string
}

type Jaeger interface {
	//Ping(timeout time.Duration) (time.Duration, string, error)
}

type jaegerModule struct {
	jaeger Jaeger
}

// NewJaegerModule returns the jaeger health module.
func NewJaegerModule(jaeger Jaeger) JaegerModule {
	return &jaegerModule{jaeger: jaeger}
}

// HealthChecks executes all health checks for Jaeger.
func (m *jaegerModule) HealthChecks(context.Context) []JaegerHealthReport {
	var reports = []JaegerHealthReport{}
	reports = append(reports, jaegerPingCheck(m.jaeger))
	return reports
}

func jaegerPingCheck(jaeger Jaeger) JaegerHealthReport {
	var duration = time.Duration(1 * time.Second)
	var status = KO
	return JaegerHealthReport{
		Name:     "ping",
		Duration: duration.String(),
		Status:   status,
		Error:    "Not implemented",
	}
}
