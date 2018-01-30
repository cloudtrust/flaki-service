package flaki

import (
	"context"
)

// Service is the interface that the services implement.
type Service interface {
	NextID(context.Context) (string, error)
	NextValidID(context.Context) string
}

// Flaki is the interface of the distributed unique IDs generator.
type Flaki interface {
	NextIDString() (string, error)
	NextValidIDString() string
}

type basicService struct {
	flaki Flaki
}

// NewBasicService returns the Flaki basic service.
func NewBasicService(flaki Flaki) Service {
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
