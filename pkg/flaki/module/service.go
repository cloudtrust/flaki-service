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

type flakiService struct {
	flaki Flaki
}

// NewBasicService returns the Flaki basic service.
func NewBasicService(flaki Flaki) Service {
	return &flakiService{
		flaki: flaki,
	}
}

func (s *flakiService) NextID(_ context.Context) (string, error) {
	return s.flaki.NextIDString()
}

func (s *flakiService) NextValidID(_ context.Context) string {
	return s.flaki.NextValidIDString()
}
