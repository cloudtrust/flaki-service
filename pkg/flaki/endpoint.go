package flaki

import (
	"context"
	"fmt"

	"github.com/cloudtrust/flaki-service/pkg/flaki/flatbuffer/fb"
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
		switch r := req.(type) {
		case *fb.FlakiRequest:
			return c.NextID(ctx, r)
		default:
			return nil, fmt.Errorf("wrong request type: %T", req)
		}
	}
}

// MakeNextValidIDEndpoint makes the NextValidIDEndpoint.
func MakeNextValidIDEndpoint(c Component) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		switch r := req.(type) {
		case *fb.FlakiRequest:
			return c.NextValidID(ctx, r), nil
		default:
			return nil, fmt.Errorf("wrong request type: %T", req)
		}
	}
}
