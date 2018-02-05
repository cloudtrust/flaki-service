package module

import (
	"context"
	"time"
)

type JaegerHealthReport struct {
	Name     string
	Duration string
	Status   string
	Error    string
}

type Jaeger interface {
	//Ping(timeout time.Duration) (time.Duration, string, error)
}

type JaegerHealthModule struct {
	jaeger Jaeger
}

// NewJaegerHealthModule returns the jaeger health module.
func NewJaegerHealthModule(jaeger Jaeger) *JaegerHealthModule {
	return &JaegerHealthModule{jaeger: jaeger}
}

// HealthChecks executes all health checks for Jaeger.
func (s *JaegerHealthModule) HealthChecks(context.Context) []JaegerHealthReport {
	var reports = []JaegerHealthReport{}
	reports = append(reports, jaegerPingCheck(s.jaeger))
	return reports
}

func jaegerPingCheck(jaeger Jaeger) JaegerHealthReport {
	var duration = time.Duration(1 * time.Second)
	var status = "Not yet implemented"
	return JaegerHealthReport{
		Name:     "ping",
		Duration: duration.String(),
		Status:   status,
		Error:    "Not implemented",
	}
}
