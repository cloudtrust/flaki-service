package flaki

//go:generate mockgen -destination=./mock/component.go -package=mock -mock_names=IDGeneratorComponent=IDGeneratorComponent github.com/cloudtrust/flaki-service/pkg/flaki IDGeneratorComponent

import (
	"context"
	"fmt"

	"github.com/cloudtrust/flaki-service/api/fb"
	"github.com/go-kit/kit/endpoint"
)

// Endpoints wraps a service behind a set of endpoints.
type Endpoints struct {
	NextIDEndpoint      endpoint.Endpoint
	NextValidIDEndpoint endpoint.Endpoint
}

// IDGeneratorComponent is the flaki component interface.
type IDGeneratorComponent interface {
	NextID(context.Context, *fb.FlakiRequest) (*fb.FlakiReply, error)
	NextValidID(context.Context, *fb.FlakiRequest) *fb.FlakiReply
}

// MakeNextIDEndpoint makes the NextIDEndpoint.
func MakeNextIDEndpoint(c IDGeneratorComponent) endpoint.Endpoint {
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
func MakeNextValidIDEndpoint(c IDGeneratorComponent) endpoint.Endpoint {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		switch r := req.(type) {
		case *fb.FlakiRequest:
			return c.NextValidID(ctx, r), nil
		default:
			return nil, fmt.Errorf("wrong request type: %T", req)
		}
	}
}
