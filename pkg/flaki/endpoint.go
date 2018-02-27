package flaki

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/pkg/errors"
)

// Endpoints wraps a service behind a set of endpoints.
type Endpoints struct {
	NextIDEndpoint      endpoint.Endpoint
	NextValidIDEndpoint endpoint.Endpoint
}

// MakeNextIDEndpoint makes the NextIDEndpoint.
func MakeNextIDEndpoint(c Component) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		var id, err = c.NextID(ctx)
		if err != nil {
			return "", errors.Wrap(err, "component could not generate ID")
		}
		return id, nil
	}
}

// MakeNextValidIDEndpoint makes the NextValidIDEndpoint.
func MakeNextValidIDEndpoint(c Component) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return c.NextValidID(ctx), nil
	}
}
