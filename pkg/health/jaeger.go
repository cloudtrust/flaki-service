package health

import (
	"context"
)

// JaegerModule is the health check module for jaeger.
type JaegerModule interface {
	HealthChecks(context.Context) []jaegerHealthReport
}

type jaegerModule struct {
	jaeger Jaeger
}

type jaegerHealthReport struct {
	Name     string
	Duration string
	Status   Status
	Error    string
}

// Jaeger is the interface of the jaeger client.
type Jaeger interface {
	//Ping(timeout time.Duration) (time.Duration, error)
}

// NewJaegerModule returns the jaeger health module.
func NewJaegerModule(jaeger Jaeger) JaegerModule {
	return &jaegerModule{jaeger: jaeger}
}

// HealthChecks executes all health checks for Jaeger.
func (m *jaegerModule) HealthChecks(context.Context) []jaegerHealthReport {
	var reports = []jaegerHealthReport{}
	reports = append(reports, jaegerPingCheck(m.jaeger))
	return reports
}

func jaegerPingCheck(jaeger Jaeger) jaegerHealthReport {
	return jaegerHealthReport{
		Name:     "ping",
		Duration: "N/A",
		Status:   KO,
		Error:    "Not yet implemented",
	}
}
