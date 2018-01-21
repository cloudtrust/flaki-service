package endpoint

import (
	"context"
	"github.com/JohanDroz/flaki-service/service/component"
	"github.com/go-kit/kit/endpoint"
)

// Endpoints wraps a service behind a set of endpoints.
type Endpoints struct {
	NextIDEndpoint      endpoint.Endpoint
	NextValidIDEndpoint endpoint.Endpoint
}

func MakeNextIDEndpoint(s component.Service, mids ...endpoint.Middleware) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		var e endpoint.Endpoint = func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.NextID(ctx)
		}

		for _, m := range mids {
			e = m(e)
		}
		return e(ctx, nil)
	}
}

func MakeNextValidIDEndpoint(s component.Service, mids ...endpoint.Middleware) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		var e endpoint.Endpoint = func(ctx context.Context, req interface{}) (interface{}, error) {
			return s.NextValidID(ctx), nil
		}

		for _, m := range mids {
			e = m(e)
		}
		return e(ctx, nil)
	}
}
