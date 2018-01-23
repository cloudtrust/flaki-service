package endpoint

import (
	"context"

	"github.com/cloudtrust/flaki-service/service/component"
	"github.com/go-kit/kit/endpoint"
)

// Endpoints wraps a service behind a set of endpoints.
type Endpoints struct {
	mids                []endpoint.Middleware
	NextIDEndpoint      endpoint.Endpoint
	NextValidIDEndpoint endpoint.Endpoint
}

func NewEndpoints(mids ...endpoint.Middleware) *Endpoints {
	var m = append([]endpoint.Middleware{}, mids...)
	return &Endpoints{
		mids: m,
	}
}

func (es *Endpoints) MakeNextIDEndpoint(s component.Service, mids ...endpoint.Middleware) *Endpoints {
	var e endpoint.Endpoint = func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.NextID(ctx)
	}
	e = es.applyMids(e, mids...)
	es.NextIDEndpoint = e
	return es
}

func (es *Endpoints) MakeNextValidIDEndpoint(s component.Service, mids ...endpoint.Middleware) *Endpoints {
	var e endpoint.Endpoint = func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.NextValidID(ctx), nil
	}
	e = es.applyMids(e, mids...)
	es.NextValidIDEndpoint = e
	return es
}

func (es *Endpoints) applyMids(e endpoint.Endpoint, mids ...endpoint.Middleware) endpoint.Endpoint {
	for _, m := range mids {
		e = m(e)
	}
	for _, m := range es.mids {
		e = m(e)
	}
	return e
}

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

func (es *Endpoints) NextValidID(ctx context.Context) string {
	var id string
	{
		var idPreCast interface{}
		idPreCast, _ = es.NextValidIDEndpoint(ctx, nil)
		id = idPreCast.(string)
	}
	return id
}
