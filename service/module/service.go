package flaki

import (
	"context"

	flaki "github.com/cloudtrust/flaki"
)

// Service is the interface that the services implement.
type Service interface {
	NextID(context.Context) (string, error)
	NextValidID(context.Context) string
}

type basicService struct {
	flaki flaki.Flaki
}

// NewBasicService returns the basic service.
func NewBasicService(flaki flaki.Flaki) Service {
	return &basicService{
		flaki: flaki,
	}
}

func (s *basicService) NextID(_ context.Context) (string, error) {
	return s.flaki.NextIDString()
}

func (s *basicService) NextValidID(_ context.Context) string {
	return s.flaki.NextValidIDString()
}
