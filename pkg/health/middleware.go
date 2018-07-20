package health

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kit/kit/endpoint"
)

// IDGenerator is the interface of the distributed unique IDs generator.
type IDGenerator interface {
	NextValidID(context.Context) string
}

// MakeEndpointCorrelationIDMW makes a middleware that adds a correlation ID
// in the context if there is not already one.
func MakeEndpointCorrelationIDMW(g IDGenerator) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var id = ctx.Value("correlation_id")

			// If there is no correlation ID in the context, request one.
			if id == nil {
				ctx = context.WithValue(ctx, "correlation_id", g.NextValidID(ctx))
			}
			return next(ctx, req)
		}
	}
}

// MakeValidationMiddleware makes a middleware that validate the health check module comming from
// the HTTP route.
func MakeValidationMiddleware(validValues map[string]struct{}) func(HealthCheckers) HealthCheckers {
	return func(next HealthCheckers) HealthCheckers {
		return &validationMW{
			validValues: validValues,
			next:        next,
		}
	}
}

type validationMW struct {
	validValues map[string]struct{}
	next        HealthCheckers
}

type ErrInvalidHCModule struct {
	s string
}

func (e *ErrInvalidHCModule) Error() string {
	return fmt.Sprintf("no health check module with name '%s'", e.s)
}

func (m *validationMW) HealthChecks(ctx context.Context, module string) (json.RawMessage, error) {
	// Check health check module validity.
	var _, ok = m.validValues[module]
	if !ok {
		return nil, &ErrInvalidHCModule{module}
	}

	return m.next.HealthChecks(ctx, module)
}
