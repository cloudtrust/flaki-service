package flaki

import (
	"context"
	flaki "github.com/cloudtrust/flaki"
)

// Service is the interface that the services implement.
type Service interface {
	NextID(context.Context) (uint64, error)
	NextValidID(context.Context) uint64
}

type basicService struct {
	flaki flaki.Flaki
}

// NewBasicService returns the basic service
func NewBasicService(flaki flaki.Flaki) Service {
	return &basicService{
		flaki: flaki,
	}
}

func (s *basicService) NextID(_ context.Context) (uint64, error) {
	return s.flaki.NextID()
}

func (s *basicService) NextValidID(_ context.Context) uint64 {
	return s.flaki.NextValidID()
}
