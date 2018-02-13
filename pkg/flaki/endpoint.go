package flaki

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints wraps a service behind a set of endpoints.
type Endpoints struct {
	NextIDEndpoint      endpoint.Endpoint
	NextValidIDEndpoint endpoint.Endpoint
}

// MakeNextIDEndpoint makes the NextIDEndpoint.
func MakeNextIDEndpoint(c Component) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return c.NextID(ctx)
	}
}

// MakeNextValidIDEndpoint makes the NextValidIDEndpoint.
func MakeNextValidIDEndpoint(c Component) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return c.NextValidID(ctx), nil
	}
}
