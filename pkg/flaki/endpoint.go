package flaki

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints wraps a service behind a set of endpoints.
type Endpoints struct {
	mids                []endpoint.Middleware
	NextIDEndpoint      endpoint.Endpoint
	NextValidIDEndpoint endpoint.Endpoint
}

// NewEndpoints returns Endpoints with the middlware mids. Mids are used to apply middlware
// to all the endpoint in Endpoints.
func NewEndpoints(mids ...endpoint.Middleware) *Endpoints {
	var m = append([]endpoint.Middleware{}, mids...)
	return &Endpoints{
		mids: m,
	}
}

// MakeNextIDEndpoint makes the NextIDEndpoint and apply the middelwares mids and Endpoints.mids.
func (es *Endpoints) MakeNextIDEndpoint(s Component, mids ...endpoint.Middleware) *Endpoints {
	var e endpoint.Endpoint = func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.NextID(ctx)
	}
	e = es.applyMids(e, mids...)
	es.NextIDEndpoint = e
	return es
}

// MakeNextValidIDEndpoint makes the NextValidIDEndpoint and apply the middelwares mids and Endpoints.mids.
func (es *Endpoints) MakeNextValidIDEndpoint(s Component, mids ...endpoint.Middleware) *Endpoints {
	var e endpoint.Endpoint = func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.NextValidID(ctx), nil
	}
	e = es.applyMids(e, mids...)
	es.NextValidIDEndpoint = e
	return es
}

// applyMids apply first the middlware mids, then Endpoints.mids to the endpoint.
func (es *Endpoints) applyMids(e endpoint.Endpoint, mids ...endpoint.Middleware) endpoint.Endpoint {
	for _, m := range mids {
		e = m(e)
	}
	for _, m := range es.mids {
		e = m(e)
	}
	return e
}

// Implements Service.
func (es *Endpoints) NextID(ctx context.Context) (string, error) {
	var id string
	{
		var idPreCast interface{}
		var err error
		idPreCast, err = es.NextIDEndpoint(ctx, nil)
		if err != nil {
			return "", err
		}
		id = idPreCast.(string)
	}
	return id, nil
}

// Implements Service.
func (es *Endpoints) NextValidID(ctx context.Context) string {
	var id string
	{
		var idPreCast interface{}
		idPreCast, _ = es.NextValidIDEndpoint(ctx, nil)
		id = idPreCast.(string)
	}
	return id
}
