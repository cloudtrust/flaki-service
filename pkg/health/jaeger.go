package health

//go:generate mockgen -destination=./mock/jaeger.go -package=mock -mock_names=JaegerModule=JaegerModule,SystemDConn=SystemDConn  github.com/cloudtrust/flaki-service/pkg/health JaegerModule,SystemDConn

import (
	"context"
	"net/http"
	"time"

	"github.com/coreos/go-systemd/dbus"
	"github.com/pkg/errors"
)

const (
	agentSystemDUnitName = "agent.service"
)

// JaegerModule is the health check module for jaeger.
type JaegerModule interface {
	HealthChecks(context.Context) []JaegerReport
}

type jaegerModule struct {
	conn                    SystemDConn
	collectorHealthCheckURL string
	httpClient              JaegerHTTPClient
	enabled                 bool
}

// JaegerReport is the health report returned by the jaeger module.
type JaegerReport struct {
	Name     string
	Duration time.Duration
	Status   Status
	Error    error
}

// SystemDConn is interface of systemd D-Bus connection.
type SystemDConn interface {
	ListUnitsByNames(units []string) ([]dbus.UnitStatus, error)
}

// JaegerHTTPClient is the interface of the http client.
type JaegerHTTPClient interface {
	Get(string) (*http.Response, error)
}

// NewJaegerModule returns the jaeger health module.
func NewJaegerModule(conn SystemDConn, httpClient JaegerHTTPClient, collectorHealthCheckURL string, enabled bool) JaegerModule {
	return &jaegerModule{
		conn:                    conn,
		httpClient:              httpClient,
		collectorHealthCheckURL: collectorHealthCheckURL,
		enabled:                 enabled,
	}
}

// HealthChecks executes all health checks for Jaeger.
func (m *jaegerModule) HealthChecks(context.Context) []JaegerReport {
	var reports = []JaegerReport{}
	reports = append(reports, m.jaegerSystemDCheck())
	reports = append(reports, m.jaegerCollectorPing())
	return reports
}

func (m *jaegerModule) jaegerSystemDCheck() JaegerReport {
	var healthCheckName = "jaeger agent systemd unit check"

	if !m.enabled {
		return JaegerReport{
			Name:   healthCheckName,
			Status: Deactivated,
		}
	}

	var now = time.Now()
	var units, err = m.conn.ListUnitsByNames([]string{agentSystemDUnitName})
	var duration = time.Since(now)

	var hcErr error
	var s Status
	switch {
	case err != nil:
		hcErr = errors.Wrapf(err, "could not list '%s' systemd unit", agentSystemDUnitName)
		s = KO
	case len(units) == 0:
		hcErr = errors.Wrapf(err, "systemd unit '%s' not found", agentSystemDUnitName)
		s = KO
	case units[0].ActiveState != "active":
		hcErr = errors.Wrapf(err, "systemd unit '%s' is not active", agentSystemDUnitName)
		s = KO
	default:
		s = OK
	}

	return JaegerReport{
		Name:     healthCheckName,
		Duration: duration,
		Status:   s,
		Error:    hcErr,
	}
}

func (m *jaegerModule) jaegerCollectorPing() JaegerReport {
	var healthCheckName = "ping jaeger collector"

	if !m.enabled {
		return JaegerReport{
			Name:   healthCheckName,
			Status: Deactivated,
		}
	}

	// query jaeger collector health check URL
	var now = time.Now()
	var res, err = m.httpClient.Get("http://" + m.collectorHealthCheckURL)
	var duration = time.Since(now)

	var hcErr error
	var s Status
	switch {
	case err != nil:
		hcErr = errors.Wrap(err, "could not query jaeger collector health check service")
		s = KO
	case res.StatusCode != 204:
		hcErr = errors.Wrapf(err, "jaeger health check service returned invalid status code: %v", res.StatusCode)
		s = KO
	default:
		s = OK
	}

	return JaegerReport{
		Name:     healthCheckName,
		Duration: duration,
		Status:   s,
		Error:    hcErr,
	}
}
