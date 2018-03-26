package health

//go:generate mockgen -destination=./mock/flakiModule.go -package=mock -mock_names=Module=FlakiModule github.com/cloudtrust/flaki-service/pkg/flaki Module

import (
	"context"

	"github.com/cloudtrust/flaki-service/pkg/flaki"
	"github.com/go-kit/kit/endpoint"
)

// IDGenerator is the interface of the distributed unique IDs generator.
type IDGenerator interface {
	NextValidIDString(context.Context) (string, error)
}

// MakeEndpointCorrelationIDMW makes a middleware that adds a correlation ID
// in the context if there is not already one.
func MakeEndpointCorrelationIDMW(g flaki.Module) endpoint.Middleware {
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
