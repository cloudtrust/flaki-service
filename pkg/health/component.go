package health

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"
)

// ModuleNames contains the list of all valid module names.
var ModuleNames = map[string]struct{}{
	"":          struct{}{},
	"cockroach": struct{}{},
	"influx":    struct{}{},
	"jaeger":    struct{}{},
	"redis":     struct{}{},
	"sentry":    struct{}{},
}

const (
	// Names of the units in the health check http response and in the DB.
	cockroachUnitName = "cockroach"
	influxUnitName    = "influx"
	jaegerUnitName    = "jaeger"
	redisUnitName     = "redis"
	sentryUnitName    = "sentry"
)

// HealthCheckStorage is the interface of the module that stores the health reports
// in the DB.
type HealthCheckStorage interface {
	Read(module, healthcheck string) (json.RawMessage, error)
	Update(module, jsonReport json.RawMessage, validity time.Duration) error 
}

// HealthChecker is the interface of the health check modules.
type HealthChecker interface {
	HealthCheck(context.Context, string) (json.RawMessage, error)
}

// Component is the Health component.
type Component struct {
	healthCheckModules  map[string]HealthChecker
	healthCheckValidity map[string]time.Duration
	storage             HealthCheckStorage
}

// NewComponent returns the health component.
func NewComponent(healthCheckModules map[string]HealthChecker, healthCheckValidity map[string]time.Duration, storage HealthCheckStorage) *Component {
	return &Component{
		healthCheckModules:  healthCheckModules,
		healthCheckValidity: healthCheckValidity,
		storage:             storage,
	}
}

func (c *Component) HealthChecks(ctx context.Context, moduleName string) (json.RawMessage, error) {
	var healthCheckName = ctx.Value("healthcheck").(string)
	var noCache bool
	{
		if ctx.Value("nocache").(string) == "1" {
			noCache = true
		}
	}

	ctx = filterContext(ctx)

	if moduleName == "" {
		return c.allHealthChecks(ctx, noCache)
	}
	return c.healthCheck(ctx, moduleName, healthCheckName, noCache)
}

func (c *Component) allHealthChecks(ctx context.Context, noCache bool) (json.RawMessage, error) {
	var reports = []json.RawMessage{}

	var names = allKeys(c.healthCheckModules)
	sort.Strings(names)

	for _, k := range names {
		var module, ok = c.healthCheckModules[k]
		if !ok {
			// Should not happen: there is a middleware validating the inputs.
			panic(fmt.Sprintf("Unknown health check module: %v", module))
		}

		var r, err = module.HealthCheck(ctx, "")
		if err != nil {
			return nil, errors.Wrapf(err, "health checks for module %s failed", module)
		}
		reports = append(reports, r)
	}
	var jsonReports json.RawMessage
	{
		var err error
		jsonReports, err = json.MarshalIndent(reports, "", "  ")
		if err != nil {
			return nil, errors.Wrap(err, "could not marshall reports")
		}
	}
	return jsonReports, nil
}

func (c *Component) (ctx context.Context, moduleName, healthCheckName string) (json.RawMessage, error) {
	var report, err = c.storage.Read(moduleName, healthCheckName)
	if err != nil {
		switch err {
		case ErrInvalid:

			fmt.Println(err.Error())
		}
	}
}

// Single health check
func (c *Component) healthCheck(ctx context.Context, moduleName, healthCheckName string, noCache bool) (json.RawMessage, error) {
	if noCache {
		var module, ok = c.healthCheckModules[moduleName]
		if !ok {
			// Should not happen: there is a middleware validating the inputs.
			panic(fmt.Sprintf("Unknown health check module: %v", module))
		}
		return module.HealthCheck(ctx, healthCheckName)
	}

	// If there is no report or the report is stale, execute the test

}

// // ExecInfluxHealthChecks executes the health checks for Influx.
// func (c *Component) ExecInfluxHealthChecks(ctx context.Context) json.RawMessage {

// 	c.storage.Update(influxUnitName, c.healthCheckValidity[influxUnitName], jsonReports)
// 	return json.RawMessage(jsonReports)
// }

// // ReadInfluxHealthChecks read the health checks status in DB.
// func (c *Component) ReadInfluxHealthChecks(ctx context.Context) json.RawMessage {
// 	return c.readFromDB(influxUnitName)
// }

func allKeys(m map[string]HealthChecker) []string {
	var keys = []string{}

	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

// The modules get a clean version of the context. This function create a new empty context
// and copy only the required keys into it.
func filterContext(ctx context.Context) context.Context {
	// New context for the modules
	var mctx = context.Background()

	mctx = context.WithValue(mctx, "correlation_id", ctx.Value("correlation_id").(string))

	return mctx
}
